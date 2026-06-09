#!/usr/bin/env bash
set -euo pipefail

APP_NAME="sub2api"
DEPLOY_DIR="${DEPLOY_DIR:-/opt/sub2api-deploy}"
COMPOSE_FILE="${DEPLOY_DIR}/docker-compose.yml"
ENV_FILE="${DEPLOY_DIR}/.env"
ENV_EXAMPLE_FILE="${DEPLOY_DIR}/.env.example"

COMPOSE_URL="https://raw.githubusercontent.com/Wei-Shaw/sub2api/main/deploy/docker-compose.local.yml"
ENV_EXAMPLE_URL="https://raw.githubusercontent.com/Wei-Shaw/sub2api/main/deploy/.env.example"

IMAGE="${IMAGE:-}"

if [ -z "$IMAGE" ]; then
  echo "[ERROR] IMAGE 为空，请在云效主机部署脚本中 export IMAGE=完整ACR镜像地址"
  exit 1
fi

echo "[INFO] Deploy image: $IMAGE"

run_as_root() {
  if [ "$(id -u)" = "0" ]; then
    "$@"
  elif command -v sudo >/dev/null 2>&1; then
    sudo "$@"
  else
    echo "[ERROR] 当前用户不是 root，且系统没有 sudo"
    exit 1
  fi
}

install_basic_tools() {
  if command -v curl >/dev/null 2>&1 && command -v openssl >/dev/null 2>&1; then
    return
  fi

  if command -v apt-get >/dev/null 2>&1; then
    run_as_root apt-get update -y
    run_as_root apt-get install -y curl openssl ca-certificates
  elif command -v yum >/dev/null 2>&1; then
    run_as_root yum install -y curl openssl ca-certificates
  elif command -v dnf >/dev/null 2>&1; then
    run_as_root dnf install -y curl openssl ca-certificates
  else
    echo "[ERROR] 未识别系统包管理器，请手动安装 curl 和 openssl"
    exit 1
  fi
}

install_docker_if_needed() {
  if command -v docker >/dev/null 2>&1; then
    echo "[INFO] Docker already installed"
    return
  fi

  echo "[INFO] Docker not found, installing..."
  install_basic_tools
  curl -fsSL https://get.docker.com | run_as_root sh

  if command -v systemctl >/dev/null 2>&1; then
    run_as_root systemctl enable docker
    run_as_root systemctl start docker
  else
    run_as_root service docker start || true
  fi
}

check_docker_compose() {
  if docker compose version >/dev/null 2>&1 || run_as_root docker compose version >/dev/null 2>&1; then
    echo "[INFO] Docker Compose v2 available"
    return
  fi

  echo "[ERROR] docker compose 插件不可用，请检查 Docker 安装"
  exit 1
}

docker_cmd() {
  if docker info >/dev/null 2>&1; then
    docker "$@"
  else
    run_as_root docker "$@"
  fi
}

generate_secret() {
  openssl rand -hex 32
}

download_file() {
  local url="$1"
  local output="$2"
  curl -fsSL "$url" -o "$output"
}

ensure_postgres_data_permissions() {
  local pg_dir="$DEPLOY_DIR/postgres_data"
  local pg_version="$pg_dir/PG_VERSION"
  local permission_status

  if [ ! -e "$pg_version" ]; then
    return
  fi

  permission_status="$(docker_cmd run --rm \
    -v "$pg_dir:/var/lib/postgresql/data" \
    --entrypoint sh \
    postgres:18-alpine \
    -c 'probe=/var/lib/postgresql/data/PG_VERSION; [ -e /var/lib/postgresql/data/global/pg_filenode.map ] && probe=/var/lib/postgresql/data/global/pg_filenode.map; expected="$(id -u postgres):$(id -g postgres)"; actual="$(stat -c "%u:%g" "$probe")"; mode="$(stat -c "%a" /var/lib/postgresql/data)"; if [ "$actual" = "$expected" ] && [ "$mode" = "700" ]; then echo ok; else echo repair; fi')"

  if [ "$permission_status" = "ok" ]; then
    return
  fi

  echo "[WARN] PostgreSQL data directory ownership/permissions need repair."
  echo "[INFO] Stop app and PostgreSQL containers before repair..."
  docker_cmd stop sub2api sub2api-postgres >/dev/null 2>&1 || true

  echo "[INFO] Repair PostgreSQL data directory ownership..."
  docker_cmd run --rm \
    -v "$pg_dir:/var/lib/postgresql/data" \
    --entrypoint sh \
    postgres:18-alpine \
    -c 'chown -R postgres:postgres /var/lib/postgresql/data && chmod 700 /var/lib/postgresql/data'
}

ensure_deploy_files() {
  echo "[INFO] Ensure deploy directory: $DEPLOY_DIR"

  run_as_root mkdir -p "$DEPLOY_DIR"
  run_as_root chown "$(id -u):$(id -g)" "$DEPLOY_DIR" || true

  mkdir -p "$DEPLOY_DIR/data" "$DEPLOY_DIR/postgres_data" "$DEPLOY_DIR/redis_data"
  # Do not recursively chown the whole deploy directory: postgres_data and
  # redis_data are bind-mounted service data directories whose ownership is
  # managed by their containers. Changing them from the host can make
  # PostgreSQL unable to read files such as global/pg_filenode.map.
  run_as_root chown -R "$(id -u):$(id -g)" "$DEPLOY_DIR/data" || true
  ensure_postgres_data_permissions

  if [ ! -f "$COMPOSE_FILE" ]; then
    echo "[INFO] Download docker-compose.yml..."
    download_file "$COMPOSE_URL" "$COMPOSE_FILE"
  fi

  if [ ! -f "$ENV_EXAMPLE_FILE" ]; then
    echo "[INFO] Download .env.example..."
    download_file "$ENV_EXAMPLE_URL" "$ENV_EXAMPLE_FILE"
  fi

  if [ ! -f "$ENV_FILE" ]; then
    echo "[INFO] Generate .env..."
    cp "$ENV_EXAMPLE_FILE" "$ENV_FILE"

    sed -i "s#^POSTGRES_PASSWORD=.*#POSTGRES_PASSWORD=$(generate_secret)#" "$ENV_FILE"
    sed -i "s#^JWT_SECRET=.*#JWT_SECRET=$(generate_secret)#" "$ENV_FILE"
    sed -i "s#^TOTP_ENCRYPTION_KEY=.*#TOTP_ENCRYPTION_KEY=$(generate_secret)#" "$ENV_FILE"
    sed -i "s#^ADMIN_EMAIL=.*#ADMIN_EMAIL=admin@sub2api.local#" "$ENV_FILE"

    chmod 600 "$ENV_FILE"
    echo "[WARN] ADMIN_PASSWORD 为空时，首次启动会自动生成；请通过日志查看"
  fi
}

set_env_value() {
  local key="$1"
  local value="$2"

  if grep -q "^${key}=" "$ENV_FILE"; then
    sed -i "s#^${key}=.*#${key}=${value}#" "$ENV_FILE"
  else
    echo "${key}=${value}" >> "$ENV_FILE"
  fi
}

patch_compose_image() {
  if grep -q "image: weishaw/sub2api:latest" "$COMPOSE_FILE"; then
    sed -i 's#image: weishaw/sub2api:latest#image: ${SUB2API_IMAGE}#' "$COMPOSE_FILE"
  fi

  set_env_value "SUB2API_IMAGE" "$IMAGE"
}

login_acr_if_configured() {
  if [ -n "${ACR_REGISTRY:-}" ] && [ -n "${ACR_USERNAME:-}" ] && [ -n "${ACR_PASSWORD:-}" ]; then
    echo "[INFO] Login ACR: $ACR_REGISTRY"
    echo "$ACR_PASSWORD" | docker_cmd login "$ACR_REGISTRY" -u "$ACR_USERNAME" --password-stdin
  else
    echo "[WARN] 未配置 ACR 登录变量。如果仓库是私有的，请确保 ECS 已经 docker login 过，或配置 ACR_REGISTRY/ACR_USERNAME/ACR_PASSWORD。"
  fi
}

deploy() {
  cd "$DEPLOY_DIR"

  echo "[INFO] Pull image..."
  docker_cmd pull "$IMAGE"

  echo "[INFO] Start services..."
  docker_cmd compose up -d

  echo "[INFO] Wait for health..."
  for _ in $(seq 1 40); do
    status="$(docker_cmd inspect --format='{{.State.Health.Status}}' "$APP_NAME" 2>/dev/null || true)"

    if [ "$status" = "healthy" ]; then
      echo "[SUCCESS] $APP_NAME is healthy"
      docker_cmd compose ps
      exit 0
    fi

    echo "[INFO] waiting... status=${status:-unknown}"
    sleep 3
  done

  echo "[ERROR] Health check failed. Recent logs:"
  docker_cmd compose logs --tail=120 sub2api
  docker_cmd compose ps
  exit 1
}

install_basic_tools
install_docker_if_needed
check_docker_compose
ensure_deploy_files
patch_compose_image
login_acr_if_configured
deploy

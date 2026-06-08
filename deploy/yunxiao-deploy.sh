#!/usr/bin/env bash
set -euo pipefail

APP_NAME="sub2api"
DEPLOY_DIR="${DEPLOY_DIR:-/opt/sub2api-deploy}"
COMPOSE_FILE="${DEPLOY_DIR}/docker-compose.yml"
ENV_FILE="${DEPLOY_DIR}/.env"
ENV_EXAMPLE_FILE="${DEPLOY_DIR}/.env.example"
SOURCE_DIR="${SOURCE_DIR:-}"
IMAGE="sub2api:local"

COMPOSE_URL="https://raw.githubusercontent.com/Wei-Shaw/sub2api/main/deploy/docker-compose.local.yml"
ENV_EXAMPLE_URL="https://raw.githubusercontent.com/Wei-Shaw/sub2api/main/deploy/.env.example"

echo "[INFO] Deploy local image: $IMAGE"

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

  echo "[INFO] Install curl/openssl..."

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

build_local_image() {
  if [ -z "$SOURCE_DIR" ]; then
    echo "[ERROR] SOURCE_DIR 为空，请在主机部署脚本里设置解压后的源码目录"
    exit 1
  fi

  if [ ! -f "$SOURCE_DIR/Dockerfile" ]; then
    echo "[ERROR] $SOURCE_DIR/Dockerfile 不存在，无法本地构建镜像"
    exit 1
  fi

  echo "[INFO] Build local image from source: $SOURCE_DIR"
  docker_cmd build \
    --build-arg GOPROXY=https://goproxy.cn,direct \
    --build-arg GOSUMDB=sum.golang.google.cn \
    -t "$IMAGE" \
    "$SOURCE_DIR"
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

ensure_deploy_files() {
  echo "[INFO] Ensure deploy directory: $DEPLOY_DIR"

  run_as_root mkdir -p "$DEPLOY_DIR"
  run_as_root chown -R "$(id -u):$(id -g)" "$DEPLOY_DIR" || true

  mkdir -p "$DEPLOY_DIR/data" "$DEPLOY_DIR/postgres_data" "$DEPLOY_DIR/redis_data"

  if [ ! -f "$COMPOSE_FILE" ]; then
    if [ -n "$SOURCE_DIR" ] && [ -f "$SOURCE_DIR/deploy/docker-compose.local.yml" ]; then
      echo "[INFO] Copy docker-compose.yml from source package..."
      cp "$SOURCE_DIR/deploy/docker-compose.local.yml" "$COMPOSE_FILE"
    else
      echo "[INFO] Download docker-compose.yml..."
      download_file "$COMPOSE_URL" "$COMPOSE_FILE"
    fi
  fi

  if [ ! -f "$ENV_EXAMPLE_FILE" ]; then
    if [ -n "$SOURCE_DIR" ] && [ -f "$SOURCE_DIR/deploy/.env.example" ]; then
      echo "[INFO] Copy .env.example from source package..."
      cp "$SOURCE_DIR/deploy/.env.example" "$ENV_EXAMPLE_FILE"
    else
      echo "[INFO] Download .env.example..."
      download_file "$ENV_EXAMPLE_URL" "$ENV_EXAMPLE_FILE"
    fi
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

deploy() {
  cd "$DEPLOY_DIR"

  echo "[INFO] Use local image: $IMAGE"

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
build_local_image
ensure_deploy_files
patch_compose_image
deploy

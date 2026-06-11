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
DOCKERHUB_IMAGE_PREFIX="${DOCKERHUB_IMAGE_PREFIX:-m.daocloud.io/docker.io/library}"
POSTGRES_IMAGE="${POSTGRES_IMAGE:-${DOCKERHUB_IMAGE_PREFIX}/postgres:18-alpine}"
REDIS_IMAGE="${REDIS_IMAGE:-${DOCKERHUB_IMAGE_PREFIX}/redis:8-alpine}"
POSTGRES_MAINTENANCE_IMAGE="${POSTGRES_MAINTENANCE_IMAGE:-$POSTGRES_IMAGE}"

SCRIPT_VERSION="2026-06-11-alinux-yum-explicit-baseurl"
echo "[INFO] Yunxiao deploy script version: $SCRIPT_VERSION"

if [ -z "$IMAGE" ]; then
  echo "[ERROR] IMAGE 为空，请在云效主机部署脚本中 export IMAGE=完整ACR镜像地址"
  exit 1
fi

echo "[INFO] Deploy image: $IMAGE"

run_as_root() {
  local env_args=()

  while [ "$#" -gt 0 ]; do
    case "$1" in
      *=*)
        env_args+=("$1")
        shift
        ;;
      *)
        break
        ;;
    esac
  done

  if [ "$(id -u)" = "0" ]; then
    if [ "${#env_args[@]}" -gt 0 ]; then
      env "${env_args[@]}" "$@"
    else
      "$@"
    fi
  elif command -v sudo >/dev/null 2>&1; then
    if [ "${#env_args[@]}" -gt 0 ]; then
      sudo env "${env_args[@]}" "$@"
    else
      sudo "$@"
    fi
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

probe_url() {
  local url="$1"
  curl -fsSLI --connect-timeout 8 --retry 2 --retry-delay 1 "$url" >/dev/null 2>&1
}

report_probe() {
  local name="$1"
  local url="$2"

  if probe_url "$url"; then
    echo "[INFO] Network probe OK: $name ($url)"
    return 0
  fi

  echo "[WARN] Network probe failed: $name ($url)"
  return 1
}

get_os_release_value() {
  local key="$1"

  if [ ! -r /etc/os-release ]; then
    return 1
  fi

  awk -F= -v target="$key" '$1 == target { gsub(/^"|"$/, "", $2); print $2; exit }' /etc/os-release
}

get_centos_compat_version() {
  local os_id version_id
  os_id="$(get_os_release_value ID)"
  version_id="$(get_os_release_value VERSION_ID)"
  version_id="${version_id%%.*}"

  case "$os_id" in
    alinux|alios|alibaba)
      case "$version_id" in
        2) echo "7" ;;
        3) echo "8" ;;
        4) echo "9" ;;
        *) echo "$version_id" ;;
      esac
      ;;
    *) echo "$version_id" ;;
  esac
}

get_deb_arch() {
  case "$(dpkg --print-architecture 2>/dev/null || uname -m)" in
    amd64|x86_64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    armhf|armv7l) echo "armhf" ;;
    *) echo "amd64" ;;
  esac
}

start_docker_service() {
  if command -v systemctl >/dev/null 2>&1; then
    run_as_root systemctl enable docker
    run_as_root systemctl start docker
  else
    run_as_root service docker start || true
  fi
}

install_docker_via_aliyun_apt() {
  local distro
  local codename
  local arch
  local mirror_repo
  local mirror_key_url
  local official_key_url

  distro="$(get_os_release_value ID)"
  codename="$(get_os_release_value VERSION_CODENAME)"
  arch="$(get_deb_arch)"

  if [ -z "$distro" ] || [ -z "$codename" ]; then
    echo "[WARN] 无法识别当前 apt 系统的发行版或代号，跳过阿里云 Docker 源安装"
    return 1
  fi

  case "$distro" in
    ubuntu|debian) ;;
    *)
      echo "[WARN] 当前 apt 系统发行版为 $distro，阿里云 Docker 源安装仅对 ubuntu/debian 启用"
      return 1
      ;;
  esac

  mirror_repo="https://mirrors.aliyun.com/docker-ce/linux/${distro}"
  mirror_key_url="${mirror_repo}/gpg"
  official_key_url="https://download.docker.com/linux/${distro}/gpg"

  echo "[INFO] Prepare Docker apt repo via Aliyun mirror: distro=${distro}, codename=${codename}, arch=${arch}"
  report_probe "Aliyun Docker repo" "${mirror_repo}/dists/${codename}/Release" || return 1

  run_as_root install -m 0755 -d /etc/apt/keyrings

  if probe_url "$mirror_key_url"; then
    echo "[INFO] Download Docker GPG key from Aliyun mirror"
    run_as_root curl -fsSL "$mirror_key_url" -o /etc/apt/keyrings/docker.asc
  elif probe_url "$official_key_url"; then
    echo "[WARN] Aliyun mirror GPG key unreachable, fallback to official Docker GPG key"
    run_as_root curl -fsSL "$official_key_url" -o /etc/apt/keyrings/docker.asc
  else
    echo "[WARN] Docker GPG key is unreachable from both Aliyun mirror and official source"
    return 1
  fi

  run_as_root chmod a+r /etc/apt/keyrings/docker.asc
  printf 'deb [arch=%s signed-by=/etc/apt/keyrings/docker.asc] %s %s stable\n' "$arch" "$mirror_repo" "$codename" \
    | run_as_root tee /etc/apt/sources.list.d/docker.list >/dev/null

  echo "[INFO] Install Docker CE from Aliyun mirror"
  run_as_root apt-get update -y
  run_as_root env DEBIAN_FRONTEND=noninteractive apt-get install -y \
    docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
}

install_docker_via_apt_fallback() {
  echo "[WARN] Fallback to distro packages: apt install docker.io"
  run_as_root apt-get update -y
  run_as_root env DEBIAN_FRONTEND=noninteractive apt-get install -y docker.io

  if run_as_root env DEBIAN_FRONTEND=noninteractive apt-get install -y docker-compose-v2; then
    echo "[INFO] Installed docker-compose-v2"
    return 0
  fi

  if run_as_root env DEBIAN_FRONTEND=noninteractive apt-get install -y docker-compose-plugin; then
    echo "[INFO] Installed docker-compose-plugin"
    return 0
  fi

  if run_as_root env DEBIAN_FRONTEND=noninteractive apt-get install -y docker-compose; then
    echo "[INFO] Installed legacy docker-compose binary"
    return 0
  fi

  echo "[WARN] Compose package installation skipped; will rely on whichever compose command is already available"
}

install_docker_via_aliyun_yum() {
  local os_id centos_ver rpm_arch repo_baseurl

  os_id="$(get_os_release_value ID)"
  centos_ver="$(get_centos_compat_version)"
  rpm_arch="$(uname -m)"
  repo_baseurl="https://mirrors.aliyun.com/docker-ce/linux/centos/${centos_ver}/${rpm_arch}/stable"

  echo "[INFO] Install Docker CE via Aliyun yum mirror: os=${os_id}, centos_compat=${centos_ver}, arch=${rpm_arch}"
  report_probe "Aliyun Docker yum repo" "${repo_baseurl}/" || return 1

  # Write repo file with explicit baseurl to avoid $releasever resolving to alinux version number
  run_as_root rm -f /etc/yum.repos.d/docker-ce.repo
  printf '[docker-ce-stable]\nname=Docker CE Stable - %s\nbaseurl=%s\nenabled=1\ngpgcheck=0\n' \
    "$rpm_arch" "$repo_baseurl" \
    | run_as_root tee /etc/yum.repos.d/docker-ce-aliyun.repo > /dev/null

  if command -v dnf >/dev/null 2>&1; then
    run_as_root dnf install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  elif command -v yum >/dev/null 2>&1; then
    run_as_root yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  else
    echo "[WARN] yum/dnf not found"
    return 1
  fi
}

install_docker_if_needed() {
  if command -v docker >/dev/null 2>&1; then
    echo "[INFO] Docker already installed"
    return
  fi

  echo "[INFO] Docker not found, installing..."
  install_basic_tools

  if command -v apt-get >/dev/null 2>&1; then
    if install_docker_via_aliyun_apt; then
      echo "[INFO] Docker installed from Aliyun mirror"
    else
      echo "[WARN] Aliyun mirror installation failed, trying distro fallback packages"
      install_docker_via_apt_fallback
    fi
  elif command -v yum >/dev/null 2>&1 || command -v dnf >/dev/null 2>&1; then
    if install_docker_via_aliyun_yum; then
      echo "[INFO] Docker installed via Aliyun yum mirror"
    else
      echo "[WARN] Aliyun yum installation failed, falling back to get.docker.com"
      report_probe "get.docker.com" "https://get.docker.com" || true
      report_probe "download.docker.com" "https://download.docker.com" || true
      curl -fsSL https://get.docker.com | run_as_root sh
    fi
  else
    report_probe "get.docker.com" "https://get.docker.com" || true
    report_probe "download.docker.com" "https://download.docker.com" || true
    curl -fsSL https://get.docker.com | run_as_root sh
  fi

  start_docker_service
}

check_docker_compose() {
  if docker compose version >/dev/null 2>&1 || run_as_root docker compose version >/dev/null 2>&1; then
    echo "[INFO] Docker Compose v2 available"
    return
  fi

  if docker-compose version >/dev/null 2>&1 || run_as_root docker-compose version >/dev/null 2>&1; then
    echo "[INFO] Legacy docker-compose available"
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

docker_compose() {
  if docker compose version >/dev/null 2>&1 || run_as_root docker compose version >/dev/null 2>&1; then
    docker_cmd compose "$@"
    return
  fi

  if docker-compose version >/dev/null 2>&1; then
    docker-compose "$@"
    return
  fi

  if run_as_root docker-compose version >/dev/null 2>&1; then
    run_as_root docker-compose "$@"
    return
  fi

  echo "[ERROR] docker compose / docker-compose 都不可用"
  exit 1
}

generate_secret() {
  openssl rand -hex 32
}

download_file() {
  local url="$1"
  local output="$2"
  curl -fsSL "$url" -o "$output"
}

postgres_mount_needs_repair() {
  local mount_spec="$1"
  local permission_status

  if ! permission_status="$(docker_cmd run --rm \
    -v "$mount_spec:/var/lib/postgresql/data" \
    --entrypoint sh \
    "$POSTGRES_MAINTENANCE_IMAGE" \
    -c '
      set -eu
      data=/var/lib/postgresql/data
      if [ ! -e "$data/PG_VERSION" ]; then
        echo empty
        exit 0
      fi

      expected="$(id -u postgres):$(id -g postgres)"
      data_owner="$(stat -c "%u:%g" "$data")"
      data_mode="$(stat -c "%a" "$data")"
      pg_version_owner="$(stat -c "%u:%g" "$data/PG_VERSION")"

      if [ "$data_owner" != "$expected" ] || [ "$data_mode" != "700" ] || [ "$pg_version_owner" != "$expected" ]; then
        echo repair
        exit 0
      fi

      if [ -e "$data/global" ]; then
        global_owner="$(stat -c "%u:%g" "$data/global")"
        if [ "$global_owner" != "$expected" ]; then
          echo repair
          exit 0
        fi
      fi

      if [ -e "$data/global/pg_filenode.map" ]; then
        filenode_owner="$(stat -c "%u:%g" "$data/global/pg_filenode.map")"
        if [ "$filenode_owner" != "$expected" ]; then
          echo repair
          exit 0
        fi
        if command -v su-exec >/dev/null 2>&1 && ! su-exec postgres test -r "$data/global/pg_filenode.map"; then
          echo repair
          exit 0
        fi
      fi

      echo ok
    ' 2>/dev/null)"; then
    permission_status="repair"
  fi

  [ "$permission_status" = "repair" ]
}

repair_postgres_data_mount() {
  local mount_spec="$1"

  echo "[WARN] PostgreSQL data directory ownership/permissions need repair."
  echo "[INFO] Stop app and PostgreSQL containers before repair..."
  docker_cmd stop sub2api sub2api-postgres >/dev/null 2>&1 || true

  echo "[INFO] Repair PostgreSQL data directory ownership..."
  docker_cmd run --rm \
    -v "$mount_spec:/var/lib/postgresql/data" \
    --entrypoint sh \
    "$POSTGRES_MAINTENANCE_IMAGE" \
    -c 'chown -R postgres:postgres /var/lib/postgresql/data && chmod 700 /var/lib/postgresql/data'
}

ensure_postgres_data_permissions_for_mount() {
  local mount_spec="$1"

  if [ -z "$mount_spec" ]; then
    return
  fi

  echo "[INFO] Check PostgreSQL data mount: $mount_spec"

  if [ -d "$mount_spec" ] && [ ! -e "$mount_spec/PG_VERSION" ]; then
    echo "[INFO] PostgreSQL data mount has no PG_VERSION yet; skip repair and let first startup initialize it."
    return
  fi

  if [ "${POSTGRES_DATA_FORCE_REPAIR:-true}" = "true" ]; then
    echo "[INFO] Force repair PostgreSQL data mount ownership before startup."
    repair_postgres_data_mount "$mount_spec"
    return
  fi

  if postgres_mount_needs_repair "$mount_spec"; then
    repair_postgres_data_mount "$mount_spec"
  else
    echo "[INFO] PostgreSQL data mount permissions look OK."
  fi
}

existing_postgres_data_mount() {
  docker_cmd inspect sub2api-postgres \
    --format '{{range .Mounts}}{{if eq .Destination "/var/lib/postgresql/data"}}{{if eq .Type "volume"}}{{.Name}}{{else}}{{.Source}}{{end}}{{end}}{{end}}' \
    2>/dev/null || true
}

ensure_postgres_data_permissions() {
  local local_mount="$DEPLOY_DIR/postgres_data"
  local current_mount

  current_mount="$(existing_postgres_data_mount)"
  ensure_postgres_data_permissions_for_mount "$current_mount"

  if [ "$current_mount" != "$local_mount" ]; then
    ensure_postgres_data_permissions_for_mount "$local_mount"
  fi
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

  ensure_postgres_data_permissions
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

  if grep -q "image: postgres:18-alpine" "$COMPOSE_FILE"; then
    sed -i 's#image: postgres:18-alpine#image: ${POSTGRES_IMAGE}#' "$COMPOSE_FILE"
  fi

  if grep -q "image: redis:8-alpine" "$COMPOSE_FILE"; then
    sed -i 's#image: redis:8-alpine#image: ${REDIS_IMAGE}#' "$COMPOSE_FILE"
  fi

  set_env_value "SUB2API_IMAGE" "$IMAGE"
  set_env_value "POSTGRES_IMAGE" "$POSTGRES_IMAGE"
  set_env_value "REDIS_IMAGE" "$REDIS_IMAGE"
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
  docker_compose up -d

  echo "[INFO] Wait for health..."
  for _ in $(seq 1 40); do
    status="$(docker_cmd inspect --format='{{.State.Health.Status}}' "$APP_NAME" 2>/dev/null || true)"

    if [ "$status" = "healthy" ]; then
      echo "[SUCCESS] $APP_NAME is healthy"
      docker_compose ps
      exit 0
    fi

    echo "[INFO] waiting... status=${status:-unknown}"
    sleep 3
  done

  echo "[ERROR] Health check failed. Recent logs:"
  docker_compose logs --tail=120 sub2api
  docker_compose ps
  exit 1
}

install_basic_tools
install_docker_if_needed
check_docker_compose
ensure_deploy_files
patch_compose_image
login_acr_if_configured
deploy

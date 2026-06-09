#!/usr/bin/env bash
set -Eeuo pipefail

ACR_REGISTRY="${ACR_REGISTRY:-crpi-eyduewu7n0l0vj8i.ap-northeast-1.personal.cr.aliyuncs.com}"
ACR_NAMESPACE="${ACR_NAMESPACE:-erqishi}"
PLATFORMS="${PLATFORMS:-linux/amd64}"
ENGINE="${ENGINE:-auto}"
COPY_MANIFEST="${COPY_MANIFEST:-false}"

NODE_IMAGE="${NODE_IMAGE:-node:24-alpine}"
GOLANG_IMAGE="${GOLANG_IMAGE:-golang:1.26.4-alpine}"
ALPINE_IMAGE="${ALPINE_IMAGE:-alpine:3.21}"
POSTGRES_IMAGE="${POSTGRES_IMAGE:-postgres:18-alpine}"

usage() {
  cat <<'EOF'
Sync Sub2API Docker base images to Alibaba Cloud ACR.

Required environment variables:
  ACR_USERNAME      Alibaba Cloud ACR username
  ACR_PASSWORD      Alibaba Cloud ACR password or access token

Optional environment variables:
  ACR_REGISTRY      Default: crpi-eyduewu7n0l0vj8i.ap-northeast-1.personal.cr.aliyuncs.com
  ACR_NAMESPACE     Default: erqishi
  PLATFORMS         Default: linux/amd64
  ENGINE            Default: auto. Supported: auto, docker, skopeo
  COPY_MANIFEST     Default: false. Set true to copy all platforms with skopeo
                    or Docker buildx imagetools.
  NODE_IMAGE        Default: node:24-alpine
  GOLANG_IMAGE      Default: golang:1.26.4-alpine
  ALPINE_IMAGE      Default: alpine:3.21
  POSTGRES_IMAGE    Default: postgres:18-alpine

Examples:
  export ACR_USERNAME='your-acr-user'
  export ACR_PASSWORD='your-acr-password'
  ./tools/push-base-images-to-acr.sh

  # Without Docker Desktop:
  brew install skopeo
  ENGINE=skopeo ./tools/push-base-images-to-acr.sh

EOF
}

log() {
  printf '[INFO] %s\n' "$*"
}

err() {
  printf '[ERROR] %s\n' "$*" >&2
}

has_cmd() {
  command -v "$1" >/dev/null 2>&1
}

require_cmd() {
  if ! has_cmd "$1"; then
    err "Missing command: $1"
    exit 1
  fi
}

acr_ref_for() {
  local source_ref="$1"
  local repo="${source_ref%%:*}"
  local tag="${source_ref#*:}"

  if [ "$repo" = "$tag" ]; then
    err "Image must include an explicit tag: $source_ref"
    exit 1
  fi

  printf '%s/%s/%s:%s' "$ACR_REGISTRY" "$ACR_NAMESPACE" "$repo" "$tag"
}

docker_sync() {
  local source_ref="$1"
  local target_ref="$2"

  if [ "$COPY_MANIFEST" = "true" ]; then
    log "Copying manifest $source_ref -> $target_ref"
    docker buildx imagetools create --tag "$target_ref" "$source_ref"
    return
  fi

  IFS=',' read -r first_platform _rest <<<"$PLATFORMS"
  if [ "$PLATFORMS" != "$first_platform" ]; then
    err "Docker pull/tag/push mode supports one platform only. Use COPY_MANIFEST=true for full manifest copy."
    exit 1
  fi

  log "Pulling $source_ref for $first_platform"
  docker pull --platform "$first_platform" "$source_ref"
  log "Tagging $source_ref -> $target_ref"
  docker tag "$source_ref" "$target_ref"
  log "Pushing $target_ref"
  docker push "$target_ref"
}

skopeo_sync() {
  local source_ref="$1"
  local target_ref="$2"

  if [ "$COPY_MANIFEST" = "true" ]; then
    log "Copying all platforms $source_ref -> $target_ref"
    skopeo copy --all "docker://$source_ref" "docker://$target_ref"
    return
  fi

  IFS=',' read -r first_platform _rest <<<"$PLATFORMS"
  if [ "$PLATFORMS" != "$first_platform" ]; then
    err "Skopeo single-platform mode supports one platform only. Use COPY_MANIFEST=true for full manifest copy."
    exit 1
  fi

  local os arch variant
  os="${first_platform%%/*}"
  arch="${first_platform#*/}"
  variant=""
  if [ "$arch" != "${arch#*/}" ]; then
    variant="${arch#*/}"
    arch="${arch%%/*}"
  fi

  log "Copying $source_ref -> $target_ref ($first_platform)"
  if [ -n "$variant" ]; then
    skopeo copy --override-os "$os" --override-arch "$arch" --override-variant "$variant" "docker://$source_ref" "docker://$target_ref"
  else
    skopeo copy --override-os "$os" --override-arch "$arch" "docker://$source_ref" "docker://$target_ref"
  fi
}

main() {
  if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
    usage
    exit 0
  fi

  local engine="$ENGINE"
  if [ "$engine" = "auto" ]; then
    if has_cmd docker; then
      engine="docker"
    elif has_cmd skopeo; then
      engine="skopeo"
    else
      err "Missing command: docker or skopeo"
      err "On macOS without Docker Desktop, run: brew install skopeo"
      exit 1
    fi
  fi

  case "$engine" in
    docker)
      require_cmd docker
      if [ "$COPY_MANIFEST" = "true" ] && ! docker buildx version >/dev/null 2>&1; then
        err "Docker buildx is required when COPY_MANIFEST=true."
        exit 1
      fi
      ;;
    skopeo)
      require_cmd skopeo
      ;;
    *)
      err "Unsupported ENGINE: $ENGINE"
      exit 1
      ;;
  esac

  if [ -z "${ACR_USERNAME:-}" ] || [ -z "${ACR_PASSWORD:-}" ]; then
    usage
    err "ACR_USERNAME and ACR_PASSWORD are required."
    exit 1
  fi

  log "Using image copy engine: $engine"
  log "Logging in to ACR registry: $ACR_REGISTRY"
  if [ "$engine" = "docker" ]; then
    printf '%s' "$ACR_PASSWORD" | docker login "$ACR_REGISTRY" -u "$ACR_USERNAME" --password-stdin
  else
    printf '%s' "$ACR_PASSWORD" | skopeo login "$ACR_REGISTRY" -u "$ACR_USERNAME" --password-stdin
  fi

  local node_target golang_target alpine_target postgres_target
  node_target="$(acr_ref_for "$NODE_IMAGE")"
  golang_target="$(acr_ref_for "$GOLANG_IMAGE")"
  alpine_target="$(acr_ref_for "$ALPINE_IMAGE")"
  postgres_target="$(acr_ref_for "$POSTGRES_IMAGE")"

  if [ "$engine" = "docker" ]; then
    docker_sync "$NODE_IMAGE" "$node_target"
    docker_sync "$GOLANG_IMAGE" "$golang_target"
    docker_sync "$ALPINE_IMAGE" "$alpine_target"
    docker_sync "$POSTGRES_IMAGE" "$postgres_target"
  else
    skopeo_sync "$NODE_IMAGE" "$node_target"
    skopeo_sync "$GOLANG_IMAGE" "$golang_target"
    skopeo_sync "$ALPINE_IMAGE" "$alpine_target"
    skopeo_sync "$POSTGRES_IMAGE" "$postgres_target"
  fi

  cat <<EOF

[OK] Base images have been pushed to ACR.

Use these Docker build args in Aliyun Flow:

--build-arg NODE_IMAGE=$node_target
--build-arg GOLANG_IMAGE=$golang_target
--build-arg ALPINE_IMAGE=$alpine_target
--build-arg POSTGRES_IMAGE=$postgres_target

EOF
}

main "$@"

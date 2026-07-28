#!/usr/bin/env bash
# RouteX 启动 / 更新二合一:拉取 IMAGE_TAG 通道(默认 latest)的最新镜像并让改动生效。
# - 服务没起:直接全量启动
# - 服务在跑:只重建镜像有更新的容器(postgres/redis 没动就不重启)
# 放在 docker-compose.yml 同目录,执行:./restart.sh
set -euo pipefail
cd "$(dirname "$0")"

if [ ! -f .env ]; then
  echo "!! .env not found. Run: cp .env.example .env  (then edit it)"; exit 1
fi

env_value() {
  local key="$1"
  local line value
  line=$(grep -E "^${key}=" .env | tail -n 1 || true)
  value="${line#*=}"
  value="${value%$'\r'}"
  if [[ "$value" == \"*\" && "$value" == *\" ]]; then
    value="${value:1:${#value}-2}"
  elif [[ "$value" == \'*\' && "$value" == *\' ]]; then
    value="${value:1:${#value}-2}"
  fi
  printf '%s' "$value"
}

fail_preflight() {
  echo "!! Deployment preflight failed: $1" >&2
  exit 1
}

# ---- 必填项预检 ----
# 这几项是上线前必须改成真实值的;缺了/还是占位符就中止,不重启容器。
jwt_secret=$(env_value APP_JWT_SECRET)
pg_password=$(env_value POSTGRES_PASSWORD)
public_host=$(env_value PUBLIC_HOST)
reality_pub=$(env_value REALITY_PUBLIC_KEY)

[[ -n "$jwt_secret" && "$jwt_secret" != *REPLACE-ME* && "$jwt_secret" != "dev-secret-change-me" ]] \
  || fail_preflight "APP_JWT_SECRET 未设置或仍是占位符(用 openssl rand -hex 32 生成)"
[[ -n "$pg_password" && "$pg_password" != *REPLACE-ME* ]] \
  || fail_preflight "POSTGRES_PASSWORD 未设置或仍是占位符"
[[ -n "$public_host" && "$public_host" != *yourdomain* ]] \
  || fail_preflight "PUBLIC_HOST 未设置(填你的公网域名或 IP)"
[[ -n "$reality_pub" && "$reality_pub" != *REPLACE-ME* ]] \
  || fail_preflight "REALITY_PUBLIC_KEY 未设置(跑 docker run --rm teddysun/xray xray x25519)"

# Xray 真实配置必须存在(含 Reality 私钥,生成见同目录 README.md)。
[[ -f ./xray/config.json ]] \
  || fail_preflight "./xray/config.json 不存在(见 README:cp deploy/xray/config.example.json 并填 privateKey)"
! grep -q "REPLACE_WITH_xray_x25519_PRIVATE_KEY" ./xray/config.json \
  || fail_preflight "./xray/config.json 的 Reality privateKey 还是占位符"

image_tag=$(env_value IMAGE_TAG)
echo ">> Deployment preflight passed: IMAGE_TAG=${image_tag:-latest} PUBLIC_HOST=$public_host"

if [[ "${PREFLIGHT_ONLY:-0}" == "1" ]]; then
  exit 0
fi

echo ">> Pulling images..."
docker compose pull

echo ">> Starting / recreating changed containers..."
docker compose up -d --remove-orphans

echo ">> Cleaning up superseded image layers..."
docker image prune -f >/dev/null

echo
echo ">> Running versions:"
docker compose images
echo
docker compose ps

#!/usr/bin/env bash
# RouteX VPS 一键安装
#
# 用法（在 VPS 上）：
#   git clone <repo> /srv/routex && cd /srv/routex
#   ./install.sh
#
# 第二次运行也安全：已存在的 .env / Reality 密钥不会被覆盖。
#
# 它会自动：
#   1. 装 Docker（若未装）
#   2. 跑 xray x25519 生成 Reality 密钥，写入 deploy/xray/config.json
#   3. 生成强随机 JWT_SECRET / POSTGRES_PASSWORD，写入 .env
#   4. 问你一次公网地址（PUBLIC_HOST）
#   5. docker compose up -d --build
#   6. 打印默认管理员账号密码

set -euo pipefail
cd "$(dirname "$0")"

# ---------- 工具 ----------
red()    { printf "\033[31m%s\033[0m\n" "$*"; }
green()  { printf "\033[32m%s\033[0m\n" "$*"; }
yellow() { printf "\033[33m%s\033[0m\n" "$*"; }
bold()   { printf "\033[1m%s\033[0m\n" "$*"; }

sed_inplace() {
  if [[ "$(uname)" == "Darwin" ]]; then sed -i '' "$@"
  else sed -i "$@"; fi
}

random_hex() {
  local n=$1
  if command -v openssl >/dev/null 2>&1; then openssl rand -hex "$n"
  else head -c "$((n*2))" /dev/urandom | od -An -tx1 | tr -d ' \n' | head -c "$((n*2))"
  fi
}

# ---------- 1. Docker ----------
if ! command -v docker >/dev/null 2>&1; then
  yellow "==> Docker 未安装，准备装"
  if [ "$(id -u)" -ne 0 ] && ! sudo -n true 2>/dev/null; then
    red "需要 sudo 权限装 Docker。请先以 root 运行或确认 sudo 可用。"
    exit 1
  fi
  curl -fsSL https://get.docker.com | sh
  if [ "$(id -u)" -ne 0 ]; then
    sudo usermod -aG docker "$USER" || true
    yellow "    已把当前用户加入 docker 组；如果接下来 docker 命令报 permission denied，请重新登录后再跑 ./install.sh"
  fi
else
  green "==> Docker 已就绪"
fi

# ---------- 2. Reality 密钥 ----------
if [ ! -f deploy/xray/config.json ] || grep -q "REPLACE_WITH_xray_x25519_PRIVATE_KEY" deploy/xray/config.json; then
  bold "==> 生成 Reality 密钥对"
  cp -n deploy/xray/config.example.json deploy/xray/config.json
  KEYS=$(docker run --rm teddysun/xray xray x25519)
  REALITY_PRIV=$(echo "$KEYS" | awk -F': *' '/PrivateKey:/ {print $2}')
  REALITY_PUB=$(echo "$KEYS"  | awk -F': *' '/PublicKey/ {print $2}')
  if [ -z "$REALITY_PRIV" ] || [ -z "$REALITY_PUB" ]; then
    red "解析 x25519 输出失败，原始输出："
    echo "$KEYS"
    exit 1
  fi
  sed_inplace "s|REPLACE_WITH_xray_x25519_PRIVATE_KEY|$REALITY_PRIV|" deploy/xray/config.json
  green "    Reality PrivateKey 已写入 deploy/xray/config.json"
  green "    Reality PublicKey:  $REALITY_PUB"
else
  green "==> Reality 密钥已存在，跳过"
  REALITY_PUB=""  # 稍后从 .env 里读（如果有）
fi

# ---------- 3. .env ----------
if [ ! -f .env ]; then
  bold "==> 生成 .env"

  # 问公网地址
  if [ -z "${PUBLIC_HOST:-}" ]; then
    read -rp "你的 VPS 公网域名或 IP (客户端用这个连): " PUBLIC_HOST
    while [ -z "$PUBLIC_HOST" ]; do
      yellow "    不能为空"
      read -rp "    再来一次: " PUBLIC_HOST
    done
  fi

  cp .env.prod.example .env

  JWT_SECRET=$(random_hex 32)
  PG_PASSWORD=$(random_hex 16)

  sed_inplace "s|<REPLACE-ME-with-\`openssl rand -hex 32\`>|$JWT_SECRET|" .env
  sed_inplace "s|<REPLACE-ME-strong-random>|$PG_PASSWORD|" .env
  sed_inplace "s|vpn.yourdomain.com|$PUBLIC_HOST|" .env
  sed_inplace "s|admin.yourdomain.com|$PUBLIC_HOST|" .env

  if [ -n "$REALITY_PUB" ]; then
    sed_inplace "s|<REPLACE-ME-with-xray-x25519-PublicKey>|$REALITY_PUB|" .env
  fi

  chmod 600 .env
  green "    .env 已生成 (chmod 600)"
else
  green "==> .env 已存在，跳过"
fi

# ---------- 4. 起服务 ----------
bold "==> docker compose up（首次约 2 分钟编译，喝口水）"
docker compose -f docker-compose.prod.yml up -d --build

# ---------- 5. 等 admin bootstrap + 打印凭证 ----------
bold "==> 等后端启动 + 自动创建默认管理员"
ADMIN_LINE=""
for i in $(seq 1 60); do
  ADMIN_LINE=$(docker compose -f docker-compose.prod.yml logs server 2>/dev/null | grep "BOOTSTRAP_ADMIN username" | head -1 || true)
  if [ -n "$ADMIN_LINE" ]; then break; fi
  sleep 2
done

echo ""
green "============================================================"
green "✓ 部署完成"
green "============================================================"
echo ""
echo "管理面板:    http://$(grep -E '^PUBLIC_HOST=' .env | cut -d= -f2-)"
echo "             （如果用了非 80 端口，加 :端口）"
echo ""
echo "VLESS 入口:  $(grep -E '^PUBLIC_HOST=' .env | cut -d= -f2-):$(grep -E '^PUBLIC_PORT=' .env | cut -d= -f2-)"
echo ""
if [ -n "$ADMIN_LINE" ]; then
  bold "默认管理员（一定要抄下来）:"
  echo "    $ADMIN_LINE"
else
  yellow "没等到 BOOTSTRAP_ADMIN 行。手动看："
  echo "    docker compose -f docker-compose.prod.yml logs server | grep BOOTSTRAP_ADMIN"
fi
echo ""
echo "常用命令:"
echo "    docker compose -f docker-compose.prod.yml ps        # 看状态"
echo "    docker compose -f docker-compose.prod.yml logs -f   # 看日志"
echo "    docker compose -f docker-compose.prod.yml restart   # 重启"
echo "    docker compose -f docker-compose.prod.yml down      # 停（保留数据）"
echo ""

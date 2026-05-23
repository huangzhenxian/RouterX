#!/usr/bin/env bash
# 节点 VPS 一键安装。只装 Xray + agent，不装控制台。
#
# 你需要预先在【控制台】上准备好：
#   1. 控制台对外可达 URL（如 http://your-control:8891 或 https://api.yourdomain.com）
#   2. 在控制台 → 节点 → 新增节点 拿到的 token（弹窗里那串）
#   3. 控制台的 Reality privateKey （在控制台机器 deploy/xray/config.json 里复制）
#
# 这台节点机器上：
#   git clone <repo> /srv/routex-node && cd /srv/routex-node
#   ./install-node.sh

set -euo pipefail
cd "$(dirname "$0")"

red()    { printf "\033[31m%s\033[0m\n" "$*"; }
green()  { printf "\033[32m%s\033[0m\n" "$*"; }
yellow() { printf "\033[33m%s\033[0m\n" "$*"; }
bold()   { printf "\033[1m%s\033[0m\n" "$*"; }

sed_inplace() {
  if [[ "$(uname)" == "Darwin" ]]; then sed -i '' "$@"
  else sed -i "$@"; fi
}

# ---------- 1. Docker ----------
if ! command -v docker >/dev/null 2>&1; then
  yellow "==> 装 Docker"
  curl -fsSL https://get.docker.com | sh
  if [ "$(id -u)" -ne 0 ]; then
    sudo usermod -aG docker "$USER" || true
    yellow "    已加入 docker 组；如下面权限报错，重新登录后再跑 ./install-node.sh"
  fi
else
  green "==> Docker 已就绪"
fi

# ---------- 2. 准备 .env.node ----------
if [ -f .env.node ]; then
  green "==> .env.node 已存在，跳过交互配置"
  set -a; source .env.node; set +a
else
  bold "==> 配置节点参数"

  read -rp "控制台 URL (如 http://1.2.3.4:8891): " ROUTEX_API_URL
  while [ -z "$ROUTEX_API_URL" ]; do read -rp "    不能为空: " ROUTEX_API_URL; done

  read -rp "节点 token (控制台新增节点弹窗里那串): " ROUTEX_NODE_TOKEN
  while [ -z "$ROUTEX_NODE_TOKEN" ]; do read -rp "    不能为空: " ROUTEX_NODE_TOKEN; done

  read -rp "VLESS 对外端口 [8895]: " PUBLIC_PORT
  PUBLIC_PORT=${PUBLIC_PORT:-8895}

  read -rp "心跳间隔 [30s]: " ROUTEX_HEARTBEAT_INTERVAL
  ROUTEX_HEARTBEAT_INTERVAL=${ROUTEX_HEARTBEAT_INTERVAL:-30s}

  cat > .env.node <<EOF
# 节点参数（install-node.sh 生成）
ROUTEX_API_URL=$ROUTEX_API_URL
ROUTEX_NODE_TOKEN=$ROUTEX_NODE_TOKEN
ROUTEX_HEARTBEAT_INTERVAL=$ROUTEX_HEARTBEAT_INTERVAL
PUBLIC_PORT=$PUBLIC_PORT
EOF
  chmod 600 .env.node
  green "    .env.node 已生成 (chmod 600)"
fi

# ---------- 3. Reality privateKey ----------
need_key=true
if [ -f deploy/xray/config.json ] && ! grep -q "REPLACE_WITH_xray_x25519_PRIVATE_KEY" deploy/xray/config.json; then
  need_key=false
fi

if $need_key; then
  bold "==> 配置 Reality privateKey"
  echo "    需要跟控制台用同一对 Reality 密钥，否则订阅链接连不上这个节点。"
  echo "    在【控制台机器】上执行："
  echo "      grep privateKey deploy/xray/config.json"
  echo "    复制双引号里那串，粘到下面。"
  read -rp "Reality privateKey: " REALITY_PRIV
  while [ -z "$REALITY_PRIV" ]; do read -rp "    不能为空: " REALITY_PRIV; done

  cp -n deploy/xray/config.example.json deploy/xray/config.json
  sed_inplace "s|REPLACE_WITH_xray_x25519_PRIVATE_KEY|$REALITY_PRIV|" deploy/xray/config.json
  green "    privateKey 已写入 deploy/xray/config.json (gitignored)"
else
  green "==> deploy/xray/config.json 已配置好 privateKey"
fi

# ---------- 4. 起服务 ----------
bold "==> docker compose up（首次约 1 分钟编译 agent）"
docker compose -f docker-compose.node.yml --env-file .env.node up -d --build

# ---------- 5. 验证 ----------
sleep 5
bold "==> 状态"
docker compose -f docker-compose.node.yml ps

echo ""
green "============================================================"
green "✓ 节点已上线"
green "============================================================"
echo ""
echo "VLESS 监听:   :${PUBLIC_PORT}"
echo "Agent 上报:   $ROUTEX_API_URL ← 每 $ROUTEX_HEARTBEAT_INTERVAL 一次"
echo ""
echo "去控制台面板 → 节点 ，几秒后应该看到这台节点变成 🟢 在线 + CPU/内存数据"
echo ""
echo "常用命令:"
echo "    docker compose -f docker-compose.node.yml logs -f xray    # Xray 日志"
echo "    docker compose -f docker-compose.node.yml logs -f agent   # agent 心跳"
echo "    docker compose -f docker-compose.node.yml restart         # 重启"
echo "    docker compose -f docker-compose.node.yml down            # 停"
echo ""
yellow "⚠️  当前限制：控制台还不会把【加用户 / 切出口】push 到节点。"
yellow "    这台节点能上线显示心跳，但 VLESS 连进来还连不上。"
yellow "    多节点用户同步是下一步要做的工作。"

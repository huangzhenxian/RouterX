#!/usr/bin/env bash
# RouteX 一键启动：docker 基础设施 + Go 后端 + React 前端，并 tail 全部日志。
# Ctrl+C 只退出 tail，服务继续在后台运行；要真正停服务请运行 ./stop.sh

set -e
ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

mkdir -p logs .pids

# ---------- 检查依赖 ----------
for cmd in docker go node npm; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "✗ 缺少依赖：$cmd" >&2
    exit 1
  fi
done

# ---------- 检查 .env ----------
if [ ! -f .env ]; then
  echo "==> .env 不存在，从 .env.example 拷贝"
  cp .env.example .env
fi

# ---------- 防重复启动 ----------
already_running=false
for pidfile in .pids/*.pid; do
  [ -f "$pidfile" ] || continue
  pid=$(cat "$pidfile" 2>/dev/null || echo "")
  if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
    already_running=true
    break
  fi
done
if [ "$already_running" = true ]; then
  echo "✗ 检测到服务已在运行（.pids/ 里有活进程）。请先 ./stop.sh"
  exit 1
fi

# ---------- 1. Docker 基础设施 ----------
echo "==> 启动 docker 服务（postgres / redis / xray / sing-box）"
docker compose up -d postgres redis xray sing-box

# 把 docker 服务日志聚合输出到一个文件，方便 tail -f
( docker compose logs -f --no-log-prefix postgres redis xray sing-box \
    > logs/docker.log 2>&1 ) &
DOCKER_LOG_PID=$!
echo $DOCKER_LOG_PID > .pids/docker-logs.pid
disown $DOCKER_LOG_PID 2>/dev/null || true

# ---------- 2. Go 后端 ----------
echo "==> 启动 Go 后端（server/cmd/server）"
: > logs/server.log
( cd server && exec go run ./cmd/server ) >> logs/server.log 2>&1 &
SERVER_PID=$!
echo $SERVER_PID > .pids/server.pid
disown $SERVER_PID 2>/dev/null || true

# ---------- 3. React 前端 ----------
if [ ! -d web/node_modules ]; then
  echo "==> 第一次启动，安装前端依赖（这一步可能耗时 1-2 分钟）"
  ( cd web && npm install )
fi

echo "==> 启动 React 前端（web）"
: > logs/web.log
( cd web && exec npm run dev ) >> logs/web.log 2>&1 &
WEB_PID=$!
echo $WEB_PID > .pids/web.pid
disown $WEB_PID 2>/dev/null || true

# ---------- 提示 ----------
sleep 1
cat <<EOF

✓ 服务已启动
  - 后端:  http://localhost:8080/v1/health   (pid $SERVER_PID, 日志 logs/server.log)
  - 前端:  http://localhost:5173             (pid $WEB_PID,    日志 logs/web.log)
  - Docker:                                  (pid $DOCKER_LOG_PID, 日志 logs/docker.log)

==> 实时日志（Ctrl+C 仅退出日志窗口，服务不停；./stop.sh 才会停服务）

EOF

# tail 全部日志，多文件会自动打 ==> filename <== 分隔
exec tail -F logs/docker.log logs/server.log logs/web.log

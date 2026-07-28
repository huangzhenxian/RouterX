#!/usr/bin/env bash
# 停止 RouteX 所有服务：Go 后端 / React 前端 / docker 基础设施。
# 数据卷（postgres/redis）保留，不会删数据。

set -u
ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

kill_pidfile() {
  local name=$1 pidfile=$2
  [ -f "$pidfile" ] || return 0
  local pid
  pid=$(cat "$pidfile" 2>/dev/null || echo "")
  if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
    # 先杀子进程（go run / npm 都会 fork 真正的子）
    pkill -TERM -P "$pid" 2>/dev/null || true
    kill -TERM "$pid" 2>/dev/null || true
    sleep 0.5
    if kill -0 "$pid" 2>/dev/null; then
      pkill -KILL -P "$pid" 2>/dev/null || true
      kill -KILL "$pid" 2>/dev/null || true
    fi
    echo "  ✓ 停止 $name (pid $pid)"
  else
    echo "  - $name 未在运行"
  fi
  rm -f "$pidfile"
}

# 兜底：按端口杀（防 pidfile 丢失或孤儿进程残留）
kill_by_port() {
  local port=$1 name=$2
  local pids
  pids=$(lsof -ti:"$port" 2>/dev/null || true)
  if [ -n "$pids" ]; then
    echo "  ✓ 兜底杀 $name 端口 $port: $pids"
    echo "$pids" | xargs kill -TERM 2>/dev/null || true
    sleep 0.3
    pids=$(lsof -ti:"$port" 2>/dev/null || true)
    [ -n "$pids" ] && echo "$pids" | xargs kill -KILL 2>/dev/null || true
  fi
}

echo "==> 停止本地进程"
kill_pidfile "Go 后端"   .pids/server.pid
kill_pidfile "React 前端" .pids/web.pid
kill_pidfile "docker 日志" .pids/docker-logs.pid

kill_by_port 8891 "Go 后端"
kill_by_port 8890 "React 前端"

echo "==> 停止 docker 服务"
docker compose down

# 清理空目录
rmdir .pids 2>/dev/null || true

echo "✓ 全部停止（数据卷保留，下次 ./start.sh 直接接着用）"

#!/usr/bin/env bash
# 比较 .env 与 .env.example 是否漂移：
#   1. .env.example 里新增的 key 在 .env 里缺失 → 大概率会跑挂
#   2. PORT/HOST 类配置值与模板不一致 → 可能有意为之，也可能忘了同步
#
# 永远 exit 0：本脚本只警告不阻断；用户自己判断要不要重置 .env。

set -u
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
EXAMPLE="$ROOT/.env.example"
ENV_FILE="$ROOT/.env"

[ -f "$ENV_FILE" ] || exit 0
[ -f "$EXAMPLE" ]  || exit 0

extract_keys() {
  grep -E '^[A-Z_][A-Z0-9_]*=' "$1" 2>/dev/null | sed 's/=.*//'
}

example_keys=$(extract_keys "$EXAMPLE" | sort -u)
env_keys=$(extract_keys "$ENV_FILE" | sort -u)
missing=$(comm -23 <(echo "$example_keys") <(echo "$env_keys"))

drifted=0

if [ -n "$missing" ]; then
  echo ""
  echo "⚠️  .env 缺少以下 key（.env.example 里有新增）："
  while IFS= read -r k; do
    [ -n "$k" ] && echo "     • $k"
  done <<< "$missing"
  drifted=1
fi

# PORT/HOST 类 key 值漂移检测
port_host_drift=""
while IFS= read -r key; do
  [ -z "$key" ] && continue
  ev=$(grep -E "^${key}=" "$EXAMPLE"  | head -1 | cut -d= -f2-)
  uv=$(grep -E "^${key}=" "$ENV_FILE" | head -1 | cut -d= -f2-)
  if [ -n "$uv" ] && [ "$ev" != "$uv" ]; then
    port_host_drift+="     • ${key}: 你的=${uv}   模板=${ev}"$'\n'
  fi
done <<< "$(extract_keys "$EXAMPLE" | grep -E '(PORT|HOST)$' || true)"

if [ -n "$port_host_drift" ]; then
  echo ""
  echo "⚠️  以下端口/主机配置与 .env.example 不一致（可能是你有意改的，请确认）："
  printf '%s' "$port_host_drift"
  drifted=1
fi

if [ "$drifted" -eq 1 ]; then
  echo ""
  echo "   修复方式："
  echo "     make reset           # 一键重置 .env（保留 DB 数据）"
  echo "     或手动编辑 .env 与 .env.example 对齐"
  echo ""
fi

exit 0

.PHONY: help up down logs ps server web migrate fmt reset nuke check-env

help:
	@echo "Targets:"
	@echo "  up        启动 docker compose 基础设施（pg/redis/xray/sing-box）"
	@echo "  down      停止全部容器"
	@echo "  logs      查看 compose 实时日志"
	@echo "  ps        查看容器状态"
	@echo "  server    本地启动 Go 后端"
	@echo "  web       本地启动 React 前端"
	@echo "  check-env 检查 .env 是否与 .env.example 漂移"
	@echo "  reset     停服 + 重置 .env（保留 DB 数据）"
	@echo "  nuke      停服 + 重置 .env + 删 DB/Redis 数据卷（不可恢复）"

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f --tail=200

ps:
	docker compose ps

server:
	cd server && go run ./cmd/server

web:
	cd web && npm run dev

fmt:
	cd server && go fmt ./...

check-env:
	./scripts/check-env.sh

reset:
	-./stop.sh
	rm -f .env
	@echo ""
	@echo "✓ 已重置 .env（DB 数据卷保留）。运行 ./start.sh 继续。"

nuke:
	-./stop.sh
	rm -f .env
	docker compose down -v
	@echo ""
	@echo "✓ 已清空 .env + DB/Redis 数据卷。运行 ./start.sh 全新起。"

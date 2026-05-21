.PHONY: help up down logs ps server web migrate fmt

help:
	@echo "Targets:"
	@echo "  up        启动 docker compose 基础设施（pg/redis/xray/sing-box）"
	@echo "  down      停止全部容器"
	@echo "  logs      查看 compose 实时日志"
	@echo "  ps        查看容器状态"
	@echo "  server    本地启动 Go 后端"
	@echo "  web       本地启动 React 前端"

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

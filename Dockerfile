# RouteX 单镜像全栈：一个容器里跑 nginx(前端+反代) + Go 后端。
# postgres / redis / xray 仍走官方镜像，由 deploy/registry/docker-compose.yml 编排。
#
# 构建上下文是仓库根：
#   docker build -t registry.cn-hongkong.aliyuncs.com/mewtwo_zero/routex:latest .

# ---------- 阶段 1：前端构建 ----------
FROM node:20-alpine AS web-builder
WORKDIR /src
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# ---------- 阶段 2：后端构建 ----------
FROM golang:1.26-alpine AS server-builder
WORKDIR /src
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---------- 阶段 3：运行时 ----------
# alpine + nginx + supervisor：supervisord 作为 PID 1 拉起并守护两个进程。
FROM nginx:1.27-alpine
RUN apk add --no-cache ca-certificates tzdata supervisor

# 后端二进制
COPY --from=server-builder /out/server /app/server
# 前端静态产物
COPY --from=web-builder /src/dist /usr/share/nginx/html
# nginx 站点配置（反代 /v1 -> 127.0.0.1:8891，SPA fallback）
COPY deploy/nginx/default.conf /etc/nginx/conf.d/default.conf
# 进程编排
COPY deploy/supervisord.conf /etc/supervisord.conf

# 80 = 管理面板(前端+API 同源)；后端 8891 只走容器内回环，不对外。
EXPOSE 80

ENTRYPOINT ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]

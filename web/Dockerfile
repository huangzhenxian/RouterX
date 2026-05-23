# 多阶段：node 构建静态产物 → 用 nginx 提供 + 反代 /v1
# 构建上下文是项目根（见 docker-compose.prod.yml 的 build.context: .），
# 所以这里所有 COPY 路径都从仓库根开始。
FROM node:20-alpine AS builder
WORKDIR /src
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM nginx:1.27-alpine
COPY --from=builder /src/dist /usr/share/nginx/html
COPY deploy/nginx/default.conf /etc/nginx/conf.d/default.conf
EXPOSE 80

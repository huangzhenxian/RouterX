# RouteX

基于 Xray-core + Go + React 的多用户代理管理平台。

## 目录结构

```
RouteX/
├── server/             # Go 后端（Gin + GORM + Zap）
├── web/                # React 前端（Vite + TS + 手写 shadcn/ui）
├── deploy/
│   ├── xray/           # Xray-core 配置
│   └── registry/       # 服务器镜像版部署栈（compose + restart.sh）
├── migrations/         # 数据库迁移 SQL
├── docker-compose.yml  # 一键起本地依赖（pg/redis/xray）
├── Dockerfile          # 单镜像全栈（前端+后端+nginx+supervisord）
├── .env.example
└── 需求.md
```

## 快速启动（开发环境）

```bash
./start.sh    # 启动所有服务（docker + Go 后端 + React 前端），并实时打印日志
./stop.sh     # 停止所有服务（数据卷保留）
```

- `Ctrl+C` 只退出日志窗口，**服务继续在后台运行**
- 想真的停服务，运行 `./stop.sh`
- 服务地址：
  - 前端 `http://localhost:8890`（浏览器入口）
  - 后端 `http://localhost:8891/v1/health`
- 其它端口：PostgreSQL `8892`，Redis `8893`，Xray API `8894`，Xray 入站 `8895`
- 日志文件：`logs/docker.log`、`logs/server.log`、`logs/web.log`
- 首次启动会自动 `cp .env.example .env` 并 `npm install`

## 生产部署（镜像版）

整套从镜像仓库运行，服务器不 checkout 代码。push 到 main/master 或打 tag →
GitHub Actions 自动构建推 `registry.cn-hongkong.aliyuncs.com/mewtwo_zero/routex:latest`。
服务器部署步骤见 [deploy/registry/README.md](deploy/registry/README.md)。

## 阶段规划

- **一期**：用户管理 + Xray 集成 + 流量统计 + 节点管理
- **二期**：出口代理池 + 监控告警
- **三期**：订阅系统 + 支付 + 多节点集群

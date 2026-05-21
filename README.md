# RouteX

基于 Xray-core + Go + React 的多用户代理管理平台。

## 目录结构

```
RouteX/
├── server/             # Go 后端（Gin + GORM + Zap）
├── web/                # React 前端（Vite + TS + AntD）
├── deploy/
│   ├── xray/           # Xray-core 配置
│   └── sing-box/       # sing-box 配置（可选出口路由）
├── migrations/         # 数据库迁移 SQL
├── docker-compose.yml  # 一键起本地全套依赖
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
- 其它端口：PostgreSQL `8892`，Redis `8893`，Xray API `8894`，Xray 入站 `8895`，sing-box API `8896`
- 日志文件：`logs/docker.log`、`logs/server.log`、`logs/web.log`
- 首次启动会自动 `cp .env.example .env` 并 `npm install`

## Xray / sing-box 安装方式

本项目默认通过 **Docker Compose** 拉取官方镜像运行（见 `docker-compose.yml`），无需手动安装。

如果需要在裸机上独立部署：

| 工具 | macOS | Linux |
|---|---|---|
| Xray-core | `brew install xray` | `bash -c "$(curl -L https://github.com/XTLS/Xray-install/raw/main/install-release.sh)" @ install` |
| sing-box | `brew install sing-box` | `bash -c "$(curl -fsSL https://sing-box.app/install.sh)"` |

## 阶段规划

- **一期**：用户管理 + Xray 集成 + 流量统计 + 节点管理
- **二期**：出口代理池 + 监控告警
- **三期**：订阅系统 + 支付 + 多节点集群

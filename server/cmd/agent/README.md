# RouteX Node Agent

每个代理节点上跑一个 agent 进程，周期性把 CPU / 内存 / 网络流量上报到控制平面，
让后台能监控节点在线状态和负载。

## 构建

```bash
cd server
go build -o bin/agent ./cmd/agent
```

或者用 Docker：

```bash
docker build -f cmd/agent/Dockerfile -t routex-agent:latest .
```

## 配置（环境变量）

| 变量 | 必填 | 说明 |
|---|---|---|
| `ROUTEX_API_URL` | 是 | 控制平面地址，如 `http://control.example.com:8080` |
| `ROUTEX_NODE_TOKEN` | 是 | 控制台 `POST /v1/nodes` 创建节点时一次性返回的 token |
| `ROUTEX_HEARTBEAT_INTERVAL` | 否 | 心跳间隔，默认 `30s` |

## 运行

```bash
ROUTEX_API_URL=http://1.2.3.4:8080 \
ROUTEX_NODE_TOKEN=xxxxx \
./bin/agent
```

或 systemd 服务：

```ini
[Unit]
Description=RouteX Node Agent
After=network.target

[Service]
Environment=ROUTEX_API_URL=http://1.2.3.4:8080
Environment=ROUTEX_NODE_TOKEN=xxxxx
ExecStart=/usr/local/bin/agent
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## 它做什么 / 不做什么

做：
- 启动时立即发一次心跳（节点在控制台立刻显示"在线"）
- 每个周期采集系统 CPU%、内存%、本周期净流量字节数
- POST `/v1/nodes/heartbeat`，请求头 `X-Node-Token` 鉴权

不做（后续可扩展）：
- 节点 Xray 进程拉起/重启（需要节点 agent 接管 Xray 生命周期）
- 节点本地配置同步（控制端 push 新用户/路由）
- 节点日志回传

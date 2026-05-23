# 节点部署（单独 VPS）

把 Xray + agent 部署到一台独立 VPS 上，作为代理节点。控制台部署见 `deploy/README-prod.md`。

## 角色分工提醒

| | 控制台 | 节点 |
|---|---|---|
| 跑什么 | 后端 + 前端 + DB + Redis | 只跑 Xray + agent |
| 部署用 | `./install.sh` (or `docker-compose.prod.yml`) | **`./install-node.sh`** (or `docker-compose.node.yml`) |
| Reality privateKey | 控制台生成 | **跟控制台用同一个**（手动复制过去） |

## 部署步骤

### 1. 在【控制台】上准备

a. 拿 Reality privateKey：

```bash
grep privateKey deploy/xray/config.json
# 复制双引号里那串
```

b. 浏览器打开控制台 → 节点 → 新增节点：
- 名称：`hk-1`（任意）
- 公网地址：这台**节点 VPS** 的公网域名或 IP
- 公网端口：`8895`（默认）或 `443`
- 提交后弹窗显示 **agent token**，**复制下来**（关了再也找不回）

### 2. 在【节点 VPS】上跑

```bash
# a. 拉代码（只为了拿 install-node.sh + agent 源码）
git clone <你的repo> /srv/routex-node
cd /srv/routex-node

# b. 一键安装
./install-node.sh
```

脚本会问你 3 件事：

| 提示 | 回答 |
|---|---|
| 控制台 URL | 控制台对外可达的地址，如 `http://1.2.3.4:8891` |
| 节点 token | 上面步骤 1b 复制的那串 |
| Reality privateKey | 上面步骤 1a 复制的那串 |

回车默认：VLESS 端口 `8895`，心跳间隔 `30s`。

### 3. 验证

几秒后，回控制台 → 节点页面，应该看到这台节点 🟢 在线 + 显示真实 CPU/内存。

## 防火墙

节点机器只需要放行：
- TCP `8895`（或你改的 VLESS 端口）— 给客户端连
- 出站全开 — agent 要 POST 心跳到控制台

**节点的 xray gRPC（容器内 10085）故意不暴露主机端口**，避免无认证 RCE。

## 重新跑 install-node.sh 也安全

`.env.node` 和 `deploy/xray/config.json` 已存在时跳过交互。所以脚本是幂等的，重启脚本不会重新问问题。

## ⚠️ 当前限制

目前控制台**不会**把"创建用户 / 切出口"push 到远程节点。节点能心跳上报，但客户端 VLESS 连过去会被拒（节点 Xray 不知道用户 UUID）。

这是**多节点同步**的工作，下一步会做：让控制台通过 agent 把用户列表 / 出口配置同步到所有节点，agent 在本地调 Xray gRPC 应用。

## 常用命令（节点机器）

```bash
docker compose -f docker-compose.node.yml ps
docker compose -f docker-compose.node.yml logs -f agent      # 心跳上报
docker compose -f docker-compose.node.yml logs -f xray       # Xray 日志
docker compose -f docker-compose.node.yml restart            # 重启全部
docker compose -f docker-compose.node.yml down               # 停（保留数据）
```

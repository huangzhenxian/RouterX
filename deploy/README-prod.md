# RouteX 生产部署（单 VPS 全栈）

把整套（后台 + Xray + DB）跑在一台 VPS 上。出口分散由你后台里加的 SOCKS5 代理池负责，不靠多入口节点。

## 一、配置 VPS

最低配置：1 核 2G，20G 硬盘，带宽建议 ≥ 100Mbps。  
推荐：2 核 4G + 1Gbps（100~500 用户）。

必须装：

```bash
# Debian/Ubuntu 示例
curl -fsSL https://get.docker.com | sh
sudo systemctl enable --now docker
sudo usermod -aG docker $USER   # 重新登录生效
```

防火墙放行（云服务商安全组里也放）：
- TCP **80**（管理面板，后面建议套 HTTPS）
- TCP **443**（VLESS+Reality 入站，或者你改成 8895）

## 二、上传代码到 VPS

```bash
# 在本地
rsync -avz --exclude='.git' --exclude='node_modules' --exclude='logs' \
  --exclude='.pids' --exclude='bin' --exclude='.env' \
  /path/to/RouteX/ user@your-vps:/srv/routex/
```

或者直接 `git clone`，如果你把仓库放上 GitHub 的话。

## 三、首次部署

### 1. 生成 Reality 密钥

在 VPS 上：

```bash
cd /srv/routex
docker run --rm teddysun/xray xray x25519
```

记下输出的 `PrivateKey` 和 `PublicKey`，下面要用。

### 2. 准备 Xray 配置

```bash
cp deploy/xray/config.example.json deploy/xray/config.json
# 把 config.json 里的 REPLACE_WITH_xray_x25519_PRIVATE_KEY 改成上面的 PrivateKey
vim deploy/xray/config.json
```

### 3. 准备 .env

```bash
cp .env.prod.example .env
vim .env
```

至少要改这几项：

| key | 改成什么 |
|---|---|
| `APP_JWT_SECRET` | `openssl rand -hex 32` 的输出 |
| `POSTGRES_PASSWORD` | 随便一个 16 位以上强密码 |
| `PUBLIC_HOST` | 你 VPS 的公网域名或 IP（客户端连这里） |
| `PUBLIC_PORT` | 生产建议 `443`（同时把 docker-compose.prod.yml 的 8895:443 改成 443:443，或者由 `PUBLIC_PORT` env 自动驱动） |
| `REALITY_PUBLIC_KEY` | 上面 x25519 输出的 PublicKey |
| `WEB_ORIGIN` | 你管理面板的 URL，例如 `http://1.2.3.4` 或 `https://admin.yourdomain.com` |

### 4. 起服务

```bash
docker compose -f docker-compose.prod.yml --env-file .env up -d --build
```

第一次会编 Go + 编前端，约 1-3 分钟。完成后看状态：

```bash
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs -f --tail=200
```

### 5. 拿默认管理员密码

```bash
docker compose -f docker-compose.prod.yml logs server | grep BOOTSTRAP_ADMIN
```

抄下 `password=` 后面那串。

### 6. 登录

浏览器打开 `http://<你VPS的IP>` 或 `http://<你的域名>`，用 `admin` + 上面那个密码登录。

## 四、日常运维

```bash
# 看日志
docker compose -f docker-compose.prod.yml logs -f --tail=200 server
docker compose -f docker-compose.prod.yml logs -f --tail=200 xray

# 改代码后重新部署
git pull && docker compose -f docker-compose.prod.yml up -d --build

# 备份数据库
docker compose -f docker-compose.prod.yml exec postgres \
  pg_dump -U routex routex | gzip > backup-$(date +%F).sql.gz

# 停服（保留数据）
docker compose -f docker-compose.prod.yml down

# 全清（含 DB 数据卷！慎用）
docker compose -f docker-compose.prod.yml down -v
```

## 五、安全清单

- [ ] `.env` 文件 `chmod 600`，绝不上传到 git
- [ ] `APP_JWT_SECRET` 是真随机的，不是默认 placeholder
- [ ] `POSTGRES_PASSWORD` 同样
- [ ] **xray gRPC API（10085）只在 docker 内部网络可达**，prod compose 没暴露端口，确认 `docker compose ps` 没有 10085 那一行
- [ ] 管理面板 80 端口建议套 Caddy/Nginx 反代加 Let's Encrypt 自动 HTTPS
- [ ] VPS 防火墙只开 80 / 443 / 22，其他全关

## 六、加 SOCKS5 出口代理

登录后台 → 出口代理 → 新增。健康检查每 2 分钟一轮，RouteX 会自动把 vless 流量切到最健康的那条。详细见主 README。

## 七、HTTPS 建议（Caddy 一行版）

在 VPS 上跑 Caddy 反代到 docker 的 web:80：

```caddyfile
admin.yourdomain.com {
    reverse_proxy localhost:80
}
```

把 docker-compose.prod.yml 里 web 的端口映射改成 `"127.0.0.1:80:80"`（只监听本地），让 Caddy 来对外提供 HTTPS。

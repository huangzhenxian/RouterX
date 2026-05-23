# RouteX 生产部署（单 VPS 全栈）

把整套（后台 + Xray + DB）跑在一台 VPS 上。出口分散由你后台里加的 SOCKS5 代理池负责，不靠多入口节点。

最低配置：1 核 2G，20G 硬盘，带宽 ≥ 100Mbps。  
推荐：2 核 4G + 1Gbps（100~500 用户）。

## 🚀 一键部署（推荐）

```bash
# 1. 把代码搬上 VPS（任选其一）
git clone <你的repo> /srv/routex
# 或：rsync -avz ./ user@vps:/srv/routex/

# 2. 跑安装脚本
cd /srv/routex
./install.sh
```

脚本会自动：
- 装 Docker（没装的话）
- 跑 `xray x25519` 生成 Reality 密钥
- 生成强随机 JWT_SECRET / POSTGRES_PASSWORD
- 问你**一个问题**：公网地址（域名或 IP）
- `docker compose up -d --build`
- 打印默认管理员密码

完事。打开 `http://<你的VPS>` 用 admin + 终端打印的密码登录。

防火墙记得放行：
- TCP `80`（管理面板）
- TCP `8895`（VLESS 入站，可改 443）

## 重新跑 install.sh 也安全

已存在的 `.env` 和 `deploy/xray/config.json` 不会被覆盖，相当于幂等。

## 手动部署（不想用 install.sh 的话）

<details>
<summary>展开</summary>

### 1. 生成 Reality 密钥

```bash
docker run --rm teddysun/xray xray x25519
```

记下 PrivateKey 和 PublicKey。

### 2. Xray 配置

```bash
cp deploy/xray/config.example.json deploy/xray/config.json
vim deploy/xray/config.json   # 填 PrivateKey
```

### 3. .env

```bash
cp .env.prod.example .env
vim .env
```

至少要改：

| key | 改成什么 |
|---|---|
| `APP_JWT_SECRET` | `openssl rand -hex 32` |
| `POSTGRES_PASSWORD` | 强随机 |
| `PUBLIC_HOST` | 你的域名或 IP |
| `REALITY_PUBLIC_KEY` | 上面那个 PublicKey |
| `WEB_ORIGIN` | `http://你的域名` |

### 4. 起服务

```bash
docker compose -f docker-compose.prod.yml up -d --build
```

### 5. 拿默认管理员密码

```bash
docker compose -f docker-compose.prod.yml logs server | grep BOOTSTRAP_ADMIN
```

</details>

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

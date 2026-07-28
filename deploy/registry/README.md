# RouteX 服务器部署(镜像版)

整套系统(postgres + redis + xray + 应用)全部从镜像运行,服务器**不需要 checkout 代码**。
应用镜像是**单镜像全栈**:一个容器里跑 nginx(前端+反代) + Go 后端,由 supervisord 编排;
postgres/redis/xray 用官方镜像随本栈一起跑。

镜像:`registry.cn-hongkong.aliyuncs.com/mewtwo_zero/routex:latest`(本仓库只推 latest 通道)

## 服务器上要哪些文件

把本目录(`deploy/registry/`)里的 3 个文件传上去,外加一份填好私钥的 Xray 配置:

```
~/routex/
├── docker-compose.yml     # 本目录的
├── restart.sh             # 本目录的(chmod +x)
├── .env                   # 由 .env.example 复制后逐项改
└── xray/
    └── config.json        # 由仓库 deploy/xray/config.example.json 复制,填 Reality privateKey
```

## 首次部署

```bash
# 0. 装好 Docker + Compose,登录 ACR(私有仓库要登录才能拉)
docker login registry.cn-hongkong.aliyuncs.com   # 用户名/密码同 GitHub Secrets

# 1. 建目录、传文件
mkdir -p ~/routex/xray && cd ~/routex
# 传:docker-compose.yml / restart.sh / .env(改成 .env)
# 传:xray/config.json
chmod +x restart.sh

# 2. 生成 Reality 密钥,把 privateKey 填进 xray/config.json,publicKey 填进 .env
docker run --rm teddysun/xray xray x25519

# 3. 起栈(restart.sh 会先预检 .env 必填项,占位符没改会拒绝启动)
./restart.sh

# 4. 拿默认管理员密码(只打一次,务必抄下来)
docker compose logs app | grep BOOTSTRAP_ADMIN
```

打开管理面板:默认只绑 `127.0.0.1:8890`,对外由 nginx/caddy 反代 80/443 进来;
本机验证 `curl -s http://localhost:8890/v1/health`。

## 日常更新

代码 push 到 main/master(或打 tag)→ GitHub Actions 自动构建推 ACR → 服务器上:

```bash
cd ~/routex && ./restart.sh
```

`restart.sh` = 启动/更新二合一:`docker compose pull` → `up -d --remove-orphans`
→ `image prune`,结尾打印各服务实际运行的镜像版本。postgres/redis 没动就不会重启。

## 防火墙 / 端口

- 管理面板 `APP_PORT`(默认 8890)**只绑 127.0.0.1**。Docker 发布端口会绕过宿主防火墙,
  直接 `0.0.0.0` 就是公网裸奔 —— 对外一定要走反代。
- VLESS 入站 `PUBLIC_PORT`(默认 8895,可改 443)**必须公网可达**,这是客户端连的口子。
- Xray gRPC `10085`、postgres `5432`、redis `6379` 都不对外,只在 docker 内网。
- 云安全组建议只开 80/443/22 + PUBLIC_PORT,其余全关。

## HTTPS(Caddy 一行版)

反代到本机的面板端口:

```caddyfile
admin.yourdomain.com {
    reverse_proxy localhost:8890
}
```

## 常用命令

```bash
docker compose logs -f app       # 应用(后端+nginx)日志
docker compose logs -f xray      # xray 日志
docker compose ps                # 状态
docker compose down              # 停(保留数据卷)
docker compose down -v           # 全清(含 DB 数据,慎用)

# 备份数据库
docker compose exec postgres pg_dump -U routex routex | gzip > backup-$(date +%F).sql.gz
```

## GitHub Actions 需要的 Secrets

| Secret | 说明 |
|---|---|
| `ACR_USERNAME` | 阿里云 ACR 用户名 |
| `ACR_PASSWORD` | 阿里云 ACR 密码(建议用独立访问凭证) |

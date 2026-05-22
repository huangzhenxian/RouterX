# Xray 配置说明

## 文件分工

| 文件 | 提交? | 作用 |
|---|---|---|
| `config.example.json` | ✅ 是 | 模板。`privateKey` 是占位符 `REPLACE_WITH_xray_x25519_PRIVATE_KEY` |
| `config.json` | ❌ **gitignored** | 本地实际生效的配置，含真实 Reality 私钥 |

## 首次启动会自动做的事

`./start.sh` 首次启动时若 `deploy/xray/config.json` 不存在，会自动：

1. `cp config.example.json config.json`
2. 跑一次 `docker run --rm teddysun/xray xray x25519` 生成 Reality 密钥对
3. 把 PrivateKey 替换进 `config.json`，并在终端打印 PublicKey（给客户端用）

## 手动生成 / 重新生成

```bash
docker run --rm teddysun/xray xray x25519
```

输出：

```
PrivateKey: <把这个填进 config.json 的 realitySettings.privateKey>
Password (PublicKey): <这个给客户端订阅链接用>
```

## API 端口

容器内 `10085` 跑 Xray gRPC API，由 docker-compose 映射到宿主机 `8894`。Go 后端通过它动态增删用户、拉取流量统计。**生产环境一定要用防火墙锁住，只允许后端访问。**

## 用户管理

`vless-in.settings.clients` 里的占位用户只是为了让 Xray 启动不报错。真实用户由 Go 后端通过 `HandlerService.AlterInbound` 动态添加，**不会**写回这份配置文件。Xray 容器重启后内存中的用户会丢失，由 `scheduler.XrayWatcher` 检测到恢复后自动 `SyncAll`。

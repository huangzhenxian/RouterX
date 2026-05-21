# Xray 配置说明

## 生成 Reality 密钥对

```bash
docker run --rm teddysun/xray xray x25519
```

输出形如：

```
Private key: ...
Public key:  ...
```

把 `Private key` 填到 `config.json` 的 `realitySettings.privateKey`，前端订阅链接用 `Public key`。

## API 端口

`10085` 端口跑的是 Xray gRPC API，Go 后端通过它动态增删用户、拉取流量统计。**生产环境一定要用防火墙锁住，只允许后端访问。**

## 用户管理

`vless-in.settings.clients` 里的占位用户只是为了让 Xray 启动不报错。真实用户由 Go 后端通过 `HandlerService.AlterInbound` 动态添加，**不会**写回这份配置文件。

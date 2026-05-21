package xray

// 占位：后续封装 Xray gRPC API（HandlerService / StatsService）。
//
// 关键能力：
//   - AddUser(inboundTag, user) error            动态添加用户
//   - RemoveUser(inboundTag, email) error        动态删除用户
//   - QueryUserTraffic(email) (up, down int64)   查询用户流量
//   - ResetUserTraffic(email) error              重置流量计数
//
// 依赖：
//   google.golang.org/grpc
//   github.com/xtls/xray-core/app/proxyman/command
//   github.com/xtls/xray-core/app/stats/command
//   github.com/xtls/xray-core/proxy/vless
//
// 一期接通后再补完整实现。

type Client struct {
	addr      string
	inboundTag string
}

func NewClient(addr, inboundTag string) *Client {
	return &Client{addr: addr, inboundTag: inboundTag}
}

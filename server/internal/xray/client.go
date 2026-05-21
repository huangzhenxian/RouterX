package xray

import (
	"context"
	"fmt"

	handlerCmd "github.com/xtls/xray-core/app/proxyman/command"
	statsCmd "github.com/xtls/xray-core/app/stats/command"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client 封装 Xray gRPC API。零值不可用，必须 New。
//
// 生命周期：进程启动时建一个，进程退出时 Close。
// 注意 grpc.NewClient 是惰性连接的，构造时不会立即报错；用 Ping 验真。
type Client struct {
	addr       string
	inboundTag string
	conn       *grpc.ClientConn
	handler    handlerCmd.HandlerServiceClient
	stats      statsCmd.StatsServiceClient
}

func New(addr, inboundTag string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("xray grpc.NewClient %s: %w", addr, err)
	}
	return &Client{
		addr:       addr,
		inboundTag: inboundTag,
		conn:       conn,
		handler:    handlerCmd.NewHandlerServiceClient(conn),
		stats:      statsCmd.NewStatsServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// InboundTag 返回客户端默认操作的入站 tag。
func (c *Client) InboundTag() string { return c.inboundTag }

// Ping 通过 GetSysStats 验证 gRPC 通道和 StatsService 是否可用。
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.stats.GetSysStats(ctx, &statsCmd.SysStatsRequest{})
	if err != nil {
		return fmt.Errorf("xray ping %s: %w", c.addr, err)
	}
	return nil
}

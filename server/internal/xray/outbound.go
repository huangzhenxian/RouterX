package xray

import (
	"context"
	"fmt"

	handlerCmd "github.com/xtls/xray-core/app/proxyman/command"
	xnet "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/proxy/freedom"
	xhttp "github.com/xtls/xray-core/proxy/http"
	"github.com/xtls/xray-core/proxy/socks"
)

// AddSocksOutbound 注册一个 SOCKS5 出站 handler。
// tag 已存在时会冲突，调用方需要先 RemoveOutbound。
func (c *Client) AddSocksOutbound(ctx context.Context, tag, host string, port uint32, user, pass string) error {
	settings := &socks.ClientConfig{
		Server: &protocol.ServerEndpoint{
			Address: xnet.NewIPOrDomain(xnet.ParseAddress(host)),
			Port:    port,
			User:    socksUser(user, pass),
		},
	}
	return c.addOutbound(ctx, tag, serial.ToTypedMessage(settings))
}

// AddHTTPOutbound 注册一个 HTTP(S) CONNECT 出站 handler。
func (c *Client) AddHTTPOutbound(ctx context.Context, tag, host string, port uint32, user, pass string) error {
	settings := &xhttp.ClientConfig{
		Server: &protocol.ServerEndpoint{
			Address: xnet.NewIPOrDomain(xnet.ParseAddress(host)),
			Port:    port,
			User:    httpUser(user, pass),
		},
	}
	return c.addOutbound(ctx, tag, serial.ToTypedMessage(settings))
}

// AddFreedomOutbound 注册一个 freedom（直连）出站，用作 proxy-out 的兜底。
func (c *Client) AddFreedomOutbound(ctx context.Context, tag string) error {
	return c.addOutbound(ctx, tag, serial.ToTypedMessage(&freedom.Config{}))
}

// RemoveOutbound 按 tag 删除出站 handler。tag 不存在时返回错误（调用方决定是否容忍）。
func (c *Client) RemoveOutbound(ctx context.Context, tag string) error {
	_, err := c.handler.RemoveOutbound(ctx, &handlerCmd.RemoveOutboundRequest{Tag: tag})
	if err != nil {
		return fmt.Errorf("xray RemoveOutbound %s: %w", tag, err)
	}
	return nil
}

func (c *Client) addOutbound(ctx context.Context, tag string, settings *serial.TypedMessage) error {
	_, err := c.handler.AddOutbound(ctx, &handlerCmd.AddOutboundRequest{
		Outbound: &core.OutboundHandlerConfig{
			Tag:           tag,
			ProxySettings: settings,
		},
	})
	if err != nil {
		return fmt.Errorf("xray AddOutbound %s: %w", tag, err)
	}
	return nil
}

func socksUser(user, pass string) *protocol.User {
	if user == "" {
		return nil
	}
	return &protocol.User{
		Account: serial.ToTypedMessage(&socks.Account{Username: user, Password: pass}),
	}
}

func httpUser(user, pass string) *protocol.User {
	if user == "" {
		return nil
	}
	return &protocol.User{
		Account: serial.ToTypedMessage(&xhttp.Account{Username: user, Password: pass}),
	}
}

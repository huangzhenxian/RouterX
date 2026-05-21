package xray

import (
	"context"
	"fmt"

	handlerCmd "github.com/xtls/xray-core/app/proxyman/command"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/proxy/vless"
)

// AddUser 在指定入站里动态添加一个用户。
// 重复 email 会返回错误（Xray 不允许同 email 重复），调用方决定是先删后加还是忽略。
func (c *Client) AddUser(ctx context.Context, u UserSpec) error {
	account, err := accountForProtocol(u)
	if err != nil {
		return err
	}
	_, err = c.handler.AlterInbound(ctx, &handlerCmd.AlterInboundRequest{
		Tag: c.inboundTag,
		Operation: serial.ToTypedMessage(&handlerCmd.AddUserOperation{
			User: &protocol.User{
				Level:   u.Level,
				Email:   u.Email,
				Account: account,
			},
		}),
	})
	if err != nil {
		return fmt.Errorf("xray AddUser %s: %w", u.Email, err)
	}
	return nil
}

// RemoveUser 按 email 从入站里删除用户。
// email 不存在时 Xray 返回错误，调用方需要决定是否容忍 not-found。
func (c *Client) RemoveUser(ctx context.Context, email string) error {
	_, err := c.handler.AlterInbound(ctx, &handlerCmd.AlterInboundRequest{
		Tag: c.inboundTag,
		Operation: serial.ToTypedMessage(&handlerCmd.RemoveUserOperation{
			Email: email,
		}),
	})
	if err != nil {
		return fmt.Errorf("xray RemoveUser %s: %w", email, err)
	}
	return nil
}

func accountForProtocol(u UserSpec) (*serial.TypedMessage, error) {
	proto := u.Protocol
	if proto == "" {
		proto = ProtocolVLESS
	}
	switch proto {
	case ProtocolVLESS:
		if u.UUID == "" {
			return nil, fmt.Errorf("xray VLESS user requires UUID")
		}
		return serial.ToTypedMessage(&vless.Account{
			Id:   u.UUID,
			Flow: u.Flow,
		}), nil
	// Trojan / Shadowsocks 待补：分别引 proxy/trojan 和 proxy/shadowsocks 的 Account。
	default:
		return nil, fmt.Errorf("xray protocol %q not implemented", proto)
	}
}

package xray

import (
	"context"
	"fmt"
	"strings"

	statsCmd "github.com/xtls/xray-core/app/stats/command"
)

// QueryUserTraffic 查询单个用户的累计上下行流量。
// reset=true 时 Xray 会在返回当前值后把计数清零（适合定时落库的场景）。
func (c *Client) QueryUserTraffic(ctx context.Context, email string, reset bool) (TrafficStat, error) {
	return c.queryPair(ctx,
		fmt.Sprintf("user>>>%s>>>traffic>>>uplink", email),
		fmt.Sprintf("user>>>%s>>>traffic>>>downlink", email),
		reset,
	)
}

// QueryInboundTraffic 查询整个入站的累计上下行流量。
func (c *Client) QueryInboundTraffic(ctx context.Context, tag string, reset bool) (TrafficStat, error) {
	return c.queryPair(ctx,
		fmt.Sprintf("inbound>>>%s>>>traffic>>>uplink", tag),
		fmt.Sprintf("inbound>>>%s>>>traffic>>>downlink", tag),
		reset,
	)
}

func (c *Client) queryPair(ctx context.Context, upName, downName string, reset bool) (TrafficStat, error) {
	up, err := c.queryOne(ctx, upName, reset)
	if err != nil {
		return TrafficStat{}, err
	}
	down, err := c.queryOne(ctx, downName, reset)
	if err != nil {
		return TrafficStat{}, err
	}
	return TrafficStat{Uplink: up, Downlink: down}, nil
}

func (c *Client) queryOne(ctx context.Context, name string, reset bool) (int64, error) {
	resp, err := c.stats.GetStats(ctx, &statsCmd.GetStatsRequest{
		Name:   name,
		Reset_: reset,
	})
	if err != nil {
		// 用户/inbound 从未产生流量时 Xray 返回 not found 的文本错误，视为 0。
		if isNotFound(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("xray GetStats %s: %w", name, err)
	}
	if resp.GetStat() == nil {
		return 0, nil
	}
	return resp.Stat.Value, nil
}

func isNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

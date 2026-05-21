// Command agent runs on each proxy node and posts heartbeats to the
// RouteX control plane. Build:
//
//	go build -o bin/agent ./cmd/agent
//
// Usage (env-driven, no flags):
//
//	ROUTEX_API_URL=http://control.example.com:8080 \
//	ROUTEX_NODE_TOKEN=<token-from-/v1/nodes-create> \
//	ROUTEX_HEARTBEAT_INTERVAL=30s   # optional, default 30s
//	./bin/agent
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"
)

const agentVersion = "0.1.0"

type heartbeatReq struct {
	CPU       float64 `json:"cpu"`
	Memory    float64 `json:"memory"`
	Bandwidth int64   `json:"bandwidth"` // 当前周期下总流量字节差值（双向）
	Version   string  `json:"version"`
}

func main() {
	api := mustEnv("ROUTEX_API_URL")
	token := mustEnv("ROUTEX_NODE_TOKEN")
	interval := envDuration("ROUTEX_HEARTBEAT_INTERVAL", 30*time.Second)

	log.Printf("agent v%s starting, target=%s interval=%s", agentVersion, api, interval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-stop; cancel() }()

	client := &http.Client{Timeout: 10 * time.Second}
	var lastNetBytes uint64
	var lastTickAt time.Time

	tick := func() {
		hb, newCounter, err := collect(ctx, lastNetBytes, lastTickAt)
		if err != nil {
			log.Printf("collect: %v", err)
			return
		}
		lastNetBytes = newCounter
		lastTickAt = time.Now()
		if err := send(ctx, client, api, token, hb); err != nil {
			log.Printf("send: %v", err)
		}
	}

	// 启动时先发一次，让控制端立刻看到节点上线
	tick()

	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Printf("agent exiting")
			return
		case <-t.C:
			tick()
		}
	}
}

// collect 读 CPU/内存/网卡累计计数，把 bandwidth 折算成上次到本次的差值字节数。
func collect(ctx context.Context, prevBytes uint64, prevAt time.Time) (heartbeatReq, uint64, error) {
	cpuPcts, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err != nil {
		return heartbeatReq{}, 0, fmt.Errorf("cpu: %w", err)
	}
	var cpuPct float64
	if len(cpuPcts) > 0 {
		cpuPct = cpuPcts[0]
	}

	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return heartbeatReq{}, 0, fmt.Errorf("mem: %w", err)
	}

	io, err := psnet.IOCountersWithContext(ctx, false) // 总和
	if err != nil {
		return heartbeatReq{}, 0, fmt.Errorf("net: %w", err)
	}
	var total uint64
	if len(io) > 0 {
		total = io[0].BytesSent + io[0].BytesRecv
	}

	var delta int64
	if !prevAt.IsZero() && total >= prevBytes {
		delta = int64(total - prevBytes)
	}

	return heartbeatReq{
		CPU:       cpuPct,
		Memory:    vm.UsedPercent,
		Bandwidth: delta,
		Version:   agentVersion,
	}, total, nil
}

func send(ctx context.Context, c *http.Client, api, token string, hb heartbeatReq) error {
	body, _ := json.Marshal(hb)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, api+"/v1/nodes/heartbeat", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Node-Token", token)

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d body=%s", resp.StatusCode, string(b))
	}
	return nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env %s not set", key)
	}
	return v
}

func envDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Printf("invalid %s=%q, using default %s", key, v, def)
		return def
	}
	return d
}

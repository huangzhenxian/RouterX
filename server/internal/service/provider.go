package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/routex/routex/internal/model"
	"golang.org/x/net/proxy"
	"gorm.io/gorm"
)

type ProviderService struct {
	db *gorm.DB
}

func NewProviderService(db *gorm.DB) *ProviderService { return &ProviderService{db: db} }

type CreateProviderInput struct {
	Name     string `json:"name"     binding:"required"`
	Type     string `json:"type"     binding:"required,oneof=socks5 http https"`
	Host     string `json:"host"     binding:"required"`
	Port     int    `json:"port"     binding:"required,min=1,max=65535"`
	Username string `json:"username"`
	Password string `json:"password"`
	Region   string `json:"region"`
	Tags     string `json:"tags"`
	Priority int    `json:"priority"`
}

func (s *ProviderService) Create(ctx context.Context, in CreateProviderInput) (*model.ProxyProvider, error) {
	p := &model.ProxyProvider{
		Name:     in.Name,
		Type:     in.Type,
		Host:     in.Host,
		Port:     in.Port,
		Username: in.Username,
		Password: in.Password,
		Region:   in.Region,
		Tags:     in.Tags,
		Priority: in.Priority,
		Status:   1,
	}
	if err := s.db.WithContext(ctx).Create(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProviderService) List(ctx context.Context) ([]model.ProxyProvider, error) {
	var ps []model.ProxyProvider
	return ps, s.db.WithContext(ctx).Order("priority asc, id asc").Find(&ps).Error
}

func (s *ProviderService) Get(ctx context.Context, id int64) (*model.ProxyProvider, error) {
	var p model.ProxyProvider
	if err := s.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *ProviderService) Delete(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Delete(&model.ProxyProvider{}, id).Error
}

func (s *ProviderService) Enable(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Model(&model.ProxyProvider{}).Where("id = ?", id).
		Updates(map[string]any{"status": 1, "fail_count": 0}).Error
}

func (s *ProviderService) Disable(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Model(&model.ProxyProvider{}).Where("id = ?", id).
		Update("status", 0).Error
}

// CheckResult 一次健康检查的结论。
type CheckResult struct {
	Healthy   bool
	LatencyMs int
	Err       string
}

// Check 用一次 HEAD 请求测试代理是否能通到外网。
// testURL 走 https://www.cloudflare.com/cdn-cgi/trace —— 这条 URL 几乎全球都能访问、返回快。
func (s *ProviderService) Check(ctx context.Context, p *model.ProxyProvider) CheckResult {
	const testURL = "https://www.cloudflare.com/cdn-cgi/trace"
	const timeout = 8 * time.Second

	client, err := buildProxyClient(p, timeout)
	if err != nil {
		return CheckResult{Err: "build client: " + err.Error()}
	}

	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(cctx, http.MethodGet, testURL, nil)
	if err != nil {
		return CheckResult{Err: err.Error()}
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return CheckResult{Err: err.Error()}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	latency := int(time.Since(start).Milliseconds())
	if resp.StatusCode/100 != 2 {
		return CheckResult{LatencyMs: latency, Err: fmt.Sprintf("status %d", resp.StatusCode)}
	}
	return CheckResult{Healthy: true, LatencyMs: latency}
}

// PersistCheck 把一次检查结果写回 DB，并维护 FailCount / 自动禁用。
func (s *ProviderService) PersistCheck(ctx context.Context, p *model.ProxyProvider, r CheckResult, autoDisableThreshold int) error {
	now := time.Now()
	updates := map[string]any{
		"healthy":         r.Healthy,
		"latency_millis":  r.LatencyMs,
		"last_checked_at": &now,
		"last_error":      truncate(r.Err, 255),
	}
	if r.Healthy {
		updates["fail_count"] = 0
	} else {
		updates["fail_count"] = p.FailCount + 1
		if autoDisableThreshold > 0 && p.FailCount+1 >= autoDisableThreshold {
			updates["status"] = 0
		}
	}
	return s.db.WithContext(ctx).Model(p).Updates(updates).Error
}

// EnabledForCheck 调度器用：拿所有启用的 provider 做一轮检查。
func (s *ProviderService) EnabledForCheck(ctx context.Context) ([]model.ProxyProvider, error) {
	var ps []model.ProxyProvider
	return ps, s.db.WithContext(ctx).Where("status = ?", 1).Find(&ps).Error
}

func buildProxyClient(p *model.ProxyProvider, timeout time.Duration) (*http.Client, error) {
	transport := &http.Transport{
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
	}
	switch strings.ToLower(p.Type) {
	case "socks5":
		var auth *proxy.Auth
		if p.Username != "" {
			auth = &proxy.Auth{User: p.Username, Password: p.Password}
		}
		dialer, err := proxy.SOCKS5("tcp", net.JoinHostPort(p.Host, fmt.Sprintf("%d", p.Port)), auth, &net.Dialer{Timeout: timeout})
		if err != nil {
			return nil, err
		}
		transport.DialContext = func(_ context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
	case "http", "https":
		scheme := "http"
		if p.Type == "https" {
			scheme = "https"
		}
		u := &url.URL{
			Scheme: scheme,
			Host:   net.JoinHostPort(p.Host, fmt.Sprintf("%d", p.Port)),
		}
		if p.Username != "" {
			u.User = url.UserPassword(p.Username, p.Password)
		}
		transport.Proxy = http.ProxyURL(u)
	default:
		return nil, errors.New("unsupported proxy type: " + p.Type)
	}
	return &http.Client{Transport: transport, Timeout: timeout}, nil
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

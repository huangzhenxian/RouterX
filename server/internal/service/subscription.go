package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/routex/routex/internal/config"
	"github.com/routex/routex/internal/model"
	"gorm.io/gorm"
)

type SubscriptionService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewSubscriptionService(db *gorm.DB, cfg *config.Config) *SubscriptionService {
	return &SubscriptionService{db: db, cfg: cfg}
}

var ErrSubscriptionNotFound = errors.New("subscription token not found")

// PerNodeLink 一条 vless:// 链接 + 它对应的展示信息。
type PerNodeLink struct {
	NodeName string `json:"node_name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	URL      string `json:"url"`
}

type Payload struct {
	Username     string        `json:"username"`
	UUID         string        `json:"uuid"`
	Links        []PerNodeLink `json:"links"`
	Base64       string        `json:"base64"`        // 客户端订阅常用格式
	RawText      string        `json:"raw_text"`      // base64 解码后的明文
}

// BuildForUser 给指定用户生成所有节点上的连接链接。
//
// 节点策略：
//   - 若 nodes 表有 status=1 的记录：每个启用节点一条链接（host/port 没填则回落 cfg.PublicHost/PublicPort）
//   - 若一条也没有：用 cfg.PublicHost/PublicPort 生成一条"默认节点"链接，方便单机起步
func (s *SubscriptionService) BuildForUser(ctx context.Context, user *model.User) (*Payload, error) {
	if user.SubscriptionToken == "" {
		return nil, errors.New("user has no subscription token")
	}

	var nodes []model.Node
	if err := s.db.WithContext(ctx).Where("status = ?", 1).Order("id asc").Find(&nodes).Error; err != nil {
		return nil, err
	}

	links := make([]PerNodeLink, 0)
	if len(nodes) == 0 {
		links = append(links, s.buildLink(user, "default", s.cfg.PublicHost, s.cfg.PublicPort))
	} else {
		for _, n := range nodes {
			host := n.PublicHost
			if host == "" {
				host = s.cfg.PublicHost
			}
			port := n.PublicPort
			if port == 0 {
				port = s.cfg.PublicPort
			}
			links = append(links, s.buildLink(user, n.Name, host, port))
		}
	}

	raw := joinURLs(links)
	return &Payload{
		Username: user.Username,
		UUID:     user.UUID,
		Links:    links,
		RawText:  raw,
		Base64:   base64.StdEncoding.EncodeToString([]byte(raw)),
	}, nil
}

// FindByToken 让 public 订阅 endpoint 用 token 反查用户。
func (s *SubscriptionService) FindByToken(ctx context.Context, token string) (*model.User, error) {
	var u model.User
	err := s.db.WithContext(ctx).Where("subscription_token = ?", token).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *SubscriptionService) buildLink(user *model.User, nodeName, host string, port int) PerNodeLink {
	q := url.Values{}
	q.Set("encryption", "none")
	q.Set("flow", "xtls-rprx-vision")
	q.Set("security", "reality")
	q.Set("type", "tcp")
	if s.cfg.RealitySNI != "" {
		q.Set("sni", s.cfg.RealitySNI)
	}
	q.Set("fp", "chrome")
	if s.cfg.RealityPublicKey != "" {
		q.Set("pbk", s.cfg.RealityPublicKey)
	}
	if s.cfg.RealityShortID != "" {
		q.Set("sid", s.cfg.RealityShortID)
	}

	// fragment 用客户端能读的展示名："<user>@<node>"
	remark := fmt.Sprintf("%s@%s", user.Username, nodeName)
	urlStr := fmt.Sprintf(
		"vless://%s@%s:%s?%s#%s",
		user.UUID, host, strconv.Itoa(port), q.Encode(), url.QueryEscape(remark),
	)
	return PerNodeLink{NodeName: nodeName, Host: host, Port: port, URL: urlStr}
}

func joinURLs(links []PerNodeLink) string {
	urls := make([]string, 0, len(links))
	for _, l := range links {
		urls = append(urls, l.URL)
	}
	return strings.Join(urls, "\n")
}

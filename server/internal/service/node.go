package service

import (
	"context"
	"time"

	"github.com/routex/routex/internal/model"
	"github.com/routex/routex/internal/util"
	"gorm.io/gorm"
)

type NodeService struct {
	db *gorm.DB
}

func NewNodeService(db *gorm.DB) *NodeService { return &NodeService{db: db} }

type CreateNodeInput struct {
	Name       string `json:"name"        binding:"required"`
	IP         string `json:"ip"`          // 运维 IP（ssh / 监控用），与 PublicHost 可不同（如反代时）
	Region     string `json:"region"`
	PublicHost string `json:"public_host"` // 客户端 vless:// 连接的 host，留空则订阅服务回落到 .env 的 PUBLIC_HOST
	PublicPort int    `json:"public_port"` // 客户端 vless 端口，留空则回落到 .env 的 PUBLIC_PORT
}

type CreateNodeResult struct {
	Node      *model.Node `json:"node"`
	AuthToken string      `json:"auth_token"` // 仅创建时返回明文，让调用方记下来给 agent
}

// Create 新增节点并生成 agent 用的认证 token。
// 明文 token 只在本函数返回值里出现一次，后续 List/Get 都不会暴露。
func (s *NodeService) Create(ctx context.Context, in CreateNodeInput) (*CreateNodeResult, error) {
	token := util.RandPassword(40)
	node := &model.Node{
		Name:       in.Name,
		IP:         in.IP,
		Region:     in.Region,
		PublicHost: in.PublicHost,
		PublicPort: in.PublicPort,
		Status:     1,
		AuthToken:  token,
	}
	if err := s.db.WithContext(ctx).Create(node).Error; err != nil {
		return nil, err
	}
	return &CreateNodeResult{Node: node, AuthToken: token}, nil
}

func (s *NodeService) List(ctx context.Context) ([]model.Node, error) {
	var nodes []model.Node
	return nodes, s.db.WithContext(ctx).Order("id desc").Find(&nodes).Error
}

func (s *NodeService) Get(ctx context.Context, id int64) (*model.Node, error) {
	var n model.Node
	if err := s.db.WithContext(ctx).First(&n, id).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (s *NodeService) Delete(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Delete(&model.Node{}, id).Error
}

// FindByToken 给 NodeAuth 中间件用。
func (s *NodeService) FindByToken(ctx context.Context, token string) (*model.Node, error) {
	var n model.Node
	if err := s.db.WithContext(ctx).Where("auth_token = ?", token).First(&n).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

type HeartbeatInput struct {
	CPU       float64 `json:"cpu"`
	Memory    float64 `json:"memory"`
	Bandwidth int64   `json:"bandwidth"`
	Version   string  `json:"version"`
}

// Heartbeat 更新指标 + last_seen。被 NodeAuth 鉴权过的节点调用。
func (s *NodeService) Heartbeat(ctx context.Context, nodeID int64, in HeartbeatInput) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&model.Node{}).Where("id = ?", nodeID).
		Updates(map[string]any{
			"cpu":       in.CPU,
			"memory":    in.Memory,
			"bandwidth": in.Bandwidth,
			"version":   in.Version,
			"last_seen": &now,
		}).Error
}

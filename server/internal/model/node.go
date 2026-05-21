package model

import "time"

type Node struct {
	ID        int64   `gorm:"primaryKey" json:"id"`
	Name      string  `gorm:"size:64;uniqueIndex" json:"name"`
	IP        string  `gorm:"size:64" json:"ip"`
	Region    string  `gorm:"size:64" json:"region"`
	Status    int     `gorm:"default:1" json:"status"` // 1=enabled
	CPU       float64 `json:"cpu"`
	Memory    float64 `json:"memory"`
	Bandwidth int64   `json:"bandwidth"`
	Version   string  `gorm:"size:32" json:"version"`

	// AuthToken 是节点 agent 上报心跳时携带的凭证。
	// 仅在创建时返回明文给调用方一次，后续接口里通过 json:"-" 屏蔽。
	AuthToken string `gorm:"size:64;uniqueIndex" json:"-"`

	LastSeen  *time.Time `json:"last_seen"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

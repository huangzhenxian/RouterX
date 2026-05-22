package model

import "time"

// ProxyProvider 一条住宅 / 数据中心出口代理。
// 多条记录构成一个"代理池"，调度器周期检查可用性、延迟。
type ProxyProvider struct {
	ID       int64  `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"size:64" json:"name"`
	Type     string `gorm:"size:32" json:"type"` // socks5 / http / https
	Host     string `gorm:"size:128" json:"host"`
	Port     int    `json:"port"`
	Username string `gorm:"size:128" json:"username"`
	Password string `gorm:"size:128" json:"-"`
	Region   string `gorm:"size:64" json:"region"`
	Tags     string `gorm:"size:255" json:"tags"`     // 逗号分隔的标签，便于路由策略
	Priority int    `gorm:"default:10" json:"priority"` // 数字越小越优先

	Status        int        `gorm:"default:1" json:"status"`        // 1=enabled, 0=disabled
	Healthy       bool       `json:"healthy"`                        // 最近一次健康检查结果
	LatencyMillis int        `json:"latency_ms"`                     // 最近一次测得的延迟（ms）
	FailCount     int        `json:"fail_count"`                     // 连续失败次数，触发自动禁用
	LastCheckedAt *time.Time `json:"last_checked_at"`
	LastError     string     `gorm:"size:255" json:"last_error"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

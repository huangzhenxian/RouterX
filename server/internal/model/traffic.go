package model

import "time"

// UserTraffic 用户每周期流量明细（按 scheduler 拉取频率写入）。
// 真正聚合查询走天/月预聚合表，后续二期再补 user_traffic_daily。
type UserTraffic struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    int64     `gorm:"index" json:"user_id"`
	Upload    int64     `json:"upload"`
	Download  int64     `json:"download"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

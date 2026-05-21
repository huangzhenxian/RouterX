package model

import "time"

type User struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex" json:"username"`
	UUID         string    `gorm:"size:128;uniqueIndex" json:"uuid"`
	Password     string    `gorm:"size:128" json:"-"`
	Status       int       `gorm:"default:1" json:"status"` // 1=enabled, 0=disabled
	TrafficLimit int64     `json:"traffic_limit"`           // bytes
	UsedTraffic  int64     `json:"used_traffic"`            // bytes
	ExpireTime   time.Time `json:"expire_time"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

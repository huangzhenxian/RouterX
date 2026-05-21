package model

import "time"

type ProxyProvider struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Type      string    `gorm:"size:32" json:"type"` // socks5 / http / https
	Host      string    `gorm:"size:128" json:"host"`
	Port      int       `json:"port"`
	Username  string    `gorm:"size:128" json:"username"`
	Password  string    `gorm:"size:128" json:"-"`
	Region    string    `gorm:"size:64" json:"region"`
	Status    int       `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

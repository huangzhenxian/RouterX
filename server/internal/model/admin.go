package model

import "time"

// Admin 后台管理员账号。与 User（代理用户）完全隔离。
type Admin struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex" json:"username"`
	PasswordHash string    `gorm:"size:128" json:"-"`
	Role         string    `gorm:"size:32;default:super" json:"role"` // super / op / readonly
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

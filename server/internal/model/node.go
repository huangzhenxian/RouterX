package model

import "time"

type Node struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:64" json:"name"`
	IP        string    `gorm:"size:64" json:"ip"`
	Region    string    `gorm:"size:64" json:"region"`
	Status    int       `gorm:"default:1" json:"status"`
	CPU       float64   `json:"cpu"`
	Memory    float64   `json:"memory"`
	Bandwidth int64     `json:"bandwidth"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

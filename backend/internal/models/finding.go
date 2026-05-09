package models

import "time"

type Finding struct {
	ID          string    `gorm:"primaryKey;type:uuid" json:"id"`
	AssetID     string    `gorm:"not null" json:"asset_id"`
	CheckType   string    `json:"check_type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

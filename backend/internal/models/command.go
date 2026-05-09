package models

import "time"

type Command struct {
	ID          string     `gorm:"primaryKey;type:uuid" json:"id"`
	AssetID     string     `gorm:"not null;index" json:"asset_id"`
	Type        string     `gorm:"type:varchar(50);not null" json:"type"`
	Params      string     `gorm:"type:text" json:"params"`
	Status      string     `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Result      string     `gorm:"type:text" json:"result"`
	CreatedAt   time.Time  `json:"created_at"`
	SentAt      *time.Time `json:"sent_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

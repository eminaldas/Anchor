package models

import "time"

// AdminWhitelist — Yönetilebilir izinli admin listesi
type AdminWhitelist struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"type:varchar(255);unique;not null" json:"username"`
	Reason    string    `json:"reason"`
	AddedBy   string    `json:"added_by"`
	CreatedAt time.Time `json:"created_at"`
}

package models

import "time"

// FIMConfig — İzlenecek dosya yapılandırması (yönetilebilir)
type FIMConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FilePath  string    `gorm:"type:varchar(500);unique;not null" json:"file_path"`
	Label     string    `json:"label"`
	Priority  string    `gorm:"type:varchar(20);default:'High'" json:"priority"` // Critical, High, Medium
	CreatedAt time.Time `json:"created_at"`
}

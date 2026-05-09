package models

import "time"

type SoftwarePolicy struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(255);unique;not null" json:"name"`
	Status    string    `gorm:"type:varchar(50);default:'Banned'" json:"status"`
	RiskLevel string    `gorm:"type:varchar(50);default:'High'" json:"risk_level"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

type SoftwareCatalog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255);unique;not null" json:"name"`
	FirstSeenAt time.Time `json:"first_seen_at"`
}

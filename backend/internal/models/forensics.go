package models

import (
	"time"

	"gorm.io/datatypes"
)

// ForensicSnapshot — Her denetim çalışmasının ham kaydı
type ForensicSnapshot struct {
	ID        string         `gorm:"primaryKey;type:uuid" json:"id"`
	AssetID   string         `gorm:"not null;index" json:"asset_id"`
	Type      string         `gorm:"type:varchar(50)" json:"type"` // local_admins, events, fim
	Data      datatypes.JSON `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
}

// SecurityEvent — Normalize edilmiş Windows Event Log kaydı
type SecurityEvent struct {
	ID              string    `gorm:"primaryKey;type:uuid" json:"id"`
	AssetID         string    `gorm:"not null;index" json:"asset_id"`
	EventID         int       `json:"event_id"`
	EventType       string    `gorm:"type:varchar(50)" json:"event_type"` // logon_success, logon_failed, logoff, user_created, group_member_added
	TargetUser      string    `json:"target_user"`
	SourceIP        string    `json:"source_ip"`
	WorkstationName string    `json:"workstation_name"`
	LogonType       int       `json:"logon_type"`
	EventTime       time.Time `json:"event_time"`
	CreatedAt       time.Time `json:"created_at"`
}

// FileIntegrityRecord — Dosya bütünlüğü hash kaydı
type FileIntegrityRecord struct {
	ID           string    `gorm:"primaryKey;type:uuid" json:"id"`
	AssetID      string    `gorm:"not null;index" json:"asset_id"`
	FilePath     string    `json:"file_path"`
	SHA256       string    `gorm:"type:varchar(64)" json:"sha256"`
	FileSize     int64     `json:"file_size"`
	LastModified time.Time `json:"last_modified"`
	Status       string    `gorm:"type:varchar(20);default:'baseline'" json:"status"` // baseline, unchanged, changed
	PreviousHash string    `json:"previous_hash,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

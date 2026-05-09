package models

import (
	"time"

	"gorm.io/datatypes"
)

type Asset struct {
	ID              string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Hostname        string         `gorm:"type:varchar(255)" json:"hostname"`
	LastSeen        time.Time      `json:"last_seen"`
	CPUUsage        string         `json:"cpu_usage"`
	RAMUsage        string         `json:"ram_usage"`
	Status          string         `gorm:"type:varchar(50);default:'Online'" json:"status"`
	Uptime          uint64         `json:"uptime"`
	DefenderEnabled bool           `json:"defender_enabled"`
	ComplianceScore int            `gorm:"default:100" json:"compliance_score"`
	SoftwareList    datatypes.JSON `json:"software_list"`
}

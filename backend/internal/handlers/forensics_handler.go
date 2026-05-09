package handlers

import (
	"audit-backend/internal/database"
	"audit-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetForensicSnapshots — Denetim snapshot'larını listeler
func GetForensicSnapshots(c *gin.Context) {
	var snapshots []models.ForensicSnapshot
	query := database.DB.Order("created_at desc")

	if assetID := c.Query("asset_id"); assetID != "" {
		query = query.Where("asset_id = ?", assetID)
	}
	if snapType := c.Query("type"); snapType != "" {
		query = query.Where("type = ?", snapType)
	}

	if err := query.Limit(100).Find(&snapshots).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Snapshot'lar getirilemedi"})
		return
	}

	c.JSON(http.StatusOK, snapshots)
}

// GetSecurityEvents — Güvenlik olaylarını listeler (filtrelemeli)
func GetSecurityEvents(c *gin.Context) {
	var events []models.SecurityEvent
	query := database.DB.Order("event_time desc")

	if assetID := c.Query("asset_id"); assetID != "" {
		query = query.Where("asset_id = ?", assetID)
	}
	if eventType := c.Query("type"); eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if from := c.Query("from"); from != "" {
		query = query.Where("event_time >= ?", from)
	}
	if to := c.Query("to"); to != "" {
		query = query.Where("event_time <= ?", to)
	}

	if err := query.Limit(500).Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Güvenlik olayları getirilemedi"})
		return
	}

	c.JSON(http.StatusOK, events)
}

// GetFileIntegrityRecords — FIM kayıtlarını listeler
func GetFileIntegrityRecords(c *gin.Context) {
	var records []models.FileIntegrityRecord
	query := database.DB.Order("created_at desc")

	if assetID := c.Query("asset_id"); assetID != "" {
		query = query.Where("asset_id = ?", assetID)
	}

	if err := query.Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "FIM kayıtları getirilemedi"})
		return
	}

	c.JSON(http.StatusOK, records)
}

// GetForensicsSummary — Tek bir cihaz için forensics özet
func GetForensicsSummary(c *gin.Context) {
	assetID := c.Param("asset_id")

	// Son admin snapshot'ı
	var lastAdminSnap models.ForensicSnapshot
	database.DB.Where("asset_id = ? AND type = ?", assetID, "local_admins").Order("created_at desc").First(&lastAdminSnap)

	// Güvenlik olayı istatistikleri
	var failedLogins int64
	database.DB.Model(&models.SecurityEvent{}).Where("asset_id = ? AND event_type = ?", assetID, "logon_failed").Count(&failedLogins)

	var successLogins int64
	database.DB.Model(&models.SecurityEvent{}).Where("asset_id = ? AND event_type = ?", assetID, "logon_success").Count(&successLogins)

	var totalEvents int64
	database.DB.Model(&models.SecurityEvent{}).Where("asset_id = ?", assetID).Count(&totalEvents)

	// FIM durumu
	var fimTotal int64
	database.DB.Model(&models.FileIntegrityRecord{}).Where("asset_id = ?", assetID).Count(&fimTotal)

	var fimChanged int64
	database.DB.Model(&models.FileIntegrityRecord{}).Where("asset_id = ? AND status = ?", assetID, "changed").Count(&fimChanged)

	// Deep Audit Finding'leri
	var deepFindings int64
	database.DB.Model(&models.Finding{}).Where("asset_id = ? AND status = ? AND check_type LIKE ?", assetID, "Open", "Deep_%").Count(&deepFindings)

	c.JSON(http.StatusOK, gin.H{
		"last_admin_snapshot": lastAdminSnap,
		"event_stats": gin.H{
			"total":          totalEvents,
			"failed_logins":  failedLogins,
			"success_logins": successLogins,
		},
		"fim_stats": gin.H{
			"total":   fimTotal,
			"changed": fimChanged,
		},
		"open_deep_findings": deepFindings,
	})
}

// GetForensicsDashboard — Global forensics dashboard verileri
func GetForensicsDashboard(c *gin.Context) {
	// Son kritik finding'ler (tüm cihazlardan)
	var recentFindings []models.Finding
	database.DB.Where("status = ?", "Open").Order("created_at desc").Limit(50).Find(&recentFindings)

	// Son güvenlik olayları
	var recentEvents []models.SecurityEvent
	database.DB.Order("event_time desc").Limit(30).Find(&recentEvents)

	// FIM ihlalleri
	var fimAlerts []models.FileIntegrityRecord
	database.DB.Where("status = ?", "changed").Order("created_at desc").Limit(20).Find(&fimAlerts)

	// Genel istatistikler
	var totalOpenFindings int64
	database.DB.Model(&models.Finding{}).Where("status = ?", "Open").Count(&totalOpenFindings)

	var criticalCount int64
	database.DB.Model(&models.Finding{}).Where("status = ? AND severity = ?", "Open", "Critical").Count(&criticalCount)

	var highCount int64
	database.DB.Model(&models.Finding{}).Where("status = ? AND severity = ?", "Open", "High").Count(&highCount)

	var mediumCount int64
	database.DB.Model(&models.Finding{}).Where("status = ? AND severity = ?", "Open", "Medium").Count(&mediumCount)

	var totalFailedLogins int64
	database.DB.Model(&models.SecurityEvent{}).Where("event_type = ?", "logon_failed").Count(&totalFailedLogins)

	var totalFIMChanges int64
	database.DB.Model(&models.FileIntegrityRecord{}).Where("status = ?", "changed").Count(&totalFIMChanges)

	// Aktif cihaz sayısı
	var onlineAssets int64
	database.DB.Model(&models.Asset{}).Where("status = ?", "Online").Count(&onlineAssets)

	var totalAssets int64
	database.DB.Model(&models.Asset{}).Count(&totalAssets)

	c.JSON(http.StatusOK, gin.H{
		"recent_findings": recentFindings,
		"recent_events":   recentEvents,
		"fim_alerts":      fimAlerts,
		"stats": gin.H{
			"total_open_findings": totalOpenFindings,
			"critical_count":     criticalCount,
			"high_count":         highCount,
			"medium_count":       mediumCount,
			"failed_logins":      totalFailedLogins,
			"fim_changes":        totalFIMChanges,
			"online_assets":      onlineAssets,
			"total_assets":       totalAssets,
		},
	})
}

package handlers

import (
	"audit-backend/internal/database"
	"audit-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetFindings fetches findings — opsiyonel asset_id ve status filtresi
func GetFindings(c *gin.Context) {
	var findings []models.Finding
	query := database.DB.Order("created_at desc")

	if assetID := c.Query("asset_id"); assetID != "" {
		query = query.Where("asset_id = ?", assetID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&findings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Veriler getirilemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total_findings": len(findings),
		"data":           findings,
	})
}

// UpdateFindingStatus updates the status of a finding
func UpdateFindingStatus(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	result := database.DB.Model(&models.Finding{}).Where("id = ?", id).Update("status", input.Status)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Güncelleme başarısız"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Durum güncellendi"})
}

// GetStats fetches the overall system statistics
func GetStats(c *gin.Context) {
	var stats struct {
		OnlineCount  int64 `json:"online_count"`
		TotalAssets  int64 `json:"total_assets"`
		FindingStats []struct {
			Severity string `json:"severity"`
			Count    int64  `json:"count"`
		} `json:"finding_stats"`
	}

	database.DB.Model(&models.Asset{}).Count(&stats.TotalAssets)
	database.DB.Model(&models.Asset{}).Where("status = ?", "Online").Count(&stats.OnlineCount)

	database.DB.Model(&models.Finding{}).
		Select("severity, count(*) as count").
		Where("status = ?", "Open").
		Group("severity").
		Scan(&stats.FindingStats)

	c.JSON(http.StatusOK, stats)
}

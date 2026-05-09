package handlers

import (
	"audit-backend/internal/database"
	"audit-backend/internal/models"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetAssets fetches all assets
func GetAssets(c *gin.Context) {
	var assets []models.Asset
	if err := database.DB.Find(&assets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cihazlar getirilemedi"})
		return
	}
	c.JSON(http.StatusOK, assets)
}

// GetAssetByID fetches a single asset by its ID
func GetAssetByID(c *gin.Context) {
	id := c.Param("id")
	var asset models.Asset
	if err := database.DB.First(&asset, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cihaz bulunamadı"})
		return
	}
	c.JSON(http.StatusOK, asset)
}

// GetUniqueSoftwareList extracts unique software names from all assets
func GetUniqueSoftwareList(c *gin.Context) {
	var assets []models.Asset
	if err := database.DB.Find(&assets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Yazılımlar getirilemedi"})
		return
	}

	softwareSet := make(map[string]bool)

	for _, asset := range assets {
		if asset.SoftwareList != nil {
			var list []string
			if err := json.Unmarshal(asset.SoftwareList, &list); err == nil {
				for _, app := range list {
					softwareSet[app] = true
				}
			}
		}
	}

	var uniqueList []string
	for app := range softwareSet {
		uniqueList = append(uniqueList, app)
	}

	c.JSON(http.StatusOK, uniqueList)
}

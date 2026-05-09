package handlers

import (
	"audit-backend/internal/database"
	"audit-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetFIMConfigs — Tüm FIM yapılandırmalarını listeler
func GetFIMConfigs(c *gin.Context) {
	var configs []models.FIMConfig
	if err := database.DB.Order("priority asc, created_at desc").Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "FIM yapılandırmaları getirilemedi"})
		return
	}
	c.JSON(http.StatusOK, configs)
}

// AddFIMConfig — Yeni FIM dosya kaydı ekler
func AddFIMConfig(c *gin.Context) {
	var config models.FIMConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Create(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "FIM yapılandırması oluşturulamadı"})
		return
	}

	c.JSON(http.StatusCreated, config)
}

// DeleteFIMConfig — FIM dosya kaydını siler
func DeleteFIMConfig(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.FIMConfig{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "FIM yapılandırması silinemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "FIM yapılandırması silindi"})
}

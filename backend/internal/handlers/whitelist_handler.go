package handlers

import (
	"audit-backend/internal/database"
	"audit-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetWhitelist — Tüm whitelist'i listeler
func GetWhitelist(c *gin.Context) {
	var list []models.AdminWhitelist
	if err := database.DB.Order("created_at desc").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Whitelist getirilemedi"})
		return
	}
	c.JSON(http.StatusOK, list)
}

// AddWhitelistEntry — Whitelist'e yeni kullanıcı ekler
func AddWhitelistEntry(c *gin.Context) {
	var entry models.AdminWhitelist
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Create(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Whitelist kaydı oluşturulamadı"})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

// DeleteWhitelistEntry — Whitelist'ten kullanıcı siler
func DeleteWhitelistEntry(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.AdminWhitelist{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Whitelist kaydı silinemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Whitelist kaydı silindi"})
}

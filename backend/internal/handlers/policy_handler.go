package handlers

import (
	"audit-backend/internal/database"
	"audit-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetPolicies fetches all software policies
func GetPolicies(c *gin.Context) {
	var policies []models.SoftwarePolicy
	if err := database.DB.Find(&policies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Politikalar getirilemedi"})
		return
	}
	c.JSON(http.StatusOK, policies)
}

// AddPolicy creates a new software policy
func AddPolicy(c *gin.Context) {
	var policy models.SoftwarePolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Create(&policy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Politika oluşturulamadı"})
		return
	}

	c.JSON(http.StatusCreated, policy)
}

// DeletePolicy removes a software policy
func DeletePolicy(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.SoftwarePolicy{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Politika silinemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Politika silindi"})
}

// GetCatalog fetches all unique software from the persistent catalog
func GetCatalog(c *gin.Context) {
	var catalog []models.SoftwareCatalog
	if err := database.DB.Order("name asc").Find(&catalog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Katalog getirilemedi"})
		return
	}
	c.JSON(http.StatusOK, catalog)
}

// BulkUpdatePolicies allows banning or approving multiple applications at once
func BulkUpdatePolicies(c *gin.Context) {
	var input struct {
		Apps      []string `json:"apps"`
		Action    string   `json:"action"` // "ban" or "approve"
		RiskLevel string   `json:"risk_level"`
		Reason    string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	for _, appName := range input.Apps {
		if input.Action == "ban" {
			var policy models.SoftwarePolicy
			// Varsa bul ve güncelle, yoksa oluştur
			result := database.DB.Where("name = ?", appName).First(&policy)
			if result.Error != nil {
				// Oluştur
				policy = models.SoftwarePolicy{
					Name:      appName,
					Status:    "Banned",
					RiskLevel: input.RiskLevel,
					Reason:    input.Reason,
				}
				database.DB.Create(&policy)
			} else {
				// Güncelle
				policy.Status = "Banned"
				policy.RiskLevel = input.RiskLevel
				policy.Reason = input.Reason
				database.DB.Save(&policy)
			}
		} else if input.Action == "approve" {
			// Yasağı kaldır (politikayı sil)
			database.DB.Where("name = ?", appName).Delete(&models.SoftwarePolicy{})
			
			// Bu uygulama için açılmış olan Policy_Violation finding'leri otomatik kapat
			database.DB.Model(&models.Finding{}).
				Where("check_type = ? AND description LIKE ? AND status = ?", "Policy_Violation", "%"+appName+"%", "Open").
				Update("status", "Closed")
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Toplu güncelleme başarılı"})
}

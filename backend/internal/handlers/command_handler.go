package handlers

import (
	"audit-backend/internal/database"
	"audit-backend/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateCommand — Dashboard'dan yeni C2 komutu oluşturur
func CreateCommand(c *gin.Context) {
	var input struct {
		AssetID string `json:"asset_id" binding:"required"`
		Type    string `json:"type" binding:"required"`
		Params  string `json:"params"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Sadece izin verilen komut tiplerini kabul et
	allowedTypes := map[string]bool{
		"screenshot":     true,
		"list_processes": true,
		"sysinfo":        true,
	}

	if !allowedTypes[input.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz komut tipi: " + input.Type})
		return
	}

	cmd := models.Command{
		ID:        uuid.New().String(),
		AssetID:   input.AssetID,
		Type:      input.Type,
		Params:    input.Params,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	if err := database.DB.Create(&cmd).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Komut oluşturulamadı"})
		return
	}

	c.JSON(http.StatusCreated, cmd)
}

// GetCommands — Komut geçmişini listeler (opsiyonel asset_id filtresi)
func GetCommands(c *gin.Context) {
	var commands []models.Command
	query := database.DB.Order("created_at desc")

	if assetID := c.Query("asset_id"); assetID != "" {
		query = query.Where("asset_id = ?", assetID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Limit(50).Find(&commands).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Komutlar getirilemedi"})
		return
	}

	c.JSON(http.StatusOK, commands)
}

// GetCommandByID — Tek bir komutun detayını getirir
func GetCommandByID(c *gin.Context) {
	id := c.Param("id")
	var cmd models.Command

	if err := database.DB.First(&cmd, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Komut bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, cmd)
}

// SubmitCommandResult — Ajan komut sonucunu gönderir
func SubmitCommandResult(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		Status string `json:"status" binding:"required"` // completed veya failed
		Result string `json:"result"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cmd models.Command
	if err := database.DB.First(&cmd, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Komut bulunamadı"})
		return
	}

	now := time.Now()
	cmd.Status = input.Status
	cmd.Result = input.Result
	cmd.CompletedAt = &now

	if err := database.DB.Save(&cmd).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Sonuç kaydedilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Komut sonucu kaydedildi"})
}

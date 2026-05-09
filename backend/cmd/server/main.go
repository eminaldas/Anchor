package main

import (
	"audit-backend/internal/database"
	"audit-backend/internal/handlers"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

func main() {
	// Veritabanı bağlantısını başlat
	database.InitDB()

	r := gin.Default()
	r.Use(cors.Default())
	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "Sentinel Brain is Active"})
		})

		v1.POST("/report", handlers.ReportFinding)
		v1.GET("/findings", handlers.GetFindings)
		v1.PATCH("/findings/:id/status", handlers.UpdateFindingStatus)
		v1.GET("/stats", handlers.GetStats)
		v1.GET("/assets", handlers.GetAssets)
		v1.GET("/assets/:id", handlers.GetAssetByID)
		
		// Policies & Envanter
		v1.GET("/inventory/software", handlers.GetUniqueSoftwareList)
		v1.GET("/inventory/catalog", handlers.GetCatalog)

		// Politikalar
		v1.GET("/policies", handlers.GetPolicies)
		v1.POST("/policies", handlers.AddPolicy)
		v1.POST("/policies/bulk", handlers.BulkUpdatePolicies)
		v1.DELETE("/policies/:id", handlers.DeletePolicy)

		// C2 Command & Control
		v1.POST("/commands", handlers.CreateCommand)
		v1.GET("/commands", handlers.GetCommands)
		v1.GET("/commands/:id", handlers.GetCommandByID)
		v1.POST("/commands/:id/result", handlers.SubmitCommandResult)

		// Forensics & Deep Audit
		v1.GET("/forensics/snapshots", handlers.GetForensicSnapshots)
		v1.GET("/forensics/events", handlers.GetSecurityEvents)
		v1.GET("/forensics/fim", handlers.GetFileIntegrityRecords)
		v1.GET("/forensics/summary/:asset_id", handlers.GetForensicsSummary)
		v1.GET("/forensics/dashboard", handlers.GetForensicsDashboard)

		// Admin — Whitelist Yönetimi
		v1.GET("/admin/whitelist", handlers.GetWhitelist)
		v1.POST("/admin/whitelist", handlers.AddWhitelistEntry)
		v1.DELETE("/admin/whitelist/:id", handlers.DeleteWhitelistEntry)

		// Admin — FIM Yapılandırma
		v1.GET("/admin/fim-config", handlers.GetFIMConfigs)
		v1.POST("/admin/fim-config", handlers.AddFIMConfig)
		v1.DELETE("/admin/fim-config/:id", handlers.DeleteFIMConfig)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
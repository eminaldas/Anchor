package main

import (
	"audit-backend/internal/database"
	"audit-backend/internal/models"
	"encoding/json"
	"fmt"
	"time"
)

func main() {
	database.InitDB()

	var assets []models.Asset
	database.DB.Find(&assets)

	count := 0
	for _, asset := range assets {
		if asset.SoftwareList != nil {
			var list []string
			if err := json.Unmarshal(asset.SoftwareList, &list); err == nil {
				for _, app := range list {
					result := database.DB.FirstOrCreate(&models.SoftwareCatalog{}, models.SoftwareCatalog{Name: app, FirstSeenAt: time.Now().Add(-48 * time.Hour)})
					if result.RowsAffected > 0 {
						count++
					}
				}
			}
		}
	}
	fmt.Printf("Eski cihazlardan toplam %d yazılım kataloga aktarıldı.\n", count)
}

package handlers

import (
	"audit-backend/internal/database"
	"audit-backend/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func ReportFinding(c *gin.Context) {
	var incoming models.Finding

	if err := c.ShouldBindJSON(&incoming); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parts := strings.Split(incoming.AssetID, "_")
	hostname := parts[0]

	var existing models.Asset
	database.DB.FirstOrCreate(&existing, models.Asset{ID: incoming.AssetID, Hostname: hostname})

	existing.LastSeen = time.Now()
	existing.Status = "Online"

	if incoming.CheckType == "Software_Inventory" {
		existing.SoftwareList = datatypes.JSON(incoming.Description)
		
		// KURAL KONTROLÜ (POLICY CHECK)
		// 1. Gelen yazılım listesini parse et
		var installedApps []string
		if err := json.Unmarshal([]byte(incoming.Description), &installedApps); err == nil {
			// KATALOG BESLEMESİ
			for _, app := range installedApps {
				database.DB.FirstOrCreate(&models.SoftwareCatalog{}, models.SoftwareCatalog{Name: app, FirstSeenAt: time.Now()})
			}

			// 2. Veritabanındaki yasaklı kuralları çek
			var policies []models.SoftwarePolicy
			database.DB.Where("status = ?", "Banned").Find(&policies)
			
			// Hızlı arama için bir map oluşturalım
			bannedMap := make(map[string]models.SoftwarePolicy)
			for _, p := range policies {
				bannedMap[p.Name] = p
			}

			// 3. Her bir yüklü yazılım kural ihlali yaratıyor mu kontrol et
			for _, app := range installedApps {
				if policy, exists := bannedMap[app]; exists {
					// İhlal bulundu! Finding oluştur.
					violation := models.Finding{
						ID:          uuid.New().String(),
						AssetID:     incoming.AssetID,
						CheckType:   "Policy_Violation",
						Severity:    policy.RiskLevel,
						Description: fmt.Sprintf("Yasaklı Yazılım Tespit Edildi: %s (Sebep: %s)", app, policy.Reason),
						Status:      "Open",
						CreatedAt:   time.Now(),
					}
					database.DB.Create(&violation)
				}
			}
		}
	} else if incoming.CheckType == "Heartbeat_Metrics" {
		var payload struct {
			CPU             string `json:"cpu"`
			RAM             string `json:"ram"`
			Uptime          uint64 `json:"uptime"`
			DefenderEnabled bool   `json:"defender_enabled"`
		}

		if err := json.Unmarshal([]byte(incoming.Description), &payload); err == nil {
			existing.CPUUsage = payload.CPU
			existing.RAMUsage = payload.RAM
			existing.Uptime = payload.Uptime
			existing.DefenderEnabled = payload.DefenderEnabled

			// KURAL 1: Windows Defender Kapalıysa "High" Risk
			if !payload.DefenderEnabled {
				var count int64
				database.DB.Model(&models.Finding{}).Where("asset_id = ? AND status = ? AND check_type = ?", existing.ID, "Open", "Defender_Check").Count(&count)
				if count == 0 {
					violation := models.Finding{
						ID:          uuid.New().String(),
						AssetID:     existing.ID,
						CheckType:   "Defender_Check",
						Severity:    "High",
						Description: "Windows Defender devre dışı bırakılmış!",
						Status:      "Open",
						CreatedAt:   time.Now(),
					}
					database.DB.Create(&violation)
				}
			}

			// KURAL 2: Uptime > 15 gün (1296000 saniye) ise "Medium" Risk
			if payload.Uptime > 1296000 {
				var count int64
				database.DB.Model(&models.Finding{}).Where("asset_id = ? AND status = ? AND check_type = ?", existing.ID, "Open", "Uptime_Check").Count(&count)
				if count == 0 {
					violation := models.Finding{
						ID:          uuid.New().String(),
						AssetID:     existing.ID,
						CheckType:   "Uptime_Check",
						Severity:    "Medium",
						Description: fmt.Sprintf("Sistem 15 günden uzun süredir açık (Uptime: %d saniye). Yeniden başlatılması önerilir.", payload.Uptime),
						Status:      "Open",
						CreatedAt:   time.Now(),
					}
					database.DB.Create(&violation)
				}
			}
		} else {
			fmt.Println("JSON Parse Hatasi:", err)
			existing.CPUUsage = incoming.Description
		}
	} else if incoming.CheckType == "Deep_Audit_Admins" {
		processDeepAuditAdmins(incoming)
	} else if incoming.CheckType == "Deep_Audit_Events" {
		processDeepAuditEvents(incoming)
	} else if incoming.CheckType == "Deep_Audit_FIM" {
		processDeepAuditFIM(incoming)
	}

	// Calculate Compliance Score
	var openFindings []models.Finding
	database.DB.Where("asset_id = ? AND status = ?", existing.ID, "Open").Find(&openFindings)
	
	score := 100
	for _, f := range openFindings {
		switch f.Severity {
		case "Critical":
			score -= 30
		case "High":
			score -= 20
		case "Medium":
			score -= 10
		case "Low":
			score -= 5
		}
	}
	if score < 0 {
		score = 0
	}
	existing.ComplianceScore = score

	if err := database.DB.Save(&existing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Asset güncellenemedi"})
		return
	}

	// C2: Bekleyen komutları bul ve response'a ekle
	var pendingCmds []models.Command
	database.DB.Where("asset_id = ? AND status = ?", existing.ID, "pending").Find(&pendingCmds)

	now := time.Now()
	for i := range pendingCmds {
		pendingCmds[i].Status = "sent"
		pendingCmds[i].SentAt = &now
		database.DB.Save(&pendingCmds[i])
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Asset status updated",
		"type":     incoming.CheckType,
		"commands": pendingCmds,
	})
}

// ──────────────────── Deep Audit Processors ────────────────────

func processDeepAuditAdmins(incoming models.Finding) {
	// 1. Snapshot kaydet
	snapshot := models.ForensicSnapshot{
		ID:        uuid.New().String(),
		AssetID:   incoming.AssetID,
		Type:      "local_admins",
		Data:      datatypes.JSON(incoming.Description),
		CreatedAt: time.Now(),
	}
	database.DB.Create(&snapshot)

	// 2. Gelen admin listesini parse et
	var adminList []string
	if err := json.Unmarshal([]byte(incoming.Description), &adminList); err != nil {
		return
	}

	// 3. Whitelist'i veritabanından çek
	var whitelist []models.AdminWhitelist
	database.DB.Find(&whitelist)
	
	whitelistMap := make(map[string]bool)
	for _, w := range whitelist {
		whitelistMap[strings.ToLower(w.Username)] = true
	}

	// 4. Whitelist dışı admin kontrolü
	for _, admin := range adminList {
		adminLower := strings.ToLower(admin)
		// Domain prefix'i temizle (DOMAIN\user → user)
		if parts := strings.Split(adminLower, `\`); len(parts) > 1 {
			adminLower = parts[len(parts)-1]
		}

		if !whitelistMap[adminLower] {
			// Daha önce aynı finding var mı kontrol et
			var count int64
			database.DB.Model(&models.Finding{}).Where(
				"asset_id = ? AND status = ? AND check_type = ? AND description LIKE ?",
				incoming.AssetID, "Open", "Deep_Admin_Violation", "%"+admin+"%",
			).Count(&count)

			if count == 0 {
				violation := models.Finding{
					ID:          uuid.New().String(),
					AssetID:     incoming.AssetID,
					CheckType:   "Deep_Admin_Violation",
					Severity:    "Critical",
					Description: fmt.Sprintf("Yetkisiz Administrator Tespit Edildi: %s", admin),
					Status:      "Open",
					CreatedAt:   time.Now(),
				}
				database.DB.Create(&violation)
			}
		}
	}
}

func processDeepAuditEvents(incoming models.Finding) {
	// 1. Snapshot kaydet
	snapshot := models.ForensicSnapshot{
		ID:        uuid.New().String(),
		AssetID:   incoming.AssetID,
		Type:      "events",
		Data:      datatypes.JSON(incoming.Description),
		CreatedAt: time.Now(),
	}
	database.DB.Create(&snapshot)

	// 2. Event'leri parse et ve veritabanına yaz
	var events []struct {
		EventID         int    `json:"event_id"`
		TimeCreated     string `json:"time_created"`
		TargetUserName  string `json:"target_user_name"`
		LogonType       int    `json:"logon_type"`
		IpAddress       string `json:"ip_address"`
		WorkstationName string `json:"workstation_name"`
	}

	if err := json.Unmarshal([]byte(incoming.Description), &events); err != nil {
		return
	}

	for _, evt := range events {
		eventTime, _ := time.Parse(time.RFC3339, evt.TimeCreated)
		
		eventType := "unknown"
		switch evt.EventID {
		case 4624:
			eventType = "logon_success"
		case 4625:
			eventType = "logon_failed"
		case 4634:
			eventType = "logoff"
		case 4720:
			eventType = "user_created"
		case 4732:
			eventType = "group_member_added"
		}

		secEvent := models.SecurityEvent{
			ID:              uuid.New().String(),
			AssetID:         incoming.AssetID,
			EventID:         evt.EventID,
			EventType:       eventType,
			TargetUser:      evt.TargetUserName,
			SourceIP:        evt.IpAddress,
			WorkstationName: evt.WorkstationName,
			LogonType:       evt.LogonType,
			EventTime:       eventTime,
			CreatedAt:       time.Now(),
		}
		database.DB.Create(&secEvent)
	}

	// 3. KURAL KONTROLLERI

	// Kural A: Brute Force — Son 5 dk'da aynı kullanıcı için 5+ başarısız giriş
	fiveMinAgo := time.Now().Add(-5 * time.Minute)
	type BruteForceResult struct {
		TargetUser string
		Count      int64
	}
	var bfResults []BruteForceResult
	database.DB.Model(&models.SecurityEvent{}).
		Select("target_user, count(*) as count").
		Where("asset_id = ? AND event_type = ? AND event_time > ?", incoming.AssetID, "logon_failed", fiveMinAgo).
		Group("target_user").
		Having("count(*) >= 5").
		Scan(&bfResults)

	for _, bf := range bfResults {
		var count int64
		database.DB.Model(&models.Finding{}).Where(
			"asset_id = ? AND status = ? AND check_type = ? AND description LIKE ? AND created_at > ?",
			incoming.AssetID, "Open", "Deep_BruteForce", "%"+bf.TargetUser+"%", fiveMinAgo,
		).Count(&count)

		if count == 0 {
			violation := models.Finding{
				ID:          uuid.New().String(),
				AssetID:     incoming.AssetID,
				CheckType:   "Deep_BruteForce",
				Severity:    "High",
				Description: fmt.Sprintf("Brute Force Şüphesi: %s için %d başarısız giriş denemesi (son 5 dakika)", bf.TargetUser, bf.Count),
				Status:      "Open",
				CreatedAt:   time.Now(),
			}
			database.DB.Create(&violation)
		}
	}

	// Kural B: Yeni Kullanıcı Oluşturma (Event 4720)
	for _, evt := range events {
		if evt.EventID == 4720 {
			violation := models.Finding{
				ID:          uuid.New().String(),
				AssetID:     incoming.AssetID,
				CheckType:   "Deep_UserCreated",
				Severity:    "High",
				Description: fmt.Sprintf("Yeni Kullanıcı Hesabı Oluşturuldu: %s", evt.TargetUserName),
				Status:      "Open",
				CreatedAt:   time.Now(),
			}
			database.DB.Create(&violation)
		}

		// Kural C: Güvenlik grubuna ekleme (Event 4732)
		if evt.EventID == 4732 {
			violation := models.Finding{
				ID:          uuid.New().String(),
				AssetID:     incoming.AssetID,
				CheckType:   "Deep_GroupChange",
				Severity:    "Critical",
				Description: fmt.Sprintf("Kullanıcı Güvenlik Grubuna Eklendi: %s", evt.TargetUserName),
				Status:      "Open",
				CreatedAt:   time.Now(),
			}
			database.DB.Create(&violation)
		}

		// Kural D: Mesai dışı oturum açma (23:00-06:00)
		if evt.EventID == 4624 {
			eventTime, _ := time.Parse(time.RFC3339, evt.TimeCreated)
			hour := eventTime.Hour()
			if hour >= 23 || hour < 6 {
				violation := models.Finding{
					ID:          uuid.New().String(),
					AssetID:     incoming.AssetID,
					CheckType:   "Deep_AfterHoursLogin",
					Severity:    "Medium",
					Description: fmt.Sprintf("Mesai Dışı Oturum Açma: %s (Saat: %02d:%02d)", evt.TargetUserName, hour, eventTime.Minute()),
					Status:      "Open",
					CreatedAt:   time.Now(),
				}
				database.DB.Create(&violation)
			}
		}
	}
}

func processDeepAuditFIM(incoming models.Finding) {
	// 1. Snapshot kaydet
	snapshot := models.ForensicSnapshot{
		ID:        uuid.New().String(),
		AssetID:   incoming.AssetID,
		Type:      "fim",
		Data:      datatypes.JSON(incoming.Description),
		CreatedAt: time.Now(),
	}
	database.DB.Create(&snapshot)

	// 2. Hash verilerini parse et
	var fileHashes []struct {
		Path         string `json:"path"`
		SHA256       string `json:"sha256"`
		SizeBytes    int64  `json:"size_bytes"`
		LastModified string `json:"last_modified"`
	}

	if err := json.Unmarshal([]byte(incoming.Description), &fileHashes); err != nil {
		return
	}

	for _, fh := range fileHashes {
		lastMod, _ := time.Parse(time.RFC3339, fh.LastModified)

		// Mevcut kaydı bul
		var existing models.FileIntegrityRecord
		result := database.DB.Where("asset_id = ? AND file_path = ?", incoming.AssetID, fh.Path).
			Order("created_at desc").First(&existing)

		if result.Error != nil {
			// İlk kayıt — baseline oluştur
			record := models.FileIntegrityRecord{
				ID:           uuid.New().String(),
				AssetID:      incoming.AssetID,
				FilePath:     fh.Path,
				SHA256:       fh.SHA256,
				FileSize:     fh.SizeBytes,
				LastModified: lastMod,
				Status:       "baseline",
				CreatedAt:    time.Now(),
			}
			database.DB.Create(&record)
		} else if existing.SHA256 != fh.SHA256 {
			// Hash değişti — alarm!
			record := models.FileIntegrityRecord{
				ID:           uuid.New().String(),
				AssetID:      incoming.AssetID,
				FilePath:     fh.Path,
				SHA256:       fh.SHA256,
				FileSize:     fh.SizeBytes,
				LastModified: lastMod,
				Status:       "changed",
				PreviousHash: existing.SHA256,
				CreatedAt:    time.Now(),
			}
			database.DB.Create(&record)

			// Finding oluştur
			var count int64
			database.DB.Model(&models.Finding{}).Where(
				"asset_id = ? AND status = ? AND check_type = ? AND description LIKE ?",
				incoming.AssetID, "Open", "Deep_FIM_Violation", "%"+fh.Path+"%",
			).Count(&count)

			if count == 0 {
				violation := models.Finding{
					ID:          uuid.New().String(),
					AssetID:     incoming.AssetID,
					CheckType:   "Deep_FIM_Violation",
					Severity:    "Critical",
					Description: fmt.Sprintf("Dosya Bütünlüğü İhlali: %s — SHA-256 değişti (Eski: %s... → Yeni: %s...)", fh.Path, existing.SHA256[:16], fh.SHA256[:16]),
					Status:      "Open",
					CreatedAt:   time.Now(),
				}
				database.DB.Create(&violation)
			}
		} else {
			// Hash aynı — güncelle
			existing.Status = "unchanged"
			existing.CreatedAt = time.Now()
			database.DB.Save(&existing)
		}
	}
}

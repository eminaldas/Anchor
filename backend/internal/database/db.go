package database

import (
	"log"
	"os"
	"time"

	"audit-backend/internal/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// .env dosyasını yükle (eğer varsa)
	err := godotenv.Load()
	if err != nil {
		log.Println("Uyarı: .env dosyası bulunamadı veya okunamadı. Çevresel değişkenler kullanılacak.")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("HATA: DATABASE_URL çevre değişkeni tanımlanmamış!")
	}
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Veritabanına bağlanılamadı: %v", err)
	}

	DB = db

	log.Println("PostgreSQL bağlantısı (GORM) başarılı!")

	// Migrate the schema
	err = DB.AutoMigrate(
		&models.Finding{},
		&models.Asset{},
		&models.SoftwarePolicy{},
		&models.SoftwareCatalog{},
		&models.Command{},
		&models.ForensicSnapshot{},
		&models.SecurityEvent{},
		&models.FileIntegrityRecord{},
		&models.AdminWhitelist{},
		&models.FIMConfig{},
	)
	if err != nil {
		log.Fatalf("Veritabanı migration hatası: %v", err)
	}
	
	log.Println("Veritabanı migration başarılı!")

	// Varsayılan verileri ekle (sadece boşsa)
	seedDefaults()
}

func seedDefaults() {
	// Varsayılan Admin Whitelist
	defaultAdmins := []models.AdminWhitelist{
		{Username: "Administrator", Reason: "Windows varsayılan admin hesabı", AddedBy: "system"},
	}

	for _, admin := range defaultAdmins {
		var count int64
		DB.Model(&models.AdminWhitelist{}).Where("username = ?", admin.Username).Count(&count)
		if count == 0 {
			admin.CreatedAt = time.Now()
			DB.Create(&admin)
			log.Printf("Varsayılan whitelist eklendi: %s", admin.Username)
		}
	}

	// Varsayılan FIM Yapılandırması
	defaultFIMConfigs := []models.FIMConfig{
		{FilePath: `C:\Windows\System32\drivers\etc\hosts`, Label: "DNS Yapılandırması", Priority: "Critical"},
		{FilePath: `C:\Windows\System32\config\SAM`, Label: "SAM Veritabanı", Priority: "Critical"},
		{FilePath: `C:\Windows\System32\GroupPolicy\Machine\Registry.pol`, Label: "Grup Politikası", Priority: "High"},
	}

	for _, cfg := range defaultFIMConfigs {
		var count int64
		DB.Model(&models.FIMConfig{}).Where("file_path = ?", cfg.FilePath).Count(&count)
		if count == 0 {
			cfg.CreatedAt = time.Now()
			DB.Create(&cfg)
			log.Printf("Varsayılan FIM dosyası eklendi: %s", cfg.Label)
		}
	}
}
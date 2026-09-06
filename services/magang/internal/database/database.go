package database

import (
	"fmt"
	"log"
	"time"

	"magang-be/services/magang/config"
	"magang-be/services/magang/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() {
	DB = New(config.App.Database)
	autoMigrate(DB)
}

func New(cfg config.DatabaseConfig) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
		sslMode(cfg.SSLMode), "Asia/Jakarta",
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect postgres [%s@%s:%d/%s]: %v",
			cfg.User, cfg.Host, cfg.Port, cfg.DBName, err)
	}

	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
		sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	}

	log.Printf("connected to postgres: %s@%s:%d/%s", cfg.User, cfg.Host, cfg.Port, cfg.DBName)
	return db
}

// autoMigrate otomatis menambah/mengupdate kolom yang belum ada di DB
func autoMigrate(db *gorm.DB) {
	// Pastikan schema magang ada
	db.Exec("CREATE SCHEMA IF NOT EXISTS magang")

	// Pastikan kolom-kolom baru ditambahkan ke tabel magang.users jika belum ada
	db.Exec("ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS no_emp VARCHAR(50)")
	db.Exec("ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS id_emp VARCHAR(50)")
	db.Exec("ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS email VARCHAR(255)")
	db.Exec("ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS title VARCHAR(150)")
	db.Exec("ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS company VARCHAR(255)")
	db.Exec("ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS section VARCHAR(100)")
	db.Exec("ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS department VARCHAR(100)")
	db.Exec("ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS division VARCHAR(100)")
	db.Exec(`DO $$
	BEGIN
		IF NOT EXISTS (
			SELECT 1 FROM pg_constraint WHERE conname = 'users_role_check'
		) THEN
			ALTER TABLE magang.users
			ADD CONSTRAINT users_role_check
			CHECK (role IN ('admin', 'user', 'approval', 'external_audit'));
		END IF;
	END $$`)

	if err := db.AutoMigrate(&models.User{}, &models.Item{}, &models.ChangeNote{}); err != nil {
		log.Printf("auto migrate note: %v", err)
	}
	log.Println("database migration completed")
}

func sslMode(v string) string {
	if v == "" {
		return "disable"
	}
	return v
}

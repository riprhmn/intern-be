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
	db.Exec("ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS code_name VARCHAR(50)")
	db.Exec("ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS email VARCHAR(255)")
	db.Exec("ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS title VARCHAR(150)")
	db.Exec("ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS company VARCHAR(255)")
	db.Exec("ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS section VARCHAR(100)")
	db.Exec("ALTER TABLE magang.users ADD COLUMN IF NOT EXISTS department VARCHAR(100)")
	db.Exec(`
		CREATE TABLE IF NOT EXISTS magang.change_notes (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			initiator VARCHAR(255) NOT NULL,
			emp_id_initiator VARCHAR(50),
			department VARCHAR(100),
			change_pertains_to TEXT NOT NULL,
			document_name VARCHAR(255),
			document_data TEXT,
			status VARCHAR(50) NOT NULL DEFAULT 'SUBMITTED',
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);
	`)

	if err := db.AutoMigrate(&models.User{}, &models.Item{}, &models.ChangeNote{}, &models.CNApprovalConfig{}, &models.CNNumberCounter{}, &models.DocumentTemplate{}, &models.ANCRRequest{}, &models.ApprovalMapping{}, &models.ApprovalDelegation{}); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	db.Exec("ALTER TABLE magang.change_notes ADD COLUMN IF NOT EXISTS document_name VARCHAR(255)")
	db.Exec("ALTER TABLE magang.change_notes ADD COLUMN IF NOT EXISTS document_data TEXT")
	// Normalize the MQD approval-stage label without discarding existing assignments.
	db.Exec(`UPDATE magang.cn_approval_configs SET stages = REPLACE(stages, 'MOD Manager', 'MQD Manager') WHERE stages LIKE '%MOD Manager%'`)
	db.Exec(`UPDATE magang.change_notes SET stages = REPLACE(stages, 'MOD Manager', 'MQD Manager') WHERE stages LIKE '%MOD Manager%'`)
	db.Exec(`UPDATE magang.cn_approval_configs SET stages = REPLACE(stages, '"label":"MQD"', '"label":"MQD Manager"') WHERE stages LIKE '%"label":"MQD"%'`)
	db.Exec(`UPDATE magang.change_notes SET stages = REPLACE(stages, '"label":"MQD"', '"label":"MQD Manager"') WHERE stages LIKE '%"label":"MQD"%'`)
	// Change Note Number is a system-generated numeric identifier, never a file name.
	db.Exec(`UPDATE magang.change_notes SET document_code = EXTRACT(YEAR FROM created_at)::INTEGER::TEXT || '-' || id::TEXT WHERE document_code <> '' AND document_code !~ '^[0-9]{4}-[0-9]+$'`)
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_change_notes_document_code_unique ON magang.change_notes (document_code) WHERE document_code <> ''`)
	if err := migrateChangeNoteNumbers(db); err != nil {
		log.Fatalf("change note number migration failed: %v", err)
	}
	db.Exec(`UPDATE magang.users SET title = 'Manager' WHERE LOWER(TRIM(title)) = 'concern manager'`)
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_code_name_unique ON magang.users (LOWER(BTRIM(code_name))) WHERE BTRIM(COALESCE(code_name, '')) <> ''`)
	db.Exec(`ALTER TABLE magang.approval_mappings ALTER COLUMN sequence DROP NOT NULL`)
	db.Exec(`ALTER TABLE magang.approval_mappings ALTER COLUMN user_id DROP NOT NULL`)
	db.Exec(`ALTER TABLE magang.approval_mappings ADD COLUMN IF NOT EXISTS seq_1 VARCHAR(50)`)
	db.Exec(`ALTER TABLE magang.approval_mappings ADD COLUMN IF NOT EXISTS seq_2 VARCHAR(50)`)
	db.Exec(`ALTER TABLE magang.approval_mappings ADD COLUMN IF NOT EXISTS seq_3 VARCHAR(50)`)
	db.Exec(`ALTER TABLE magang.approval_mappings ADD COLUMN IF NOT EXISTS seq_4 VARCHAR(50)`)
	db.Exec(`ALTER TABLE magang.approval_mappings DROP CONSTRAINT IF EXISTS approval_mappings_scope_check`)
	db.Exec(`DROP INDEX IF EXISTS magang.idx_approval_mapping_entry_unique`)
	db.Exec(`DROP INDEX IF EXISTS magang.idx_approval_mapping_cn_sequence_unique`)

	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_approval_delegation_active_scope ON magang.approval_delegations (from_user_id, approval_type, section_code) WHERE is_active = true`)
	db.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'approval_delegations_scope_check') THEN
			ALTER TABLE magang.approval_delegations ADD CONSTRAINT approval_delegations_scope_check
			CHECK (approval_type IN ('ALL', 'CN', 'ANCR') AND BTRIM(section_code) <> '' AND from_user_id <> to_user_id AND (starts_at IS NULL OR ends_at IS NULL OR ends_at > starts_at));
		END IF;
	END $$`)
	log.Println("database migration completed")
}

func migrateChangeNoteNumbers(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Remove the legacy non-partial index: empty values must remain valid while
		// old rows are being backfilled during the first deployment.
		if err := tx.Exec(`DROP INDEX IF EXISTS magang.idx_magang_change_notes_registration_number`).Error; err != nil {
			return err
		}

		var initialized int64
		if err := tx.Model(&models.CNNumberCounter{}).Count(&initialized).Error; err != nil {
			return err
		}

		if initialized == 0 {
			if err := tx.Exec(`
				WITH numbered AS (
					SELECT id,
						EXTRACT(YEAR FROM created_at)::INTEGER AS issue_year,
						ROW_NUMBER() OVER (
							PARTITION BY EXTRACT(YEAR FROM created_at)::INTEGER
							ORDER BY created_at, id
						) AS issue_number
					FROM magang.change_notes
				)
				UPDATE magang.change_notes AS cn
				SET registration_number = 'CN-' || numbered.issue_year || '-' || LPAD(numbered.issue_number::TEXT, 3, '0')
				FROM numbered
				WHERE cn.id = numbered.id
			`).Error; err != nil {
				return err
			}

			// Move existing values out of the final namespace before resequencing,
			// avoiding transient collisions with the existing unique index.
			if err := tx.Exec(`
				UPDATE magang.change_notes
				SET document_code = '__cn_resequence__' || id
				WHERE document_code <> ''
			`).Error; err != nil {
				return err
			}
			if err := tx.Exec(`
				WITH numbered AS (
					SELECT id,
						EXTRACT(YEAR FROM created_at)::INTEGER AS issue_year,
						ROW_NUMBER() OVER (
							PARTITION BY EXTRACT(YEAR FROM created_at)::INTEGER
							ORDER BY created_at, id
						) AS issue_number
					FROM magang.change_notes
					WHERE document_code LIKE '__cn_resequence__%'
				)
				UPDATE magang.change_notes AS cn
				SET document_code = numbered.issue_year || '-' || numbered.issue_number
				FROM numbered
				WHERE cn.id = numbered.id
			`).Error; err != nil {
				return err
			}
		}

		if err := tx.Exec(`
			INSERT INTO magang.cn_number_counters (kind, year, last_number)
			SELECT 'registration', EXTRACT(YEAR FROM created_at)::INTEGER, COUNT(*)
			FROM magang.change_notes
			GROUP BY EXTRACT(YEAR FROM created_at)::INTEGER
			ON CONFLICT (kind, year) DO UPDATE
			SET last_number = GREATEST(magang.cn_number_counters.last_number, EXCLUDED.last_number)
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			INSERT INTO magang.cn_number_counters (kind, year, last_number)
			SELECT 'change_note', EXTRACT(YEAR FROM created_at)::INTEGER, COUNT(*)
			FROM magang.change_notes
			WHERE document_code <> ''
			GROUP BY EXTRACT(YEAR FROM created_at)::INTEGER
			ON CONFLICT (kind, year) DO UPDATE
			SET last_number = GREATEST(magang.cn_number_counters.last_number, EXCLUDED.last_number)
		`).Error; err != nil {
			return err
		}
		return tx.Exec(`
			CREATE UNIQUE INDEX IF NOT EXISTS idx_change_notes_registration_number_unique
			ON magang.change_notes (registration_number)
			WHERE registration_number <> ''
		`).Error
	})
}

func sslMode(v string) string {
	if v == "" {
		return "disable"
	}
	return v
}

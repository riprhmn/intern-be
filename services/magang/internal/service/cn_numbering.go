package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

const (
	cnNumberKindRegistration = "registration"
	cnNumberKindChangeNote   = "change_note"
)

func cnYear(value time.Time) int {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return value.Year()
	}
	return value.In(location).Year()
}

func nextCNSerial(tx *gorm.DB, kind string, year int) (uint64, error) {
	if err := tx.Exec(`
		INSERT INTO magang.cn_number_counters (kind, year, last_number)
		VALUES (?, ?, 0)
		ON CONFLICT (kind, year) DO NOTHING
	`, kind, year).Error; err != nil {
		return 0, err
	}

	var result struct {
		LastNumber uint64 `gorm:"column:last_number"`
	}
	if err := tx.Raw(`
		UPDATE magang.cn_number_counters
		SET last_number = last_number + 1
		WHERE kind = ? AND year = ?
		RETURNING last_number
	`, kind, year).Scan(&result).Error; err != nil {
		return 0, err
	}
	if result.LastNumber == 0 {
		return 0, fmt.Errorf("counter nomor %s tahun %d tidak dapat dialokasikan", kind, year)
	}
	return result.LastNumber, nil
}

func formatCNRegistrationNumber(year int, number uint64) string {
	return fmt.Sprintf("CN-%d-%03d", year, number)
}

func formatChangeNoteNumber(year int, number uint64) string {
	return fmt.Sprintf("%d-%d", year, number)
}

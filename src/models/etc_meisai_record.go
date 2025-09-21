package models

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ETCMeisaiRecord represents an ETC toll record with comprehensive validation
type ETCMeisaiRecord struct {
	ID              int64              `gorm:"primaryKey;autoIncrement" json:"id"`
	Hash            string             `gorm:"uniqueIndex;size:64;not null" json:"hash"`
	Date            time.Time          `gorm:"index;not null" json:"date"`
	Time            string             `gorm:"size:8;not null" json:"time"`
	EntranceIC      string             `gorm:"size:100;not null" json:"entrance_ic"`
	ExitIC          string             `gorm:"size:100;not null" json:"exit_ic"`
	TollAmount      int                `gorm:"not null" json:"toll_amount"`
	CarNumber       string             `gorm:"index;size:20;not null" json:"car_number"`
	ETCCardNumber   string             `gorm:"index;size:20;not null" json:"etc_card_number"`
	ETCNum          *string            `gorm:"index;size:50" json:"etc_num,omitempty"`
	DtakoRowID      *int64             `gorm:"index" json:"dtako_row_id,omitempty"`
	CreatedAt       time.Time          `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time          `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt     `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName returns the table name for GORM
func (ETCMeisaiRecord) TableName() string {
	return "etc_meisai_records"
}

// BeforeCreate hook to generate hash before creating record
func (r *ETCMeisaiRecord) BeforeCreate(tx *gorm.DB) error {
	if err := r.validate(); err != nil {
		return err
	}

	if r.Hash == "" {
		r.Hash = r.generateHash()
	}

	return nil
}

// BeforeSave hook to validate data before saving
func (r *ETCMeisaiRecord) BeforeSave(tx *gorm.DB) error {
	return r.validate()
}

// validate performs comprehensive validation of the record
func (r *ETCMeisaiRecord) validate() error {
	// Validate date (not in future)
	if r.Date.After(time.Now()) {
		return fmt.Errorf("date cannot be in the future")
	}

	// Validate time format (HH:MM:SS)
	timeRegex := regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$`)
	if !timeRegex.MatchString(r.Time) {
		return fmt.Errorf("time must be in HH:MM:SS format")
	}

	// Validate entrance and exit IC names
	if len(strings.TrimSpace(r.EntranceIC)) == 0 {
		return fmt.Errorf("entrance IC cannot be empty")
	}
	if len(strings.TrimSpace(r.ExitIC)) == 0 {
		return fmt.Errorf("exit IC cannot be empty")
	}
	if len(r.EntranceIC) > 100 {
		return fmt.Errorf("entrance IC name too long (max 100 characters)")
	}
	if len(r.ExitIC) > 100 {
		return fmt.Errorf("exit IC name too long (max 100 characters)")
	}

	// Validate toll amount
	if r.TollAmount < 0 {
		return fmt.Errorf("toll amount must be non-negative")
	}
	if r.TollAmount > 999999 {
		return fmt.Errorf("toll amount too large (max 999999)")
	}

	// Validate car number (Japanese vehicle number format)
	if err := r.validateCarNumber(); err != nil {
		return err
	}

	// Validate ETC card number (16-19 digits)
	if err := r.validateETCCardNumber(); err != nil {
		return err
	}

	// Validate ETC number if provided
	if r.ETCNum != nil && *r.ETCNum != "" {
		if err := r.validateETCNum(); err != nil {
			return err
		}
	}

	return nil
}

// validateCarNumber validates Japanese vehicle number format
func (r *ETCMeisaiRecord) validateCarNumber() error {
	if len(strings.TrimSpace(r.CarNumber)) == 0 {
		return fmt.Errorf("car number cannot be empty")
	}

	// Japanese vehicle number patterns
	patterns := []string{
		`^\d{3}-\d{2}$`,                    // 軽自動車: 123-45
		`^\d{3}\s\d{2}$`,                   // 軽自動車: 123 45
		`^[あ-ん]{1}\d{3}$`,                // ひらがな + 数字: あ123
		`^[ア-ン]{1}\d{3}$`,                // カタカナ + 数字: ア123
		`^\d{2}-\d{2}$`,                    // 二輪: 12-34
		`^\d{4}$`,                          // 4桁数字
		`^[a-zA-Z0-9\-\s]{3,20}$`,         // 一般的なパターン
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, r.CarNumber); matched {
			return nil
		}
	}

	return fmt.Errorf("invalid car number format")
}

// validateETCCardNumber validates ETC card number format
func (r *ETCMeisaiRecord) validateETCCardNumber() error {
	if len(strings.TrimSpace(r.ETCCardNumber)) == 0 {
		return fmt.Errorf("ETC card number cannot be empty")
	}

	// Remove spaces and hyphens for validation
	cleaned := strings.ReplaceAll(strings.ReplaceAll(r.ETCCardNumber, " ", ""), "-", "")

	// Check if it's 16-19 digits
	if len(cleaned) < 16 || len(cleaned) > 19 {
		return fmt.Errorf("ETC card number must be 16-19 digits")
	}

	// Check if all characters are digits
	if _, err := strconv.ParseInt(cleaned, 10, 64); err != nil {
		return fmt.Errorf("ETC card number must contain only digits")
	}

	return nil
}

// validateETCNum validates ETC 2.0 device number format
func (r *ETCMeisaiRecord) validateETCNum() error {
	if r.ETCNum == nil || *r.ETCNum == "" {
		return nil // Optional field
	}

	// ETC 2.0 device number is typically alphanumeric, 10-20 characters
	etcNum := strings.TrimSpace(*r.ETCNum)
	if len(etcNum) < 5 || len(etcNum) > 50 {
		return fmt.Errorf("ETC number must be 5-50 characters")
	}

	// Allow alphanumeric characters and some special characters
	pattern := `^[a-zA-Z0-9\-_]+$`
	if matched, _ := regexp.MatchString(pattern, etcNum); !matched {
		return fmt.Errorf("ETC number contains invalid characters")
	}

	return nil
}

// generateHash creates a SHA256 hash for the record
func (r *ETCMeisaiRecord) generateHash() string {
	data := fmt.Sprintf("%s|%s|%s|%s|%d|%s|%s",
		r.Date.Format("2006-01-02"),
		r.Time,
		r.EntranceIC,
		r.ExitIC,
		r.TollAmount,
		r.CarNumber,
		r.ETCCardNumber,
	)

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// GetDateString returns the date in YYYY-MM-DD format
func (r *ETCMeisaiRecord) GetDateString() string {
	return r.Date.Format("2006-01-02")
}

// GetMaskedETCCardNumber returns a masked version of the ETC card number
func (r *ETCMeisaiRecord) GetMaskedETCCardNumber() string {
	if len(r.ETCCardNumber) <= 4 {
		return "****"
	}

	// Show only last 4 digits
	return "****-****-****-" + r.ETCCardNumber[len(r.ETCCardNumber)-4:]
}

// IsValidForMapping checks if the record has required fields for mapping
func (r *ETCMeisaiRecord) IsValidForMapping() bool {
	return r.ID > 0 && r.Hash != "" && !r.Date.IsZero()
}
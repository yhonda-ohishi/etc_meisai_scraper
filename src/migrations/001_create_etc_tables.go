package migrations

import (
	"fmt"
	"log"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"gorm.io/gorm"
)

// Migration001CreateETCTables represents the initial migration to create ETC-related tables
type Migration001CreateETCTables struct{}

// ID returns the migration identifier
func (m *Migration001CreateETCTables) ID() string {
	return "001_create_etc_tables"
}

// Description returns the migration description
func (m *Migration001CreateETCTables) Description() string {
	return "Create ETC meisai records, mappings, and import sessions tables"
}

// Up executes the migration
func (m *Migration001CreateETCTables) Up(db *gorm.DB) error {
	log.Printf("Running migration: %s - %s", m.ID(), m.Description())

	// Create tables using GORM AutoMigrate
	if err := db.AutoMigrate(
		&models.ETCMeisaiRecord{},
		&models.ETCMapping{},
		&models.ImportSession{},
	); err != nil {
		return fmt.Errorf("failed to auto-migrate tables: %w", err)
	}

	// Add composite indexes for performance optimization
	if err := m.createCompositeIndexes(db); err != nil {
		return fmt.Errorf("failed to create composite indexes: %w", err)
	}

	// Add foreign key constraints
	if err := m.createForeignKeys(db); err != nil {
		return fmt.Errorf("failed to create foreign keys: %w", err)
	}

	// Add check constraints
	if err := m.createCheckConstraints(db); err != nil {
		return fmt.Errorf("failed to create check constraints: %w", err)
	}

	log.Printf("Successfully completed migration: %s", m.ID())
	return nil
}

// Down reverts the migration
func (m *Migration001CreateETCTables) Down(db *gorm.DB) error {
	log.Printf("Reverting migration: %s", m.ID())

	// Drop tables in reverse order to respect foreign key constraints
	if err := db.Migrator().DropTable(&models.ETCMapping{}); err != nil {
		return fmt.Errorf("failed to drop etc_mappings table: %w", err)
	}

	if err := db.Migrator().DropTable(&models.ETCMeisaiRecord{}); err != nil {
		return fmt.Errorf("failed to drop etc_meisai_records table: %w", err)
	}

	if err := db.Migrator().DropTable(&models.ImportSession{}); err != nil {
		return fmt.Errorf("failed to drop import_sessions table: %w", err)
	}

	log.Printf("Successfully reverted migration: %s", m.ID())
	return nil
}

// createCompositeIndexes creates performance-optimized composite indexes
func (m *Migration001CreateETCTables) createCompositeIndexes(db *gorm.DB) error {
	indexes := []struct {
		table string
		name  string
		sql   string
	}{
		{
			table: "etc_meisai_records",
			name:  "idx_etc_date_car",
			sql:   "CREATE INDEX IF NOT EXISTS idx_etc_date_car ON etc_meisai_records(date, car_number)",
		},
		{
			table: "etc_meisai_records",
			name:  "idx_etc_card_date",
			sql:   "CREATE INDEX IF NOT EXISTS idx_etc_card_date ON etc_meisai_records(etc_card_number, date)",
		},
		{
			table: "etc_mappings",
			name:  "idx_mapping_entity",
			sql:   "CREATE INDEX IF NOT EXISTS idx_mapping_entity ON etc_mappings(mapped_entity_type, mapped_entity_id)",
		},
		{
			table: "etc_mappings",
			name:  "idx_mapping_status_type",
			sql:   "CREATE INDEX IF NOT EXISTS idx_mapping_status_type ON etc_mappings(status, mapping_type)",
		},
		{
			table: "import_sessions",
			name:  "idx_import_account",
			sql:   "CREATE INDEX IF NOT EXISTS idx_import_account ON import_sessions(account_type, account_id, status)",
		},
		{
			table: "import_sessions",
			name:  "idx_import_status_date",
			sql:   "CREATE INDEX IF NOT EXISTS idx_import_status_date ON import_sessions(status, created_at)",
		},
	}

	for _, idx := range indexes {
		if err := db.Exec(idx.sql).Error; err != nil {
			return fmt.Errorf("failed to create index %s on table %s: %w", idx.name, idx.table, err)
		}
		log.Printf("Created index: %s on table: %s", idx.name, idx.table)
	}

	return nil
}

// createForeignKeys creates foreign key constraints
func (m *Migration001CreateETCTables) createForeignKeys(db *gorm.DB) error {
	// Check if we're using a database that supports foreign keys
	dialect := db.Dialector.Name()
	if dialect == "sqlite3" || dialect == "sqlite" {
		// SQLite supports foreign keys but they need to be enabled
		if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
			return fmt.Errorf("failed to enable foreign keys in SQLite: %w", err)
		}
	}

	// Add foreign key constraint for etc_mappings -> etc_meisai_records
	fkSQL := `
		ALTER TABLE etc_mappings
		ADD CONSTRAINT fk_etc_record
		FOREIGN KEY (etc_record_id)
		REFERENCES etc_meisai_records(id)
		ON DELETE CASCADE
	`

	// For MySQL and PostgreSQL, add the constraint if it doesn't exist
	if dialect == "mysql" {
		// Check if foreign key already exists
		checkSQL := `
			SELECT COUNT(*) as count
			FROM information_schema.KEY_COLUMN_USAGE
			WHERE CONSTRAINT_NAME = 'fk_etc_record'
			AND TABLE_SCHEMA = DATABASE()
		`
		var count int64
		if err := db.Raw(checkSQL).Scan(&count).Error; err != nil {
			return fmt.Errorf("failed to check existing foreign key: %w", err)
		}

		if count == 0 {
			if err := db.Exec(fkSQL).Error; err != nil {
				return fmt.Errorf("failed to create foreign key constraint: %w", err)
			}
			log.Printf("Created foreign key constraint: fk_etc_record")
		}
	} else if dialect != "sqlite3" && dialect != "sqlite" {
		// For other databases (PostgreSQL, etc.)
		if err := db.Exec(fkSQL).Error; err != nil {
			// Log the error but don't fail the migration if FK already exists
			log.Printf("Warning: failed to create foreign key constraint (may already exist): %v", err)
		} else {
			log.Printf("Created foreign key constraint: fk_etc_record")
		}
	}

	return nil
}

// createCheckConstraints creates check constraints for data validation
func (m *Migration001CreateETCTables) createCheckConstraints(db *gorm.DB) error {
	dialect := db.Dialector.Name()

	// Check constraints are not supported by all databases in the same way
	constraints := []struct {
		table string
		name  string
		sql   string
	}{
		{
			table: "etc_meisai_records",
			name:  "chk_toll_amount",
			sql:   "ALTER TABLE etc_meisai_records ADD CONSTRAINT chk_toll_amount CHECK (toll_amount >= 0)",
		},
		{
			table: "etc_mappings",
			name:  "chk_confidence",
			sql:   "ALTER TABLE etc_mappings ADD CONSTRAINT chk_confidence CHECK (confidence >= 0 AND confidence <= 1)",
		},
		{
			table: "import_sessions",
			name:  "chk_file_size",
			sql:   "ALTER TABLE import_sessions ADD CONSTRAINT chk_file_size CHECK (file_size > 0)",
		},
		{
			table: "import_sessions",
			name:  "chk_rows",
			sql: `ALTER TABLE import_sessions ADD CONSTRAINT chk_rows CHECK (
				processed_rows >= 0 AND
				success_rows >= 0 AND
				error_rows >= 0 AND
				duplicate_rows >= 0 AND
				processed_rows = success_rows + error_rows + duplicate_rows
			)`,
		},
	}

	for _, constraint := range constraints {
		if dialect == "mysql" || dialect == "postgres" || dialect == "postgresql" {
			// Check if constraint already exists for MySQL
			if dialect == "mysql" {
				checkSQL := `
					SELECT COUNT(*) as count
					FROM information_schema.TABLE_CONSTRAINTS
					WHERE CONSTRAINT_NAME = ?
					AND TABLE_SCHEMA = DATABASE()
				`
				var count int64
				if err := db.Raw(checkSQL, constraint.name).Scan(&count).Error; err != nil {
					log.Printf("Warning: failed to check existing constraint %s: %v", constraint.name, err)
					continue
				}

				if count > 0 {
					log.Printf("Constraint %s already exists, skipping", constraint.name)
					continue
				}
			}

			if err := db.Exec(constraint.sql).Error; err != nil {
				// Log the error but don't fail the migration
				log.Printf("Warning: failed to create check constraint %s on table %s: %v", constraint.name, constraint.table, err)
			} else {
				log.Printf("Created check constraint: %s on table: %s", constraint.name, constraint.table)
			}
		} else {
			log.Printf("Skipping check constraint %s for database type: %s", constraint.name, dialect)
		}
	}

	return nil
}

// Migration interface implementation
type Migration interface {
	ID() string
	Description() string
	Up(db *gorm.DB) error
	Down(db *gorm.DB) error
}

// GetMigration001 returns the migration instance
func GetMigration001() Migration {
	return &Migration001CreateETCTables{}
}
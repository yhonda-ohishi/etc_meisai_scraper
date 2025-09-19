package repositories

import (
	"database/sql"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// NewETCDtakoMappingRepository creates a new mapping repository
func NewETCDtakoMappingRepository(db *sql.DB) *ETCDtakoMappingRepository {
	return &ETCDtakoMappingRepository{db: db}
}

// ETCDtakoMappingRepository handles mapping operations
type ETCDtakoMappingRepository struct {
	db *sql.DB
}

// CreateMapping creates a new mapping
func (r *ETCDtakoMappingRepository) CreateMapping(mapping *models.ETCDtakoMapping) error {
	query := `
		INSERT INTO etc_dtako_mapping (etc_meisai_id, dtako_row_id, vehicle_id, mapping_type, notes, created_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, mapping.ETCMeisaiID, mapping.DtakoRowID, mapping.VehicleID, mapping.MappingType, mapping.Notes, mapping.CreatedBy)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	mapping.ID = id
	return nil
}

// GetMappingByETCMeisaiID gets mapping by ETC meisai ID
func (r *ETCDtakoMappingRepository) GetMappingByETCMeisaiID(etcMeisaiID int64) (*models.ETCDtakoMapping, error) {
	query := `
		SELECT id, etc_meisai_id, dtako_row_id, vehicle_id, mapping_type, notes, created_at, updated_at, created_by
		FROM etc_dtako_mapping
		WHERE etc_meisai_id = ?
		LIMIT 1
	`
	var mapping models.ETCDtakoMapping
	err := r.db.QueryRow(query, etcMeisaiID).Scan(
		&mapping.ID, &mapping.ETCMeisaiID, &mapping.DtakoRowID, &mapping.VehicleID,
		&mapping.MappingType, &mapping.Notes, &mapping.CreatedAt, &mapping.UpdatedAt, &mapping.CreatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}
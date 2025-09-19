package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// ETCRepository handles database operations for ETC meisai
type ETCRepository struct {
	db *sql.DB
}

// NewETCRepository creates a new ETC repository
func NewETCRepository(db *sql.DB) *ETCRepository {
	return &ETCRepository{db: db}
}

// GetByID retrieves an ETC record by row ID
func (r *ETCRepository) GetByID(rowID string) (*models.ETCRow, error) {
	query := `
		SELECT row_id, date, time, entry_ic, exit_ic,
		       card_number, amount, created_at
		FROM etc_rows
		WHERE row_id = ?
	`

	var row models.ETCRow
	err := r.db.QueryRow(query, rowID).Scan(
		&row.RowID,
		&row.Date,
		&row.Time,
		&row.EntryIC,
		&row.ExitIC,
		&row.CardNumber,
		&row.Amount,
		&row.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ETC row by ID: %w", err)
	}

	return &row, nil
}

// CountByDateRange counts ETC records in a date range
func (r *ETCRepository) CountByDateRange(startDate, endDate time.Time) (int, error) {
	query := `
		SELECT COUNT(*) FROM etc_rows
		WHERE date >= ? AND date <= ?
	`

	var count int
	err := r.db.QueryRow(query, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count ETC records: %w", err)
	}

	return count, nil
}

// GetByDateRangeBatch retrieves ETC records in batches
func (r *ETCRepository) GetByDateRangeBatch(startDate, endDate time.Time, limit, offset int) ([]models.ETCRow, error) {
	query := `
		SELECT row_id, date, time, entry_ic, exit_ic,
		       card_number, amount, created_at
		FROM etc_rows
		WHERE date >= ? AND date <= ?
		ORDER BY date, time
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query ETC rows: %w", err)
	}
	defer rows.Close()

	var etcRows []models.ETCRow
	for rows.Next() {
		var row models.ETCRow
		err := rows.Scan(
			&row.RowID,
			&row.Date,
			&row.Time,
			&row.EntryIC,
			&row.ExitIC,
			&row.CardNumber,
			&row.Amount,
			&row.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ETC row: %w", err)
		}
		etcRows = append(etcRows, row)
	}

	return etcRows, nil
}

// GetUnmappedRecords retrieves ETC records without mappings
func (r *ETCRepository) GetUnmappedETCRecords(startDate, endDate time.Time, limit, offset int) ([]*models.ETCRow, error) {
	query := `
		SELECT e.row_id, e.date, e.time, e.entry_ic, e.exit_ic,
		       e.card_number, e.amount, e.created_at
		FROM etc_rows e
		LEFT JOIN etc_dtako_mapping m ON e.row_id = m.etc_row_id
		WHERE m.id IS NULL
		  AND e.date >= ? AND e.date <= ?
		ORDER BY e.date, e.time
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query unmapped ETC rows: %w", err)
	}
	defer rows.Close()

	var etcRows []*models.ETCRow
	for rows.Next() {
		var row models.ETCRow
		err := rows.Scan(
			&row.RowID,
			&row.Date,
			&row.Time,
			&row.EntryIC,
			&row.ExitIC,
			&row.CardNumber,
			&row.Amount,
			&row.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan unmapped ETC row: %w", err)
		}
		etcRows = append(etcRows, &row)
	}

	return etcRows, nil
}

// GetByDateRange retrieves ETC meisai records within a date range
func (r *ETCRepository) GetByDateRange(fromDate, toDate time.Time) ([]models.ETCMeisai, error) {
	query := `
		SELECT id, unko_no, date, time, ic_entry, ic_exit, vehicle_no,
		       card_no, amount, discount_amount, total_amount, usage_type,
		       payment_method, route_code, distance, created_at, updated_at
		FROM etc_meisai
		WHERE date >= ? AND date <= ?
		ORDER BY date, time
	`

	rows, err := r.db.Query(query, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query ETC meisai: %w", err)
	}
	defer rows.Close()

	var results []models.ETCMeisai
	for rows.Next() {
		var m models.ETCMeisai
		err := rows.Scan(
			&m.ID, &m.UnkoNo, &m.Date, &m.Time, &m.ICEntry, &m.ICExit,
			&m.VehicleNo, &m.CardNo, &m.Amount, &m.DiscountAmount,
			&m.TotalAmount, &m.UsageType, &m.PaymentMethod, &m.RouteCode,
			&m.Distance, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, m)
	}

	return results, nil
}

// GetByUnkoNo retrieves ETC meisai records by unko_no
func (r *ETCRepository) GetByUnkoNo(unkoNo string) ([]models.ETCMeisai, error) {
	query := `
		SELECT id, unko_no, date, time, ic_entry, ic_exit, vehicle_no,
		       card_no, amount, discount_amount, total_amount, usage_type,
		       payment_method, route_code, distance, created_at, updated_at
		FROM etc_meisai
		WHERE unko_no = ?
		ORDER BY date, time
	`

	rows, err := r.db.Query(query, unkoNo)
	if err != nil {
		return nil, fmt.Errorf("failed to query ETC meisai by unko_no: %w", err)
	}
	defer rows.Close()

	var results []models.ETCMeisai
	for rows.Next() {
		var m models.ETCMeisai
		err := rows.Scan(
			&m.ID, &m.UnkoNo, &m.Date, &m.Time, &m.ICEntry, &m.ICExit,
			&m.VehicleNo, &m.CardNo, &m.Amount, &m.DiscountAmount,
			&m.TotalAmount, &m.UsageType, &m.PaymentMethod, &m.RouteCode,
			&m.Distance, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, m)
	}

	return results, nil
}

// Insert creates a new ETC meisai record
func (r *ETCRepository) Insert(m *models.ETCMeisai) error {
	query := `
		INSERT INTO etc_meisai (
			unko_no, date, time, ic_entry, ic_exit, vehicle_no,
			card_no, amount, discount_amount, total_amount, usage_type,
			payment_method, route_code, distance, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	result, err := r.db.Exec(query,
		m.UnkoNo, m.Date, m.Time, m.ICEntry, m.ICExit, m.VehicleNo,
		m.CardNo, m.Amount, m.DiscountAmount, m.TotalAmount, m.UsageType,
		m.PaymentMethod, m.RouteCode, m.Distance,
	)
	if err != nil {
		return fmt.Errorf("failed to insert ETC meisai: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}
	m.ID = id

	return nil
}

// BulkInsert inserts multiple ETC meisai records
func (r *ETCRepository) BulkInsert(records []models.ETCMeisai) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO etc_meisai (
			unko_no, date, time, ic_entry, ic_exit, vehicle_no,
			card_no, amount, discount_amount, total_amount, usage_type,
			payment_method, route_code, distance, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, m := range records {
		_, err := stmt.Exec(
			m.UnkoNo, m.Date, m.Time, m.ICEntry, m.ICExit, m.VehicleNo,
			m.CardNo, m.Amount, m.DiscountAmount, m.TotalAmount, m.UsageType,
			m.PaymentMethod, m.RouteCode, m.Distance,
		)
		if err != nil {
			return fmt.Errorf("failed to insert record: %w", err)
		}
	}

	return tx.Commit()
}

// GetSummaryByDateRange gets summary statistics for a date range
func (r *ETCRepository) GetSummaryByDateRange(fromDate, toDate time.Time) ([]models.ETCSummary, error) {
	query := `
		SELECT date, vehicle_no,
		       SUM(total_amount) as total_amount,
		       COUNT(*) as total_count,
		       SUM(distance) as total_distance
		FROM etc_meisai
		WHERE date >= ? AND date <= ?
		GROUP BY date, vehicle_no
		ORDER BY date, vehicle_no
	`

	rows, err := r.db.Query(query, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query summary: %w", err)
	}
	defer rows.Close()

	var results []models.ETCSummary
	for rows.Next() {
		var s models.ETCSummary
		err := rows.Scan(&s.Date, &s.VehicleNo, &s.TotalAmount, &s.TotalCount, &s.TotalDistance)
		if err != nil {
			return nil, fmt.Errorf("failed to scan summary row: %w", err)
		}
		results = append(results, s)
	}

	return results, nil
}
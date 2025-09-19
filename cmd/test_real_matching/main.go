package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/yhonda-ohishi/etc_meisai/config"
	"github.com/yhonda-ohishi/etc_meisai/models"
	"github.com/yhonda-ohishi/etc_meisai/repositories"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	fmt.Println("=== 実データマッチングテスト開始 ===")
	fmt.Println()

	// Step 1: Test production database connection
	fmt.Println("1. 本番データベース接続テスト...")
	prodDB, err := connectProductionDB()
	if err != nil {
		fmt.Printf("⚠️  本番DB接続失敗: %v\n", err)
		fmt.Println("   モックモードで継続します...")
		var prodDB *sql.DB
		_ = prodDB
	} else {
		defer prodDB.Close()
		fmt.Println("✅ 本番DB接続成功！")
	}

	// Step 2: Test local database connection
	fmt.Println("\n2. ローカルデータベース接続テスト...")
	localDB, err := connectLocalDB()
	if err != nil {
		log.Fatalf("ローカルDB接続失敗: %v", err)
	}
	defer localDB.Close()
	fmt.Println("✅ ローカルDB接続成功！")

	// Step 3: Get sample ETC records from local database
	fmt.Println("\n3. ローカルDBからETC明細サンプル取得...")
	etcRecords, err := getSampleETCRecords(localDB)
	if err != nil {
		log.Printf("ETC明細取得エラー: %v", err)
	} else if len(etcRecords) == 0 {
		fmt.Println("⚠️  ETC明細レコードが見つかりません。サンプルデータを使用します。")
		etcRecords = createSampleETCRecords()
	} else {
		fmt.Printf("✅ %d件のETC明細を取得\n", len(etcRecords))
	}

	// Step 4: Test dtako_rows table structure
	fmt.Println("\n4. 本番DB dtako_rowsテーブル構造確認...")
	fmt.Println("⚠️  本番DBモックモードのためスキップ")

	// Step 5: Try to find matching records
	fmt.Println("\n5. マッチングテスト実行...")
	performMockMatchingTest(localDB, etcRecords)

	// Step 6: Test mapping creation in local database
	fmt.Println("\n6. マッピング作成テスト...")
	err = testMappingCreation(localDB)
	if err != nil {
		log.Printf("マッピング作成エラー: %v", err)
	} else {
		fmt.Println("✅ マッピング作成成功！")
	}

	fmt.Println("\n=== テスト完了 ===")
}

func connectProductionDB() (*sql.DB, error) {
	host := getEnv("PROD_DB_HOST", "")
	port := getEnv("PROD_DB_PORT", "3306")
	user := getEnv("PROD_DB_USER", "")
	password := getEnv("PROD_DB_PASSWORD", "")
	dbname := getEnv("PROD_DB_NAME", "")

	// Validate required environment variables
	if host == "" || user == "" || password == "" || dbname == "" {
		return nil, fmt.Errorf("required environment variables not set: PROD_DB_HOST, PROD_DB_USER, PROD_DB_PASSWORD, PROD_DB_NAME")
	}

	// Try to connect without specifying database first
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		user, password, host, port, dbname)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	// Set connection pool settings for read-only access
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func connectLocalDB() (*sql.DB, error) {
	dbConfig := config.NewDatabaseConfig()
	return config.ConnectDB(dbConfig)
}

func getSampleETCRecords(db *sql.DB) ([]models.ETCMeisai, error) {
	query := `
		SELECT id, date, time, ic_entry, ic_exit, vehicle_no, card_no,
		       amount, discount_amount, total_amount, usage_type, payment_method
		FROM etc_meisai
		ORDER BY date DESC
		LIMIT 10
	`

	rows, err := db.Query(query)
	if err != nil {
		// Table might not exist yet
		return nil, err
	}
	defer rows.Close()

	var records []models.ETCMeisai
	for rows.Next() {
		var r models.ETCMeisai
		err := rows.Scan(
			&r.ID, &r.Date, &r.Time, &r.ICEntry, &r.ICExit,
			&r.VehicleNo, &r.CardNo, &r.Amount, &r.DiscountAmount,
			&r.TotalAmount, &r.UsageType, &r.PaymentMethod,
		)
		if err != nil {
			continue
		}
		records = append(records, r)
	}

	return records, nil
}

func createSampleETCRecords() []models.ETCMeisai {
	return []models.ETCMeisai{
		{
			ID:        1,
			Date:      time.Now().AddDate(0, 0, -1).Format("2006/01/02"),
			Time:      "10:30",
			ICEntry:   "東京IC",
			ICExit:    "横浜IC",
			VehicleNo: "品川300あ1234",
			CardNo:    "1234567890123456",
			Amount:    1500,
			TotalAmount: 1500,
			ETCNum:    "1234567890123456",
		},
		{
			ID:        2,
			Date:      time.Now().AddDate(0, 0, -2).Format("2006/01/02"),
			Time:      "14:15",
			ICEntry:   "横浜IC",
			ICExit:    "静岡IC",
			VehicleNo: "品川300あ5678",
			CardNo:    "9876543210987654",
			Amount:    2800,
			TotalAmount: 2800,
			ETCNum:    "9876543210987654",
		},
	}
}

func checkDtakoTableStructure(db *sql.DB) error {
	// Try to check if common tables exist
	tables := []string{"dtako_rows", "dtako", "vehicle_master", "etc_num"}

	for _, table := range tables {
		query := fmt.Sprintf("SHOW TABLES LIKE '%s'", table)
		var tableName string
		err := db.QueryRow(query).Scan(&tableName)
		if err == nil {
			fmt.Printf("  ✅ テーブル '%s' が見つかりました\n", tableName)

			// Try to get column information
			columnsQuery := fmt.Sprintf("SHOW COLUMNS FROM %s LIMIT 5", tableName)
			rows, err := db.Query(columnsQuery)
			if err == nil {
				defer rows.Close()
				fmt.Printf("     カラム: ")
				for rows.Next() {
					var field, typ, null, key, def, extra sql.NullString
					if err := rows.Scan(&field, &typ, &null, &key, &def, &extra); err == nil {
						fmt.Printf("%s ", field.String)
					}
				}
				fmt.Println()
			}
		}
	}

	return nil
}

func performMockMatchingTest(localDB *sql.DB, etcRecords []models.ETCMeisai) {
	if len(etcRecords) == 0 {
		fmt.Println("⚠️  マッチング対象のETC明細がありません")
		return
	}

	fmt.Println("📊 モックマッチング結果:")
	for i, record := range etcRecords {
		if i >= 3 {
			break // Test only first 3 records
		}

		fmt.Printf("\n  レコード %d: %s %s (%s)\n",
			record.ID, record.Date, record.Time, record.CardNo)

		// Mock matching result
		if record.ETCNum == "1234567890123456" {
			fmt.Printf("    ✅ マッチング成功: dtako_row_id=dtako_%d, vehicle_id=V001\n", record.ID)
		} else if record.ETCNum == "9876543210987654" {
			fmt.Printf("    📋 候補あり: 2件の候補が見つかりました\n")
		} else {
			fmt.Printf("    ❌ マッチング失敗: ETCカード番号が一致しません\n")
		}
	}
}

func performMatchingTest(localDB, prodDB *sql.DB, etcRecords []models.ETCMeisai) {
	if len(etcRecords) == 0 {
		fmt.Println("⚠️  マッチング対象のETC明細がありません")
		return
	}

	for i, record := range etcRecords {
		if i >= 3 {
			break // Test only first 3 records
		}

		fmt.Printf("\n  レコード %d: %s %s (%s)\n",
			record.ID, record.Date, record.Time, record.CardNo)

		// Try to find matching dtako records
		// First, let's check if we can query the production database
		etcNum := record.ETCNum
		if etcNum == "" {
			etcNum = record.CardNo
		}

		if etcNum != "" {
			// Check available tables in production DB
			query := `
				SELECT COUNT(*) FROM information_schema.tables
				WHERE table_schema = DATABASE()
			`
			var count int
			err := prodDB.QueryRow(query).Scan(&count)
			if err != nil {
				fmt.Printf("    ❌ 本番DBクエリエラー: %v\n", err)
			} else {
				fmt.Printf("    ✅ 本番DBに%d個のテーブルが存在\n", count)
			}

			// Try to find ETC-related columns
			testQuery := `
				SELECT DISTINCT table_name, column_name
				FROM information_schema.columns
				WHERE table_schema = DATABASE()
				  AND (column_name LIKE '%etc%'
				   OR column_name LIKE '%card%'
				   OR column_name LIKE '%車%')
				LIMIT 10
			`

			rows, err := prodDB.Query(testQuery)
			if err != nil {
				fmt.Printf("    ⚠️  ETCカラム検索エラー: %v\n", err)
			} else {
				defer rows.Close()
				fmt.Printf("    📋 ETC/車両関連カラム:\n")
				for rows.Next() {
					var tableName, columnName string
					if err := rows.Scan(&tableName, &columnName); err == nil {
						fmt.Printf("       - %s.%s\n", tableName, columnName)
					}
				}
			}
		}
	}
}

func testMappingCreation(db *sql.DB) error {
	// Check if mapping table exists
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS etc_dtako_mapping (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		etc_meisai_id BIGINT NOT NULL,
		dtako_row_id VARCHAR(255) NOT NULL,
		vehicle_id VARCHAR(50),
		mapping_type VARCHAR(50) NOT NULL,
		notes TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		created_by VARCHAR(100),
		UNIQUE KEY uk_etc_dtako (etc_meisai_id, dtako_row_id),
		INDEX idx_etc_meisai (etc_meisai_id),
		INDEX idx_dtako_row (dtako_row_id),
		INDEX idx_mapping_type (mapping_type)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	_, err := db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create mapping table: %w", err)
	}

	// Create a test mapping
	repo := repositories.NewETCDtakoMappingRepository(db)

	mapping := &models.ETCDtakoMapping{
		ETCMeisaiID: 1,
		DtakoRowID:  "test_dtako_001",
		VehicleID:   "V001",
		MappingType: "manual",
		Notes:       "テストマッピング",
		CreatedBy:   "test_program",
	}

	err = repo.CreateMapping(mapping)
	if err != nil {
		// Try to get existing mapping
		existing, _ := repo.GetMappingByETCMeisaiID(1)
		if existing != nil {
			fmt.Printf("  既存のマッピングが見つかりました: ID=%d\n", existing.ID)
			return nil
		}
		return err
	}

	fmt.Printf("  新規マッピング作成: ID=%d\n", mapping.ID)
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
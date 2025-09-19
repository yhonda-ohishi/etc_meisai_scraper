# Data Model - ETC明細スクレイピング＆DBインポート（シンプル版）

**Date**: 2025-09-19 | **Branch**: `003-api-project-git-simple`

## Overview

スクレイピングとDBインポートに必要な最小限のデータモデル。複雑な関連やマッチング機能は除外。

## Core Entities

### 1. ETCMeisai（簡素化）

**Table**: `etc_meisai`
**Purpose**: スクレイピングしたETC明細データの保存

```go
type ETCMeisai struct {
    ID             int64     `db:"id" json:"id"`
    UsageDate      time.Time `db:"usage_date" json:"usage_date"`
    EntryIC        string    `db:"entry_ic" json:"entry_ic"`
    ExitIC         string    `db:"exit_ic" json:"exit_ic"`
    TollAmount     int       `db:"toll_amount" json:"toll_amount"`
    VehicleNumber  string    `db:"vehicle_number" json:"vehicle_number"`
    ETCCardNumber  string    `db:"etc_card_number" json:"etc_card_number"`
    AccountType    string    `db:"account_type" json:"account_type"` // corporate/personal
    ImportedAt     time.Time `db:"imported_at" json:"imported_at"`
    CreatedAt      time.Time `db:"created_at" json:"created_at"`
}
```

**Validation Rules**:
- UsageDate: 必須、未来日付不可
- TollAmount: 0以上の整数
- AccountType: "corporate" または "personal"
- 重複チェック: (usage_date, entry_ic, exit_ic, vehicle_number, etc_card_number)

### 2. ImportSession（オプション）

**Table**: `import_sessions`
**Purpose**: インポート実行履歴（ログ用）

```go
type ImportSession struct {
    ID           int64     `db:"id"`
    AccountType  string    `db:"account_type"`
    StartDate    time.Time `db:"start_date"`
    EndDate      time.Time `db:"end_date"`
    RecordCount  int       `db:"record_count"`
    Status       string    `db:"status"` // success/failed
    ExecutedAt   time.Time `db:"executed_at"`
    ErrorMessage string    `db:"error_message"`
}
```

## Database Schema

```sql
-- メインテーブル
CREATE TABLE IF NOT EXISTS etc_meisai (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    usage_date DATE NOT NULL,
    entry_ic VARCHAR(100) NOT NULL,
    exit_ic VARCHAR(100) NOT NULL,
    toll_amount INT NOT NULL,
    vehicle_number VARCHAR(20),
    etc_card_number VARCHAR(20),
    account_type VARCHAR(20) NOT NULL,
    imported_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- 重複防止用ユニークインデックス
    UNIQUE KEY unique_record (
        usage_date,
        entry_ic,
        exit_ic,
        vehicle_number,
        etc_card_number
    ),

    -- 検索用インデックス
    INDEX idx_usage_date (usage_date),
    INDEX idx_account_type (account_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- インポート履歴テーブル（オプション）
CREATE TABLE IF NOT EXISTS import_sessions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    account_type VARCHAR(20) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    record_count INT DEFAULT 0,
    status VARCHAR(20) NOT NULL,
    executed_at TIMESTAMP NOT NULL,
    error_message TEXT,

    INDEX idx_executed_at (executed_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## Repository Interface

```go
type ETCRepository interface {
    // 基本操作
    Insert(record *ETCMeisai) error
    BulkInsert(records []ETCMeisai) error

    // 重複チェック
    Exists(record *ETCMeisai) (bool, error)

    // 検索
    FindByDateRange(start, end time.Time) ([]ETCMeisai, error)

    // トランザクション
    WithTransaction(fn func(*sql.Tx) error) error
}
```

## CSV Mapping

CSVファイルからETCMeisaiへのマッピング:

| CSV列 | フィールド | 変換 |
|-------|-----------|------|
| 利用日 | UsageDate | 日付パース |
| 入口IC | EntryIC | そのまま |
| 出口IC | ExitIC | そのまま |
| 料金 | TollAmount | 整数変換 |
| 車両番号 | VehicleNumber | そのまま |
| カード番号 | ETCCardNumber | そのまま |

## Transaction Strategy

```go
// バルクインポートのトランザクション処理
func ImportCSVData(csvPath string, accountType string) error {
    records := ParseCSV(csvPath)

    return repo.WithTransaction(func(tx *sql.Tx) error {
        for _, record := range records {
            if exists, _ := repo.Exists(&record); !exists {
                if err := repo.Insert(&record); err != nil {
                    return err // ロールバック
                }
            }
        }
        return nil // コミット
    })
}
```

## Performance Considerations

- **バッチサイズ**: 100レコード/トランザクション
- **重複チェック**: UNIQUEインデックスで自動処理
- **接続プール**: 最大10接続
- **タイムアウト**: 30秒/クエリ

## Migration Script

```sql
-- 既存テーブルからの移行（必要な場合）
INSERT INTO etc_meisai (
    usage_date, entry_ic, exit_ic, toll_amount,
    vehicle_number, etc_card_number, account_type, imported_at
)
SELECT
    usage_date, entrance_ic, exit_ic, toll_amount,
    vehicle_number, etc_card_no, account_type, NOW()
FROM old_etc_table
ON DUPLICATE KEY UPDATE imported_at = NOW();
```

---
*Simplified Data Model - Focus on core functionality*
# Data Model Specification: etc_meisai Server Repository Integration

**Branch**: `001-db-service-integration`
**Date**: 2025-09-21
**Status**: Draft

## 概要

etc_meisai統合のためのデータモデル定義。GORM統一モデルとProtocol Buffers型定義の対応関係を明確化し、バリデーションルール、状態遷移、関連性を定義する。

## 主要エンティティ

### 1. ETCMeisaiRecord (ETC明細レコード)

#### GORMモデル定義
```go
type ETCMeisaiRecord struct {
    ID              int64     `gorm:"primaryKey;autoIncrement"`
    Hash            string    `gorm:"uniqueIndex;size:64;not null"`
    Date            time.Time `gorm:"index;not null"`
    Time            string    `gorm:"size:8;not null"`
    EntranceIC      string    `gorm:"size:100;not null"`
    ExitIC          string    `gorm:"size:100;not null"`
    TollAmount      int       `gorm:"not null"`
    CarNumber       string    `gorm:"index;size:20;not null"`
    ETCCardNumber   string    `gorm:"index;size:20;not null"`
    ETCNum          *string   `gorm:"index;size:50"`
    DtakoRowID      *int64    `gorm:"index"`
    CreatedAt       time.Time `gorm:"autoCreateTime"`
    UpdatedAt       time.Time `gorm:"autoUpdateTime"`
    DeletedAt       gorm.DeletedAt `gorm:"index"`
}
```

#### Protocol Buffers定義
```protobuf
message ETCMeisaiRecord {
  int64 id = 1;
  string hash = 2;
  string date = 3;  // YYYY-MM-DD format
  string time = 4;  // HH:MM:SS format
  string entrance_ic = 5;
  string exit_ic = 6;
  int32 toll_amount = 7;
  string car_number = 8;
  string etc_card_number = 9;
  optional string etc_num = 10;
  optional int64 dtako_row_id = 11;
  google.protobuf.Timestamp created_at = 12;
  google.protobuf.Timestamp updated_at = 13;
}
```

#### バリデーションルール
- **hash**: 必須、64文字、SHA256形式、ユニーク制約
- **date**: 必須、有効な日付、未来日付不可
- **time**: 必須、HH:MM:SS形式
- **entrance_ic/exit_ic**: 必須、最大100文字、有効なIC名
- **toll_amount**: 必須、0以上、最大999999
- **car_number**: 必須、日本の車両番号形式
- **etc_card_number**: 必須、16-19桁の数字
- **etc_num**: オプション、ETC2.0車載器番号形式

#### インデックス
- PRIMARY: id
- UNIQUE: hash
- INDEX: date, car_number, etc_card_number, etc_num, dtako_row_id
- COMPOSITE: (date, car_number)

### 2. ETCMapping (マッピング情報)

#### GORMモデル定義
```go
type ETCMapping struct {
    ID              int64           `gorm:"primaryKey;autoIncrement"`
    ETCRecordID     int64           `gorm:"not null;index"`
    ETCRecord       ETCMeisaiRecord `gorm:"foreignKey:ETCRecordID"`
    MappingType     string          `gorm:"size:50;not null;index"`
    MappedEntityID  int64           `gorm:"not null;index"`
    MappedEntityType string         `gorm:"size:50;not null;index"`
    Confidence      float32         `gorm:"default:1.0"`
    Status          string          `gorm:"size:20;default:'active';index"`
    Metadata        datatypes.JSON  `gorm:"type:json"`
    CreatedBy       string          `gorm:"size:100"`
    CreatedAt       time.Time       `gorm:"autoCreateTime"`
    UpdatedAt       time.Time       `gorm:"autoUpdateTime"`
}
```

#### Protocol Buffers定義
```protobuf
message ETCMapping {
  int64 id = 1;
  int64 etc_record_id = 2;
  ETCMeisaiRecord etc_record = 3;
  string mapping_type = 4;
  int64 mapped_entity_id = 5;
  string mapped_entity_type = 6;
  float confidence = 7;
  string status = 8;
  google.protobuf.Struct metadata = 9;
  string created_by = 10;
  google.protobuf.Timestamp created_at = 11;
  google.protobuf.Timestamp updated_at = 12;
}

enum MappingStatus {
  MAPPING_STATUS_UNSPECIFIED = 0;
  MAPPING_STATUS_ACTIVE = 1;
  MAPPING_STATUS_INACTIVE = 2;
  MAPPING_STATUS_PENDING = 3;
  MAPPING_STATUS_REJECTED = 4;
}
```

#### バリデーションルール
- **etc_record_id**: 必須、存在するレコードID
- **mapping_type**: 必須、定義済みタイプ（dtako, expense, invoice）
- **mapped_entity_id**: 必須、正の整数
- **mapped_entity_type**: 必須、定義済みエンティティタイプ
- **confidence**: 0.0-1.0の範囲
- **status**: active, inactive, pending, rejected のいずれか

#### 状態遷移
```
pending → active (承認時)
pending → rejected (却下時)
active → inactive (無効化時)
inactive → active (再有効化時)
```

### 3. ImportSession (インポートセッション)

#### GORMモデル定義
```go
type ImportSession struct {
    ID              string          `gorm:"primaryKey;size:36"`  // UUID
    AccountType     string          `gorm:"size:20;not null;index"`
    AccountID       string          `gorm:"size:50;not null;index"`
    FileName        string          `gorm:"size:255;not null"`
    FileSize        int64           `gorm:"not null"`
    Status          string          `gorm:"size:20;not null;index"`
    TotalRows       int             `gorm:"default:0"`
    ProcessedRows   int             `gorm:"default:0"`
    SuccessRows     int             `gorm:"default:0"`
    ErrorRows       int             `gorm:"default:0"`
    DuplicateRows   int             `gorm:"default:0"`
    StartedAt       time.Time       `gorm:"not null"`
    CompletedAt     *time.Time
    ErrorLog        datatypes.JSON  `gorm:"type:json"`
    CreatedBy       string          `gorm:"size:100"`
    CreatedAt       time.Time       `gorm:"autoCreateTime"`
}
```

#### Protocol Buffers定義
```protobuf
message ImportSession {
  string id = 1;  // UUID
  string account_type = 2;
  string account_id = 3;
  string file_name = 4;
  int64 file_size = 5;
  ImportStatus status = 6;
  int32 total_rows = 7;
  int32 processed_rows = 8;
  int32 success_rows = 9;
  int32 error_rows = 10;
  int32 duplicate_rows = 11;
  google.protobuf.Timestamp started_at = 12;
  google.protobuf.Timestamp completed_at = 13;
  repeated ImportError error_log = 14;
  string created_by = 15;
  google.protobuf.Timestamp created_at = 16;
}

enum ImportStatus {
  IMPORT_STATUS_UNSPECIFIED = 0;
  IMPORT_STATUS_PENDING = 1;
  IMPORT_STATUS_PROCESSING = 2;
  IMPORT_STATUS_COMPLETED = 3;
  IMPORT_STATUS_FAILED = 4;
  IMPORT_STATUS_CANCELLED = 5;
}

message ImportError {
  int32 row_number = 1;
  string error_type = 2;
  string error_message = 3;
  string raw_data = 4;
}
```

#### バリデーションルール
- **id**: UUID v4形式
- **account_type**: corporate, personal のいずれか
- **account_id**: 必須、存在するアカウントID
- **file_name**: 必須、.csv拡張子
- **file_size**: 0より大きい、最大100MB
- **status**: 定義済みステータスのみ

#### 状態遷移
```
pending → processing (処理開始時)
processing → completed (正常終了時)
processing → failed (エラー発生時)
processing → cancelled (ユーザーキャンセル時)
```

## リレーション定義

### 1対多リレーション
- ImportSession (1) → ETCMeisaiRecord (多)
  - インポートセッションで取り込まれたレコード
- ETCMeisaiRecord (1) → ETCMapping (多)
  - 1つのETCレコードに複数のマッピング可能

### 多対多リレーション
なし（現時点では不要）

## インデックス戦略

### パフォーマンス最適化インデックス
```sql
-- 頻繁な検索パターン用
CREATE INDEX idx_etc_date_car ON etc_meisai_records(date, car_number);
CREATE INDEX idx_etc_card_date ON etc_meisai_records(etc_card_number, date);
CREATE INDEX idx_mapping_entity ON etc_mappings(mapped_entity_type, mapped_entity_id);
CREATE INDEX idx_import_account ON import_sessions(account_type, account_id, status);

-- 全文検索用（将来的に必要な場合）
CREATE INDEX idx_etc_ic_fulltext ON etc_meisai_records USING gin(to_tsvector('japanese', entrance_ic || ' ' || exit_ic));
```

## データ整合性制約

### 外部キー制約
```sql
ALTER TABLE etc_mappings
  ADD CONSTRAINT fk_etc_record
  FOREIGN KEY (etc_record_id)
  REFERENCES etc_meisai_records(id)
  ON DELETE CASCADE;
```

### チェック制約
```sql
ALTER TABLE etc_meisai_records
  ADD CONSTRAINT chk_toll_amount CHECK (toll_amount >= 0);

ALTER TABLE etc_mappings
  ADD CONSTRAINT chk_confidence CHECK (confidence >= 0 AND confidence <= 1);

ALTER TABLE import_sessions
  ADD CONSTRAINT chk_rows CHECK (
    processed_rows >= 0 AND
    success_rows >= 0 AND
    error_rows >= 0 AND
    duplicate_rows >= 0 AND
    processed_rows = success_rows + error_rows + duplicate_rows
  );
```

## マイグレーション戦略

### 初期マイグレーション
```go
func Migrate001_CreateETCTables(db *gorm.DB) error {
    // テーブル作成
    if err := db.AutoMigrate(
        &ETCMeisaiRecord{},
        &ETCMapping{},
        &ImportSession{},
    ); err != nil {
        return err
    }

    // インデックス追加
    db.Exec("CREATE INDEX IF NOT EXISTS idx_etc_date_car ON etc_meisai_records(date, car_number)")
    db.Exec("CREATE INDEX IF NOT EXISTS idx_mapping_entity ON etc_mappings(mapped_entity_type, mapped_entity_id)")

    return nil
}
```

### データ移行（既存データがある場合）
```go
func MigrateExistingData(oldDB, newDB *gorm.DB) error {
    // バッチ処理で既存データを移行
    var oldRecords []OldETCRecord
    batchSize := 1000

    return oldDB.FindInBatches(&oldRecords, batchSize, func(tx *gorm.DB, batch int) error {
        var newRecords []ETCMeisaiRecord
        for _, old := range oldRecords {
            newRecords = append(newRecords, convertToNewFormat(old))
        }
        return newDB.CreateInBatches(newRecords, 100).Error
    }).Error
}
```

## パフォーマンス考慮事項

### バルクインサート最適化
```go
// 大量データインポート時の最適化
func BulkInsertETCRecords(db *gorm.DB, records []ETCMeisaiRecord) error {
    // トランザクション内でバッチ処理
    return db.Transaction(func(tx *gorm.DB) error {
        // 一時的にインデックスを無効化（PostgreSQLの場合）
        tx.Exec("SET LOCAL synchronous_commit = OFF")

        // バッチサイズを調整してメモリ使用量を制御
        batchSize := 500
        for i := 0; i < len(records); i += batchSize {
            end := i + batchSize
            if end > len(records) {
                end = len(records)
            }
            if err := tx.CreateInBatches(records[i:end], 100).Error; err != nil {
                return err
            }
        }
        return nil
    })
}
```

### クエリ最適化
```go
// 効率的なページネーション
func GetETCRecordsPaginated(db *gorm.DB, page, size int) ([]ETCMeisaiRecord, int64, error) {
    var records []ETCMeisaiRecord
    var total int64

    // カウントとデータ取得を別クエリで実行
    db.Model(&ETCMeisaiRecord{}).Count(&total)

    err := db.Scopes(Paginate(page, size)).
        Preload("Mappings", "status = ?", "active").
        Find(&records).Error

    return records, total, err
}
```

## セキュリティ考慮事項

### 個人情報保護
- ETCカード番号は部分的にマスキング
- 車両番号の表示制限
- アクセスログの記録

### データ暗号化
```go
// 機密データの暗号化
func (r *ETCMeisaiRecord) BeforeSave(tx *gorm.DB) error {
    // ETCカード番号を暗号化
    if r.ETCCardNumber != "" {
        encrypted, err := encryptSensitiveData(r.ETCCardNumber)
        if err != nil {
            return err
        }
        r.ETCCardNumber = encrypted
    }
    return nil
}
```

## 監査ログ

### 変更履歴追跡
```go
type AuditLog struct {
    ID          int64          `gorm:"primaryKey"`
    TableName   string         `gorm:"size:50;not null;index"`
    RecordID    int64          `gorm:"not null;index"`
    Action      string         `gorm:"size:20;not null"`
    ChangedBy   string         `gorm:"size:100;not null"`
    ChangedAt   time.Time      `gorm:"not null;index"`
    OldValues   datatypes.JSON `gorm:"type:json"`
    NewValues   datatypes.JSON `gorm:"type:json"`
}
```

## 今後の拡張ポイント

1. **全文検索機能**
   - IC名の曖昧検索
   - 複合条件検索

2. **統計情報テーブル**
   - 月次集計データ
   - 利用傾向分析

3. **キャッシュテーブル**
   - 頻繁にアクセスされるデータのキャッシュ
   - マテリアライズドビュー

---
*Data Model Specification v1.0 - 2025-09-21*
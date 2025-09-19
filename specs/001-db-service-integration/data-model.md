# データモデル設計: データベースサービス統合

**作成日**: 2025-09-19
**対象**: ETC明細システム ⇔ db_service統合

## 統合データモデル概要

### 設計原則
1. **db_serviceモデルを基準**: GORMベースの正規化されたモデル
2. **後方互換性維持**: 既存38フィールドのエイリアス対応
3. **データ整合性確保**: ハッシュベース重複検出統一
4. **型安全性**: Protocol Buffers⇔Go struct変換

## 主要エンティティ設計

### 1. ETCMeisai (統合メインモデル)

#### 基本構造 (db_serviceベース)
```go
type ETCMeisai struct {
    // Primary Key
    ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`

    // 利用情報
    UseDate     time.Time `gorm:"not null" json:"use_date"`
    UseTime     string    `gorm:"size:8" json:"use_time"`

    // 料金所情報
    EntryIC     string    `gorm:"size:100" json:"entry_ic"`
    ExitIC      string    `gorm:"size:100" json:"exit_ic"`

    // 金額情報
    Amount      int32     `gorm:"not null" json:"amount"`

    // 車両情報
    CarNumber   string    `gorm:"size:20" json:"car_number"`

    // ETC情報
    ETCNumber   string    `gorm:"size:20;index" json:"etc_number"`

    // データ整合性
    Hash        string    `gorm:"size:64;uniqueIndex" json:"hash"`

    // タイムスタンプ
    CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
```

#### 互換性レイヤー (既存38フィールド対応)
```go
// ETCMeisaiCompat: 既存システム互換性用構造体
type ETCMeisaiCompat struct {
    *ETCMeisai

    // エイリアスフィールド（既存38フィールド対応）
    ICEntry      *string `json:"ic_entry,omitempty"`      // EntryIC のエイリアス
    EntryICName  *string `json:"entry_ic_name,omitempty"` // 未使用フィールド
    ExitICName   *string `json:"exit_ic_name,omitempty"`  // 未使用フィールド
    UsageDate    *string `json:"usage_date,omitempty"`    // UseDate のエイリアス
    UsageTime    *string `json:"usage_time,omitempty"`    // UseTime のエイリアス
    // ... 他の既存フィールド
}

// 変換メソッド
func (e *ETCMeisai) ToCompat() *ETCMeisaiCompat {
    compat := &ETCMeisaiCompat{ETCMeisai: e}

    // エイリアス設定
    compat.ICEntry = &e.EntryIC
    compat.UsageDate = &e.UseDate.Format("2006-01-02")
    compat.UsageTime = &e.UseTime

    return compat
}
```

### 2. ETCMeisaiMapping (マッピング管理)

```go
type ETCMeisaiMapping struct {
    ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    ETCMeisaiID int64     `gorm:"not null;index" json:"etc_meisai_id"`
    DTakoRowID  string    `gorm:"size:50;not null;index" json:"dtako_row_id"`
    MappingType string    `gorm:"size:20;not null" json:"mapping_type"` // auto, manual
    Confidence  float32   `gorm:"default:0" json:"confidence"`
    CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

    // リレーション
    ETCMeisai *ETCMeisai `gorm:"foreignKey:ETCMeisaiID"`
}
```

### 3. バッチ処理用モデル

```go
type ETCImportBatch struct {
    ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    BatchHash       string    `gorm:"size:64;uniqueIndex" json:"batch_hash"`
    FileName        string    `gorm:"size:255" json:"file_name"`
    TotalRecords    int32     `gorm:"not null" json:"total_records"`
    ProcessedCount  int32     `gorm:"default:0" json:"processed_count"`
    ErrorCount      int32     `gorm:"default:0" json:"error_count"`
    Status          string    `gorm:"size:20;default:'pending'" json:"status"`
    StartTime       *time.Time `json:"start_time"`
    CompleteTime    *time.Time `json:"complete_time"`
    CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}
```

## リポジトリインターフェース統合

### 統合ETCRepository
```go
type ETCRepository interface {
    // 基本CRUD (db_serviceベース)
    Create(data *ETCMeisai) error
    GetByID(id int64) (*ETCMeisai, error)
    Update(data *ETCMeisai) error
    DeleteByID(id int64) error
    List(params *ETCListParams) ([]*ETCMeisai, int64, error)

    // ETC明細固有機能
    BulkInsert(records []*ETCMeisai) error
    GetByDateRange(start, end time.Time) ([]*ETCMeisai, error)
    GetByHash(hash string) (*ETCMeisai, error)
    GetUnmappedRecords(start, end time.Time) ([]*ETCMeisai, error)

    // ハッシュベース操作
    ListByHashBatch(hashes []string) ([]*ETCMeisai, error)
    CheckDuplicatesByHash(hashes []string) (map[string]bool, error)

    // 集計・統計
    GetSummaryByDateRange(start, end time.Time) (*ETCSummary, error)
    GetMonthlyStats(year int, month int) (*ETCMonthlyStats, error)
}
```

### マッピングRepository
```go
type ETCMappingRepository interface {
    Create(mapping *ETCMeisaiMapping) error
    GetByETCMeisaiID(etcMeisaiID int64) ([]*ETCMeisaiMapping, error)
    GetByDTakoRowID(dtakoRowID string) (*ETCMeisaiMapping, error)
    UpdateMappingType(id int64, mappingType string, confidence float32) error
    DeleteByID(id int64) error

    // 自動マッチング支援
    FindPotentialMatches(etcMeisaiID int64) ([]*PotentialMatch, error)
    BulkCreateMappings(mappings []*ETCMeisaiMapping) error
}
```

## データフロー設計

### 1. CSVインポートフロー
```
CSV File → Parser → CompatLayer → ETCMeisai → Repository → Database
                                      ↓
                              Hash Check → Duplicate Detection
```

### 2. API レスポンスフロー
```
Database → Repository → ETCMeisai → CompatLayer → JSON Response
                                          ↓
                                    Legacy Format
```

### 3. マッピングフロー
```
ETCMeisai → Auto Match Service → DTako Data → ETCMeisaiMapping → Repository
                ↓
         Confidence Score → Manual Review Queue
```

## バリデーション規則

### ETCMeisai
- `UseDate`: 必須、過去1年以内
- `Amount`: 必須、正の整数
- `ETCNumber`: 必須、ETC番号形式チェック
- `Hash`: 必須、重複チェック

### ETCMeisaiMapping
- `ETCMeisaiID`: 存在チェック
- `DTakoRowID`: 形式チェック
- `MappingType`: enum値チェック (auto, manual)
- `Confidence`: 0.0-1.0の範囲

## パフォーマンス最適化

### インデックス戦略
```sql
-- 主要検索パターン用インデックス
CREATE INDEX idx_etc_use_date ON etc_meisais(use_date);
CREATE INDEX idx_etc_number ON etc_meisais(etc_number);
CREATE INDEX idx_etc_hash ON etc_meisais(hash);
CREATE INDEX idx_etc_created_at ON etc_meisais(created_at);

-- マッピング用インデックス
CREATE INDEX idx_mapping_etc_id ON etc_meisai_mappings(etc_meisai_id);
CREATE INDEX idx_mapping_dtako_id ON etc_meisai_mappings(dtako_row_id);
```

### バッチ処理最適化
- バルクインサート: 1000件単位のバッチ処理
- ハッシュチェック: バッチでの重複検出
- トランザクション: バッチ単位でのコミット

## 移行戦略

### Phase 1: モデル統合
1. db_serviceモデルの導入
2. 互換性レイヤーの実装
3. バリデーション統一

### Phase 2: Repository統合
1. 統合インターフェースの実装
2. GORM移行
3. 既存クエリの変換

### Phase 3: データ移行
1. 既存データのバックアップ
2. スキーマ移行
3. データ整合性検証

### Phase 4: 最適化
1. パフォーマンステスト
2. インデックス最適化
3. クリーンアップ

## テスト戦略

### ユニットテスト
- モデルバリデーション
- 変換メソッド
- ハッシュ生成

### 統合テスト
- Repository CRUD操作
- バッチ処理
- データ整合性

### 契約テスト
- API互換性
- gRPC契約
- データ形式

この設計により、既存システムとの互換性を保ちながら、より保守性の高いデータモデルへの移行が可能になります。
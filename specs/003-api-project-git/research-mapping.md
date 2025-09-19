# リサーチ: ETCDtakoMappingテーブル実装

**日付**: 2025-09-18
**目的**: ETCMeisaiとDtakoRowIDのマッピング機能の技術調査

## 1. 既存システム分析

### DB接続パターン
**Decision**: 既存のMySQLデータベース使用（ローカル）
**Rationale**:
- 既存のetc_meisaiテーブルがMySQLに存在
- schema.sqlファイルが既に定義済み
- リポジトリパターンが確立済み

**Alternatives considered**:
- SQLite: 軽量だが、既存システムとの不整合
- PostgreSQL: 高機能だが、既存環境との互換性問題

### リポジトリパターン
**Decision**: 既存のリポジトリパターンを踏襲
**Rationale**:
- repositories/ディレクトリに既存実装あり
- etc_dtako_mapping_repository.goは既に存在（更新必要）
- sqlxライブラリ使用で型安全なクエリ実装

**Alternatives considered**:
- ORMフレームワーク（GORM）: オーバースペック
- 生SQL直接実行: 型安全性の欠如

### APIハンドラー登録
**Decision**: registry.goでの自動登録メカニズム使用
**Rationale**:
- v0.0.15でハンドラー自動登録メカニズムが実装済み
- api_handlers_mapping.goが既に存在
- 統一的な登録方法

**Alternatives considered**:
- main.goでの手動登録: メンテナンス性が低い
- ルーターライブラリ使用: 既存システムとの不整合

## 2. データモデル設計

### テーブル構造の決定
**Decision**: MatchScoreフィールドを削除したシンプルな構造
**Rationale**:
- ユーザー要求により不要と判断
- シンプルな設計原則に従う
- パフォーマンス向上（不要なフィールド削除）

**最終テーブル構造**:
```sql
CREATE TABLE IF NOT EXISTS etc_dtako_mapping (
    id INT AUTO_INCREMENT PRIMARY KEY,
    etc_meisai_id INT NOT NULL,
    dtako_row_id VARCHAR(100) NOT NULL,
    vehicle_id VARCHAR(50),
    mapping_type VARCHAR(20) NOT NULL DEFAULT 'manual',
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    FOREIGN KEY (etc_meisai_id) REFERENCES etc_meisai(id) ON DELETE CASCADE,
    INDEX idx_etc_meisai_id (etc_meisai_id),
    INDEX idx_dtako_row_id (dtako_row_id),
    INDEX idx_vehicle_id (vehicle_id)
)
```

## 3. API設計

### エンドポイント構成
**Decision**: RESTfulなエンドポイント設計
**Rationale**:
- 既存APIとの一貫性
- 標準的なHTTPメソッド使用
- Swagger統合済み

**エンドポイント一覧**:
1. `POST /api/etc/mapping` - 新規マッピング作成
2. `GET /api/etc/mapping/{id}` - ID指定で取得
3. `GET /api/etc/mapping/by-meisai/{meisai_id}` - ETC明細IDで検索
4. `GET /api/etc/mapping/by-dtako/{dtako_row_id}` - デジタコIDで検索
5. `PUT /api/etc/mapping/{id}` - 更新
6. `DELETE /api/etc/mapping/{id}` - 削除
7. `GET /api/etc/mapping/list` - ページネーション付き一覧

## 4. 実装優先度

### 必須機能
1. テーブル作成SQL
2. モデル修正（MatchScore削除）
3. 基本CRUD操作
4. 検索機能

### オプション機能
1. バルクインポート
2. 履歴管理
3. 自動マッピング提案

## 5. テスト戦略

### 単体テスト
- モデルのバリデーション
- リポジトリメソッド

### 統合テスト
- API全エンドポイント
- DB接続とトランザクション
- エラーハンドリング

## 6. パフォーマンス考慮

### インデックス戦略
**Decision**: 3つのインデックス作成
**Rationale**:
- etc_meisai_id: 頻繁な検索対象
- dtako_row_id: 逆引き検索に必要
- vehicle_id: 車両ベースの検索最適化

### クエリ最適化
- ページネーション実装（LIMIT/OFFSET）
- 必要なカラムのみSELECT
- JOINを避けたシンプルなクエリ

## まとめ

すべての技術的不明点が解決され、実装準備が整いました：
- ローカルMySQL使用確定
- MatchScore削除によるシンプル化
- 既存パターンとの整合性確保
- パフォーマンスを考慮した設計

次のフェーズ（Phase 1: Design & Contracts）へ進む準備完了。
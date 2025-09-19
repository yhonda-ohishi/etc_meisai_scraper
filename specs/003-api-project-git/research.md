# 技術調査: ETC明細APIシステム完成（prod db使用dtako_row_idマッチング）

**最終更新**: 2025-09-18
**対象**: ETC明細システムにおける本番dtakoデータベース統合

## 主要技術決定事項

### Decision: dtako_row_idマッチングアルゴリズム
**選択決定**: ハッシュベース類似度マッチング + 手動確認フロー
### Rationale:
既存のハッシュベースシステムを拡張し、dtako_row_idとの自動マッチングを実現する：
1. 日付・時刻の近似マッチング（±30分許容）
2. 金額の完全一致または近似一致（±100円許容）
3. 車両情報の部分マッチング

### Alternatives considered:
- 機械学習ベースの類似度計算 → 計算コストが高く、現在の要件に過剰
- 完全一致のみのマッチング → マッチ率が低く、手動作業が増加
- 外部APIによるマッチング → 本番データへの外部アクセス制限により不可

## 技術スタックの最適化

### 1. データベース統合 (`config/` package) の改善

**Decision**: デュアルデータベース接続管理の最適化
**Rationale**:
- 現在の実装では単一のDB接続プールを使用
- 本番DBは読み取り専用で接続プール数を制限する必要がある
- ローカルDBは読み書き可能で十分なプール数を確保

**実装戦略**:
```go
// デュアルDB管理の改善
type DatabaseManager struct {
    LocalDB *sql.DB
    ProdDB  *sql.DB
}
```

### 2. マッチングサービス (`services/` package) の設計

**Decision**: 複数段階のマッチング戦略
**Rationale**:
- ステップ1: 完全一致マッチング（高精度）
- ステップ2: 近似マッチング（中精度）
- ステップ3: 候補提示（手動確認）

**実装戦略**:
- `AutoMatchingService`: 自動マッチングロジック
- `ManualMatchingService`: 手動マッチング管理
- `MatchingValidationService`: 結果検証

### 3. パフォーマンス最適化の決定

**Decision**: バッチ処理 + 並列化アプローチ
**Rationale**:
- 大量データ処理時のメモリ効率
- 本番DBへの負荷分散
- ユーザーへの進捗フィードバック

**実装戦略**:
- 1000件/バッチでの処理
- 最大10 goroutinesでの並列実行
- 進捗追跡とキャンセル機能

## 本番環境との統合戦略

### 1. セキュリティ要件

**Decision**: 読み取り専用アクセス + 接続プール制限
**Rationale**:
- 本番データへの安全なアクセス
- システムリソースの保護
- 監査ログの必要性

### 2. エラーハンドリング

**Decision**: 段階的フォールバック戦略
**Rationale**:
- 本番DB接続失敗時のローカル処理継続
- 部分的な結果でも価値提供
- エラー状況の詳細ログ記録

### 3. データ同期

**Decision**: オンデマンド同期 + キャッシュ戦略なし
**Rationale**:
ユーザーからの明確な指示により、キャッシュ機能は不要：
- リアルタイムデータの重要性
- キャッシュ管理の複雑さ回避
- ストレージ要件の簡素化

## APIエンドポイント設計の改善

### 1. 新規エンドポイント決定

**Decision**: RESTful API + 非同期処理サポート
**Rationale**:
- 既存のAPI設計パターンとの一貫性
- 大量データ処理の非同期化
- 進捗追跡機能の必要性

**新規エンドポイント**:
- `POST /api/etc/mapping`: マッピング作成
- `GET /api/etc/mapping/list`: マッピング一覧
- `PUT /api/etc/mapping/{id}`: マッピング更新
- `POST /api/etc/auto-match`: 自動マッチング実行

### 2. レスポンス設計

**Decision**: 統一されたレスポンス形式 + 詳細エラー情報
**Rationale**:
- クライアント側の実装簡素化
- デバッグとトラブルシューティングの向上
- API使用者の開発効率向上

## テストストラテジーの決定

### 1. 統合テスト拡張

**Decision**: 本番DB接続を含む統合テスト環境
**Rationale**:
- 実際の動作環境での検証
- データ品質の確認
- パフォーマンス特性の測定

### 2. モック戦略

**Decision**: 本番DB操作のモック化 + 実データでの検証
**Rationale**:
- 開発時の迅速なテスト実行
- 本番環境への影響回避
- CI/CD パイプラインでの安定実行

## セキュリティと監査

### 1. アクセス制御

**Decision**: 読み取り専用権限 + 接続ログ記録
**Rationale**:
- 本番データの保護
- 操作履歴の追跡
- セキュリティ監査の要件

### 2. データプライバシー

**Decision**: 必要最小限のデータアクセス
**Rationale**:
- プライバシー保護
- 法的要件の遵守
- システムリスクの最小化

## 実装優先順位の決定

### Phase 1 (最優先): 基盤システム完成
1. **失敗テスト修正**: 既存の統合テストを通す（完了済み）
2. **デュアルDB接続**: 本番DB統合
3. **基本マッピングモデル**: ETCDtakoMapping

### Phase 2: マッチング機能実装
1. **自動マッチングサービス**: 基本アルゴリズム
2. **手動マッピング管理**: CRUD操作
3. **バッチ処理**: 大量データ対応

### Phase 3: API完成
1. **新規エンドポイント**: マッピング管理API群
2. **統合テスト**: 完全なE2Eテスト
3. **ドキュメント**: quickstart.mdとAPI仕様

## 設定とデプロイメント

### 1. 環境変数設計

**Decision**: 環境別設定ファイル + .env オーバーライド
**Rationale**:
- 開発・本番環境の明確な分離
- 機密情報の安全な管理
- デプロイメントの自動化対応

### 2. ログ戦略

**Decision**: 構造化ログ + レベル別出力
**Rationale**:
- 運用時の問題特定
- パフォーマンス監視
- セキュリティ監査証跡

---

## ryohi_sub_calシステム参考分析

### Decision: etc_numフィールドを活用した精密マッチング
**選択決定**: ryohi_sub_calのマッチングロジックを参考にしたetc_num活用戦略
### Rationale:
ryohi_sub_calシステムで実証されたアプローチを参考に、以下を実装：
1. **etc_numによる車両特定**: ETCカード番号を使った車両マッピング
2. **時間範囲マッチング**: dtako運行時間との時間差チェック
3. **金額整合性確認**: 料金データと運行距離の妥当性検証

### ryohi_sub_cal参考ポイント:
- ETCカード番号(etc_num)による車両識別の精度向上
- 時間軸での運行データとの関連付け手法
- 本番データベースへの効率的なクエリ最適化

**実装戦略**:
```go
// etc_num活用マッチング
type ETCNumMatcher struct {
    ETCNum      string    // ETCカード番号
    TimeRange   TimeWindow // 運行時間範囲
    VehicleID   string    // 車両ID
    AmountRange AmountRange // 料金範囲
}
```

## 本番データベース統合の詳細設計

### Decision: dtako_rowsテーブルとの直接結合クエリ
**選択決定**: リアルタイム本番DBクエリ + ローカルキャッシュなし戦略
**Rationale**:
- ユーザー要求により、キャッシュ機能は実装しない
- 本番データの最新性を重視
- クエリ最適化による性能確保

**本番DBクエリ戦略**:
```sql
-- etc_numとdtako_row_idのマッチングクエリ例
SELECT dr.dtako_row_id, dr.start_time, dr.end_time, dr.vehicle_id
FROM dtako_rows dr
WHERE dr.etc_num = ?
  AND dr.start_time <= ?
  AND dr.end_time >= ?
  AND dr.vehicle_id IN (SELECT vehicle_id FROM vehicle_etc_mapping WHERE etc_num = ?)
```

### Decision: マルチステージマッチングアルゴリズム
**選択決定**: 段階的マッチング精度向上
**Rationale**:
1. **完全一致段階**: etc_num + 時刻完全マッチ
2. **近似マッチング段階**: etc_num + 時間範囲マッチ（±30分）
3. **候補抽出段階**: etc_num + 車両IDマッチ（手動確認用）

## etc_num活用による性能最適化

### Decision: インデックス戦略最適化
**選択決定**: etc_numを主キーとした複合インデックス
**Rationale**:
- 本番DBでのクエリ性能向上
- etc_numの選択性を活用
- 時間範囲クエリの最適化

**インデックス設計**:
```sql
-- 推奨複合インデックス
CREATE INDEX idx_dtako_etc_time ON dtako_rows (etc_num, start_time, end_time);
CREATE INDEX idx_dtako_vehicle_etc ON dtako_rows (vehicle_id, etc_num);
```

## API設計への影響

### Decision: etc_num中心のAPIエンドポイント設計
**選択決定**: ETCカード番号を中心としたAPI構造
**Rationale**:
- ryohi_sub_calとの連携を考慮
- etc_numによる効率的なデータ取得
- 車両管理システムとの統合容易性

**新規エンドポイント**:
- `GET /api/etc/mapping/by-etc-num/{etc_num}`: ETCカード別マッピング一覧
- `POST /api/etc/auto-match/by-etc-num`: ETCカード番号指定自動マッチング
- `GET /api/etc/vehicles/etc-mapping`: 車両-ETCカード関連付け一覧

**結論**: Phase 1 (テスト修正、デュアルDB、基本モデル) → Phase 2 (etc_num活用マッチング機能) → Phase 3 (ryohi_sub_cal連携API完成)の順序で実装を進める。現在のハッシュベースシステムを基盤として、ryohi_sub_calの知見を活用し、etc_numフィールドを中心とした精密なdtako_row_idマッチング機能を実現する。
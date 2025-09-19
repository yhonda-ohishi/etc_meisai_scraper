# 明確化事項と回答（Goモジュール版）

**機能**: ETC明細Goモジュール完成
**作成日**: 2025-09-18
**更新日**: 2025-09-18

## 重要な前提
**このプロジェクトは他のプロジェクトからGoモジュールとして呼び出されるライブラリです。**
HTTPルーティングは呼び出し側が実装するため、このモジュールの責務ではありません。

## 明確化済み事項

### 1. 認証戦略
**質問**: APIアクセスにはどの認証方法を使用すべきか？
**回答**: localで実行するので、認証は必要ない
**決定**: 認証なし（呼び出し側の責務）

### 2. データ保持ポリシー
**質問**: ダウンロードしたETCデータはどのくらいの期間保存すべきか？
**回答**: 7回分
**決定**: 最新7回分のダウンロードデータを保持する機能を提供

### 3. パフォーマンス要件
**質問**: ターゲットレスポンス時間と同時ユーザー制限は何か？
**回答**: 特になし
**決定**: Goモジュールとして効率的な実装を提供

### 4. レート制限
**質問**: ユーザー/時間あたりのAPIリクエストに制限を設けるべきか？
**回答**: 特になし
**決定**: レート制限は呼び出し側で実装

### 5. セキュリティ要件
**質問**: データ保存と送信に関する特定のセキュリティ要件はあるか？
**回答**: 特になし
**決定**: セキュリティは呼び出し側の責務

### 6. 未実装のGoモジュール公開関数

**現在公開されている主要な関数/型**:

#### メインクライアント (api.go)
- ✅ `NewETCClient(*ClientConfig) *ETCClient` - ETCクライアントの作成
- ✅ `(*ETCClient) DownloadETCData(accounts []SimpleAccount, from, to time.Time) ([]DownloadResult, error)` - 複数アカウントのデータダウンロード
- ✅ `(*ETCClient) DownloadETCDataSingle(userID, password string, from, to time.Time) (*DownloadResult, error)` - 単一アカウントのダウンロード
- ✅ `ParseETCCSV(csvPath string) ([]ETCMeisai, error)` - CSVファイルの解析
- ✅ `LoadCorporateAccounts() ([]SimpleAccount, error)` - 法人アカウント読み込み
- ✅ `LoadPersonalAccounts() ([]SimpleAccount, error)` - 個人アカウント読み込み

#### モジュール初期化 (module.go)
- ✅ `NewModule(db *sql.DB) *Module` - モジュールの初期化
- ✅ `InitializeWithRouter(db *sql.DB, r *chi.Mux) (*Module, error)` - ルーター付き初期化
- ✅ `(*Module) GetHandler() *ETCHandler` - ハンドラー取得
- ✅ `(*Module) GetService() *ETCService` - サービス取得
- ✅ `(*Module) GetRepository() *ETCRepository` - リポジトリ取得

#### HTTPハンドラー（呼び出し側で利用）
各種ハンドラー関数は公開されているが、呼び出し側でルーティングに組み込む

**追加実装が必要な可能性のある公開関数**:
- ❌ `(*ETCClient) ExportToExcel(records []ETCMeisai, path string) error` - Excel出力
- ❌ `(*ETCClient) ExportToPDF(records []ETCMeisai, path string) error` - PDF出力
- ❌ `(*ETCClient) GetStatistics(records []ETCMeisai) *Statistics` - 統計情報生成
- ❌ `(*ETCClient) FilterRecords(records []ETCMeisai, criteria FilterCriteria) []ETCMeisai` - レコードフィルタリング
- ❌ `(*ETCClient) ValidateAccounts(accounts []SimpleAccount) []AccountValidationResult` - アカウント検証
- ❌ `(*ETCClient) MergeRecords(records1, records2 []ETCMeisai) []ETCMeisai` - レコードマージ

### 7. APIバージョニング
**質問**: 後方互換性のためのAPIバージョニングはどのように処理すべきか？
**回答**: 提案して

**提案**: Goモジュールのセマンティックバージョニング
```
github.com/yhonda-ohishi/etc_meisai (v0.x.x - 現在)
github.com/yhonda-ohishi/etc_meisai/v2 (v2.x.x - メジャーバージョンアップ時)
```

**実装方法**:
1. 現在のv0.0.16は開発版として維持
2. v1.0.0リリース時に安定版API確定
3. 破壊的変更が必要な場合はv2として別モジュール

**go.modでの利用例**:
```go
require (
    github.com/yhonda-ohishi/etc_meisai v1.0.0
)
```

## まとめ

Goモジュールとしての実装方針：
- HTTPハンドラーは提供するが、ルーティングは呼び出し側
- 認証・セキュリティは呼び出し側の責務
- データ処理のコア機能に集中
- セマンティックバージョニングで後方互換性管理
- 追加機能はExport機能や統計処理などのユーティリティ関数
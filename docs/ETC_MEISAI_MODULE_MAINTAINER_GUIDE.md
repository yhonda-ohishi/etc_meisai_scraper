# ETC Meisai Module Maintainer Guide

## 概要

このドキュメントは、`github.com/yhonda-ohishi/etc_meisai`モジュールのメンテナンス方法について説明します。

## モジュール構成

```
etc_meisai/
├── api.go                    # 外部公開API
├── api_handlers.go           # HTTP APIハンドラー (Swagger対応)
├── cmd/
│   ├── server/              # スタンドアロンサーバー
│   └── scraper/             # スクレイパーCLI
├── config/
│   ├── accounts.go          # アカウント設定
│   ├── simple_accounts.go   # シンプルアカウント管理
│   └── database.go          # DB設定
├── scraper/
│   ├── etc_scraper_actual.go # メインスクレイパー実装
│   ├── etc_scraper.go       # スクレイパー設定
│   └── multi_account_scraper.go # 複数アカウント処理
├── parser/
│   └── csv_parser.go         # CSVパーサー
├── models/
│   └── models.go             # データモデル
├── docs/
│   ├── swagger.yaml          # OpenAPI仕様
│   ├── swagger.html          # Swagger UI
│   └── ETC_MEISAI_MODULE_MAINTAINER_GUIDE.md
└── README.md
```

## バージョン管理

### 現在のバージョン
- **v0.0.1** - 初回リリース (2025-09-15)

### バージョニング規則
セマンティックバージョニング (SemVer) を採用：
- **MAJOR**: 破壊的変更
- **MINOR**: 後方互換性のある機能追加
- **PATCH**: バグ修正

### リリース手順

1. **変更をコミット**
```bash
git add .
git commit -m "機能追加/修正内容"
```

2. **タグを作成**
```bash
git tag -a v0.0.2 -m "リリースノート"
```

3. **プッシュ**
```bash
git push origin master
git push origin v0.0.2
```

## API仕様

### 公開関数（api.go）

#### ETCClient
```go
// クライアント作成
func NewETCClient(config *ClientConfig) *ETCClient

// 複数アカウントダウンロード
func (c *ETCClient) DownloadETCData(accounts []config.SimpleAccount, fromDate, toDate time.Time) ([]DownloadResult, error)

// 単一アカウントダウンロード
func (c *ETCClient) DownloadETCDataSingle(userID, password string, fromDate, toDate time.Time) (*DownloadResult, error)
```

#### ユーティリティ関数
```go
// CSVパース
func ParseETCCSV(csvPath string) ([]models.ETCMeisai, error)

// アカウント読み込み
func LoadCorporateAccounts() ([]config.SimpleAccount, error)
func LoadPersonalAccounts() ([]config.SimpleAccount, error)
```

### Swagger API仕様

Swagger UIは `docs/swagger.html` で確認可能。

主要エンドポイント：
- `POST /api/etc/download` - 明細ダウンロード
- `POST /api/etc/download-single` - 単一アカウントダウンロード
- `POST /api/etc/parse-csv` - CSVパース
- `GET /api/etc/meisai` - 明細取得
- `GET /api/etc/summary` - サマリー取得

## スクレイパーの保守

### ページ構造の違い

ETCサイトはアカウントによってページ構造が異なります：

#### Type 1: ohishiexp（標準構造）
- 日付選択ドロップダウンが直接表示
- fromYYYY, fromMM, fromDD セレクタ使用

#### Type 2: ohishiexp1（条件表示型）
- 「検索条件」クリックで日付選択表示
- 月ボタンによる選択も可能

### デバッグ方法

1. **ヘッドレスモードを無効化**
```go
config := &ScraperConfig{
    Headless: false,  // ブラウザを表示
}
```

2. **スクリーンショット有効化**
```go
s.takeScreenshot("debug_point")
```

3. **詳細ログ出力**
```go
log.Printf("DEBUG: %v", variable)
```

### よくある問題と対処

#### 問題: ログイン失敗
**原因**: セレクタ変更、サイトメンテナンス
**対処**:
```go
// etc_scraper_actual.go の Login() 内のセレクタを確認
userSelectors := []string{
    "input[name='usrid']",
    "input[name='userId']",
    // 新しいセレクタを追加
}
```

#### 問題: 日付選択失敗
**原因**: アカウントタイプの違い
**対処**:
```go
// SearchAndDownloadCSV() 内で構造検出ロジックを確認
if count, _ := s.page.Locator("select[name='fromYYYY']").Count(); count > 0 {
    hasDateSelects = true
}
```

#### 問題: CSVダウンロード失敗
**原因**: ボタンセレクタ変更
**対処**:
```go
downloadSelectors := []string{
    "input[type='button'][value='利用明細ＣＳＶ出力']",
    // 新しいセレクタを追加
}
```

## テスト

### ユニットテスト実行
```bash
go test ./...
```

### 統合テスト
```bash
# 単一アカウントテスト
go run test_ohishiexp1_only.go

# 複数アカウントテスト
go run test_multi_accounts.go
```

### テストアカウント設定
`.env.example`を`.env`にコピーして設定：
```env
ETC_CORP_ACCOUNTS=["test1:pass1","test2:pass2"]
```

## 依存関係

### 主要依存パッケージ
- `playwright-go` - ブラウザ自動化
- `golang.org/x/text` - 文字エンコーディング
- `joho/godotenv` - 環境変数管理
- `go-chi/chi` - HTTPルーター
- `go-sql-driver/mysql` - MySQLドライバー

### 依存関係更新
```bash
go get -u ./...
go mod tidy
```

## トラブルシューティング

### Playwright関連

#### ブラウザのインストール
```bash
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install
```

#### ブラウザが起動しない
```bash
# 依存関係確認
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install-deps
```

### 文字化け問題
CSVファイルはShift-JISエンコーディング：
```go
reader := transform.NewReader(file, japanese.ShiftJIS.NewDecoder())
```

## セキュリティ

### 認証情報の管理
- 環境変数使用を推奨
- `.env`ファイルはGit管理外
- 本番環境ではシークレット管理システム使用

### アクセス制限
```go
// レート制限の実装例
time.Sleep(5 * time.Second) // アカウント間の待機
```

## パフォーマンス最適化

### 並列処理
```go
// multi_account_scraper.go
concurrent: 2, // 同時実行数を制限
```

### メモリ管理
```go
defer s.Close() // リソースの確実な解放
```

### タイムアウト設定
```go
Timeout: 30000, // 30秒
```

## モニタリング

### ログ出力
```go
log.Printf("[%s] %s", level, message)
```

### メトリクス収集ポイント
- ログイン成功率
- ダウンロード成功率
- 処理時間
- エラー率

## リリースチェックリスト

- [ ] テスト実行（単体・統合）
- [ ] README.md更新
- [ ] swagger.yaml更新
- [ ] バージョン番号更新
- [ ] CHANGELOGエントリ追加
- [ ] タグ作成
- [ ] GitHubへプッシュ

## 連絡先

### 問題報告
GitHubのIssuesで報告

### 開発者
- リポジトリ: https://github.com/yhonda-ohishi/etc_meisai

## 更新履歴

### v0.0.1 (2025-09-15)
- 初回リリース
- スクレイピング機能実装
- 複数アカウント対応
- Swagger API仕様追加

## ライセンス

内部使用のみ。無断での外部公開・使用を禁止。
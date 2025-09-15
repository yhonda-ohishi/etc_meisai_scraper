# ETC Meisai Module

ETCメイサイ（明細）データを管理・自動取得するGoモジュール

## 概要

このモジュールは、ETC利用明細データの自動取得と管理機能を提供します。ryohi_sub_cal2からGitHub経由で呼び出されることを想定しています。

## 機能

### スクレイピング機能
- ETC利用照会サービスからの自動ログイン
- 複数アカウント対応（法人・個人）
- 日付範囲指定でのCSVダウンロード
- アカウントタイプに応じた適応的ページ処理
- Shift-JISエンコーディング対応

### データ管理機能
- ETC明細データのインポート
- 運行番号（UnkoNo）による明細検索
- 日付範囲による明細検索
- 明細データの一括インポート
- 利用状況のサマリー取得

## インストール

```bash
go get github.com/yhonda-ohishi/etc_meisai@v0.0.8
```

## 使用方法

### ライブラリとして使用（スクレイピング機能）

```go
import (
    etc "github.com/yhonda-ohishi/etc_meisai"
)

// ETCクライアント作成
client := etc.NewETCClient(&etc.ClientConfig{
    DownloadPath: "./downloads",
    Headless:     true,
    Timeout:      30000,
    RetryCount:   3,
})

// アカウント情報（環境変数から読み込みも可能）
accounts, err := etc.LoadCorporateAccounts()

// 日付範囲指定
fromDate := time.Date(2025, 8, 1, 0, 0, 0, 0, time.Local)
toDate := time.Now()

// ETC明細をダウンロード
results, err := client.DownloadETCData(accounts, fromDate, toDate)

// 単一アカウントでのダウンロード
result, err := client.DownloadETCDataSingle(userID, password, fromDate, toDate)

// 既存のCSVファイルをパース
records, err := etc.ParseETCCSV("path/to/file.csv")
```

### go.modでの設定（ローカル開発時）

```go
module your-module

require (
    github.com/yhonda-ohishi/etc_meisai v0.0.1
)

// ローカル開発時は以下を追加
replace github.com/yhonda-ohishi/etc_meisai => ../etc_meisai
```

### データベース連携モジュールとして使用

```go
import (
    "database/sql"
    "github.com/go-chi/chi/v5"
    "github.com/yhonda-ohishi/etc_meisai"
)

// データベース接続
db, err := sql.Open("mysql", dsn)

// ルーター作成
r := chi.NewRouter()

// モジュール初期化
module, err := etc_meisai.InitializeWithRouter(db, r)
```

### スタンドアロンサーバーとして実行

```bash
# 環境変数設定
cp .env.example .env
# .envファイルを編集

# サーバー起動
go run cmd/server/main.go
```

## API エンドポイント

### 基本エンドポイント
- `GET /health` - ヘルスチェック
- `GET /api/etc/accounts` - 利用可能なアカウント一覧（パスワード非表示）
- `POST /api/etc/import` - データインポート
- `POST /api/etc/bulk-import` - 一括インポート
- `GET /api/etc/meisai` - 明細取得
- `POST /api/etc/meisai` - 明細作成
- `GET /api/etc/meisai/{id}` - ID指定で明細取得
- `GET /api/etc/summary` - サマリー取得

### ダウンロードエンドポイント

#### `/api/etc/download` - 複数アカウント一括ダウンロード

複数のETCアカウントから明細を一括でダウンロードします。

**アカウント設定方法：**

1. **リクエストボディで指定**
```bash
curl -X POST http://localhost:8080/api/etc/download \
  -H "Content-Type: application/json" \
  -d '{
    "accounts": [
      {"user_id": "ohishiexp", "password": "pass1"},
      {"user_id": "ohishiexp1", "password": "pass2"}
    ],
    "from_date": "2025-08-01",
    "to_date": "2025-09-15"
  }'
```

2. **環境変数を使用**（accountsパラメータを省略）
```bash
# 環境変数を設定
export ETC_CORP_ACCOUNTS="ohishiexp:pass1,ohishiexp1:pass2"

# 利用可能なアカウントを確認
curl http://localhost:8080/api/etc/accounts

# APIを呼び出し
curl -X POST http://localhost:8080/api/etc/download \
  -H "Content-Type: application/json" \
  -d '{
    "from_date": "2025-08-01",
    "to_date": "2025-09-15"
  }'
```

3. **カスタム設定付き**
```bash
curl -X POST http://localhost:8080/api/etc/download \
  -H "Content-Type: application/json" \
  -d '{
    "from_date": "2025-08-01",
    "to_date": "2025-09-15",
    "config": {
      "download_path": "./custom_downloads",
      "headless": false,
      "timeout": 60000,
      "retry_count": 5
    }
  }'
```

#### `/api/etc/download-single` - 単一アカウントダウンロード

単一のETCアカウントから明細をダウンロードします。

**使用例：**
```bash
curl -X POST http://localhost:8080/api/etc/download-single \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "ohishiexp",
    "password": "password123",
    "from_date": "2025-08-01",
    "to_date": "2025-09-15"
  }'
```

**注意事項：**
- `user_id` と `password` は必須パラメータです
- このエンドポイントは環境変数を使用しません
- アカウントタイプ（ohishiexp/ohishiexp1）によってページ構造が異なるため、正しいユーザーIDを指定してください

## データ構造

### ETCMeisai

ETC明細データの主要構造体：

- `UnkoNo` - 運行番号
- `Date` - 日付
- `Time` - 時刻
- `ICEntry` - IC入口
- `ICExit` - IC出口
- `VehicleNo` - 車両番号
- `CardNo` - ETCカード番号
- `Amount` - 利用金額
- `DiscountAmount` - 割引金額
- `TotalAmount` - 請求金額
- `Distance` - 走行距離

## データベース

MySQLを使用。スキーマは`schema.sql`を参照。

## 環境変数

### スクレイピング用
- `ETC_CORP_ACCOUNTS` - 法人アカウント情報（JSON配列形式）
  ```
  ETC_CORP_ACCOUNTS=["user1:pass1","user2:pass2"]
  ```
- `ETC_PERSONAL_ACCOUNTS` - 個人アカウント情報（JSON配列形式）
- `ETC_USER_ID` - 単一アカウントのユーザーID
- `ETC_PASSWORD` - 単一アカウントのパスワード

### データベース用
- `DB_HOST` - データベースホスト
- `DB_PORT` - データベースポート
- `DB_USER` - データベースユーザー
- `DB_PASSWORD` - データベースパスワード
- `DB_NAME` - データベース名
- `SERVER_PORT` - サーバーポート（デフォルト: 8080）

## ライセンス

内部使用のみ
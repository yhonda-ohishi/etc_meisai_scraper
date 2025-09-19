# クイックスタートガイド: ETC明細スクレイピング＆DBインポート（シンプル版）

**前提条件**: Go 1.22+, MySQL 8.0+, Playwright

## 1. セットアップ (5分)

### 環境変数設定
```bash
# .env ファイル作成
cat > .env << EOF
DB_HOST=localhost
DB_PORT=3307
DB_USER=root
DB_PASSWORD=yourpassword
DB_NAME=etc_meisai

ETC_CORP_USER=corporate_user
ETC_CORP_PASS=corporate_pass
ETC_PERSONAL_USER=personal_user
ETC_PERSONAL_PASS=personal_pass
EOF
```

### データベース準備
```bash
# データベース作成
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS etc_meisai;"

# テーブル作成
mysql -u root -p etc_meisai < schema.sql
```

### 依存関係インストール
```bash
go mod download
go mod tidy

# Playwright インストール
go run github.com/playwright-community/playwright-go/cmd/playwright install chromium
```

## 2. ビルド (1分)

```bash
# CLIツールをビルド
go build -o etc_scraper cmd/scraper/main.go
```

## 3. 基本的な使い方

### 単一アカウントのスクレイピング
```bash
# 法人アカウントで今月のデータ取得
./etc_scraper --account corporate --month current

# 個人アカウントで先月のデータ取得
./etc_scraper --account personal --month last
```

### 日付範囲指定
```bash
# 特定期間のデータ取得
./etc_scraper --account corporate --from 2025-01-01 --to 2025-01-31
```

### 全アカウント処理
```bash
# 環境変数に設定された全アカウントを処理
./etc_scraper --all --month current
```

## 4. テスト実行

```bash
# 統合テストの実行
go test ./tests/integration -v

# 特定のテストのみ実行
go test ./tests/integration -run TestScraping
```

## 5. 実行例とログ

```bash
$ ./etc_scraper --account corporate --month current

2025-09-19 10:00:00 [INFO] Starting ETC scraper...
2025-09-19 10:00:01 [INFO] Logging in to corporate account...
2025-09-19 10:00:05 [INFO] Login successful
2025-09-19 10:00:06 [INFO] Downloading CSV for 2025-09-01 to 2025-09-30...
2025-09-19 10:00:15 [INFO] CSV downloaded: downloads/corporate_202509.csv
2025-09-19 10:00:16 [INFO] Parsing CSV file...
2025-09-19 10:00:17 [INFO] Found 150 records
2025-09-19 10:00:18 [INFO] Importing to database...
2025-09-19 10:00:19 [INFO] Imported: 145, Skipped: 5 (duplicates)
2025-09-19 10:00:20 [INFO] Process completed successfully
```

## 6. トラブルシューティング

### ログイン失敗
```bash
# 環境変数確認
echo $ETC_CORP_USER
echo $ETC_CORP_PASS

# Playwrightブラウザ確認
playwright install chromium
```

### DB接続エラー
```bash
# MySQL接続テスト
mysql -h localhost -P 3307 -u root -p -e "SELECT VERSION();"

# データベース存在確認
mysql -u root -p -e "SHOW DATABASES LIKE 'etc_meisai';"
```

### 重複エラー
```bash
# 重複データ確認
mysql -u root -p etc_meisai -e "
  SELECT usage_date, entry_ic, exit_ic, COUNT(*) as cnt
  FROM etc_meisai
  GROUP BY usage_date, entry_ic, exit_ic
  HAVING cnt > 1;
"
```

## 7. cron設定例

```bash
# 毎日深夜2時に前日のデータを取得
0 2 * * * /path/to/etc_scraper --all --day yesterday >> /var/log/etc_scraper.log 2>&1

# 毎月1日に先月のデータを取得
0 0 1 * * /path/to/etc_scraper --all --month last >> /var/log/etc_scraper.log 2>&1
```

## 8. データ確認

```sql
-- インポートされたデータの確認
SELECT COUNT(*) FROM etc_meisai;

-- 日付別集計
SELECT usage_date, COUNT(*) as count, SUM(toll_amount) as total
FROM etc_meisai
GROUP BY usage_date
ORDER BY usage_date DESC
LIMIT 10;

-- アカウント別集計
SELECT account_type, COUNT(*) as count
FROM etc_meisai
GROUP BY account_type;
```

## 9. クリーンアップ

```bash
# 古いCSVファイル削除
find downloads/ -name "*.csv" -mtime +30 -delete

# 古いログ削除
find logs/ -name "*.log" -mtime +30 -delete
```

## 10. パフォーマンス目安

- **スクレイピング**: 1アカウント約3-5分
- **CSV解析**: 1000レコード/秒
- **DB保存**: 500レコード/秒
- **メモリ使用**: 約200MB

---
*Simplified Quickstart - Focus on essential operations*
# クイックスタートガイド: ETC明細APIシステム

**最終更新**: 2025-09-18
**対象**: 開発者・QA・システム管理者

## 🚀 概要

このガイドでは、ETC明細APIシステムの開発環境セットアップから基本的なテスト実行まで、最短時間で開始できる手順を提供します。**重要**: このプロジェクトの目標はテストを通すことであり、コンパイルエラーの修正だけではありません。

## 📋 前提条件

### 必須環境
- **Go**: 1.22以上
- **MySQL**: 8.0以上（ローカル開発用）
- **Git**: 最新版
- **OS**: Windows 10/11 (開発環境) または Linux (本番環境)

### 推奨ツール
- **IDE**: VS Code + Go Extension
- **API テストツール**: Postman または curl
- **データベースツール**: MySQL Workbench または DBeaver

## 🔧 環境セットアップ

### 1. リポジトリのクローンと依存関係インストール

```bash
# プロジェクトディレクトリに移動
cd C:/go/etc_meisai

# Go モジュールの依存関係を更新
go mod download
go mod tidy

# ビルドテスト（コンパイルエラーがないことを確認）
go build -o /dev/null ./...
```

### 2. データベース設定

#### ローカルMySQL設定
```bash
# MySQL接続確認
mysql -u root -p -e "SELECT VERSION();"

# データベース作成（必要に応じて）
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS etc_meisai_local CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

#### 環境変数設定 (.env)
```bash
# .envファイルの確認と必要に応じて設定
cat .env

# 基本的な設定例
DB_HOST=localhost
DB_PORT=3307
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=etc_meisai_local

# 本番データベース設定
PROD_DB_HOST=172.18.21.35
PROD_DB_PORT=3306
PROD_DB_USER=pbi
PROD_DB_PASSWORD=kikuraku
PROD_DB_NAME=production_db
```

## 🧪 テスト実行手順（最重要）

### 1. 現在のテスト状況確認

```bash
# 全体のテスト実行（現在の失敗状況を確認）
go test ./...

# 統合テスト実行（失敗している主要テスト）
go test ./tests/integration/...

# 詳細なテスト結果表示
go test -v ./tests/integration/...
```

### 2. 失敗しているテストの特定

```bash
# ハッシュインポートテストのみ実行
go test -v ./tests/integration/ -run TestHashImportHandler

# 重複検出テストのみ実行
go test -v ./tests/integration/ -run TestDuplicateDetection

# 変更検出テストのみ実行
go test -v ./tests/integration/ -run TestChangeDetection
```

### 3. テスト失敗の根本原因分析

現在のテスト失敗は以下の問題が原因：

#### A. CSVパーサーの日付フォーマット不整合
```bash
# エラー例：parsing time "2025/09/01" as "06/01/02"
# 原因：パーサーが期待する形式と実際のテストデータの形式が異なる
```

#### B. 文字エンコーディング問題
```bash
# エラー例：invalid UTF-8 encoding
# 原因：Shift-JIS→UTF-8変換がテストデータに不適切に適用されている
```

#### C. テストデータの検証
```bash
# 現在のテストデータ確認
head -5 tests/integration/testdata/sample.csv

# 実際のCSVファイル確認（比較用）
head -5 downloads/202509151527.csv
```

## 🔍 テスト失敗の具体的な修正アプローチ

### 1. CSVパーサーの改善

**問題**: 複数の日付フォーマットに対応していない
**修正目標**: `"06/01/02"`, `"2025/09/01"`, `"25/07/30"` 全てに対応

```go
// parser/etc_csv_parser.go の改善例
dateFormats := []string{
    "06/01/02",      // 現在の形式
    "2006/01/02",    // フル年形式
    "06/1/2",        // 0埋めなし
    "2006/1/2",      // フル年+0埋めなし
}
```

### 2. テストデータの修正

**問題**: テストデータが実際のCSV形式と一致していない
**修正目標**: 実際のETCサイトCSVファイル形式に準拠

```bash
# テストデータの形式確認と修正
# - Shift-JISエンコーディング
# - 適切な日付フォーマット
# - 正しい列数（15列）
```

### 3. エンコーディング処理の改善

**問題**: UTF-8テストデータに対する不適切なShift-JIS変換
**修正目標**: エンコーディング自動検出機能

## 🏃‍♂️ サーバー起動とAPI確認

### 1. 開発サーバー起動

```bash
# 開発サーバー起動
go run cmd/server/main.go

# サーバーの起動確認
curl http://localhost:8080/health
```

### 2. 基本的なAPI確認

#### ハッシュインポート機能
```bash
# ハッシュベースインポートのテスト
curl -X POST http://localhost:8080/api/etc/import/hash \
  -H "Content-Type: application/json" \
  -d '{
    "csv_path": "./downloads/202509151527.csv",
    "options": {
      "validate_only": true
    }
  }'
```

#### ハッシュ統計情報の確認
```bash
# ハッシュインデックス統計
curl http://localhost:8080/api/etc/hash/stats
```

## 🔧 トラブルシューティング

### テスト失敗パターンと解決方法

#### 1. 日付パーシングエラー
```
Error: parsing time "2025/09/01" as "06/01/02": cannot parse "25/09/01" as "/"
```

**解決方法**:
1. `parser/etc_csv_parser.go` で複数フォーマット対応を実装
2. テストデータの日付形式を統一
3. フォールバック機能の実装

#### 2. レコード処理数ゼロ
```
Should have processed some records
Import results: Added=0, Updated=0, Duplicates=0
```

**解決方法**:
1. CSVパーシングが正常に動作しているか確認
2. テストデータの形式とパーサーの期待形式の整合性確認
3. エラーログの詳細確認

#### 3. 重複検出失敗
```
Should have detected duplicates within the file
```

**解決方法**:
1. ハッシュ計算が正常に実行されているか確認
2. テストデータに実際の重複が含まれているか確認
3. ハッシュインデックスの状態確認

## 📊 成功の測定指標

### テスト成功の条件
- [ ] `TestHashImportHandler`: ProcessedCount > 0
- [ ] `TestDuplicateDetection`: 重複検出数 > 0
- [ ] `TestChangeDetection`: 変更検出数 > 0
- [ ] 全統合テストがPASS
- [ ] コンパイルエラーなし

### API動作確認
- [ ] ヘルスチェックAPI応答
- [ ] ハッシュインポートAPI正常動作
- [ ] 統計情報API正常応答
- [ ] エラーハンドリング適切

## 🎯 完了チェックリスト

### セットアップ完了確認
- [ ] Go 1.22+がインストール済み
- [ ] MySQL接続が成功
- [ ] 依存関係のインストール完了
- [ ] `.env`ファイル設定済み

### **重要: テスト成功確認**
- [ ] 全統合テストがPASS
- [ ] CSVパーシング正常動作
- [ ] ハッシュ計算正常動作
- [ ] 重複検出機能動作
- [ ] 変更検出機能動作

### 基本動作確認
- [ ] コンパイルエラーなし (`go build`)
- [ ] サーバー起動成功
- [ ] ヘルスチェックAPI応答
- [ ] データベース接続確認

## 📞 サポート

### 重要な注意点
**このプロジェクトの成功の定義は「テストが通ること」です。コンパイルが通るだけでは不十分です。**

### 追加情報
- **仕様書**: `specs/003-api-project-git/`
- **API仕様**: `specs/003-api-project-git/contracts/`
- **データモデル**: `specs/003-api-project-git/data-model.md`
- **技術調査結果**: `specs/003-api-project-git/research.md`

### 次のステップ
1. **テスト修正**: 失敗テストの根本原因修正
2. **CSVパーサー改善**: 複数フォーマット対応
3. **統合テスト拡張**: 新しいテストケース追加
4. **API完成**: 全エンドポイントの動作確認

---

**テストの成功を目指して頑張りましょう！** このガイドで問題が解決しない場合は、開発チームまでお問い合わせください。
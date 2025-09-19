# Phase 0: Research - ETC明細スクレイピング＆DBインポート（シンプル版）

**Date**: 2025-09-19 | **Status**: Complete | **Branch**: `003-api-project-git-simple`

## Executive Summary

現在のプロジェクトを分析し、スクレイピングとDBインポートのコア機能のみに簡素化する方針を確立。不要な機能を削除し、実装とメンテナンスを容易にする。

## 1. 現在の実装状況分析

### 既存のスクレイピング機能
- **実装済み**: `scraper/scraper.go`にPlaywright-goベースの実装
- **ログイン処理**: Corporate/Personal両対応
- **CSV取得**: ダウンロード機能実装済み
- **保存先**: `downloads/`ディレクトリ

### 既存のDB機能
- **MySQL接続**: `config/database.go`で実装
- **Repository**: `repositories/etc_repository.go`でCRUD操作
- **モデル**: `models/etc_meisai.go`定義済み

## 2. 削除対象の機能

### APIレイヤー（削除）
- `api/handlers/`ディレクトリ全体
- `routes.go`, `registry.go`
- HTTPサーバー関連コード
- Swagger/OpenAPI定義

### 複雑な統合機能（削除）
- 本番DB接続（`config/prod_database.go`）
- マッチング機能（`services/matching_service.go`）
- 非同期ジョブ管理
- 統計・レポート機能

### 不要な依存関係（削除）
- Chi router
- Swagger関連
- 本番DB接続ライブラリ

## 3. 保持・強化する機能

### コアスクレイピング
```go
// 保持する主要関数
func (s *Scraper) Login() error
func (s *Scraper) DownloadCSV(from, to time.Time) (string, error)
func (s *Scraper) ParseCSV(filepath string) ([]ETCMeisai, error)
```

### シンプルなDB操作
```go
// 保持するDB操作
func (r *ETCRepository) BulkInsert(records []ETCMeisai) error
func (r *ETCRepository) CheckDuplicate(record ETCMeisai) bool
```

## 4. 技術選択の決定

### Decision: Playwright-go継続使用
**選択**: 現在のPlaywright-go実装を維持
**理由**:
- 既に動作確認済み
- JavaScript実行が必要なサイトに対応
- 安定している

### Decision: MySQL単一DB
**選択**: ローカルMySQLのみ使用
**理由**:
- 設定がシンプル
- 本番DB統合の複雑性を排除
- トランザクション処理で十分

### Decision: CLI実行
**選択**: コマンドラインツールとして実装
**理由**:
- HTTPサーバー不要
- cronやタスクスケジューラで自動化可能
- デバッグが容易

## 5. 簡素化されたアーキテクチャ

```
CLI Entry Point
    ↓
Scraper Service
    ↓
CSV Download
    ↓
Parse CSV
    ↓
DB Import (with duplicate check)
    ↓
Complete
```

## 6. 環境変数設計（簡素化）

```bash
# 必須環境変数のみ
DB_HOST=localhost
DB_PORT=3307
DB_USER=root
DB_PASSWORD=password
DB_NAME=etc_meisai

# ETCアカウント
ETC_CORP_USER=user1
ETC_CORP_PASS=pass1
ETC_PERSONAL_USER=user2
ETC_PERSONAL_PASS=pass2
```

## 7. パフォーマンス要件（簡素化）

- **スクレイピング**: 1アカウント5分以内
- **DB保存**: 1000レコード/秒
- **メモリ使用**: 500MB以下
- **エラー処理**: リトライ3回

## 8. テスト戦略（簡素化）

### 統合テストのみ
- スクレイピング成功テスト
- DB保存成功テスト
- 重複検出テスト
- エラーハンドリングテスト

### モックなし
- 実際のDBを使用（テスト用DB）
- 実際のCSVファイルを使用

## 9. 実装優先順位

1. **クリーンアップ** [最優先]
   - 不要なAPIコード削除
   - 依存関係整理
   - ディレクトリ構造簡素化

2. **コア機能整備**
   - スクレイピング関数整理
   - DB操作簡素化
   - エラーハンドリング

3. **CLI実装**
   - コマンドライン引数処理
   - 実行フロー実装
   - ログ出力

4. **テスト作成**
   - 統合テスト実装
   - 実行確認

## 次のステップ

Phase 1に進み、簡素化されたデータモデルとクイックスタートガイドを作成する。

---
*Research completed: シンプル化方針確立*
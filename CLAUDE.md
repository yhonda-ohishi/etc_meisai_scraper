# CLAUDE.md - ETC明細Goモジュール プロジェクトコンチE��スチE

## プロジェクト概要E
ETC明細チE�EタをWebスクレイピングで取得し、データベ�Eスに保存するGoモジュール、E
db-handler-serverパターンに従ったハンドラー実裁E��の移行中、E

## 技術スタチE��
- **言誁E*: Go 1.21+
- **フレームワーク**: gRPC + grpc-gateway (chi から移衁E
- **Protocol Buffers**: API定義とコード生戁E
- **チE�Eタベ�Eス**: db_service (Fiber実裁E via gRPC
- **通信**: gRPC (server_repo統吁E
- **スクレイピング**: Playwright-go
- **チE��チE��ング**: testify/mock, table-driven tests (100%カバレチE��目樁E
- **依存管琁E*: Go Modules, buf (Protocol Buffers)
- **アーキチE��チャ**: etc_meisai ↁEgRPC ↁEdb_service (Fiber)

## プロジェクト構造
```
etc_meisai/
├── src/
━E  ├── proto/           # Protocol Buffers定義
━E  ├── pb/              # 生�EされたgRPCコーチE
━E  ├── grpc/            # gRPCサーバ�E実裁E
━E  ├── services/        # ビジネスロジチE��層
━E  ├── repositories/    # チE�Eタアクセス層
━E  ├── models/          # GORMチE�EタモチE��
━E  └── adapters/        # 互換性レイヤー
├── handlers/            # HTTPハンドラー�E�レガシー�E�E
├── parser/              # CSV解极E
├── config/              # 設定管琁E
└── downloads/           # CSVファイル保存�E
```

## 主要機�E
1. **ETC明細ダウンローチE*: 褁E��アカウント対応、E��同期処琁E
2. **チE�Eタ処琁E*: CSV解析、データ変換、バルク保孁E
3. **マッピング管琁E*: ETC明細とチE��タコチE�Eタの関連付け�E�Etc_num活用�E�E
4. **進捗追跡**: リアルタイム進捗通知�E�ESE対応！E
5. **自動�EチE��ング**: dtako_row_idとの精寁E�EチE��ング

## 最近�E変更 (v0.0.19 - gRPC統吁E
- **gRPC移衁E*: go-chiからgRPC + grpc-gatewayへの移行完亁E
- **Protocol Buffers**: API定義をprotoファイルで一允E��琁E
- **Swagger統吁E*: OpenAPI仕様�E自動生成とSwagger UI統吁E
- **server_repo統吁E*: 統一されたサービス登録とルーチE��ング

## 開発中の機�E (統合フェーズ)
- **モチE��統吁E*: db_serviceのGORMモチE�� + 互換性レイヤー実裁E
- **Repository統吁E*: 統吁Eepository interface + gRPCクライアント実裁E
- **サービス統吁E*: 既孁Eervices/のgRPCクライアント化

## スコープ外�E機�E
- Excel/PDF エクスポ�Eト機�E
- 統計情報生�E機�E
- キャチE��ュ機�E�E�ユーザー要求により除外！E

## パフォーマンス目樁E
- CSVファイル1丁E��を5秒以冁E��処琁E
- メモリ使用釁E00MB以丁E
- 同時ダウンローチEアカウントまで
- チE��ト実行時閁E0秒以冁E���EチE��トスイート！E
- チE��トカバレチE��100%維持E

## 統合アーキチE��チャ (db_service via gRPC)
```
etc_meisai:
  HTTP API (handlers/) ↁEService Layer (services/) ↁEgRPC Client
                                                           ↁE
                                                     [gRPC Protocol]
                                                           ↁE
db_service (Fiber):
  gRPC Server ↁERepository (GORM) ↁEMySQL Database
```

### 統合コンポ�EネンチE
- **統吁Eepository**: `src/repositories/etc_integrated_repo.go`
- **gRPCクライアンチE*: `src/clients/db_service_client.go`
- **互換性レイヤー**: `src/adapters/etc_compat_adapter.go`
- **統合テスチE*: `tests/contract/`, `tests/integration/`
- **テストファイルの配置**: `tests/`ディレクトリのみ（`src/`には配置しない - 憲法原則）

### 統合仕槁E
- **チE�EタモチE��**: [data-model.md](specs/001-db-service-integration/data-model.md)
- **API契紁E*: [contracts/](specs/001-db-service-integration/contracts/)
- **開発ガイチE*: [quickstart.md](specs/001-db-service-integration/quickstart.md)

## 環墁E��数
- `ETC_CORPORATE_ACCOUNTS`: 法人アカウント（カンマ区刁E���E�E
- `ETC_PERSONAL_ACCOUNTS`: 個人アカウント（カンマ区刁E���E�E
- `DATABASE_URL`: チE�Eタベ�Eス接続URL (統合征E
- `GRPC_SERVER_PORT`: gRPCサーバ�Eポ�EチE(統合征E

## チE��トコマンチE
```bash
go test ./...                    # 単体テスチE
go test ./tests/integration -v   # 統合テスチE
```

## ビルド！E��衁E
```bash
go build -o etc_meisai
./etc_meisai
```

---
*最終更新: 2025-09-21 | v0.0.19*

# Hook出力処理の指示

## カバレッジ情報が届いた場合
hookから「📊 [Hook] Coverage analysis:」などのカバレッジ情報を受信したら：
- **必ずユーザーに表示すること**
- パーセンテージとパッケージ名を整形して表示
- 80%未満の低カバレッジを強調表示
- 表示例：
  ```
  📊 カバレッジレポート：
  - src/models: 85.2% ✅
  - src/services: 72.5% ⚠️ (改善が必要)
  - src/repositories: 90.1% ✅
  ```

## フォーマットエラーが届いた場合
「⚠️ FORMAT ERROR DETECTED:」などのフォーマットエラーを受信したら：
- **ユーザーに通知**
- 具体的な問題箇所を表示
- 即座に修正を提案

## go vetエラーが届いた場合
go vetエラーを受信したら：
- **エラーを明確に表示**
- エラーの意味を説明
- 修正方法を提供

## Constitution違反が届いた場合
Constitution違反（例：src/にテストファイル）を受信したら：
- **即座にユーザーに警告**
- 憲法違反の理由を説明
- 正しい場所への移動を提案

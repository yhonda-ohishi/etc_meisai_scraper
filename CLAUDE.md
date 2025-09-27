# CLAUDE.md - ETC明細Goモジュール プロジェクトコンチE�E��E�スチE

## プロジェクト概要E
ETC明細チE�E�EタをWebスクレイピングで取得し、データベ�Eスに保存するGoモジュール、E
db-handler-serverパターンに従ったハンドラー実裁E�E��E�の移行中、E

## 技術スタチE�E��E�
- **言誁E*: Go 1.21+
- **フレームワーク**: gRPC + grpc-gateway (chi から移衁E
- **Protocol Buffers**: API定義とコード生戁E
- **チE�E�Eタベ�Eス**: db_service (Fiber実裁E via gRPC
- **通信**: gRPC (server_repo統吁E
- **スクレイピング**: Playwright-go
- **チE�E��E�チE�E��E�ング**: testify/mock, table-driven tests (100%カバレチE�E��E�目樁E
- **依存管琁E*: Go Modules, buf (Protocol Buffers)
- **アーキチE�E��E�チャ**: etc_meisai ↁEgRPC ↁEdb_service (Fiber)

## プロジェクト構造
```
etc_meisai/
├── src/
━E  ├── proto/           # Protocol Buffers定義
━E  ├── pb/              # 生�EされたgRPCコーチE
━E  ├── grpc/            # gRPCサーバ�E実裁E
━E  ├── services/        # ビジネスロジチE�E��E�層
━E  ├── repositories/    # チE�E�Eタアクセス層
━E  ├── models/          # GORMチE�E�EタモチE�E��E�
━E  └── adapters/        # 互換性レイヤー
├── handlers/            # HTTPハンドラー�E�E�E�レガシー�E�E�E�E
├── parser/              # CSV解极E
├── config/              # 設定管琁E
└── downloads/           # CSVファイル保存�E
```

## 主要機�E
1. **ETC明細ダウンローチE*: 褁E�E��E�アカウント対応、E�E��E�同期処琁E
2. **チE�E�Eタ処琁E*: CSV解析、データ変換、バルク保孁E
3. **マッピング管琁E*: ETC明細とチE�E��E�タコチE�E�Eタの関連付け�E�E�E�Etc_num活用�E�E�E�E
4. **進捗追跡**: リアルタイム進捗通知�E�E�E�ESE対応！E
5. **自動�EチE�E��E�ング**: dtako_row_idとの精寁E�E�EチE�E��E�ング

## 最近�E変更 (v0.0.19 - gRPC統吁E
- **gRPC移衁E*: go-chiからgRPC + grpc-gatewayへの移行完亁E
- **Protocol Buffers**: API定義をprotoファイルで一允E�E��E�琁E
- **Swagger統吁E*: OpenAPI仕様�E自動生成とSwagger UI統吁E
- **server_repo統吁E*: 統一されたサービス登録とルーチE�E��E�ング

## 開発中の機�E (統合フェーズ)
- **モチE�E��E�統吁E*: db_serviceのGORMモチE�E��E� + 互換性レイヤー実裁E
- **Repository統吁E*: 統吁Eepository interface + gRPCクライアント実裁E
- **サービス統吁E*: 既孁Eervices/のgRPCクライアント化

## スコープ外�E機�E
- Excel/PDF エクスポ�Eト機�E
- 統計情報生�E機�E
- キャチE�E��E�ュ機�E�E�E�E�ユーザー要求により除外！E

## パフォーマンス目樁E
- CSVファイル1丁E�E��E�めE秒以冁E�E��E�処琁E
- メモリ使用釁E00MB以丁E
- 同時ダウンローチEアカウントまで
- チE�E��E�ト実行時閁E0秒以冁E�E��E��E�EチE�E��E�トスイート！E
- チE�E��E�トカバレチE�E��E�100%維持E

## 統合アーキチE�E��E�チャ (db_service via gRPC)
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
- **チE��トファイルの配置**: `tests/`チE��レクトリのみ�E�Esrc/`には配置しなぁE- 憲法原剁E��E

### 統合仕槁E
- **チE�E�EタモチE�E��E�**: [data-model.md](specs/001-db-service-integration/data-model.md)
- **API契紁E*: [contracts/](specs/001-db-service-integration/contracts/)
- **開発ガイチE*: [quickstart.md](specs/001-db-service-integration/quickstart.md)

## 環墁E�E��E�数
- `ETC_CORPORATE_ACCOUNTS`: 法人アカウント（カンマ区刁E�E��E��E�E�E�E
- `ETC_PERSONAL_ACCOUNTS`: 個人アカウント（カンマ区刁E�E��E��E�E�E�E
- `DATABASE_URL`: チE�E�Eタベ�Eス接続URL (統合征E
- `GRPC_SERVER_PORT`: gRPCサーバ�Eポ�EチE(統合征E

## チE�E��E�トコマンチE
```bash
go test ./...                    # 単体テスチE
go test ./tests/integration -v   # 統合テスチE
```

## ビルド！E�E��E�衁E
```bash
go build -o etc_meisai
./etc_meisai
```

---
*最終更新: 2025-09-21 | v0.0.19*

# Hook出力�E琁E�E持E��

## カバレチE��惁E��が届いた場吁E
hookから「📁E[Hook] Coverage analysis:」などのカバレチE��惁E��を受信したら！E
- **忁E��ユーザーに表示すること**
- パ�EセンチE�Eジとパッケージ名を整形して表示
- 80%未満の低カバレチE��を強調表示
- 表示例！E
  ```
  📊 カバレチE��レポ�Eト！E
  - src/models: 85.2% ✁E
  - src/services: 72.5% ⚠�E�E(改喁E��忁E��E
  - src/repositories: 90.1% ✁E
  ```

## フォーマットエラーが届いた場吁E
「⚠�E�EFORMAT ERROR DETECTED:」などのフォーマットエラーを受信したら！E
- **ユーザーに通知**
- 具体的な問題箁E��を表示
- 即座に修正を提桁E

## go vetエラーが届いた場吁E
go vetエラーを受信したら！E
- **エラーを�E確に表示**
- エラーの意味を説昁E
- 修正方法を提侁E

## Constitution違反が届いた場吁E
Constitution違反�E�例：src/にチE��トファイル�E�を受信したら！E
- **即座にユーザーに警呁E*
- 憲法違反�E琁E��を説昁E
- 正しい場所への移動を提桁E

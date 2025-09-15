# ETC Meisai ハンドラー自動認識問題の解決指示書

## 問題の概要

v0.0.14でDownloadAsyncHandlerとGetDownloadStatusHandlerを実装したが、ryohi_sub_cal2がこれらのハンドラーを認識できない。

## 現在の問題分析

### 1. なぜ認識できないのか

#### 根本原因
ryohi_sub_cal2の`handler_registry.go`でハンドラーがハードコードされており、存在チェックが不完全：

```go
// handler_registry.go の問題箇所
"DownloadAsyncHandler": func() func(http.ResponseWriter, *http.Request) {
    defer func() { recover() }()
    // このハンドラーはまだ存在しない ← コメントが間違っている
    return nil  // ← 実際にはetc_meisai.DownloadAsyncHandlerが存在するのにnilを返している
},
```

#### 技術的制約
1. **ryohi_sub_cal2側の制約**: 「ryohi_sub_cal2をいじるな」という指示により修正不可
2. **リフレクション制限**: Goのリフレクションでは、パッケージレベルの関数を名前文字列から動的に取得できない
3. **コンパイル時バインディング**: Goは静的型付け言語のため、コンパイル時に関数の存在を確認する必要がある

## 解決策の提案

### 方法1: ハンドラーレジストリパターン（推奨）

etc_meisai側に自己登録メカニズムを実装する。

#### 実装手順

1. **etc_meisai/registry.go を作成**
```go
package etc_meisai

import (
    "net/http"
    "sync"
)

// HandlerRegistry は利用可能なハンドラーを管理
type HandlerRegistry struct {
    mu       sync.RWMutex
    handlers map[string]func(http.ResponseWriter, *http.Request)
}

// グローバルレジストリ
var Registry = &HandlerRegistry{
    handlers: make(map[string]func(http.ResponseWriter, *http.Request)),
}

// Register はハンドラーを登録
func (r *HandlerRegistry) Register(name string, handler func(http.ResponseWriter, *http.Request)) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.handlers[name] = handler
}

// Get はハンドラーを取得
func (r *HandlerRegistry) Get(name string) func(http.ResponseWriter, *http.Request) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.handlers[name]
}

// GetAll は全ハンドラーを取得
func (r *HandlerRegistry) GetAll() map[string]func(http.ResponseWriter, *http.Request) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    result := make(map[string]func(http.ResponseWriter, *http.Request))
    for k, v := range r.handlers {
        result[k] = v
    }
    return result
}

// init で自動登録
func init() {
    // 実装済みハンドラーを自動登録
    Registry.Register("HealthCheckHandler", HealthCheckHandler)
    Registry.Register("GetAvailableAccountsHandler", GetAvailableAccountsHandler)
    Registry.Register("DownloadETCDataHandler", DownloadETCDataHandler)
    Registry.Register("DownloadSingleAccountHandler", DownloadSingleAccountHandler)
    Registry.Register("ParseCSVHandler", ParseCSVHandler)
    Registry.Register("DownloadAsyncHandler", DownloadAsyncHandler)
    Registry.Register("GetDownloadStatusHandler", GetDownloadStatusHandler)
}
```

2. **ryohi_sub_cal2側の利用方法**
```go
// ryohi_sub_cal2側で以下のように利用可能
handlers := etc_meisai.Registry.GetAll()
for name, handler := range handlers {
    // 自動的に全ハンドラーを取得して登録
    router.HandleFunc(pathForHandler(name), handler)
}
```

### 方法2: インターフェースベースのアプローチ

#### 実装手順

1. **etc_meisai/module_interface.go を作成**
```go
package etc_meisai

import "net/http"

// ModuleInterface はモジュールが提供する機能のインターフェース
type ModuleInterface interface {
    GetHandlers() map[string]func(http.ResponseWriter, *http.Request)
    GetVersion() string
    GetSwaggerPath() string
}

// Module は実装
type Module struct{}

// GetHandlers は利用可能な全ハンドラーを返す
func (m *Module) GetHandlers() map[string]func(http.ResponseWriter, *http.Request) {
    return map[string]func(http.ResponseWriter, *http.Request){
        "HealthCheckHandler":           HealthCheckHandler,
        "GetAvailableAccountsHandler":  GetAvailableAccountsHandler,
        "DownloadETCDataHandler":       DownloadETCDataHandler,
        "DownloadSingleAccountHandler": DownloadSingleAccountHandler,
        "ParseCSVHandler":              ParseCSVHandler,
        "DownloadAsyncHandler":         DownloadAsyncHandler,
        "GetDownloadStatusHandler":     GetDownloadStatusHandler,
    }
}

// GetVersion はモジュールバージョンを返す
func (m *Module) GetVersion() string {
    return Version
}

// GetSwaggerPath はSwagger定義のパスを返す
func (m *Module) GetSwaggerPath() string {
    return "https://raw.githubusercontent.com/yhonda-ohishi/etc_meisai/master/docs/swagger.yaml"
}

// NewModule はモジュールインスタンスを作成
func NewModule() ModuleInterface {
    return &Module{}
}
```

### 方法3: プラグインシステム（Go Plugin）

**注意**: Windows環境では使用不可

### 方法4: コード生成アプローチ

自動生成ツールでryohi_sub_cal2側のコードを更新する（ただし「ryohi_sub_cal2をいじるな」制約に抵触）。

## 推奨実装プラン

### ステップ1: etc_meisai側の実装（v0.0.15）

1. **registry.go** を作成し、ハンドラーレジストリを実装
2. **module_interface.go** を作成し、モジュールインターフェースを実装
3. バージョンを0.0.15に更新
4. テストを追加
5. GitHubにプッシュ

### ステップ2: ryohi_sub_cal2側での利用方法の文書化

1. **ETC_MEISAI_INTEGRATION_V2.md** を作成
2. 新しい統合方法を説明
3. コード例を提供

### ステップ3: 移行ガイド

1. 既存のハードコードされた方法から新しい方法への移行手順
2. 後方互換性の確保

## 実装優先順位

1. **最優先**: ハンドラーレジストリパターン（方法1）
   - 実装が簡単
   - ryohi_sub_cal2側の変更が最小限
   - 自動検出が可能

2. **次善策**: インターフェースベース（方法2）
   - よりクリーンなアーキテクチャ
   - 将来の拡張に対応しやすい

## 成功基準

- [ ] etc_meisai.Registry.GetAll()で全ハンドラーを取得可能
- [ ] 新しいハンドラー追加時に自動的に認識される
- [ ] ryohi_sub_cal2側のコード変更が最小限
- [ ] 後方互換性が保たれる

## リスクと対策

### リスク
- ryohi_sub_cal2側が新しい方法を採用しない可能性

### 対策
- 従来の個別ハンドラーエクスポートも維持
- 段階的な移行をサポート
- 詳細なドキュメントとサンプルコードを提供

## まとめ

現在の問題はryohi_sub_cal2側のハードコードされたハンドラーチェックが原因。etc_meisai側でハンドラーレジストリを実装することで、ryohi_sub_cal2側の最小限の変更で自動認識を実現できる。

---
作成日: 2025-09-15
作成者: Claude
バージョン: 1.0
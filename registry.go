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

// GlobalRegistry はグローバルハンドラーレジストリ
var GlobalRegistry = &HandlerRegistry{
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

// GetNames は登録されているハンドラー名のリストを取得
func (r *HandlerRegistry) GetNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.handlers))
	for k := range r.handlers {
		names = append(names, k)
	}
	return names
}

// init で自動登録
func init() {
	// 実装済みハンドラーを自動登録
	GlobalRegistry.Register("HealthCheckHandler", HealthCheckHandler)
	GlobalRegistry.Register("GetAvailableAccountsHandler", GetAvailableAccountsHandler)
	GlobalRegistry.Register("DownloadETCDataHandler", DownloadETCDataHandler)
	GlobalRegistry.Register("DownloadSingleAccountHandler", DownloadSingleAccountHandler)
	GlobalRegistry.Register("ParseCSVHandler", ParseCSVHandler)
	GlobalRegistry.Register("DownloadAsyncHandler", DownloadAsyncHandler)
	GlobalRegistry.Register("GetDownloadStatusHandler", GetDownloadStatusHandler)
}

// GetHandlerByName は名前でハンドラーを取得（後方互換性のため）
func GetHandlerByName(name string) func(http.ResponseWriter, *http.Request) {
	return GlobalRegistry.Get(name)
}

// GetAllHandlers は全ハンドラーを取得（後方互換性のため）
func GetAllHandlers() map[string]func(http.ResponseWriter, *http.Request) {
	return GlobalRegistry.GetAll()
}
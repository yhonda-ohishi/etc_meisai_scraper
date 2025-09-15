package etc_meisai

import "net/http"

// ModuleInterface はモジュールが提供する機能のインターフェース
type ModuleInterface interface {
	GetHandlers() map[string]func(http.ResponseWriter, *http.Request)
	GetVersion() string
	GetSwaggerURL() string
}

// ETCModule は実装
type ETCModule struct{}

// GetHandlers は利用可能な全ハンドラーを返す
func (m *ETCModule) GetHandlers() map[string]func(http.ResponseWriter, *http.Request) {
	return GlobalRegistry.GetAll()
}

// GetVersion はモジュールバージョンを返す
func (m *ETCModule) GetVersion() string {
	return Version
}

// GetSwaggerURL はSwagger定義のURLを返す
func (m *ETCModule) GetSwaggerURL() string {
	return "https://raw.githubusercontent.com/yhonda-ohishi/etc_meisai/master/docs/swagger.yaml"
}

// NewETCModule はモジュールインスタンスを作成
func NewETCModule() ModuleInterface {
	return &ETCModule{}
}

// GetModule はシングルトンのモジュールインスタンスを返す
var moduleInstance = &ETCModule{}

func GetModule() ModuleInterface {
	return moduleInstance
}
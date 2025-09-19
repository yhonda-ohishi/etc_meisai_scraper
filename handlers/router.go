package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/yhonda-ohishi/etc_meisai/services"
)

// Router はAPIルーターを設定
func SetupRouter(db *sql.DB, logger *log.Logger) http.Handler {
	r := chi.NewRouter()

	// ミドルウェアの設定
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// 基底ハンドラーの作成
	base := BaseHandler{
		DB:     db,
		Logger: logger,
	}

	// 各ハンドラーの初期化
	accountsHandler := NewAccountsHandler(base)
	downloadService := services.NewDownloadService(db, logger)
	downloadHandler := NewDownloadHandler(base, downloadService)
	parseHandler := NewParseHandler(base)
	mappingHandler := NewMappingHandler(base)

	// ヘルスチェック
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"etc_meisai"}`))
	})

	// APIルートの設定
	r.Route("/api", func(r chi.Router) {
		// アカウント関連
		r.Get("/accounts", accountsHandler.GetAccounts)

		// ダウンロード関連
		r.Post("/download", downloadHandler.DownloadSync)
		r.Post("/download/async", downloadHandler.DownloadAsync)
		r.Get("/download/status/{jobId}", downloadHandler.GetDownloadStatus)

		// CSV解析関連
		r.Post("/parse", parseHandler.ParseCSV)

		// マッピング関連
		r.Get("/mapping", mappingHandler.GetMappings)
		r.Post("/mapping", mappingHandler.CreateMapping)
		r.Put("/mapping/{id}", mappingHandler.UpdateMapping)
		r.Post("/auto-match", mappingHandler.AutoMatch)
	})

	return r
}
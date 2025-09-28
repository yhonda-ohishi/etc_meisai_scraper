package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/yhonda-ohishi/etc_meisai_scraper/src/grpc"
	"github.com/yhonda-ohishi/etc_meisai_scraper/src/handlers"
	"github.com/yhonda-ohishi/etc_meisai_scraper/src/services"
)

func main() {
	// コマンドラインフラグ
	var (
		useGRPC    = flag.Bool("grpc", false, "Use gRPC server instead of HTTP")
		grpcPort   = flag.String("grpc-port", "50051", "gRPC server port")
		httpPort   = flag.String("http-port", "8080", "HTTP server port")
	)
	flag.Parse()

	// ロガー設定
	logger := log.New(os.Stdout, "[ETC-MEISAI] ", log.LstdFlags|log.Lshortfile)

	// DB接続（TODO: 実装）
	var db *sql.DB

	if *useGRPC {
		// gRPCサーバーモード
		runGRPCServer(db, logger, *grpcPort)
	} else {
		// HTTPサーバーモード
		runHTTPServer(db, logger, *httpPort)
	}
}

func runGRPCServer(db *sql.DB, logger *log.Logger, port string) {
	server := grpc.NewServer(db, logger)

	// シグナルハンドリング
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		server.Stop()
		os.Exit(0)
	}()

	if err := server.Start(port); err != nil {
		logger.Fatalf("Failed to start gRPC server: %v", err)
	}
}

func runHTTPServer(db *sql.DB, logger *log.Logger, port string) {
	// ダウンロードサービス初期化
	downloadService := services.NewDownloadService(db, logger)

	// ハンドラー初期化
	downloadHandler := handlers.NewDownloadHandler(downloadService)

	// ルーティング設定
	http.HandleFunc("/api/download/sync", downloadHandler.DownloadSync)
	http.HandleFunc("/api/download/async", downloadHandler.DownloadAsync)
	http.HandleFunc("/api/download/status", downloadHandler.GetDownloadStatus)

	logger.Printf("Starting HTTP server on port %s", port)
	logger.Printf("GitHub repository: https://github.com/yhonda-ohishi/etc_meisai_scraper")
	logger.Printf("Download endpoints:")
	logger.Printf("  POST /api/download/sync  - 同期ダウンロード")
	logger.Printf("  POST /api/download/async - 非同期ダウンロード")
	logger.Printf("  GET  /api/download/status?job_id={id} - ステータス確認")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Fatalf("HTTP server failed to start: %v", err)
	}
}
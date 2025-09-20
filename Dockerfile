# ETC明細システム Dockerfile
FROM golang:1.21-alpine AS builder

# 必要なパッケージをインストール
RUN apk add --no-cache git ca-certificates

# 作業ディレクトリを設定
WORKDIR /app

# Goモジュールファイルをコピー
COPY go.mod go.sum ./

# 依存関係をダウンロード
RUN go mod download

# ソースコードをコピー
COPY . .

# バイナリをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# 本番用イメージ
FROM alpine:latest

# 必要なパッケージをインストール
RUN apk --no-cache add ca-certificates tzdata

# 作業ディレクトリを設定
WORKDIR /root/

# ビルドしたバイナリをコピー
COPY --from=builder /app/main .

# 設定ファイルをコピー（存在する場合）
COPY --from=builder /app/config.json* ./

# データディレクトリを作成
RUN mkdir -p ./data ./logs ./temp

# ポートを公開
EXPOSE 8080

# 環境変数を設定
ENV DB_PATH=./data/etc_meisai.db
ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=8080
ENV LOG_PATH=./logs
ENV IMPORT_TEMP_DIR=./temp

# ヘルスチェックを追加
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3   CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# アプリケーションを実行
CMD ["./main"]

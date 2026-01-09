# Dockerfile
# 本番環境用のマルチステージビルド
# CI/CDパイプラインやECS Fargateデプロイ時に使用
# 開発環境では Dockerfile.dev を使用すること

FROM golang:1.25-alpine AS builder

WORKDIR /app

# 依存関係をコピーしてインストール
COPY go.mod go.sum ./
RUN go mod download

# アプリケーションコードをコピー
COPY . .

# アプリケーションをビルド
RUN go build -o /app/bin/api ./cmd/api

# 実行ステージ
FROM alpine:latest

# --- タイムゾーンデータをインストール ---
RUN apk add --no-cache tzdata
# タイムゾーンを設定
ENV TZ=Asia/Tokyo

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/bin/api .

EXPOSE 8080

CMD ["./api"]
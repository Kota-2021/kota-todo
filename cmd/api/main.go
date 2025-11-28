package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	log.Println("=== PostgreSQL 接続テスト開始 ===")

	// 環境変数から接続情報を取得
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")
	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" || dbSSLMode == "" {
		log.Fatalf("環境変数が設定されていません")
	}

	// ★ 修正箇所: パスワードをURLエンコードする ★
	// パスワードに含まれる可能性のある特殊文字を安全にURLに含める
	encodedPassword := url.QueryEscape(dbPassword)

	// 接続文字列を構築
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, encodedPassword, dbHost, dbPort, dbName, dbSSLMode,
	)

	// 本番時修正：詳細情報はログに出力しない
	log.Printf("接続先: %s:%s/%s (ユーザー: %s)", dbHost, dbPort, dbName, dbUser)

	ctx := context.Background()

	// 接続テスト（最大30秒間リトライ）
	maxRetries := 30
	retryInterval := 1 * time.Second
	var conn *pgx.Conn
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err = pgx.Connect(ctx, connStr)
		if err == nil {
			log.Println("✓ PostgreSQLへの接続に成功しました！")
			break
		}

		if i < maxRetries-1 {
			log.Printf("接続試行 %d/%d 失敗: %v (再試行します...)", i+1, maxRetries, err)
			time.Sleep(retryInterval)
		} else {
			log.Fatalf("接続試行 %d/%d 失敗: %v", i+1, maxRetries, err)
		}
	}
	defer conn.Close(ctx)

	// データベース情報を取得して表示
	var (
		version     string
		currentDB   string
		currentUser string
	)

	err = conn.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		log.Fatalf("バージョン情報の取得に失敗しました: %v", err)
	}

	err = conn.QueryRow(ctx, "SELECT current_database()").Scan(&currentDB)
	if err != nil {
		log.Fatalf("データベース名の取得に失敗しました: %v", err)
	}

	err = conn.QueryRow(ctx, "SELECT current_user").Scan(&currentUser)
	if err != nil {
		log.Fatalf("ユーザー名の取得に失敗しました: %v", err)
	}

	log.Println("\n=== 接続情報 ===")
	log.Printf("PostgreSQL バージョン: %s", version)
	log.Printf("現在のデータベース: %s", currentDB)
	log.Printf("現在のユーザー: %s", currentUser)

	// テーブル一覧を取得
	rows, err := conn.Query(ctx, `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
		ORDER BY table_name
	`)
	if err != nil {
		log.Printf("テーブル一覧の取得に失敗しました: %v", err)
	} else {
		defer rows.Close()

		var tables []string
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				log.Printf("テーブル名の読み取りに失敗しました: %v", err)
				continue
			}
			tables = append(tables, tableName)
		}

		if len(tables) > 0 {
			log.Println("\n=== テーブル一覧 ===")
			for _, table := range tables {
				log.Printf("  - %s", table)
			}
		} else {
			log.Println("\n=== テーブル一覧 ===")
			log.Println("  (テーブルはまだ作成されていません)")
		}
	}

	log.Println("\n=== 接続テスト完了 ===")
}

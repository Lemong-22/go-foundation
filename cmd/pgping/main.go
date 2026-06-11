// Package main verifies Postgres connectivity.
// It opens a connection, runs SELECT 1, and reports success or failure with exit code 0/1.
//
// Usage:
//
//	export DATABASE_URL="postgres://go_foundation:***@localhost:5432/go_foundation?sslmode=disable"
//	go run ./cmd/pgping
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is not set")
		fmt.Fprintln(os.Stderr, "export DATABASE_URL=\"postgres://go_foundation:***@localhost:5432/go_foundation?sslmode=disable\"")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	var n int
	if err := pool.QueryRow(ctx, "SELECT 1").Scan(&n); err != nil {
		fmt.Fprintf(os.Stderr, "query: %v\n", err)
		os.Exit(1)
	}

	if n != 1 {
		fmt.Fprintf(os.Stderr, "expected 1, got %d\n", n)
		os.Exit(1)
	}

	fmt.Println("OK — connected to Postgres, SELECT 1 returned 1")
}

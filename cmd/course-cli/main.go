package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

// rootCmd adalah perintah paling luar/utama.
var rootCmd = &cobra.Command{
	Use:   "course-cli",
	Short: "course-cli adalah tools terminal untuk manajemen course",
}

// dbPool menampung koneksi pgxpool yang dishare ke sub-command via Package-level ini.
// Pola ini adalah skeleton minimal: di production akan lebih baik pakai dependency
// injection lewat constructor / PersistentPreRunE yang menerima *App struct.
// Tapi untuk CLI skala kecil, global pool sudah cukup jelas.
var dbPool *pgxpool.Pool

func main() {
	// 1. Buka koneksi DB kalau DATABASE_URL tersedia.
	// Kalau tidak ada, CLI tetap jalan tapi sub-command repository akan return error.
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "[warn] DATABASE_URL tidak di-set — sub-command repository akan gagal")
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		pool, err := pgxpool.New(ctx, dsn)
		cancel()
		if err != nil {
			fmt.Fprintf(os.Stderr, "gagal membuat pool: %v\n", err)
			os.Exit(1)
		}

		// 2. Verify connectivity (SELECT 1) — fail fast kalau pool dibuat tapi DB gak reachable.
		ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
		if err := pool.Ping(ctxPing); err != nil {
			cancelPing()
			pool.Close()
			fmt.Fprintf(os.Stderr, "gagal ping DB: %v\n", err)
			os.Exit(1)
		}
		cancelPing()

		dbPool = pool
		defer func() {
			// Tutup pool bersih-bersih di akhir program.
			dbPool.Close()
		}()
	}

	// 3. Graceful shutdown kalau dapat SIGINT / SIGTERM.
	// Tidak wajib untuk CLI cepat, tapi untuk konsistensi & safety.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		if dbPool != nil {
			dbPool.Close()
		}
		fmt.Fprintln(os.Stderr, "\n[interrupt] pool DB ditutup, bye")
		os.Exit(130)
	}()

	// 4. Jalankan CLI.
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Waduh error pas jalanin CLI:", err)
		os.Exit(1)
	}
}

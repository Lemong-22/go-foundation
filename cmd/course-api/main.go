package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/lemong-22/go-foundation/internal/course"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 1. Ambil DSN dari env var (.env)
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Println("DATABASE_URL is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 2. Open kolam koneksi DB Postgres
	dbPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Printf("Database connection failed: %v\n", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// 3. Inisialisasi Repo & Handler
	repo := course.NewPostgresCourseRepository(dbPool)
	handler := course.NewCourseHandler(repo)

	// 4. Daftarin Rute HTTP (Routing bawaan Go 1.22)
	http.HandleFunc("POST /courses", handler.CreateCourseHandler)
	http.HandleFunc("GET /courses", handler.ListCoursesHandler)

	fmt.Println("Server course-api mengudara di port 8080...")
	// 5. Nyalain server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server crash: %v\n", err)
	}
}

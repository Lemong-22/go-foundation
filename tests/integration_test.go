package tests

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lemong-22/go-foundation/internal/course"
)

func TestCLIToRESTParityIntegration(t *testing.T) {
	// 1. Mengambil DSN Kunci Database dari Environment
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("Test diloncat: DATABASE_URL tidak diset di env terminal")
	}

	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("Gagal connect ke Postgres Test: %v", err)
	}
	defer dbPool.Close()

	// Fungsi sapu bersih (Clean up) agar data test sebelumnya gak mengotori test baru
	cleanup := func() {
		_, _ = dbPool.Exec(ctx, "TRUNCATE TABLE lessons, courses CASCADE")
	}
	cleanup()
	defer cleanup()

	// 2. Jalankan HTTP Server tiruan (Mock Server) memakai httptest
	repo := course.NewPostgresCourseRepository(dbPool)
	handler := course.NewCourseHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /courses", handler.ListCoursesHandler)

	server := httptest.NewServer(mux)
	defer server.Close()

	// 3. ROBOT ACT: Tembak Perintah via CLI Bawaan Cobra
	// Menjalankan: go run ./cmd/course-cli course create --title "Go Integration" --slug "go-int" --output json
	cmd := exec.Command("go", "run", "../cmd/course-cli", "course", "create",
		"--title", "Go Integration",
		"--slug", "go-int",
		"--output", "json", // Flag baru sesuai spec spec §8
	)

	// Suntikkan DATABASE_URL env agar CLI tahu jalan ke Postgres canister
	cmd.Env = append(os.Environ(), "DATABASE_URL="+dsn)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Robot CLI Gagal beroperasi: %v, Output: %s", err, string(output))
	}

	// 4. ROBOT VERIFY: Ambil data lewat pintu REST API web server
	resp, err := http.Get(server.URL + "/courses")
	if err != nil {
		t.Fatalf("Robot gagal nembak GET HTTP REST API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Eror! REST API tidak mengembalikan status 200 OK, tapi: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Gagal membaca dokumen response web: %v", err)
	}

	var courses []course.Course
	if err := json.Unmarshal(body, &courses); err != nil {
		t.Fatalf("Gagal transkrip JSON REST ke Struct Go: %v", err)
	}

	// ASERSION: Pastikan jumlah data tepat 1 dan nilainya presisi (Parity)
	if len(courses) != 1 {
		t.Fatalf("Data gak sinkron! Ekspektasi ada 1 course di DB, tapi malah ada: %d", len(courses))
	}

	if courses[0].Title != "Go Integration" || courses[0].Slug != "go-int" {
		t.Errorf("Parity Cacat! Hasil REST berbeda dengan input CLI. Dapat Title: %s, Slug: %s", courses[0].Title, courses[0].Slug)
	}
}

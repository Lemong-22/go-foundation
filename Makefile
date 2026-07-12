# Jalankan aplikasi pgping ke database
ping:
	go run ./cmd/pgping

# Bikin course baru (persist ke Postgres)
course-create:
	DATABASE_URL="postgres://go_foundation:go_foundation@localhost:5432/go_foundation?sslmode=disable" \
		go run ./cmd/course-cli course create --title "Intro to Go" --slug "intro-go"

# Tampilkan daftar course
course-list:
	DATABASE_URL="postgres://go_foundation:go_foundation@localhost:5432/go_foundation?sslmode=disable" \
		go run ./cmd/course-cli course list

# Cari course by id
course-find:
	DATABASE_URL="postgres://go_foundation:go_foundation@localhost:5432/go_foundation?sslmode=disable" \
		go run ./cmd/course-cli course find --id "$(ID)"

# Hapus course by id
course-delete:
	DATABASE_URL="postgres://go_foundation:***@localhost:5432/go_foundation?sslmode=disable" \
		go run ./cmd/course-cli course delete --id "$(ID)"

# Bikin lesson baru di bawah course tertentu (persist ke Postgres).
# Contoh pakai:
#   make lesson-create COURSE=CRSM-1783154565172307654 TITLE="Bab 1" SLUG="bab-1" ORDER=0
lesson-create:
	DATABASE_URL="postgres://go_foundation:***@localhost:5432/go_foundation?sslmode=disable" \
		go run ./cmd/course-cli lesson create \
			--course-id "$(COURSE)" --title "$(TITLE)" --slug "$(SLUG)" --order "$(ORDER)"

# Tampilkan semua lesson milik course tertentu.
# Contoh pakai:
#   make lesson-list COURSE=CRSM-1783154565172307654
lesson-list:
	DATABASE_URL="postgres://go_foundation:***@localhost:5432/go_foundation?sslmode=disable" \
		go run ./cmd/course-cli lesson list --course-id "$(COURSE)"

# Hapus lesson by id.
# Contoh pakai:
#   make lesson-delete LESSON=LES-1234567890
lesson-delete:
	DATABASE_URL="postgres://go_foundation:***@localhost:5432/go_foundation?sslmode=disable" \
		go run ./cmd/course-cli lesson delete --id "$(LESSON)"

# Jalanin HTTP REST API server di port 8080
api-start:
	DATABASE_URL="postgres://go_foundation:***@localhost:5432/go_foundation?sslmode=disable" \
		go run ./cmd/course-api

# Apply migrasi schema courses ke Postgres
migrate-up:
	psql "postgres://go_foundation:go_foundation@localhost:5432/go_foundation?sslmode=disable" \
		-f migrations/000001_create_courses.up.sql

# Drop schema courses
migrate-down:
	psql "postgres://go_foundation:go_foundation@localhost:5432/go_foundation?sslmode=disable" \
		-f migrations/000001_create_courses.down.sql

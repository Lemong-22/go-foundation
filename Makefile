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
	DATABASE_URL="postgres://go_foundation:go_foundation@localhost:5432/go_foundation?sslmode=disable" \
		go run ./cmd/course-cli course delete --id "$(ID)"

# Apply migrasi schema courses ke Postgres
migrate-up:
	psql "postgres://go_foundation:go_foundation@localhost:5432/go_foundation?sslmode=disable" \
		-f migrations/000001_create_courses.up.sql

# Drop schema courses
migrate-down:
	psql "postgres://go_foundation:go_foundation@localhost:5432/go_foundation?sslmode=disable" \
		-f migrations/000001_create_courses.down.sql

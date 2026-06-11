package migrations

import (
	"strings"
	"testing"
)

func TestRegisteredIncludesPhaseDMigrationsInOrder(t *testing.T) {
	registered, err := Registered()
	if err != nil {
		t.Fatalf("expected migrations to register, got %v", err)
	}

	wantVersions := []string{courseSchemaVersion, contentBlocksVersion, quizSchemaVersion, practiceSchemaVersion, testSchemaVersion}
	if len(registered) != len(wantVersions) {
		t.Fatalf("expected %d migrations, got %d", len(wantVersions), len(registered))
	}

	for index, want := range wantVersions {
		got := registered[index]
		if got.Version != want {
			t.Fatalf("expected migration %q at index %d, got %q", want, index, got.Version)
		}
		if strings.TrimSpace(got.UpSQL) == "" || strings.TrimSpace(got.DownSQL) == "" {
			t.Fatalf("expected migration %q to include up and down sql", want)
		}
	}
}

func TestTestMigrationIsRegisteredInRuntimeFlow(t *testing.T) {
	registered, err := Registered()
	if err != nil {
		t.Fatalf("expected migrations to register, got %v", err)
	}

	latest := registered[len(registered)-1]
	if latest.Version != testSchemaVersion {
		t.Fatalf("expected test migration to be latest, got %q", latest.Version)
	}
	if !strings.Contains(strings.ToLower(latest.UpSQL), "create table tests") {
		t.Fatalf("expected test migration sql to create tests")
	}
}

func TestPhaseEAddsNoImportPersistenceMigration(t *testing.T) {
	registered, err := Registered()
	if err != nil {
		t.Fatalf("expected migrations to register, got %v", err)
	}

	if len(registered) != 5 {
		t.Fatalf("expected Phase E to add no migrations, got %d registered migrations", len(registered))
	}
	for _, migration := range registered {
		sql := strings.ToLower(migration.UpSQL)
		if strings.Contains(sql, "create table import") || strings.Contains(sql, "create table imports") {
			t.Fatalf("expected no import persistence table in migration %s", migration.Version)
		}
	}
}

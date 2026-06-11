package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	configDBURL        = "postgres://config:pass@localhost:5432/course_config"
	configInstructorID = "550e8400-e29b-41d4-a716-446655440030"
	configAPIToken     = "config-token"
	envDBURL           = "postgres://env:pass@localhost:5432/course_env"
	envInstructorID    = "550e8400-e29b-41d4-a716-446655440031"
	envAPIToken        = "env-token"
	flagDBURL          = "postgres://flag:pass@localhost:5432/course_flag"
	flagInstructorID   = "550e8400-e29b-41d4-a716-446655440032"
	flagAPIToken       = "flag-token"
)

func TestLoadConfigReadsConfigFileKeys(t *testing.T) {
	clearConfigEnv(t)
	config := viper.New()
	config.SetConfigFile(writeConfigFile(t, configDBURL, configInstructorID, configAPIToken))

	got, err := LoadConfig(config)
	if err != nil {
		t.Fatalf("expected config to load, got %v", err)
	}

	if got.DBURL != configDBURL || got.InstructorID != configInstructorID || got.APIToken != configAPIToken {
		t.Fatalf("expected config file values, got %+v", got)
	}
}

func TestLoadConfigReadsEnvironmentValues(t *testing.T) {
	t.Setenv("COURSE_CLI_DB_URL", envDBURL)
	t.Setenv("COURSE_CLI_INSTRUCTOR_ID", envInstructorID)
	t.Setenv("COURSE_CLI_API_TOKEN", envAPIToken)

	got, err := LoadConfig(viper.New())
	if err != nil {
		t.Fatalf("expected config to load, got %v", err)
	}

	if got.DBURL != envDBURL || got.InstructorID != envInstructorID || got.APIToken != envAPIToken {
		t.Fatalf("expected env values, got %+v", got)
	}
}

func TestLoadConfigGivesFlagsPriorityOverEnvironmentAndFile(t *testing.T) {
	t.Setenv("COURSE_CLI_DB_URL", envDBURL)
	t.Setenv("COURSE_CLI_INSTRUCTOR_ID", envInstructorID)
	t.Setenv("COURSE_CLI_API_TOKEN", envAPIToken)

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String(dbURLKey, "", "database url")
	flags.String(instructorKey, "", "instructor id")
	flags.String(apiTokenKey, "", "api token")
	if err := flags.Set(dbURLKey, flagDBURL); err != nil {
		t.Fatalf("expected db url flag to set, got %v", err)
	}
	if err := flags.Set(instructorKey, flagInstructorID); err != nil {
		t.Fatalf("expected instructor flag to set, got %v", err)
	}
	if err := flags.Set(apiTokenKey, flagAPIToken); err != nil {
		t.Fatalf("expected api token flag to set, got %v", err)
	}

	config := viper.New()
	config.SetConfigFile(writeConfigFile(t, configDBURL, configInstructorID, configAPIToken))
	if err := config.BindPFlag(dbURLKey, flags.Lookup(dbURLKey)); err != nil {
		t.Fatalf("expected db url flag to bind, got %v", err)
	}
	if err := config.BindPFlag(instructorKey, flags.Lookup(instructorKey)); err != nil {
		t.Fatalf("expected instructor flag to bind, got %v", err)
	}
	if err := config.BindPFlag(apiTokenKey, flags.Lookup(apiTokenKey)); err != nil {
		t.Fatalf("expected api token flag to bind, got %v", err)
	}

	got, err := LoadConfig(config)
	if err != nil {
		t.Fatalf("expected config to load, got %v", err)
	}

	if got.DBURL != flagDBURL || got.InstructorID != flagInstructorID || got.APIToken != flagAPIToken {
		t.Fatalf("expected flag values, got %+v", got)
	}
}

func TestLoadConfigWrapsMalformedConfigFile(t *testing.T) {
	clearConfigEnv(t)
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("db_url: ["), 0o600); err != nil {
		t.Fatalf("expected config fixture to write, got %v", err)
	}

	config := viper.New()
	config.SetConfigFile(path)

	_, err := LoadConfig(config)
	if err == nil {
		t.Fatalf("expected malformed config to fail")
	}
	if !strings.HasPrefix(err.Error(), "read config") {
		t.Fatalf("expected read config context, got %v", err)
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()

	t.Setenv("COURSE_CLI_DB_URL", "")
	t.Setenv("COURSE_CLI_INSTRUCTOR_ID", "")
	t.Setenv("COURSE_CLI_API_TOKEN", "")
}

func writeConfigFile(t *testing.T, dbURL string, instructorID string, apiToken string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := "db_url: " + dbURL + "\ninstructor_id: " + instructorID + "\napi_token: " + apiToken + "\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("expected config fixture to write, got %v", err)
	}

	return path
}

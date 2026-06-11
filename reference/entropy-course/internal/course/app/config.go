package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	envPrefix         = "COURSE_CLI"
	configName        = "config"
	configType        = "yaml"
	configDirectory   = ".config/course-cli"
	dbURLKey          = "db-url"
	dbURLFileKey      = "db_url"
	instructorKey     = "instructor-id"
	instructorFileKey = "instructor_id"
	apiTokenKey       = "api-token"
	apiTokenFileKey   = "api_token"
)

type Config struct {
	DBURL        string
	InstructorID string
	APIToken     string
}

func LoadConfig(config *viper.Viper) (Config, error) {
	config = ConfigureViper(config)
	if err := config.ReadInConfig(); err != nil && !isMissingConfigFile(err) {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	return ConfigFromViper(config), nil
}

func ConfigFromViper(config *viper.Viper) Config {
	config = ConfigureViper(config)

	return Config{
		DBURL:        configString(config, dbURLKey, dbURLFileKey),
		InstructorID: configString(config, instructorKey, instructorFileKey),
		APIToken:     configString(config, apiTokenKey, apiTokenFileKey),
	}
}

func ConfigureViper(config *viper.Viper) *viper.Viper {
	if config == nil {
		config = viper.New()
	}

	if config.ConfigFileUsed() == "" {
		config.SetConfigName(configName)
		config.SetConfigType(configType)
		if home, err := os.UserHomeDir(); err == nil {
			config.AddConfigPath(filepath.Join(home, configDirectory))
		}
	}

	config.SetEnvPrefix(envPrefix)
	config.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	config.AutomaticEnv()

	config.SetDefault(dbURLKey, "")
	config.SetDefault(instructorKey, "")
	config.SetDefault(apiTokenKey, "")

	return config
}

func configString(config *viper.Viper, keys ...string) string {
	for _, key := range keys {
		if value := config.GetString(key); value != "" {
			return value
		}
	}

	return ""
}

func isMissingConfigFile(err error) bool {
	var missing viper.ConfigFileNotFoundError
	return errors.As(err, &missing)
}

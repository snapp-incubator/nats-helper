package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// NATSConfig holds NATS connection configuration
type NATSConfig struct {
	URL      string `mapstructure:"url"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// Config holds all configuration
type Config struct {
	NATS NATSConfig `mapstructure:"nats"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("nats.url", "nats://localhost:4222")

	// Read from environment variables
	v.SetEnvPrefix("NATS_HELPER")
	v.AutomaticEnv()
	if err := v.BindEnv("nats.url", "NATS_URL"); err != nil {
		return nil, fmt.Errorf("failed to bind env nats.url: %w", err)
	}
	if err := v.BindEnv("nats.user", "NATS_USER"); err != nil {
		return nil, fmt.Errorf("failed to bind env nats.user: %w", err)
	}
	if err := v.BindEnv("nats.password", "NATS_PASSWORD"); err != nil {
		return nil, fmt.Errorf("failed to bind env nats.password: %w", err)
	}

	// If config file is specified, use it
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Otherwise, look for config in default locations
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}

		v.AddConfigPath(home)
		v.AddConfigPath(".")
		v.SetConfigName(".nats-helper")
		v.SetConfigType("yaml")
	}

	// Read config file if it exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// Save saves the configuration to a file
func (c *Config) Save(configPath string) error {
	v := viper.New()

	// Set values
	v.Set("nats.url", c.NATS.URL)
	v.Set("nats.user", c.NATS.User)
	v.Set("nats.password", c.NATS.Password)

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

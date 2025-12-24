package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App      App
	Database Database
}

type App struct {
	Mode string `env:"MODE" env-required:"debug"` // debug, release
	Port string `env:"PORT" env-required:"8080"`
}

type Database struct {
	Host     string `env:"POSTGRES_HOST" env-required:"true"`
	Port     string `env:"POSTGRES_PORT" env-required:"true"`
	User     string `env:"POSTGRES_USER" env-required:"true"`
	DBName   string `env:"POSTGRES_DB" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	SSLMode  string `env:"POSTGRES_SSLMODE" env-required:"true"`
}

func (d Database) DSN() string {
	return fmt.Sprintf(
		`host=%s port=%s user=%s password=%s dbname=%s sslmode=%s`,
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

var (
	instance *Config
	once     sync.Once
	errInit  error
)

func Get() *Config {
	once.Do(func() {
		cfg := &Config{}

		if err := readConfig(cfg); err != nil {
			errInit = fmt.Errorf("failed to load config: %w", err)
			return
		}
		instance = cfg
	})

	if errInit != nil {
		panic(errInit)
	}
	return instance
}

func readConfig(cfg *Config) error {
	if _, err := os.Stat(".env"); err == nil {
		if err := cleanenv.ReadConfig(".env", cfg); err != nil {
			return fmt.Errorf("read .env file: %w", err)
		}
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return fmt.Errorf("invalid or missing environment variables: %w", err)
	}
	return nil
}

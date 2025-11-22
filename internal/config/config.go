package config

import (
	"log" //nolint
	"os"
	"path/filepath"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	EnvDebug = "debug"
	EnvProd  = "prod"
)

type Config struct {
	Env        string     `yaml:"env"         env-default:"prod"`
	HTTPServer HTTPServer `yaml:"http_server"`
	DB         DB
}

type HTTPServer struct {
	Address     string        `yaml:"address"      env-default:"localhost"`
	Port        int           `yaml:"port"         env-default:"8080"`
	Timeout     time.Duration `yaml:"timeout"      env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type DB struct {
	Host     string `env:"DB_HOST"     env-required:"true"`
	Port     int    `env:"DB_PORT"     env-required:"true"`
	User     string `env:"DB_USER"     env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	Name     string `env:"DB_NAME"     env-required:"true"`
	SslMode  string `env:"DB_SSL_MODE" env-required:"true"`
}

const VarConfigPath = "CONFIG_PATH"

func Must() *Config {
	configPath := os.Getenv(VarConfigPath)
	if configPath == "" {
		log.Fatalf("env variable %s not set", VarConfigPath)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file %s does not exist", configPath)
	}

	var config Config
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("cannot get working directory: %v", err)
	}

	dotEnvPath := filepath.Join(wd, ".env")
	if err = cleanenv.ReadConfig(dotEnvPath, &config); err != nil {
		log.Fatalf("cannot read .env file: %v", err)
	}

	if err = cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatalf("cannot read config file: %v", err)
	}

	return &config
}

package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type AuthConfig struct {
	JWTSecret string `mapstructure:"jwt_secret"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
}

var App Config

func Load(configPath string) {
	dir, _ := os.Getwd()
	_ = godotenv.Load(filepath.Join(dir, ".env"))
	_ = godotenv.Load(filepath.Join(dir, "..", "..", ".env"))

	viper.SetConfigFile(configPath)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.dbname", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")
	viper.BindEnv("server.port", "MAGANG_PORT")
	viper.BindEnv("auth.jwt_secret", "MAGANG_JWT_SECRET")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config [%s]: %v", configPath, err)
	}
	if err := viper.Unmarshal(&App); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	if strings.TrimSpace(App.Auth.JWTSecret) == "" {
		log.Fatal("MAGANG_JWT_SECRET is required but not set")
	}
}

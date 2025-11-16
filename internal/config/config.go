package config

import (
  "fmt"
  "os"

  "github.com/joho/godotenv"
)

type Config struct {
  Host     string
  Port     string
  DBHost   string
  DBPort   string
  DBName   string
  DBUser   string
  DBPass   string
  SMTPHost string
  SMTPPort string
  SMTPUser string
  SMTPPass string
}

func LoadConfig() (*Config, error) {
  godotenv.Load()

  cfg := &Config{
    Host:     getEnv("HOST", "0.0.0.0"),
    Port:     getEnv("PORT", "8080"),
    DBHost:   getEnv("PG_HOST", "localhost"),
    DBPort:   getEnv("PG_PORT", "5432"),
    DBName:   getEnv("PG_DATABASE", "newsletter"),
    DBUser:   getEnv("PG_USER", "postgres"),
    DBPass:   getEnv("PG_PASSWORD", ""),
    SMTPHost: getEnv("SMTP_SERVER", ""),
    SMTPPort: getEnv("SMTP_PORT", "587"),
    SMTPUser: getEnv("SMTP_USER", ""),
    SMTPPass: getEnv("SMTP_PASSWORD", ""),
  }

  if cfg.SMTPHost == "" {
    return nil, fmt.Errorf("SMTP_SERVER not configured")
  }

  return cfg, nil
}

func getEnv(key, defaultVal string) string {
  if value, exists := os.LookupEnv(key); exists {
    return value
  }
  return defaultVal
}
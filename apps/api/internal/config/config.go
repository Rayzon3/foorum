package config

import (
  "os"
  "strings"
)

type Config struct {
  Port        string
  DatabaseURL string
  JWTSecret   string
  CorsOrigins []string
}

func Load() Config {
  port := getenv("PORT", "8080")
  dbURL := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/jabber?sslmode=disable")
  jwtSecret := getenv("JWT_SECRET", "change-me")
  origins := getenv("CORS_ORIGINS", "http://localhost:5173")

  return Config{
    Port:        port,
    DatabaseURL: dbURL,
    JWTSecret:   jwtSecret,
    CorsOrigins: splitAndTrim(origins),
  }
}

func getenv(key, fallback string) string {
  if val := os.Getenv(key); val != "" {
    return val
  }
  return fallback
}

func splitAndTrim(raw string) []string {
  parts := strings.Split(raw, ",")
  out := make([]string, 0, len(parts))
  for _, p := range parts {
    trimmed := strings.TrimSpace(p)
    if trimmed != "" {
      out = append(out, trimmed)
    }
  }
  if len(out) == 0 {
    return []string{"*"}
  }
  return out
}

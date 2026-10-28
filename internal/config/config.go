package config

import "os"

// Config holds settings read from environment variables.
type Config struct {
	Addr    string // host:port the HTTP sever listens on
	DataDir string // where the SQLite DB and uploads will live
	Env     string // "dev" or "prod"
}

func Load() Config {
	return Config{
		Addr:    getenv("ADDR", "127.0.01:8080"),
		DataDir: getenv("DATA_DIR", "./data"),
		Env:     getenv("APP_ENV", "dev"),
	}
}

func (c Config) IsDev() bool { return c.Env == "dev" }

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

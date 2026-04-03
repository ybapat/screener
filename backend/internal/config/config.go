package config

import (
	"os"
)

type Config struct {
	DatabaseURL       string
	RedisURL          string
	JWTSecret         string
	Port              string
	Env               string
	SolanaRPCURL      string
	SolanaWSURL       string
	SolanaKeypairPath string
	SolanaProgramID   string
}

func Load() *Config {
	return &Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://screener_user:devpassword@localhost:5432/screener?sslmode=disable"),
		RedisURL:          getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:         getEnv("JWT_SECRET", "dev-secret-change-in-prod"),
		Port:              getEnv("PORT", "8080"),
		Env:               getEnv("ENV", "development"),
		SolanaRPCURL:      getEnv("SOLANA_RPC_URL", ""),
		SolanaWSURL:       getEnv("SOLANA_WS_URL", ""),
		SolanaKeypairPath: getEnv("SOLANA_KEYPAIR_PATH", "/app/data/solana-keypair.json"),
		SolanaProgramID:   getEnv("SOLANA_PROGRAM_ID", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

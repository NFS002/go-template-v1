package utils

import (
	"fmt"
	"os"
	"strconv"

	dotenv "github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func GetEnvOrDefault(key, fallback string) string {
	if valueStr, exists := os.LookupEnv(key); exists && len(valueStr) > 0 {
		return valueStr
	}
	return fallback
}

func GetIntEnvOrDefault(key string, fallback int) int {
	if valueStr, exists := os.LookupEnv(key); exists && len(valueStr) > 0 {
		if valueInt, err := strconv.ParseInt(valueStr, 0, 0); err == nil {
			return int(valueInt)
		}

	}
	return fallback
}

func GetBoolEnvOrDefault(key string, fallback bool) bool {
	valueStr, exists := os.LookupEnv(key)
	if exists {
		switch valueStr {
		case "true", "True", "t", "T":
			return true
		}
		return false
	}
	return fallback
}

// If the APP_ENV variable is set and the file .env.{APP_ENV} exists, load that
// Otherwise, default to .env
func LoadEnv() {
	value, exists := os.LookupEnv("APP_ENV")
	log.Info().Str("APP_ENV", value).Msg("Loading env: reading $APP_ENV")
	if exists && len(value) > 0 {
		filename := fmt.Sprintf(".env.%s", value)
		if _, err := os.Stat(filename); err == nil {
			log.Info().Str("filename", filename).Msg("Loading env: Found file")
			dotenv.Load(filename)
			return
		} else {
			log.Info().Str("filename", filename).Msg("loading env: File not found")
		}
	}
	log.Info().Str("filename", ".env").Msg("Loading env: Using default file")
	dotenv.Load(".env")
}

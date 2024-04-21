package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"

	dotenv "github.com/joho/godotenv"
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
func LoadEnv() (*log.Logger, *log.Logger) {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	value, exists := os.LookupEnv("APP_ENV")
	infoLog.Printf("Loading environment variables, found APP_ENV=%s", value)
	if exists && len(value) > 0 {
		filename := fmt.Sprintf(".env.%s", value)
		if _, err := os.Stat(filename); err == nil {
			infoLog.Printf("Loading environment variables, found file=%s", filename)
			dotenv.Load(filename)
			return infoLog, errorLog
		} else {
			errorLog.Printf("Loading environment variables, no %s file found. Defaulting to .env", filename)
		}
	}
	infoLog.Printf("Loading environment variables from default .env file")
	dotenv.Load(".env")
	return infoLog, errorLog
}

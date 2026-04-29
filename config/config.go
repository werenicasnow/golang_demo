package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	ServerPort string
	CacheTTL   int
}

func LoadConfig() *Config {
	// Переменные для БД с значениями по умолчанию
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnvAsInt("DB_PORT", 5432)
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "1")
	dbName := getEnv("DB_NAME", "golang_db")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")
	
	// Порт сервера
	serverPort := getEnv("SERVER_PORT", "8082")
	
	// Время жизни кеша в секундах
	cacheTTL := getEnvAsInt("CACHE_TTL_SECONDS", 60)
	
	return &Config{
		DBHost:     dbHost,
		DBPort:     dbPort,
		DBUser:     dbUser,
		DBPassword: dbPassword,
		DBName:     dbName,
		DBSSLMode:  dbSSLMode,
		ServerPort: serverPort,
		CacheTTL:   cacheTTL,
	}
}

// getEnv получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает переменную окружения как int или возвращает значение по умолчанию
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetDBConnectionString формирует строку подключения к БД
func (c *Config) GetDBConnectionString() string {
	return "postgres://" + c.DBUser + ":" + c.DBPassword + "@" + c.DBHost + ":" + strconv.Itoa(c.DBPort) + "/" + c.DBName + "?sslmode=" + c.DBSSLMode
}
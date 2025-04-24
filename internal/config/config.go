package config

import (
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	TELEGRAM_BOT_TOKEN string
}

func LoadConfig() AppConfig {
	err := godotenv.Load("./config/.env")
	if err != nil {
		// TODO
		// Если файл не найден, можно продолжить с переменными окружения из системы
		// или завершить с ошибкой, в зависимости от ваших требований
		// Здесь оставим как опциональное логирование в main
		return AppConfig{}
	}

	return AppConfig{
		DBHost:             os.Getenv("DB_HOST"),
		DBPort:             os.Getenv("DB_PORT"),
		DBUser:             os.Getenv("DB_USER"),
		DBPassword:         os.Getenv("DB_PASSWORD"),
		DBName:             os.Getenv("DB_NAME"),
		TELEGRAM_BOT_TOKEN: os.Getenv("TELEGRAM_BOT_TOKEN"),
	}
}

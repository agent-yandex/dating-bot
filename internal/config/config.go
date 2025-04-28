package config

import (
	"os"

	"github.com/joho/godotenv"
)

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

type AppConfig struct {
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	TELEGRAM_BOT_TOKEN string
	Minio              MinioConfig // Новое поле для MinIO
}

func LoadConfig() AppConfig {
	err := godotenv.Load("./config/.env")
	if err != nil {
		return AppConfig{}
	}

	return AppConfig{
		DBHost:             os.Getenv("DB_HOST"),
		DBPort:             os.Getenv("DB_PORT"),
		DBUser:             os.Getenv("DB_USER"),
		DBPassword:         os.Getenv("DB_PASSWORD"),
		DBName:             os.Getenv("DB_NAME"),
		TELEGRAM_BOT_TOKEN: os.Getenv("TELEGRAM_BOT_TOKEN"),
		Minio: MinioConfig{
			Endpoint:  os.Getenv("MINIO_ENDPOINT"),
			AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
			SecretKey: os.Getenv("MINIO_SECRET_KEY"),
			Bucket:    os.Getenv("MINIO_BUCKET"),
		},
	}
}

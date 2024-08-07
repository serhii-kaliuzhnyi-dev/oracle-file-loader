package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DBUrl      string `envconfig:"DB_URL"`
	DBUser     string `envconfig:"DB_USER"`
	DBPassword string `envconfig:"DB_PASSWORD"`
	FilePath   string `envconfig:"FILE_PATH"`
	TableName  string `envconfig:"TABLE_NAME"`
	CtlFilePath string `envconfig:"CTL_FILE_PATH"`
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	} else {
		log.Println(".env file loaded successfully")
	}

	var config Config
	err = envconfig.Process("", &config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}

package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
}

var Init Config

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found!")
	}

	Init = Config{
		Port: os.Getenv("PORT"),
	}
}

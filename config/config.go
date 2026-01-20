package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken      string
	BemVindoChannelID string
	CargoAutoID       string
}

var AppConfig Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env não encontrado, usando variáveis de ambiente")
	}

	AppConfig.DiscordToken = os.Getenv("DISCORD_TOKEN")
	AppConfig.BemVindoChannelID = os.Getenv("BEM_VINDO_CHANNEL_ID")
	AppConfig.CargoAutoID = os.Getenv("CARGO_AUTO_ID")

	if AppConfig.DiscordToken == "" {
		log.Fatal("❌ DISCORD_TOKEN não configurado!")
	}
}

package main

import (
	"fmt"
	"os"
	"github.com/joho/godotenv"
	"feints/internal/discord"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("❌ Error loading .env file, proceeding with environment variables")
	}
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		fmt.Println("❌ Missing DISCORD_TOKEN env variable")
		return
	}
	if err := discord.StartBot(token); err != nil {
		fmt.Println("❌", err)
	}
}

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/Games-Gamers/StarBot/bot"
)

func main() {
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	// check the contents of the loaded .env
	_, exists := os.LookupEnv("Token")
	if !exists {
		fmt.Println("Token value not found in .env")
		return
	}
	_, exists = os.LookupEnv("StarboardChannel")
	if !exists {
		fmt.Println("StarboardChannel value not found in .env")
		return
	}
	_, exists = os.LookupEnv("LoggingChannel")
	if !exists {
		fmt.Println("LoggingChannel value not found in .env")
		return
	}
	_, exists = os.LookupEnv("GuildID")
	if !exists {
		fmt.Println("GuildID value not found in .env")
		return
	}

	bot.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-c
	// bot.Stop()
}

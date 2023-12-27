package main

import (
	"log"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/Games-Gamers/StarBot/bot"
)

var logg *log.Logger = log.New(os.Stdout, "Info: ", log.Ldate|log.Ltime|log.Lshortfile)

func main() {
	// Load the .env file
	_ := godotenv.Load()

	// check the contents of the loaded .env
	_, exists := os.LookupEnv("Token")
	if !exists {
		logg.Println("Token value not found in .env")
		return
	}
	_, exists = os.LookupEnv("StarboardChannel")
	if !exists {
		logg.Println("StarboardChannel value not found in .env")
		return
	}
	_, exists = os.LookupEnv("LoggingChannel")
	if !exists {
		logg.Println("LoggingChannel value not found in .env")
		return
	}
	_, exists = os.LookupEnv("GuildID")
	if !exists {
		logg.Println("GuildID value not found in .env")
		return
	}

	bot.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-c
	bot.Stop()
}

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Games-Gamers/StarBot/bot"
	"github.com/Games-Gamers/StarBot/config"
)

func main() {
	err := config.ReadConfig()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	bot.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-c
	bot.Stop()
}

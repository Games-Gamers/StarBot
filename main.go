package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"testbot.com/TestBot/bot"
	"testbot.com/TestBot/config"
)

func main() {
	err := config.ReadConfig()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	bot.Start()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		bot.Stop()
		os.Exit(0)
	}()
	<-make(chan struct{})

}

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var (
	Token            string
	BotPrefix        string
	StarboardChannel string
	LoggingChannel   string

	config *configStruct
)

type configStruct struct {
	Token            string `json:"Token"`
	BotPrefix        string `json:"BotPrefix"`
	StarboardChannel string `json:"StarboardChannel"`
	LoggingChannel   string `json:"LoggingChannel"`
}

func ReadConfig() error {
	fmt.Println("Reading config file...")

	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	Token = config.Token
	BotPrefix = config.BotPrefix
	StarboardChannel = config.StarboardChannel
	LoggingChannel = config.LoggingChannel

	return nil
}

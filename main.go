package main

import (
	"log"
	"os"

	pills_bot "github.com/fulviodenza/pills-of-cs/bot"
) //This the token you receive from botfather

type Env struct {
	TelegramToken string `json:"TELEGRAM_TOKEN"`
}

func main() {

	bot, err := pills_bot.NewBotWithConfig()
	if err != nil {
		log.Fatalf("[NewBotWithConfig()]: %v", err)
		os.Exit(1)
	}

	err = bot.Bot.Run()
	if err != nil {
		log.Fatalf("[bot.Bot.Run()]: %v", err)
		os.Exit(1)
	}

	bot.Start()
}

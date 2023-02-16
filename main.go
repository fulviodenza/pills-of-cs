package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pills-of-cs/adapters/ent"
	"github.com/pills-of-cs/bot"
)

type Env struct {
	TelegramToken string `json:"TELEGRAM_TOKEN"`
}

const uri = "mongodb://localhost:27017/"

func main() {

	var err error
	var ctx context.Context

	client, err := ent.SetupAndConnectDatabase(uri, "pills")
	if err != nil {
		log.Fatalf("[ent.SetupAndConnectDatabase]: %v", err)
	}

	bot, err := bot.NewBotWithConfig(client)
	if err != nil {
		log.Fatalf("[NewBotWithConfig]: %v", err)
		os.Exit(1)
	}

	// Create the db and the collection
	bot.UserRepo.Client = client

	err = bot.Bot.Run()
	if err != nil {
		log.Fatalf("[bot.Bot.Run]: %v", err)
		os.Exit(1)
	}

	bot.Start(ctx)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

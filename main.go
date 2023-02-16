package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	pills_bot "github.com/fulviodenza/pills-of-cs/bot"
	"github.com/pills-of-cs/adapters/ent"
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

	bot, err := pills_bot.NewBotWithConfig(client)
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

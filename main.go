package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	adapters "github.com/pills-of-cs/adapters/repositories"
	"github.com/pills-of-cs/bot"
)

func main() {

	var err error
	ctx := context.Background()

	botWithConfig, client, err := bot.NewBotWithConfig()
	if err != nil {
		log.Fatalf("got error: %v", err)
	}

	// Create the db and the collection
	botWithConfig.UserRepo = adapters.NewUserRepo(client)

	err = botWithConfig.TelegramClient.Run(false)
	if err != nil {
		log.Fatalf("got error: %v", err)
	}
	go botWithConfig.Start(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

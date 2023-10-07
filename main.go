package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	botWithConfig.UserRepo.Client = client

	err = botWithConfig.TelegramClient.Run()
	if err != nil {
		log.Fatalf("got error: %v", err)
	}
	go botWithConfig.Start(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

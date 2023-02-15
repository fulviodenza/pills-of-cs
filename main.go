package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"pills-of-cs/adapters/ent"
	"syscall"

	pills_bot "github.com/fulviodenza/pills-of-cs/bot"
	"go.mongodb.org/mongo-driver/mongo"
)

type Env struct {
	TelegramToken string `json:"TELEGRAM_TOKEN"`
}

const uri = "mongodb://localhost:27017/"

func main() {

	var client *mongo.Client
	var err error

	for {

		client, err := ent.SetupAndConnectDatabase(uri, "pills")
		if err != nil {
			log.Fatalf("[ent.SetupAndConnectDatabase]: %v", err)
		}

		// client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
		// if err != nil {
		// 	log.Fatalf("[mongo.Connect]: %v", err)
		// 	continue
		// }
		// defer func() {
		// 	if err = client.Disconnect(ctx); err != nil {
		// 		log.Fatalf("[client.Disconnect]: %v", err)
		// 	}
		// }()

		// if err := client.Ping(ctx, readpref.Primary()); err != nil {
		// 	log.Fatalf("[client.Ping]: %v", err)
		// 	continue
		// }
		break
	}
	fmt.Println("BRUH: Successfully connected and pinged.")

	bot, err := pills_bot.NewBotWithConfig(client)
	if err != nil {
		log.Fatalf("[NewBotWithConfig]: %v", err)
		os.Exit(1)
	}

	// Create the db and the collection
	bot.UserRepo.Client = client

	log.Println(dbName)
	err = bot.Bot.Run()
	if err != nil {
		log.Fatalf("[bot.Bot.Run]: %v", err)
		os.Exit(1)
	}

	bot.Start()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

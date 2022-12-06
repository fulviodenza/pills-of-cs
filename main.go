package main

import (
	"context"
	"fmt"
	"log"
	"os"

	pills_bot "github.com/fulviodenza/pills-of-cs/bot"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Env struct {
	TelegramToken string `json:"TELEGRAM_TOKEN"`
}

const uri = "mongodb://localhost:27017/"

func main() {

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("[mongo.Connect]: %v", err)
		os.Exit(1)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal("[client.Disconnect]: %v", err)
			os.Exit(1)
		}
	}()

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal("[client.Ping]: %v", err)
		os.Exit(1)
	}
	fmt.Println("Successfully connected and pinged.")

	bot, err := pills_bot.NewBotWithConfig(client)
	if err != nil {
		log.Fatalf("[NewBotWithConfig]: %v", err)
		os.Exit(1)
	}

	err = bot.Bot.Run()
	if err != nil {
		log.Fatalf("[bot.Bot.Run]: %v", err)
		os.Exit(1)
	}

	bot.Start()
}

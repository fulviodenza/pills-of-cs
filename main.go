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

	var client *mongo.Client
	for {
		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
		if err != nil {
			log.Fatalf("[mongo.Connect]: %v", err)
			continue
		}
		defer func() {
			if err = client.Disconnect(context.TODO()); err != nil {
				log.Fatalf("[client.Disconnect]: %v", err)
			}
		}()

		if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
			log.Fatalf("[client.Ping]: %v", err)
			continue
		}
		break
	}
	fmt.Println("Successfully connected and pinged.")

	bot, err := pills_bot.NewBotWithConfig(client)
	if err != nil {
		log.Fatalf("[NewBotWithConfig]: %v", err)
		os.Exit(1)
	}

	// Create the db and the collection
	dbName := bot.UserRepo.Client.Database("pills").Collection("users")

	log.Println(dbName)
	err = bot.Bot.Run()
	if err != nil {
		log.Fatalf("[bot.Bot.Run]: %v", err)
		os.Exit(1)
	}

	bot.Start()
}

package bot

import (
	"context"
	"fmt"
	"log"
	"os"

	bt "github.com/SakoDroid/telego"
	cfg "github.com/SakoDroid/telego/configs"
	"github.com/SakoDroid/telego/objects"
	"github.com/joho/godotenv"
	"github.com/jomei/notionapi"
)

type Bot struct {
	TelegramToken string
	Cfg           cfg.BotConfigs
	Bot           bt.Bot
	NotionClient  notionapi.Client
}

func NewBotWithConfig() (*Bot, error) {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("[NewBotWithConfig]: %v", err)
		return nil, err
	}

	token := os.Getenv("TELEGRAM_TOKEN")

	cf := cfg.DefaultUpdateConfigs()

	bot_config := cfg.BotConfigs{
		BotAPI: cfg.DefaultBotAPI,
		APIKey: token, UpdateConfigs: cf,
		Webhook:        false,
		LogFileAddress: cfg.DefaultLogFile,
	}

	b, err := bt.NewBot(&bot_config)
	if err != nil {
		return nil, err
	}

	notion_client := notionapi.NewClient("secret_cHeqVXYXaRURTy8PJiLn81PL4G27Mxnm5hYA0BtWvyD")

	return &Bot{
		TelegramToken: token,
		Cfg:           bot_config,
		Bot:           *b,
		NotionClient:  *notion_client,
	}, nil
}

func (b *Bot) Start() error {
	//Register the channel
	messageChannel, _ := b.Bot.AdvancedMode().RegisterChannel("", "message")

	for {
		up := <-*messageChannel
		b.handleMessage(up)
	}
}

func (b *Bot) handleMessage(up *objects.Update) {
	switch {
	case up.Message.Text == "/start":
		_, err := b.Bot.SendMessage(up.Message.Chat.Id, "Hello from the server!", "", up.Message.MessageId, false, false)
		if err != nil {
			return
		}
	case up.Message.Text == "/pill":
		page, err := b.NotionClient.Page.Get(context.Background(), "48b530629463419ca92e22cc6ef50dab")
		if err != nil {
			fmt.Println(page)
		}
	}
}

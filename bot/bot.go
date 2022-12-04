package bot

import (
	"math/rand"
	"os"
	"time"

	bt "github.com/SakoDroid/telego"
	cfg "github.com/SakoDroid/telego/configs"
	"github.com/SakoDroid/telego/objects"
	notionapi "github.com/dstotijn/go-notion"
)

const PAGE_ID = "48b530629463419ca92e22cc6ef50dab"

type Bot struct {
	TelegramToken string
	Cfg           cfg.BotConfigs
	Bot           bt.Bot
	NotionClient  notionapi.Client
}

func NewBotWithConfig() (*Bot, error) {

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

	notion_client := notionapi.NewClient(os.Getenv("NOTION_TOKEN"))

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
		serializedPills, err := parse()
		if err != nil {
			return
		}
		rand.Seed(time.Now().Unix())
		randomIndex := rand.Intn(len(serializedPills.Pills))
		_, err = b.Bot.SendMessage(
			up.Message.Chat.Id,
			"BRUH: "+serializedPills.Pills[randomIndex].Title+": "+serializedPills.Pills[randomIndex].Body, "", up.Message.MessageId, false, false)
	}
}

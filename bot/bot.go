package bot

import (
	"encoding/json"
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
		_, err := b.Bot.SendMessage(up.Message.Chat.Id, "Welcome to the pills-of-cs bot! Press `/pill` to request a pill or `/help` to get informations about the bot", "", up.Message.MessageId, false, false)
		if err != nil {
			return
		}
	case up.Message.Text == "/pill":
		var dst []byte
		_, err := parse(&dst)
		if err != nil {
			return
		}

		sp := SerializedPills{}
		err = json.Unmarshal(dst, &sp)
		if err != nil {
			return
		}
		rand.Seed(time.Now().Unix())

		randomIndex := rand.Intn(len(sp.Pills))
		_, err = b.Bot.SendMessage(
			up.Message.Chat.Id,
			sp.Pills[randomIndex].Title+": "+sp.Pills[randomIndex].Body, "", up.Message.MessageId, false, false)
	case up.Message.Text == "/help":
		var dst []byte
		_, err := parse(&dst)
		if err != nil {
			return
		}
		_, err = b.Bot.SendMessage(up.Message.Chat.Id, string(dst), "", up.Message.MessageId, false, false)
		if err != nil {
			return
		}
	}
}

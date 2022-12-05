package bot

import (
	"encoding/json"
	"math/rand"
	"os"
	"strings"
	"time"

	bt "github.com/SakoDroid/telego"
	cfg "github.com/SakoDroid/telego/configs"
	"github.com/SakoDroid/telego/objects"
	notionapi "github.com/dstotijn/go-notion"
)

const (
	PAGE_ID        = "48b530629463419ca92e22cc6ef50dab"
	NOTION_TOKEN   = "NOTION_TOKEN"
	TELEGRAM_TOKEN = "TELEGRAM_TOKEN"
)

type Bot struct {
	TelegramToken string
	Cfg           cfg.BotConfigs
	Bot           bt.Bot
	NotionClient  notionapi.Client
}

func NewBotWithConfig() (*Bot, error) {

	var (
		telegramToken string
		notionToken   string
	)

	// The function does not work?
	// F*** off, I implement it by myself
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if pair[0] == TELEGRAM_TOKEN {
			telegramToken = pair[1]
		}
		if pair[0] == NOTION_TOKEN {
			notionToken = pair[1]
		}
	}

	cf := cfg.DefaultUpdateConfigs()

	bot_config := cfg.BotConfigs{
		BotAPI: cfg.DefaultBotAPI,
		APIKey: telegramToken, UpdateConfigs: cf,
		Webhook:        false,
		LogFileAddress: cfg.DefaultLogFile,
	}

	b, err := bt.NewBot(&bot_config)
	if err != nil {
		return nil, err
	}

	notion_client := notionapi.NewClient(notionToken)

	return &Bot{
		TelegramToken: telegramToken,
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
		_, err := b.Bot.SendMessage(up.Message.Chat.Id, "Welcome to the pills-of-cs bot! Press `/pill` to request a pill or `/help` to get informations about the bot", "Markdown", up.Message.MessageId, false, false)
		if err != nil {
			return
		}
	case up.Message.Text == "/pill":
		var dst []byte
		_, err := parse(PILLS_ASSET, &dst)
		if err != nil {
			return
		}

		sp := SerializedPills{}
		err = json.Unmarshal(dst, &sp)
		if err != nil {
			return
		}
		rand.Seed(time.Now().UnixNano())

		randomIndex := makeTimestamp(len(sp.Pills))
		_, err = b.Bot.SendMessage(
			up.Message.Chat.Id,
			sp.Pills[randomIndex].Title+": "+sp.Pills[randomIndex].Body, "Markdown", up.Message.MessageId, false, false)
	case up.Message.Text == "/help":
		var dst []byte
		_, err := parse(HELP_MESSAGE_ASSET, &dst)
		if err != nil {
			return
		}
		_, err = b.Bot.SendMessage(up.Message.Chat.Id, string(dst), "Markdown", up.Message.MessageId, false, false)
		if err != nil {
			return
		}
	}
}

func makeTimestamp(len int) int64 {
	return (time.Now().UnixNano() / int64(time.Millisecond)) % int64(len)
}

package bot

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	adapters "pills-of-cs/adapters"
	"pills-of-cs/entities"
	"pills-of-cs/parser"
	"strings"
	"time"

	bt "github.com/SakoDroid/telego"
	cfg "github.com/SakoDroid/telego/configs"
	"github.com/SakoDroid/telego/objects"
	"github.com/joho/godotenv"
	"github.com/jomei/notionapi"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	NOTION_TOKEN       = "NOTION_TOKEN"
	TELEGRAM_TOKEN     = "TELEGRAM_TOKEN"
	PAGE_ID            = "48b530629463419ca92e22cc6ef50dab"
	PILLS_ASSET        = "./assets/pills.json"
	HELP_MESSAGE_ASSET = "./assets/help_message.txt"
)

type Bot struct {
	TelegramToken string
	Cfg           cfg.BotConfigs
	Bot           bt.Bot
	NotionClient  notionapi.Client
	HelpMessage   string
	UserRepo      adapters.UserRepo
	Pills         []entities.Pill
	Categories    map[string][]entities.Pill
}

func NewBotWithConfig(client *mongo.Client) (*Bot, error) {

	var (
		telegramToken string
		notionToken   string
	)

	var dst []byte
	_, err := parser.Parse(PILLS_ASSET, &dst)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	sp := entities.SerializedPills{}
	err = json.Unmarshal(dst, &sp)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	dst = []byte{}
	_, err = parser.Parse(HELP_MESSAGE_ASSET, &dst)
	if err != nil {
		return nil, err
	}

	err = godotenv.Load(".env")
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

	notion_client := notionapi.NewClient(notionapi.Token(notionToken))

	categories := map[string][]entities.Pill{}
	for _, p := range sp.Pills {
		for _, category := range p.Tags {
			categories[category] = []entities.Pill{p}
		}
	}
	return &Bot{
		TelegramToken: telegramToken,
		Cfg:           bot_config,
		Bot:           *b,
		NotionClient:  *notion_client,
		Pills:         sp.Pills,
		Categories:    categories,
		HelpMessage:   string(dst), // dst will contain bytes of the help message
		UserRepo: adapters.UserRepo{
			Client: client,
		},
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
	case strings.Contains(up.Message.Text, "/start"):
		_, err := b.Bot.SendMessage(up.Message.Chat.Id, "Welcome to the pills-of-cs bot! Press `/pill` to request a pill or `/help` to get informations about the bot", "Markdown", up.Message.MessageId, false, false)
		if err != nil {
			return
		}
	case strings.Contains(up.Message.Text, "/pill"):
		subscribedTags, err := b.UserRepo.GetTagsByUserId(up.Message.Chat.Id)
		if subscribedTags == nil {
			_, err := b.Bot.SendMessage(up.Message.Chat.Id, string(b.HelpMessage), "Markdown", up.Message.MessageId, false, false)
			if err != nil {
				return
			}
		}
		randomCategory := makeTimestamp(len(subscribedTags.Categories))
		randomIndex := makeTimestamp(len(b.Categories[subscribedTags.Categories[randomCategory]]))

		_, err = b.Bot.SendMessage(
			up.Message.Chat.Id,
			b.Categories[subscribedTags.Categories[randomCategory]][randomIndex].Title+": "+b.Categories[subscribedTags.Categories[randomCategory]][randomIndex].Body, "Markdown", up.Message.MessageId, false, false)
		if err != nil {
			return
		}
	case strings.Contains(up.Message.Text, "/help"):
		_, err := b.Bot.SendMessage(up.Message.Chat.Id, string(b.HelpMessage), "Markdown", up.Message.MessageId, false, false)
		if err != nil {
			return
		}
	case strings.Contains(up.Message.Text, "/choose_tags"):

		// /cmd args[0] args[1]
		args := strings.SplitN(up.Message.Text, " ", -1)

		// Replacing the underscores with spaces in the arguments.
		for i, a := range args {
			if strings.Contains(a, "_") {
				twoWordArg := strings.SplitN(a, "_", 2)
				args[i] = twoWordArg[0] + " " + twoWordArg[1]
			}
		}

		err := b.UserRepo.AddTagsToUser(up.Message.Chat.Id, args[1:])
		if err != nil {
			return
		}
		log.Printf("Return operation exit")
		_, err = b.Bot.SendMessage(up.Message.Chat.Id, "tags updated", "Markdown", up.Message.MessageId, false, false)
		if err != nil {
			return
		}

	}
}

func makeTimestamp(len int) int64 {
	return (time.Now().UnixNano() / int64(time.Millisecond)) % int64(len)
}

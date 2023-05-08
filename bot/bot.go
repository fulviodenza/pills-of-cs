package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pills-of-cs/adapters/ent"
	repositories "github.com/pills-of-cs/adapters/repositories"
	"github.com/pills-of-cs/parser"

	"github.com/pills-of-cs/entities"

	bt "github.com/SakoDroid/telego"
	cfg "github.com/SakoDroid/telego/configs"
	"github.com/SakoDroid/telego/objects"
	"github.com/jomei/notionapi"
)

// APIs constants
const (
	NOTION_TOKEN       = "NOTION_TOKEN"
	TELEGRAM_TOKEN     = "TELEGRAM_TOKEN"
	PAGE_ID            = "48b530629463419ca92e22cc6ef50dab"
	PILLS_ASSET        = "./assets/pills.json"
	HELP_MESSAGE_ASSET = "./assets/help_message.txt"
	DATABASE_URL       = "DATABASE_URL"
)

var (
	databaseUrl   string
	telegramToken string
	notionToken   string
)

type Bot struct {
	*entities.BotConf
}

// this cast force us to follow the given interface
// if the interface will not be followed, this will not compile
var _ entities.IBot = (*Bot)(nil)

func NewBotWithConfig() (*Bot, *ent.Client, error) {
	var dst []byte
	_, err := parser.Parse(PILLS_ASSET, &dst)
	if err != nil {
		return nil, nil, errors.New(err.Error())
	}

	sp := entities.SerializedPills{}
	err = json.Unmarshal(dst, &sp)
	if err != nil {
		return nil, nil, err
	}

	dst = []byte{}
	_, err = parser.Parse(HELP_MESSAGE_ASSET, &dst)
	if err != nil {
		return nil, nil, err
	}

	telegramToken = os.Getenv(TELEGRAM_TOKEN)
	databaseUrl = os.Getenv(DATABASE_URL)

	// connect to database with the env db uri
	client, err := ent.SetupAndConnectDatabase(databaseUrl)
	fmt.Println(client)
	if err != nil {
		log.Fatalf("[ent.SetupAndConnectDatabase]: error in db setup or connection: %v", err.Error())
	}

	// Configure telegram bot
	cf := cfg.DefaultUpdateConfigs()

	bot_config := cfg.BotConfigs{
		BotAPI: cfg.DefaultBotAPI,
		APIKey: telegramToken, UpdateConfigs: cf,
		Webhook:        false,
		LogFileAddress: cfg.DefaultLogFile,
	}

	b, err := bt.NewBot(&bot_config)
	if err != nil {
		return nil, nil, err
	}
	// end telegram configuration

	notion_client := notionapi.NewClient(notionapi.Token(notionToken))

	categories := map[string][]entities.Pill{}
	for _, p := range sp.Pills {
		for _, category := range p.Tags {
			categories[category] = []entities.Pill{p}
		}
	}
	return &Bot{
		&entities.BotConf{
			TelegramToken: telegramToken,
			Cfg:           bot_config,
			Bot:           *b,
			NotionClient:  *notion_client,
			Pills:         sp.Pills,
			Categories:    categories,
			HelpMessage:   string(dst),
			UserRepo: repositories.UserRepo{
				Client: client,
			},
		},
	}, client, err
}

func (b Bot) Start(ctx context.Context) error {
	//Register the channel
	messageChannel, _ := b.Bot.AdvancedMode().RegisterChannel("", "message")

	for {
		up := <-*messageChannel
		b.HandleMessage(ctx, up)
	}
}

func (ba Bot) HandleMessage(ctx context.Context, up *objects.Update) {
	switch {
	case strings.Contains(up.Message.Text, "/start"):
		ba.Run(ctx, up)
	case strings.Contains(up.Message.Text, "/pill"):
		ba.Pill(ctx, up)
	case strings.Contains(up.Message.Text, "/help"):
		ba.Help(ctx, up)
	case strings.Contains(up.Message.Text, "/choose_tags"):
		ba.ChooseTags(ctx, up)
	case strings.Contains(up.Message.Text, "/get_tags"):
		ba.GetTags(ctx, up)
	case strings.Contains(up.Message.Text, "/get_subscribed_categories"):
		ba.GetSubscribedTags(ctx, up)
	}
}

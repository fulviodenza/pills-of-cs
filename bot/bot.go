package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	objs "github.com/SakoDroid/telego/objects"
	"github.com/pills-of-cs/adapters/ent"
	repositories "github.com/pills-of-cs/adapters/repositories"
	"github.com/pills-of-cs/parser"
	"github.com/robfig/cron/v3"

	"github.com/pills-of-cs/entities"

	bt "github.com/SakoDroid/telego"
	cfg "github.com/SakoDroid/telego/configs"
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
		return nil, nil, err
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

	bot_config := &cfg.BotConfigs{
		BotAPI:         cfg.DefaultBotAPI,
		APIKey:         telegramToken,
		UpdateConfigs:  cfg.DefaultUpdateConfigs(),
		Webhook:        false,
		LogFileAddress: cfg.DefaultLogFile,
	}

	b, err := bt.NewBot(bot_config)
	if err != nil {
		return nil, nil, err
	}
	// end telegram configuration

	notion_client := notionapi.NewClient(notionapi.Token(notionToken))

	categories := map[string][]entities.Pill{}
	for _, p := range sp.Pills {
		for _, category := range p.Tags {
			categories[category] = append(categories[category], p)
		}
	}

	s := cron.New()
	s.Start()
	return &Bot{
		&entities.BotConf{
			Bot:          *b,
			NotionClient: *notion_client,
			Pills:        sp.Pills,
			Categories:   categories,
			HelpMessage:  string(dst),
			UserRepo: repositories.UserRepo{
				Client: client,
			},
			Schedules: map[string]time.Time{},
			Scheduler: s,
		},
	}, client, err
}

func (b *Bot) Start(ctx context.Context) error {
	var err error

	b.Bot.AddHandler("start", func(u *objs.Update) {
		err = b.Run(ctx, u)
		if err != nil {
			log.Printf("[Start]: %v\n", err)
		}
	}, "private")

	b.Bot.AddHandler("pill", func(u *objs.Update) {
		err = b.Pill(ctx, u)
		if err != nil {
			log.Printf("[Start]: %v\n", err)
		}
	}, "private")

	b.Bot.AddHandler("help", func(u *objs.Update) {
		err = b.Help(ctx, u)
		if err != nil {
			log.Printf("[Start]: %v\n", err)
		}
	}, "private")

	b.Bot.AddHandler("choose_tags", func(u *objs.Update) {
		err = b.ChooseTags(ctx, u)
		if err != nil {
			log.Printf("[Start]: %v\n", err)
		}
	}, "private")

	b.Bot.AddHandler("get_subscribed_categories", func(u *objs.Update) {
		err = b.GetSubscribedTags(ctx, u)
		if err != nil {
			log.Printf("[Start]: %v\n", err)
		}
	}, "private")

	b.Bot.AddHandler("schedule_pill", func(u *objs.Update) {
		err = b.SchedulePill(ctx, u)
		if err != nil {
			log.Printf("[Start]: %v\n", err)
		}
	}, "private")

	return err
}

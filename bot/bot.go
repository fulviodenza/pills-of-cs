package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/pills-of-cs/adapters/ent"
	repositories "github.com/pills-of-cs/adapters/repositories"
	"github.com/pills-of-cs/entities"
	"github.com/pills-of-cs/parser"

	bt "github.com/SakoDroid/telego"
	cfg "github.com/SakoDroid/telego/configs"
	"github.com/SakoDroid/telego/objects"
	objs "github.com/SakoDroid/telego/objects"
	"github.com/jomei/notionapi"
	"github.com/robfig/cron/v3"
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
	COMMAND_START                     = "/start"
	COMMAND_PILL                      = "/pill"
	COMMAND_HELP                      = "/help"
	COMMAND_CHOOSE_TAGS               = "/choose_tags"
	COMMAND_GET_SUBSCRIBED_CATEGORIES = "get_subscribed_categories"
	COMMAND_SCHEDULE_PILL             = "/schedule_pill"
)

var PRIVATE_CHAT_TYPE = "private"
var GROUP_CHAT_TYPE = "group"
var SUPERGROUP_CHAT_TYPE = "supergroup"

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
	ctx := context.Background()
	var dst []byte
	_, err := parser.Parse(PILLS_ASSET, &dst)
	if err != nil {
		return nil, nil, err
	}

	sp := entities.SerializedPills{}
	json.Unmarshal(dst, &sp)
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
		log.Printf("[ent.SetupAndConnectDatabase]: error in db setup or connection: %v", err.Error())
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

	userRepo := repositories.UserRepo{
		Client: client,
	}

	bot := &Bot{
		&entities.BotConf{
			Bot:          *b,
			NotionClient: *notion_client,
			Pills:        sp.Pills,
			Categories:   categories,
			HelpMessage:  string(dst),
			UserRepo:     userRepo,
			Schedules:    map[string]time.Time{},
		},
	}

	// setup the cron
	s := cron.New()

	// recovery crons from database
	crontabs, err := userRepo.GetAllCrontabs(ctx)
	for uid, cron := range crontabs {
		userId, err := strconv.Atoi(uid)
		if err != nil {
			continue
		}
		s.AddFunc(cron, func() {
			bot.Pill(context.Background(), &objs.Update{
				Message: &objs.Message{
					Chat: &objs.Chat{
						Id: userId,
					},
				},
			})
		})
	}
	s.Start()
	bot.Scheduler = s

	return bot, client, err
}

func (b *Bot) Start(ctx context.Context) {
	updateCh := b.Bot.GetUpdateChannel()
	go func() {
		for {
			update := <-*updateCh
			log.Printf("got update: %v\n", update.Update_id)
		}
	}()

	var handlers = map[string]func(ctx context.Context, up *objects.Update){
		COMMAND_START:                     b.Run,
		COMMAND_PILL:                      b.Pill,
		COMMAND_HELP:                      b.Help,
		COMMAND_CHOOSE_TAGS:               b.ChooseTags,
		COMMAND_GET_SUBSCRIBED_CATEGORIES: b.GetSubscribedTags,
		COMMAND_SCHEDULE_PILL:             b.SchedulePill,
	}
	for c, f := range handlers {
		b.Bot.AddHandler(c, func(u *objs.Update) {
			f(ctx, u)
		}, PRIVATE_CHAT_TYPE, GROUP_CHAT_TYPE, SUPERGROUP_CHAT_TYPE)
	}
}

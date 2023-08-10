package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/barthr/newsapi"
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
	NEWS_TOKEN         = "NEWS_TOKEN"
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
	COMMAND_GET_SUBSCRIBED_CATEGORIES = "/get_subscribed_categories"
	COMMAND_SCHEDULE_PILL             = "/schedule_pill"
	COMMAND_GET_TAGS                  = "/get_tags"
	COMMAND_NEWS                      = "/news"
	COMMAND_SCHEDULE_NEWS             = "/schedule_news"
)

var PRIVATE_CHAT_TYPE = "private"
var GROUP_CHAT_TYPE = "group"
var SUPERGROUP_CHAT_TYPE = "supergroup"

var (
	databaseUrl   string
	telegramToken string
	newsToken     string
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
	newsToken = os.Getenv(NEWS_TOKEN)

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

	// set news client
	newsClient := newsapi.NewClient(newsToken, newsapi.WithHTTPClient(http.DefaultClient), newsapi.WithUserAgent("pills-of-cs"))
	if err != nil {
		return nil, nil, err
	}

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
			NewsClient:   newsClient,
			Pills:        sp.Pills,
			Categories:   categories,
			HelpMessage:  string(dst),
			UserRepo:     userRepo,
			Schedules:    map[string]time.Time{},
		},
	}

	// setup the cron
	// recovery crons from database
	err = bot.recoverCrontabs(ctx, "pill")
	if err != nil {
		return nil, nil, err
	}
	err = bot.recoverCrontabs(ctx, "news")
	if err != nil {
		return nil, nil, err
	}

	return bot, client, err
}

func (b *Bot) recoverCrontabs(ctx context.Context, schedulerType string) error {
	s := cron.New()
	crontabs := map[string]string{}
	var err error

	switch schedulerType {
	case "news":
		crontabs, err = b.UserRepo.GetAllNewsCrontabs(ctx)
		if err != nil {
			return err
		}
	case "pill":
		crontabs, err = b.UserRepo.GetAllPillCrontabs(ctx)
		if err != nil {
			return err
		}
	}

	for uid, cron := range crontabs {
		userId, err := strconv.Atoi(uid)
		if err != nil {
			continue
		}
		s.AddFunc(cron, func() {
			switch schedulerType {
			case "news":
				b.News(context.Background(), &objs.Update{
					Message: &objs.Message{
						Chat: &objs.Chat{
							Id: userId,
						},
					},
				})
			case "pill":
				b.Pill(context.Background(), &objs.Update{
					Message: &objs.Message{
						Chat: &objs.Chat{
							Id: userId,
						},
					},
				})
			}
		})
	}

	s.Start()
	switch schedulerType {
	case "news":
		b.NewsScheduler = s
	case "pill":
		b.PillScheduler = s
	}

	return nil
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
		COMMAND_GET_TAGS:                  b.GetTags,
		COMMAND_START:                     b.Run,
		COMMAND_PILL:                      b.Pill,
		COMMAND_HELP:                      b.Help,
		COMMAND_CHOOSE_TAGS:               b.ChooseTags,
		COMMAND_GET_SUBSCRIBED_CATEGORIES: b.GetSubscribedTags,
		COMMAND_SCHEDULE_PILL:             b.SchedulePill,
		COMMAND_NEWS:                      b.News,
		COMMAND_SCHEDULE_NEWS:             b.ScheduleNews,
	}
	for c, f := range handlers {
		c := c
		f := f
		b.Bot.AddHandler(c, func(u *objs.Update) {
			if strings.Contains(u.Message.Text, c) {
				f(ctx, u)
			}
		}, PRIVATE_CHAT_TYPE, GROUP_CHAT_TYPE, SUPERGROUP_CHAT_TYPE)
	}
}

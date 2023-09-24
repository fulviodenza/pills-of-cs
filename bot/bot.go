package bot

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pills-of-cs/adapters/ent"
	repositories "github.com/pills-of-cs/adapters/repositories"
	"github.com/pills-of-cs/parser"

	bt "github.com/SakoDroid/telego"
	cfg "github.com/SakoDroid/telego/configs"
	"github.com/SakoDroid/telego/objects"
	objs "github.com/SakoDroid/telego/objects"
	"github.com/barthr/newsapi"
	"github.com/jomei/notionapi"
	"github.com/robfig/cron/v3"
)

// APIs constants
const (
	NOTION_TOKEN       = "NOTION_TOKEN"
	NOTION_DATABASE_ID = "NOTION_DATABASE_ID"
	TELEGRAM_TOKEN     = "TELEGRAM_TOKEN"
	NEWS_TOKEN         = "NEWS_TOKEN"
	DATABASE_URL       = "DATABASE_URL"
)

const (
	CATEGORIES_ASSET   = "./assets/categories.txt"
	HELP_MESSAGE_ASSET = "./assets/help_message.txt"
)

const (
	COMMAND_START                     = "/start"
	COMMAND_PILL                      = "/pill"
	COMMAND_HELP                      = "/help"
	COMMAND_CHOOSE_TAGS               = "/choose_tags"
	COMMAND_GET_SUBSCRIBED_CATEGORIES = "/get_subscribed_categories"
	COMMAND_SCHEDULE_PILL             = "/schedule_pill"
	COMMAND_GET_TAGS                  = "/get_tags"
	COMMAND_NEWS                      = "/news"
	COMMAND_SCHEDULE_NEWS             = "/schedule_news"
	COMMAND_UNSCHEDULE_NEWS           = "/unschedule_news"
	COMMAND_UNSCHEDULE_PILL           = "/unschedule_pill"
)

const (
	PRIVATE_CHAT_TYPE    = "private"
	GROUP_CHAT_TYPE      = "group"
	SUPERGROUP_CHAT_TYPE = "supergroup"
)

type Bot struct {
	*BotConf
}

type BotConf struct {
	TelegramClient bt.Bot
	NotionClient   notionapi.Client
	NewsClient     *newsapi.Client

	HelpMessage string
	Categories  []string

	UserRepo  repositories.UserRepo
	Schedules map[string]time.Time

	PillScheduler *cron.Cron
	PillsMu       sync.Mutex
	PillMap       map[string]cron.EntryID

	NewsScheduler *cron.Cron
	NewsMu        sync.Mutex
	NewsMap       map[string]cron.EntryID
}

type IBot interface {
	Start(ctx context.Context)
}

// this cast force us to follow the given interface
// if the interface will not be followed, this will not compile
var _ IBot = (*Bot)(nil)

// get variables from env
var (
	notionToken      = os.Getenv(NOTION_TOKEN)
	notionDatabaseId = os.Getenv(NOTION_DATABASE_ID)
	telegramToken    = os.Getenv(TELEGRAM_TOKEN)
	newsToken        = os.Getenv(NEWS_TOKEN)
	databaseUrl      = os.Getenv(DATABASE_URL)
)

func NewBotWithConfig() (*Bot, *ent.Client, error) {
	ctx := context.Background()

	helpMessage := make([]byte, 0)
	if err := parser.Read(HELP_MESSAGE_ASSET, &helpMessage); err != nil {
		return nil, nil, err
	}

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

	categories, err := parser.ParseCategories(CATEGORIES_ASSET)
	if err != nil {
		return nil, nil, err
	}

	userRepo := repositories.UserRepo{
		Client: client,
	}

	bot := &Bot{
		&BotConf{
			// client initialization
			NotionClient:   *notion_client,
			NewsClient:     newsClient,
			TelegramClient: *b,
			// static assets initialization
			Categories:  categories,
			HelpMessage: string(helpMessage),
			// database initialization
			UserRepo:  userRepo,
			Schedules: map[string]time.Time{},

			NewsMu:  sync.Mutex{},
			NewsMap: make(map[string]cron.EntryID),

			PillsMu: sync.Mutex{},
			PillMap: make(map[string]cron.EntryID),
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
		cId, err := s.AddFunc(cron, func() {
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
		switch schedulerType {
		case "news":
			b.BotConf.NewsMu.Lock()
			b.BotConf.NewsMap[uid] = cId
			b.BotConf.NewsMu.Unlock()
		case "pill":
			b.BotConf.PillsMu.Lock()
			b.BotConf.PillMap[uid] = cId
			b.BotConf.PillsMu.Unlock()
		}
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
	updateCh := b.TelegramClient.GetUpdateChannel()
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
		COMMAND_UNSCHEDULE_NEWS:           b.UnscheduleNews,
		COMMAND_UNSCHEDULE_PILL:           b.UnschedulePill,
	}
	for c, f := range handlers {
		c := c
		f := f
		b.TelegramClient.AddHandler(c, func(u *objs.Update) {
			if strings.Contains(u.Message.Text, c) {
				f(ctx, u)
			}
		}, PRIVATE_CHAT_TYPE, GROUP_CHAT_TYPE, SUPERGROUP_CHAT_TYPE)
	}
}

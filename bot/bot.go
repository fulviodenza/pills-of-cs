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

	"github.com/pills-of-cs/bot/types"
	"github.com/pills-of-cs/utils"

	bt "github.com/SakoDroid/telego/v2"
	cfg "github.com/SakoDroid/telego/v2/configs"
	objs "github.com/SakoDroid/telego/v2/objects"
	"github.com/barthr/newsapi"
	"github.com/jomei/notionapi"
	"github.com/robfig/cron/v3"
)

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
	COMMAND_PILL                = "/pill"
	COMMAND_HELP                = "/help"
	COMMAND_CHOOSE_TAGS         = "/choose_tags"
	COMMAND_GET_SUBSCRIBED_TAGS = "/get_subscribed_tags"
	COMMAND_SCHEDULE_PILL       = "/schedule_pill"
	COMMAND_GET_TAGS            = "/get_tags"
	COMMAND_NEWS                = "/news"
	COMMAND_SCHEDULE_NEWS       = "/schedule_news"
	COMMAND_UNSCHEDULE_NEWS     = "/unschedule_news"
	COMMAND_UNSCHEDULE_PILL     = "/unschedule_pill"
	COMMAND_QUIZ                = "/quiz"
)

const (
	PRIVATE_CHAT_TYPE    = "private"
	GROUP_CHAT_TYPE      = "group"
	SUPERGROUP_CHAT_TYPE = "supergroup"
)

type Bot struct {
	TelegramClient bt.Bot
	NotionClient   notionapi.Client
	NewsClient     *newsapi.Client
	UserRepo       repositories.IUserRepo

	helpMessage string
	Categories  []string

	Schedules map[string]time.Time

	PillScheduler *cron.Cron
	PillsMu       sync.Mutex
	PillMap       map[string]cron.EntryID

	NewsScheduler *cron.Cron
	NewsMu        sync.Mutex
	NewsMap       map[string]cron.EntryID
}

// this cast force us to follow the given interface
// if the interface will not be followed, this will not compile
var _ types.IBot = (*Bot)(nil)

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

	bot_config := &cfg.BotConfigs{
		BotAPI:         cfg.DefaultBotAPI,
		APIKey:         telegramToken,
		UpdateConfigs:  cfg.DefaultUpdateConfigs(),
		Webhook:        false,
		LogFileAddress: cfg.DefaultLogFile,
	}

	notionClient := notionapi.NewClient(notionapi.Token(notionToken))

	newsClient := newsapi.NewClient(newsToken, newsapi.WithHTTPClient(http.DefaultClient), newsapi.WithUserAgent("pills-of-cs"))

	bot := &Bot{
		NotionClient: *notionClient,
		NewsClient:   newsClient,
		Schedules:    map[string]time.Time{},

		NewsMu:  sync.Mutex{},
		NewsMap: make(map[string]cron.EntryID),

		PillsMu: sync.Mutex{},
		PillMap: make(map[string]cron.EntryID),
	}

	bot.loadHelpMessage()

	client, err := ent.SetupAndConnectDatabase(databaseUrl)
	fmt.Println(client)
	if err != nil {
		log.Printf("[ent.SetupAndConnectDatabase]: error in db setup or connection: %v", err.Error())
	}
	bot.SetUserRepo(repositories.NewUserRepo(client), nil)

	b, err := bt.NewBot(bot_config)
	if err != nil {
		return nil, nil, err
	}
	bot.SetTelegramClient(*b)

	categories, err := parser.ParseCategories(CATEGORIES_ASSET)
	if err != nil {
		return nil, nil, err
	}
	bot.SetCategories(categories)

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

func (b *Bot) Start(ctx context.Context) {
	updateCh := b.TelegramClient.GetUpdateChannel()
	go func() {
		for {
			update := <-*updateCh
			log.Printf("got update: %v\n", update.Update_id)
		}
	}()

	var handlers = b.initializeHandlers()
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

func (b *Bot) GetHelpMessage() string {
	return b.helpMessage
}

func (b *Bot) SetCategories(categories []string) {
	b.Categories = categories
}

func (b *Bot) GetCategories() []string {
	return b.Categories
}

func (b *Bot) SendMessage(msg string, up *objs.Update, formatMarkdown bool) error {
	parseMode := ""
	if formatMarkdown {
		parseMode = "Markdown"
	}

	if len(msg) >= MAX_LEN_MESSAGE {
		msgs := utils.SplitString(msg)
		for _, m := range msgs {
			_, err := b.TelegramClient.SendMessage(up.Message.Chat.Id, m, parseMode, 0, false, false)
			if err != nil {
				log.Printf("[SendMessage]: sending message to user: %v", err.Error())
				return err
			}
		}
	} else {
		_, err := b.TelegramClient.SendMessage(up.Message.Chat.Id, msg, parseMode, 0, false, false)
		if err != nil {
			log.Printf("[SendMessage]: sending message to user: %v", err.Error())
			return err
		}
	}
	return nil
}

func (b *Bot) SetUserRepo(userRepo repositories.IUserRepo, ch chan interface{}) {
	b.UserRepo = userRepo
}

func (b *Bot) GetUserRepo() repositories.IUserRepo {
	return b.UserRepo
}

func (b *Bot) loadHelpMessage() {
	helpMessage := make([]byte, 0)
	err := parser.Read(HELP_MESSAGE_ASSET, &helpMessage)
	if err != nil {
		log.Fatalf("Failed to load help message: %v", err)
	}
	b.helpMessage = string(helpMessage)
}

func (b *Bot) setCron(ctx context.Context, up *objs.Update, schedulerType string) (strings.Builder, error) {
	var (
		crontab string
		err     error
		msg     strings.Builder
	)
	id := strconv.Itoa(up.Message.Chat.Id)
	// args[1] contains the time HH:MM, args[2] contains the timezone
	args := strings.SplitN(up.Message.Text, " ", -1)
	if len(args) != 3 {
		msg.WriteString("Failed parsing provided time")
	} else {
		crontab, err = parser.ValidateSchedule(args[1], args[2])
		if err != nil {
			msg.WriteString("Failed parsing provided time")
		}
	}
	switch schedulerType {
	case "pill":
		err = b.UserRepo.SavePillSchedule(ctx, id, crontab)
		if err != nil {
			log.Printf("[SchedulePill]: failed saving time: %v", err.Error())
			msg.WriteString("failed saving time")
		}
	case "news":
		err = b.UserRepo.SaveNewsSchedule(ctx, id, crontab)
		if err != nil {
			log.Printf("[SchedulePill]: failed saving time: %v", err.Error())
			msg.WriteString("failed saving time")
		}

	}

	// run the goroutine with the cron
	go func(ctx context.Context, u *objs.Update) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("[SchedulePill]: Recovering from panic:", r)
			}
		}()
		switch schedulerType {
		case "pill":
			uid := strconv.Itoa(up.Message.Chat.Id)
			cronId, err := b.PillScheduler.AddFunc(crontab, func() {
				NewPillCommand(b)(ctx, up)
			})
			if err != nil {
				log.Println("[SchedulePill]: got error:", err)
				return
			}
			b.PillsMu.Lock()
			b.PillMap[uid] = cronId
			b.PillsMu.Unlock()
		case "news":
			uid := strconv.Itoa(up.Message.Chat.Id)
			cronId, err := b.NewsScheduler.AddFunc(crontab, func() {
				NewNewsCommand(b)(ctx, up)
			})
			if err != nil {
				log.Println("[ScheduleNews]: got error:", err)
				return
			}
			b.NewsMu.Lock()
			b.NewsMap[uid] = cronId
			b.NewsMu.Unlock()
		}
	}(ctx, up)

	// the human readable format is with times[0] (hours) first
	msg.WriteString(fmt.Sprintf("Crontab for your pill `%s`", crontab))
	return msg, nil
}

func (b *Bot) SetTelegramClient(bot bt.Bot) {
	b.TelegramClient = bot
}
func (b *Bot) GetTelegramClient() *bt.Bot {
	return &b.TelegramClient
}

func (b *Bot) recoverCrontabs(ctx context.Context, schedulerType string) error {
	s := cron.New()
	crontabs := map[string]string{}
	var err error

	newsCommand := NewNewsCommand(b)
	pillCommand := NewPillCommand(b)

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
				newsCommand(ctx, &objs.Update{
					Message: &objs.Message{
						Chat: &objs.Chat{
							Id: userId,
						},
					},
				})
			case "pill":
				pillCommand(ctx, &objs.Update{
					Message: &objs.Message{
						Chat: &objs.Chat{
							Id: userId,
						},
					},
				})
			}
		})
		if err != nil {
			continue
		}

		switch schedulerType {
		case "news":
			b.NewsMu.Lock()
			b.NewsMap[uid] = cId
			b.NewsMu.Unlock()
		case "pill":
			b.PillsMu.Lock()
			b.PillMap[uid] = cId
			b.PillsMu.Unlock()
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

func (b *Bot) initializeHandlers() map[string]func(ctx context.Context, up *objs.Update) {
	return map[string]func(ctx context.Context, up *objs.Update){
		COMMAND_GET_TAGS:            NewGetTagsCommand(b),
		COMMAND_PILL:                NewPillCommand(b),
		COMMAND_HELP:                NewHelpCommand(b),
		COMMAND_CHOOSE_TAGS:         NewChooseTagsCommand(b),
		COMMAND_GET_SUBSCRIBED_TAGS: NewGetSubscribedTagsCommand(b),
		COMMAND_SCHEDULE_PILL:       NewSchedulePillCommand(b),
		COMMAND_NEWS:                NewNewsCommand(b),
		COMMAND_SCHEDULE_NEWS:       NewScheduleNewsCommand(b),
		COMMAND_UNSCHEDULE_NEWS:     NewUnscheduleNewsCommand(b),
		COMMAND_UNSCHEDULE_PILL:     NewUnschedulePillCommand(b),
		COMMAND_QUIZ:                NewQuizCommand(b),
	}
}

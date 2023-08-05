package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	objs "github.com/SakoDroid/telego/objects"
	"github.com/go-co-op/gocron"
	"github.com/pills-of-cs/adapters/ent"
	repositories "github.com/pills-of-cs/adapters/repositories"
	"github.com/pills-of-cs/parser"

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

	s := gocron.NewScheduler(time.UTC)

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

// func (b Bot) Start(ctx context.Context) error {
// var err error = nil
// //Register the channel
// messageChannel, err := b.Bot.AdvancedMode().RegisterChannel("", "message")
// if err != nil {
// 	log.Fatalf("[Start]: got error: %v", err)
// 	return err
// }
// defer func() {
// 	b.Bot.AdvancedMode().UnRegisterChannel("", "message")
// 	close(*messageChannel)
// }()

// if err != nil {
// 	log.Fatalf("[Start]: got error: %v", err)
// 	return err
// }

// for {
// 	up := <-*messageChannel
// 	err = b.HandleMessage(ctx, up)
// 	if err != nil {
// 		log.Fatalf("[Start]: got error: %v", err)
// 		return err
// 	}
// }
// }

func (b *Bot) Start(ctx context.Context) error {

	var err error
	// updateChannel := b.Bot.GetUpdateChannel()
	b.Bot.AddHandler("hi", func(u *objs.Update) {

		//Register channel for receiving messages from this chat.
		cc, _ := b.Bot.AdvancedMode().RegisterChannel(strconv.Itoa(u.Message.Chat.Id), "message")

		//Sends back a message
		_, err := b.Bot.SendMessage(u.Message.Chat.Id, "hi to you too, send me a location", "", u.Message.MessageId, false, false)
		if err != nil {
			fmt.Println(err)
		}

		//Waits for an update from this chat
		up := <-*cc

		//Sends back the received location
		_, err = b.Bot.SendLocation(up.Message.Chat.Id, false, false, up.Message.Location.Latitude, up.Message.Location.Longitude, up.Message.Location.HorizontalAccuracy, up.Message.MessageId)

		if err != nil {
			log.Printf("[Start]: %v\n", err)
		}
	}, "private")

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

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

const (
	NOTION_TOKEN       = "NOTION_TOKEN"
	TELEGRAM_TOKEN     = "TELEGRAM_TOKEN"
	PAGE_ID            = "48b530629463419ca92e22cc6ef50dab"
	PILLS_ASSET        = "./assets/pills.json"
	HELP_MESSAGE_ASSET = "./assets/help_message.txt"
)

var databaseUrl, pgUser, pgPwd, pgPort, pgHost, phPort, pgDatabase, telegramToken string

const (
	DATABASE_URL = "DATABASE_URL"
	PG_USER      = "PGUSER"
	PG_PWD       = "PGPASSWORD"
	PG_HOST      = "PGHOST"
	PG_PORT      = "PGPORT"
	PG_DATABASE  = "PGDATABASE"
)

type Bot struct {
	*entities.BotConf
}

// this cast force us to follow the given interface
// if the interface will not be followed, this will not compile
var _ entities.IBot = (*Bot)(nil)

func NewBotWithConfig() (*Bot, *ent.Client, error) {
	var (
		notionToken string
		dbUri       string
	)

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

	load()

	dbUri = databaseUrl + "://" + pgUser + ":" + pgPwd + "@" + pgHost + ":" + pgPort + "/" + pgDatabase

	// connect to database with the env db uri
	client, err := ent.SetupAndConnectDatabase(dbUri)
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
		ba.start(ctx, up)
	case strings.Contains(up.Message.Text, "/pill"):
		ba.pill(ctx, up)
	case strings.Contains(up.Message.Text, "/help"):
		ba.help(ctx, up)
	case strings.Contains(up.Message.Text, "/choose_tags"):
		ba.chooseTags(ctx, up)
	case strings.Contains(up.Message.Text, "/get_tags"):
		ba.getTags(ctx, up)
	}
}

func load() {
	// get environment variables from env
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if pair[0] == TELEGRAM_TOKEN {
			telegramToken = pair[1]
		}
		// postgresql://postgres:changeme@localhost:5435/notion_on_the_go
		if pair[0] == DATABASE_URL {
			databaseUrl = pair[1]
		}

		if pair[0] == PG_DATABASE {
			pgDatabase = pair[1]
		}

		if pair[0] == PG_HOST {
			pgHost = pair[1]
		}

		if pair[0] == PG_USER {
			pgUser = pair[1]
		}

		if pair[0] == PG_PWD {
			pgPwd = pair[1]
		}

		if pair[0] == PG_PORT {
			pgPort = pair[1]
		}
	}
}

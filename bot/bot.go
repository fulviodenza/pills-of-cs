package bot

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pills-of-cs/adapters/ent"
	repositories "github.com/pills-of-cs/adapters/repositories"
	"github.com/pills-of-cs/parser"

	"github.com/pills-of-cs/entities"

	bt "github.com/SakoDroid/telego"
	cfg "github.com/SakoDroid/telego/configs"
	"github.com/SakoDroid/telego/objects"
	"github.com/joho/godotenv"
	"github.com/jomei/notionapi"
)

const (
	NOTION_TOKEN       = "NOTION_TOKEN"
	TELEGRAM_TOKEN     = "TELEGRAM_TOKEN"
	PAGE_ID            = "48b530629463419ca92e22cc6ef50dab"
	PILLS_ASSET        = "./assets/pills.json"
	HELP_MESSAGE_ASSET = "./assets/help_message.txt"
)

type Bot struct {
	*entities.BotConf
}

// this cast force us to follow the given interface
// if the interface will not be followed, this will not compile
var _ entities.IBot = (*Bot)(nil)

func NewBotWithConfig(client *ent.Client) (*Bot, error) {
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
		return nil, err
	}

	dst = []byte{}
	_, err = parser.Parse(HELP_MESSAGE_ASSET, &dst)
	if err != nil {
		return nil, err
	}

	err = godotenv.Load(".env")
	if err != nil {
		log.Fatalf("[godotenv.Load]: failed loading .env file: %v", err.Error())
		return nil, err
	}
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
		&entities.BotConf{
			TelegramToken: telegramToken,
			Cfg:           bot_config,
			Bot:           *b,
			NotionClient:  *notion_client,
			Pills:         sp.Pills,
			Categories:    categories,
			HelpMessage:   string(dst), // dst will contain bytes of the help message
			UserRepo: repositories.UserRepo{
				Client: client,
			},
		},
	}, nil
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
	case strings.Contains(up.Message.Text, "/pill"):
		subscribedTags, err := ba.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
		if err != nil {
			log.Fatalf("[b.UserRepo.GetTagsByUserId]: failed getting tags: %v", err.Error())
			return
		}
		if subscribedTags == nil {
			_, err := ba.Bot.SendMessage(up.Message.Chat.Id, string(ba.HelpMessage), "Markdown", up.Message.MessageId, false, false)
			if err != nil {
				return
			}
		}

		var randomCategory, randomIndex int64
		var randomCategoryP []entities.Pill
		rand.Seed(time.Now().Unix())

		if len(subscribedTags) > 0 {
			randomCategory = makeTimestamp(len(subscribedTags))
			randomIndex = makeTimestamp(len(ba.Categories[subscribedTags[randomCategory]]))
			_, err = ba.Bot.SendMessage(
				up.Message.Chat.Id,
				ba.Categories[subscribedTags[randomCategory]][randomIndex].Title+": "+ba.Categories[subscribedTags[randomCategory]][randomIndex].Body, "Markdown", up.Message.MessageId, false, false)
			if err != nil {
				return
			}
		} else {
			randomCategoryP = pick(ba.Categories)
			randomIndex = makeTimestamp(len(randomCategoryP))
			_, err = ba.Bot.SendMessage(
				up.Message.Chat.Id,
				randomCategoryP[randomIndex].Title+": "+randomCategoryP[randomIndex].Body, "Markdown", up.Message.MessageId, false, false)
			if err != nil {
				return
			}

		}

	case strings.Contains(up.Message.Text, "/help"):
		_, err := ba.Bot.SendMessage(up.Message.Chat.Id, string(ba.HelpMessage), "Markdown", up.Message.MessageId, false, false)
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

		err := ba.UserRepo.AddTagsToUser(ctx, strconv.Itoa(up.Message.Chat.Id), args[1:])
		if err != nil {
			return
		}

		log.Printf("Return operation exit")
		_, err = ba.Bot.SendMessage(up.Message.Chat.Id, "tags updated", "Markdown", up.Message.MessageId, false, false)
		if err != nil {
			return
		}
	case strings.Contains(up.Message.Text, "/get_tags"):
		msg := ""
		for k := range ba.Categories {
			msg += "- " + k + "\n"
		}
		_, err := ba.Bot.SendMessage(up.Message.Chat.Id, msg, "Markdown", up.Message.MessageId, false, false)
		if err != nil {
			return
		}
	}
}

func makeTimestamp(len int) int64 {
	millisec := int64(time.Millisecond)
	now := time.Now().UnixNano()
	division := now / millisec
	return (division) % int64(len)
}

func pick[K comparable, V any](m map[K]V) V {
	k := rand.Intn(len(m))
	for _, x := range m {
		if k == 0 {
			return x
		}
		k--
	}
	panic("unreachable")
}

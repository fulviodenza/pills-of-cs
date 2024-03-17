package bot

import (
	"context"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/barthr/newsapi"
	"github.com/pills-of-cs/bot/types"
)

const MAX_LEN_MESSAGE = 4096

var _ types.ICommand = (*NewsCommand)(nil)

type NewsCommand struct {
	Bot *Bot
}

func NewNewsCommand(bot *Bot) types.Command {
	hc := HelpCommand{
		Bot: bot,
	}
	return hc.Execute
}

func (nc NewsCommand) Execute(ctx context.Context, up *objects.Update) {
	var msg strings.Builder
	newsCategories := ""
	categories, err := nc.Bot.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil || len(categories) == 0 {
		newsCategories += "technology"
	}

	for _, c := range categories {
		sourceParams := &newsapi.EverythingParameters{
			Keywords: c + "&",
			Language: "en",
		}
		sources, err := nc.Bot.NewsClient.GetEverything(ctx, sourceParams)
		sort.Slice(sources.Articles, func(i, j int) bool {
			return sources.Articles[i].PublishedAt.After(sources.Articles[j].PublishedAt)
		})
		if err == nil && len(sources.Articles) != 0 {
			articles := sources.Articles[:10]
			rand.New(rand.NewSource(time.Now().UnixNano()))
			rand.Shuffle(len(articles), func(i, j int) { (articles)[i], (articles)[j] = (articles)[j], (articles)[i] })

			for _, a := range articles {
				if len(msg.String()) < MAX_LEN_MESSAGE-1 {
					description := strings.Trim(a.Description, "\n")
					msg.WriteString("🔴 " + a.Title + "\n" + description + "\n" + "from " + a.URL + "\n")
				} else {
					break
				}
			}
		} else {
			log.Printf("err: %v; articles len: %v", err, len(sources.Articles))
			msg.WriteString("sources missing!")
		}
	}
	nc.Bot.SendMessage(msg.String(), up, false)
}

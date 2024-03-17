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

const (
	MAX_LEN_MESSAGE = 4096
	SOURCE_MISSING  = "sources missing!"
)

var _ types.ICommand = (*NewsCommand)(nil)

type NewsCommand struct {
	Bot types.IBot
}

func NewNewsCommand(bot *Bot) types.Command {
	hc := NewsCommand{
		Bot: bot,
	}
	return hc.Execute
}

func (nc NewsCommand) Execute(ctx context.Context, up *objects.Update) {
	var msg strings.Builder
	categories, err := nc.Bot.GetUserRepo().GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil || len(categories) == 0 {
		categories = append(categories, "technology")
	}

	for _, c := range categories {
		sourceParams := &newsapi.EverythingParameters{
			Keywords: c + "&",
			Language: "en",
		}

		newsClient := nc.Bot.GetNewsClient()
		sources, err := newsClient.GetEverything(ctx, sourceParams)
		sort.Slice(sources.Articles, func(i, j int) bool {
			return sources.Articles[i].PublishedAt.After(sources.Articles[j].PublishedAt)
		})
		if err != nil || len(sources.Articles) == 0 {
			log.Printf("err: %v; articles len: %v", err, len(sources.Articles))
			msg.WriteString(SOURCE_MISSING)
			continue
		}

		i := 0
		var articles = make([]newsapi.Article, 0)
		for i < 10 && i < len(sources.Articles) {
			articles = append(articles, sources.Articles[i])
			i++
		}
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
	}
	nc.Bot.SendMessage(msg.String(), up, false)
}

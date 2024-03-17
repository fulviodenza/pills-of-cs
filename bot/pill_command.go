package bot

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/jomei/notionapi"
	"github.com/pills-of-cs/bot/types"
	"github.com/pills-of-cs/utils"
)

var _ types.ICommand = (*PillCommand)(nil)

type PillCommand struct {
	Bot *Bot
}

func NewPillCommand(bot *Bot) types.Command {
	hc := HelpCommand{
		Bot: bot,
	}
	return hc.Execute
}

func (pc PillCommand) Execute(ctx context.Context, up *objects.Update) {
	var msg strings.Builder
	subscribedTags, err := pc.Bot.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil {
		log.Printf("[Pill]: failed getting tags: %v", err.Error())
	}

	choosenCategory := pc.Bot.Categories[utils.MakeTimestamp(len(pc.Bot.Categories))]
	if len(subscribedTags) > 0 {
		rand.New(rand.NewSource(time.Now().UnixNano()))
		choosenCategory = pc.Bot.Categories[utils.MakeTimestamp(len(subscribedTags))]
	}

	res, err := pc.Bot.NotionClient.Database.Query(ctx, notionapi.DatabaseID(notionDatabaseId), &notionapi.DatabaseQueryRequest{
		Filter: notionapi.PropertyFilter{
			Property: "Tags",
			MultiSelect: &notionapi.MultiSelectFilterCondition{
				Contains: choosenCategory,
			},
		},
	})
	if err != nil {
		log.Printf("[Pill]: failed retrieving pill: %v", err.Error())
	}

	if res != nil {
		row := types.NotionDbRow{}
		var rowProps []byte

		if rowProps, err = json.Marshal(res.Results[utils.MakeTimestamp(len(res.Results))].Properties); err != nil {
			log.Printf("[Pill]: failed marshaling pill: %v", err.Error())
		}
		if err = json.Unmarshal(rowProps, &row); err != nil {
			log.Printf("[Pill]: failed unmarshaling pill: %v", err.Error())
		}

		msg.WriteString(row.Name.Title[0].Text.Content + ": ")
		for _, c := range row.Text.RichText {
			msg.WriteString(c.Text.Content)
		}
		pc.Bot.SendMessage(msg.String(), up, true)
	}
}

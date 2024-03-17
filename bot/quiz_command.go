package bot

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/jomei/notionapi"
	"github.com/pills-of-cs/bot/types"
	"github.com/pills-of-cs/utils"
)

var _ types.ICommand = (*QuizCommand)(nil)

const QuizCategory = "quiz"

type QuizCommand struct {
	Bot *Bot
}

func NewQuizCommand(bot *Bot) types.Command {
	hc := QuizCommand{
		Bot: bot,
	}
	return hc.Execute
}

// Execute method to process the help command.
func (qc *QuizCommand) Execute(ctx context.Context, update *objects.Update) {
	var optionsAnswer []string
	var options []string
	question, optionsRaw := "", ""
	correctIndex := -1

	rawPoll, err := qc.Bot.NotionClient.Database.Query(ctx, notionapi.DatabaseID(notionDatabaseId), &notionapi.DatabaseQueryRequest{
		Filter: notionapi.PropertyFilter{
			Property: "Tags",
			MultiSelect: &notionapi.MultiSelectFilterCondition{
				Contains: QuizCategory,
			},
		},
	})
	if err != nil {
		log.Printf("[Pill]: failed retrieving pill: %v", err.Error())
	}
	if rawPoll != nil {
		row := types.NotionDbRow{}
		var rowProps []byte

		if rowProps, err = json.Marshal(rawPoll.Results[utils.MakeTimestamp(len(rawPoll.Results))].Properties); err != nil {
			log.Printf("[Pill]: failed marshaling pill: %v", err.Error())
		}
		if err = json.Unmarshal(rowProps, &row); err != nil {
			log.Printf("[Pill]: failed unmarshaling pill: %v", err.Error())
		}
		question = row.Name.Title[0].Text.Content
		for _, c := range row.Text.RichText {
			optionsRaw += c.Text.Content
		}
		optionsAnswer = strings.Split(optionsRaw, ";")
		options = strings.Split(optionsAnswer[0], ",")
		for i, o := range options {
			if o == optionsAnswer[1] {
				correctIndex = i
				break
			}
		}
	}
	poll, err := qc.Bot.TelegramClient.CreatePoll(update.Message.Chat.Id, question, QuizCategory)
	if err != nil {
		log.Printf("[Quiz] error creating poll: %v", err)
	}
	for _, o := range options {
		poll.AddOption(o)
	}
	poll.SetCorrectOption(correctIndex)
	poll.Send(false, false, update.Message.MessageId)
}

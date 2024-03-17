package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/pills-of-cs/bot/types"
)

var _ types.ICommand = (*GetTagsCommand)(nil)

const EMPTY_CATEGORIES = "empty categories"

type GetTagsCommand struct {
	Bot types.IBot
}

func NewGetTagsCommand(bot *Bot) types.Command {
	hc := GetTagsCommand{
		Bot: bot,
	}
	return hc.Execute
}

func (gc *GetTagsCommand) Execute(ctx context.Context, update *objects.Update) {
	var msg strings.Builder

	categories := gc.Bot.GetCategories()
	if len(categories) == 0 {
		msg.WriteString(EMPTY_CATEGORIES)
	} else {
		for _, v := range categories {
			msg.WriteString(fmt.Sprintf("- %s\n", v))
		}
	}

	gc.Bot.SendMessage(msg.String(), update, false)
}

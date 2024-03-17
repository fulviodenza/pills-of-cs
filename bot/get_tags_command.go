package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/pills-of-cs/bot/types"
)

var _ types.ICommand = (*GetTagsCommand)(nil)

type GetTagsCommand struct {
	Bot *Bot
}

func NewGetTagsCommand(bot *Bot) types.Command {
	hc := HelpCommand{
		Bot: bot,
	}
	return hc.Execute
}

func (gc *GetTagsCommand) Execute(ctx context.Context, update *objects.Update) {
	var msg strings.Builder

	categories := gc.Bot.GetCategories()
	if len(categories) == 0 {
		msg.WriteString("empty categories")
	} else {
		for _, v := range categories {
			msg.WriteString(fmt.Sprintf("- %s\n", v))
		}
	}

	gc.Bot.SendMessage(msg.String(), update, false)
}

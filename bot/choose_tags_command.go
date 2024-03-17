package bot

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/pills-of-cs/bot/types"
)

var _ types.ICommand = (*ChooseTagsCommand)(nil)

const TAGS_UPDATED = "tags updated"

type ChooseTagsCommand struct {
	Bot types.IBot
}

func NewChooseTagsCommand(bot *Bot) types.Command {
	hc := ChooseTagsCommand{
		Bot: bot,
	}
	return hc.Execute
}

func (cc *ChooseTagsCommand) Execute(ctx context.Context, update *objects.Update) {
	// /cmd args[0] args[1]
	args := strings.Split(update.Message.Text, " ")

	// Replacing the underscores with spaces in the arguments.
	// This is done for more-than-one-word tags.
	// Indeed, /choose_tags command requires:
	// /choose_tags distributed_systems for example
	for i, a := range args {
		if strings.Contains(a, "_") {
			twoWordArg := strings.SplitN(a, "_", 2)
			args[i] = twoWordArg[0] + " " + twoWordArg[1]
		}
	}

	err := cc.Bot.GetUserRepo().AddTagsToUser(ctx, strconv.Itoa(update.Message.Chat.Id), args)
	if err != nil {
		log.Printf("[ChooseTags]: failed adding tag to user: %v", err.Error())
		return
	}

	cc.Bot.SendMessage(TAGS_UPDATED, update, false)
}

package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/pills-of-cs/bot/types"
)

var _ types.ICommand = (*RemoveCommand)(nil)

type RemoveCommand struct {
	Bot types.IBot
}

func NewRemoveCommand(bot *Bot) types.Command {
	hc := RemoveCommand{
		Bot: bot,
	}
	return hc.Execute
}

func (hc *RemoveCommand) Execute(ctx context.Context, update *objects.Update) {
	args := strings.Split(update.Message.Text, " ")

	for i, a := range args {
		if strings.Contains(a, "_") {
			twoWordArg := strings.SplitN(a, "_", 2)
			args[i] = twoWordArg[0] + " " + twoWordArg[1]
		}
	}

	var msg strings.Builder
	newTags, err := hc.Bot.GetUserRepo().RemoveTagsFromUser(ctx, strconv.Itoa(update.Message.Chat.Id), args)
	if err != nil {
		log.Printf("Failed to removing tags: %v", err)
	}

	if len(newTags) > 0 {
		msg.WriteString("Tags updated:\n")
		for _, v := range newTags {
			msg.WriteString(fmt.Sprintf("- %s\n", v))
		}
	} else {
		msg.WriteString("No tags anymore!")
	}

	if err = hc.Bot.SendMessage(msg.String(), update, false); err != nil {
		log.Printf("Failed sending message: %v", err)
	}
}

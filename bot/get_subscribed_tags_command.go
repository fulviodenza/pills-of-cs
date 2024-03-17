package bot

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/pills-of-cs/bot/types"
	"github.com/pills-of-cs/utils"
)

var _ types.ICommand = (*GetSubscribedTagsCommand)(nil)

type GetSubscribedTagsCommand struct {
	Bot *Bot
}

func NewGetSubscribedTagsCommand(bot *Bot) types.Command {
	hc := GetSubscribedTagsCommand{
		Bot: bot,
	}
	return hc.Execute
}

// Execute method to process the GetSubscribedTags command.
func (gc *GetSubscribedTagsCommand) Execute(ctx context.Context, update *objects.Update) {
	var msg strings.Builder
	tags, err := gc.Bot.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(update.Message.Chat.Id))
	if err != nil {
		log.Printf("[getSubscribedTags]: failed getting tags by user id: %v", err.Error())
	}
	msg.WriteString(utils.AggregateTags(tags))
	if len(tags) == 0 {
		msg.WriteString("You are not subscribed to any tag!\nSubscribe one with /choose_tags [tag] command!")
	}

	gc.Bot.SendMessage(msg.String(), update, false)
}

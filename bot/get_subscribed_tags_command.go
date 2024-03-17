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

const NOT_SUBSCRIBED_MSG = "You are not subscribed to any tag!\nSubscribe one with /choose_tags [tag] command!"

type GetSubscribedTagsCommand struct {
	Bot types.IBot
}

func NewGetSubscribedTagsCommand(bot *Bot) types.Command {
	hc := GetSubscribedTagsCommand{
		Bot: bot,
	}
	return hc.Execute
}

func (gc *GetSubscribedTagsCommand) Execute(ctx context.Context, update *objects.Update) {
	var msg strings.Builder
	tags, err := gc.Bot.GetUserRepo().GetTagsByUserId(ctx, strconv.Itoa(update.Message.Chat.Id))
	if err != nil {
		log.Printf("[getSubscribedTags]: failed getting tags by user id: %v", err.Error())
	}
	msg.WriteString(utils.AggregateTags(tags))

	if len(tags) == 0 || err != nil {
		msg.WriteString(NOT_SUBSCRIBED_MSG)
	}

	gc.Bot.SendMessage(msg.String(), update, false)
}

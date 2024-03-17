package bot

import (
	"context"
	"log"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/pills-of-cs/bot/types"
)

var _ types.ICommand = (*ScheduleNewsCommand)(nil)

type ScheduleNewsCommand struct {
	Bot *Bot
}

func NewScheduleNewsCommand(bot *Bot) types.Command {
	hc := ScheduleNewsCommand{
		Bot: bot,
	}
	return hc.Execute
}

func (sc *ScheduleNewsCommand) Execute(ctx context.Context, update *objects.Update) {
	msg, err := sc.Bot.setCron(ctx, update, "news")
	if err != nil {
		log.Printf("[ScheduleNews] got error: %v", err)
	}
	sc.Bot.SendMessage(msg.String(), update, true)
}

package bot

import (
	"context"
	"log"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/pills-of-cs/bot/types"
)

var _ types.ICommand = (*SchedulePillCommand)(nil)

type SchedulePillCommand struct {
	Bot *Bot
}

func NewSchedulePillCommand(bot *Bot) types.Command {
	hc := SchedulePillCommand{
		Bot: bot,
	}
	return hc.Execute
}

func (sc *SchedulePillCommand) Execute(ctx context.Context, update *objects.Update) {
	msg, err := sc.Bot.setCron(ctx, update, "pill")
	if err != nil {
		sc.Bot.SendMessage("failed validating the inserted time, try using the format `/schedule_pill HH:MM Timezone`", update, true)
		log.Printf("[SchedulePill] got error: %v", err)
	}
	sc.Bot.SendMessage(msg.String(), update, true)
}

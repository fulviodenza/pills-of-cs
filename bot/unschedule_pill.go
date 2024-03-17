package bot

import (
	"context"
	"log"
	"strconv"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/pills-of-cs/bot/types"
)

var _ types.ICommand = (*UnschedulePillCommand)(nil)

type UnschedulePillCommand struct {
	Bot *Bot
}

func NewUnschedulePillCommand(bot *Bot) types.Command {
	hc := UnscheduleNewsCommand{
		Bot: bot,
	}
	return hc.Execute
}

func (uc *UnschedulePillCommand) Execute(ctx context.Context, update *objects.Update) {
	id := strconv.Itoa(update.Message.Chat.Id)
	cronId, ok := uc.Bot.PillMap[id]
	if !ok {
		log.Printf("[UnschedulePill] id not found in pillMap: %v", id)
		uc.Bot.SendMessage("user not found in schedules", update, false)
	} else {
		uc.Bot.PillScheduler.Remove(cronId)
		uc.Bot.PillsMu.Lock()
		delete(uc.Bot.PillMap, id)
		uc.Bot.PillsMu.Unlock()
		err := uc.Bot.UserRepo.RemovePillSchedule(ctx, id)
		if err != nil {
			log.Printf("[UnschedulePill] error from db: %v", err)
			uc.Bot.SendMessage("user not found in db", update, false)
		} else {
			uc.Bot.SendMessage("pill unscheduled", update, false)
		}
	}
}

package bot

import (
	"context"
	"log"
	"strconv"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/pills-of-cs/bot/types"
)

var _ types.ICommand = (*UnscheduleNewsCommand)(nil)

type UnscheduleNewsCommand struct {
	Bot *Bot
}

func NewUnscheduleNewsCommand(bot *Bot) types.Command {
	hc := UnscheduleNewsCommand{
		Bot: bot,
	}
	return hc.Execute
}

func (uc *UnscheduleNewsCommand) Execute(ctx context.Context, update *objects.Update) {
	id := strconv.Itoa(update.Message.Chat.Id)
	cronId, ok := uc.Bot.NewsMap[id]
	if !ok {
		log.Printf("[UnscheduleNews] id not found in newsMap: %v", id)
		uc.Bot.SendMessage("user not found in schedules", update, false)
	} else {
		uc.Bot.NewsScheduler.Remove(cronId)
		uc.Bot.NewsMu.Lock()
		delete(uc.Bot.NewsMap, id)
		uc.Bot.NewsMu.Unlock()
		err := uc.Bot.UserRepo.RemoveNewsSchedule(ctx, id)
		if err != nil {
			log.Printf("[UnscheduleNews] error from db: %v", err)
			uc.Bot.SendMessage("user not found in db", update, false)
		}
		uc.Bot.SendMessage("news unscheduled", update, false)
	}
}

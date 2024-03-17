package bot

import (
	"context"
	"log"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/pills-of-cs/bot/types"
)

var _ types.ICommand = (*HelpCommand)(nil)

type HelpCommand struct {
	Bot types.IBot
}

func NewHelpCommand(bot *Bot) types.Command {
	hc := HelpCommand{
		Bot: bot,
	}
	return hc.Execute
}

func (hc *HelpCommand) Execute(ctx context.Context, update *objects.Update) {
	chatID := update.Message.Chat.Id
	helpMessage := hc.Bot.GetHelpMessage()

	_, err := hc.Bot.GetTelegramClient().SendMessage(chatID, helpMessage, "", 0, false, false)
	if err != nil {
		log.Printf("Failed to send help message: %v", err)
	}
}

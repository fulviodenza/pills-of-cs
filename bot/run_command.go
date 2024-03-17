package bot

import (
	"context"
	"log"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/pills-of-cs/bot/types"
)

var _ types.ICommand = (*RunCommand)(nil)

type RunCommand struct {
	Bot *Bot
}

func NewRunCommand(bot *Bot) types.Command {
	rc := HelpCommand{
		Bot: bot,
	}
	return rc.Execute
}

// Execute method to process the help command.
func (rc *RunCommand) Execute(ctx context.Context, update *objects.Update) {
	err := rc.Bot.SendMessage("Welcome to the pills-of-cs bot! Press `/pill` to request a pill or `/help` to get informations about the bot", update, true)
	if err != nil {
		log.Printf("Failed to send help message: %v", err)
	}
}

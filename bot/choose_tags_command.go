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

const (
	TAGS_UPDATED  = "tags updated"
	NO_VALID_TAGS = "no valid tags found"
)

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

	tags, err := cc.Bot.GetUserRepo().GetTagsByUserId(ctx, strconv.Itoa(update.Message.Chat.Id))
	if err != nil {
		log.Printf("[ChooseTags]: failed getting tag by user: %v", err.Error())
	}

	validatedArgs := make([]string, 0)
	for _, a := range args {
		if contains(a, cc.Bot.GetCategories()) && !contains(a, tags) {
			validatedArgs = append(validatedArgs, a)
		}
	}

	if len(validatedArgs) == 0 {
		cc.Bot.SendMessage(NO_VALID_TAGS, update, false)
		return
	}

	if err = cc.Bot.GetUserRepo().AddTagsToUser(ctx, strconv.Itoa(update.Message.Chat.Id), validatedArgs); err != nil {
		log.Printf("[ChooseTags]: failed adding tag to user: %v", err.Error())
		return
	}

	cc.Bot.SendMessage(TAGS_UPDATED, update, false)
}

func contains(s string, ss []string) bool {
	for _, t := range ss {
		if strings.EqualFold(s, t) {
			return true
		}
	}
	return false
}

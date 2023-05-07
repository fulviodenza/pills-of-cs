package bot

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/SakoDroid/telego/objects"
	"github.com/pills-of-cs/entities"
	"github.com/pills-of-cs/utils"
)

func (ba Bot) start(ctx context.Context, up *objects.Update) {
	_, err := ba.Bot.SendMessage(up.Message.Chat.Id, "Welcome to the pills-of-cs bot! Press `/pill` to request a pill or `/help` to get informations about the bot", "Markdown", up.Message.MessageId, false, false)
	if err != nil {
		return
	}
}

func (ba Bot) pill(ctx context.Context, up *objects.Update) {
	subscribedTags, err := ba.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil {
		log.Fatalf("[b.UserRepo.GetTagsByUserId]: failed getting tags: %v", err.Error())
		return
	}
	if subscribedTags == nil {
		_, err := ba.Bot.SendMessage(up.Message.Chat.Id, string(ba.HelpMessage), "Markdown", up.Message.MessageId, false, false)
		if err != nil {
			return
		}
	}

	var randomCategory, randomIndex int64
	var randomCategoryP []entities.Pill
	rand.Seed(time.Now().Unix())

	if len(subscribedTags) > 0 {

		randomCategory = utils.MakeTimestamp(len(subscribedTags))
		randomIndex = utils.MakeTimestamp(len(ba.Categories[subscribedTags[randomCategory]]))
		_, err = ba.Bot.SendMessage(
			up.Message.Chat.Id,
			ba.Categories[subscribedTags[randomCategory]][randomIndex].Title+": "+ba.Categories[subscribedTags[randomCategory]][randomIndex].Body, "Markdown", up.Message.MessageId, false, false)
		if err != nil {
			return
		}
	} else {
		randomCategoryP = utils.Pick(ba.Categories)
		randomIndex = utils.MakeTimestamp(len(randomCategoryP))
		_, err = ba.Bot.SendMessage(
			up.Message.Chat.Id,
			randomCategoryP[randomIndex].Title+": "+randomCategoryP[randomIndex].Body, "Markdown", up.Message.MessageId, false, false)
		if err != nil {
			return
		}
	}
}

func (ba Bot) help(ctx context.Context, up *objects.Update) {
	_, err := ba.Bot.SendMessage(up.Message.Chat.Id, string(ba.HelpMessage), "Markdown", up.Message.MessageId, false, false)
	if err != nil {
		return
	}
}

func (ba Bot) chooseTags(ctx context.Context, up *objects.Update) {
	// /cmd args[0] args[1]
	args := strings.SplitN(up.Message.Text, " ", -1)

	// Replacing the underscores with spaces in the arguments.
	for i, a := range args {
		if strings.Contains(a, "_") {
			twoWordArg := strings.SplitN(a, "_", 2)
			args[i] = twoWordArg[0] + " " + twoWordArg[1]
		}
	}

	err := ba.UserRepo.AddTagsToUser(ctx, strconv.Itoa(up.Message.Chat.Id), args[1:])
	if err != nil {
		return
	}

	log.Printf("Return operation exit")
	_, err = ba.Bot.SendMessage(up.Message.Chat.Id, "tags updated", "Markdown", up.Message.MessageId, false, false)
	if err != nil {
		return
	}
}

func (ba Bot) getTags(ctx context.Context, up *objects.Update) {
	msg := ""
	for k := range ba.Categories {
		msg += "- " + k + "\n"
	}
	_, err := ba.Bot.SendMessage(up.Message.Chat.Id, msg, "Markdown", up.Message.MessageId, false, false)
	if err != nil {
		return
	}
}

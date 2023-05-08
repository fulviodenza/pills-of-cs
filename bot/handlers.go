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

var _ IBot = &Bot{}

type IBot interface {
	Run(ctx context.Context, up *objects.Update)
	Pill(ctx context.Context, up *objects.Update)
	Help(ctx context.Context, up *objects.Update)
	ChooseTags(ctx context.Context, up *objects.Update)
	GetTags(ctx context.Context, up *objects.Update)
}

func (b Bot) Run(ctx context.Context, up *objects.Update) {
	_, err := b.Bot.SendMessage(up.Message.Chat.Id, "Welcome to the pills-of-cs bot! Press `/pill` to request a pill or `/help` to get informations about the bot", "Markdown", up.Message.MessageId, false, false)
	if err != nil {
		return
	}
}

func (b Bot) Pill(ctx context.Context, up *objects.Update) {

	subscribedTags, err := b.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil {
		log.Fatalf("[b.UserRepo.GetTagsByUserId]: failed getting tags: %v", err.Error())
		return
	}

	var randomCategory, randomIndex int64
	var randomCategoryP []entities.Pill
	rand.Seed(time.Now().Unix())

	if len(subscribedTags) > 0 {

		randomCategory = utils.MakeTimestamp(len(subscribedTags))
		randomIndex = utils.MakeTimestamp(len(b.Categories[subscribedTags[randomCategory]]))
		_, err = b.Bot.SendMessage(
			up.Message.Chat.Id,
			b.Categories[subscribedTags[randomCategory]][randomIndex].Title+": "+b.Categories[subscribedTags[randomCategory]][randomIndex].Body, "Markdown", up.Message.MessageId, false, false)
		if err != nil {
			log.Fatalf("[b.Bot.SendMessage]: failed sending message: %v", err.Error())
			return
		}
	} else {
		randomCategoryP = utils.Pick(b.Categories)
		randomIndex = utils.MakeTimestamp(len(randomCategoryP))
		_, err = b.Bot.SendMessage(
			up.Message.Chat.Id,
			randomCategoryP[randomIndex].Title+": "+randomCategoryP[randomIndex].Body, "Markdown", up.Message.MessageId, false, false)
		if err != nil {
			log.Fatalf("[b.Bot.SendMessage]: failed sending message: %v", err.Error())
			return
		}
	}
}

func (b Bot) Help(ctx context.Context, up *objects.Update) {
	_, err := b.Bot.SendMessage(up.Message.Chat.Id, string(b.HelpMessage), "Markdown", up.Message.MessageId, false, false)
	if err != nil {
		log.Fatalf("[b.Bot.SendMessage]: failed sending message: %v", err.Error())
		return
	}
}

func (b Bot) ChooseTags(ctx context.Context, up *objects.Update) {
	// /cmd args[0] args[1]
	args := strings.SplitN(up.Message.Text, " ", -1)

	// Replacing the underscores with spaces in the arguments.
	for i, a := range args {
		if strings.Contains(a, "_") {
			twoWordArg := strings.SplitN(a, "_", 2)
			args[i] = twoWordArg[0] + " " + twoWordArg[1]
		}
	}

	err := b.UserRepo.AddTagsToUser(ctx, strconv.Itoa(up.Message.Chat.Id), args[1:])
	if err != nil {
		return
	}

	log.Printf("Return operation exit")
	_, err = b.Bot.SendMessage(up.Message.Chat.Id, "tags updated", "Markdown", up.Message.MessageId, false, false)
	if err != nil {
		log.Fatalf("[b.UserRepo.AddTagsToUser]: failed adding tag to user: %v", err.Error())
		return
	}

	_, err = b.Bot.SendMessage(up.Message.Chat.Id, "tags updated", "Markdown", up.Message.MessageId, false, false)
	if err != nil {
		log.Fatalf("[b.Bot.SendMessage]: failed sending message: %v", err.Error())
		return
	}
}

func (b Bot) GetTags(ctx context.Context, up *objects.Update) {
	msg := ""
	for k := range b.Categories {
		msg += "- " + k + "\n"
	}
	_, err := b.Bot.SendMessage(up.Message.Chat.Id, msg, "Markdown", up.Message.MessageId, false, false)
	log.Fatalf("[b.Bot.SendMessage]: failed sending message: %v", err.Error())
	return
}

func (b Bot) GetSubscribedTags(ctx context.Context, up *objects.Update) {

	tags, err := b.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil {
		log.Fatalf("[b.UserRepo.GetTagsByUserId]: failed getting tags by user id: %v", err.Error())
		return
	}

	msg := aggregateTags(tags)

	_, err = b.Bot.SendMessage(up.Message.Chat.Id, msg, "Markdown", up.Message.MessageId, false, false)
	if err != nil {
		log.Fatalf("[b.Bot.SendMessage]: failed sending message: %v", err.Error())
		return
	}
}

func aggregateTags(tags []string) string {
	msg := ""
	for _, s := range tags {
		msg += "- " + s + "\n"
	}

	return msg
}

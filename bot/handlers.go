package bot

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/SakoDroid/telego/objects"
	"github.com/pills-of-cs/entities"
	"github.com/pills-of-cs/parser"
	"github.com/pills-of-cs/utils"
)

var _ IBot = (*Bot)(nil)

type IBot interface {
	Run(ctx context.Context, up *objects.Update)
	Pill(ctx context.Context, up *objects.Update)
	Help(ctx context.Context, up *objects.Update)
	ChooseTags(ctx context.Context, up *objects.Update)
	GetTags(ctx context.Context, up *objects.Update)
	SchedulePill(ctx context.Context, up *objects.Update)
}

func (b *Bot) Run(ctx context.Context, up *objects.Update) {
	_, err := b.Bot.SendMessage(up.Message.Chat.Id, "Welcome to the pills-of-cs bot! Press `/pill` to request a pill or `/help` to get informations about the bot", "", 0, false, false)
	if err != nil {
		log.Printf("[Run]: failed sending message: %v", err.Error())
	}
}

func (b *Bot) Pill(ctx context.Context, up *objects.Update) {
	msg, err := b.pill(ctx, up)
	if err != nil {
		log.Printf("[Pill]: failed building message: %v", err.Error())
	}
	_, err = b.Bot.SendMessage(up.Message.Chat.Id, msg, "", 0, false, false)
	if err != nil {
		log.Printf("[Pill]: failed sending message: %v", err.Error())
	}
}

func (b *Bot) pill(ctx context.Context, up *objects.Update) (msg string, err error) {
	subscribedTags, err := b.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil {
		log.Fatalf("[Pill]: failed getting tags: %v", err.Error())
		return "", err
	}

	var randomCategory, randomIndex int64
	var randomCategoryP []entities.Pill
	rand.Seed(time.Now().Unix())

	if len(subscribedTags) > 0 {
		randomCategory = utils.MakeTimestamp(len(subscribedTags))
		randomIndex = utils.MakeTimestamp(len(b.Categories[subscribedTags[randomCategory]]))
		msg = b.Categories[subscribedTags[randomCategory]][randomIndex].Title + ": " + b.Categories[subscribedTags[randomCategory]][randomIndex].Body
	} else {
		randomCategoryP = utils.Pick(b.Categories)
		randomIndex = utils.MakeTimestamp(len(randomCategoryP))
		msg = randomCategoryP[randomIndex].Title + ": " + randomCategoryP[randomIndex].Body
	}
	return msg, nil
}

func (b *Bot) Help(ctx context.Context, up *objects.Update) {
	_, err := b.Bot.SendMessage(up.Message.Chat.Id, string(b.HelpMessage), "Markdown", 0, false, false)
	if err != nil {
		log.Printf("[Help]: failed sending message: %v", err.Error())
	}
}

func (b *Bot) GetTags(ctx context.Context, up *objects.Update) {
	_, err := b.Bot.SendMessage(up.Message.Chat.Id, b.getTags(ctx, up), "Markdown", 0, false, false)
	if err != nil {
		log.Printf("[GetTags]: failed sending message: %v", err.Error())
	}
}

func (b *Bot) getTags(ctx context.Context, up *objects.Update) (msg string) {
	msg = ""
	for k := range b.Categories {
		msg += fmt.Sprintf("- %s\n", k)
	}
	return msg
}

func (b *Bot) GetSubscribedTags(ctx context.Context, up *objects.Update) {
	msg, err := b.getSubscribedTags(ctx, up)
	if err != nil {
		log.Fatalf("[GetSubscribedTags]: failed building message: %v", err.Error())
	}
	_, err = b.Bot.SendMessage(up.Message.Chat.Id, msg, "Markdown", 0, false, false)
	if err != nil {
		log.Fatalf("[GetSubscribedTags]: failed sending message: %v", err.Error())
	}
}

func (b *Bot) getSubscribedTags(ctx context.Context, up *objects.Update) (msg string, err error) {
	tags, err := b.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil {
		log.Fatalf("[getSubscribedTags]: failed getting tags by user id: %v", err.Error())
		return "", err
	}

	return utils.AggregateTags(tags), nil
}

func (b *Bot) ChooseTags(ctx context.Context, up *objects.Update) {
	msg, err := b.chooseTags(ctx, up)
	if err != nil {
		log.Printf("[ChooseTags]: failed building message: %v", err.Error())
	}
	_, err = b.Bot.SendMessage(up.Message.Chat.Id, msg, "Markdown", 0, false, false)
	if err != nil {
		log.Printf("[ChooseTags]: failed adding tag to user: %v", err.Error())
	}
}

func (b *Bot) chooseTags(ctx context.Context, up *objects.Update) (msg string, err error) {
	// /cmd args[0] args[1]
	args := strings.SplitN(up.Message.Text, " ", -1)

	// Replacing the underscores with spaces in the arguments.
	// This is done for more-than-one-word tags.
	// Indeed, /choose_tags command requires:
	///choose_tags distributed_systems for example
	for i, a := range args {
		if strings.Contains(a, "_") {
			twoWordArg := strings.SplitN(a, "_", 2)
			args[i] = twoWordArg[0] + " " + twoWordArg[1]
		}
	}

	err = b.UserRepo.AddTagsToUser(ctx, strconv.Itoa(up.Message.Chat.Id), args[1:])
	if err != nil {
		log.Fatalf("[ChooseTags]: failed adding tag to user: %v", err.Error())
		return "", err
	}

	log.Printf("[ChooseTags]: return operation exit")
	return "tags updated", nil
}

// /schedule_pill 08:00
func (b *Bot) SchedulePill(ctx context.Context, up *objects.Update) {
	msg, err := b.schedulePill(ctx, up)
	if err != nil {
		log.Printf("[SchedulePill]: failed building message: %v", err.Error())
	}

	_, err = b.Bot.SendMessage(up.Message.Chat.Id, msg, "Markdown", 0, false, false)
	if err != nil {
		log.Printf("[SchedulePill]: failed sending message: %v", err.Error())
	}
}

func (b *Bot) schedulePill(ctx context.Context, up *objects.Update) (msg string, err error) {
	id := strconv.Itoa(up.Message.Chat.Id)
	// args[1] contains the time HH:MM, args[2] contains the timezone
	args := strings.SplitN(up.Message.Text, " ", -1)

	if len(args) != 3 {
		msg = "Failed parsing provided time"
		return msg, nil
	}
	crontab, err := parser.ParseSchedule(args[1], args[2])
	if err != nil {
		msg = "Failed parsing provided time"
		return msg, err
	}

	err = b.UserRepo.SaveSchedule(ctx, id, crontab)
	if err != nil {
		log.Fatalf("[SchedulePill]: failed saving time: %v", err.Error())
		msg = "failed saving time"
		return msg, err
	}

	// run the goroutine with the cron
	go func(ctx context.Context, u *objects.Update) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("[SchedulePill]: Recovering from panic:", r)
			}
		}()
		_, err = b.Scheduler.AddFunc(crontab, func() {
			b.Pill(ctx, u)
		})
		if err != nil {
			log.Println("[SchedulePill]: got error:", err)
			return
		}
	}(ctx, up)

	// the human readable format is with times[0] (hours] first
	msg = fmt.Sprintf("Crontab for you pill `%s`", crontab)
	return msg, nil
}

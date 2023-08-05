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
	"github.com/pills-of-cs/utils"
)

var _ IBot = (*Bot)(nil)

type IBot interface {
	Run(ctx context.Context, up *objects.Update) error
	Pill(ctx context.Context, up *objects.Update) error
	Help(ctx context.Context, up *objects.Update) error
	ChooseTags(ctx context.Context, up *objects.Update) error
	GetTags(ctx context.Context, up *objects.Update) error
	SchedulePill(ctx context.Context, up *objects.Update) error
}

func (b *Bot) Run(ctx context.Context, up *objects.Update) error {
	_, err := b.Bot.SendMessage(up.Message.Chat.Id, "Welcome to the pills-of-cs bot! Press `/pill` to request a pill or `/help` to get informations about the bot", "Markdown", up.Message.MessageId, false, false)
	if err != nil {
		log.Fatalf("[Run]: failed sending message: %v", err.Error())
		return err
	}
	return nil
}

func (b *Bot) Pill(ctx context.Context, up *objects.Update) error {
	subscribedTags, err := b.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil {
		log.Fatalf("[Pill]: failed getting tags: %v", err.Error())
		return err
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
			log.Fatalf("[Pill]: failed sending message: %v", err.Error())
			return err
		}
	} else {
		randomCategoryP = utils.Pick(b.Categories)
		randomIndex = utils.MakeTimestamp(len(randomCategoryP))
		_, err = b.Bot.SendMessage(
			up.Message.Chat.Id,
			randomCategoryP[randomIndex].Title+": "+randomCategoryP[randomIndex].Body, "Markdown", up.Message.MessageId, false, false)
		if err != nil {
			log.Fatalf("[Pill]: failed sending message: %v", err.Error())
			return err
		}
	}
	return nil
}

func (b *Bot) Help(ctx context.Context, up *objects.Update) error {
	_, err := b.Bot.SendMessage(up.Message.Chat.Id, string(b.HelpMessage), "Markdown", up.Message.MessageId, false, false)
	if err != nil {
		log.Fatalf("[Help]: failed sending message: %v", err.Error())
		return err
	}
	return nil
}

func (b *Bot) GetTags(ctx context.Context, up *objects.Update) error {
	msg := ""
	for k := range b.Categories {
		msg += "- " + k + "\n"
	}
	_, err := b.Bot.SendMessage(up.Message.Chat.Id, msg, "Markdown", up.Message.MessageId, false, false)
	if err != nil {
		log.Fatalf("[GetTags]: failed sending message: %v", err.Error())
		return err
	}
	return nil
}

func (b *Bot) GetSubscribedTags(ctx context.Context, up *objects.Update) error {
	tags, err := b.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil {
		log.Fatalf("[GetSubscribedTags]: failed getting tags by user id: %v", err.Error())
		return err
	}

	msg := utils.AggregateTags(tags)

	_, err = b.Bot.SendMessage(up.Message.Chat.Id, msg, "Markdown", up.Message.MessageId, false, false)
	if err != nil {
		log.Fatalf("[GetSubscribedTags]: failed sending message: %v", err.Error())
		return err
	}
	return nil
}

func (b *Bot) ChooseTags(ctx context.Context, up *objects.Update) error {
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

	err := b.UserRepo.AddTagsToUser(ctx, strconv.Itoa(up.Message.Chat.Id), args[1:])
	if err != nil {
		log.Fatalf("[ChooseTags]: failed adding tag to user: %v", err.Error())
		return err
	}

	log.Printf("[ChooseTags]: return operation exit")
	_, err = b.Bot.SendMessage(up.Message.Chat.Id, "tags updated", "Markdown", up.Message.MessageId, false, false)
	if err != nil {
		log.Fatalf("[ChooseTags]: failed adding tag to user: %v", err.Error())
		return err
	}

	_, err = b.Bot.SendMessage(up.Message.Chat.Id, "tags updated", "Markdown", up.Message.MessageId, false, false)
	if err != nil {
		log.Fatalf("[ChooseTags]: failed sending message: %v", err.Error())
		return err
	}
	return nil
}

// /schedule_pill 08:00
func (b *Bot) SchedulePill(ctx context.Context, up *objects.Update) error {
	id := strconv.Itoa(up.Message.Chat.Id)

	// args[1] contains the time HH:MM
	args := strings.SplitN(up.Message.Text, " ", -1)
	sched := args[1]

	err := b.UserRepo.SaveSchedule(ctx, id, sched)
	if err != nil {
		log.Fatalf("[SchedulePill]: failed saving time: %v", err.Error())
		return err
	}

	message := fmt.Sprintf("I'll send you a pill every day at: %s", sched)
	_, err = b.Bot.SendMessage(up.Message.Chat.Id, message, "Markdown", up.Message.MessageId, false, false)
	if err != nil {
		log.Fatalf("[SchedulePill]: failed sending message: %v", err.Error())
		return err
	}

	// times contains an array with two elements [Hours, Minutes]
	times := strings.SplitN(sched, ":", -1)
	crontab := fmt.Sprintf("%s %s * * *", times[1], times[0])

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

	return nil
}

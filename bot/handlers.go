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
	"github.com/barthr/newsapi"
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
	News(ctx context.Context, up *objects.Update)
	ScheduleNews(ctx context.Context, up *objects.Update)
}

func (b *Bot) sendMessage(msg string, up *objects.Update, formatMarkdown bool) {
	parseMode := ""
	if formatMarkdown {
		parseMode = "Markdown"
	}
	_, err := b.Bot.SendMessage(up.Message.Chat.Id, msg, parseMode, 0, false, false)
	if err != nil {
		log.Printf("[SendMessage]: sending message to user: %v", err.Error())
	}
}

func (b *Bot) Run(ctx context.Context, up *objects.Update) {
	b.sendMessage("Welcome to the pills-of-cs bot! Press `/pill` to request a pill or `/help` to get informations about the bot", up, false)
}

func (b *Bot) Help(ctx context.Context, up *objects.Update) {
	b.sendMessage(string(b.HelpMessage), up, false)
}

func (b *Bot) Pill(ctx context.Context, up *objects.Update) {
	msg := ""
	subscribedTags, err := b.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil {
		log.Printf("[Pill]: failed getting tags: %v", err.Error())
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
	b.sendMessage(msg, up, false)
}

func (b *Bot) News(ctx context.Context, up *objects.Update) {
	msg := ""
	newsCategories := ""

	categories, err := b.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil || len(categories) == 0 {
		newsCategories += "technology"
	}

	sourceParams := &newsapi.EverythingParameters{}
	for _, c := range categories {
		newsCategories += c + "&"
	}
	sourceParams.Keywords = newsCategories
	sourceParams.Language = "en"

	sources, err := b.NewsClient.GetEverything(ctx, sourceParams)
	if err == nil && len(sources.Articles) != 0 {
		for i := 0; i < 10; i++ {
			msg += "- *" + sources.Articles[i].Title + "*\n"
			msg += sources.Articles[i].Description + "\n"
			msg += "from " + sources.Articles[i].URL + "\n"
		}
	} else {
		msg += "sources missing!"
	}
	b.sendMessage(msg, up, true)
}

func (b *Bot) GetTags(ctx context.Context, up *objects.Update) {
	var msg = ""
	for k := range b.Categories {
		msg += fmt.Sprintf("- %s\n", k)
	}
	if len(b.Categories) == 0 {
		msg = "empty categories"
	}
	b.sendMessage(msg, up, false)
}

func (b *Bot) GetSubscribedTags(ctx context.Context, up *objects.Update) {
	msg := ""
	tags, err := b.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil {
		log.Printf("[getSubscribedTags]: failed getting tags by user id: %v", err.Error())
	}
	msg += utils.AggregateTags(tags)
	if len(tags) == 0 {
		msg += "You are not subscribed to any tag!\nSubscribe one with /choose_tags [tag] command!"
	}

	b.sendMessage(msg, up, false)
}

func (b *Bot) ChooseTags(ctx context.Context, up *objects.Update) {
	// /cmd args[0] args[1]
	args := strings.SplitN(up.Message.Text, " ", -1)

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

	err := b.UserRepo.AddTagsToUser(ctx, strconv.Itoa(up.Message.Chat.Id), args[1:])
	if err != nil {
		log.Printf("[ChooseTags]: failed adding tag to user: %v", err.Error())
	}

	log.Printf("[ChooseTags]: return operation exit")
	b.sendMessage("tags updated", up, false)
}

func (b *Bot) SchedulePill(ctx context.Context, up *objects.Update) {
	msg, err := b.setCron(ctx, up, "pill")
	if err != nil {
		log.Printf("[SchedulePill] got error: %v", err)
	}
	b.sendMessage(msg, up, false)
}

func (b *Bot) ScheduleNews(ctx context.Context, up *objects.Update) {
	msg, err := b.setCron(ctx, up, "news")
	if err != nil {
		log.Printf("[ScheduleNews] got error: %v", err)
	}
	b.sendMessage(msg, up, true)
}

func (b *Bot) setCron(ctx context.Context, up *objects.Update, schedulerType string) (string, error) {
	var msg string
	id := strconv.Itoa(up.Message.Chat.Id)
	// args[1] contains the time HH:MM, args[2] contains the timezone
	args := strings.SplitN(up.Message.Text, " ", -1)

	if len(args) != 3 {
		msg = "Failed parsing provided time"
	}
	crontab, err := parser.ParseSchedule(args[1], args[2])
	if err != nil {
		msg = "Failed parsing provided time"
	}

	switch schedulerType {
	case "pill":
		err = b.UserRepo.SavePillSchedule(ctx, id, crontab)
		if err != nil {
			log.Printf("[SchedulePill]: failed saving time: %v", err.Error())
			msg = "failed saving time"
		}
	case "news":
		err = b.UserRepo.SaveNewsSchedule(ctx, id, crontab)
		if err != nil {
			log.Printf("[SchedulePill]: failed saving time: %v", err.Error())
			msg = "failed saving time"
		}

	}

	// run the goroutine with the cron
	go func(ctx context.Context, u *objects.Update) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("[SchedulePill]: Recovering from panic:", r)
			}
		}()
		switch schedulerType {
		case "pill":
			_, err = b.PillScheduler.AddFunc(crontab, func() {
				b.Pill(ctx, u)
			})
			if err != nil {
				log.Println("[SchedulePill]: got error:", err)
				return
			}
		case "news":
			_, err = b.NewsScheduler.AddFunc(crontab, func() {
				b.News(ctx, u)
			})
			if err != nil {
				log.Println("[ScheduleNews]: got error:", err)
				return
			}
		}
	}(ctx, up)

	// the human readable format is with times[0] (hours] first
	msg = fmt.Sprintf("Crontab for your pill `%s`", crontab)
	return msg, nil
}

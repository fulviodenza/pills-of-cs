package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/pills-of-cs/parser"
	"github.com/pills-of-cs/utils"

	"github.com/SakoDroid/telego/objects"
	"github.com/barthr/newsapi"
	"github.com/jomei/notionapi"
)

var _ IBot = (*Bot)(nil)

type notionDbRow struct {
	Tags notionapi.MultiSelectProperty `json:"Tags"`
	Text notionapi.RichTextProperty    `json:"Text"`
	Name notionapi.TitleProperty       `json:"Name"`
}

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

	if len(msg) >= 4096 {
		msgs := splitString(msg)
		for _, m := range msgs {
			_, err := b.TelegramClient.SendMessage(up.Message.Chat.Id, m, parseMode, 0, false, false)
			if err != nil {
				log.Printf("[SendMessage]: sending message to user: %v", err.Error())
			}
		}
	} else {
		_, err := b.TelegramClient.SendMessage(up.Message.Chat.Id, msg, parseMode, 0, false, false)
		if err != nil {
			log.Printf("[SendMessage]: sending message to user: %v", err.Error())
		}
	}
}

func splitString(s string) []string {
	if len(s) <= 0 {
		return nil
	}

	maxGroupLen := 4095
	if len(s) < maxGroupLen {
		maxGroupLen = len(s)
	}
	group := s[:maxGroupLen]
	return append([]string{group}, splitString(s[maxGroupLen:])...)
}

func (b *Bot) Run(ctx context.Context, up *objects.Update) {
	b.sendMessage("Welcome to the pills-of-cs bot! Press `/pill` to request a pill or `/help` to get informations about the bot", up, true)
}

func (b *Bot) Help(ctx context.Context, up *objects.Update) {
	b.sendMessage(string(b.HelpMessage), up, true)
}

func (b *Bot) Pill(ctx context.Context, up *objects.Update) {
	msg := ""
	subscribedTags, err := b.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil {
		log.Printf("[Pill]: failed getting tags: %v", err.Error())
	}

	choosenCategory := b.Categories[utils.MakeTimestamp(len(b.Categories))]
	if len(subscribedTags) > 0 {
		choosenCategory = b.Categories[utils.MakeTimestamp(len(subscribedTags))]
	}

	res, err := b.NotionClient.Database.Query(ctx, notionapi.DatabaseID(notionDatabaseId), &notionapi.DatabaseQueryRequest{
		Filter: notionapi.PropertyFilter{
			Property: "Tags",
			MultiSelect: &notionapi.MultiSelectFilterCondition{
				Contains: choosenCategory,
			},
		},
	})
	if err != nil {
		log.Printf("[Pill]: failed retrieving pill: %v", err.Error())
	}

	if res != nil {
		row := notionDbRow{}
		rowProps := make([]byte, 0)

		if rowProps, err = json.Marshal(res.Results[utils.MakeTimestamp(len(res.Results))].Properties); err != nil {
			log.Printf("[Pill]: failed marshaling pill: %v", err.Error())
		}
		if err = json.Unmarshal(rowProps, &row); err != nil {
			log.Printf("[Pill]: failed unmarshaling pill: %v", err.Error())
		}

		msg = row.Name.Title[0].Text.Content + ": "
		for _, c := range row.Text.RichText {
			msg += c.Text.Content
		}
		b.sendMessage(msg, up, true)
	}
}

func (b *Bot) News(ctx context.Context, up *objects.Update) {
	msg := ""
	newsCategories := ""

	categories, err := b.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil || len(categories) == 0 {
		newsCategories += "technology"
	}

	for _, c := range categories {
		sourceParams := &newsapi.EverythingParameters{
			Keywords: c + "&",
			Language: "en",
		}
		sources, err := b.NewsClient.GetEverything(ctx, sourceParams)
		if err == nil && len(sources.Articles) != 0 {
			articles := sources.Articles
			sort.Slice(articles, func(i, j int) bool {
				return sources.Articles[i].PublishedAt.Before(sources.Articles[i].PublishedAt)
			})

			for i := 0; i < 3; i++ {
				msg += "- *" + sources.Articles[i].Title + "*\n"
				msg += sources.Articles[i].Description + "\n"
				msg += "from " + sources.Articles[i].URL + "\n"
			}
		} else {
			log.Printf("err: %v", err)
			log.Printf("articles len: %v", len(sources.Articles))
			msg += "sources missing!"
		}
	}

	b.sendMessage(msg, up, true)
}

func (b *Bot) GetTags(ctx context.Context, up *objects.Update) {
	var msg = ""
	for _, v := range b.Categories {
		msg += fmt.Sprintf("- %s\n", v)
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
	b.sendMessage(msg, up, true)
}

func (b *Bot) ScheduleNews(ctx context.Context, up *objects.Update) {
	msg, err := b.setCron(ctx, up, "news")
	if err != nil {
		log.Printf("[ScheduleNews] got error: %v", err)
	}
	b.sendMessage(msg, up, true)
}

func (b *Bot) setCron(ctx context.Context, up *objects.Update, schedulerType string) (crontab string, err error) {
	var msg string
	id := strconv.Itoa(up.Message.Chat.Id)
	// args[1] contains the time HH:MM, args[2] contains the timezone
	args := strings.SplitN(up.Message.Text, " ", -1)
	if len(args) != 3 {
		msg = "Failed parsing provided time"
	} else {
		crontab, err = parser.ValidateSchedule(args[1], args[2])
		if err != nil {
			msg = "Failed parsing provided time"
		}
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

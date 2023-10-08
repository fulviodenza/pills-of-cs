package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pills-of-cs/parser"
	"github.com/pills-of-cs/utils"

	"github.com/SakoDroid/telego/objects"
	"github.com/barthr/newsapi"
	"github.com/google/go-cmp/cmp"
	"github.com/jomei/notionapi"
)

const QuizCategory = "quiz"

type notionDbRow struct {
	Tags notionapi.MultiSelectProperty `json:"Tags"`
	Text notionapi.RichTextProperty    `json:"Text"`
	Name notionapi.TitleProperty       `json:"Name"`
}

func (b *Bot) sendMessage(msg string, up *objects.Update, formatMarkdown bool) {
	parseMode := ""
	if formatMarkdown {
		parseMode = "Markdown"
	}

	if len(msg) >= 4096 {
		msgs := parser.SplitString(msg)
		for _, m := range msgs {
			_, err := b.TelegramClient.SendMessage(up.Message.Chat.Id, m, parseMode, 0, false, false)
			if err != nil {
				log.Printf("[SendMessage]: sending message to user: %v", err.Error())
			}
			break
		}
	} else {
		_, err := b.TelegramClient.SendMessage(up.Message.Chat.Id, msg, parseMode, 0, false, false)
		if err != nil {
			log.Printf("[SendMessage]: sending message to user: %v", err.Error())
		}
	}
}

func (b *Bot) Run(ctx context.Context, up *objects.Update) {
	b.sendMessage("Welcome to the pills-of-cs bot! Press `/pill` to request a pill or `/help` to get informations about the bot", up, true)
}

func (b *Bot) Help(ctx context.Context, up *objects.Update) {
	b.sendMessageFunc(string(b.HelpMessage), up, true)
}

func (b *Bot) Pill(ctx context.Context, up *objects.Update) {
	var msg strings.Builder
	subscribedTags, err := b.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil {
		log.Printf("[Pill]: failed getting tags: %v", err.Error())
	}

	chosenCategory := b.Categories[utils.MakeTimestamp(len(b.Categories))]
	if len(subscribedTags) > 0 {
		rand.Seed(time.Now().Unix())
		chosenCategory = b.Categories[utils.MakeTimestamp(len(subscribedTags))]
	}

	res, err := b.NotionClient.Database.Query(ctx, notionapi.DatabaseID(notionDatabaseId), &notionapi.DatabaseQueryRequest{
		Filter: notionapi.PropertyFilter{
			Property: "Tags",
			MultiSelect: &notionapi.MultiSelectFilterCondition{
				Contains: chosenCategory,
			},
		},
	})
	if err != nil {
		log.Printf("[Pill]: failed retrieving pill: %v", err.Error())
	}

	if res != nil {
		var row *notionDbRow = &notionDbRow{}
		rowProps := make([]byte, 0)

		resultsLen := len(res.Results)
		if resultsLen > 0 { // we will use resultsLen to be divided by 0
			if rowProps, err = json.Marshal(res.Results[utils.MakeTimestamp(resultsLen)].Properties); err != nil {
				log.Printf("[Pill]: failed marshaling pill: %v", err.Error())
			}
			if err = json.Unmarshal(rowProps, row); err != nil {
				log.Printf("[Pill]: failed unmarshaling pill: %v", err.Error())
			}

			if diff := cmp.Diff(*row, notionDbRow{}); diff != "" { // the row is not empty
				msg.WriteString(row.Name.Title[0].Text.Content + ": ")
				for _, c := range row.Text.RichText {
					msg.WriteString(c.Text.Content)
				}
			}
		}
		b.sendMessageFunc(msg.String(), up, false)
	}
}

func (b *Bot) News(ctx context.Context, up *objects.Update) {
	var msg strings.Builder
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
		sort.Slice(sources.Articles, func(i, j int) bool {
			return sources.Articles[i].PublishedAt.After(sources.Articles[j].PublishedAt)
		})
		if err == nil && len(sources.Articles) != 0 {
			articles := sources.Articles[:10]

			for _, a := range articles {
				if len(msg.String()) < 4095 {
					description := strings.Trim(a.Description, "\n")

					msg.WriteString("🔴 " + a.Title + "\n" + description + "\n" + "from " + a.URL + "\n")
				} else {
					break
				}
			}
		} else {
			log.Printf("err: %v; articles len: %v", err, len(sources.Articles))
			msg.WriteString("sources missing!")
		}
	}
	b.sendMessage(msg.String(), up, false)
}

func (b *Bot) GetTags(ctx context.Context, up *objects.Update) {
	var msg strings.Builder
	for _, v := range b.Categories {
		msg.WriteString(fmt.Sprintf("- %s\n", v))
	}
	if len(b.Categories) == 0 {
		msg.WriteString("empty categories")
	}
	b.sendMessageFunc(msg.String(), up, false)
}

func (b *Bot) GetSubscribedTags(ctx context.Context, up *objects.Update) {
	var msg strings.Builder
	tags, err := b.UserRepo.GetTagsByUserId(ctx, strconv.Itoa(up.Message.Chat.Id))
	if err != nil {
		log.Printf("[getSubscribedTags]: failed getting tags by user id: %v", err.Error())
	}
	msg.WriteString(utils.AggregateTags(tags))
	if len(tags) == 0 {
		msg.WriteString("You are not subscribed to any tag!\nSubscribe one with /choose_tags [tag] command!")
	}

	b.sendMessage(msg.String(), up, false)
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

	b.sendMessage("tags updated", up, false)
}

func (b *Bot) SchedulePill(ctx context.Context, up *objects.Update) {
	msg, err := b.setCron(ctx, up, "pill")
	if err != nil {
		b.sendMessage("failed validating the inserted time, try using the format `/schedule_pill HH:MM Timezone`", up, true)
		log.Printf("[SchedulePill] got error: %v", err)
	}
	b.sendMessage(msg.String(), up, true)
}

func (b *Bot) ScheduleNews(ctx context.Context, up *objects.Update) {
	msg, err := b.setCron(ctx, up, "news")
	if err != nil {
		log.Printf("[ScheduleNews] got error: %v", err)
	}
	b.sendMessage(msg.String(), up, true)
}

func (b *Bot) UnscheduleNews(ctx context.Context, up *objects.Update) {
	id := strconv.Itoa(up.Message.Chat.Id)
	cronId, ok := b.NewsMap[id]
	if !ok {
		log.Printf("[UnscheduleNews] id not found in newsMap: %v", id)
		b.sendMessage("user not found in schedules", up, false)
	} else {
		b.NewsScheduler.Remove(cronId)
		b.NewsMu.Lock()
		delete(b.NewsMap, id)
		b.NewsMu.Unlock()
		err := b.UserRepo.RemoveNewsSchedule(ctx, id)
		if err != nil {
			log.Printf("[UnscheduleNews] error from db: %v", err)
			b.sendMessage("user not found in db", up, false)
		}
		b.sendMessage("news unscheduled", up, false)
	}
}

func (b *Bot) UnschedulePill(ctx context.Context, up *objects.Update) {
	id := strconv.Itoa(up.Message.Chat.Id)
	cronId, ok := b.PillMap[id]
	if !ok {
		log.Printf("[UnschedulePill] id not found in pillMap: %v", id)
		b.sendMessage("user not found in schedules", up, false)
	} else {
		b.PillScheduler.Remove(cronId)
		b.PillsMu.Lock()
		delete(b.PillMap, id)
		b.PillsMu.Unlock()
		err := b.UserRepo.RemovePillSchedule(ctx, id)
		if err != nil {
			log.Printf("[UnschedulePill] error from db: %v", err)
			b.sendMessage("user not found in db", up, false)
		} else {
			b.sendMessage("pill unscheduled", up, false)
		}
	}
}

func (b *Bot) Quiz(ctx context.Context, up *objects.Update) {
	var optionsAnswer []string
	var options []string
	question, optionsRaw := "", ""
	correctIndex := -1

	rawPoll, err := b.NotionClient.Database.Query(ctx, notionapi.DatabaseID(notionDatabaseId), &notionapi.DatabaseQueryRequest{
		Filter: notionapi.PropertyFilter{
			Property: "Tags",
			MultiSelect: &notionapi.MultiSelectFilterCondition{
				Contains: QuizCategory,
			},
		},
	})
	if err != nil {
		log.Printf("[Pill]: failed retrieving pill: %v", err.Error())
	}
	if rawPoll != nil {
		row := notionDbRow{}
		rowProps := make([]byte, 0)

		if rowProps, err = json.Marshal(rawPoll.Results[utils.MakeTimestamp(len(rawPoll.Results))].Properties); err != nil {
			log.Printf("[Pill]: failed marshaling pill: %v", err.Error())
		}
		if err = json.Unmarshal(rowProps, &row); err != nil {
			log.Printf("[Pill]: failed unmarshaling pill: %v", err.Error())
		}
		question = row.Name.Title[0].Text.Content
		for _, c := range row.Text.RichText {
			optionsRaw += c.Text.Content
		}
		optionsAnswer = strings.Split(optionsRaw, ";")
		options = strings.Split(optionsAnswer[0], ",")
		for i, o := range options {
			if o == optionsAnswer[1] {
				correctIndex = i
				break
			}
		}
	}
	poll, err := b.TelegramClient.CreatePoll(up.Message.Chat.Id, question, QuizCategory)
	if err != nil {
		log.Printf("[Quiz] error creating poll: %v", err)
	}
	for _, o := range options {
		poll.AddOption(o)
	}
	poll.SetCorrectOption(correctIndex)
	poll.Send(false, false, up.Message.MessageId)
}

func (b *Bot) setCron(ctx context.Context, up *objects.Update, schedulerType string) (strings.Builder, error) {
	var (
		crontab string
		err     error
		msg     strings.Builder
	)
	id := strconv.Itoa(up.Message.Chat.Id)
	// args[1] contains the time HH:MM, args[2] contains the timezone
	args := strings.SplitN(up.Message.Text, " ", -1)
	if len(args) != 3 {
		msg.WriteString("Failed parsing provided time")
	} else {
		crontab, err = parser.ValidateSchedule(args[1], args[2])
		if err != nil {
			msg.WriteString("Failed parsing provided time")
		}
	}
	switch schedulerType {
	case "pill":
		err = b.UserRepo.SavePillSchedule(ctx, id, crontab)
		if err != nil {
			log.Printf("[SchedulePill]: failed saving time: %v", err.Error())
			msg.WriteString("failed saving time")
		}
	case "news":
		err = b.UserRepo.SaveNewsSchedule(ctx, id, crontab)
		if err != nil {
			log.Printf("[SchedulePill]: failed saving time: %v", err.Error())
			msg.WriteString("failed saving time")
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
			uid := strconv.Itoa(up.Message.Chat.Id)
			cronId, err := b.PillScheduler.AddFunc(crontab, func() {
				b.Pill(ctx, u)
			})
			if err != nil {
				log.Println("[SchedulePill]: got error:", err)
				return
			}
			b.PillsMu.Lock()
			b.PillMap[uid] = cronId
			b.PillsMu.Unlock()
		case "news":
			uid := strconv.Itoa(up.Message.Chat.Id)
			cronId, err := b.NewsScheduler.AddFunc(crontab, func() {
				b.News(ctx, u)
			})
			if err != nil {
				log.Println("[ScheduleNews]: got error:", err)
				return
			}
			b.NewsMu.Lock()
			b.NewsMap[uid] = cronId
			b.NewsMu.Unlock()
		}
	}(ctx, up)

	// the human readable format is with times[0] (hours) first
	msg.WriteString(fmt.Sprintf("Crontab for your pill `%s`", crontab))
	return msg, nil
}

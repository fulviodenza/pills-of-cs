package entities

import (
	"context"
	"time"

	repositories "github.com/pills-of-cs/adapters/repositories"

	bt "github.com/SakoDroid/telego"
	"github.com/barthr/newsapi"
	"github.com/jomei/notionapi"
	"github.com/robfig/cron/v3"
)

type BotConf struct {
	TelegramClient bt.Bot
	NotionClient   notionapi.Client
	NewsClient     *newsapi.Client

	HelpMessage string
	Categories  []string

	UserRepo  repositories.UserRepo
	Schedules map[string]time.Time

	PillScheduler *cron.Cron
	NewsScheduler *cron.Cron
}

type IBot interface {
	Start(ctx context.Context)
}

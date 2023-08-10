package entities

import (
	"context"
	"time"

	"github.com/barthr/newsapi"
	repositories "github.com/pills-of-cs/adapters/repositories"

	bt "github.com/SakoDroid/telego"
	"github.com/jomei/notionapi"
	"github.com/robfig/cron/v3"
)

type BotConf struct {
	Bot           bt.Bot
	NotionClient  notionapi.Client
	NewsClient    *newsapi.Client
	HelpMessage   string
	UserRepo      repositories.UserRepo
	Pills         []Pill
	Categories    map[string][]Pill
	Schedules     map[string]time.Time
	PillScheduler *cron.Cron
	NewsScheduler *cron.Cron
}

type IBot interface {
	Start(ctx context.Context)
}

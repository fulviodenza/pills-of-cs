package entities

import (
	"context"
	"time"

	repositories "github.com/pills-of-cs/adapters/repositories"

	bt "github.com/SakoDroid/telego"
	"github.com/jomei/notionapi"
	"github.com/robfig/cron/v3"
)

type BotConf struct {
	Bot          bt.Bot
	NotionClient notionapi.Client
	HelpMessage  string
	UserRepo     repositories.UserRepo
	Pills        []Pill
	Categories   map[string][]Pill
	Schedules    map[string]time.Time
	Scheduler    *cron.Cron
}

type IBot interface {
	Start(ctx context.Context) error
}

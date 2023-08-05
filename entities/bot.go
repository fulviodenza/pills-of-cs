package entities

import (
	"context"
	"time"

	bt "github.com/SakoDroid/telego"
	"github.com/go-co-op/gocron"
	"github.com/jomei/notionapi"
	repositories "github.com/pills-of-cs/adapters/repositories"
)

type BotConf struct {
	Bot          bt.Bot
	NotionClient notionapi.Client
	HelpMessage  string
	UserRepo     repositories.UserRepo
	Pills        []Pill
	Categories   map[string][]Pill
	Schedules    map[string]time.Time
	Scheduler    *gocron.Scheduler
}

type IBot interface {
	Start(ctx context.Context) error
}

package entities

import (
	"context"

	bt "github.com/SakoDroid/telego"
	cfg "github.com/SakoDroid/telego/configs"
	"github.com/SakoDroid/telego/objects"
	"github.com/jomei/notionapi"
	repositories "github.com/pills-of-cs/adapters/repositories"
)

type BotConf struct {
	TelegramToken string
	Cfg           cfg.BotConfigs
	Bot           bt.Bot
	NotionClient  notionapi.Client
	HelpMessage   string
	UserRepo      repositories.UserRepo
	Pills         []Pill
	Categories    map[string][]Pill
}

type IBot interface {
	Start(ctx context.Context) error
	HandleMessage(ctx context.Context, up *objects.Update)
}

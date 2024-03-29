package types

import (
	"context"

	bt "github.com/SakoDroid/telego/v2"
	objs "github.com/SakoDroid/telego/v2/objects"
	"github.com/jomei/notionapi"
	"github.com/pills-of-cs/adapters/news"
	adapters "github.com/pills-of-cs/adapters/repositories"
)

type NotionDbRow struct {
	Tags notionapi.MultiSelectProperty `json:"Tags"`
	Text notionapi.RichTextProperty    `json:"Text"`
	Name notionapi.TitleProperty       `json:"Name"`
}

type NotionInterface interface {
	DatabaseQuery(ctx context.Context, databaseID notionapi.DatabaseID, query *notionapi.DatabaseQueryRequest) (*notionapi.DatabaseQueryResponse, error)
}

type TelegramInterface interface {
	SendMessage(chatID int, msg string, parseMode string, replyToMessageID int, disableWebPagePreview, disableNotification bool) (*objs.Message, error)
}

type Command func(ctx context.Context, update *objs.Update)

type ICommand interface {
	Execute(ctx context.Context, update *objs.Update)
}

// Getters and Setters methods Bot instances
type IBot interface {
	Start(ctx context.Context)
	SendMessage(msg string, up *objs.Update, formatMarkdown bool) error
	GetHelpMessage() string
	GetCategories() []string
	SetCategories([]string)
	SetUserRepo(adapters.IUserRepo, chan interface{})
	GetUserRepo() adapters.IUserRepo
	SetTelegramClient(bot bt.Bot)
	GetTelegramClient() *bt.Bot
	SetNewsClient(news.INews)
	GetNewsClient() news.INews
}

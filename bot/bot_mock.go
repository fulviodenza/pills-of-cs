package bot

import (
	"context"

	bt "github.com/SakoDroid/telego/v2"
	"github.com/SakoDroid/telego/v2/objects"
	"github.com/pills-of-cs/adapters/news"
	repositories "github.com/pills-of-cs/adapters/repositories"
	"github.com/pills-of-cs/bot/types"
)

var _ types.IBot = (*MockBot)(nil)

type MockBot struct {
	repositories.IUserRepo
	news.INews
	Resp       string
	Err        error
	categories []string
}

func NewMockBot() *MockBot {
	return &MockBot{}
}

func (b *MockBot) SendMessage(msg string, up *objects.Update, formatMarkdown bool) error {
	if b.Err != nil {
		return b.Err
	}
	b.Resp = msg
	return nil
}

func (b *MockBot) Start(ctx context.Context) {}

func (b *MockBot) GetCategories() []string {
	return b.categories
}

func (b *MockBot) SetCategories(categories []string) {
	b.categories = categories
}

func (b *MockBot) GetHelpMessage() string {
	return ""
}

func (b *MockBot) SetUserRepo(repo repositories.IUserRepo, ch chan interface{}) {
	b.IUserRepo = repo
}

func (b *MockBot) GetUserRepo() repositories.IUserRepo {
	return b.IUserRepo
}

func (b *MockBot) SetTelegramClient(bot bt.Bot) {}

func (b *MockBot) GetTelegramClient() *bt.Bot {
	return nil
}

func (b *MockBot) SetNewsClient(news news.INews) {
	b.INews = news
}

func (b *MockBot) GetNewsClient() news.INews { return b.INews }

var (
	update = func(opts ...func(*objects.Update)) *objects.Update {
		update := &objects.Update{
			Message: &objects.Message{
				Chat: &objects.Chat{
					Id: 1,
				},
			},
		}
		for _, o := range opts {
			o(update)
		}

		return update
	}
	withMessage = func(msg string) func(*objects.Update) {
		return func(up *objects.Update) {
			up.Message.Text = msg
		}
	}
)

var (
	bot = func(opts ...func(*MockBot)) *MockBot {
		bot := NewMockBot()
		for _, o := range opts {
			o(bot)
		}
		return bot
	}
	withCategories = func(categories []string) func(*MockBot) {
		return func(mb *MockBot) {
			mb.SetCategories(categories)
		}
	}
)

package bot

import (
	"context"

	"github.com/SakoDroid/telego/v2/objects"
	repositories "github.com/pills-of-cs/adapters/repositories"
	"github.com/pills-of-cs/bot/types"
)

var _ types.IBot = (*MockBot)(nil)

type MockBot struct {
	repositories.IUserRepo
	Resp string
	Err  error
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
	return nil
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

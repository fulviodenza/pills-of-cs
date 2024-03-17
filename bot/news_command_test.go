package bot

import (
	"bytes"
	"context"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/barthr/newsapi"
	"github.com/google/go-cmp/cmp"
	"github.com/pills-of-cs/adapters/news"
	adapters "github.com/pills-of-cs/adapters/repositories"
)

func TestNewsCommand_Execute(t *testing.T) {
	articleMsg := "🔴 article_test_title\narticle_test_description\nfrom example.com\n"

	newsApiArticle := func(opts ...func(newsapi.Article)) newsapi.Article {
		a := newsapi.Article{
			Description: "article_test_description",
			Title:       "article_test_title",
			URL:         "example.com",
			PublishedAt: time.Unix(0, 0),
		}

		for _, o := range opts {
			o(a)
		}

		return a
	}
	withPublishedAt := func(t time.Time) func(newsapi.Article) {
		return func(a newsapi.Article) {
			a.PublishedAt = t
		}
	}
	type fields struct {
		Bot          *MockBot
		existingTags []string
		news         []newsapi.Article
		err          error
	}
	type args struct {
		update *objects.Update
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantMsg string
		wantLog string
	}{
		{
			"get news",
			fields{
				Bot:          bot(),
				existingTags: []string{"test"},
				news:         []newsapi.Article{newsApiArticle()},
			},
			args{
				update: update(),
			},
			articleMsg,
			"",
		},
		{
			"get news",
			fields{
				Bot:          bot(),
				existingTags: []string{"test"},
				news: []newsapi.Article{
					newsApiArticle(withPublishedAt(time.Unix(0, 0))),
					newsApiArticle(withPublishedAt(time.Unix(0, 1))),
				},
			},
			args{
				update: update(),
			},
			articleMsg + articleMsg,
			"",
		},

		{
			"get news",
			fields{
				Bot:          bot(),
				existingTags: []string{"test"},
			},
			args{
				update: update(),
			},
			SOURCE_MISSING,
			"",
		},
		{
			"get news",
			fields{
				Bot: bot(),
			},
			args{
				update: update(),
			},
			SOURCE_MISSING,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)

			cc := &NewsCommand{
				Bot: tt.fields.Bot,
			}

			ch := make(chan interface{}, 10)
			defer close(ch)

			userRepo := adapters.NewMockUserRepo(ch, tt.fields.err, tt.fields.existingTags)
			cc.Bot.SetUserRepo(userRepo, ch)

			cc.Bot.SetNewsClient(news.NewMockNewsClient(tt.fields.news))
			cc.Execute(context.TODO(), tt.args.update)

			mockBot := cc.Bot.(*MockBot)
			if diff := cmp.Diff(mockBot.Resp, tt.wantMsg); diff != "" {
				t.Errorf("unexpected message: want: \n%s\n, got \n%s", tt.wantMsg, mockBot.Resp)
			}

			defer log.SetOutput(os.Stderr)
			got := buf.String()
			if !strings.Contains(got, tt.wantLog) {
				t.Errorf("Execute() = %q, want %q", got, tt.wantLog)
			}
		})
	}
}

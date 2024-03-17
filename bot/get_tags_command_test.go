package bot

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/SakoDroid/telego/v2/objects"
	adapters "github.com/pills-of-cs/adapters/repositories"
)

func TestGetTagsCommand_Execute(t *testing.T) {
	type fields struct {
		Bot          *MockBot
		existingTags []string
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
			"get tags",
			fields{
				bot(
					withCategories([]string{"test"}),
				),
				nil,
				errors.New("error"),
			},
			args{
				update: update(),
			},
			"- test\n",
			"",
		},
		{
			"no categories found",
			fields{
				bot(),
				nil,
				errors.New("error"),
			},
			args{
				update: update(),
			},
			EMPTY_CATEGORIES,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)

			cc := &GetTagsCommand{
				Bot: tt.fields.Bot,
			}

			ch := make(chan interface{}, 10)
			defer close(ch)

			userRepo := adapters.NewMockUserRepo(ch, tt.fields.err, tt.fields.existingTags)
			cc.Bot.SetUserRepo(userRepo, ch)

			cc.Execute(context.TODO(), tt.args.update)

			mockBot := cc.Bot.(*MockBot)
			if mockBot.Resp != tt.wantMsg {
				t.Errorf("unexpected message: want: %s, got %s", tt.wantMsg, mockBot.Resp)
			}

			defer log.SetOutput(os.Stderr)
			got := buf.String()
			if !strings.Contains(got, tt.wantLog) {
				t.Errorf("Execute() = %q, want %q", got, tt.wantLog)
			}
		})
	}
}

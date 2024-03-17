package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/SakoDroid/telego/v2/objects"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	adapters "github.com/pills-of-cs/adapters/repositories"
)

func TestChooseTagsCommand_Execute(t *testing.T) {
	type fields struct {
		Bot *MockBot
		err error
	}
	type args struct {
		update *objects.Update
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantMsg       string
		wantSavedTags []string
		wantLog       string
	}{
		{
			"update tags",
			fields{
				bot(withCategories([]string{"test"})),
				nil,
			},
			args{
				update: update(
					withMessage("test"),
				),
			},
			"tags updated",
			[]string{"test"},
			"",
		},
		{
			"update tags with underscore",
			fields{
				bot(withCategories([]string{"test underscore"})),
				nil,
			},
			args{
				update: update(
					withMessage("test_underscore"),
				),
			},
			"tags updated",
			[]string{"test underscore"},
			"",
		},
		{
			"error updating tags",
			fields{
				bot(withCategories([]string{"test underscore"})),
				errors.New("error updating tags"),
			},
			args{
				update: update(
					withMessage("test_underscore"),
				),
			},
			"",
			nil,
			"failed adding tag to user: error updating tags",
		},
		{
			"fail tag validation",
			fields{
				bot(withCategories([]string{"test"})),
				errors.New("error updating tags"),
			},
			args{
				update: update(
					withMessage("test-non-existing"),
				),
			},
			NO_VALID_TAGS,
			nil,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)

			cc := &ChooseTagsCommand{
				Bot: tt.fields.Bot,
			}

			ch := make(chan interface{}, 10)
			defer close(ch)

			userRepo := adapters.NewMockUserRepo(ch, tt.fields.err, nil)
			cc.Bot.SetUserRepo(userRepo, ch)

			cc.Execute(context.TODO(), tt.args.update)

			defer log.SetOutput(os.Stderr)
			got := buf.String()
			if !strings.Contains(got, tt.wantLog) {
				t.Errorf("Execute() = %q, want %q", got, tt.wantLog)
			}

			mockBot := cc.Bot.(*MockBot)
			if mockBot.Resp != tt.wantMsg {
				t.Errorf("unexpected message: want: %s, got %s", tt.wantMsg, mockBot.Resp)
			}

			if tt.fields.err == nil {
				select {
				case res := <-ch:
					r := []string{}
					r = res.([]string)
					less := func(x, y interface{}) bool {
						b1, _ := json.Marshal(r)
						b2, _ := json.Marshal(tt.wantSavedTags)
						return string(b1) < string(b2)
					}

					if diff := cmp.Diff(r, tt.wantSavedTags, cmpopts.SortSlices(less)); diff != "" {
						t.Errorf("unexpected saved tags: want: %s, got %s", tt.wantSavedTags, res)
					}
				case <-time.After(5 * time.Second):
					t.Errorf("waited more than 5 seconds, exiting")
				}
			}
		})
	}
}

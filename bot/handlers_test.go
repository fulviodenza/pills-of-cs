package bot

import (
	"context"
	"testing"

	"github.com/SakoDroid/telego/objects"
	"github.com/jomei/notionapi"
	"github.com/pills-of-cs/mocks"
)

func TestBot_GetTags(t *testing.T) {
	ctx := context.Background()
	up := &objects.Update{}
	type fields struct {
		Categories []string
	}
	type args struct {
		ctx context.Context
		up  *objects.Update
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantMsg string
	}{
		{
			name: "get tags",
			fields: fields{
				Categories: []string{"Database"},
			},
			args: args{
				ctx,
				up,
			},
			wantMsg: "- Database\n",
		},
		{
			name: "empty tags",
			fields: fields{
				Categories: []string{},
			},
			args: args{
				ctx,
				up,
			},
			wantMsg: "empty categories",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Bot{
				Categories: tt.fields.Categories,
				sendMessageFunc: func(msg string, up *objects.Update, formatMarkdown bool) {
					if msg != tt.wantMsg {
						t.Errorf("Wrong message sent")
					}
				},
			}
			b.GetTags(tt.args.ctx, tt.args.up)
		})
	}
}

func TestBot_Help(t *testing.T) {
	ctx := context.Background()
	up := &objects.Update{}

	type fields struct {
		HelpMessage string
	}
	type args struct {
		ctx context.Context
		up  *objects.Update
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantMsg string
	}{
		{
			"print help message",
			fields{
				HelpMessage: "help message",
			},
			args{
				ctx: ctx,
				up:  up,
			},
			"help message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bot{
				HelpMessage: tt.fields.HelpMessage,
				sendMessageFunc: func(msg string, up *objects.Update, formatMarkdown bool) {
					if msg != tt.wantMsg {
						t.Errorf("Wrong message sent")
					}
				},
			}
			b.Help(tt.args.ctx, tt.args.up)
		})
	}
}

func TestBot_Pill(t *testing.T) {
	ctx := context.Background()
	up := &objects.Update{
		Message: &objects.Message{
			Chat: &objects.Chat{
				Id: 1,
			},
		},
	}

	queryVal := &notionapi.DatabaseQueryResponse{
		Results: []notionapi.Page{
			{
				Properties: notionapi.Properties{
					"Tags": notionapi.MultiSelectProperty{
						ID:   "tag",
						Type: notionapi.PropertyTypeText,
					},
					"Text": notionapi.RichTextProperty{
						ID:   "text",
						Type: notionapi.PropertyTypeText,
						RichText: []notionapi.RichText{
							{
								PlainText: "example text",
								Text: &notionapi.Text{
									Content: "example text",
								},
							},
						},
					},
					"Name": notionapi.TitleProperty{
						ID:   "text",
						Type: notionapi.PropertyTypeText,
						Title: []notionapi.RichText{
							{
								Text: &notionapi.Text{
									Content: "example title",
								},
							},
						},
					},
				},
			},
		},
	}

	type UserRepoMock struct {
		GetTagsByUserIdValue []string
		GetTagsByUserIdError error
	}

	type NotionClientMock struct {
		QueryVal *notionapi.DatabaseQueryResponse
	}
	type fields struct {
		Categories   []string
		UserRepo     UserRepoMock
		NotionClient NotionClientMock
	}
	type args struct {
		ctx context.Context
		up  *objects.Update
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantMsg string
	}{
		{
			"send pill",
			fields{
				Categories: []string{"Database"},
				UserRepo: UserRepoMock{
					GetTagsByUserIdValue: []string{"Database"},
					GetTagsByUserIdError: nil,
				},
				NotionClient: NotionClientMock{
					QueryVal: queryVal,
				},
			},
			args{
				ctx: ctx,
				up:  up,
			},
			"example title: example text",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bot{
				Categories: tt.fields.Categories,
				UserRepo: mocks.UserRepoMock{
					GetTagsByUserIdValue: tt.fields.UserRepo.GetTagsByUserIdValue,
					GetTagsByUserIdError: tt.fields.UserRepo.GetTagsByUserIdError,
				},
				NotionClient: notionapi.Client{
					Database: mocks.NotionDatabaseServiceMock{
						QueryVal: tt.fields.NotionClient.QueryVal,
					},
				},
				sendMessageFunc: func(msg string, up *objects.Update, formatMarkdown bool) {
					if msg != tt.wantMsg {
						t.Errorf("Wrong message sent: \n want: %v, \ngot: %v", tt.wantMsg, msg)
					}
				},
			}
			b.Pill(tt.args.ctx, tt.args.up)
		})
	}
}

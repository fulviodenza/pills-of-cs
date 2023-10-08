package entities

import "context"

type Document struct {
	UserId     int      `json:"userId" bson:"userId"`
	Categories []string `json:"categories" bson:"categories"`
}
type User interface {
	AddTagsToUser(ctx context.Context, id string, topics []string) error
	RemovePillSchedule(ctx context.Context, id string) error
	RemoveNewsSchedule(ctx context.Context, id string) error
	SavePillSchedule(ctx context.Context, id string, pillSchedule string) error
	GetTagsByUserId(ctx context.Context, id string) ([]string, error)
	GetAllPillCrontabs(ctx context.Context) (map[string]string, error)
	GetAllNewsCrontabs(ctx context.Context) (map[string]string, error)
	SaveNewsSchedule(ctx context.Context, id string, news_schedule string) error
}

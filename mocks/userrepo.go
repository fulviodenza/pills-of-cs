package mocks

import (
	"context"
	"github.com/pills-of-cs/entities"
)

type UserRepoMock struct {
	GetTagsByUserIdValue []string
	GetTagsByUserIdError error
}

var _ entities.User = (*UserRepoMock)(nil)

func (u UserRepoMock) AddTagsToUser(ctx context.Context, id string, topics []string) error {
	return nil
}

func (u UserRepoMock) RemovePillSchedule(ctx context.Context, id string) error {
	return nil
}

func (u UserRepoMock) RemoveNewsSchedule(ctx context.Context, id string) error {
	return nil
}

func (u UserRepoMock) SavePillSchedule(ctx context.Context, id string, pillSchedule string) error {
	return nil
}

func (u UserRepoMock) GetTagsByUserId(ctx context.Context, id string) ([]string, error) {
	return u.GetTagsByUserIdValue, u.GetTagsByUserIdError
}

func (u UserRepoMock) GetAllPillCrontabs(ctx context.Context) (map[string]string, error) {
	return nil, nil
}

func (u UserRepoMock) GetAllNewsCrontabs(ctx context.Context) (map[string]string, error) {
	return nil, nil
}

func (u UserRepoMock) SaveNewsSchedule(ctx context.Context, id string, news_schedule string) error {
	return nil
}

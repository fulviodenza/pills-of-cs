package adapters

import "context"

var _ IUserRepo = (*mockUserRepo)(nil)

type mockUserRepo struct {
	ch   chan interface{}
	err  error
	tags []string
}

func NewMockUserRepo(ch chan interface{}, err error, tags []string) IUserRepo {
	return &mockUserRepo{
		ch:   ch,
		err:  err,
		tags: tags,
	}
}
func (ur *mockUserRepo) AddTagsToUser(ctx context.Context, id string, topics []string) error {
	if ur.err != nil {
		return ur.err
	}

	ur.ch <- topics
	return nil
}

func (ur *mockUserRepo) RemoveTagsFromUser(ctx context.Context, id string, topics []string) ([]string, error) {
	return nil, nil
}

func (*mockUserRepo) RemovePillSchedule(ctx context.Context, id string) error { return nil }
func (*mockUserRepo) RemoveNewsSchedule(ctx context.Context, id string) error { return nil }
func (*mockUserRepo) SavePillSchedule(ctx context.Context, id string, pillSchedule string) error {
	return nil
}
func (ur *mockUserRepo) GetTagsByUserId(ctx context.Context, id string) ([]string, error) {
	if ur.err != nil {
		return nil, ur.err
	}
	return ur.tags, nil
}
func (*mockUserRepo) GetAllPillCrontabs(ctx context.Context) (map[string]string, error) {
	return nil, nil
}
func (*mockUserRepo) GetAllNewsCrontabs(ctx context.Context) (map[string]string, error) {
	return nil, nil
}
func (*mockUserRepo) SaveNewsSchedule(ctx context.Context, id string, news_schedule string) error {
	return nil
}

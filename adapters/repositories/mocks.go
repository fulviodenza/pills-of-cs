package adapters

import "context"

var _ IUserRepo = (*mockUserRepo)(nil)

type mockUserRepo struct {
	ch  chan interface{}
	err error
}

func NewMockUserRepo(ch chan interface{}, err error) IUserRepo {
	return &mockUserRepo{
		ch:  ch,
		err: err,
	}
}
func (ur *mockUserRepo) AddTagsToUser(ctx context.Context, id string, topics []string) error {
	if ur.err != nil {
		return ur.err
	}

	ur.ch <- topics
	return nil
}

func (*mockUserRepo) RemovePillSchedule(ctx context.Context, id string) error { return nil }
func (*mockUserRepo) RemoveNewsSchedule(ctx context.Context, id string) error { return nil }
func (*mockUserRepo) SavePillSchedule(ctx context.Context, id string, pillSchedule string) error {
	return nil
}
func (*mockUserRepo) GetTagsByUserId(ctx context.Context, id string) ([]string, error) {
	return nil, nil
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

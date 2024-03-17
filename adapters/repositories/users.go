package adapters

import (
	"context"
	"errors"
	"log"

	"github.com/pills-of-cs/adapters/ent"
	"github.com/pills-of-cs/adapters/ent/user"
)

var _ IUserRepo = (*UserRepo)(nil)

type IUserRepo interface {
	AddTagsToUser(ctx context.Context, id string, topics []string) error
	RemovePillSchedule(ctx context.Context, id string) error
	RemoveNewsSchedule(ctx context.Context, id string) error
	SavePillSchedule(ctx context.Context, id string, pillSchedule string) error
	GetTagsByUserId(ctx context.Context, id string) ([]string, error)
	GetAllPillCrontabs(ctx context.Context) (map[string]string, error)
	GetAllNewsCrontabs(ctx context.Context) (map[string]string, error)
	SaveNewsSchedule(ctx context.Context, id string, news_schedule string) error
}

type UserRepo struct {
	*ent.Client
}

func NewUserRepo(client *ent.Client) *UserRepo {
	return &UserRepo{Client: client}
}

func (ur *UserRepo) AddTagsToUser(ctx context.Context, id string, topics []string) error {
	userEl, err := ur.User.Query().
		Where(user.IDEQ(id)).
		First(ctx)
	var notFoundError *ent.NotFoundError
	if errors.As(err, &notFoundError) {
		err = ur.User.Create().
			SetID(id).
			SetCategories(topics).
			SetPillSchedule("").
			SetNewsSchedule("").
			Exec(ctx)
		if err != nil {
			log.Printf("[ur.User.Create]: error executing the query: %v", err)
			return err
		}
	}
	if err != nil {
		var notFoundError *ent.NotFoundError
		if !errors.As(err, &notFoundError) {
			log.Printf("[ur.User.Create]: error executing the query: %v", err)
			return err
		}
	}

	toAdd := findCategoriesToAdd(topics, userEl.Categories)
	userEl.Categories = append(userEl.Categories, toAdd...)

	if toAdd != nil {
		err = ur.User.Update().SetCategories(userEl.Categories).Where(user.IDEQ(userEl.ID)).Exec(ctx)

		if err != nil {
			log.Printf("[ur.User.Update]: error executing the query: %v", err)
			return err
		}
	}

	return nil
}

func (ur *UserRepo) RemovePillSchedule(ctx context.Context, id string) error {
	err := ur.User.Update().SetPillSchedule("").Where(user.IDEQ(id)).Exec(ctx)
	if err != nil {
		log.Printf("[RemovePillSchedule]: error executing the query: %v", err)
	}
	return err
}

func (ur *UserRepo) RemoveNewsSchedule(ctx context.Context, id string) error {
	err := ur.User.Update().SetNewsSchedule("").Where(user.IDEQ(id)).Exec(ctx)
	if err != nil {
		log.Printf("[RemoveNewsSchedule]: error executing the query: %v", err)
	}
	return err
}

func (ur *UserRepo) SavePillSchedule(ctx context.Context, id string, pillSchedule string) error {
	_, err := ur.User.Query().
		Where(user.IDEQ(id)).
		First(ctx)
	if _, ok := err.(*ent.NotFoundError); ok {
		err = ur.User.Create().
			SetID(id).
			SetCategories([]string{}).
			SetPillSchedule(pillSchedule).
			SetNewsSchedule("").
			Exec(ctx)
		if err != nil {
			log.Printf("[ur.User.Create]: error executing the query: %v", err)
			return err
		}
	}
	if err != nil {
		if _, ok := err.(*ent.NotFoundError); !ok {
			log.Printf("[ur.User.Create]: error executing the query: %v", err)
			return err
		}
	}

	err = ur.User.Update().SetPillSchedule(pillSchedule).Where(user.IDEQ(id)).Exec(ctx)
	if err != nil {
		log.Printf("[ur.User.Update]: error executing the query: %v", err)
		return err
	}

	return nil
}

func (ur *UserRepo) GetTagsByUserId(ctx context.Context, id string) ([]string, error) {
	exists, err := ur.User.Query().
		Where(user.IDEQ(id)).Exist(ctx)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	first, err := ur.Client.User.Query().
		Where(user.IDEQ(id)).First(ctx)
	if err != nil {
		return nil, err
	}

	return first.Categories, err
}

func (ur *UserRepo) GetAllPillCrontabs(ctx context.Context) (map[string]string, error) {
	users, err := ur.Client.User.Query().All(ctx)
	if err != nil {
		return nil, err
	}

	crontabs := make(map[string]string, len(users))
	for _, u := range users {
		crontabs[u.ID] = u.PillSchedule
	}
	return crontabs, nil
}

func (ur *UserRepo) GetAllNewsCrontabs(ctx context.Context) (map[string]string, error) {
	users, err := ur.Client.User.Query().All(ctx)
	if err != nil {
		return nil, err
	}

	crontabs := make(map[string]string, len(users))
	for _, u := range users {
		crontabs[u.ID] = u.NewsSchedule
	}
	return crontabs, nil
}

func (ur *UserRepo) SaveNewsSchedule(ctx context.Context, id string, news_schedule string) error {
	_, err := ur.User.Query().
		Where(user.IDEQ(id)).
		First(ctx)
	if _, ok := err.(*ent.NotFoundError); ok {
		err = ur.User.Create().
			SetID(id).
			SetCategories([]string{}).
			SetPillSchedule("").
			SetNewsSchedule(news_schedule).
			Exec(ctx)
		if err != nil {
			log.Printf("[ur.User.Create]: error executing the query: %v", err)
			return err
		}
	}
	if err != nil {
		if _, ok := err.(*ent.NotFoundError); !ok {
			log.Printf("[ur.User.Create]: error executing the query: %v", err)
			return err
		}
	}

	err = ur.User.Update().SetNewsSchedule(news_schedule).Where(user.IDEQ(id)).Exec(ctx)
	if err != nil {
		log.Printf("[ur.User.Update]: error executing the query: %v", err)
		return err
	}

	return nil
}

func findCategoriesToAdd(s1, s2 []string) []string {
	var toAdd []string

	for _, s := range s1 {
		found := false
		for _, r := range s2 {
			if s == r {
				found = true
				break
			}
		}
		if !found {
			toAdd = append(toAdd, s)
		}
	}

	return toAdd
}

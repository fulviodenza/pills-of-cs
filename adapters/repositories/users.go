package adapters

import (
	"context"

	"github.com/pills-of-cs/adapters/ent"
	"github.com/pills-of-cs/adapters/ent/user"
)

type UserRepo struct {
	Client *ent.Client
}

func (ur *UserRepo) AddTagsToUser(ctx context.Context, id string, topics []string) error {

	user, err := ur.Client.User.Query().
		Where(user.IDEQ(id)).
		First(ctx)
	if _, ok := err.(*ent.NotFoundError); ok {
		err = ur.Client.User.Create().
			SetID(id).
			SetCategories(topics).
			Exec(ctx)
		if err != nil {
			return err
		}
	}
	if err != nil {
		if _, ok := err.(*ent.NotFoundError); !ok {
			return err
		}
	}

	toAdd := findCategoriesToAdd(topics, user.Categories)
	err = ur.Client.User.Update().AppendCategories(toAdd).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (ur *UserRepo) GetTagsByUserId(ctx context.Context, id string) ([]string, error) {
	user, err := ur.Client.User.Query().
		Where(user.IDEQ(id)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	return user.Categories, err
}

func findCategoriesToAdd(s1, s2 []string) []string {
	toAdd := make([]string, len(s1))

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

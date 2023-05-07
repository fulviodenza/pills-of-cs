package adapters

import (
	"context"
	"log"

	"github.com/pills-of-cs/adapters/ent"
	"github.com/pills-of-cs/adapters/ent/user"
)

type UserRepo struct {
	*ent.Client
}

func (ur *UserRepo) AddTagsToUser(ctx context.Context, id string, topics []string) error {

	userEl, err := ur.User.Query().
		Where(user.IDEQ(id)).
		First(ctx)
	if _, ok := err.(*ent.NotFoundError); ok {

		err = ur.User.Create().
			SetID(id).
			SetCategories(topics).
			Exec(ctx)
		if err != nil {
			log.Fatalf("[ur.User.Create]: error executing the query: %v", err)
			return err
		}
	}
	if err != nil {
		if _, ok := err.(*ent.NotFoundError); !ok {
			log.Fatalf("[ur.User.Create]: error executing the query: %v", err)
			return err
		}
	}

	toAdd := findCategoriesToAdd(topics, userEl.Categories)
	userEl.Categories = append(userEl.Categories, toAdd...)

	if toAdd != nil {
		err = ur.User.Update().SetCategories(userEl.Categories).Where(user.IDEQ(userEl.ID)).Exec(ctx)

		if err != nil {
			log.Fatalf("[ur.User.Update]: error executing the query: %v", err)
			return err
		}
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

	user, err := ur.Client.User.Query().
		Where(user.IDEQ(id)).First(ctx)
	if err != nil {
		return nil, err
	}

	return user.Categories, err
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

package adapters

import (
	"context"
	"pills-of-cs/adapters/ent"
	"pills-of-cs/adapters/ent/user"
)

type UserRepo struct {
	Client *ent.Client
}

func (ur *UserRepo) AddTagsToUser(ctx context.Context, id, topic string) error {

	user, err := ur.Client.User.Query().
		Where(user.IDEQ(id)).
		First(ctx)
	if _, ok := err.(*ent.NotFoundError); ok {
		err = ur.Client.User.Create().
			SetID(id).
			SetCategories([]string{topic}).
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
	for _, c := range user.Categories {
		if c == topic {
			// the topic is already present, exit
			return nil
		}
	}

	user.Categories = append(user.Categories, topic)

	err = ur.Client.User.Update().SetCategories(user.Categories).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

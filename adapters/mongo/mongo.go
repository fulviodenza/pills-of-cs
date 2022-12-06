package adapters

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepo struct {
	Client *mongo.Client
}

func (c *UserRepo) AddTagsToUser(userId string, categories []string) error {

	// TODO: Add db and collection
	coll := c.Client.Database("myDB").Collection("users")
	for _, cat := range categories {
		doc := bson.M{"userId": userId, "categories": cat}
		_, err := coll.InsertOne(context.TODO(), doc)
		if err != nil {
			return errors.New(err.Error())
		}
	}
	return nil
}

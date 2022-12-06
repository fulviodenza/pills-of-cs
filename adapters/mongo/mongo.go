package adapters

import (
	"context"
	"errors"
	"log"
	"pills-of-cs/entities"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepo struct {
	Client *mongo.Client
}

// Adding tags to a user.
func (c *UserRepo) AddTagsToUser(userId int, tags []string) error {

	coll := c.Client.Database("pills").Collection("users")

	for _, cat := range tags {

		filter := bson.M{
			"userId": userId,
		}

		doc := bson.M{
			"$push": bson.M{
				"categories": cat,
			},
		}

		_, err := coll.UpdateOne(context.TODO(), filter, doc, options.Update().SetUpsert(true))
		if err != nil {
			log.Fatalf("[AddTagsToUser]: %v", err)
			return errors.New(err.Error())
		}
	}
	return nil
}

func (c *UserRepo) GetTagsByUserId(userId int) (*entities.Document, error) {

	coll := c.Client.Database("pills").Collection("users")
	filter := bson.M{"userId": userId}
	var result bson.D

	err := coll.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if strings.Contains(err.Error(), "no documents") {
			return nil, nil
		}
		log.Fatalf("[GetTagsByUserId]: %v", err)
		return nil, errors.New(err.Error())
	}

	res, err := bson.Marshal(result)
	if err != nil {
		log.Fatalf("[GetTagsByUserId]: %v", err)
		return nil, errors.New(err.Error())
	}

	var document entities.Document
	err = bson.Unmarshal(res, &document)

	return &document, err
}

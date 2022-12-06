package entities

type Document struct {
	UserId     int      `json:"userId" bson:"userId"`
	Categories []string `json:"categories" bson:"categories"`
}
type User interface {
	AddTagsToUser(userId string, categories []string) error
}

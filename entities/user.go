package entities

type User interface {
	AddTagsToUser(userId string, categories []string) error
}

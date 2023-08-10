package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("user_id"),
		field.JSON("categories", []string{}).Optional(),
		field.String("pill_schedule").StorageKey("pill_schedule"),
		field.String("news_schedule").StorageKey("news_schedule"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}

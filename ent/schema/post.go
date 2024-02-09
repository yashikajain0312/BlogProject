package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "time"
)

// Post holds the schema definition for the Post entity.
type Post struct {
    ent.Schema
}

// Fields of the Post.
func (Post) Fields() []ent.Field {
    return []ent.Field{
        field.Int("id").Unique(),
        field.String("title"),
        field.String("content"),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

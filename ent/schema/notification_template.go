package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NotificationTemplate holds the schema definition for the NotificationTemplate entity.
type NotificationTemplate struct {
	ent.Schema
}

// Fields of the NotificationTemplate.
func (NotificationTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").Unique().Immutable(),
		field.String("name").NotEmpty(),
		field.String("type").NotEmpty(), // email, sms, webhook
		field.String("subject_template").Optional(),
		field.Text("content_template").NotEmpty(),
		field.String("format").Default("html"), // html, text, json
		field.Int("user_id").Optional(), // null for system default templates
		field.Bool("is_system_default").Default(false),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the NotificationTemplate.
func (NotificationTemplate) Edges() []ent.Edge {
	return nil
}

// Indexes of the NotificationTemplate.
func (NotificationTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type", "is_system_default").Unique(),
		index.Fields("user_id", "type"),
		index.Fields("is_active"),
		index.Fields("name", "user_id").Unique(),
	}
}
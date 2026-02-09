package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Role holds the schema definition for the Role entity.
type Role struct {
	ent.Schema
}

// Fields of the Role.
func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Unique().
			Comment("角色名称"),
		field.String("description").
			Optional().
			Comment("角色描述"),
		field.Bool("is_system").
			Default(false).
			Comment("是否为系统角色"),
	}
}

// Edges of the Role.
func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("users", User.Type).
			Comment("角色用户"),
		edge.To("permissions", Permission.Type).
			Comment("角色权限"),
	}
}

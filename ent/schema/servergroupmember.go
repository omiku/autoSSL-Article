package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ServerGroupMember 服务器组成员中间表
type ServerGroupMember struct {
	ent.Schema
}

// Indexes 定义ServerGroupMember模型的索引
func (ServerGroupMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("server_group", "access").Unique(),
	}
}

func (ServerGroupMember) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("type").Values("ssh", "api", "other").Comment("访问类型"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (ServerGroupMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("server_group", ServerGroup.Type).Unique().Required().Comment("关联的服务器组"),
		edge.From("access", Access.Type).Ref("server_group_members").Unique().Comment("关联的访问配置"),
	}
}

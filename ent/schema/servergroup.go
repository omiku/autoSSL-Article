package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ServerGroup 服务器组模型
type ServerGroup struct {
	ent.Schema
}

func (ServerGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Comment("组名称"),
		field.String("description").Optional().Comment("组描述"),
		field.Enum("type").Values("ssh", "api", "tencentcloud-teo").Comment("组访问类型"),
		field.Enum("status").Values("active", "inactive").Default("active").Comment("组状态"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (ServerGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("server_groups").Unique(),
		edge.To("group_memberships", ServerGroupMember.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)).
			StorageKey(edge.Column("server_group_member_server_group")),
		edge.To("certificate_deployments", CertificateDeployment.Type).Comment("关联的证书部署记录"),
	}
}

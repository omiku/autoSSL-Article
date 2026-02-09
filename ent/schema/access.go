package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Access 访问权限配置模型
type Access struct {
	ent.Schema
}

// Fields 定义Access模型的字段
func (Access) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Comment("权限配置名称"),
		field.JSON("config", map[string]interface{}{}).Sensitive().Comment("访问配置详情"),
		field.Enum("type").Values("ssh", "api", "tencentcloud-teo", "other").Comment("访问类型"),
		field.Bool("is_active").Default(true).Comment("是否启用"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges 定义Access模型的关联关系
func (Access) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("access_configs").Unique().Required().Comment("所属用户"),
		edge.To("server_group_members", ServerGroupMember.Type).Comment("关联的服务器组成员"),
		edge.To("deployments", CertificateDeployment.Type).Comment("部署记录"),
	}
}

// Indexes 定义Access模型的索引
func (Access) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Edges("user").Unique(),
	}
}

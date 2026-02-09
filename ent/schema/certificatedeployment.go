package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CertificateDeployment 证书部署记录模型
type CertificateDeployment struct {
	ent.Schema
}

func (CertificateDeployment) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("status").Values("pending", "success", "failed").Default("pending").Comment("部署状态"),
		field.Time("deployed_at").Optional().Comment("部署时间"),
		field.String("log").Optional().Comment("部署日志"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Indexes 定义CertificateDeployment模型的索引
func (CertificateDeployment) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("certificate", "server_group").Unique(),
	}
}

func (CertificateDeployment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("certificate", Certificate.Type).Ref("deployments").Unique().Required(),
		edge.From("access", Access.Type).Ref("deployments").Required().Comment("访问配置"),
		edge.From("server_group", ServerGroup.Type).Ref("certificate_deployments").Unique().Required().Comment("关联的服务器组"),
	}
}

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Certificate SSL证书模型
type Certificate struct {
	ent.Schema
}

func (Certificate) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("domains", []string{}).Comment("证书域名列表"),
		field.String("issued_by").Comment("签发机构"),
		field.Time("issued_at").Optional().Comment("签发时间"),
		field.Time("expired_at").Optional().Comment("过期时间"),
		field.String("acme_account_email").Comment("ACME账号邮箱"),
		field.String("dns_provider_name").Comment("DNS提供商名称"),
		field.Int("renewal_days_before").Default(1).Comment("距离过期前多少天开始续期"),
		field.Bool("renewal_enabled").Default(false).Comment("是否启用自动续期"),
		field.Enum("validation_type").Values("dns", "http").Comment("验证类型"),
		field.Text("certificate").Sensitive().Comment("证书数据"),
		field.Text("private_key").Sensitive().Comment("私钥数据"),
		field.Enum("status").Values("pending", "processing", "issued", "deploying", "completed", "failed").Default("pending"),
		field.Text("error_message").Optional().Comment("错误信息"),
		field.Int("server_group_id").Optional().Comment("服务器组ID"),
		field.JSON("access_ids", []int{}).Optional().Comment("访问配置ID列表"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Certificate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("certificates").Unique().Required(),
		edge.From("dns_provider", DNSProvider.Type).Ref("certificates").Unique().Required(),
		edge.From("acme_account", ACMEAccount.Type).Ref("certificates").Unique(),
		edge.To("deployments", CertificateDeployment.Type),
		edge.To("workflow_steps", WorkflowStep.Type),
		edge.To("workflows", Workflow.Type).Comment("关联的工作流"),
	}
}

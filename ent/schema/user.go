package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User 用户模型
type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("is_admin").Default(false).Comment("管理员标识"),
		field.String("email").SchemaType(map[string]string{"mysql": "varchar(50)"}).Unique().Comment("登录邮箱"),
		field.String("password_hash").Sensitive().Comment("密码哈希"),

		field.Bool("is_active").Default(true).Comment("账户状态"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("orders", Order.Type).Comment("用户订单"),
		edge.To("certificates", Certificate.Type).Comment("用户证书"),
		edge.To("dns_providers", DNSProvider.Type).Comment("DNS提供商配置"),
		edge.To("server_groups", ServerGroup.Type).Comment("服务器组"),
		edge.To("access_configs", Access.Type).Comment("访问配置"),
		edge.To("notification_configs", NotificationConfig.Type).Comment("通知配置"),
		edge.To("workflow_steps", WorkflowStep.Type).Comment("工作流步骤"),
		edge.To("workflows", Workflow.Type).Comment("用户工作流"),
		edge.To("acme_accounts", ACMEAccount.Type).Comment("用户的ACME账户"),
		edge.From("roles", Role.Type).
			Ref("users").
			Comment("用户角色"),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email").Unique(),
	}
}

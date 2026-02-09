package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NotificationConfig 通知配置模型
type NotificationConfig struct {
	ent.Schema
}

// Fields 定义通知配置字段
func (NotificationConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Comment("配置名称"),
		field.Enum("type").
			Values("webhook", "email", "sms").
			Comment("通知类型"),
		field.JSON("config", map[string]interface{}{}).
			Comment(`通知配置参数(JSON格式)
				- webhook: {url:string, secret:string}
				- email: {smtp_server:string, smtp_port:int, username:string, password:string, from:string}
				- sms: {api_key:string, template_id:string, sign_name:string}`),
		field.Bool("is_active").
			Default(true).
			Comment("是否启用"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
	}
}

// Edges 定义关联关系
func (NotificationConfig) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("notification_configs").
			Unique().
			Required().
			Comment("所属用户"),
	}
}

// Indexes 定义索引
func (NotificationConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").
			Edges("user").
			Unique(),
		index.Fields("type").
			Edges("user"),
	}
}

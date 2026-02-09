package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// GlobalNotificationConfig 全局通知配置模型
type GlobalNotificationConfig struct {
	ent.Schema
}

// Fields 定义全局通知配置字段
func (GlobalNotificationConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Unique().
			Comment("配置名称，如：platform_email, platform_sms"),
		field.Enum("type").
			Values("email", "sms").
			Comment("通知类型：email或sms"),
		field.JSON("config", map[string]interface{}{}).
			Comment(`全局通知配置参数(JSON格式)
				- email: {smtp_server:string, smtp_port:int, username:string, password:string, from:string, from_name:string}
				- sms: {api_key:string, api_secret:string, sign_name:string, template_code:string}`),
		field.Bool("is_active").
			Default(true).
			Comment("是否启用"),
		field.String("description").
			Optional().
			Comment("配置描述"),
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

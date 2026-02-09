package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// DNSProvider DNS供应商模型
type DNSProvider struct {
	ent.Schema
}

func (DNSProvider) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("provider_type").Values("cloudflare", "alidns", "tencentcloud").Comment("DNS服务商类型"),
		field.JSON("config", map[string]interface{}{}).Comment("供应商配置 (JSON格式，包含所有认证信息)"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (DNSProvider) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("dns_providers").Unique().Required(),
		edge.To("certificates", Certificate.Type).Comment("关联的证书"),
	}
}

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ACMEAccount holds the schema definition for the ACMEAccount entity.
type ACMEAccount struct {
	ent.Schema
}

// Fields of the ACMEAccount.
func (ACMEAccount) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").
			Comment("ACME账号邮箱地址"),
		field.String("provider").
			Default("letsencrypt").
			Comment("SSL证书提供商: letsencrypt, zerossl"),
		field.String("private_key").
			Comment("ACME账号私钥，PEM格式"),
		field.String("registration_uri").
			Comment("ACME注册URI"),
		field.String("registration_body").
			Comment("ACME注册响应体，JSON格式"),
		field.Time("created_at").
			Default(time.Now).
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
	}
}

// Indexes of the ACMEAccount.
func (ACMEAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email", "provider").Unique(),
	}
}

// Edges of the ACMEAccount.
func (ACMEAccount) Edges() []ent.Edge {
	return []ent.Edge{
		// 每个ACME账户属于一个用户
		edge.From("user", User.Type).Ref("acme_accounts").Unique().Required(),
		// ACME账号可以被多个证书使用
		edge.To("certificates", Certificate.Type),
	}
}

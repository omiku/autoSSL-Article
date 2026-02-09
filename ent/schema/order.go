package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Order 证书订单模型
type Order struct {
	ent.Schema
}

func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.Float("price").Comment("订单金额"),
		field.Enum("order_type").Values("new", "renew", "upgrade").Comment("订单类型"),
		field.Int("cert_num").Comment("证书数量"),
		field.Int("deploy_num").Comment("部署服务器数量"),
		field.Time("start_date").Comment("订单生效时间"),
		field.Time("end_date").Comment("订单失效时间"),
		field.Enum("status").Values("pending", "active", "expired").Default("pending"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Order) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("orders").Unique().Required(),
	}
}

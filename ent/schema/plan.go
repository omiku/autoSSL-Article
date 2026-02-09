package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Plan 套餐模型
type Plan struct {
	ent.Schema
}

func (Plan) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Comment("套餐名称"),
		field.Float("price").Comment("套餐价格"),
		field.Time("valid_from").Comment("有效期起始时间"),
		field.Time("valid_to").Comment("有效期结束时间"),
		field.Int("cert_num").Comment("证书数量限制"),
		field.Int("deploy_num").Comment("部署服务器数量限制"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Plan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("users", User.Type).Comment("使用该套餐的用户"),
	}
}

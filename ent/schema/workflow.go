package schema

import (
	"fmt"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Workflow 工作流模型
type Workflow struct {
	ent.Schema
}

// Fields of the Workflow.
func (Workflow) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("domains", []string{}).Comment("域名列表"),
		field.String("status").Comment("工作流状态: pending, processing, issued, deploying, completed, failed, cancelled").
			Validate(func(s string) error {
				allowed := map[string]bool{
					"pending":    true,
					"processing": true,
					"issued":     true,
					"deploying":  true,
					"completed":  true,
					"failed":     true,
					"cancelled":  true,
				}
				if !allowed[s] {
					return fmt.Errorf("invalid status: %s", s)
				}
				return nil
			}),
		field.JSON("extra_config", map[string]interface{}{}).Sensitive().Comment("工作流额外配置"),
		field.JSON("notification_config", map[string]interface{}{}).Sensitive().Comment("工作流通知配置"),
		field.Int("current_step").Default(0).Comment("当前工作流步骤"),
		field.Int("total_steps").Default(0).Comment("总工作流步骤数"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("completed_at").
			Optional().
			Comment("完成时间"),
	}
}

// Edges of the Workflow.
func (Workflow) Edges() []ent.Edge {
	return []ent.Edge{
		// 关联到证书
		edge.From("certificate", Certificate.Type).
			Ref("workflows").
			Unique().
			Required(),
		// 关联到用户
		edge.From("user", User.Type).
			Ref("workflows").
			Unique().
			Required(),
		// 包含多个工作流步骤
		edge.To("steps", WorkflowStep.Type),
	}
}

// Indexes of the Workflow.
func (Workflow) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("created_at"),
		index.Edges("certificate").Unique(),
		index.Fields("status", "created_at"),
	}
}

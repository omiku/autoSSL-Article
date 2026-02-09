package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// WorkflowStep 工作流步骤模型
type WorkflowStep struct {
	ent.Schema
}

func (WorkflowStep) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Optional().Comment("步骤名称"),
		field.Enum("status").Values("pending", "processing", "issued", "deploying", "completed", "failed", "cancelled").Default("pending").Comment("步骤状态"),
		field.Int("step_number").Comment("步骤编号"),
		field.Text("log").Optional().Comment("步骤日志"),
		field.Time("started_at").Optional().Comment("开始时间"),
		field.Time("completed_at").Optional().Comment("完成时间"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (WorkflowStep) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("certificate", Certificate.Type).Ref("workflow_steps").Unique().Required(),
		edge.From("user", User.Type).Ref("workflow_steps").Required().Comment("所属用户"),
		edge.From("workflow", Workflow.Type).Ref("steps").Unique().Required().Comment("所属工作流"),
	}
}

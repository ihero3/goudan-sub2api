package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
)

// VideoTask holds the schema definition for the VideoTask entity.
// 视频生成异步任务表：记录用户请求、上游 channel、任务状态及最终结果。
// 与 Channel/Account 通过 account_id 弱关联（不做 ent edge，降低耦合）。
type VideoTask struct {
	ent.Schema
}

func (VideoTask) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "video_tasks"},
	}
}

func (VideoTask) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (VideoTask) Fields() []ent.Field {
	return []ent.Field{
		field.String("local_id").
			MaxLen(128).
			Unique().
			NotEmpty().
			Comment("对用户暴露的唯一任务 ID，如 vid_xxx"),
		field.Int64("user_id").
			NonNegative(),
		field.Int64("api_key_id").
			Optional().
			NonNegative(),
		field.String("public_model").
			MaxLen(100).
			NotEmpty().
			Comment("用户请求的统一模型名，如 seedance-2.5"),
		field.String("upstream_model").
			MaxLen(100).
			NotEmpty().
			Comment("model_mapping 翻译后上游实际模型名"),
		field.Int64("account_id").
			NonNegative().
			Comment("实际使用的上游 channel (Account) ID"),
		field.String("upstream_task_id").
			MaxLen(255).
			Optional().
			Comment("上游返回的任务 ID"),
		field.String("status").
			MaxLen(20).
			Default("processing").
			Comment("processing / succeeded / failed / cancelled"),
		field.String("resolution").
			MaxLen(20).
			Optional(),
		field.Int("duration_sec").
			Optional().
			NonNegative(),
		field.Text("video_url").
			Optional(),
		field.Text("thumbnail_url").
			Optional(),
		field.JSON("request_body", map[string]any{}).
			Optional().
			Comment("原始请求快照，便于排障"),
		field.Text("error_message").
			Optional(),
		field.Float("cost_usd").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Time("finished_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (VideoTask) Edges() []ent.Edge {
	return nil
}

func (VideoTask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status", "created_at"),
		index.Fields("user_id", "created_at"),
		index.Fields("account_id"),
	}
}

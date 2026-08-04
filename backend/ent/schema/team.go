package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
)

// Team holds the schema definition for the Team entity.
type Team struct {
	ent.Schema
}

func (Team) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "teams"},
	}
}

func (Team) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (Team) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		field.String("slug").
			MaxLen(100).
			NotEmpty(),
		field.String("description").
			MaxLen(500).
			Optional().
			Default(""),
		field.String("timezone").
			MaxLen(50).
			Optional().
			Default("Asia/Shanghai"),
		field.String("language").
			MaxLen(10).
			Optional().
			Default("zh-CN"),
		field.Int64("owner_id").
			Comment("Owner user ID, same as users.id").
			StorageKey("owner_user_id"),
		field.String("billing_email").
			MaxLen(255).
			Optional().
			Nillable(),
		field.JSON("settings", map[string]interface{}{}).
			Default(map[string]interface{}{}).
			StorageKey("metadata"),
		field.String("status").
			MaxLen(20).
			Default("active"),
	}
}

func (Team) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("owned_teams").
			Field("owner_id").
			Unique().
			Required(),
		edge.To("members", TeamMember.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("departments", Department.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("consumers", Consumer.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("api_keys", APIKey.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("usage_logs", UsageLog.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("audit_logs", TeamAuditLog.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("usage_team_daily", TeamUsageTeamDaily.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("usage_dept_daily", TeamUsageDeptDaily.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("usage_consumer_daily", TeamUsageConsumerDaily.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("usage_model_daily", TeamUsageModelDaily.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Team) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").
			Unique(),
		index.Fields("owner_id"),
		index.Fields("status"),
		index.Fields("deleted_at"),
	}
}

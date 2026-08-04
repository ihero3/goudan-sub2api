package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
)

// TeamUsageTeamDaily holds the schema definition for the TeamUsageTeamDaily entity.
type TeamUsageTeamDaily struct {
	ent.Schema
}

func (TeamUsageTeamDaily) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "team_usage_team_daily"},
	}
}

func (TeamUsageTeamDaily) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (TeamUsageTeamDaily) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("team_id").
			Comment("Team ID"),
		field.Time("bucket_date").
			Comment("Aggregation date"),
		field.Int64("total_requests").
			Default(0),
		field.Int64("input_tokens").
			Default(0),
		field.Int64("output_tokens").
			Default(0),
		field.Int64("cache_creation_tokens").
			Default(0),
		field.Int64("cache_read_tokens").
			Default(0),
		field.Float("total_cost").
			Default(0).
			SchemaType(map[string]string{"postgres": "decimal(20,10)"}),
		field.Float("actual_cost").
			Default(0).
			SchemaType(map[string]string{"postgres": "decimal(20,10)"}),
		field.Time("computed_at").
			Default(time.Now),
	}
}

func (TeamUsageTeamDaily) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("team", Team.Type).
			Ref("usage_team_daily").
			Field("team_id").
			Unique().
			Required(),
	}
}

func (TeamUsageTeamDaily) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("team_id", "bucket_date").
			Unique(),
	}
}

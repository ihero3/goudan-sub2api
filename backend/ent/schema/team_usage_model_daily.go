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

// TeamUsageModelDaily holds the schema definition for the TeamUsageModelDaily entity.
type TeamUsageModelDaily struct {
	ent.Schema
}

func (TeamUsageModelDaily) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "team_usage_model_daily"},
	}
}

func (TeamUsageModelDaily) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (TeamUsageModelDaily) Fields() []ent.Field {
	return []ent.Field {
		field.Int64("team_id").
			Comment("Team ID"),
		field.Int64("department_id").
			Optional().
			Nillable().
			Comment("Department ID"),
		field.Int64("consumer_id").
			Optional().
			Nillable().
			Comment("Consumer ID"),
		field.Time("bucket_date").
			Comment("Aggregation date"),
		field.String("model_name").
			MaxLen(100).
			NotEmpty(),
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

func (TeamUsageModelDaily) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("team", Team.Type).
			Ref("usage_model_daily").
			Field("team_id").
			Unique().
			Required(),
	}
}

func (TeamUsageModelDaily) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("team_id", "bucket_date"),
		index.Fields("department_id", "bucket_date"),
		index.Fields("consumer_id", "bucket_date"),
		index.Fields("model_name", "bucket_date"),
		index.Fields("team_id", "department_id", "consumer_id", "bucket_date", "model_name").
			Unique(),
	}
}

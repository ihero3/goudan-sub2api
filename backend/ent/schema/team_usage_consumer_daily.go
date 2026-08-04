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

// TeamUsageConsumerDaily holds the schema definition for the TeamUsageConsumerDaily entity.
type TeamUsageConsumerDaily struct {
	ent.Schema
}

func (TeamUsageConsumerDaily) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "team_usage_consumer_daily"},
	}
}

func (TeamUsageConsumerDaily) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (TeamUsageConsumerDaily) Fields() []ent.Field {
	return []ent.Field {
		field.Int64("team_id").
			Comment("Team ID"),
		field.Int64("consumer_id").
			Comment("Consumer ID"),
		field.String("consumer_name").
			MaxLen(100).
			Optional().
			Nillable(),
		field.String("consumer_type").
			MaxLen(20).
			Optional().
			Nillable(),
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

func (TeamUsageConsumerDaily) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("team", Team.Type).
			Ref("usage_consumer_daily").
			Field("team_id").
			Unique().
			Required(),
		edge.From("consumer", Consumer.Type).
			Ref("usage_consumer_daily").
			Field("consumer_id").
			Unique().
			Required(),
	}
}

func (TeamUsageConsumerDaily) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("team_id", "bucket_date"),
		index.Fields("consumer_id", "bucket_date"),
		index.Fields("team_id", "consumer_id", "bucket_date").
			Unique(),
	}
}

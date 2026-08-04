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

// TeamMember holds the schema definition for the TeamMember entity.
type TeamMember struct {
	ent.Schema
}

func (TeamMember) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "team_members"},
	}
}

func (TeamMember) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (TeamMember) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("team_id").
			Comment("Team ID"),
		field.Int64("user_id").
			Comment("User ID"),
		field.String("role").
			MaxLen(20).
			Default("member").
			Comment("owner|admin|member|viewer"),
		field.String("status").
			MaxLen(20).
			Default("active").
			Comment("active|inactive|removed"),
		field.String("display_name").
			MaxLen(100).
			Optional().
			Comment("Display name set by inviter, falls back to username if empty"),
		field.Int64("department_id").
			Optional().
			Nillable().
			Comment("Department ID for row-level permission filtering"),
		field.Int64("consumer_id").
			Optional().
			Nillable().
			Comment("Consumer ID for ScopeSelf permission filtering"),
		field.Time("joined_at").
			Default(time.Now),
	}
}

func (TeamMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("team", Team.Type).
			Ref("members").
			Field("team_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("team_memberships").
			Field("user_id").
			Unique().
			Required(),
		edge.From("department", Department.Type).
			Ref("members").
			Field("department_id").
			Unique(),
		edge.From("consumer", Consumer.Type).
			Ref("members").
			Field("consumer_id").
			Unique(),
	}
}

func (TeamMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("team_id", "user_id").
			Unique(),
		index.Fields("user_id"),
		index.Fields("role"),
		index.Fields("department_id"),
		index.Fields("consumer_id"),
	}
}

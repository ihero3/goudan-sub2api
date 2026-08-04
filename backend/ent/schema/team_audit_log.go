package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// TeamAuditLog holds the schema definition for the TeamAuditLog entity.
type TeamAuditLog struct {
	ent.Schema
}

func (TeamAuditLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "team_audit_logs"},
	}
}

func (TeamAuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("team_id").
			Comment("Team ID"),
		field.Int64("user_id").
			Optional().
			Nillable().
			Comment("User ID who performed the action"),
		field.String("action").
			MaxLen(50).
			NotEmpty(),
		field.String("operation_type").
			MaxLen(50).
			Optional().
			Nillable(),
		field.String("resource_type").
			MaxLen(50).
			NotEmpty(),
		field.Int64("resource_id").
			Optional().
			Nillable(),
		field.JSON("changes", map[string]interface{}{}).
			Optional(),
		field.String("ip").
			MaxLen(45).
			Optional().
			Nillable(),
		field.String("user_agent").
			SchemaType(map[string]string{"postgres": "text"}).
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (TeamAuditLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("team", Team.Type).
			Ref("audit_logs").
			Field("team_id").
			Unique().
			Required(),
	}
}

func (TeamAuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("team_id"),
		index.Fields("created_at"),
		index.Fields("operation_type"),
		index.Fields("resource_type", "resource_id"),
	}
}

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

// Consumer holds the schema definition for the Consumer entity.
type Consumer struct {
	ent.Schema
}

func (Consumer) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "consumers"},
	}
}

func (Consumer) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (Consumer) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("team_id").
			Comment("Team ID"),
		field.Int64("department_id").
			Optional().
			Nillable().
			Comment("Department ID"),
		field.String("type").
			MaxLen(20).
			Default("person").
			Comment("Type: person|application|service_account"),
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		field.String("email").
			MaxLen(255).
			Optional().
			Nillable(),
		field.String("phone").
			MaxLen(50).
			Optional().
			Nillable(),
		field.String("title").
			MaxLen(100).
			Optional().
			Nillable(),
		field.String("app_id").
			MaxLen(100).
			Optional().
			Nillable().
			Comment("Unique identifier for application type"),
		field.String("app_description").
			SchemaType(map[string]string{"postgres": "text"}).
			Optional().
			Nillable(),
		field.String("external_id").
			MaxLen(255).
			Optional().
			Nillable().
			Comment("External system ID (LDAP UID/OIDC sub/employee ID)"),
		field.String("source").
			MaxLen(50).
			Default("manual").
			Comment("Source: manual|ldap|oidc|google"),
		field.Time("deactivated_at").
			Optional().
			Nillable(),
		field.String("status").
			MaxLen(20).
			Default("active"),
		field.JSON("settings", map[string]interface{}{}).
			Default(map[string]interface{}{}),
	}
}

func (Consumer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("team", Team.Type).
			Ref("consumers").
			Field("team_id").
			Unique().
			Required(),
		edge.From("department", Department.Type).
			Ref("consumers").
			Field("department_id").
			Unique(),
		edge.To("members", TeamMember.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("api_keys", APIKey.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("usage_consumer_daily", TeamUsageConsumerDaily.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Consumer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("team_id"),
		index.Fields("department_id"),
		index.Fields("type"),
		index.Fields("external_id"),
		index.Fields("status"),
		index.Fields("team_id", "app_id").
			Unique(),
	}
}

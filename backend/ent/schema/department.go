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

// Department holds the schema definition for the Department entity.
type Department struct {
	ent.Schema
}

func (Department) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "departments"},
	}
}

func (Department) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (Department) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("team_id").
			Comment("Team ID"),
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		field.String("description").
			SchemaType(map[string]string{"postgres": "text"}).
			Default(""),
		field.String("cost_center_code").
			MaxLen(50).
			Optional().
			Nillable(),
		field.Int64("parent_id").
			Optional().
			Nillable().
			Comment("Parent department ID for hierarchical structure"),
		field.Int("level").
			Default(0).
			Comment("Hierarchy depth, root department is 0"),
		field.String("path").
			MaxLen(500).
			Default("/").
			Comment("Full path ID string, e.g. /1/5/12/"),
		field.String("external_id").
			MaxLen(255).
			Optional().
			Nillable().
			Comment("External system ID (LDAP/OIDC/Google Workspace)"),
		field.String("source").
			MaxLen(50).
			Default("manual").
			Comment("Source: manual|ldap|oidc|google"),
		field.Int("sort_order").
			Default(0),
		field.String("status").
			MaxLen(20).
			Default("active"),
	}
}

func (Department) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("team", Team.Type).
			Ref("departments").
			Field("team_id").
			Unique().
			Required(),
		edge.To("members", TeamMember.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("consumers", Consumer.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("api_keys", APIKey.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("usage_dept_daily", TeamUsageDeptDaily.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (Department) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("team_id"),
		index.Fields("parent_id"),
		index.Fields("external_id"),
		index.Fields("path"),
		index.Fields("team_id", "name").
			Unique(),
	}
}

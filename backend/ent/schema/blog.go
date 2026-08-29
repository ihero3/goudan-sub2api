package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Blog holds the schema definition for the Blog entity.
type Blog struct {
	ent.Schema
}

func (Blog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "blogs"},
	}
}

func (Blog) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").
			MaxLen(500).
			NotEmpty().
			Comment("博客标题"),
		field.String("content").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			NotEmpty().
			Comment("博客内容（支持 Markdown）"),
		field.String("summary").
			MaxLen(1000).
			Optional().
			Default("").
			Comment("博客摘要"),
		field.String("cover_image").
			MaxLen(1000).
			Optional().
			Default("").
			Comment("封面图片 URL"),
		field.String("status").
			MaxLen(20).
			Default("draft").
			Comment("状态: draft, published"),
		field.String("tags").
			Optional().
			Default("").
			Comment("标签（逗号分隔）"),
		field.Time("published_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("发布时间"),
		field.Int64("created_by").
			Optional().
			Nillable().
			Comment("创建人用户ID"),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("deleted_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (Blog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("created_at"),
		index.Fields("published_at"),
	}
}
// File updated by protoc-gen-ent.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type CustomConfigResource struct {
	ent.Schema
}

func (CustomConfigResource) Fields() []ent.Field {
	return []ent.Field{field.String("resource_id").Unique(), field.String("name").Immutable(), field.String("config").Immutable(), field.String("description").Optional(), field.String("tenant_id").Immutable(), field.String("created_at").Immutable().SchemaType(map[string]string{"postgres": "TIMESTAMP"}), field.String("updated_at").SchemaType(map[string]string{"postgres": "TIMESTAMP"})}
}
func (CustomConfigResource) Edges() []ent.Edge {
	return []ent.Edge{edge.To("instances", InstanceResource.Type)}
}
func (CustomConfigResource) Annotations() []schema.Annotation {
	return nil
}
func (CustomConfigResource) Indexes() []ent.Index {
	return []ent.Index{index.Fields("name", "tenant_id").Unique(), index.Fields("tenant_id")}
}

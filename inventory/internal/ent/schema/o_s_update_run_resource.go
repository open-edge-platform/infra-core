// File updated by protoc-gen-ent.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type OSUpdateRunResource struct {
	ent.Schema
}

func (OSUpdateRunResource) Fields() []ent.Field {
	return []ent.Field{field.String("resource_id").Unique(), field.String("name").Immutable(), field.String("description").Optional().Immutable(), field.Enum("status_indicator").Values("STATUS_INDICATION_UNSPECIFIED", "STATUS_INDICATION_ERROR", "STATUS_INDICATION_IN_PROGRESS", "STATUS_INDICATION_IDLE"), field.String("status"), field.String("status_details"), field.String("status_timestamp").SchemaType(map[string]string{"postgres": "TIMESTAMP"}), field.String("start_time").Immutable().SchemaType(map[string]string{"postgres": "TIMESTAMP"}), field.String("end_time").SchemaType(map[string]string{"postgres": "TIMESTAMP"}), field.String("tenant_id").Immutable(), field.String("created_at").Immutable().SchemaType(map[string]string{"postgres": "TIMESTAMP"}), field.String("updated_at").SchemaType(map[string]string{"postgres": "TIMESTAMP"})}
}
func (OSUpdateRunResource) Edges() []ent.Edge {
	return []ent.Edge{edge.To("applied_policy", OSUpdatePolicyResource.Type).Required().Unique(), edge.To("instance", InstanceResource.Type).Required().Unique()}
}
func (OSUpdateRunResource) Annotations() []schema.Annotation {
	return nil
}

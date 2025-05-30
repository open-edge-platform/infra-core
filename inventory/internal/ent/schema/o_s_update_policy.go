// File updated by protoc-gen-ent.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type OSUpdatePolicy struct {
	ent.Schema
}

func (OSUpdatePolicy) Fields() []ent.Field {
	return []ent.Field{field.String("resource_id").Unique(), field.String("installed_packages").Optional(), field.String("update_sources").Optional(), field.String("kernel_command").Optional(), field.Enum("update_policy").Optional().Values("UPDATE_POLICY_UNSPECIFIED", "UPDATE_POLICY_LATEST", "UPDATE_POLICY_TARGET"), field.String("tenant_id").Immutable(), field.String("created_at").Immutable().SchemaType(map[string]string{"postgres": "TIMESTAMP"}), field.String("updated_at").SchemaType(map[string]string{"postgres": "TIMESTAMP"})}
}
func (OSUpdatePolicy) Edges() []ent.Edge {
	return []ent.Edge{edge.To("target_os", OperatingSystemResource.Type).Unique()}
}
func (OSUpdatePolicy) Annotations() []schema.Annotation {
	return nil
}

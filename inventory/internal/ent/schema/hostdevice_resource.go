// File updated by protoc-gen-ent.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type HostdeviceResource struct {
	ent.Schema
}

func (HostdeviceResource) Fields() []ent.Field {
	return []ent.Field{field.String("resource_id").Unique(), field.String("version").Optional(), field.String("hostname").Optional(), field.String("operational_state").Optional(), field.String("build_number").Optional(), field.String("sku").Optional(), field.String("features").Optional(), field.String("device_guid").Optional(), field.String("control_mode").Optional(), field.String("dns_suffix").Optional(), field.String("network_status").Optional(), field.String("remote_status").Optional(), field.String("remote_trigger").Optional(), field.String("mps_hostname").Optional(), field.String("tenant_id").Immutable(), field.String("created_at").Immutable().SchemaType(map[string]string{"postgres": "TIMESTAMP"}), field.String("updated_at").SchemaType(map[string]string{"postgres": "TIMESTAMP"})}
}
func (HostdeviceResource) Edges() []ent.Edge {
	return []ent.Edge{edge.To("host", HostResource.Type).Required().Unique()}
}
func (HostdeviceResource) Annotations() []schema.Annotation {
	return nil
}
func (HostdeviceResource) Indexes() []ent.Index {
	return []ent.Index{index.Fields("tenant_id")}
}

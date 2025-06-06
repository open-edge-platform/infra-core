// File updated by protoc-gen-ent.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type InstanceResource struct {
	ent.Schema
}

func (InstanceResource) Fields() []ent.Field {
	return []ent.Field{field.String("resource_id").Unique(), field.Enum("kind").Optional().Values("INSTANCE_KIND_UNSPECIFIED", "INSTANCE_KIND_VM", "INSTANCE_KIND_METAL"), field.String("name").Optional(), field.Enum("desired_state").Optional().Values("INSTANCE_STATE_UNSPECIFIED", "INSTANCE_STATE_RUNNING", "INSTANCE_STATE_DELETED", "INSTANCE_STATE_UNTRUSTED"), field.Enum("current_state").Optional().Values("INSTANCE_STATE_UNSPECIFIED", "INSTANCE_STATE_RUNNING", "INSTANCE_STATE_DELETED", "INSTANCE_STATE_UNTRUSTED"), field.Uint64("vm_memory_bytes").Optional(), field.Uint32("vm_cpu_cores").Optional(), field.Uint64("vm_storage_bytes").Optional(), field.Enum("security_feature").Optional().Immutable().Values("SECURITY_FEATURE_UNSPECIFIED", "SECURITY_FEATURE_NONE", "SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION"), field.String("instance_status").Optional(), field.Enum("instance_status_indicator").Optional().Values("STATUS_INDICATION_UNSPECIFIED", "STATUS_INDICATION_ERROR", "STATUS_INDICATION_IN_PROGRESS", "STATUS_INDICATION_IDLE"), field.Uint64("instance_status_timestamp").Optional(), field.String("provisioning_status").Optional(), field.Enum("provisioning_status_indicator").Optional().Values("STATUS_INDICATION_UNSPECIFIED", "STATUS_INDICATION_ERROR", "STATUS_INDICATION_IN_PROGRESS", "STATUS_INDICATION_IDLE"), field.Uint64("provisioning_status_timestamp").Optional(), field.String("update_status").Optional(), field.Enum("update_status_indicator").Optional().Values("STATUS_INDICATION_UNSPECIFIED", "STATUS_INDICATION_ERROR", "STATUS_INDICATION_IN_PROGRESS", "STATUS_INDICATION_IDLE"), field.Uint64("update_status_timestamp").Optional(), field.String("update_status_detail").Optional(), field.String("trusted_attestation_status").Optional(), field.Enum("trusted_attestation_status_indicator").Optional().Values("STATUS_INDICATION_UNSPECIFIED", "STATUS_INDICATION_ERROR", "STATUS_INDICATION_IN_PROGRESS", "STATUS_INDICATION_IDLE"), field.Uint64("trusted_attestation_status_timestamp").Optional(), field.String("existing_cves").Optional(), field.String("tenant_id").Immutable(), field.String("instance_status_detail").Optional(), field.String("created_at").Immutable().SchemaType(map[string]string{"postgres": "TIMESTAMP"}), field.String("updated_at").SchemaType(map[string]string{"postgres": "TIMESTAMP"})}
}
func (InstanceResource) Edges() []ent.Edge {
	return []ent.Edge{edge.To("host", HostResource.Type).Unique(), edge.To("desired_os", OperatingSystemResource.Type).Required().Unique(), edge.To("current_os", OperatingSystemResource.Type).Unique(), edge.From("workload_members", WorkloadMember.Type).Ref("instance"), edge.To("provider", ProviderResource.Type).Unique(), edge.To("localaccount", LocalAccountResource.Type).Unique()}
}
func (InstanceResource) Annotations() []schema.Annotation {
	return nil
}
func (InstanceResource) Indexes() []ent.Index {
	return []ent.Index{index.Fields("tenant_id")}
}

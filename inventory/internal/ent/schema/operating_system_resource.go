// File updated by protoc-gen-ent.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OperatingSystemResource struct {
	ent.Schema
}

func (OperatingSystemResource) Fields() []ent.Field {
	return []ent.Field{field.String("resource_id").Unique(), field.String("name").Optional().Immutable(), field.String("architecture").Optional(), field.String("kernel_command").Optional(), field.String("update_sources").Optional(), field.String("image_url").Optional().Immutable(), field.String("image_id").Optional().Immutable(), field.String("sha256").Optional().Immutable(), field.String("profile_name").Optional().Immutable(), field.String("profile_version").Optional().Immutable(), field.String("installed_packages").Optional(), field.Enum("security_feature").Optional().Immutable().Values("SECURITY_FEATURE_UNSPECIFIED", "SECURITY_FEATURE_NONE", "SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION"), field.Enum("os_type").Optional().Immutable().Values("OS_TYPE_UNSPECIFIED", "OS_TYPE_MUTABLE", "OS_TYPE_IMMUTABLE"), field.Enum("os_provider").Immutable().Values("OS_PROVIDER_KIND_UNSPECIFIED", "OS_PROVIDER_KIND_INFRA", "OS_PROVIDER_KIND_LENOVO"), field.String("platform_bundle").Optional().Immutable(), field.String("description").Optional().Immutable(), field.String("metadata").Optional(), field.String("existing_cves_url").Optional().Immutable(), field.String("existing_cves").Optional(), field.String("fixed_cves_url").Optional().Immutable(), field.String("fixed_cves").Optional(), field.String("tenant_id").Immutable(), field.String("created_at").Immutable().SchemaType(map[string]string{"postgres": "TIMESTAMP"}), field.String("updated_at").SchemaType(map[string]string{"postgres": "TIMESTAMP"})}
}
func (OperatingSystemResource) Edges() []ent.Edge {
	return nil
}
func (OperatingSystemResource) Annotations() []schema.Annotation {
	return nil
}
func (OperatingSystemResource) Indexes() []ent.Index {
	return []ent.Index{index.Fields("name", "tenant_id").Unique(), index.Fields("tenant_id")}
}

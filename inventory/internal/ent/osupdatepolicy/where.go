// Code generated by ent, DO NOT EDIT.

package osupdatepolicy

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldLTE(FieldID, id))
}

// ResourceID applies equality check predicate on the "resource_id" field. It's identical to ResourceIDEQ.
func ResourceID(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldResourceID, v))
}

// InstalledPackages applies equality check predicate on the "installed_packages" field. It's identical to InstalledPackagesEQ.
func InstalledPackages(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldInstalledPackages, v))
}

// UpdateSources applies equality check predicate on the "update_sources" field. It's identical to UpdateSourcesEQ.
func UpdateSources(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldUpdateSources, v))
}

// KernelCommand applies equality check predicate on the "kernel_command" field. It's identical to KernelCommandEQ.
func KernelCommand(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldKernelCommand, v))
}

// TenantID applies equality check predicate on the "tenant_id" field. It's identical to TenantIDEQ.
func TenantID(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldTenantID, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldCreatedAt, v))
}

// UpdatedAt applies equality check predicate on the "updated_at" field. It's identical to UpdatedAtEQ.
func UpdatedAt(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldUpdatedAt, v))
}

// ResourceIDEQ applies the EQ predicate on the "resource_id" field.
func ResourceIDEQ(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldResourceID, v))
}

// ResourceIDNEQ applies the NEQ predicate on the "resource_id" field.
func ResourceIDNEQ(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNEQ(FieldResourceID, v))
}

// ResourceIDIn applies the In predicate on the "resource_id" field.
func ResourceIDIn(vs ...string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldIn(FieldResourceID, vs...))
}

// ResourceIDNotIn applies the NotIn predicate on the "resource_id" field.
func ResourceIDNotIn(vs ...string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNotIn(FieldResourceID, vs...))
}

// ResourceIDGT applies the GT predicate on the "resource_id" field.
func ResourceIDGT(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldGT(FieldResourceID, v))
}

// ResourceIDGTE applies the GTE predicate on the "resource_id" field.
func ResourceIDGTE(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldGTE(FieldResourceID, v))
}

// ResourceIDLT applies the LT predicate on the "resource_id" field.
func ResourceIDLT(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldLT(FieldResourceID, v))
}

// ResourceIDLTE applies the LTE predicate on the "resource_id" field.
func ResourceIDLTE(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldLTE(FieldResourceID, v))
}

// ResourceIDContains applies the Contains predicate on the "resource_id" field.
func ResourceIDContains(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldContains(FieldResourceID, v))
}

// ResourceIDHasPrefix applies the HasPrefix predicate on the "resource_id" field.
func ResourceIDHasPrefix(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldHasPrefix(FieldResourceID, v))
}

// ResourceIDHasSuffix applies the HasSuffix predicate on the "resource_id" field.
func ResourceIDHasSuffix(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldHasSuffix(FieldResourceID, v))
}

// ResourceIDEqualFold applies the EqualFold predicate on the "resource_id" field.
func ResourceIDEqualFold(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEqualFold(FieldResourceID, v))
}

// ResourceIDContainsFold applies the ContainsFold predicate on the "resource_id" field.
func ResourceIDContainsFold(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldContainsFold(FieldResourceID, v))
}

// InstalledPackagesEQ applies the EQ predicate on the "installed_packages" field.
func InstalledPackagesEQ(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldInstalledPackages, v))
}

// InstalledPackagesNEQ applies the NEQ predicate on the "installed_packages" field.
func InstalledPackagesNEQ(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNEQ(FieldInstalledPackages, v))
}

// InstalledPackagesIn applies the In predicate on the "installed_packages" field.
func InstalledPackagesIn(vs ...string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldIn(FieldInstalledPackages, vs...))
}

// InstalledPackagesNotIn applies the NotIn predicate on the "installed_packages" field.
func InstalledPackagesNotIn(vs ...string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNotIn(FieldInstalledPackages, vs...))
}

// InstalledPackagesGT applies the GT predicate on the "installed_packages" field.
func InstalledPackagesGT(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldGT(FieldInstalledPackages, v))
}

// InstalledPackagesGTE applies the GTE predicate on the "installed_packages" field.
func InstalledPackagesGTE(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldGTE(FieldInstalledPackages, v))
}

// InstalledPackagesLT applies the LT predicate on the "installed_packages" field.
func InstalledPackagesLT(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldLT(FieldInstalledPackages, v))
}

// InstalledPackagesLTE applies the LTE predicate on the "installed_packages" field.
func InstalledPackagesLTE(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldLTE(FieldInstalledPackages, v))
}

// InstalledPackagesContains applies the Contains predicate on the "installed_packages" field.
func InstalledPackagesContains(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldContains(FieldInstalledPackages, v))
}

// InstalledPackagesHasPrefix applies the HasPrefix predicate on the "installed_packages" field.
func InstalledPackagesHasPrefix(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldHasPrefix(FieldInstalledPackages, v))
}

// InstalledPackagesHasSuffix applies the HasSuffix predicate on the "installed_packages" field.
func InstalledPackagesHasSuffix(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldHasSuffix(FieldInstalledPackages, v))
}

// InstalledPackagesIsNil applies the IsNil predicate on the "installed_packages" field.
func InstalledPackagesIsNil() predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldIsNull(FieldInstalledPackages))
}

// InstalledPackagesNotNil applies the NotNil predicate on the "installed_packages" field.
func InstalledPackagesNotNil() predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNotNull(FieldInstalledPackages))
}

// InstalledPackagesEqualFold applies the EqualFold predicate on the "installed_packages" field.
func InstalledPackagesEqualFold(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEqualFold(FieldInstalledPackages, v))
}

// InstalledPackagesContainsFold applies the ContainsFold predicate on the "installed_packages" field.
func InstalledPackagesContainsFold(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldContainsFold(FieldInstalledPackages, v))
}

// UpdateSourcesEQ applies the EQ predicate on the "update_sources" field.
func UpdateSourcesEQ(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldUpdateSources, v))
}

// UpdateSourcesNEQ applies the NEQ predicate on the "update_sources" field.
func UpdateSourcesNEQ(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNEQ(FieldUpdateSources, v))
}

// UpdateSourcesIn applies the In predicate on the "update_sources" field.
func UpdateSourcesIn(vs ...string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldIn(FieldUpdateSources, vs...))
}

// UpdateSourcesNotIn applies the NotIn predicate on the "update_sources" field.
func UpdateSourcesNotIn(vs ...string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNotIn(FieldUpdateSources, vs...))
}

// UpdateSourcesGT applies the GT predicate on the "update_sources" field.
func UpdateSourcesGT(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldGT(FieldUpdateSources, v))
}

// UpdateSourcesGTE applies the GTE predicate on the "update_sources" field.
func UpdateSourcesGTE(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldGTE(FieldUpdateSources, v))
}

// UpdateSourcesLT applies the LT predicate on the "update_sources" field.
func UpdateSourcesLT(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldLT(FieldUpdateSources, v))
}

// UpdateSourcesLTE applies the LTE predicate on the "update_sources" field.
func UpdateSourcesLTE(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldLTE(FieldUpdateSources, v))
}

// UpdateSourcesContains applies the Contains predicate on the "update_sources" field.
func UpdateSourcesContains(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldContains(FieldUpdateSources, v))
}

// UpdateSourcesHasPrefix applies the HasPrefix predicate on the "update_sources" field.
func UpdateSourcesHasPrefix(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldHasPrefix(FieldUpdateSources, v))
}

// UpdateSourcesHasSuffix applies the HasSuffix predicate on the "update_sources" field.
func UpdateSourcesHasSuffix(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldHasSuffix(FieldUpdateSources, v))
}

// UpdateSourcesIsNil applies the IsNil predicate on the "update_sources" field.
func UpdateSourcesIsNil() predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldIsNull(FieldUpdateSources))
}

// UpdateSourcesNotNil applies the NotNil predicate on the "update_sources" field.
func UpdateSourcesNotNil() predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNotNull(FieldUpdateSources))
}

// UpdateSourcesEqualFold applies the EqualFold predicate on the "update_sources" field.
func UpdateSourcesEqualFold(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEqualFold(FieldUpdateSources, v))
}

// UpdateSourcesContainsFold applies the ContainsFold predicate on the "update_sources" field.
func UpdateSourcesContainsFold(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldContainsFold(FieldUpdateSources, v))
}

// KernelCommandEQ applies the EQ predicate on the "kernel_command" field.
func KernelCommandEQ(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldKernelCommand, v))
}

// KernelCommandNEQ applies the NEQ predicate on the "kernel_command" field.
func KernelCommandNEQ(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNEQ(FieldKernelCommand, v))
}

// KernelCommandIn applies the In predicate on the "kernel_command" field.
func KernelCommandIn(vs ...string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldIn(FieldKernelCommand, vs...))
}

// KernelCommandNotIn applies the NotIn predicate on the "kernel_command" field.
func KernelCommandNotIn(vs ...string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNotIn(FieldKernelCommand, vs...))
}

// KernelCommandGT applies the GT predicate on the "kernel_command" field.
func KernelCommandGT(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldGT(FieldKernelCommand, v))
}

// KernelCommandGTE applies the GTE predicate on the "kernel_command" field.
func KernelCommandGTE(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldGTE(FieldKernelCommand, v))
}

// KernelCommandLT applies the LT predicate on the "kernel_command" field.
func KernelCommandLT(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldLT(FieldKernelCommand, v))
}

// KernelCommandLTE applies the LTE predicate on the "kernel_command" field.
func KernelCommandLTE(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldLTE(FieldKernelCommand, v))
}

// KernelCommandContains applies the Contains predicate on the "kernel_command" field.
func KernelCommandContains(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldContains(FieldKernelCommand, v))
}

// KernelCommandHasPrefix applies the HasPrefix predicate on the "kernel_command" field.
func KernelCommandHasPrefix(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldHasPrefix(FieldKernelCommand, v))
}

// KernelCommandHasSuffix applies the HasSuffix predicate on the "kernel_command" field.
func KernelCommandHasSuffix(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldHasSuffix(FieldKernelCommand, v))
}

// KernelCommandIsNil applies the IsNil predicate on the "kernel_command" field.
func KernelCommandIsNil() predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldIsNull(FieldKernelCommand))
}

// KernelCommandNotNil applies the NotNil predicate on the "kernel_command" field.
func KernelCommandNotNil() predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNotNull(FieldKernelCommand))
}

// KernelCommandEqualFold applies the EqualFold predicate on the "kernel_command" field.
func KernelCommandEqualFold(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEqualFold(FieldKernelCommand, v))
}

// KernelCommandContainsFold applies the ContainsFold predicate on the "kernel_command" field.
func KernelCommandContainsFold(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldContainsFold(FieldKernelCommand, v))
}

// UpdatePolicyEQ applies the EQ predicate on the "update_policy" field.
func UpdatePolicyEQ(v UpdatePolicy) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldUpdatePolicy, v))
}

// UpdatePolicyNEQ applies the NEQ predicate on the "update_policy" field.
func UpdatePolicyNEQ(v UpdatePolicy) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNEQ(FieldUpdatePolicy, v))
}

// UpdatePolicyIn applies the In predicate on the "update_policy" field.
func UpdatePolicyIn(vs ...UpdatePolicy) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldIn(FieldUpdatePolicy, vs...))
}

// UpdatePolicyNotIn applies the NotIn predicate on the "update_policy" field.
func UpdatePolicyNotIn(vs ...UpdatePolicy) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNotIn(FieldUpdatePolicy, vs...))
}

// UpdatePolicyIsNil applies the IsNil predicate on the "update_policy" field.
func UpdatePolicyIsNil() predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldIsNull(FieldUpdatePolicy))
}

// UpdatePolicyNotNil applies the NotNil predicate on the "update_policy" field.
func UpdatePolicyNotNil() predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNotNull(FieldUpdatePolicy))
}

// TenantIDEQ applies the EQ predicate on the "tenant_id" field.
func TenantIDEQ(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldTenantID, v))
}

// TenantIDNEQ applies the NEQ predicate on the "tenant_id" field.
func TenantIDNEQ(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNEQ(FieldTenantID, v))
}

// TenantIDIn applies the In predicate on the "tenant_id" field.
func TenantIDIn(vs ...string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldIn(FieldTenantID, vs...))
}

// TenantIDNotIn applies the NotIn predicate on the "tenant_id" field.
func TenantIDNotIn(vs ...string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNotIn(FieldTenantID, vs...))
}

// TenantIDGT applies the GT predicate on the "tenant_id" field.
func TenantIDGT(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldGT(FieldTenantID, v))
}

// TenantIDGTE applies the GTE predicate on the "tenant_id" field.
func TenantIDGTE(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldGTE(FieldTenantID, v))
}

// TenantIDLT applies the LT predicate on the "tenant_id" field.
func TenantIDLT(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldLT(FieldTenantID, v))
}

// TenantIDLTE applies the LTE predicate on the "tenant_id" field.
func TenantIDLTE(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldLTE(FieldTenantID, v))
}

// TenantIDContains applies the Contains predicate on the "tenant_id" field.
func TenantIDContains(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldContains(FieldTenantID, v))
}

// TenantIDHasPrefix applies the HasPrefix predicate on the "tenant_id" field.
func TenantIDHasPrefix(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldHasPrefix(FieldTenantID, v))
}

// TenantIDHasSuffix applies the HasSuffix predicate on the "tenant_id" field.
func TenantIDHasSuffix(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldHasSuffix(FieldTenantID, v))
}

// TenantIDEqualFold applies the EqualFold predicate on the "tenant_id" field.
func TenantIDEqualFold(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEqualFold(FieldTenantID, v))
}

// TenantIDContainsFold applies the ContainsFold predicate on the "tenant_id" field.
func TenantIDContainsFold(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldContainsFold(FieldTenantID, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldLTE(FieldCreatedAt, v))
}

// CreatedAtContains applies the Contains predicate on the "created_at" field.
func CreatedAtContains(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldContains(FieldCreatedAt, v))
}

// CreatedAtHasPrefix applies the HasPrefix predicate on the "created_at" field.
func CreatedAtHasPrefix(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldHasPrefix(FieldCreatedAt, v))
}

// CreatedAtHasSuffix applies the HasSuffix predicate on the "created_at" field.
func CreatedAtHasSuffix(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldHasSuffix(FieldCreatedAt, v))
}

// CreatedAtEqualFold applies the EqualFold predicate on the "created_at" field.
func CreatedAtEqualFold(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEqualFold(FieldCreatedAt, v))
}

// CreatedAtContainsFold applies the ContainsFold predicate on the "created_at" field.
func CreatedAtContainsFold(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldContainsFold(FieldCreatedAt, v))
}

// UpdatedAtEQ applies the EQ predicate on the "updated_at" field.
func UpdatedAtEQ(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEQ(FieldUpdatedAt, v))
}

// UpdatedAtNEQ applies the NEQ predicate on the "updated_at" field.
func UpdatedAtNEQ(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNEQ(FieldUpdatedAt, v))
}

// UpdatedAtIn applies the In predicate on the "updated_at" field.
func UpdatedAtIn(vs ...string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldIn(FieldUpdatedAt, vs...))
}

// UpdatedAtNotIn applies the NotIn predicate on the "updated_at" field.
func UpdatedAtNotIn(vs ...string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldNotIn(FieldUpdatedAt, vs...))
}

// UpdatedAtGT applies the GT predicate on the "updated_at" field.
func UpdatedAtGT(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldGT(FieldUpdatedAt, v))
}

// UpdatedAtGTE applies the GTE predicate on the "updated_at" field.
func UpdatedAtGTE(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldGTE(FieldUpdatedAt, v))
}

// UpdatedAtLT applies the LT predicate on the "updated_at" field.
func UpdatedAtLT(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldLT(FieldUpdatedAt, v))
}

// UpdatedAtLTE applies the LTE predicate on the "updated_at" field.
func UpdatedAtLTE(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldLTE(FieldUpdatedAt, v))
}

// UpdatedAtContains applies the Contains predicate on the "updated_at" field.
func UpdatedAtContains(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldContains(FieldUpdatedAt, v))
}

// UpdatedAtHasPrefix applies the HasPrefix predicate on the "updated_at" field.
func UpdatedAtHasPrefix(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldHasPrefix(FieldUpdatedAt, v))
}

// UpdatedAtHasSuffix applies the HasSuffix predicate on the "updated_at" field.
func UpdatedAtHasSuffix(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldHasSuffix(FieldUpdatedAt, v))
}

// UpdatedAtEqualFold applies the EqualFold predicate on the "updated_at" field.
func UpdatedAtEqualFold(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldEqualFold(FieldUpdatedAt, v))
}

// UpdatedAtContainsFold applies the ContainsFold predicate on the "updated_at" field.
func UpdatedAtContainsFold(v string) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.FieldContainsFold(FieldUpdatedAt, v))
}

// HasTargetOs applies the HasEdge predicate on the "target_os" edge.
func HasTargetOs() predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, TargetOsTable, TargetOsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasTargetOsWith applies the HasEdge predicate on the "target_os" edge with a given conditions (other predicates).
func HasTargetOsWith(preds ...predicate.OperatingSystemResource) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(func(s *sql.Selector) {
		step := newTargetOsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.OSUpdatePolicy) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.OSUpdatePolicy) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.OSUpdatePolicy) predicate.OSUpdatePolicy {
	return predicate.OSUpdatePolicy(sql.NotPredicates(p))
}

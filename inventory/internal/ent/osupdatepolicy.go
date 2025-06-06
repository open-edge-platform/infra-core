// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/operatingsystemresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/osupdatepolicy"
)

// OSUpdatePolicy is the model entity for the OSUpdatePolicy schema.
type OSUpdatePolicy struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// ResourceID holds the value of the "resource_id" field.
	ResourceID string `json:"resource_id,omitempty"`
	// InstalledPackages holds the value of the "installed_packages" field.
	InstalledPackages string `json:"installed_packages,omitempty"`
	// UpdateSources holds the value of the "update_sources" field.
	UpdateSources string `json:"update_sources,omitempty"`
	// KernelCommand holds the value of the "kernel_command" field.
	KernelCommand string `json:"kernel_command,omitempty"`
	// UpdatePolicy holds the value of the "update_policy" field.
	UpdatePolicy osupdatepolicy.UpdatePolicy `json:"update_policy,omitempty"`
	// TenantID holds the value of the "tenant_id" field.
	TenantID string `json:"tenant_id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt string `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt string `json:"updated_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the OSUpdatePolicyQuery when eager-loading is set.
	Edges                      OSUpdatePolicyEdges `json:"edges"`
	os_update_policy_target_os *int
	selectValues               sql.SelectValues
}

// OSUpdatePolicyEdges holds the relations/edges for other nodes in the graph.
type OSUpdatePolicyEdges struct {
	// TargetOs holds the value of the target_os edge.
	TargetOs *OperatingSystemResource `json:"target_os,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// TargetOsOrErr returns the TargetOs value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e OSUpdatePolicyEdges) TargetOsOrErr() (*OperatingSystemResource, error) {
	if e.TargetOs != nil {
		return e.TargetOs, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: operatingsystemresource.Label}
	}
	return nil, &NotLoadedError{edge: "target_os"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*OSUpdatePolicy) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case osupdatepolicy.FieldID:
			values[i] = new(sql.NullInt64)
		case osupdatepolicy.FieldResourceID, osupdatepolicy.FieldInstalledPackages, osupdatepolicy.FieldUpdateSources, osupdatepolicy.FieldKernelCommand, osupdatepolicy.FieldUpdatePolicy, osupdatepolicy.FieldTenantID, osupdatepolicy.FieldCreatedAt, osupdatepolicy.FieldUpdatedAt:
			values[i] = new(sql.NullString)
		case osupdatepolicy.ForeignKeys[0]: // os_update_policy_target_os
			values[i] = new(sql.NullInt64)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the OSUpdatePolicy fields.
func (oup *OSUpdatePolicy) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case osupdatepolicy.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			oup.ID = int(value.Int64)
		case osupdatepolicy.FieldResourceID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field resource_id", values[i])
			} else if value.Valid {
				oup.ResourceID = value.String
			}
		case osupdatepolicy.FieldInstalledPackages:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field installed_packages", values[i])
			} else if value.Valid {
				oup.InstalledPackages = value.String
			}
		case osupdatepolicy.FieldUpdateSources:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field update_sources", values[i])
			} else if value.Valid {
				oup.UpdateSources = value.String
			}
		case osupdatepolicy.FieldKernelCommand:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field kernel_command", values[i])
			} else if value.Valid {
				oup.KernelCommand = value.String
			}
		case osupdatepolicy.FieldUpdatePolicy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field update_policy", values[i])
			} else if value.Valid {
				oup.UpdatePolicy = osupdatepolicy.UpdatePolicy(value.String)
			}
		case osupdatepolicy.FieldTenantID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field tenant_id", values[i])
			} else if value.Valid {
				oup.TenantID = value.String
			}
		case osupdatepolicy.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				oup.CreatedAt = value.String
			}
		case osupdatepolicy.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				oup.UpdatedAt = value.String
			}
		case osupdatepolicy.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for edge-field os_update_policy_target_os", value)
			} else if value.Valid {
				oup.os_update_policy_target_os = new(int)
				*oup.os_update_policy_target_os = int(value.Int64)
			}
		default:
			oup.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the OSUpdatePolicy.
// This includes values selected through modifiers, order, etc.
func (oup *OSUpdatePolicy) Value(name string) (ent.Value, error) {
	return oup.selectValues.Get(name)
}

// QueryTargetOs queries the "target_os" edge of the OSUpdatePolicy entity.
func (oup *OSUpdatePolicy) QueryTargetOs() *OperatingSystemResourceQuery {
	return NewOSUpdatePolicyClient(oup.config).QueryTargetOs(oup)
}

// Update returns a builder for updating this OSUpdatePolicy.
// Note that you need to call OSUpdatePolicy.Unwrap() before calling this method if this OSUpdatePolicy
// was returned from a transaction, and the transaction was committed or rolled back.
func (oup *OSUpdatePolicy) Update() *OSUpdatePolicyUpdateOne {
	return NewOSUpdatePolicyClient(oup.config).UpdateOne(oup)
}

// Unwrap unwraps the OSUpdatePolicy entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (oup *OSUpdatePolicy) Unwrap() *OSUpdatePolicy {
	_tx, ok := oup.config.driver.(*txDriver)
	if !ok {
		panic("ent: OSUpdatePolicy is not a transactional entity")
	}
	oup.config.driver = _tx.drv
	return oup
}

// String implements the fmt.Stringer.
func (oup *OSUpdatePolicy) String() string {
	var builder strings.Builder
	builder.WriteString("OSUpdatePolicy(")
	builder.WriteString(fmt.Sprintf("id=%v, ", oup.ID))
	builder.WriteString("resource_id=")
	builder.WriteString(oup.ResourceID)
	builder.WriteString(", ")
	builder.WriteString("installed_packages=")
	builder.WriteString(oup.InstalledPackages)
	builder.WriteString(", ")
	builder.WriteString("update_sources=")
	builder.WriteString(oup.UpdateSources)
	builder.WriteString(", ")
	builder.WriteString("kernel_command=")
	builder.WriteString(oup.KernelCommand)
	builder.WriteString(", ")
	builder.WriteString("update_policy=")
	builder.WriteString(fmt.Sprintf("%v", oup.UpdatePolicy))
	builder.WriteString(", ")
	builder.WriteString("tenant_id=")
	builder.WriteString(oup.TenantID)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(oup.CreatedAt)
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(oup.UpdatedAt)
	builder.WriteByte(')')
	return builder.String()
}

// OSUpdatePolicies is a parsable slice of OSUpdatePolicy.
type OSUpdatePolicies []*OSUpdatePolicy

// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/regionresource"
)

// RegionResource is the model entity for the RegionResource schema.
type RegionResource struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// ResourceID holds the value of the "resource_id" field.
	ResourceID string `json:"resource_id,omitempty"`
	// Name holds the value of the "name" field.
	Name string `json:"name,omitempty"`
	// RegionKind holds the value of the "region_kind" field.
	RegionKind string `json:"region_kind,omitempty"`
	// Metadata holds the value of the "metadata" field.
	Metadata string `json:"metadata,omitempty"`
	// TenantID holds the value of the "tenant_id" field.
	TenantID string `json:"tenant_id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt string `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt string `json:"updated_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the RegionResourceQuery when eager-loading is set.
	Edges                         RegionResourceEdges `json:"edges"`
	region_resource_parent_region *int
	selectValues                  sql.SelectValues
}

// RegionResourceEdges holds the relations/edges for other nodes in the graph.
type RegionResourceEdges struct {
	// ParentRegion holds the value of the parent_region edge.
	ParentRegion *RegionResource `json:"parent_region,omitempty"`
	// Children holds the value of the children edge.
	Children []*RegionResource `json:"children,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
}

// ParentRegionOrErr returns the ParentRegion value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e RegionResourceEdges) ParentRegionOrErr() (*RegionResource, error) {
	if e.ParentRegion != nil {
		return e.ParentRegion, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: regionresource.Label}
	}
	return nil, &NotLoadedError{edge: "parent_region"}
}

// ChildrenOrErr returns the Children value or an error if the edge
// was not loaded in eager-loading.
func (e RegionResourceEdges) ChildrenOrErr() ([]*RegionResource, error) {
	if e.loadedTypes[1] {
		return e.Children, nil
	}
	return nil, &NotLoadedError{edge: "children"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*RegionResource) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case regionresource.FieldID:
			values[i] = new(sql.NullInt64)
		case regionresource.FieldResourceID, regionresource.FieldName, regionresource.FieldRegionKind, regionresource.FieldMetadata, regionresource.FieldTenantID, regionresource.FieldCreatedAt, regionresource.FieldUpdatedAt:
			values[i] = new(sql.NullString)
		case regionresource.ForeignKeys[0]: // region_resource_parent_region
			values[i] = new(sql.NullInt64)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the RegionResource fields.
func (rr *RegionResource) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case regionresource.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			rr.ID = int(value.Int64)
		case regionresource.FieldResourceID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field resource_id", values[i])
			} else if value.Valid {
				rr.ResourceID = value.String
			}
		case regionresource.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				rr.Name = value.String
			}
		case regionresource.FieldRegionKind:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field region_kind", values[i])
			} else if value.Valid {
				rr.RegionKind = value.String
			}
		case regionresource.FieldMetadata:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field metadata", values[i])
			} else if value.Valid {
				rr.Metadata = value.String
			}
		case regionresource.FieldTenantID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field tenant_id", values[i])
			} else if value.Valid {
				rr.TenantID = value.String
			}
		case regionresource.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				rr.CreatedAt = value.String
			}
		case regionresource.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				rr.UpdatedAt = value.String
			}
		case regionresource.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for edge-field region_resource_parent_region", value)
			} else if value.Valid {
				rr.region_resource_parent_region = new(int)
				*rr.region_resource_parent_region = int(value.Int64)
			}
		default:
			rr.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the RegionResource.
// This includes values selected through modifiers, order, etc.
func (rr *RegionResource) Value(name string) (ent.Value, error) {
	return rr.selectValues.Get(name)
}

// QueryParentRegion queries the "parent_region" edge of the RegionResource entity.
func (rr *RegionResource) QueryParentRegion() *RegionResourceQuery {
	return NewRegionResourceClient(rr.config).QueryParentRegion(rr)
}

// QueryChildren queries the "children" edge of the RegionResource entity.
func (rr *RegionResource) QueryChildren() *RegionResourceQuery {
	return NewRegionResourceClient(rr.config).QueryChildren(rr)
}

// Update returns a builder for updating this RegionResource.
// Note that you need to call RegionResource.Unwrap() before calling this method if this RegionResource
// was returned from a transaction, and the transaction was committed or rolled back.
func (rr *RegionResource) Update() *RegionResourceUpdateOne {
	return NewRegionResourceClient(rr.config).UpdateOne(rr)
}

// Unwrap unwraps the RegionResource entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (rr *RegionResource) Unwrap() *RegionResource {
	_tx, ok := rr.config.driver.(*txDriver)
	if !ok {
		panic("ent: RegionResource is not a transactional entity")
	}
	rr.config.driver = _tx.drv
	return rr
}

// String implements the fmt.Stringer.
func (rr *RegionResource) String() string {
	var builder strings.Builder
	builder.WriteString("RegionResource(")
	builder.WriteString(fmt.Sprintf("id=%v, ", rr.ID))
	builder.WriteString("resource_id=")
	builder.WriteString(rr.ResourceID)
	builder.WriteString(", ")
	builder.WriteString("name=")
	builder.WriteString(rr.Name)
	builder.WriteString(", ")
	builder.WriteString("region_kind=")
	builder.WriteString(rr.RegionKind)
	builder.WriteString(", ")
	builder.WriteString("metadata=")
	builder.WriteString(rr.Metadata)
	builder.WriteString(", ")
	builder.WriteString("tenant_id=")
	builder.WriteString(rr.TenantID)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(rr.CreatedAt)
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(rr.UpdatedAt)
	builder.WriteByte(')')
	return builder.String()
}

// RegionResources is a parsable slice of RegionResource.
type RegionResources []*RegionResource

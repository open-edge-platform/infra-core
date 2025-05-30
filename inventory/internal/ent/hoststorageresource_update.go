// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hoststorageresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/predicate"
)

// HoststorageResourceUpdate is the builder for updating HoststorageResource entities.
type HoststorageResourceUpdate struct {
	config
	hooks    []Hook
	mutation *HoststorageResourceMutation
}

// Where appends a list predicates to the HoststorageResourceUpdate builder.
func (hru *HoststorageResourceUpdate) Where(ps ...predicate.HoststorageResource) *HoststorageResourceUpdate {
	hru.mutation.Where(ps...)
	return hru
}

// SetResourceID sets the "resource_id" field.
func (hru *HoststorageResourceUpdate) SetResourceID(s string) *HoststorageResourceUpdate {
	hru.mutation.SetResourceID(s)
	return hru
}

// SetNillableResourceID sets the "resource_id" field if the given value is not nil.
func (hru *HoststorageResourceUpdate) SetNillableResourceID(s *string) *HoststorageResourceUpdate {
	if s != nil {
		hru.SetResourceID(*s)
	}
	return hru
}

// SetKind sets the "kind" field.
func (hru *HoststorageResourceUpdate) SetKind(s string) *HoststorageResourceUpdate {
	hru.mutation.SetKind(s)
	return hru
}

// SetNillableKind sets the "kind" field if the given value is not nil.
func (hru *HoststorageResourceUpdate) SetNillableKind(s *string) *HoststorageResourceUpdate {
	if s != nil {
		hru.SetKind(*s)
	}
	return hru
}

// ClearKind clears the value of the "kind" field.
func (hru *HoststorageResourceUpdate) ClearKind() *HoststorageResourceUpdate {
	hru.mutation.ClearKind()
	return hru
}

// SetProviderStatus sets the "provider_status" field.
func (hru *HoststorageResourceUpdate) SetProviderStatus(s string) *HoststorageResourceUpdate {
	hru.mutation.SetProviderStatus(s)
	return hru
}

// SetNillableProviderStatus sets the "provider_status" field if the given value is not nil.
func (hru *HoststorageResourceUpdate) SetNillableProviderStatus(s *string) *HoststorageResourceUpdate {
	if s != nil {
		hru.SetProviderStatus(*s)
	}
	return hru
}

// ClearProviderStatus clears the value of the "provider_status" field.
func (hru *HoststorageResourceUpdate) ClearProviderStatus() *HoststorageResourceUpdate {
	hru.mutation.ClearProviderStatus()
	return hru
}

// SetWwid sets the "wwid" field.
func (hru *HoststorageResourceUpdate) SetWwid(s string) *HoststorageResourceUpdate {
	hru.mutation.SetWwid(s)
	return hru
}

// SetNillableWwid sets the "wwid" field if the given value is not nil.
func (hru *HoststorageResourceUpdate) SetNillableWwid(s *string) *HoststorageResourceUpdate {
	if s != nil {
		hru.SetWwid(*s)
	}
	return hru
}

// ClearWwid clears the value of the "wwid" field.
func (hru *HoststorageResourceUpdate) ClearWwid() *HoststorageResourceUpdate {
	hru.mutation.ClearWwid()
	return hru
}

// SetSerial sets the "serial" field.
func (hru *HoststorageResourceUpdate) SetSerial(s string) *HoststorageResourceUpdate {
	hru.mutation.SetSerial(s)
	return hru
}

// SetNillableSerial sets the "serial" field if the given value is not nil.
func (hru *HoststorageResourceUpdate) SetNillableSerial(s *string) *HoststorageResourceUpdate {
	if s != nil {
		hru.SetSerial(*s)
	}
	return hru
}

// ClearSerial clears the value of the "serial" field.
func (hru *HoststorageResourceUpdate) ClearSerial() *HoststorageResourceUpdate {
	hru.mutation.ClearSerial()
	return hru
}

// SetVendor sets the "vendor" field.
func (hru *HoststorageResourceUpdate) SetVendor(s string) *HoststorageResourceUpdate {
	hru.mutation.SetVendor(s)
	return hru
}

// SetNillableVendor sets the "vendor" field if the given value is not nil.
func (hru *HoststorageResourceUpdate) SetNillableVendor(s *string) *HoststorageResourceUpdate {
	if s != nil {
		hru.SetVendor(*s)
	}
	return hru
}

// ClearVendor clears the value of the "vendor" field.
func (hru *HoststorageResourceUpdate) ClearVendor() *HoststorageResourceUpdate {
	hru.mutation.ClearVendor()
	return hru
}

// SetModel sets the "model" field.
func (hru *HoststorageResourceUpdate) SetModel(s string) *HoststorageResourceUpdate {
	hru.mutation.SetModel(s)
	return hru
}

// SetNillableModel sets the "model" field if the given value is not nil.
func (hru *HoststorageResourceUpdate) SetNillableModel(s *string) *HoststorageResourceUpdate {
	if s != nil {
		hru.SetModel(*s)
	}
	return hru
}

// ClearModel clears the value of the "model" field.
func (hru *HoststorageResourceUpdate) ClearModel() *HoststorageResourceUpdate {
	hru.mutation.ClearModel()
	return hru
}

// SetCapacityBytes sets the "capacity_bytes" field.
func (hru *HoststorageResourceUpdate) SetCapacityBytes(u uint64) *HoststorageResourceUpdate {
	hru.mutation.ResetCapacityBytes()
	hru.mutation.SetCapacityBytes(u)
	return hru
}

// SetNillableCapacityBytes sets the "capacity_bytes" field if the given value is not nil.
func (hru *HoststorageResourceUpdate) SetNillableCapacityBytes(u *uint64) *HoststorageResourceUpdate {
	if u != nil {
		hru.SetCapacityBytes(*u)
	}
	return hru
}

// AddCapacityBytes adds u to the "capacity_bytes" field.
func (hru *HoststorageResourceUpdate) AddCapacityBytes(u int64) *HoststorageResourceUpdate {
	hru.mutation.AddCapacityBytes(u)
	return hru
}

// ClearCapacityBytes clears the value of the "capacity_bytes" field.
func (hru *HoststorageResourceUpdate) ClearCapacityBytes() *HoststorageResourceUpdate {
	hru.mutation.ClearCapacityBytes()
	return hru
}

// SetDeviceName sets the "device_name" field.
func (hru *HoststorageResourceUpdate) SetDeviceName(s string) *HoststorageResourceUpdate {
	hru.mutation.SetDeviceName(s)
	return hru
}

// SetNillableDeviceName sets the "device_name" field if the given value is not nil.
func (hru *HoststorageResourceUpdate) SetNillableDeviceName(s *string) *HoststorageResourceUpdate {
	if s != nil {
		hru.SetDeviceName(*s)
	}
	return hru
}

// ClearDeviceName clears the value of the "device_name" field.
func (hru *HoststorageResourceUpdate) ClearDeviceName() *HoststorageResourceUpdate {
	hru.mutation.ClearDeviceName()
	return hru
}

// SetUpdatedAt sets the "updated_at" field.
func (hru *HoststorageResourceUpdate) SetUpdatedAt(s string) *HoststorageResourceUpdate {
	hru.mutation.SetUpdatedAt(s)
	return hru
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (hru *HoststorageResourceUpdate) SetNillableUpdatedAt(s *string) *HoststorageResourceUpdate {
	if s != nil {
		hru.SetUpdatedAt(*s)
	}
	return hru
}

// SetHostID sets the "host" edge to the HostResource entity by ID.
func (hru *HoststorageResourceUpdate) SetHostID(id int) *HoststorageResourceUpdate {
	hru.mutation.SetHostID(id)
	return hru
}

// SetHost sets the "host" edge to the HostResource entity.
func (hru *HoststorageResourceUpdate) SetHost(h *HostResource) *HoststorageResourceUpdate {
	return hru.SetHostID(h.ID)
}

// Mutation returns the HoststorageResourceMutation object of the builder.
func (hru *HoststorageResourceUpdate) Mutation() *HoststorageResourceMutation {
	return hru.mutation
}

// ClearHost clears the "host" edge to the HostResource entity.
func (hru *HoststorageResourceUpdate) ClearHost() *HoststorageResourceUpdate {
	hru.mutation.ClearHost()
	return hru
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (hru *HoststorageResourceUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, hru.sqlSave, hru.mutation, hru.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (hru *HoststorageResourceUpdate) SaveX(ctx context.Context) int {
	affected, err := hru.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (hru *HoststorageResourceUpdate) Exec(ctx context.Context) error {
	_, err := hru.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (hru *HoststorageResourceUpdate) ExecX(ctx context.Context) {
	if err := hru.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (hru *HoststorageResourceUpdate) check() error {
	if hru.mutation.HostCleared() && len(hru.mutation.HostIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "HoststorageResource.host"`)
	}
	return nil
}

func (hru *HoststorageResourceUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := hru.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(hoststorageresource.Table, hoststorageresource.Columns, sqlgraph.NewFieldSpec(hoststorageresource.FieldID, field.TypeInt))
	if ps := hru.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := hru.mutation.ResourceID(); ok {
		_spec.SetField(hoststorageresource.FieldResourceID, field.TypeString, value)
	}
	if value, ok := hru.mutation.Kind(); ok {
		_spec.SetField(hoststorageresource.FieldKind, field.TypeString, value)
	}
	if hru.mutation.KindCleared() {
		_spec.ClearField(hoststorageresource.FieldKind, field.TypeString)
	}
	if value, ok := hru.mutation.ProviderStatus(); ok {
		_spec.SetField(hoststorageresource.FieldProviderStatus, field.TypeString, value)
	}
	if hru.mutation.ProviderStatusCleared() {
		_spec.ClearField(hoststorageresource.FieldProviderStatus, field.TypeString)
	}
	if value, ok := hru.mutation.Wwid(); ok {
		_spec.SetField(hoststorageresource.FieldWwid, field.TypeString, value)
	}
	if hru.mutation.WwidCleared() {
		_spec.ClearField(hoststorageresource.FieldWwid, field.TypeString)
	}
	if value, ok := hru.mutation.Serial(); ok {
		_spec.SetField(hoststorageresource.FieldSerial, field.TypeString, value)
	}
	if hru.mutation.SerialCleared() {
		_spec.ClearField(hoststorageresource.FieldSerial, field.TypeString)
	}
	if value, ok := hru.mutation.Vendor(); ok {
		_spec.SetField(hoststorageresource.FieldVendor, field.TypeString, value)
	}
	if hru.mutation.VendorCleared() {
		_spec.ClearField(hoststorageresource.FieldVendor, field.TypeString)
	}
	if value, ok := hru.mutation.Model(); ok {
		_spec.SetField(hoststorageresource.FieldModel, field.TypeString, value)
	}
	if hru.mutation.ModelCleared() {
		_spec.ClearField(hoststorageresource.FieldModel, field.TypeString)
	}
	if value, ok := hru.mutation.CapacityBytes(); ok {
		_spec.SetField(hoststorageresource.FieldCapacityBytes, field.TypeUint64, value)
	}
	if value, ok := hru.mutation.AddedCapacityBytes(); ok {
		_spec.AddField(hoststorageresource.FieldCapacityBytes, field.TypeUint64, value)
	}
	if hru.mutation.CapacityBytesCleared() {
		_spec.ClearField(hoststorageresource.FieldCapacityBytes, field.TypeUint64)
	}
	if value, ok := hru.mutation.DeviceName(); ok {
		_spec.SetField(hoststorageresource.FieldDeviceName, field.TypeString, value)
	}
	if hru.mutation.DeviceNameCleared() {
		_spec.ClearField(hoststorageresource.FieldDeviceName, field.TypeString)
	}
	if value, ok := hru.mutation.UpdatedAt(); ok {
		_spec.SetField(hoststorageresource.FieldUpdatedAt, field.TypeString, value)
	}
	if hru.mutation.HostCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   hoststorageresource.HostTable,
			Columns: []string{hoststorageresource.HostColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(hostresource.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := hru.mutation.HostIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   hoststorageresource.HostTable,
			Columns: []string{hoststorageresource.HostColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(hostresource.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, hru.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{hoststorageresource.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	hru.mutation.done = true
	return n, nil
}

// HoststorageResourceUpdateOne is the builder for updating a single HoststorageResource entity.
type HoststorageResourceUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *HoststorageResourceMutation
}

// SetResourceID sets the "resource_id" field.
func (hruo *HoststorageResourceUpdateOne) SetResourceID(s string) *HoststorageResourceUpdateOne {
	hruo.mutation.SetResourceID(s)
	return hruo
}

// SetNillableResourceID sets the "resource_id" field if the given value is not nil.
func (hruo *HoststorageResourceUpdateOne) SetNillableResourceID(s *string) *HoststorageResourceUpdateOne {
	if s != nil {
		hruo.SetResourceID(*s)
	}
	return hruo
}

// SetKind sets the "kind" field.
func (hruo *HoststorageResourceUpdateOne) SetKind(s string) *HoststorageResourceUpdateOne {
	hruo.mutation.SetKind(s)
	return hruo
}

// SetNillableKind sets the "kind" field if the given value is not nil.
func (hruo *HoststorageResourceUpdateOne) SetNillableKind(s *string) *HoststorageResourceUpdateOne {
	if s != nil {
		hruo.SetKind(*s)
	}
	return hruo
}

// ClearKind clears the value of the "kind" field.
func (hruo *HoststorageResourceUpdateOne) ClearKind() *HoststorageResourceUpdateOne {
	hruo.mutation.ClearKind()
	return hruo
}

// SetProviderStatus sets the "provider_status" field.
func (hruo *HoststorageResourceUpdateOne) SetProviderStatus(s string) *HoststorageResourceUpdateOne {
	hruo.mutation.SetProviderStatus(s)
	return hruo
}

// SetNillableProviderStatus sets the "provider_status" field if the given value is not nil.
func (hruo *HoststorageResourceUpdateOne) SetNillableProviderStatus(s *string) *HoststorageResourceUpdateOne {
	if s != nil {
		hruo.SetProviderStatus(*s)
	}
	return hruo
}

// ClearProviderStatus clears the value of the "provider_status" field.
func (hruo *HoststorageResourceUpdateOne) ClearProviderStatus() *HoststorageResourceUpdateOne {
	hruo.mutation.ClearProviderStatus()
	return hruo
}

// SetWwid sets the "wwid" field.
func (hruo *HoststorageResourceUpdateOne) SetWwid(s string) *HoststorageResourceUpdateOne {
	hruo.mutation.SetWwid(s)
	return hruo
}

// SetNillableWwid sets the "wwid" field if the given value is not nil.
func (hruo *HoststorageResourceUpdateOne) SetNillableWwid(s *string) *HoststorageResourceUpdateOne {
	if s != nil {
		hruo.SetWwid(*s)
	}
	return hruo
}

// ClearWwid clears the value of the "wwid" field.
func (hruo *HoststorageResourceUpdateOne) ClearWwid() *HoststorageResourceUpdateOne {
	hruo.mutation.ClearWwid()
	return hruo
}

// SetSerial sets the "serial" field.
func (hruo *HoststorageResourceUpdateOne) SetSerial(s string) *HoststorageResourceUpdateOne {
	hruo.mutation.SetSerial(s)
	return hruo
}

// SetNillableSerial sets the "serial" field if the given value is not nil.
func (hruo *HoststorageResourceUpdateOne) SetNillableSerial(s *string) *HoststorageResourceUpdateOne {
	if s != nil {
		hruo.SetSerial(*s)
	}
	return hruo
}

// ClearSerial clears the value of the "serial" field.
func (hruo *HoststorageResourceUpdateOne) ClearSerial() *HoststorageResourceUpdateOne {
	hruo.mutation.ClearSerial()
	return hruo
}

// SetVendor sets the "vendor" field.
func (hruo *HoststorageResourceUpdateOne) SetVendor(s string) *HoststorageResourceUpdateOne {
	hruo.mutation.SetVendor(s)
	return hruo
}

// SetNillableVendor sets the "vendor" field if the given value is not nil.
func (hruo *HoststorageResourceUpdateOne) SetNillableVendor(s *string) *HoststorageResourceUpdateOne {
	if s != nil {
		hruo.SetVendor(*s)
	}
	return hruo
}

// ClearVendor clears the value of the "vendor" field.
func (hruo *HoststorageResourceUpdateOne) ClearVendor() *HoststorageResourceUpdateOne {
	hruo.mutation.ClearVendor()
	return hruo
}

// SetModel sets the "model" field.
func (hruo *HoststorageResourceUpdateOne) SetModel(s string) *HoststorageResourceUpdateOne {
	hruo.mutation.SetModel(s)
	return hruo
}

// SetNillableModel sets the "model" field if the given value is not nil.
func (hruo *HoststorageResourceUpdateOne) SetNillableModel(s *string) *HoststorageResourceUpdateOne {
	if s != nil {
		hruo.SetModel(*s)
	}
	return hruo
}

// ClearModel clears the value of the "model" field.
func (hruo *HoststorageResourceUpdateOne) ClearModel() *HoststorageResourceUpdateOne {
	hruo.mutation.ClearModel()
	return hruo
}

// SetCapacityBytes sets the "capacity_bytes" field.
func (hruo *HoststorageResourceUpdateOne) SetCapacityBytes(u uint64) *HoststorageResourceUpdateOne {
	hruo.mutation.ResetCapacityBytes()
	hruo.mutation.SetCapacityBytes(u)
	return hruo
}

// SetNillableCapacityBytes sets the "capacity_bytes" field if the given value is not nil.
func (hruo *HoststorageResourceUpdateOne) SetNillableCapacityBytes(u *uint64) *HoststorageResourceUpdateOne {
	if u != nil {
		hruo.SetCapacityBytes(*u)
	}
	return hruo
}

// AddCapacityBytes adds u to the "capacity_bytes" field.
func (hruo *HoststorageResourceUpdateOne) AddCapacityBytes(u int64) *HoststorageResourceUpdateOne {
	hruo.mutation.AddCapacityBytes(u)
	return hruo
}

// ClearCapacityBytes clears the value of the "capacity_bytes" field.
func (hruo *HoststorageResourceUpdateOne) ClearCapacityBytes() *HoststorageResourceUpdateOne {
	hruo.mutation.ClearCapacityBytes()
	return hruo
}

// SetDeviceName sets the "device_name" field.
func (hruo *HoststorageResourceUpdateOne) SetDeviceName(s string) *HoststorageResourceUpdateOne {
	hruo.mutation.SetDeviceName(s)
	return hruo
}

// SetNillableDeviceName sets the "device_name" field if the given value is not nil.
func (hruo *HoststorageResourceUpdateOne) SetNillableDeviceName(s *string) *HoststorageResourceUpdateOne {
	if s != nil {
		hruo.SetDeviceName(*s)
	}
	return hruo
}

// ClearDeviceName clears the value of the "device_name" field.
func (hruo *HoststorageResourceUpdateOne) ClearDeviceName() *HoststorageResourceUpdateOne {
	hruo.mutation.ClearDeviceName()
	return hruo
}

// SetUpdatedAt sets the "updated_at" field.
func (hruo *HoststorageResourceUpdateOne) SetUpdatedAt(s string) *HoststorageResourceUpdateOne {
	hruo.mutation.SetUpdatedAt(s)
	return hruo
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (hruo *HoststorageResourceUpdateOne) SetNillableUpdatedAt(s *string) *HoststorageResourceUpdateOne {
	if s != nil {
		hruo.SetUpdatedAt(*s)
	}
	return hruo
}

// SetHostID sets the "host" edge to the HostResource entity by ID.
func (hruo *HoststorageResourceUpdateOne) SetHostID(id int) *HoststorageResourceUpdateOne {
	hruo.mutation.SetHostID(id)
	return hruo
}

// SetHost sets the "host" edge to the HostResource entity.
func (hruo *HoststorageResourceUpdateOne) SetHost(h *HostResource) *HoststorageResourceUpdateOne {
	return hruo.SetHostID(h.ID)
}

// Mutation returns the HoststorageResourceMutation object of the builder.
func (hruo *HoststorageResourceUpdateOne) Mutation() *HoststorageResourceMutation {
	return hruo.mutation
}

// ClearHost clears the "host" edge to the HostResource entity.
func (hruo *HoststorageResourceUpdateOne) ClearHost() *HoststorageResourceUpdateOne {
	hruo.mutation.ClearHost()
	return hruo
}

// Where appends a list predicates to the HoststorageResourceUpdate builder.
func (hruo *HoststorageResourceUpdateOne) Where(ps ...predicate.HoststorageResource) *HoststorageResourceUpdateOne {
	hruo.mutation.Where(ps...)
	return hruo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (hruo *HoststorageResourceUpdateOne) Select(field string, fields ...string) *HoststorageResourceUpdateOne {
	hruo.fields = append([]string{field}, fields...)
	return hruo
}

// Save executes the query and returns the updated HoststorageResource entity.
func (hruo *HoststorageResourceUpdateOne) Save(ctx context.Context) (*HoststorageResource, error) {
	return withHooks(ctx, hruo.sqlSave, hruo.mutation, hruo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (hruo *HoststorageResourceUpdateOne) SaveX(ctx context.Context) *HoststorageResource {
	node, err := hruo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (hruo *HoststorageResourceUpdateOne) Exec(ctx context.Context) error {
	_, err := hruo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (hruo *HoststorageResourceUpdateOne) ExecX(ctx context.Context) {
	if err := hruo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (hruo *HoststorageResourceUpdateOne) check() error {
	if hruo.mutation.HostCleared() && len(hruo.mutation.HostIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "HoststorageResource.host"`)
	}
	return nil
}

func (hruo *HoststorageResourceUpdateOne) sqlSave(ctx context.Context) (_node *HoststorageResource, err error) {
	if err := hruo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(hoststorageresource.Table, hoststorageresource.Columns, sqlgraph.NewFieldSpec(hoststorageresource.FieldID, field.TypeInt))
	id, ok := hruo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "HoststorageResource.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := hruo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, hoststorageresource.FieldID)
		for _, f := range fields {
			if !hoststorageresource.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != hoststorageresource.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := hruo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := hruo.mutation.ResourceID(); ok {
		_spec.SetField(hoststorageresource.FieldResourceID, field.TypeString, value)
	}
	if value, ok := hruo.mutation.Kind(); ok {
		_spec.SetField(hoststorageresource.FieldKind, field.TypeString, value)
	}
	if hruo.mutation.KindCleared() {
		_spec.ClearField(hoststorageresource.FieldKind, field.TypeString)
	}
	if value, ok := hruo.mutation.ProviderStatus(); ok {
		_spec.SetField(hoststorageresource.FieldProviderStatus, field.TypeString, value)
	}
	if hruo.mutation.ProviderStatusCleared() {
		_spec.ClearField(hoststorageresource.FieldProviderStatus, field.TypeString)
	}
	if value, ok := hruo.mutation.Wwid(); ok {
		_spec.SetField(hoststorageresource.FieldWwid, field.TypeString, value)
	}
	if hruo.mutation.WwidCleared() {
		_spec.ClearField(hoststorageresource.FieldWwid, field.TypeString)
	}
	if value, ok := hruo.mutation.Serial(); ok {
		_spec.SetField(hoststorageresource.FieldSerial, field.TypeString, value)
	}
	if hruo.mutation.SerialCleared() {
		_spec.ClearField(hoststorageresource.FieldSerial, field.TypeString)
	}
	if value, ok := hruo.mutation.Vendor(); ok {
		_spec.SetField(hoststorageresource.FieldVendor, field.TypeString, value)
	}
	if hruo.mutation.VendorCleared() {
		_spec.ClearField(hoststorageresource.FieldVendor, field.TypeString)
	}
	if value, ok := hruo.mutation.Model(); ok {
		_spec.SetField(hoststorageresource.FieldModel, field.TypeString, value)
	}
	if hruo.mutation.ModelCleared() {
		_spec.ClearField(hoststorageresource.FieldModel, field.TypeString)
	}
	if value, ok := hruo.mutation.CapacityBytes(); ok {
		_spec.SetField(hoststorageresource.FieldCapacityBytes, field.TypeUint64, value)
	}
	if value, ok := hruo.mutation.AddedCapacityBytes(); ok {
		_spec.AddField(hoststorageresource.FieldCapacityBytes, field.TypeUint64, value)
	}
	if hruo.mutation.CapacityBytesCleared() {
		_spec.ClearField(hoststorageresource.FieldCapacityBytes, field.TypeUint64)
	}
	if value, ok := hruo.mutation.DeviceName(); ok {
		_spec.SetField(hoststorageresource.FieldDeviceName, field.TypeString, value)
	}
	if hruo.mutation.DeviceNameCleared() {
		_spec.ClearField(hoststorageresource.FieldDeviceName, field.TypeString)
	}
	if value, ok := hruo.mutation.UpdatedAt(); ok {
		_spec.SetField(hoststorageresource.FieldUpdatedAt, field.TypeString, value)
	}
	if hruo.mutation.HostCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   hoststorageresource.HostTable,
			Columns: []string{hoststorageresource.HostColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(hostresource.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := hruo.mutation.HostIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   hoststorageresource.HostTable,
			Columns: []string{hoststorageresource.HostColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(hostresource.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &HoststorageResource{config: hruo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, hruo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{hoststorageresource.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	hruo.mutation.done = true
	return _node, nil
}

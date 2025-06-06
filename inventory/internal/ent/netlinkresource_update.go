// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/endpointresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/netlinkresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/predicate"
)

// NetlinkResourceUpdate is the builder for updating NetlinkResource entities.
type NetlinkResourceUpdate struct {
	config
	hooks    []Hook
	mutation *NetlinkResourceMutation
}

// Where appends a list predicates to the NetlinkResourceUpdate builder.
func (nru *NetlinkResourceUpdate) Where(ps ...predicate.NetlinkResource) *NetlinkResourceUpdate {
	nru.mutation.Where(ps...)
	return nru
}

// SetResourceID sets the "resource_id" field.
func (nru *NetlinkResourceUpdate) SetResourceID(s string) *NetlinkResourceUpdate {
	nru.mutation.SetResourceID(s)
	return nru
}

// SetNillableResourceID sets the "resource_id" field if the given value is not nil.
func (nru *NetlinkResourceUpdate) SetNillableResourceID(s *string) *NetlinkResourceUpdate {
	if s != nil {
		nru.SetResourceID(*s)
	}
	return nru
}

// SetKind sets the "kind" field.
func (nru *NetlinkResourceUpdate) SetKind(s string) *NetlinkResourceUpdate {
	nru.mutation.SetKind(s)
	return nru
}

// SetNillableKind sets the "kind" field if the given value is not nil.
func (nru *NetlinkResourceUpdate) SetNillableKind(s *string) *NetlinkResourceUpdate {
	if s != nil {
		nru.SetKind(*s)
	}
	return nru
}

// ClearKind clears the value of the "kind" field.
func (nru *NetlinkResourceUpdate) ClearKind() *NetlinkResourceUpdate {
	nru.mutation.ClearKind()
	return nru
}

// SetName sets the "name" field.
func (nru *NetlinkResourceUpdate) SetName(s string) *NetlinkResourceUpdate {
	nru.mutation.SetName(s)
	return nru
}

// SetNillableName sets the "name" field if the given value is not nil.
func (nru *NetlinkResourceUpdate) SetNillableName(s *string) *NetlinkResourceUpdate {
	if s != nil {
		nru.SetName(*s)
	}
	return nru
}

// ClearName clears the value of the "name" field.
func (nru *NetlinkResourceUpdate) ClearName() *NetlinkResourceUpdate {
	nru.mutation.ClearName()
	return nru
}

// SetDesiredState sets the "desired_state" field.
func (nru *NetlinkResourceUpdate) SetDesiredState(ns netlinkresource.DesiredState) *NetlinkResourceUpdate {
	nru.mutation.SetDesiredState(ns)
	return nru
}

// SetNillableDesiredState sets the "desired_state" field if the given value is not nil.
func (nru *NetlinkResourceUpdate) SetNillableDesiredState(ns *netlinkresource.DesiredState) *NetlinkResourceUpdate {
	if ns != nil {
		nru.SetDesiredState(*ns)
	}
	return nru
}

// SetCurrentState sets the "current_state" field.
func (nru *NetlinkResourceUpdate) SetCurrentState(ns netlinkresource.CurrentState) *NetlinkResourceUpdate {
	nru.mutation.SetCurrentState(ns)
	return nru
}

// SetNillableCurrentState sets the "current_state" field if the given value is not nil.
func (nru *NetlinkResourceUpdate) SetNillableCurrentState(ns *netlinkresource.CurrentState) *NetlinkResourceUpdate {
	if ns != nil {
		nru.SetCurrentState(*ns)
	}
	return nru
}

// ClearCurrentState clears the value of the "current_state" field.
func (nru *NetlinkResourceUpdate) ClearCurrentState() *NetlinkResourceUpdate {
	nru.mutation.ClearCurrentState()
	return nru
}

// SetProviderStatus sets the "provider_status" field.
func (nru *NetlinkResourceUpdate) SetProviderStatus(s string) *NetlinkResourceUpdate {
	nru.mutation.SetProviderStatus(s)
	return nru
}

// SetNillableProviderStatus sets the "provider_status" field if the given value is not nil.
func (nru *NetlinkResourceUpdate) SetNillableProviderStatus(s *string) *NetlinkResourceUpdate {
	if s != nil {
		nru.SetProviderStatus(*s)
	}
	return nru
}

// ClearProviderStatus clears the value of the "provider_status" field.
func (nru *NetlinkResourceUpdate) ClearProviderStatus() *NetlinkResourceUpdate {
	nru.mutation.ClearProviderStatus()
	return nru
}

// SetUpdatedAt sets the "updated_at" field.
func (nru *NetlinkResourceUpdate) SetUpdatedAt(s string) *NetlinkResourceUpdate {
	nru.mutation.SetUpdatedAt(s)
	return nru
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (nru *NetlinkResourceUpdate) SetNillableUpdatedAt(s *string) *NetlinkResourceUpdate {
	if s != nil {
		nru.SetUpdatedAt(*s)
	}
	return nru
}

// SetSrcID sets the "src" edge to the EndpointResource entity by ID.
func (nru *NetlinkResourceUpdate) SetSrcID(id int) *NetlinkResourceUpdate {
	nru.mutation.SetSrcID(id)
	return nru
}

// SetNillableSrcID sets the "src" edge to the EndpointResource entity by ID if the given value is not nil.
func (nru *NetlinkResourceUpdate) SetNillableSrcID(id *int) *NetlinkResourceUpdate {
	if id != nil {
		nru = nru.SetSrcID(*id)
	}
	return nru
}

// SetSrc sets the "src" edge to the EndpointResource entity.
func (nru *NetlinkResourceUpdate) SetSrc(e *EndpointResource) *NetlinkResourceUpdate {
	return nru.SetSrcID(e.ID)
}

// SetDstID sets the "dst" edge to the EndpointResource entity by ID.
func (nru *NetlinkResourceUpdate) SetDstID(id int) *NetlinkResourceUpdate {
	nru.mutation.SetDstID(id)
	return nru
}

// SetNillableDstID sets the "dst" edge to the EndpointResource entity by ID if the given value is not nil.
func (nru *NetlinkResourceUpdate) SetNillableDstID(id *int) *NetlinkResourceUpdate {
	if id != nil {
		nru = nru.SetDstID(*id)
	}
	return nru
}

// SetDst sets the "dst" edge to the EndpointResource entity.
func (nru *NetlinkResourceUpdate) SetDst(e *EndpointResource) *NetlinkResourceUpdate {
	return nru.SetDstID(e.ID)
}

// Mutation returns the NetlinkResourceMutation object of the builder.
func (nru *NetlinkResourceUpdate) Mutation() *NetlinkResourceMutation {
	return nru.mutation
}

// ClearSrc clears the "src" edge to the EndpointResource entity.
func (nru *NetlinkResourceUpdate) ClearSrc() *NetlinkResourceUpdate {
	nru.mutation.ClearSrc()
	return nru
}

// ClearDst clears the "dst" edge to the EndpointResource entity.
func (nru *NetlinkResourceUpdate) ClearDst() *NetlinkResourceUpdate {
	nru.mutation.ClearDst()
	return nru
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (nru *NetlinkResourceUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, nru.sqlSave, nru.mutation, nru.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (nru *NetlinkResourceUpdate) SaveX(ctx context.Context) int {
	affected, err := nru.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (nru *NetlinkResourceUpdate) Exec(ctx context.Context) error {
	_, err := nru.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (nru *NetlinkResourceUpdate) ExecX(ctx context.Context) {
	if err := nru.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (nru *NetlinkResourceUpdate) check() error {
	if v, ok := nru.mutation.DesiredState(); ok {
		if err := netlinkresource.DesiredStateValidator(v); err != nil {
			return &ValidationError{Name: "desired_state", err: fmt.Errorf(`ent: validator failed for field "NetlinkResource.desired_state": %w`, err)}
		}
	}
	if v, ok := nru.mutation.CurrentState(); ok {
		if err := netlinkresource.CurrentStateValidator(v); err != nil {
			return &ValidationError{Name: "current_state", err: fmt.Errorf(`ent: validator failed for field "NetlinkResource.current_state": %w`, err)}
		}
	}
	return nil
}

func (nru *NetlinkResourceUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := nru.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(netlinkresource.Table, netlinkresource.Columns, sqlgraph.NewFieldSpec(netlinkresource.FieldID, field.TypeInt))
	if ps := nru.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := nru.mutation.ResourceID(); ok {
		_spec.SetField(netlinkresource.FieldResourceID, field.TypeString, value)
	}
	if value, ok := nru.mutation.Kind(); ok {
		_spec.SetField(netlinkresource.FieldKind, field.TypeString, value)
	}
	if nru.mutation.KindCleared() {
		_spec.ClearField(netlinkresource.FieldKind, field.TypeString)
	}
	if value, ok := nru.mutation.Name(); ok {
		_spec.SetField(netlinkresource.FieldName, field.TypeString, value)
	}
	if nru.mutation.NameCleared() {
		_spec.ClearField(netlinkresource.FieldName, field.TypeString)
	}
	if value, ok := nru.mutation.DesiredState(); ok {
		_spec.SetField(netlinkresource.FieldDesiredState, field.TypeEnum, value)
	}
	if value, ok := nru.mutation.CurrentState(); ok {
		_spec.SetField(netlinkresource.FieldCurrentState, field.TypeEnum, value)
	}
	if nru.mutation.CurrentStateCleared() {
		_spec.ClearField(netlinkresource.FieldCurrentState, field.TypeEnum)
	}
	if value, ok := nru.mutation.ProviderStatus(); ok {
		_spec.SetField(netlinkresource.FieldProviderStatus, field.TypeString, value)
	}
	if nru.mutation.ProviderStatusCleared() {
		_spec.ClearField(netlinkresource.FieldProviderStatus, field.TypeString)
	}
	if value, ok := nru.mutation.UpdatedAt(); ok {
		_spec.SetField(netlinkresource.FieldUpdatedAt, field.TypeString, value)
	}
	if nru.mutation.SrcCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   netlinkresource.SrcTable,
			Columns: []string{netlinkresource.SrcColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(endpointresource.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := nru.mutation.SrcIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   netlinkresource.SrcTable,
			Columns: []string{netlinkresource.SrcColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(endpointresource.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if nru.mutation.DstCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   netlinkresource.DstTable,
			Columns: []string{netlinkresource.DstColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(endpointresource.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := nru.mutation.DstIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   netlinkresource.DstTable,
			Columns: []string{netlinkresource.DstColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(endpointresource.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, nru.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{netlinkresource.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	nru.mutation.done = true
	return n, nil
}

// NetlinkResourceUpdateOne is the builder for updating a single NetlinkResource entity.
type NetlinkResourceUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *NetlinkResourceMutation
}

// SetResourceID sets the "resource_id" field.
func (nruo *NetlinkResourceUpdateOne) SetResourceID(s string) *NetlinkResourceUpdateOne {
	nruo.mutation.SetResourceID(s)
	return nruo
}

// SetNillableResourceID sets the "resource_id" field if the given value is not nil.
func (nruo *NetlinkResourceUpdateOne) SetNillableResourceID(s *string) *NetlinkResourceUpdateOne {
	if s != nil {
		nruo.SetResourceID(*s)
	}
	return nruo
}

// SetKind sets the "kind" field.
func (nruo *NetlinkResourceUpdateOne) SetKind(s string) *NetlinkResourceUpdateOne {
	nruo.mutation.SetKind(s)
	return nruo
}

// SetNillableKind sets the "kind" field if the given value is not nil.
func (nruo *NetlinkResourceUpdateOne) SetNillableKind(s *string) *NetlinkResourceUpdateOne {
	if s != nil {
		nruo.SetKind(*s)
	}
	return nruo
}

// ClearKind clears the value of the "kind" field.
func (nruo *NetlinkResourceUpdateOne) ClearKind() *NetlinkResourceUpdateOne {
	nruo.mutation.ClearKind()
	return nruo
}

// SetName sets the "name" field.
func (nruo *NetlinkResourceUpdateOne) SetName(s string) *NetlinkResourceUpdateOne {
	nruo.mutation.SetName(s)
	return nruo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (nruo *NetlinkResourceUpdateOne) SetNillableName(s *string) *NetlinkResourceUpdateOne {
	if s != nil {
		nruo.SetName(*s)
	}
	return nruo
}

// ClearName clears the value of the "name" field.
func (nruo *NetlinkResourceUpdateOne) ClearName() *NetlinkResourceUpdateOne {
	nruo.mutation.ClearName()
	return nruo
}

// SetDesiredState sets the "desired_state" field.
func (nruo *NetlinkResourceUpdateOne) SetDesiredState(ns netlinkresource.DesiredState) *NetlinkResourceUpdateOne {
	nruo.mutation.SetDesiredState(ns)
	return nruo
}

// SetNillableDesiredState sets the "desired_state" field if the given value is not nil.
func (nruo *NetlinkResourceUpdateOne) SetNillableDesiredState(ns *netlinkresource.DesiredState) *NetlinkResourceUpdateOne {
	if ns != nil {
		nruo.SetDesiredState(*ns)
	}
	return nruo
}

// SetCurrentState sets the "current_state" field.
func (nruo *NetlinkResourceUpdateOne) SetCurrentState(ns netlinkresource.CurrentState) *NetlinkResourceUpdateOne {
	nruo.mutation.SetCurrentState(ns)
	return nruo
}

// SetNillableCurrentState sets the "current_state" field if the given value is not nil.
func (nruo *NetlinkResourceUpdateOne) SetNillableCurrentState(ns *netlinkresource.CurrentState) *NetlinkResourceUpdateOne {
	if ns != nil {
		nruo.SetCurrentState(*ns)
	}
	return nruo
}

// ClearCurrentState clears the value of the "current_state" field.
func (nruo *NetlinkResourceUpdateOne) ClearCurrentState() *NetlinkResourceUpdateOne {
	nruo.mutation.ClearCurrentState()
	return nruo
}

// SetProviderStatus sets the "provider_status" field.
func (nruo *NetlinkResourceUpdateOne) SetProviderStatus(s string) *NetlinkResourceUpdateOne {
	nruo.mutation.SetProviderStatus(s)
	return nruo
}

// SetNillableProviderStatus sets the "provider_status" field if the given value is not nil.
func (nruo *NetlinkResourceUpdateOne) SetNillableProviderStatus(s *string) *NetlinkResourceUpdateOne {
	if s != nil {
		nruo.SetProviderStatus(*s)
	}
	return nruo
}

// ClearProviderStatus clears the value of the "provider_status" field.
func (nruo *NetlinkResourceUpdateOne) ClearProviderStatus() *NetlinkResourceUpdateOne {
	nruo.mutation.ClearProviderStatus()
	return nruo
}

// SetUpdatedAt sets the "updated_at" field.
func (nruo *NetlinkResourceUpdateOne) SetUpdatedAt(s string) *NetlinkResourceUpdateOne {
	nruo.mutation.SetUpdatedAt(s)
	return nruo
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (nruo *NetlinkResourceUpdateOne) SetNillableUpdatedAt(s *string) *NetlinkResourceUpdateOne {
	if s != nil {
		nruo.SetUpdatedAt(*s)
	}
	return nruo
}

// SetSrcID sets the "src" edge to the EndpointResource entity by ID.
func (nruo *NetlinkResourceUpdateOne) SetSrcID(id int) *NetlinkResourceUpdateOne {
	nruo.mutation.SetSrcID(id)
	return nruo
}

// SetNillableSrcID sets the "src" edge to the EndpointResource entity by ID if the given value is not nil.
func (nruo *NetlinkResourceUpdateOne) SetNillableSrcID(id *int) *NetlinkResourceUpdateOne {
	if id != nil {
		nruo = nruo.SetSrcID(*id)
	}
	return nruo
}

// SetSrc sets the "src" edge to the EndpointResource entity.
func (nruo *NetlinkResourceUpdateOne) SetSrc(e *EndpointResource) *NetlinkResourceUpdateOne {
	return nruo.SetSrcID(e.ID)
}

// SetDstID sets the "dst" edge to the EndpointResource entity by ID.
func (nruo *NetlinkResourceUpdateOne) SetDstID(id int) *NetlinkResourceUpdateOne {
	nruo.mutation.SetDstID(id)
	return nruo
}

// SetNillableDstID sets the "dst" edge to the EndpointResource entity by ID if the given value is not nil.
func (nruo *NetlinkResourceUpdateOne) SetNillableDstID(id *int) *NetlinkResourceUpdateOne {
	if id != nil {
		nruo = nruo.SetDstID(*id)
	}
	return nruo
}

// SetDst sets the "dst" edge to the EndpointResource entity.
func (nruo *NetlinkResourceUpdateOne) SetDst(e *EndpointResource) *NetlinkResourceUpdateOne {
	return nruo.SetDstID(e.ID)
}

// Mutation returns the NetlinkResourceMutation object of the builder.
func (nruo *NetlinkResourceUpdateOne) Mutation() *NetlinkResourceMutation {
	return nruo.mutation
}

// ClearSrc clears the "src" edge to the EndpointResource entity.
func (nruo *NetlinkResourceUpdateOne) ClearSrc() *NetlinkResourceUpdateOne {
	nruo.mutation.ClearSrc()
	return nruo
}

// ClearDst clears the "dst" edge to the EndpointResource entity.
func (nruo *NetlinkResourceUpdateOne) ClearDst() *NetlinkResourceUpdateOne {
	nruo.mutation.ClearDst()
	return nruo
}

// Where appends a list predicates to the NetlinkResourceUpdate builder.
func (nruo *NetlinkResourceUpdateOne) Where(ps ...predicate.NetlinkResource) *NetlinkResourceUpdateOne {
	nruo.mutation.Where(ps...)
	return nruo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (nruo *NetlinkResourceUpdateOne) Select(field string, fields ...string) *NetlinkResourceUpdateOne {
	nruo.fields = append([]string{field}, fields...)
	return nruo
}

// Save executes the query and returns the updated NetlinkResource entity.
func (nruo *NetlinkResourceUpdateOne) Save(ctx context.Context) (*NetlinkResource, error) {
	return withHooks(ctx, nruo.sqlSave, nruo.mutation, nruo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (nruo *NetlinkResourceUpdateOne) SaveX(ctx context.Context) *NetlinkResource {
	node, err := nruo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (nruo *NetlinkResourceUpdateOne) Exec(ctx context.Context) error {
	_, err := nruo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (nruo *NetlinkResourceUpdateOne) ExecX(ctx context.Context) {
	if err := nruo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (nruo *NetlinkResourceUpdateOne) check() error {
	if v, ok := nruo.mutation.DesiredState(); ok {
		if err := netlinkresource.DesiredStateValidator(v); err != nil {
			return &ValidationError{Name: "desired_state", err: fmt.Errorf(`ent: validator failed for field "NetlinkResource.desired_state": %w`, err)}
		}
	}
	if v, ok := nruo.mutation.CurrentState(); ok {
		if err := netlinkresource.CurrentStateValidator(v); err != nil {
			return &ValidationError{Name: "current_state", err: fmt.Errorf(`ent: validator failed for field "NetlinkResource.current_state": %w`, err)}
		}
	}
	return nil
}

func (nruo *NetlinkResourceUpdateOne) sqlSave(ctx context.Context) (_node *NetlinkResource, err error) {
	if err := nruo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(netlinkresource.Table, netlinkresource.Columns, sqlgraph.NewFieldSpec(netlinkresource.FieldID, field.TypeInt))
	id, ok := nruo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "NetlinkResource.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := nruo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, netlinkresource.FieldID)
		for _, f := range fields {
			if !netlinkresource.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != netlinkresource.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := nruo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := nruo.mutation.ResourceID(); ok {
		_spec.SetField(netlinkresource.FieldResourceID, field.TypeString, value)
	}
	if value, ok := nruo.mutation.Kind(); ok {
		_spec.SetField(netlinkresource.FieldKind, field.TypeString, value)
	}
	if nruo.mutation.KindCleared() {
		_spec.ClearField(netlinkresource.FieldKind, field.TypeString)
	}
	if value, ok := nruo.mutation.Name(); ok {
		_spec.SetField(netlinkresource.FieldName, field.TypeString, value)
	}
	if nruo.mutation.NameCleared() {
		_spec.ClearField(netlinkresource.FieldName, field.TypeString)
	}
	if value, ok := nruo.mutation.DesiredState(); ok {
		_spec.SetField(netlinkresource.FieldDesiredState, field.TypeEnum, value)
	}
	if value, ok := nruo.mutation.CurrentState(); ok {
		_spec.SetField(netlinkresource.FieldCurrentState, field.TypeEnum, value)
	}
	if nruo.mutation.CurrentStateCleared() {
		_spec.ClearField(netlinkresource.FieldCurrentState, field.TypeEnum)
	}
	if value, ok := nruo.mutation.ProviderStatus(); ok {
		_spec.SetField(netlinkresource.FieldProviderStatus, field.TypeString, value)
	}
	if nruo.mutation.ProviderStatusCleared() {
		_spec.ClearField(netlinkresource.FieldProviderStatus, field.TypeString)
	}
	if value, ok := nruo.mutation.UpdatedAt(); ok {
		_spec.SetField(netlinkresource.FieldUpdatedAt, field.TypeString, value)
	}
	if nruo.mutation.SrcCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   netlinkresource.SrcTable,
			Columns: []string{netlinkresource.SrcColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(endpointresource.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := nruo.mutation.SrcIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   netlinkresource.SrcTable,
			Columns: []string{netlinkresource.SrcColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(endpointresource.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if nruo.mutation.DstCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   netlinkresource.DstTable,
			Columns: []string{netlinkresource.DstColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(endpointresource.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := nruo.mutation.DstIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   netlinkresource.DstTable,
			Columns: []string{netlinkresource.DstColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(endpointresource.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &NetlinkResource{config: nruo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, nruo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{netlinkresource.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	nruo.mutation.done = true
	return _node, nil
}

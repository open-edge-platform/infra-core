// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/predicate"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/tenant"
)

// TenantUpdate is the builder for updating Tenant entities.
type TenantUpdate struct {
	config
	hooks    []Hook
	mutation *TenantMutation
}

// Where appends a list predicates to the TenantUpdate builder.
func (tu *TenantUpdate) Where(ps ...predicate.Tenant) *TenantUpdate {
	tu.mutation.Where(ps...)
	return tu
}

// SetResourceID sets the "resource_id" field.
func (tu *TenantUpdate) SetResourceID(s string) *TenantUpdate {
	tu.mutation.SetResourceID(s)
	return tu
}

// SetNillableResourceID sets the "resource_id" field if the given value is not nil.
func (tu *TenantUpdate) SetNillableResourceID(s *string) *TenantUpdate {
	if s != nil {
		tu.SetResourceID(*s)
	}
	return tu
}

// SetCurrentState sets the "current_state" field.
func (tu *TenantUpdate) SetCurrentState(ts tenant.CurrentState) *TenantUpdate {
	tu.mutation.SetCurrentState(ts)
	return tu
}

// SetNillableCurrentState sets the "current_state" field if the given value is not nil.
func (tu *TenantUpdate) SetNillableCurrentState(ts *tenant.CurrentState) *TenantUpdate {
	if ts != nil {
		tu.SetCurrentState(*ts)
	}
	return tu
}

// ClearCurrentState clears the value of the "current_state" field.
func (tu *TenantUpdate) ClearCurrentState() *TenantUpdate {
	tu.mutation.ClearCurrentState()
	return tu
}

// SetDesiredState sets the "desired_state" field.
func (tu *TenantUpdate) SetDesiredState(ts tenant.DesiredState) *TenantUpdate {
	tu.mutation.SetDesiredState(ts)
	return tu
}

// SetNillableDesiredState sets the "desired_state" field if the given value is not nil.
func (tu *TenantUpdate) SetNillableDesiredState(ts *tenant.DesiredState) *TenantUpdate {
	if ts != nil {
		tu.SetDesiredState(*ts)
	}
	return tu
}

// SetWatcherOsmanager sets the "watcher_osmanager" field.
func (tu *TenantUpdate) SetWatcherOsmanager(b bool) *TenantUpdate {
	tu.mutation.SetWatcherOsmanager(b)
	return tu
}

// SetNillableWatcherOsmanager sets the "watcher_osmanager" field if the given value is not nil.
func (tu *TenantUpdate) SetNillableWatcherOsmanager(b *bool) *TenantUpdate {
	if b != nil {
		tu.SetWatcherOsmanager(*b)
	}
	return tu
}

// ClearWatcherOsmanager clears the value of the "watcher_osmanager" field.
func (tu *TenantUpdate) ClearWatcherOsmanager() *TenantUpdate {
	tu.mutation.ClearWatcherOsmanager()
	return tu
}

// SetUpdatedAt sets the "updated_at" field.
func (tu *TenantUpdate) SetUpdatedAt(s string) *TenantUpdate {
	tu.mutation.SetUpdatedAt(s)
	return tu
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (tu *TenantUpdate) SetNillableUpdatedAt(s *string) *TenantUpdate {
	if s != nil {
		tu.SetUpdatedAt(*s)
	}
	return tu
}

// Mutation returns the TenantMutation object of the builder.
func (tu *TenantUpdate) Mutation() *TenantMutation {
	return tu.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (tu *TenantUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, tu.sqlSave, tu.mutation, tu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (tu *TenantUpdate) SaveX(ctx context.Context) int {
	affected, err := tu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (tu *TenantUpdate) Exec(ctx context.Context) error {
	_, err := tu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tu *TenantUpdate) ExecX(ctx context.Context) {
	if err := tu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (tu *TenantUpdate) check() error {
	if v, ok := tu.mutation.CurrentState(); ok {
		if err := tenant.CurrentStateValidator(v); err != nil {
			return &ValidationError{Name: "current_state", err: fmt.Errorf(`ent: validator failed for field "Tenant.current_state": %w`, err)}
		}
	}
	if v, ok := tu.mutation.DesiredState(); ok {
		if err := tenant.DesiredStateValidator(v); err != nil {
			return &ValidationError{Name: "desired_state", err: fmt.Errorf(`ent: validator failed for field "Tenant.desired_state": %w`, err)}
		}
	}
	return nil
}

func (tu *TenantUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := tu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(tenant.Table, tenant.Columns, sqlgraph.NewFieldSpec(tenant.FieldID, field.TypeInt))
	if ps := tu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := tu.mutation.ResourceID(); ok {
		_spec.SetField(tenant.FieldResourceID, field.TypeString, value)
	}
	if value, ok := tu.mutation.CurrentState(); ok {
		_spec.SetField(tenant.FieldCurrentState, field.TypeEnum, value)
	}
	if tu.mutation.CurrentStateCleared() {
		_spec.ClearField(tenant.FieldCurrentState, field.TypeEnum)
	}
	if value, ok := tu.mutation.DesiredState(); ok {
		_spec.SetField(tenant.FieldDesiredState, field.TypeEnum, value)
	}
	if value, ok := tu.mutation.WatcherOsmanager(); ok {
		_spec.SetField(tenant.FieldWatcherOsmanager, field.TypeBool, value)
	}
	if tu.mutation.WatcherOsmanagerCleared() {
		_spec.ClearField(tenant.FieldWatcherOsmanager, field.TypeBool)
	}
	if value, ok := tu.mutation.UpdatedAt(); ok {
		_spec.SetField(tenant.FieldUpdatedAt, field.TypeString, value)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, tu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{tenant.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	tu.mutation.done = true
	return n, nil
}

// TenantUpdateOne is the builder for updating a single Tenant entity.
type TenantUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *TenantMutation
}

// SetResourceID sets the "resource_id" field.
func (tuo *TenantUpdateOne) SetResourceID(s string) *TenantUpdateOne {
	tuo.mutation.SetResourceID(s)
	return tuo
}

// SetNillableResourceID sets the "resource_id" field if the given value is not nil.
func (tuo *TenantUpdateOne) SetNillableResourceID(s *string) *TenantUpdateOne {
	if s != nil {
		tuo.SetResourceID(*s)
	}
	return tuo
}

// SetCurrentState sets the "current_state" field.
func (tuo *TenantUpdateOne) SetCurrentState(ts tenant.CurrentState) *TenantUpdateOne {
	tuo.mutation.SetCurrentState(ts)
	return tuo
}

// SetNillableCurrentState sets the "current_state" field if the given value is not nil.
func (tuo *TenantUpdateOne) SetNillableCurrentState(ts *tenant.CurrentState) *TenantUpdateOne {
	if ts != nil {
		tuo.SetCurrentState(*ts)
	}
	return tuo
}

// ClearCurrentState clears the value of the "current_state" field.
func (tuo *TenantUpdateOne) ClearCurrentState() *TenantUpdateOne {
	tuo.mutation.ClearCurrentState()
	return tuo
}

// SetDesiredState sets the "desired_state" field.
func (tuo *TenantUpdateOne) SetDesiredState(ts tenant.DesiredState) *TenantUpdateOne {
	tuo.mutation.SetDesiredState(ts)
	return tuo
}

// SetNillableDesiredState sets the "desired_state" field if the given value is not nil.
func (tuo *TenantUpdateOne) SetNillableDesiredState(ts *tenant.DesiredState) *TenantUpdateOne {
	if ts != nil {
		tuo.SetDesiredState(*ts)
	}
	return tuo
}

// SetWatcherOsmanager sets the "watcher_osmanager" field.
func (tuo *TenantUpdateOne) SetWatcherOsmanager(b bool) *TenantUpdateOne {
	tuo.mutation.SetWatcherOsmanager(b)
	return tuo
}

// SetNillableWatcherOsmanager sets the "watcher_osmanager" field if the given value is not nil.
func (tuo *TenantUpdateOne) SetNillableWatcherOsmanager(b *bool) *TenantUpdateOne {
	if b != nil {
		tuo.SetWatcherOsmanager(*b)
	}
	return tuo
}

// ClearWatcherOsmanager clears the value of the "watcher_osmanager" field.
func (tuo *TenantUpdateOne) ClearWatcherOsmanager() *TenantUpdateOne {
	tuo.mutation.ClearWatcherOsmanager()
	return tuo
}

// SetUpdatedAt sets the "updated_at" field.
func (tuo *TenantUpdateOne) SetUpdatedAt(s string) *TenantUpdateOne {
	tuo.mutation.SetUpdatedAt(s)
	return tuo
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (tuo *TenantUpdateOne) SetNillableUpdatedAt(s *string) *TenantUpdateOne {
	if s != nil {
		tuo.SetUpdatedAt(*s)
	}
	return tuo
}

// Mutation returns the TenantMutation object of the builder.
func (tuo *TenantUpdateOne) Mutation() *TenantMutation {
	return tuo.mutation
}

// Where appends a list predicates to the TenantUpdate builder.
func (tuo *TenantUpdateOne) Where(ps ...predicate.Tenant) *TenantUpdateOne {
	tuo.mutation.Where(ps...)
	return tuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (tuo *TenantUpdateOne) Select(field string, fields ...string) *TenantUpdateOne {
	tuo.fields = append([]string{field}, fields...)
	return tuo
}

// Save executes the query and returns the updated Tenant entity.
func (tuo *TenantUpdateOne) Save(ctx context.Context) (*Tenant, error) {
	return withHooks(ctx, tuo.sqlSave, tuo.mutation, tuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (tuo *TenantUpdateOne) SaveX(ctx context.Context) *Tenant {
	node, err := tuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (tuo *TenantUpdateOne) Exec(ctx context.Context) error {
	_, err := tuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tuo *TenantUpdateOne) ExecX(ctx context.Context) {
	if err := tuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (tuo *TenantUpdateOne) check() error {
	if v, ok := tuo.mutation.CurrentState(); ok {
		if err := tenant.CurrentStateValidator(v); err != nil {
			return &ValidationError{Name: "current_state", err: fmt.Errorf(`ent: validator failed for field "Tenant.current_state": %w`, err)}
		}
	}
	if v, ok := tuo.mutation.DesiredState(); ok {
		if err := tenant.DesiredStateValidator(v); err != nil {
			return &ValidationError{Name: "desired_state", err: fmt.Errorf(`ent: validator failed for field "Tenant.desired_state": %w`, err)}
		}
	}
	return nil
}

func (tuo *TenantUpdateOne) sqlSave(ctx context.Context) (_node *Tenant, err error) {
	if err := tuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(tenant.Table, tenant.Columns, sqlgraph.NewFieldSpec(tenant.FieldID, field.TypeInt))
	id, ok := tuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Tenant.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := tuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, tenant.FieldID)
		for _, f := range fields {
			if !tenant.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != tenant.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := tuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := tuo.mutation.ResourceID(); ok {
		_spec.SetField(tenant.FieldResourceID, field.TypeString, value)
	}
	if value, ok := tuo.mutation.CurrentState(); ok {
		_spec.SetField(tenant.FieldCurrentState, field.TypeEnum, value)
	}
	if tuo.mutation.CurrentStateCleared() {
		_spec.ClearField(tenant.FieldCurrentState, field.TypeEnum)
	}
	if value, ok := tuo.mutation.DesiredState(); ok {
		_spec.SetField(tenant.FieldDesiredState, field.TypeEnum, value)
	}
	if value, ok := tuo.mutation.WatcherOsmanager(); ok {
		_spec.SetField(tenant.FieldWatcherOsmanager, field.TypeBool, value)
	}
	if tuo.mutation.WatcherOsmanagerCleared() {
		_spec.ClearField(tenant.FieldWatcherOsmanager, field.TypeBool)
	}
	if value, ok := tuo.mutation.UpdatedAt(); ok {
		_spec.SetField(tenant.FieldUpdatedAt, field.TypeString, value)
	}
	_node = &Tenant{config: tuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, tuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{tenant.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	tuo.mutation.done = true
	return _node, nil
}

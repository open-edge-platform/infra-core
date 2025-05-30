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
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/providerresource"
)

// ProviderResourceUpdate is the builder for updating ProviderResource entities.
type ProviderResourceUpdate struct {
	config
	hooks    []Hook
	mutation *ProviderResourceMutation
}

// Where appends a list predicates to the ProviderResourceUpdate builder.
func (pru *ProviderResourceUpdate) Where(ps ...predicate.ProviderResource) *ProviderResourceUpdate {
	pru.mutation.Where(ps...)
	return pru
}

// SetResourceID sets the "resource_id" field.
func (pru *ProviderResourceUpdate) SetResourceID(s string) *ProviderResourceUpdate {
	pru.mutation.SetResourceID(s)
	return pru
}

// SetNillableResourceID sets the "resource_id" field if the given value is not nil.
func (pru *ProviderResourceUpdate) SetNillableResourceID(s *string) *ProviderResourceUpdate {
	if s != nil {
		pru.SetResourceID(*s)
	}
	return pru
}

// SetProviderKind sets the "provider_kind" field.
func (pru *ProviderResourceUpdate) SetProviderKind(pk providerresource.ProviderKind) *ProviderResourceUpdate {
	pru.mutation.SetProviderKind(pk)
	return pru
}

// SetNillableProviderKind sets the "provider_kind" field if the given value is not nil.
func (pru *ProviderResourceUpdate) SetNillableProviderKind(pk *providerresource.ProviderKind) *ProviderResourceUpdate {
	if pk != nil {
		pru.SetProviderKind(*pk)
	}
	return pru
}

// SetProviderVendor sets the "provider_vendor" field.
func (pru *ProviderResourceUpdate) SetProviderVendor(pv providerresource.ProviderVendor) *ProviderResourceUpdate {
	pru.mutation.SetProviderVendor(pv)
	return pru
}

// SetNillableProviderVendor sets the "provider_vendor" field if the given value is not nil.
func (pru *ProviderResourceUpdate) SetNillableProviderVendor(pv *providerresource.ProviderVendor) *ProviderResourceUpdate {
	if pv != nil {
		pru.SetProviderVendor(*pv)
	}
	return pru
}

// ClearProviderVendor clears the value of the "provider_vendor" field.
func (pru *ProviderResourceUpdate) ClearProviderVendor() *ProviderResourceUpdate {
	pru.mutation.ClearProviderVendor()
	return pru
}

// SetName sets the "name" field.
func (pru *ProviderResourceUpdate) SetName(s string) *ProviderResourceUpdate {
	pru.mutation.SetName(s)
	return pru
}

// SetNillableName sets the "name" field if the given value is not nil.
func (pru *ProviderResourceUpdate) SetNillableName(s *string) *ProviderResourceUpdate {
	if s != nil {
		pru.SetName(*s)
	}
	return pru
}

// SetAPIEndpoint sets the "api_endpoint" field.
func (pru *ProviderResourceUpdate) SetAPIEndpoint(s string) *ProviderResourceUpdate {
	pru.mutation.SetAPIEndpoint(s)
	return pru
}

// SetNillableAPIEndpoint sets the "api_endpoint" field if the given value is not nil.
func (pru *ProviderResourceUpdate) SetNillableAPIEndpoint(s *string) *ProviderResourceUpdate {
	if s != nil {
		pru.SetAPIEndpoint(*s)
	}
	return pru
}

// SetAPICredentials sets the "api_credentials" field.
func (pru *ProviderResourceUpdate) SetAPICredentials(s string) *ProviderResourceUpdate {
	pru.mutation.SetAPICredentials(s)
	return pru
}

// SetNillableAPICredentials sets the "api_credentials" field if the given value is not nil.
func (pru *ProviderResourceUpdate) SetNillableAPICredentials(s *string) *ProviderResourceUpdate {
	if s != nil {
		pru.SetAPICredentials(*s)
	}
	return pru
}

// ClearAPICredentials clears the value of the "api_credentials" field.
func (pru *ProviderResourceUpdate) ClearAPICredentials() *ProviderResourceUpdate {
	pru.mutation.ClearAPICredentials()
	return pru
}

// SetConfig sets the "config" field.
func (pru *ProviderResourceUpdate) SetConfig(s string) *ProviderResourceUpdate {
	pru.mutation.SetConfig(s)
	return pru
}

// SetNillableConfig sets the "config" field if the given value is not nil.
func (pru *ProviderResourceUpdate) SetNillableConfig(s *string) *ProviderResourceUpdate {
	if s != nil {
		pru.SetConfig(*s)
	}
	return pru
}

// ClearConfig clears the value of the "config" field.
func (pru *ProviderResourceUpdate) ClearConfig() *ProviderResourceUpdate {
	pru.mutation.ClearConfig()
	return pru
}

// SetUpdatedAt sets the "updated_at" field.
func (pru *ProviderResourceUpdate) SetUpdatedAt(s string) *ProviderResourceUpdate {
	pru.mutation.SetUpdatedAt(s)
	return pru
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (pru *ProviderResourceUpdate) SetNillableUpdatedAt(s *string) *ProviderResourceUpdate {
	if s != nil {
		pru.SetUpdatedAt(*s)
	}
	return pru
}

// Mutation returns the ProviderResourceMutation object of the builder.
func (pru *ProviderResourceUpdate) Mutation() *ProviderResourceMutation {
	return pru.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (pru *ProviderResourceUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, pru.sqlSave, pru.mutation, pru.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (pru *ProviderResourceUpdate) SaveX(ctx context.Context) int {
	affected, err := pru.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (pru *ProviderResourceUpdate) Exec(ctx context.Context) error {
	_, err := pru.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pru *ProviderResourceUpdate) ExecX(ctx context.Context) {
	if err := pru.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (pru *ProviderResourceUpdate) check() error {
	if v, ok := pru.mutation.ProviderKind(); ok {
		if err := providerresource.ProviderKindValidator(v); err != nil {
			return &ValidationError{Name: "provider_kind", err: fmt.Errorf(`ent: validator failed for field "ProviderResource.provider_kind": %w`, err)}
		}
	}
	if v, ok := pru.mutation.ProviderVendor(); ok {
		if err := providerresource.ProviderVendorValidator(v); err != nil {
			return &ValidationError{Name: "provider_vendor", err: fmt.Errorf(`ent: validator failed for field "ProviderResource.provider_vendor": %w`, err)}
		}
	}
	return nil
}

func (pru *ProviderResourceUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := pru.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(providerresource.Table, providerresource.Columns, sqlgraph.NewFieldSpec(providerresource.FieldID, field.TypeInt))
	if ps := pru.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := pru.mutation.ResourceID(); ok {
		_spec.SetField(providerresource.FieldResourceID, field.TypeString, value)
	}
	if value, ok := pru.mutation.ProviderKind(); ok {
		_spec.SetField(providerresource.FieldProviderKind, field.TypeEnum, value)
	}
	if value, ok := pru.mutation.ProviderVendor(); ok {
		_spec.SetField(providerresource.FieldProviderVendor, field.TypeEnum, value)
	}
	if pru.mutation.ProviderVendorCleared() {
		_spec.ClearField(providerresource.FieldProviderVendor, field.TypeEnum)
	}
	if value, ok := pru.mutation.Name(); ok {
		_spec.SetField(providerresource.FieldName, field.TypeString, value)
	}
	if value, ok := pru.mutation.APIEndpoint(); ok {
		_spec.SetField(providerresource.FieldAPIEndpoint, field.TypeString, value)
	}
	if value, ok := pru.mutation.APICredentials(); ok {
		_spec.SetField(providerresource.FieldAPICredentials, field.TypeString, value)
	}
	if pru.mutation.APICredentialsCleared() {
		_spec.ClearField(providerresource.FieldAPICredentials, field.TypeString)
	}
	if value, ok := pru.mutation.Config(); ok {
		_spec.SetField(providerresource.FieldConfig, field.TypeString, value)
	}
	if pru.mutation.ConfigCleared() {
		_spec.ClearField(providerresource.FieldConfig, field.TypeString)
	}
	if value, ok := pru.mutation.UpdatedAt(); ok {
		_spec.SetField(providerresource.FieldUpdatedAt, field.TypeString, value)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, pru.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{providerresource.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	pru.mutation.done = true
	return n, nil
}

// ProviderResourceUpdateOne is the builder for updating a single ProviderResource entity.
type ProviderResourceUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *ProviderResourceMutation
}

// SetResourceID sets the "resource_id" field.
func (pruo *ProviderResourceUpdateOne) SetResourceID(s string) *ProviderResourceUpdateOne {
	pruo.mutation.SetResourceID(s)
	return pruo
}

// SetNillableResourceID sets the "resource_id" field if the given value is not nil.
func (pruo *ProviderResourceUpdateOne) SetNillableResourceID(s *string) *ProviderResourceUpdateOne {
	if s != nil {
		pruo.SetResourceID(*s)
	}
	return pruo
}

// SetProviderKind sets the "provider_kind" field.
func (pruo *ProviderResourceUpdateOne) SetProviderKind(pk providerresource.ProviderKind) *ProviderResourceUpdateOne {
	pruo.mutation.SetProviderKind(pk)
	return pruo
}

// SetNillableProviderKind sets the "provider_kind" field if the given value is not nil.
func (pruo *ProviderResourceUpdateOne) SetNillableProviderKind(pk *providerresource.ProviderKind) *ProviderResourceUpdateOne {
	if pk != nil {
		pruo.SetProviderKind(*pk)
	}
	return pruo
}

// SetProviderVendor sets the "provider_vendor" field.
func (pruo *ProviderResourceUpdateOne) SetProviderVendor(pv providerresource.ProviderVendor) *ProviderResourceUpdateOne {
	pruo.mutation.SetProviderVendor(pv)
	return pruo
}

// SetNillableProviderVendor sets the "provider_vendor" field if the given value is not nil.
func (pruo *ProviderResourceUpdateOne) SetNillableProviderVendor(pv *providerresource.ProviderVendor) *ProviderResourceUpdateOne {
	if pv != nil {
		pruo.SetProviderVendor(*pv)
	}
	return pruo
}

// ClearProviderVendor clears the value of the "provider_vendor" field.
func (pruo *ProviderResourceUpdateOne) ClearProviderVendor() *ProviderResourceUpdateOne {
	pruo.mutation.ClearProviderVendor()
	return pruo
}

// SetName sets the "name" field.
func (pruo *ProviderResourceUpdateOne) SetName(s string) *ProviderResourceUpdateOne {
	pruo.mutation.SetName(s)
	return pruo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (pruo *ProviderResourceUpdateOne) SetNillableName(s *string) *ProviderResourceUpdateOne {
	if s != nil {
		pruo.SetName(*s)
	}
	return pruo
}

// SetAPIEndpoint sets the "api_endpoint" field.
func (pruo *ProviderResourceUpdateOne) SetAPIEndpoint(s string) *ProviderResourceUpdateOne {
	pruo.mutation.SetAPIEndpoint(s)
	return pruo
}

// SetNillableAPIEndpoint sets the "api_endpoint" field if the given value is not nil.
func (pruo *ProviderResourceUpdateOne) SetNillableAPIEndpoint(s *string) *ProviderResourceUpdateOne {
	if s != nil {
		pruo.SetAPIEndpoint(*s)
	}
	return pruo
}

// SetAPICredentials sets the "api_credentials" field.
func (pruo *ProviderResourceUpdateOne) SetAPICredentials(s string) *ProviderResourceUpdateOne {
	pruo.mutation.SetAPICredentials(s)
	return pruo
}

// SetNillableAPICredentials sets the "api_credentials" field if the given value is not nil.
func (pruo *ProviderResourceUpdateOne) SetNillableAPICredentials(s *string) *ProviderResourceUpdateOne {
	if s != nil {
		pruo.SetAPICredentials(*s)
	}
	return pruo
}

// ClearAPICredentials clears the value of the "api_credentials" field.
func (pruo *ProviderResourceUpdateOne) ClearAPICredentials() *ProviderResourceUpdateOne {
	pruo.mutation.ClearAPICredentials()
	return pruo
}

// SetConfig sets the "config" field.
func (pruo *ProviderResourceUpdateOne) SetConfig(s string) *ProviderResourceUpdateOne {
	pruo.mutation.SetConfig(s)
	return pruo
}

// SetNillableConfig sets the "config" field if the given value is not nil.
func (pruo *ProviderResourceUpdateOne) SetNillableConfig(s *string) *ProviderResourceUpdateOne {
	if s != nil {
		pruo.SetConfig(*s)
	}
	return pruo
}

// ClearConfig clears the value of the "config" field.
func (pruo *ProviderResourceUpdateOne) ClearConfig() *ProviderResourceUpdateOne {
	pruo.mutation.ClearConfig()
	return pruo
}

// SetUpdatedAt sets the "updated_at" field.
func (pruo *ProviderResourceUpdateOne) SetUpdatedAt(s string) *ProviderResourceUpdateOne {
	pruo.mutation.SetUpdatedAt(s)
	return pruo
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (pruo *ProviderResourceUpdateOne) SetNillableUpdatedAt(s *string) *ProviderResourceUpdateOne {
	if s != nil {
		pruo.SetUpdatedAt(*s)
	}
	return pruo
}

// Mutation returns the ProviderResourceMutation object of the builder.
func (pruo *ProviderResourceUpdateOne) Mutation() *ProviderResourceMutation {
	return pruo.mutation
}

// Where appends a list predicates to the ProviderResourceUpdate builder.
func (pruo *ProviderResourceUpdateOne) Where(ps ...predicate.ProviderResource) *ProviderResourceUpdateOne {
	pruo.mutation.Where(ps...)
	return pruo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (pruo *ProviderResourceUpdateOne) Select(field string, fields ...string) *ProviderResourceUpdateOne {
	pruo.fields = append([]string{field}, fields...)
	return pruo
}

// Save executes the query and returns the updated ProviderResource entity.
func (pruo *ProviderResourceUpdateOne) Save(ctx context.Context) (*ProviderResource, error) {
	return withHooks(ctx, pruo.sqlSave, pruo.mutation, pruo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (pruo *ProviderResourceUpdateOne) SaveX(ctx context.Context) *ProviderResource {
	node, err := pruo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (pruo *ProviderResourceUpdateOne) Exec(ctx context.Context) error {
	_, err := pruo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pruo *ProviderResourceUpdateOne) ExecX(ctx context.Context) {
	if err := pruo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (pruo *ProviderResourceUpdateOne) check() error {
	if v, ok := pruo.mutation.ProviderKind(); ok {
		if err := providerresource.ProviderKindValidator(v); err != nil {
			return &ValidationError{Name: "provider_kind", err: fmt.Errorf(`ent: validator failed for field "ProviderResource.provider_kind": %w`, err)}
		}
	}
	if v, ok := pruo.mutation.ProviderVendor(); ok {
		if err := providerresource.ProviderVendorValidator(v); err != nil {
			return &ValidationError{Name: "provider_vendor", err: fmt.Errorf(`ent: validator failed for field "ProviderResource.provider_vendor": %w`, err)}
		}
	}
	return nil
}

func (pruo *ProviderResourceUpdateOne) sqlSave(ctx context.Context) (_node *ProviderResource, err error) {
	if err := pruo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(providerresource.Table, providerresource.Columns, sqlgraph.NewFieldSpec(providerresource.FieldID, field.TypeInt))
	id, ok := pruo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "ProviderResource.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := pruo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, providerresource.FieldID)
		for _, f := range fields {
			if !providerresource.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != providerresource.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := pruo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := pruo.mutation.ResourceID(); ok {
		_spec.SetField(providerresource.FieldResourceID, field.TypeString, value)
	}
	if value, ok := pruo.mutation.ProviderKind(); ok {
		_spec.SetField(providerresource.FieldProviderKind, field.TypeEnum, value)
	}
	if value, ok := pruo.mutation.ProviderVendor(); ok {
		_spec.SetField(providerresource.FieldProviderVendor, field.TypeEnum, value)
	}
	if pruo.mutation.ProviderVendorCleared() {
		_spec.ClearField(providerresource.FieldProviderVendor, field.TypeEnum)
	}
	if value, ok := pruo.mutation.Name(); ok {
		_spec.SetField(providerresource.FieldName, field.TypeString, value)
	}
	if value, ok := pruo.mutation.APIEndpoint(); ok {
		_spec.SetField(providerresource.FieldAPIEndpoint, field.TypeString, value)
	}
	if value, ok := pruo.mutation.APICredentials(); ok {
		_spec.SetField(providerresource.FieldAPICredentials, field.TypeString, value)
	}
	if pruo.mutation.APICredentialsCleared() {
		_spec.ClearField(providerresource.FieldAPICredentials, field.TypeString)
	}
	if value, ok := pruo.mutation.Config(); ok {
		_spec.SetField(providerresource.FieldConfig, field.TypeString, value)
	}
	if pruo.mutation.ConfigCleared() {
		_spec.ClearField(providerresource.FieldConfig, field.TypeString)
	}
	if value, ok := pruo.mutation.UpdatedAt(); ok {
		_spec.SetField(providerresource.FieldUpdatedAt, field.TypeString, value)
	}
	_node = &ProviderResource{config: pruo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, pruo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{providerresource.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	pruo.mutation.done = true
	return _node, nil
}

// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/instanceresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/remoteaccessconfiguration"
)

// RemoteAccessConfigurationCreate is the builder for creating a RemoteAccessConfiguration entity.
type RemoteAccessConfigurationCreate struct {
	config
	mutation *RemoteAccessConfigurationMutation
	hooks    []Hook
}

// SetResourceID sets the "resource_id" field.
func (racc *RemoteAccessConfigurationCreate) SetResourceID(s string) *RemoteAccessConfigurationCreate {
	racc.mutation.SetResourceID(s)
	return racc
}

// SetExpirationTimestamp sets the "expiration_timestamp" field.
func (racc *RemoteAccessConfigurationCreate) SetExpirationTimestamp(u uint64) *RemoteAccessConfigurationCreate {
	racc.mutation.SetExpirationTimestamp(u)
	return racc
}

// SetLocalPort sets the "local_port" field.
func (racc *RemoteAccessConfigurationCreate) SetLocalPort(u uint32) *RemoteAccessConfigurationCreate {
	racc.mutation.SetLocalPort(u)
	return racc
}

// SetNillableLocalPort sets the "local_port" field if the given value is not nil.
func (racc *RemoteAccessConfigurationCreate) SetNillableLocalPort(u *uint32) *RemoteAccessConfigurationCreate {
	if u != nil {
		racc.SetLocalPort(*u)
	}
	return racc
}

// SetUser sets the "user" field.
func (racc *RemoteAccessConfigurationCreate) SetUser(s string) *RemoteAccessConfigurationCreate {
	racc.mutation.SetUser(s)
	return racc
}

// SetNillableUser sets the "user" field if the given value is not nil.
func (racc *RemoteAccessConfigurationCreate) SetNillableUser(s *string) *RemoteAccessConfigurationCreate {
	if s != nil {
		racc.SetUser(*s)
	}
	return racc
}

// SetCurrentState sets the "current_state" field.
func (racc *RemoteAccessConfigurationCreate) SetCurrentState(rs remoteaccessconfiguration.CurrentState) *RemoteAccessConfigurationCreate {
	racc.mutation.SetCurrentState(rs)
	return racc
}

// SetNillableCurrentState sets the "current_state" field if the given value is not nil.
func (racc *RemoteAccessConfigurationCreate) SetNillableCurrentState(rs *remoteaccessconfiguration.CurrentState) *RemoteAccessConfigurationCreate {
	if rs != nil {
		racc.SetCurrentState(*rs)
	}
	return racc
}

// SetDesiredState sets the "desired_state" field.
func (racc *RemoteAccessConfigurationCreate) SetDesiredState(rs remoteaccessconfiguration.DesiredState) *RemoteAccessConfigurationCreate {
	racc.mutation.SetDesiredState(rs)
	return racc
}

// SetConfigurationStatus sets the "configuration_status" field.
func (racc *RemoteAccessConfigurationCreate) SetConfigurationStatus(s string) *RemoteAccessConfigurationCreate {
	racc.mutation.SetConfigurationStatus(s)
	return racc
}

// SetNillableConfigurationStatus sets the "configuration_status" field if the given value is not nil.
func (racc *RemoteAccessConfigurationCreate) SetNillableConfigurationStatus(s *string) *RemoteAccessConfigurationCreate {
	if s != nil {
		racc.SetConfigurationStatus(*s)
	}
	return racc
}

// SetConfigurationStatusIndicator sets the "configuration_status_indicator" field.
func (racc *RemoteAccessConfigurationCreate) SetConfigurationStatusIndicator(rsi remoteaccessconfiguration.ConfigurationStatusIndicator) *RemoteAccessConfigurationCreate {
	racc.mutation.SetConfigurationStatusIndicator(rsi)
	return racc
}

// SetNillableConfigurationStatusIndicator sets the "configuration_status_indicator" field if the given value is not nil.
func (racc *RemoteAccessConfigurationCreate) SetNillableConfigurationStatusIndicator(rsi *remoteaccessconfiguration.ConfigurationStatusIndicator) *RemoteAccessConfigurationCreate {
	if rsi != nil {
		racc.SetConfigurationStatusIndicator(*rsi)
	}
	return racc
}

// SetConfigurationStatusTimestamp sets the "configuration_status_timestamp" field.
func (racc *RemoteAccessConfigurationCreate) SetConfigurationStatusTimestamp(u uint64) *RemoteAccessConfigurationCreate {
	racc.mutation.SetConfigurationStatusTimestamp(u)
	return racc
}

// SetNillableConfigurationStatusTimestamp sets the "configuration_status_timestamp" field if the given value is not nil.
func (racc *RemoteAccessConfigurationCreate) SetNillableConfigurationStatusTimestamp(u *uint64) *RemoteAccessConfigurationCreate {
	if u != nil {
		racc.SetConfigurationStatusTimestamp(*u)
	}
	return racc
}

// SetTenantID sets the "tenant_id" field.
func (racc *RemoteAccessConfigurationCreate) SetTenantID(s string) *RemoteAccessConfigurationCreate {
	racc.mutation.SetTenantID(s)
	return racc
}

// SetCreatedAt sets the "created_at" field.
func (racc *RemoteAccessConfigurationCreate) SetCreatedAt(s string) *RemoteAccessConfigurationCreate {
	racc.mutation.SetCreatedAt(s)
	return racc
}

// SetUpdatedAt sets the "updated_at" field.
func (racc *RemoteAccessConfigurationCreate) SetUpdatedAt(s string) *RemoteAccessConfigurationCreate {
	racc.mutation.SetUpdatedAt(s)
	return racc
}

// SetInstanceID sets the "instance" edge to the InstanceResource entity by ID.
func (racc *RemoteAccessConfigurationCreate) SetInstanceID(id int) *RemoteAccessConfigurationCreate {
	racc.mutation.SetInstanceID(id)
	return racc
}

// SetInstance sets the "instance" edge to the InstanceResource entity.
func (racc *RemoteAccessConfigurationCreate) SetInstance(i *InstanceResource) *RemoteAccessConfigurationCreate {
	return racc.SetInstanceID(i.ID)
}

// Mutation returns the RemoteAccessConfigurationMutation object of the builder.
func (racc *RemoteAccessConfigurationCreate) Mutation() *RemoteAccessConfigurationMutation {
	return racc.mutation
}

// Save creates the RemoteAccessConfiguration in the database.
func (racc *RemoteAccessConfigurationCreate) Save(ctx context.Context) (*RemoteAccessConfiguration, error) {
	return withHooks(ctx, racc.sqlSave, racc.mutation, racc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (racc *RemoteAccessConfigurationCreate) SaveX(ctx context.Context) *RemoteAccessConfiguration {
	v, err := racc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (racc *RemoteAccessConfigurationCreate) Exec(ctx context.Context) error {
	_, err := racc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (racc *RemoteAccessConfigurationCreate) ExecX(ctx context.Context) {
	if err := racc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (racc *RemoteAccessConfigurationCreate) check() error {
	if _, ok := racc.mutation.ResourceID(); !ok {
		return &ValidationError{Name: "resource_id", err: errors.New(`ent: missing required field "RemoteAccessConfiguration.resource_id"`)}
	}
	if _, ok := racc.mutation.ExpirationTimestamp(); !ok {
		return &ValidationError{Name: "expiration_timestamp", err: errors.New(`ent: missing required field "RemoteAccessConfiguration.expiration_timestamp"`)}
	}
	if v, ok := racc.mutation.CurrentState(); ok {
		if err := remoteaccessconfiguration.CurrentStateValidator(v); err != nil {
			return &ValidationError{Name: "current_state", err: fmt.Errorf(`ent: validator failed for field "RemoteAccessConfiguration.current_state": %w`, err)}
		}
	}
	if _, ok := racc.mutation.DesiredState(); !ok {
		return &ValidationError{Name: "desired_state", err: errors.New(`ent: missing required field "RemoteAccessConfiguration.desired_state"`)}
	}
	if v, ok := racc.mutation.DesiredState(); ok {
		if err := remoteaccessconfiguration.DesiredStateValidator(v); err != nil {
			return &ValidationError{Name: "desired_state", err: fmt.Errorf(`ent: validator failed for field "RemoteAccessConfiguration.desired_state": %w`, err)}
		}
	}
	if v, ok := racc.mutation.ConfigurationStatusIndicator(); ok {
		if err := remoteaccessconfiguration.ConfigurationStatusIndicatorValidator(v); err != nil {
			return &ValidationError{Name: "configuration_status_indicator", err: fmt.Errorf(`ent: validator failed for field "RemoteAccessConfiguration.configuration_status_indicator": %w`, err)}
		}
	}
	if _, ok := racc.mutation.TenantID(); !ok {
		return &ValidationError{Name: "tenant_id", err: errors.New(`ent: missing required field "RemoteAccessConfiguration.tenant_id"`)}
	}
	if _, ok := racc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "RemoteAccessConfiguration.created_at"`)}
	}
	if _, ok := racc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "RemoteAccessConfiguration.updated_at"`)}
	}
	if len(racc.mutation.InstanceIDs()) == 0 {
		return &ValidationError{Name: "instance", err: errors.New(`ent: missing required edge "RemoteAccessConfiguration.instance"`)}
	}
	return nil
}

func (racc *RemoteAccessConfigurationCreate) sqlSave(ctx context.Context) (*RemoteAccessConfiguration, error) {
	if err := racc.check(); err != nil {
		return nil, err
	}
	_node, _spec := racc.createSpec()
	if err := sqlgraph.CreateNode(ctx, racc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	racc.mutation.id = &_node.ID
	racc.mutation.done = true
	return _node, nil
}

func (racc *RemoteAccessConfigurationCreate) createSpec() (*RemoteAccessConfiguration, *sqlgraph.CreateSpec) {
	var (
		_node = &RemoteAccessConfiguration{config: racc.config}
		_spec = sqlgraph.NewCreateSpec(remoteaccessconfiguration.Table, sqlgraph.NewFieldSpec(remoteaccessconfiguration.FieldID, field.TypeInt))
	)
	if value, ok := racc.mutation.ResourceID(); ok {
		_spec.SetField(remoteaccessconfiguration.FieldResourceID, field.TypeString, value)
		_node.ResourceID = value
	}
	if value, ok := racc.mutation.ExpirationTimestamp(); ok {
		_spec.SetField(remoteaccessconfiguration.FieldExpirationTimestamp, field.TypeUint64, value)
		_node.ExpirationTimestamp = value
	}
	if value, ok := racc.mutation.LocalPort(); ok {
		_spec.SetField(remoteaccessconfiguration.FieldLocalPort, field.TypeUint32, value)
		_node.LocalPort = value
	}
	if value, ok := racc.mutation.User(); ok {
		_spec.SetField(remoteaccessconfiguration.FieldUser, field.TypeString, value)
		_node.User = value
	}
	if value, ok := racc.mutation.CurrentState(); ok {
		_spec.SetField(remoteaccessconfiguration.FieldCurrentState, field.TypeEnum, value)
		_node.CurrentState = value
	}
	if value, ok := racc.mutation.DesiredState(); ok {
		_spec.SetField(remoteaccessconfiguration.FieldDesiredState, field.TypeEnum, value)
		_node.DesiredState = value
	}
	if value, ok := racc.mutation.ConfigurationStatus(); ok {
		_spec.SetField(remoteaccessconfiguration.FieldConfigurationStatus, field.TypeString, value)
		_node.ConfigurationStatus = value
	}
	if value, ok := racc.mutation.ConfigurationStatusIndicator(); ok {
		_spec.SetField(remoteaccessconfiguration.FieldConfigurationStatusIndicator, field.TypeEnum, value)
		_node.ConfigurationStatusIndicator = value
	}
	if value, ok := racc.mutation.ConfigurationStatusTimestamp(); ok {
		_spec.SetField(remoteaccessconfiguration.FieldConfigurationStatusTimestamp, field.TypeUint64, value)
		_node.ConfigurationStatusTimestamp = value
	}
	if value, ok := racc.mutation.TenantID(); ok {
		_spec.SetField(remoteaccessconfiguration.FieldTenantID, field.TypeString, value)
		_node.TenantID = value
	}
	if value, ok := racc.mutation.CreatedAt(); ok {
		_spec.SetField(remoteaccessconfiguration.FieldCreatedAt, field.TypeString, value)
		_node.CreatedAt = value
	}
	if value, ok := racc.mutation.UpdatedAt(); ok {
		_spec.SetField(remoteaccessconfiguration.FieldUpdatedAt, field.TypeString, value)
		_node.UpdatedAt = value
	}
	if nodes := racc.mutation.InstanceIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   remoteaccessconfiguration.InstanceTable,
			Columns: []string{remoteaccessconfiguration.InstanceColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(instanceresource.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.remote_access_configuration_instance = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// RemoteAccessConfigurationCreateBulk is the builder for creating many RemoteAccessConfiguration entities in bulk.
type RemoteAccessConfigurationCreateBulk struct {
	config
	err      error
	builders []*RemoteAccessConfigurationCreate
}

// Save creates the RemoteAccessConfiguration entities in the database.
func (raccb *RemoteAccessConfigurationCreateBulk) Save(ctx context.Context) ([]*RemoteAccessConfiguration, error) {
	if raccb.err != nil {
		return nil, raccb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(raccb.builders))
	nodes := make([]*RemoteAccessConfiguration, len(raccb.builders))
	mutators := make([]Mutator, len(raccb.builders))
	for i := range raccb.builders {
		func(i int, root context.Context) {
			builder := raccb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*RemoteAccessConfigurationMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, raccb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, raccb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				if specs[i].ID.Value != nil {
					id := specs[i].ID.Value.(int64)
					nodes[i].ID = int(id)
				}
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, raccb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (raccb *RemoteAccessConfigurationCreateBulk) SaveX(ctx context.Context) []*RemoteAccessConfiguration {
	v, err := raccb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (raccb *RemoteAccessConfigurationCreateBulk) Exec(ctx context.Context) error {
	_, err := raccb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (raccb *RemoteAccessConfigurationCreateBulk) ExecX(ctx context.Context) {
	if err := raccb.Exec(ctx); err != nil {
		panic(err)
	}
}

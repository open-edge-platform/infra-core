// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/instanceresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/regionresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/siteresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/telemetrygroupresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/telemetryprofile"
)

// TelemetryProfileCreate is the builder for creating a TelemetryProfile entity.
type TelemetryProfileCreate struct {
	config
	mutation *TelemetryProfileMutation
	hooks    []Hook
}

// SetResourceID sets the "resource_id" field.
func (tpc *TelemetryProfileCreate) SetResourceID(s string) *TelemetryProfileCreate {
	tpc.mutation.SetResourceID(s)
	return tpc
}

// SetKind sets the "kind" field.
func (tpc *TelemetryProfileCreate) SetKind(t telemetryprofile.Kind) *TelemetryProfileCreate {
	tpc.mutation.SetKind(t)
	return tpc
}

// SetMetricsInterval sets the "metrics_interval" field.
func (tpc *TelemetryProfileCreate) SetMetricsInterval(u uint32) *TelemetryProfileCreate {
	tpc.mutation.SetMetricsInterval(u)
	return tpc
}

// SetNillableMetricsInterval sets the "metrics_interval" field if the given value is not nil.
func (tpc *TelemetryProfileCreate) SetNillableMetricsInterval(u *uint32) *TelemetryProfileCreate {
	if u != nil {
		tpc.SetMetricsInterval(*u)
	}
	return tpc
}

// SetLogLevel sets the "log_level" field.
func (tpc *TelemetryProfileCreate) SetLogLevel(tl telemetryprofile.LogLevel) *TelemetryProfileCreate {
	tpc.mutation.SetLogLevel(tl)
	return tpc
}

// SetNillableLogLevel sets the "log_level" field if the given value is not nil.
func (tpc *TelemetryProfileCreate) SetNillableLogLevel(tl *telemetryprofile.LogLevel) *TelemetryProfileCreate {
	if tl != nil {
		tpc.SetLogLevel(*tl)
	}
	return tpc
}

// SetTenantID sets the "tenant_id" field.
func (tpc *TelemetryProfileCreate) SetTenantID(s string) *TelemetryProfileCreate {
	tpc.mutation.SetTenantID(s)
	return tpc
}

// SetCreatedAt sets the "created_at" field.
func (tpc *TelemetryProfileCreate) SetCreatedAt(s string) *TelemetryProfileCreate {
	tpc.mutation.SetCreatedAt(s)
	return tpc
}

// SetUpdatedAt sets the "updated_at" field.
func (tpc *TelemetryProfileCreate) SetUpdatedAt(s string) *TelemetryProfileCreate {
	tpc.mutation.SetUpdatedAt(s)
	return tpc
}

// SetRegionID sets the "region" edge to the RegionResource entity by ID.
func (tpc *TelemetryProfileCreate) SetRegionID(id int) *TelemetryProfileCreate {
	tpc.mutation.SetRegionID(id)
	return tpc
}

// SetNillableRegionID sets the "region" edge to the RegionResource entity by ID if the given value is not nil.
func (tpc *TelemetryProfileCreate) SetNillableRegionID(id *int) *TelemetryProfileCreate {
	if id != nil {
		tpc = tpc.SetRegionID(*id)
	}
	return tpc
}

// SetRegion sets the "region" edge to the RegionResource entity.
func (tpc *TelemetryProfileCreate) SetRegion(r *RegionResource) *TelemetryProfileCreate {
	return tpc.SetRegionID(r.ID)
}

// SetSiteID sets the "site" edge to the SiteResource entity by ID.
func (tpc *TelemetryProfileCreate) SetSiteID(id int) *TelemetryProfileCreate {
	tpc.mutation.SetSiteID(id)
	return tpc
}

// SetNillableSiteID sets the "site" edge to the SiteResource entity by ID if the given value is not nil.
func (tpc *TelemetryProfileCreate) SetNillableSiteID(id *int) *TelemetryProfileCreate {
	if id != nil {
		tpc = tpc.SetSiteID(*id)
	}
	return tpc
}

// SetSite sets the "site" edge to the SiteResource entity.
func (tpc *TelemetryProfileCreate) SetSite(s *SiteResource) *TelemetryProfileCreate {
	return tpc.SetSiteID(s.ID)
}

// SetInstanceID sets the "instance" edge to the InstanceResource entity by ID.
func (tpc *TelemetryProfileCreate) SetInstanceID(id int) *TelemetryProfileCreate {
	tpc.mutation.SetInstanceID(id)
	return tpc
}

// SetNillableInstanceID sets the "instance" edge to the InstanceResource entity by ID if the given value is not nil.
func (tpc *TelemetryProfileCreate) SetNillableInstanceID(id *int) *TelemetryProfileCreate {
	if id != nil {
		tpc = tpc.SetInstanceID(*id)
	}
	return tpc
}

// SetInstance sets the "instance" edge to the InstanceResource entity.
func (tpc *TelemetryProfileCreate) SetInstance(i *InstanceResource) *TelemetryProfileCreate {
	return tpc.SetInstanceID(i.ID)
}

// SetGroupID sets the "group" edge to the TelemetryGroupResource entity by ID.
func (tpc *TelemetryProfileCreate) SetGroupID(id int) *TelemetryProfileCreate {
	tpc.mutation.SetGroupID(id)
	return tpc
}

// SetGroup sets the "group" edge to the TelemetryGroupResource entity.
func (tpc *TelemetryProfileCreate) SetGroup(t *TelemetryGroupResource) *TelemetryProfileCreate {
	return tpc.SetGroupID(t.ID)
}

// Mutation returns the TelemetryProfileMutation object of the builder.
func (tpc *TelemetryProfileCreate) Mutation() *TelemetryProfileMutation {
	return tpc.mutation
}

// Save creates the TelemetryProfile in the database.
func (tpc *TelemetryProfileCreate) Save(ctx context.Context) (*TelemetryProfile, error) {
	return withHooks(ctx, tpc.sqlSave, tpc.mutation, tpc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (tpc *TelemetryProfileCreate) SaveX(ctx context.Context) *TelemetryProfile {
	v, err := tpc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tpc *TelemetryProfileCreate) Exec(ctx context.Context) error {
	_, err := tpc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tpc *TelemetryProfileCreate) ExecX(ctx context.Context) {
	if err := tpc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (tpc *TelemetryProfileCreate) check() error {
	if _, ok := tpc.mutation.ResourceID(); !ok {
		return &ValidationError{Name: "resource_id", err: errors.New(`ent: missing required field "TelemetryProfile.resource_id"`)}
	}
	if _, ok := tpc.mutation.Kind(); !ok {
		return &ValidationError{Name: "kind", err: errors.New(`ent: missing required field "TelemetryProfile.kind"`)}
	}
	if v, ok := tpc.mutation.Kind(); ok {
		if err := telemetryprofile.KindValidator(v); err != nil {
			return &ValidationError{Name: "kind", err: fmt.Errorf(`ent: validator failed for field "TelemetryProfile.kind": %w`, err)}
		}
	}
	if v, ok := tpc.mutation.LogLevel(); ok {
		if err := telemetryprofile.LogLevelValidator(v); err != nil {
			return &ValidationError{Name: "log_level", err: fmt.Errorf(`ent: validator failed for field "TelemetryProfile.log_level": %w`, err)}
		}
	}
	if _, ok := tpc.mutation.TenantID(); !ok {
		return &ValidationError{Name: "tenant_id", err: errors.New(`ent: missing required field "TelemetryProfile.tenant_id"`)}
	}
	if _, ok := tpc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "TelemetryProfile.created_at"`)}
	}
	if _, ok := tpc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "TelemetryProfile.updated_at"`)}
	}
	if len(tpc.mutation.GroupIDs()) == 0 {
		return &ValidationError{Name: "group", err: errors.New(`ent: missing required edge "TelemetryProfile.group"`)}
	}
	return nil
}

func (tpc *TelemetryProfileCreate) sqlSave(ctx context.Context) (*TelemetryProfile, error) {
	if err := tpc.check(); err != nil {
		return nil, err
	}
	_node, _spec := tpc.createSpec()
	if err := sqlgraph.CreateNode(ctx, tpc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	tpc.mutation.id = &_node.ID
	tpc.mutation.done = true
	return _node, nil
}

func (tpc *TelemetryProfileCreate) createSpec() (*TelemetryProfile, *sqlgraph.CreateSpec) {
	var (
		_node = &TelemetryProfile{config: tpc.config}
		_spec = sqlgraph.NewCreateSpec(telemetryprofile.Table, sqlgraph.NewFieldSpec(telemetryprofile.FieldID, field.TypeInt))
	)
	if value, ok := tpc.mutation.ResourceID(); ok {
		_spec.SetField(telemetryprofile.FieldResourceID, field.TypeString, value)
		_node.ResourceID = value
	}
	if value, ok := tpc.mutation.Kind(); ok {
		_spec.SetField(telemetryprofile.FieldKind, field.TypeEnum, value)
		_node.Kind = value
	}
	if value, ok := tpc.mutation.MetricsInterval(); ok {
		_spec.SetField(telemetryprofile.FieldMetricsInterval, field.TypeUint32, value)
		_node.MetricsInterval = value
	}
	if value, ok := tpc.mutation.LogLevel(); ok {
		_spec.SetField(telemetryprofile.FieldLogLevel, field.TypeEnum, value)
		_node.LogLevel = value
	}
	if value, ok := tpc.mutation.TenantID(); ok {
		_spec.SetField(telemetryprofile.FieldTenantID, field.TypeString, value)
		_node.TenantID = value
	}
	if value, ok := tpc.mutation.CreatedAt(); ok {
		_spec.SetField(telemetryprofile.FieldCreatedAt, field.TypeString, value)
		_node.CreatedAt = value
	}
	if value, ok := tpc.mutation.UpdatedAt(); ok {
		_spec.SetField(telemetryprofile.FieldUpdatedAt, field.TypeString, value)
		_node.UpdatedAt = value
	}
	if nodes := tpc.mutation.RegionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   telemetryprofile.RegionTable,
			Columns: []string{telemetryprofile.RegionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(regionresource.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.telemetry_profile_region = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := tpc.mutation.SiteIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   telemetryprofile.SiteTable,
			Columns: []string{telemetryprofile.SiteColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(siteresource.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.telemetry_profile_site = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := tpc.mutation.InstanceIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   telemetryprofile.InstanceTable,
			Columns: []string{telemetryprofile.InstanceColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(instanceresource.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.telemetry_profile_instance = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := tpc.mutation.GroupIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   telemetryprofile.GroupTable,
			Columns: []string{telemetryprofile.GroupColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(telemetrygroupresource.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.telemetry_profile_group = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// TelemetryProfileCreateBulk is the builder for creating many TelemetryProfile entities in bulk.
type TelemetryProfileCreateBulk struct {
	config
	err      error
	builders []*TelemetryProfileCreate
}

// Save creates the TelemetryProfile entities in the database.
func (tpcb *TelemetryProfileCreateBulk) Save(ctx context.Context) ([]*TelemetryProfile, error) {
	if tpcb.err != nil {
		return nil, tpcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(tpcb.builders))
	nodes := make([]*TelemetryProfile, len(tpcb.builders))
	mutators := make([]Mutator, len(tpcb.builders))
	for i := range tpcb.builders {
		func(i int, root context.Context) {
			builder := tpcb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*TelemetryProfileMutation)
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
					_, err = mutators[i+1].Mutate(root, tpcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, tpcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, tpcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (tpcb *TelemetryProfileCreateBulk) SaveX(ctx context.Context) []*TelemetryProfile {
	v, err := tpcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tpcb *TelemetryProfileCreateBulk) Exec(ctx context.Context) error {
	_, err := tpcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tpcb *TelemetryProfileCreateBulk) ExecX(ctx context.Context) {
	if err := tpcb.Exec(ctx); err != nil {
		panic(err)
	}
}

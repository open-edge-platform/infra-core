// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/ouresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/providerresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/regionresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/siteresource"
)

// SiteResourceCreate is the builder for creating a SiteResource entity.
type SiteResourceCreate struct {
	config
	mutation *SiteResourceMutation
	hooks    []Hook
}

// SetResourceID sets the "resource_id" field.
func (src *SiteResourceCreate) SetResourceID(s string) *SiteResourceCreate {
	src.mutation.SetResourceID(s)
	return src
}

// SetName sets the "name" field.
func (src *SiteResourceCreate) SetName(s string) *SiteResourceCreate {
	src.mutation.SetName(s)
	return src
}

// SetNillableName sets the "name" field if the given value is not nil.
func (src *SiteResourceCreate) SetNillableName(s *string) *SiteResourceCreate {
	if s != nil {
		src.SetName(*s)
	}
	return src
}

// SetAddress sets the "address" field.
func (src *SiteResourceCreate) SetAddress(s string) *SiteResourceCreate {
	src.mutation.SetAddress(s)
	return src
}

// SetNillableAddress sets the "address" field if the given value is not nil.
func (src *SiteResourceCreate) SetNillableAddress(s *string) *SiteResourceCreate {
	if s != nil {
		src.SetAddress(*s)
	}
	return src
}

// SetSiteLat sets the "site_lat" field.
func (src *SiteResourceCreate) SetSiteLat(i int32) *SiteResourceCreate {
	src.mutation.SetSiteLat(i)
	return src
}

// SetNillableSiteLat sets the "site_lat" field if the given value is not nil.
func (src *SiteResourceCreate) SetNillableSiteLat(i *int32) *SiteResourceCreate {
	if i != nil {
		src.SetSiteLat(*i)
	}
	return src
}

// SetSiteLng sets the "site_lng" field.
func (src *SiteResourceCreate) SetSiteLng(i int32) *SiteResourceCreate {
	src.mutation.SetSiteLng(i)
	return src
}

// SetNillableSiteLng sets the "site_lng" field if the given value is not nil.
func (src *SiteResourceCreate) SetNillableSiteLng(i *int32) *SiteResourceCreate {
	if i != nil {
		src.SetSiteLng(*i)
	}
	return src
}

// SetDNSServers sets the "dns_servers" field.
func (src *SiteResourceCreate) SetDNSServers(s string) *SiteResourceCreate {
	src.mutation.SetDNSServers(s)
	return src
}

// SetNillableDNSServers sets the "dns_servers" field if the given value is not nil.
func (src *SiteResourceCreate) SetNillableDNSServers(s *string) *SiteResourceCreate {
	if s != nil {
		src.SetDNSServers(*s)
	}
	return src
}

// SetDockerRegistries sets the "docker_registries" field.
func (src *SiteResourceCreate) SetDockerRegistries(s string) *SiteResourceCreate {
	src.mutation.SetDockerRegistries(s)
	return src
}

// SetNillableDockerRegistries sets the "docker_registries" field if the given value is not nil.
func (src *SiteResourceCreate) SetNillableDockerRegistries(s *string) *SiteResourceCreate {
	if s != nil {
		src.SetDockerRegistries(*s)
	}
	return src
}

// SetMetricsEndpoint sets the "metrics_endpoint" field.
func (src *SiteResourceCreate) SetMetricsEndpoint(s string) *SiteResourceCreate {
	src.mutation.SetMetricsEndpoint(s)
	return src
}

// SetNillableMetricsEndpoint sets the "metrics_endpoint" field if the given value is not nil.
func (src *SiteResourceCreate) SetNillableMetricsEndpoint(s *string) *SiteResourceCreate {
	if s != nil {
		src.SetMetricsEndpoint(*s)
	}
	return src
}

// SetHTTPProxy sets the "http_proxy" field.
func (src *SiteResourceCreate) SetHTTPProxy(s string) *SiteResourceCreate {
	src.mutation.SetHTTPProxy(s)
	return src
}

// SetNillableHTTPProxy sets the "http_proxy" field if the given value is not nil.
func (src *SiteResourceCreate) SetNillableHTTPProxy(s *string) *SiteResourceCreate {
	if s != nil {
		src.SetHTTPProxy(*s)
	}
	return src
}

// SetHTTPSProxy sets the "https_proxy" field.
func (src *SiteResourceCreate) SetHTTPSProxy(s string) *SiteResourceCreate {
	src.mutation.SetHTTPSProxy(s)
	return src
}

// SetNillableHTTPSProxy sets the "https_proxy" field if the given value is not nil.
func (src *SiteResourceCreate) SetNillableHTTPSProxy(s *string) *SiteResourceCreate {
	if s != nil {
		src.SetHTTPSProxy(*s)
	}
	return src
}

// SetFtpProxy sets the "ftp_proxy" field.
func (src *SiteResourceCreate) SetFtpProxy(s string) *SiteResourceCreate {
	src.mutation.SetFtpProxy(s)
	return src
}

// SetNillableFtpProxy sets the "ftp_proxy" field if the given value is not nil.
func (src *SiteResourceCreate) SetNillableFtpProxy(s *string) *SiteResourceCreate {
	if s != nil {
		src.SetFtpProxy(*s)
	}
	return src
}

// SetNoProxy sets the "no_proxy" field.
func (src *SiteResourceCreate) SetNoProxy(s string) *SiteResourceCreate {
	src.mutation.SetNoProxy(s)
	return src
}

// SetNillableNoProxy sets the "no_proxy" field if the given value is not nil.
func (src *SiteResourceCreate) SetNillableNoProxy(s *string) *SiteResourceCreate {
	if s != nil {
		src.SetNoProxy(*s)
	}
	return src
}

// SetMetadata sets the "metadata" field.
func (src *SiteResourceCreate) SetMetadata(s string) *SiteResourceCreate {
	src.mutation.SetMetadata(s)
	return src
}

// SetNillableMetadata sets the "metadata" field if the given value is not nil.
func (src *SiteResourceCreate) SetNillableMetadata(s *string) *SiteResourceCreate {
	if s != nil {
		src.SetMetadata(*s)
	}
	return src
}

// SetTenantID sets the "tenant_id" field.
func (src *SiteResourceCreate) SetTenantID(s string) *SiteResourceCreate {
	src.mutation.SetTenantID(s)
	return src
}

// SetCreatedAt sets the "created_at" field.
func (src *SiteResourceCreate) SetCreatedAt(s string) *SiteResourceCreate {
	src.mutation.SetCreatedAt(s)
	return src
}

// SetUpdatedAt sets the "updated_at" field.
func (src *SiteResourceCreate) SetUpdatedAt(s string) *SiteResourceCreate {
	src.mutation.SetUpdatedAt(s)
	return src
}

// SetRegionID sets the "region" edge to the RegionResource entity by ID.
func (src *SiteResourceCreate) SetRegionID(id int) *SiteResourceCreate {
	src.mutation.SetRegionID(id)
	return src
}

// SetNillableRegionID sets the "region" edge to the RegionResource entity by ID if the given value is not nil.
func (src *SiteResourceCreate) SetNillableRegionID(id *int) *SiteResourceCreate {
	if id != nil {
		src = src.SetRegionID(*id)
	}
	return src
}

// SetRegion sets the "region" edge to the RegionResource entity.
func (src *SiteResourceCreate) SetRegion(r *RegionResource) *SiteResourceCreate {
	return src.SetRegionID(r.ID)
}

// SetOuID sets the "ou" edge to the OuResource entity by ID.
func (src *SiteResourceCreate) SetOuID(id int) *SiteResourceCreate {
	src.mutation.SetOuID(id)
	return src
}

// SetNillableOuID sets the "ou" edge to the OuResource entity by ID if the given value is not nil.
func (src *SiteResourceCreate) SetNillableOuID(id *int) *SiteResourceCreate {
	if id != nil {
		src = src.SetOuID(*id)
	}
	return src
}

// SetOu sets the "ou" edge to the OuResource entity.
func (src *SiteResourceCreate) SetOu(o *OuResource) *SiteResourceCreate {
	return src.SetOuID(o.ID)
}

// SetProviderID sets the "provider" edge to the ProviderResource entity by ID.
func (src *SiteResourceCreate) SetProviderID(id int) *SiteResourceCreate {
	src.mutation.SetProviderID(id)
	return src
}

// SetNillableProviderID sets the "provider" edge to the ProviderResource entity by ID if the given value is not nil.
func (src *SiteResourceCreate) SetNillableProviderID(id *int) *SiteResourceCreate {
	if id != nil {
		src = src.SetProviderID(*id)
	}
	return src
}

// SetProvider sets the "provider" edge to the ProviderResource entity.
func (src *SiteResourceCreate) SetProvider(p *ProviderResource) *SiteResourceCreate {
	return src.SetProviderID(p.ID)
}

// Mutation returns the SiteResourceMutation object of the builder.
func (src *SiteResourceCreate) Mutation() *SiteResourceMutation {
	return src.mutation
}

// Save creates the SiteResource in the database.
func (src *SiteResourceCreate) Save(ctx context.Context) (*SiteResource, error) {
	return withHooks(ctx, src.sqlSave, src.mutation, src.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (src *SiteResourceCreate) SaveX(ctx context.Context) *SiteResource {
	v, err := src.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (src *SiteResourceCreate) Exec(ctx context.Context) error {
	_, err := src.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (src *SiteResourceCreate) ExecX(ctx context.Context) {
	if err := src.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (src *SiteResourceCreate) check() error {
	if _, ok := src.mutation.ResourceID(); !ok {
		return &ValidationError{Name: "resource_id", err: errors.New(`ent: missing required field "SiteResource.resource_id"`)}
	}
	if _, ok := src.mutation.TenantID(); !ok {
		return &ValidationError{Name: "tenant_id", err: errors.New(`ent: missing required field "SiteResource.tenant_id"`)}
	}
	if _, ok := src.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "SiteResource.created_at"`)}
	}
	if _, ok := src.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "SiteResource.updated_at"`)}
	}
	return nil
}

func (src *SiteResourceCreate) sqlSave(ctx context.Context) (*SiteResource, error) {
	if err := src.check(); err != nil {
		return nil, err
	}
	_node, _spec := src.createSpec()
	if err := sqlgraph.CreateNode(ctx, src.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	src.mutation.id = &_node.ID
	src.mutation.done = true
	return _node, nil
}

func (src *SiteResourceCreate) createSpec() (*SiteResource, *sqlgraph.CreateSpec) {
	var (
		_node = &SiteResource{config: src.config}
		_spec = sqlgraph.NewCreateSpec(siteresource.Table, sqlgraph.NewFieldSpec(siteresource.FieldID, field.TypeInt))
	)
	if value, ok := src.mutation.ResourceID(); ok {
		_spec.SetField(siteresource.FieldResourceID, field.TypeString, value)
		_node.ResourceID = value
	}
	if value, ok := src.mutation.Name(); ok {
		_spec.SetField(siteresource.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := src.mutation.Address(); ok {
		_spec.SetField(siteresource.FieldAddress, field.TypeString, value)
		_node.Address = value
	}
	if value, ok := src.mutation.SiteLat(); ok {
		_spec.SetField(siteresource.FieldSiteLat, field.TypeInt32, value)
		_node.SiteLat = value
	}
	if value, ok := src.mutation.SiteLng(); ok {
		_spec.SetField(siteresource.FieldSiteLng, field.TypeInt32, value)
		_node.SiteLng = value
	}
	if value, ok := src.mutation.DNSServers(); ok {
		_spec.SetField(siteresource.FieldDNSServers, field.TypeString, value)
		_node.DNSServers = value
	}
	if value, ok := src.mutation.DockerRegistries(); ok {
		_spec.SetField(siteresource.FieldDockerRegistries, field.TypeString, value)
		_node.DockerRegistries = value
	}
	if value, ok := src.mutation.MetricsEndpoint(); ok {
		_spec.SetField(siteresource.FieldMetricsEndpoint, field.TypeString, value)
		_node.MetricsEndpoint = value
	}
	if value, ok := src.mutation.HTTPProxy(); ok {
		_spec.SetField(siteresource.FieldHTTPProxy, field.TypeString, value)
		_node.HTTPProxy = value
	}
	if value, ok := src.mutation.HTTPSProxy(); ok {
		_spec.SetField(siteresource.FieldHTTPSProxy, field.TypeString, value)
		_node.HTTPSProxy = value
	}
	if value, ok := src.mutation.FtpProxy(); ok {
		_spec.SetField(siteresource.FieldFtpProxy, field.TypeString, value)
		_node.FtpProxy = value
	}
	if value, ok := src.mutation.NoProxy(); ok {
		_spec.SetField(siteresource.FieldNoProxy, field.TypeString, value)
		_node.NoProxy = value
	}
	if value, ok := src.mutation.Metadata(); ok {
		_spec.SetField(siteresource.FieldMetadata, field.TypeString, value)
		_node.Metadata = value
	}
	if value, ok := src.mutation.TenantID(); ok {
		_spec.SetField(siteresource.FieldTenantID, field.TypeString, value)
		_node.TenantID = value
	}
	if value, ok := src.mutation.CreatedAt(); ok {
		_spec.SetField(siteresource.FieldCreatedAt, field.TypeString, value)
		_node.CreatedAt = value
	}
	if value, ok := src.mutation.UpdatedAt(); ok {
		_spec.SetField(siteresource.FieldUpdatedAt, field.TypeString, value)
		_node.UpdatedAt = value
	}
	if nodes := src.mutation.RegionIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   siteresource.RegionTable,
			Columns: []string{siteresource.RegionColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(regionresource.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.site_resource_region = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := src.mutation.OuIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   siteresource.OuTable,
			Columns: []string{siteresource.OuColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(ouresource.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.site_resource_ou = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := src.mutation.ProviderIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   siteresource.ProviderTable,
			Columns: []string{siteresource.ProviderColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(providerresource.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.site_resource_provider = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// SiteResourceCreateBulk is the builder for creating many SiteResource entities in bulk.
type SiteResourceCreateBulk struct {
	config
	err      error
	builders []*SiteResourceCreate
}

// Save creates the SiteResource entities in the database.
func (srcb *SiteResourceCreateBulk) Save(ctx context.Context) ([]*SiteResource, error) {
	if srcb.err != nil {
		return nil, srcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(srcb.builders))
	nodes := make([]*SiteResource, len(srcb.builders))
	mutators := make([]Mutator, len(srcb.builders))
	for i := range srcb.builders {
		func(i int, root context.Context) {
			builder := srcb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*SiteResourceMutation)
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
					_, err = mutators[i+1].Mutate(root, srcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, srcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, srcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (srcb *SiteResourceCreateBulk) SaveX(ctx context.Context) []*SiteResource {
	v, err := srcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (srcb *SiteResourceCreateBulk) Exec(ctx context.Context) error {
	_, err := srcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (srcb *SiteResourceCreateBulk) ExecX(ctx context.Context) {
	if err := srcb.Exec(ctx); err != nil {
		panic(err)
	}
}

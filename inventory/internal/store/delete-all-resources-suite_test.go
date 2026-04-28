// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store_test

import (
	"context"
	"fmt"
	"time"

	"github.com/stretchr/testify/suite"

	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	inv_util "github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/filters"
)

type hardDeleteAllResourcesSuite struct {
	suite.Suite
	createModel  func(*inv_testing.InvResourceDAO) (tenantId string, noOfResources int)
	resourceKind inv_v1.ResourceKind
}

func (ts *hardDeleteAllResourcesSuite) Test() {
	dao := inv_testing.NewInvResourceDAOOrFail(ts.T())
	t1, noOfResources := ts.createModel(dao)
	t2, _ := ts.createModel(dao)

	allTenants := []string{t1, t2}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Minute)
	defer cancel()

	resource, err := inv_util.GetResourceFromKind(ts.resourceKind)
	ts.Require().NoError(err)

	createFilter := func(tid string) *inv_v1.ResourceFilter {
		return &inv_v1.ResourceFilter{
			Resource: resource,
			Filter:   filters.NewBuilderWith(filters.ValEq("tenant_id", tid)).Build(),
		}
	}

	for _, tid := range allTenants {
		ts.Run("Read Resources for TenantID="+tid, func() {
			eps, err := dao.GetAPIClient().FindAll(ctx, createFilter(tid))
			ts.Require().NoError(err)
			ts.Require().Len(eps, noOfResources)
		})
	}

	ts.Run(fmt.Sprintf("DeleteAllResources(tenantID=%s, rk=%s)", allTenants[0], ts.resourceKind), func() {
		ts.Require().NoError(dao.DeleteAllResources(ts.T(), ctx, allTenants[0], ts.resourceKind, true))
	})

	ts.Run(fmt.Sprintf("FindAll(tenantID=%s, rk=%s)", allTenants[0], ts.resourceKind), func() {
		eps, err := dao.GetAPIClient().FindAll(ctx, createFilter(allTenants[0]))
		ts.Require().NoError(err)
		ts.Require().Empty(eps)
	})

	ts.Run(fmt.Sprintf("FindAll(tenantID=%s, rk=%s)", allTenants[1], ts.resourceKind), func() {
		eps, err := dao.GetAPIClient().FindAll(ctx, createFilter(allTenants[1]))
		ts.Require().NoError(err)
		ts.Require().NotEmpty(eps)
	})

	ts.Run(fmt.Sprintf("DeleteAllResources(tenantID=%s, rk=%s)", allTenants[1], ts.resourceKind), func() {
		ts.Require().NoError(dao.DeleteAllResources(ts.T(), ctx, allTenants[1], ts.resourceKind, true))
	})

	ts.Run(fmt.Sprintf("FindAll(tenantID=%s, rk=%s)", allTenants[1], ts.resourceKind), func() {
		eps, err := dao.GetAPIClient().FindAll(ctx, createFilter(allTenants[1]))
		ts.Require().NoError(err)
		ts.Require().Empty(eps)
	})
}

type softDeleteAllResourcesSuite struct {
	suite.Suite
	createModel      func(*inv_testing.InvResourceDAO) (tenantId string, noOfResources int)
	resourceKind     inv_v1.ResourceKind
	deletedClause    filters.Clause
	notDeletedClause filters.Clause
}

func (ts *softDeleteAllResourcesSuite) Test() {
	dao := inv_testing.NewInvResourceDAOOrFail(ts.T())
	t1, noOfResources := ts.createModel(dao)
	t2, _ := ts.createModel(dao)

	allTenants := []string{t1, t2}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Minute)
	defer cancel()

	resource, err := inv_util.GetResourceFromKind(ts.resourceKind)
	ts.Require().NoError(err)

	createFilter := func(tid string) *inv_v1.ResourceFilter {
		return &inv_v1.ResourceFilter{
			Resource: resource,
			Filter:   filters.NewBuilderWith(filters.ValEq("tenant_id", tid)).Build(),
		}
	}

	// each tenant shall own resources
	for _, tid := range allTenants {
		ts.Run("Read Resources for TenantID="+tid, func() {
			eps, err := dao.GetAPIClient().FindAll(ctx, createFilter(tid))
			ts.Require().NoError(err)
			ts.Require().Len(eps, noOfResources)
		})
	}

	// soft delete resources of first tenant
	ts.Run(fmt.Sprintf("Soft Delete Resources tenantID=%s, rk=%s", allTenants[0], ts.resourceKind), func() {
		ts.Require().NoError(dao.DeleteAllResources(ts.T(), ctx, allTenants[0], ts.resourceKind, false))
	})

	// resources of first tenant shall be: desired_state eq deleted
	ts.Run(fmt.Sprintf("Check soft deleted for tenantID=%s", allTenants[0]), func() {
		eps, err := dao.GetAPIClient().FindAll(ctx, &inv_v1.ResourceFilter{
			Resource: resource,
			Filter: filters.NewBuilderWith(filters.ValEq("tenant_id", allTenants[0])).
				And(ts.deletedClause).
				Build(),
		})
		ts.Require().NoError(err)
		ts.Require().Len(eps, noOfResources)
	})

	// resources of second tenant shall be desired_state not eq deleted
	ts.Run(fmt.Sprintf("Check not soft deleted for tenantID=%s", allTenants[1]), func() {
		eps, err := dao.GetAPIClient().FindAll(ctx, &inv_v1.ResourceFilter{
			Resource: resource,
			Filter: filters.NewBuilderWith(filters.ValEq("tenant_id", allTenants[1])).
				And(ts.notDeletedClause).
				Build(),
		})
		ts.Require().NoError(err)
		ts.Require().Len(eps, noOfResources)
	})
}

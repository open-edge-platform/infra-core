// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store_test

import (
	"context"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
)

var tenants = []string{
	"11111111-1111-1111-1111-111111111112",
	"22222222-2222-2222-2222-222222222221",
}

type resDesc struct {
	resourceID, tenantID string
	resource             *inv_v1.Resource
}

type mt struct {
	suite.Suite
	apiClient           client.TenantAwareInventoryClient
	rmClient            client.TenantAwareInventoryClient
	createResource      func(tenantID string) (resourceId string, res *inv_v1.Resource)
	resources           []*resDesc
	updateReqClientType inv_testing.ClientType
}

func (mts *mt) SetupSuite() {
	mts.apiClient = inv_testing.GetClient(mts.T(), inv_testing.APIClient).GetTenantAwareInventoryClient()
	mts.rmClient = inv_testing.GetClient(mts.T(), inv_testing.RMClient).GetTenantAwareInventoryClient()
	if mts.updateReqClientType == "" {
		mts.updateReqClientType = inv_testing.APIClient
	}
	collections.ForEach(tenants, func(tenantID string) {
		id, res := mts.createResource(tenantID)
		mts.resources = append(mts.resources, &resDesc{
			resourceID: id,
			tenantID:   tenantID,
			resource:   res,
		})
	})
}

// TestListAll confirms cross tenant capabilities of ListAll API.
func (mts *mt) TestListAll() {
	mts.Require().NotEmpty(mts.resources)
	anyResource := mts.resources[0]
	zeroResource := mts.zero(anyResource.resource)
	all, err := mts.apiClient.ListAll(context.TODO(), &inv_v1.ResourceFilter{Resource: zeroResource})
	mts.Require().NoError(err)

	returnedTenants := collections.MapSlice[*inv_v1.Resource, string](all, func(resource *inv_v1.Resource) string {
		pm, uerr := util.UnwrapResource[proto.Message](resource)
		mts.Require().NoError(uerr)
		carrier, ok := pm.(interface{ GetTenantId() string })
		mts.Require().True(ok)
		return carrier.GetTenantId()
	})

	affectedTenants := collections.MapSlice[*inv_v1.Resource, string](
		collections.MapSlice(mts.resources, func(rd *resDesc) *inv_v1.Resource { return rd.resource }),
		func(resource *inv_v1.Resource) string {
			pm, uerr := util.UnwrapResource[proto.Message](resource)
			mts.Require().NoError(uerr)
			carrier, ok := pm.(interface{ GetTenantId() string })
			mts.Require().True(ok)
			return carrier.GetTenantId()
		})
	mts.Require().NotEmpty(returnedTenants)
	mts.Require().ElementsMatch(returnedTenants, affectedTenants)
}

// TestResourceCreationWithTenantMismatch confirms resource creation is rejected when:
// request.tenantID != request.resource.tenantID.
func (mts *mt) TestResourceCreationWithTenantMismatch() {
	mts.Require().NotEmpty(mts.resources)
	anyResource := mts.resources[0]
	createReq := mts.zero(anyResource.resource)

	_, err := mts.apiClient.Create(context.TODO(), tenantIDZero, createReq)
	mts.Require().Error(err, "error shall be thrown when req.tenantID != req.resource.TenantID")
	mts.Require().True(errors.IsInvalidArgument(err))
}

// TestGetOwnedByOtherTenant checks if resource owned by one tenant can be got by another tenant.
func (mts *mt) TestGetOwnedByOtherTenant() {
	mts.Require().NotEmpty(mts.resources)
	sut := mts.resources[0]
	getResp1, err := mts.apiClient.Get(context.TODO(), sut.tenantID, sut.resourceID)
	mts.Require().NoError(err)
	mts.Require().NotNil(getResp1)
	getResp2, err := mts.apiClient.Get(context.TODO(), "anotherTenant", sut.resourceID)
	mts.Require().Error(err)
	mts.Require().Equal(codes.NotFound, status.Code(err), "not found error expected")
	mts.Require().Nil(getResp2)
}

// TestUpdateResourceOwnedByOtherTenant checks if resource owned by one tenant can be updated by another tenant.
func (mts *mt) TestUpdateResourceOwnedByOtherTenant() {
	mts.Require().NotEmpty(mts.resources)
	// to be updated
	anyResource := mts.resources[0]

	if host, ok := anyResource.resource.Resource.(*inv_v1.Resource_Host); ok {
		// Host Update with UUID & SN & AmtSku fields is not allowed - remove fields from Host resource
		host.Host.Uuid = ""
		host.Host.SerialNumber = ""
		host.Host.AmtSku = ""
		anyResource.resource = &inv_v1.Resource{Resource: host}
	}

	updateReqResp, err := mts.clientByType(mts.updateReqClientType).Update(
		context.TODO(),
		tenantIDOne,
		anyResource.resourceID,
		&fieldmaskpb.FieldMask{Paths: []string{}}, anyResource.resource)

	mts.Require().Error(err, "tenant T2 cannot update data owned by tenant T1")
	mts.Require().Equal(codes.NotFound, status.Code(err), "not found error expected")
	mts.Require().Nil(updateReqResp)
}

// TestTenantIDUpdateIsNotAllowed checks if tenantID field can be updated for requested resource.
func (mts *mt) TestTenantIDUpdateIsNotAllowed() {
	mts.Require().NotEmpty(mts.resources)
	anyResource := mts.resources[0]
	updateReq := anyResource.resource

	_, err := mts.apiClient.Update(
		context.TODO(),
		anyResource.tenantID,
		anyResource.resourceID,
		&fieldmaskpb.FieldMask{Paths: []string{"tenant_id"}}, updateReq)

	mts.Require().Error(err, "error shall be thrown when updateRequest.resource.tenantID != ''")
	mts.Require().Equalf(codes.InvalidArgument, status.Code(err),
		"invalidArgument error is expected whilst returned is: %v", err)
	mts.Require().ErrorContains(err, "tenant update is not allowed")
}

// TestDeleteOwnedByOtherTenant checks if resource owned by one tenant can be deleted by another tenant.
func (mts *mt) TestDeleteOwnedByOtherTenant() {
	for _, rd := range mts.resources {
		deletionReqResponse, err := mts.apiClient.Delete(context.TODO(), tenantIDZero, rd.resourceID)
		mts.Require().Error(err, "tenant T2 cannot update data owned by tenant T1")
		mts.Require().Contains([]codes.Code{codes.NotFound, codes.Unimplemented}, status.Code(err),
			"not found or unimplemented error expected: %v", err)
		mts.Require().Nil(deletionReqResponse)
	}
}

func (mts *mt) zero(r *inv_v1.Resource) *inv_v1.Resource {
	rk := util.GetResourceKindFromResource(r)
	zeroResource, err := util.GetResourceFromKind(rk)
	mts.Require().NoError(err)
	return zeroResource
}

func (mts *mt) clientByType(ct inv_testing.ClientType) client.TenantAwareInventoryClient {
	switch ct {
	case inv_testing.RMClient:
		return mts.rmClient
	default:
		return mts.apiClient
	}
}

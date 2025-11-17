// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package invhandlers_test

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/api/internal/types"
	"github.com/open-edge-platform/infra-core/api/internal/worker/clients"
	"github.com/open-edge-platform/infra-core/api/internal/worker/handlers"
	inv_handlers "github.com/open-edge-platform/infra-core/api/internal/worker/handlers/invhandlers"
	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	"github.com/open-edge-platform/infra-core/api/test/utils"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	statusv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/status/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
)

var (
	enableQueryDetail = true
	offset1           = 0
	pageSize1         = 1
	offset2           = 1
	pageSize2         = 2
)

func BuildFmFromHostRequest(body *api.Host) []string {
	fm := []string{}
	fm = append(fm, "name")
	if body.SerialNumber != nil {
		fm = append(fm, "serial_number")
	}
	if body.Site != nil {
		fm = append(fm, "site")
	}
	if body.Uuid != nil {
		fm = append(fm, "uuid")
	}
	if body.Metadata != nil {
		fm = append(fm, "metadata")
	}
	if body.DesiredPowerState != nil {
		fm = append(fm, "desired_power_state")
	}
	return fm
}

func BuildFmFromHostRegisterRequest(body *api.HostRegisterInfo) []string {
	fm := []string{}
	if body.Name != nil {
		fm = append(fm, "name")
	}
	if body.SerialNumber != nil {
		fm = append(fm, "serial_number")
	}
	if body.Uuid != nil {
		fm = append(fm, "uuid")
	}
	if body.AutoOnboard != nil {
		fm = append(fm, "desired_state")
	}
	return fm
}

func validateFM(actual, expected *fieldmaskpb.FieldMask) error {
	var err error
	if actual != nil {
		actual.Normalize()
		expected.Normalize()
		if !proto.Equal(expected, actual) {
			err = fmt.Errorf(
				"FieldMask is incorrectly constructed, expected: %s got: %s",
				expected.Paths,
				actual.Paths,
			)
		}
	} else {
		err = fmt.Errorf("no request in Mock Inventory")
	}
	return err
}

// check that we pass the expected filters to the inventory.
//
//nolint:funlen // it is a test
func Test_hostHandler_list(t *testing.T) {
	//nolint:goconst // it is a test
	SiteID := "site-12345678"
	//nolint:goconst // it is a test
	instanceID := "inst-12345678"
	UUID := "BFD3B398-9A4B-480D-AB53-4050ED108F5F"
	Metadata := []string{"key1=value1", "key2=value2"}
	hasSiteFilter := "has(Site)"
	bySiteFilter := `site.resource_id = "site-12345678"`
	byInstanceFilter := `instance.resource_id = "inst-12345678"`
	byUUIDFilter := `uuid = "BFD3B398-9A4B-480D-AB53-4050ED108F5F"`
	OrderBy := "ResourceId"

	type args struct {
		query api.GetComputeHostsParams
	}

	type want struct {
		callCount   int
		SiteID      string
		InstanceID  string
		UUID        string
		callLimit   uint32
		callOffset  uint32
		callOrderBy string
		callFilter  string
	}

	// NOTE that enabling details fetch the IPAddresses associated with the NICs
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"list-all",
			args{query: api.GetComputeHostsParams{Detail: &enableQueryDetail}},
			want{
				callCount: 3,
				callLimit: 20,
			},
		},
		{
			"list-all-no-detail",
			args{query: api.GetComputeHostsParams{}},
			want{
				callCount: 1,
				callLimit: 20,
			},
		},
		{
			"list-by-site",
			args{query: api.GetComputeHostsParams{SiteID: &SiteID, Detail: &enableQueryDetail}},
			want{
				callCount:  3,
				callLimit:  20,
				callFilter: bySiteFilter,
				SiteID:     SiteID,
			},
		},
		{
			"list-by-instance",
			args{query: api.GetComputeHostsParams{InstanceID: &instanceID, Detail: &enableQueryDetail}},
			want{
				callCount:  3,
				callLimit:  20,
				callFilter: byInstanceFilter,
				InstanceID: instanceID,
			},
		},
		{
			"list-by-device-guid",
			args{query: api.GetComputeHostsParams{Uuid: &UUID, Detail: &enableQueryDetail}},
			want{
				callCount:  3,
				callLimit:  20,
				callFilter: byUUIDFilter,
				UUID:       UUID,
			},
		},
		{
			"list-by-metadata",
			args{query: api.GetComputeHostsParams{Metadata: &Metadata, Detail: &enableQueryDetail}},
			want{
				callCount: 3,
				callLimit: 20,
			},
		},
		// All these calls are without detail to avoid interferences of the hidden calls
		{
			"list-all-no-detail-idx1-siz1",
			args{query: api.GetComputeHostsParams{Offset: &offset1, PageSize: &pageSize1}},
			want{
				callCount:  1,
				callLimit:  1,
				callOffset: 0,
			},
		},
		{
			"list-all-no-detail-idx1-siz2",
			args{query: api.GetComputeHostsParams{Offset: &offset1, PageSize: &pageSize2}},
			want{
				callCount:  1,
				callLimit:  2,
				callOffset: 0,
			},
		},
		{
			"list-all-no-detail-idx2-siz2",
			args{query: api.GetComputeHostsParams{Offset: &offset2, PageSize: &pageSize2}},
			want{
				callCount:  1,
				callLimit:  2,
				callOffset: 1,
			},
		},
		{
			"list-with-order-by",
			args{query: api.GetComputeHostsParams{OrderBy: &OrderBy}},
			want{
				callCount:   1,
				callLimit:   20,
				callOrderBy: OrderBy,
			},
		},
		{
			"list-with-filter",
			args{query: api.GetComputeHostsParams{Filter: &hasSiteFilter}},
			want{
				callCount:  1,
				callLimit:  20,
				callFilter: hasSiteFilter,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := utils.NewMockInventoryServiceClient(
				utils.MockResponses{
					ListResourcesResponse: &inventory.ListResourcesResponse{
						Resources: []*inventory.GetResourceResponse{},
					},
					GetResourceResponse: &inventory.GetResourceResponse{},
				},
			)
			client := &clients.InventoryClientHandler{
				InvClient: mockClient,
			}
			h := handlers.NewHandlers(client, nil)

			ctx := context.TODO()
			job := types.NewJob(ctx, types.List, types.Host, tt.args.query, nil)
			_, err := h.Do(job)

			// NOTE for now this test does not cover the return cases
			// eg: the inventory client returns an error
			require.NoError(t, err)
			assert.Equal(
				t,
				tt.want.callCount,
				mockClient.ListResourcesCallCount,
				"ListResources not called",
			)
			switch mockClient.ListResourcesCalls[0].GetResource().GetResource().(type) {
			case *inventory.Resource_Host:
				// no-op
			default:
				assert.Fail(t, "ListResources called with wrong parameter")
			}
			assert.Equal(
				t,
				tt.want.callLimit,
				mockClient.ListResourcesCalls[0].GetLimit(),
				"ListResources wrong limit",
			)
			assert.Equal(
				t,
				tt.want.callOffset,
				mockClient.ListResourcesCalls[0].GetOffset(),
				"ListResources wrong offset",
			)
			assert.Equal(
				t,
				tt.want.callOrderBy,
				mockClient.ListResourcesCalls[0].GetOrderBy(),
				"ListResources wrong orderBy",
			)
			assert.Equal(
				t,
				tt.want.callFilter,
				mockClient.ListResourcesCalls[0].GetFilter(),
				"ListResources wrong filter",
			)
		})
	}
}

func Test_hostHandler_List(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	ctx := context.TODO()
	job := types.NewJob(ctx, types.List, types.Host, nil, api.GetComputeHostsParams{})
	r, err := h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	listParams := api.GetComputeHostsParams{
		Offset:   &pgOffset,
		PageSize: &pgSize,
	}
	job = types.NewJob(ctx, types.List, types.Host, listParams, inv_handlers.HostURLParams{})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	listResources, ok := r.Payload.Data.(api.HostsList)
	require.True(t, ok)
	assert.NotNil(t, listResources)
	assert.Equal(t, 0, len(*listResources.Hosts))

	host1 := inv_testing.CreateHost(t, nil, nil)

	listParams = api.GetComputeHostsParams{
		Offset:   &pgOffset,
		PageSize: &pgSize,
	}
	job = types.NewJob(ctx, types.List, types.Host, listParams, inv_handlers.HostURLParams{})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)
	listResources, ok = r.Payload.Data.(api.HostsList)
	require.True(t, ok)
	assert.NotNil(t, listResources)
	assert.Equal(t, 1, len(*listResources.Hosts))

	filter := fmt.Sprintf("%s = %q", computev1.HostResourceFieldResourceId, host1.GetResourceId())
	orderBy := computev1.HostResourceFieldResourceId
	listParams = api.GetComputeHostsParams{
		Offset:   &pgOffset,
		PageSize: &pgSize,
		Filter:   &filter,
		OrderBy:  &orderBy,
	}
	job = types.NewJob(ctx, types.List, types.Host, listParams, inv_handlers.HostURLParams{})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)
	listResources, ok = r.Payload.Data.(api.HostsList)
	require.True(t, ok)
	assert.NotNil(t, listResources)
	assert.Equal(t, 1, len(*listResources.Hosts))

	listParams = api.GetComputeHostsParams{
		Offset:   &pgIndexWrong,
		PageSize: &pgSizeWrong,
	}
	job = types.NewJob(ctx, types.List, types.Host, listParams, inv_handlers.HostURLParams{})
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))

	job = types.NewJob(ctx, types.List, types.Host, api.GetSitesParams{}, inv_handlers.HostURLParams{})
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))
}

func Test_hostHandler_Job_Error(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	ctx := context.TODO()
	job := types.NewJob(ctx, BadOperation, types.Host, nil, inv_handlers.HostURLParams{})
	_, err := h.Do(job)
	require.Error(t, err, "Expected error")
	require.Equal(t, http.StatusNotImplemented, errors.ErrorToHTTPStatus(err))

	// Good operation but wrong params
	job = types.NewJob(ctx, types.List, types.Host, inv_handlers.HostURLParams{}, nil)
	_, err = h.Do(job)
	require.Error(t, err, "Expected error")
	require.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))
}

func Test_hostHandler_Post(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	hostUUID := uuid.New()
	body := &api.Host{
		Metadata: &api.Metadata{
			{
				Key:   "examplekey",
				Value: "examplevalue",
			}, {
				Key:   "examplekey2",
				Value: "examplevalue2",
			},
		},
		Name: "hostName",
		Uuid: &hostUUID,
	}

	ctx := context.TODO()
	job := types.NewJob(ctx, types.Post, types.Host, body, inv_handlers.HostURLParams{})
	r, err := h.Do(job)
	require.Equal(t, nil, err)
	assert.Equal(t, http.StatusCreated, r.Status)
	gotRes, ok := r.Payload.Data.(*api.Host)
	assert.True(t, ok)

	// Validate Post changes
	job = types.NewJob(ctx, types.Get, types.Host, nil, inv_handlers.HostURLParams{
		HostID: *gotRes.ResourceId,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok = r.Payload.Data.(*api.Host)
	require.True(t, ok)
	assert.NotNil(t, gotRes)
	assert.Equal(t, body.Name, gotRes.Name)
	assert.Equal(t, api.HOSTSTATEONBOARDED, *gotRes.DesiredState)

	// test Post Error - no body
	job = types.NewJob(ctx, types.Post, types.Host, nil, inv_handlers.HostURLParams{})
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))

	// test Post error - wrong body request format
	job = types.NewJob(ctx, types.Post, types.Host, &utils.Site1Request, inv_handlers.HostURLParams{})
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))
}

func Test_hostHandler_Put(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	ctx := context.TODO()
	hostResource := inv_testing.CreateHost(t, nil, nil)

	body := &api.Host{
		Name: "hostName",
	}

	// test Put
	job := types.NewJob(
		ctx,
		types.Put,
		types.Host,
		body,
		inv_handlers.HostURLParams{HostID: hostResource.ResourceId},
	)
	r, err := h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// Validate Put changes
	job = types.NewJob(ctx, types.Get, types.Host, nil, inv_handlers.HostURLParams{
		HostID: hostResource.ResourceId,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok := r.Payload.Data.(*api.Host)
	require.True(t, ok)
	assert.NotNil(t, gotRes)
	assert.Equal(t, body.Name, gotRes.Name)
	assert.Equal(t, api.HOSTSTATEONBOARDED, *gotRes.DesiredState)

	// test Put Error - wrong body format
	job = types.NewJob(
		ctx,
		types.Put,
		types.Host,
		api.Site{},
		inv_handlers.HostURLParams{},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))

	// test Put Error - wrong params
	job = types.NewJob(
		ctx,
		types.Put,
		types.Host,
		body,
		inv_handlers.OUURLParams{},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))

	// Immutable field
	hostUUID := uuid.New()
	body = &api.Host{
		Uuid: &hostUUID,
		Name: "hostName",
	}

	// test Put Error - immutable field update not allowed
	job = types.NewJob(
		ctx,
		types.Put,
		types.Host,
		body,
		inv_handlers.HostURLParams{HostID: hostResource.ResourceId},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, errors.ErrorToHTTPStatus(err))
}

func Test_hostHandler_Invalidate(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	ctx := context.TODO()

	hostResource := inv_testing.CreateHost(t, nil, nil)

	// test Put /invalidate
	job := types.NewJob(
		ctx,
		types.Put,
		types.Host,
		&api.PutComputeHostsHostIDInvalidateJSONRequestBody{
			Note: defaultOperationNote,
		},
		inv_handlers.HostURLParams{
			HostID: hostResource.ResourceId,
			Action: types.HostActionInvalidate,
		},
	)
	r, err := h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// test Put /invalidate Error - wrong params
	job = types.NewJob(
		ctx,
		types.Put,
		types.Host,
		nil,
		inv_handlers.OUURLParams{},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))
}

//nolint:funlen // it is a test
func Test_hostHandler_Register(t *testing.T) {
	h := handlers.NewHandlers(&clients.InventoryClientHandler{InvClient: inv_testing.TestClients[inv_testing.APIClient]}, nil)
	require.NotNil(t, h)

	hostUUID := uuid.New()
	hostName := "hostName"
	hostSerial := "SN001"
	autoOnboard := true
	body := &api.HostRegisterInfo{
		Name:         &hostName,
		Uuid:         &hostUUID,
		SerialNumber: &hostSerial,
		AutoOnboard:  &autoOnboard,
	}

	ctx := context.TODO()
	job := types.NewJob(ctx, types.Post, types.Host, body, inv_handlers.HostURLParams{Action: types.HostActionRegister})
	r, err := h.Do(job)
	require.Equal(t, nil, err)
	assert.Equal(t, http.StatusCreated, r.Status)
	gotRes, ok := r.Payload.Data.(*api.Host)
	assert.True(t, ok)

	// Validate Post changes - autoOnboard = true
	job = types.NewJob(ctx, types.Get, types.Host, nil, inv_handlers.HostURLParams{
		HostID: *gotRes.ResourceId,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok = r.Payload.Data.(*api.Host)
	require.True(t, ok)
	assert.NotNil(t, gotRes)
	assert.Equal(t, body.SerialNumber, gotRes.SerialNumber)
	assert.Equal(t, body.Uuid.String(), gotRes.Uuid.String())
	assert.Equal(t, api.HOSTSTATEONBOARDED, *gotRes.DesiredState)

	// test Patch - autoOnboard = false
	autoOnboard = false
	job = types.NewJob(ctx, types.Patch, types.Host, &api.HostRegisterInfo{AutoOnboard: &autoOnboard}, inv_handlers.HostURLParams{
		Action: types.HostActionRegister,
		HostID: *gotRes.ResourceId,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// Validate Patch changes - autoOnboard = false
	job = types.NewJob(ctx, types.Get, types.Host, nil, inv_handlers.HostURLParams{
		HostID: *gotRes.ResourceId,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok = r.Payload.Data.(*api.Host)
	require.True(t, ok)
	assert.NotNil(t, gotRes)
	assert.Equal(t, api.HOSTSTATEREGISTERED, *gotRes.DesiredState)

	// test Patch - autoOnboard = true
	autoOnboard = true
	job = types.NewJob(ctx, types.Patch, types.Host, &api.HostRegisterInfo{AutoOnboard: &autoOnboard}, inv_handlers.HostURLParams{
		Action: types.HostActionRegister,
		HostID: *gotRes.ResourceId,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// Validate Patch changes - autoOnboard = true
	job = types.NewJob(ctx, types.Get, types.Host, nil, inv_handlers.HostURLParams{
		HostID: *gotRes.ResourceId,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok = r.Payload.Data.(*api.Host)
	require.True(t, ok)
	assert.NotNil(t, gotRes)
	assert.Equal(t, api.HOSTSTATEONBOARDED, *gotRes.DesiredState)

	// test Post - no UUID
	hostName = "hostName2"
	hostSerial = "SN002"
	body = &api.HostRegisterInfo{
		Name:         &hostName,
		SerialNumber: &hostSerial,
	}
	job = types.NewJob(ctx, types.Post, types.Host, body, inv_handlers.HostURLParams{Action: types.HostActionRegister})
	r, err = h.Do(job)
	require.Equal(t, nil, err)
	assert.Equal(t, http.StatusCreated, r.Status)
	gotRes, ok = r.Payload.Data.(*api.Host)
	assert.True(t, ok)

	// Validate Post changes - no UUID
	job = types.NewJob(ctx, types.Get, types.Host, nil, inv_handlers.HostURLParams{
		HostID: *gotRes.ResourceId,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok = r.Payload.Data.(*api.Host)
	require.True(t, ok)
	assert.NotNil(t, gotRes)
	assert.Equal(t, body.SerialNumber, gotRes.SerialNumber)

	// test Post - no SerialNumber
	hostName = "hostName2"
	hostUUID = uuid.New()
	body = &api.HostRegisterInfo{
		Name: &hostName,
		Uuid: &hostUUID,
	}
	job = types.NewJob(ctx, types.Post, types.Host, body, inv_handlers.HostURLParams{Action: types.HostActionRegister})
	r, err = h.Do(job)
	require.Equal(t, nil, err)
	assert.Equal(t, http.StatusCreated, r.Status)
	gotRes, ok = r.Payload.Data.(*api.Host)
	assert.True(t, ok)

	// test Post error - no SerialNumber
	job = types.NewJob(ctx, types.Get, types.Host, nil, inv_handlers.HostURLParams{
		HostID: *gotRes.ResourceId,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok = r.Payload.Data.(*api.Host)
	require.True(t, ok)
	assert.NotNil(t, gotRes)
	assert.Equal(t, body.Uuid, gotRes.Uuid)
}

func Test_hostHandler_Register_Error(t *testing.T) {
	h := handlers.NewHandlers(&clients.InventoryClientHandler{InvClient: inv_testing.TestClients[inv_testing.APIClient]}, nil)
	require.NotNil(t, h)
	ctx := context.TODO()

	// test Post Error - no body
	job := types.NewJob(ctx, types.Post, types.Host, nil, inv_handlers.HostURLParams{
		Action: types.HostActionRegister,
	})
	_, err := h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))

	// test Post error - wrong body request format
	job = types.NewJob(ctx, types.Post, types.Host, &utils.Site1Request, inv_handlers.HostURLParams{
		Action: types.HostActionRegister,
	})
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))

	// test Post error - no UUID, no SerialNumber
	hostName := "hostName1"
	autoOnboard := true
	body := &api.HostRegisterInfo{
		Name:         &hostName,
		Uuid:         nil,
		SerialNumber: nil,
		AutoOnboard:  &autoOnboard,
	}
	job = types.NewJob(ctx, types.Post, types.Host, body, inv_handlers.HostURLParams{
		Action: types.HostActionRegister,
	})
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))
}

func Test_hostHandler_Onboard(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	hostUUID := uuid.New()
	hostName := "hostName"
	hostSerial := "12345678"
	body := &api.HostRegisterInfo{
		Name:         &hostName,
		Uuid:         &hostUUID,
		SerialNumber: &hostSerial,
	}

	params := inv_handlers.HostURLParams{
		Action: types.HostActionRegister,
	}

	ctx := context.TODO()
	job := types.NewJob(ctx, types.Post, types.Host, body, params)
	r, err := h.Do(job)
	require.Equal(t, nil, err)
	assert.Equal(t, http.StatusCreated, r.Status)
	gotRes, ok := r.Payload.Data.(*api.Host)
	assert.True(t, ok)

	// Validate host registered
	job = types.NewJob(ctx, types.Get, types.Host, nil, inv_handlers.HostURLParams{
		HostID: *gotRes.ResourceId,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok = r.Payload.Data.(*api.Host)
	require.True(t, ok)
	assert.NotNil(t, gotRes)
	assert.Equal(t, body.SerialNumber, gotRes.SerialNumber)
	assert.Equal(t, body.Uuid.String(), gotRes.Uuid.String())
	assert.Equal(t, api.HOSTSTATEREGISTERED, *gotRes.DesiredState)

	// test /onboard
	params = inv_handlers.HostURLParams{
		Action: types.HostActionOnboard,
		HostID: *gotRes.ResourceId,
	}

	job = types.NewJob(ctx, types.Patch, types.Host, body, params)
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// Validate host onboarded
	job = types.NewJob(ctx, types.Get, types.Host, nil, inv_handlers.HostURLParams{
		HostID: *gotRes.ResourceId,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok = r.Payload.Data.(*api.Host)
	require.True(t, ok)
	assert.NotNil(t, gotRes)
	assert.Equal(t, body.SerialNumber, gotRes.SerialNumber)
	assert.Equal(t, body.Uuid.String(), gotRes.Uuid.String())
	assert.Equal(t, api.HOSTSTATEONBOARDED, *gotRes.DesiredState)
}

func Test_hostHandler_Patch(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	ctx := context.TODO()
	hostResource := inv_testing.CreateHost(t, nil, nil)
	body := &api.Host{
		Name: "hostName",
	}

	// test Patch
	job := types.NewJob(
		ctx,
		types.Patch,
		types.Host,
		body,
		inv_handlers.HostURLParams{HostID: hostResource.ResourceId},
	)
	r, err := h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// Validate Patch changes
	job = types.NewJob(ctx, types.Get, types.Host, nil, inv_handlers.HostURLParams{
		HostID: hostResource.ResourceId,
	})
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	gotRes, ok := r.Payload.Data.(*api.Host)
	require.True(t, ok)
	assert.NotNil(t, gotRes)
	assert.Equal(t, body.Name, gotRes.Name)
	assert.Equal(t, api.HOSTSTATEONBOARDED, *gotRes.DesiredState)

	// test Patch Error - wrong data
	job = types.NewJob(
		ctx,
		types.Patch,
		types.Host,
		api.GetComputeHostsParams{},
		inv_handlers.HostURLParams{},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))

	// test Patch Error - wrong params
	job = types.NewJob(
		ctx,
		types.Patch,
		types.Host,
		body,
		inv_handlers.OUURLParams{},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))

	// Immutable field
	hostUUID := uuid.New()
	body = &api.Host{
		Uuid: &hostUUID,
		Name: "hostName",
	}

	// test Patch error - immutable field update not allowed
	job = types.NewJob(
		ctx,
		types.Patch,
		types.Host,
		body,
		inv_handlers.HostURLParams{HostID: hostResource.ResourceId},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, errors.ErrorToHTTPStatus(err))
}

func Test_hostHandler_PatchFieldMask(t *testing.T) {
	mockClient := utils.NewMockInventoryServiceClient(
		utils.MockResponses{
			UpdateResourceResponse: &inventory.Resource{
				Resource: &inventory.Resource_Host{
					Host: &computev1.HostResource{},
				},
			},
		},
	)
	client := &clients.InventoryClientHandler{
		InvClient: mockClient,
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	body := &utils.Host1Request
	body.Instance = &api.Instance{}
	// Immutable field
	body.Uuid = nil

	ctx := context.TODO()
	// test Patch
	job := types.NewJob(
		ctx,
		types.Patch,
		types.Host,
		body,
		inv_handlers.HostURLParams{},
	)
	r, err := h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// test Patch FieldMask
	expectedPatchFieldMask := BuildFmFromHostRequest(body)
	host := &computev1.HostResource{}
	expectedFieldMask, err := fieldmaskpb.New(host, expectedPatchFieldMask...)
	require.NoError(t, err)

	err = validateFM(mockClient.LastUpdateResourceRequestFieldMask, expectedFieldMask)
	assert.NoError(t, err)
}

func Test_hostHandler_PatchRegisterFieldMask(t *testing.T) {
	mockClient := utils.NewMockInventoryServiceClient(
		utils.MockResponses{
			UpdateResourceResponse: &inventory.Resource{
				Resource: &inventory.Resource_Host{
					Host: &computev1.HostResource{},
				},
			},
		},
	)
	client := &clients.InventoryClientHandler{
		InvClient: mockClient,
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	body := &utils.HostRegisterAutoOnboard
	// Immutable fields
	body.Uuid = nil
	body.SerialNumber = nil

	ctx := context.TODO()
	// test Patch - fields: name & autoOnboard
	job := types.NewJob(
		ctx,
		types.Patch,
		types.Host,
		body,
		inv_handlers.HostURLParams{Action: types.HostActionRegister},
	)
	r, err := h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// validate Patch FieldMask
	expectedPatchFieldMask := BuildFmFromHostRegisterRequest(body)
	host := &computev1.HostResource{}
	expectedFieldMask, err := fieldmaskpb.New(host, expectedPatchFieldMask...)
	require.NoError(t, err)

	err = validateFM(mockClient.LastUpdateResourceRequestFieldMask, expectedFieldMask)
	assert.NoError(t, err)

	// test Patch - fields: name only
	body.AutoOnboard = nil
	job = types.NewJob(
		ctx,
		types.Patch,
		types.Host,
		body,
		inv_handlers.HostURLParams{Action: types.HostActionRegister},
	)
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// validate Patch FieldMask
	expectedPatchFieldMask = BuildFmFromHostRegisterRequest(body)
	expectedFieldMask, err = fieldmaskpb.New(host, expectedPatchFieldMask...)
	require.NoError(t, err)

	err = validateFM(mockClient.LastUpdateResourceRequestFieldMask, expectedFieldMask)
	assert.NoError(t, err)

	// test Patch - fields: autoOnboard only
	body.Name = nil
	body.AutoOnboard = &utils.AutoOnboardTrue
	job = types.NewJob(
		ctx,
		types.Patch,
		types.Host,
		body,
		inv_handlers.HostURLParams{Action: types.HostActionRegister},
	)
	r, err = h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.Status)

	// validate Patch FieldMask
	expectedPatchFieldMask = BuildFmFromHostRegisterRequest(body)
	expectedFieldMask, err = fieldmaskpb.New(host, expectedPatchFieldMask...)
	require.NoError(t, err)

	err = validateFM(mockClient.LastUpdateResourceRequestFieldMask, expectedFieldMask)
	assert.NoError(t, err)
}

func verifyHostResourceInterfaces(t *testing.T, r *types.Response) {
	t.Helper()

	require.NotNil(t, r.Payload)
	host, ok := r.Payload.Data.(*api.Host)
	require.True(t, ok)
	require.NotNil(t, host.OnboardingStatusIndicator)
	require.NotNil(t, host.RegistrationStatusIndicator)
	require.NotNil(t, host.HostStatusIndicator)
}

func verifyHostResourceGPUs(t *testing.T, r *types.Response, expected *computev1.HostgpuResource) {
	t.Helper()

	require.NotNil(t, r.Payload)
	host, ok := r.Payload.Data.(*api.Host)
	require.True(t, ok)
	require.Len(t, *host.HostGpus, 1)
	apiHostGPU := (*host.HostGpus)[0]
	assert.Equal(t, *apiHostGPU.Description, expected.Description)
	assert.Equal(t, *apiHostGPU.PciId, expected.GetPciId())
	assert.Equal(t, *apiHostGPU.Vendor, expected.GetVendor())
	assert.Equal(t, *apiHostGPU.Product, expected.GetProduct())
	assert.Equal(t, *apiHostGPU.DeviceName, expected.GetDeviceName())
	assert.Equal(t, *apiHostGPU.Capabilities, strings.Split(expected.GetFeatures(), ","))
}

func verifyHostStatusFields(t *testing.T, host *api.Host,
	expected *computev1.HostResource,
) {
	t.Helper()

	require.NotNil(t, host.HostStatusIndicator)
	require.NotNil(t, host.OnboardingStatusIndicator)
	require.NotNil(t, host.RegistrationStatusIndicator)

	assert.Equal(t, expected.GetOnboardingStatus(), *host.OnboardingStatus)
	assert.Equal(t, expected.GetOnboardingStatusTimestamp(), *host.OnboardingStatusTimestamp)
	assert.Equal(t, *inv_handlers.GrpcToOpenAPIStatusIndicator(expected.GetOnboardingStatusIndicator()),
		*host.OnboardingStatusIndicator)

	assert.Equal(t, expected.GetRegistrationStatus(), *host.RegistrationStatus)
	assert.Equal(t, expected.GetRegistrationStatusTimestamp(), *host.RegistrationStatusTimestamp)
	assert.Equal(t, *inv_handlers.GrpcToOpenAPIStatusIndicator(expected.GetRegistrationStatusIndicator()),
		*host.RegistrationStatusIndicator)

	assert.Equal(t, expected.GetHostStatus(), *host.HostStatus)
	assert.Equal(t, expected.GetHostStatusTimestamp(), *host.HostStatusTimestamp)
	assert.Equal(t, *inv_handlers.GrpcToOpenAPIStatusIndicator(expected.GetHostStatusIndicator()),
		*host.HostStatusIndicator)
}

//nolint:funlen // it's a test
func Test_hostHandler_Get(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	ctx := context.TODO()

	provResource := inv_testing.CreateProvider(t, "TEST")
	hostResource := inv_testing.CreateHost(t, nil, provResource)
	// Adds Nic to validate Get when Nics exists, i.e., method getInterfaceToIPAddresses
	hostNicResource := inv_testing.CreateHostNic(t, hostResource)
	hostNicIPAddressResource := inv_testing.CreateIPAddress(t, hostNicResource, true)
	hostGpu := inv_testing.CreatHostGPU(t, hostResource)
	osResource := inv_testing.CreateOs(t)
	instResource := inv_testing.CreateInstance(t, hostResource, osResource)
	instResource.Os = osResource
	hostResource.Instance = instResource

	// Note that we must update HostResource to verify status fields, we may skip updating once
	// inv_testing.CreateHost will set status fields by default.
	//nolint:gosec // uint64 conversions are safe for testing
	updatedHost := &computev1.HostResource{
		CpuTopology:                 "{\"some_json\":[]}",
		OnboardingStatus:            "some onboarding status",
		OnboardingStatusIndicator:   statusv1.StatusIndication_STATUS_INDICATION_IDLE,
		OnboardingStatusTimestamp:   uint64(time.Now().Unix()),
		RegistrationStatus:          "some registration status",
		RegistrationStatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IDLE,
		RegistrationStatusTimestamp: uint64(time.Now().Unix()),
		HostStatus:                  "some host status",
		HostStatusIndicator:         statusv1.StatusIndication_STATUS_INDICATION_IN_PROGRESS,
		HostStatusTimestamp:         uint64(time.Now().Unix()),
	}

	_, err := inv_testing.TestClients[inv_testing.APIClient].Update(ctx, hostResource.GetResourceId(),
		&fieldmaskpb.FieldMask{Paths: []string{
			computev1.HostResourceFieldCpuTopology,
			computev1.HostResourceFieldOnboardingStatus,
			computev1.HostResourceFieldOnboardingStatusIndicator,
			computev1.HostResourceFieldOnboardingStatusTimestamp,
			computev1.HostResourceFieldRegistrationStatus,
			computev1.HostResourceFieldRegistrationStatusIndicator,
			computev1.HostResourceFieldRegistrationStatusTimestamp,
			computev1.HostResourceFieldHostStatus,
			computev1.HostResourceFieldHostStatusIndicator,
			computev1.HostResourceFieldHostStatusTimestamp,
		}}, &inventory.Resource{
			Resource: &inventory.Resource_Host{
				Host: updatedHost,
			},
		})
	require.NoError(t, err)

	// test Get
	job := types.NewJob(
		ctx,
		types.Get,
		types.Host,
		nil,
		inv_handlers.HostURLParams{HostID: hostResource.ResourceId},
	)
	r, err := h.Do(job)
	require.Equal(t, err, nil)
	require.Equal(t, http.StatusOK, r.Status)
	verifyHostResourceInterfaces(t, r)
	verifyHostResourceGPUs(t, r, hostGpu)
	hostResp, ok := r.Payload.Data.(*api.Host)
	require.True(t, ok)
	assert.Equal(t, updatedHost.GetCpuTopology(), *hostResp.CpuTopology)
	assert.Equal(t, hostResource.GetName(), hostResp.Name)
	assert.Equal(t, hostResource.GetResourceId(), *hostResp.ResourceId)
	require.NotNil(t, hostResp.Instance)
	assert.Equal(t, hostResource.GetInstance().GetResourceId(), *hostResp.Instance.ResourceId)
	require.NotNil(t, hostResp.Instance.Os)
	assert.Equal(t, hostResource.GetInstance().GetOs().GetResourceId(), *hostResp.Instance.Os.ResourceId)
	assert.Equal(t, hostResource.GetInstance().GetOs().GetName(), *hostResp.Instance.Os.Name)
	assert.Equal(t, hostResource.GetInstance().GetOs().GetResourceId(), *hostResp.Instance.CurrentOs.ResourceId)

	verifyHostStatusFields(t, hostResp, updatedHost)
	hostHasNicIP := false
	hostNics := *hostResp.HostNics
	for _, hostNic := range hostNics {
		if *hostNic.DeviceName == hostNicResource.GetDeviceName() {
			for _, ip := range *hostNic.Ipaddresses {
				if ip.Address.String() == hostNicIPAddressResource.GetAddress() {
					hostHasNicIP = true
				}
			}
		}
	}
	assert.True(t, hostHasNicIP)

	VerifyProvider(t, hostResp.Provider, provResource)

	// Test Get error - wrong params
	job = types.NewJob(
		ctx,
		types.Get,
		types.Host,
		api.GetComputeHostsParams{},
		inv_handlers.OUURLParams{},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))
}

func Test_hostHandler_Delete(t *testing.T) {
	client := &clients.InventoryClientHandler{
		InvClient: inv_testing.TestClients[inv_testing.APIClient],
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	hostResource := inv_testing.CreateHostNoCleanup(t, nil, nil)

	ctx := context.TODO()
	job := types.NewJob(
		ctx,
		types.Delete,
		types.Host,
		&api.DeleteComputeHostsHostIDJSONRequestBody{
			Note: defaultOperationNote,
		},
		inv_handlers.HostURLParams{HostID: hostResource.ResourceId},
	)
	r, err := h.Do(job)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, r.Status)

	// Test Delete error - wrong params
	job = types.NewJob(
		ctx,
		types.Delete,
		types.Host,
		api.GetComputeHostsParams{},
		inv_handlers.OUURLParams{},
	)
	_, err = h.Do(job)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, errors.ErrorToHTTPStatus(err))
}

func Test_Inventory_Host_Integration(t *testing.T) {
	// verify the projection of the constants to Proto first;
	// we build a map using the field names of the proto stored in the
	// ProtoHost* slices in internal/work/handlers/host.go. Elements must
	// have a mapping key otherwise we throw an error if there is no
	// alignment with host proto in Inventory. Make sure to update these
	// two slices in internal/work/handlers/host.go
	hostResource := &computev1.HostResource{}
	validateInventoryIntegration(
		t, hostResource, api.Host{}, inv_handlers.OpenAPIHostToProto,
		inv_handlers.OpenAPIHostToProtoExcluded, maps.Values(inv_handlers.OpenAPIHostToProto), false)
}

func validateInventoryIntegration(
	t *testing.T,
	msg proto.Message,
	openapi interface{},
	openAPIToProto map[string]string,
	openAPIExcluded map[string]struct{},
	protoFields []string,
	checkEmpty bool,
) {
	t.Helper()
	protoMessage := msg.ProtoReflect()
	invToAPI := make(map[string]string)
	for i := 0; i < protoMessage.Descriptor().Fields().Len(); i++ {
		field := protoMessage.Descriptor().Fields().Get(i).Name()
		invToAPI[string(protoMessage.Descriptor().Fields().Get(i).Name())] = string(field)
	}

	for _, field := range protoFields {
		_, ok := invToAPI[field]
		assert.True(t, ok, fmt.Sprintf("Cannot find %s in HostResource proto", field))
	}

	verifyStruct(t, openapi, openAPIToProto, openAPIExcluded)
	if checkEmpty {
		// verifyStruct will consume all the fields of
		// openAPIToProto. See verifyStruct documentation
		assert.Empty(t, openAPIToProto, "Translation OpenAPI to Proto broken update fm constants")
	}
}

// It is an helper function that verifies if the field is a struct and
// if the fields of this struct have a mapping into the Proto fields which are
// stored as map into openApiToProto. Exclusion rules exist and goes under the map openApiExcluded.
// If there is a change in the translation logic OpenAPI to Proto you need to
// update these mappings in internal/work/handlers.
func verifyStruct(
	t *testing.T,
	field interface{},
	openAPIToProto map[string]string,
	openAPIExcluded map[string]struct{},
) {
	t.Helper()
	rt := reflect.TypeOf(field)
	assert.Equal(t, rt.Kind(), reflect.Struct)

	for i := 0; i < rt.NumField(); i++ {
		// use split to ignore tag "options"
		v := strings.Split(rt.Field(i).Tag.Get("json"), ",")[0]
		_, ok := openAPIExcluded[v]
		// not excluded
		if !ok {
			_, ok := openAPIToProto[v]
			assert.True(t, ok, fmt.Sprintf("Cannot find %s in proto", v))
			delete(openAPIToProto, v)
		}
	}
}

// Test_hostHandler_InvMockClient_Errors evaluates all
// host handler methods with mock inventory client
// that returns errors.
func Test_hostHandler_InvMockClient_Errors(t *testing.T) {
	mockClient := utils.NewMockInventoryServiceClientError()
	client := &clients.InventoryClientHandler{
		InvClient: mockClient,
	}
	h := handlers.NewHandlers(client, nil)
	require.NotNil(t, h)

	ctx := context.TODO()
	job := types.NewJob(
		ctx,
		types.List,
		types.Host,
		api.GetComputeHostsParams{
			Offset:   &pgOffset,
			PageSize: &pgSize,
		},
		inv_handlers.HostURLParams{},
	)
	_, err := h.Do(job)
	assert.Error(t, err)

	body := &utils.Host3Request
	job = types.NewJob(ctx, types.Post, types.Host, body, inv_handlers.HostURLParams{})
	_, err = h.Do(job)
	assert.Error(t, err)

	job = types.NewJob(
		ctx,
		types.Put,
		types.Host,
		body,
		inv_handlers.HostURLParams{},
	)
	_, err = h.Do(job)
	assert.Error(t, err)

	job = types.NewJob(
		ctx,
		types.Get,
		types.Host,
		nil,
		inv_handlers.HostURLParams{HostID: "host-12345678"},
	)
	_, err = h.Do(job)
	assert.Error(t, err)

	job = types.NewJob(
		ctx,
		types.Delete,
		types.Host,
		api.GetComputeHostsParams{},
		inv_handlers.HostURLParams{},
	)
	_, err = h.Do(job)
	assert.Error(t, err)
}

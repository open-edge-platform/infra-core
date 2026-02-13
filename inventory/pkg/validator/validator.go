// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"

	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventoryv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	localaccountv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/localaccount/v1"
	locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	networkv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/network/v1"
	osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	ouv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/ou/v1"
	providerv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	remoteaccessv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/remoteaccess/v1"
	schedulev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/schedule/v1"
	telemetryv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
	tenantv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/tenant/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
)

var (
	protovalidator protovalidate.Validator

	preWarmMessages = make([]proto.Message, 0)

	zlog = logging.GetLogger("InventoryValidator")
)

var invMessages = []proto.Message{
	// all possible top-level messages from api/inventory/v1/inventory.proto
	&inventoryv1.Resource{},
	&inventoryv1.SubscribeEventsRequest{},
	&inventoryv1.SubscribeEventsResponse{},
	&inventoryv1.ChangeSubscribeEventsRequest{},
	&inventoryv1.ChangeSubscribeEventsResponse{},
	&inventoryv1.CreateResourceRequest{},
	&inventoryv1.FindResourcesRequest{},
	&inventoryv1.FindResourcesResponse{},
	&inventoryv1.GetResourceRequest{},
	&inventoryv1.GetResourceResponse{},
	&inventoryv1.UpdateResourceRequest{},
	&inventoryv1.DeleteResourceRequest{},
	&inventoryv1.DeleteResourceResponse{},
	&inventoryv1.ListResourcesRequest{},
	&inventoryv1.ListResourcesResponse{},
	&inventoryv1.ListInheritedTelemetryProfilesRequest{},
	&inventoryv1.ListInheritedTelemetryProfilesResponse{},
	&inventoryv1.DeleteAllResourcesRequest{},

	// Inventory resources
	&computev1.HostResource{},
	&computev1.InstanceResource{},
	&computev1.HostnicResource{},
	&computev1.HoststorageResource{},
	&computev1.HostgpuResource{},
	&computev1.HostusbResource{},
	&computev1.HostdeviceResource{},
	&computev1.WorkloadResource{},
	&computev1.WorkloadMember{},
	&locationv1.SiteResource{},
	&locationv1.RegionResource{},
	&networkv1.EndpointResource{},
	&networkv1.NetlinkResource{},
	&networkv1.NetworkSegment{},
	&networkv1.IPAddressResource{},
	&osv1.OperatingSystemResource{},
	&ouv1.OuResource{},
	&providerv1.ProviderResource{},
	&schedulev1.SingleScheduleResource{},
	&schedulev1.RepeatedScheduleResource{},
	&telemetryv1.TelemetryGroupResource{},
	&telemetryv1.TelemetryProfile{},
	&remoteaccessv1.RemoteAccessConfiguration{},
	&localaccountv1.LocalAccountResource{},
	&tenantv1.Tenant{},
}

//nolint:gochecknoinits // we auto-initialize protovalidator with Inventory messages when the package is imported.
func init() {
	MustInit(invMessages)
}

func startProtovalidate(preWarmMsg ...proto.Message) (protovalidate.Validator, error) {
	validator, err := protovalidate.New(
		// this warms up validator - pre-uploads message's validation constraints
		protovalidate.WithMessages(
			preWarmMsg...,
		),
	)
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msgf("Error starting validator")
		return nil, errors.Wrap(err)
	}

	return validator, nil
}

// MustInit initializes protovalidate and pre-warms it with provided preWarmMsgs.
// Note that this function does fatal in the case of error.
func MustInit(preWarmMsgs []proto.Message) {
	preWarmMessages = append(preWarmMessages, preWarmMsgs...)
	_validator, err := startProtovalidate(preWarmMessages...)
	if err != nil {
		zlog.InfraSec().Fatal().Msgf("Failed to initialize proto validate: %s", err)
	}
	protovalidator = _validator
}

func ValidateMessage(message proto.Message) error {
	if err := protovalidator.Validate(message); err != nil {
		zlog.InfraSec().InfraErr(err).Msgf("Error validating input data: %v", message)
		return errors.Wrap(err)
	}

	return nil
}

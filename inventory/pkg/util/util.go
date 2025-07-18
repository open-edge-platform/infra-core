// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/mennanov/fmutils"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	compute_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	localaccountv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/localaccount/v1"
	location_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	network_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/network/v1"
	os_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	ou_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/ou/v1"
	provider_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	remoteaccessv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/remoteaccess/v1"
	schedule_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/schedule/v1"
	telemetry_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
	tenantv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/tenant/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
)

var zlog = logging.GetLogger("InfraInvInvUtil")

type ResourcePrefix string

const (
	INT63    = 63
	Kilobyte = 1024
	Megabyte = 1024 * Kilobyte
	Gigabyte = 1024 * Megabyte
	Terabyte = 1024 * Gigabyte

	MaxResourceNestingLevel = 5
)

const (
	ResourcePrefixUnspecified      ResourcePrefix = "unspecified"
	ResourcePrefixInstance         ResourcePrefix = "inst"
	ResourcePrefixHost             ResourcePrefix = "host"
	ResourcePrefixHoststorage      ResourcePrefix = "hoststorage"
	ResourcePrefixHostnic          ResourcePrefix = "hostnic"
	ResourcePrefixHostusb          ResourcePrefix = "hostusb"
	ResourcePrefixHostgpu          ResourcePrefix = "hostgpu"
	ResourcePrefixNetworkSegment   ResourcePrefix = "netseg"
	ResourcePrefixNetlink          ResourcePrefix = "netlink"
	ResourcePrefixEndpoint         ResourcePrefix = "endpoint"
	ResourcePrefixRegion           ResourcePrefix = "region"
	ResourcePrefixSite             ResourcePrefix = "site"
	ResourcePrefixOu               ResourcePrefix = "ou"
	ResourcePrefixProject          ResourcePrefix = "proj"
	ResourcePrefixUser             ResourcePrefix = "user"
	ResourcePrefixProvider         ResourcePrefix = "provider"
	ResourcePrefixOs               ResourcePrefix = "os"
	ResourcePrefixSingleSchedule   ResourcePrefix = "singlesche"
	ResourcePrefixRepeatedSchedule ResourcePrefix = "repeatedsche"
	ResourcePrefixTelemetryGroup   ResourcePrefix = "telemetrygroup"
	ResourcePrefixTelemetryProfile ResourcePrefix = "telemetryprofile"
	ResourcePrefixWorkload         ResourcePrefix = "workload"
	ResourcePrefixWorkloadMember   ResourcePrefix = "workloadmember"
	ResourcePrefixIPAddress        ResourcePrefix = "ipaddr"
	ResourcePrefixRemoteAccessConf ResourcePrefix = "rmtacconf"
	ResourcePrefixLocalAccount     ResourcePrefix = "localaccount"
	ResourcePrefixTenant           ResourcePrefix = "tenant"
	ResourcePrefixOsUpdatePolicy   ResourcePrefix = "osupdatepolicy"
	ResourcePrefixCustomConfig     ResourcePrefix = "customconfig"
	ResourcePrefixOsUpdateRun      ResourcePrefix = "osupdaterun"
)

func ResourceKindToPrefix(kind inv_v1.ResourceKind) ResourcePrefix {
	mapResourceKindToPrefix := map[inv_v1.ResourceKind]ResourcePrefix{
		inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE:          ResourcePrefixInstance,
		inv_v1.ResourceKind_RESOURCE_KIND_HOST:              ResourcePrefixHost,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE:       ResourcePrefixHoststorage,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC:           ResourcePrefixHostnic,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB:           ResourcePrefixHostusb,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU:           ResourcePrefixHostgpu,
		inv_v1.ResourceKind_RESOURCE_KIND_NETWORKSEGMENT:    ResourcePrefixNetworkSegment,
		inv_v1.ResourceKind_RESOURCE_KIND_NETLINK:           ResourcePrefixNetlink,
		inv_v1.ResourceKind_RESOURCE_KIND_ENDPOINT:          ResourcePrefixEndpoint,
		inv_v1.ResourceKind_RESOURCE_KIND_SITE:              ResourcePrefixSite,
		inv_v1.ResourceKind_RESOURCE_KIND_REGION:            ResourcePrefixRegion,
		inv_v1.ResourceKind_RESOURCE_KIND_OU:                ResourcePrefixOu,
		inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER:          ResourcePrefixProvider,
		inv_v1.ResourceKind_RESOURCE_KIND_OS:                ResourcePrefixOs,
		inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE:    ResourcePrefixSingleSchedule,
		inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE:  ResourcePrefixRepeatedSchedule,
		inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_GROUP:   ResourcePrefixTelemetryGroup,
		inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_PROFILE: ResourcePrefixTelemetryProfile,
		inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD:          ResourcePrefixWorkload,
		inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD_MEMBER:   ResourcePrefixWorkloadMember,
		inv_v1.ResourceKind_RESOURCE_KIND_IPADDRESS:         ResourcePrefixIPAddress,
		inv_v1.ResourceKind_RESOURCE_KIND_RMT_ACCESS_CONF:   ResourcePrefixRemoteAccessConf,
		inv_v1.ResourceKind_RESOURCE_KIND_TENANT:            ResourcePrefixTenant,
		inv_v1.ResourceKind_RESOURCE_KIND_LOCALACCOUNT:      ResourcePrefixLocalAccount,
		inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATEPOLICY:    ResourcePrefixOsUpdatePolicy,
		inv_v1.ResourceKind_RESOURCE_KIND_CUSTOMCONFIG:      ResourcePrefixCustomConfig,
		inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATERUN:       ResourcePrefixOsUpdateRun,
	}

	prefix, ok := mapResourceKindToPrefix[kind]
	if !ok {
		zlog.InfraSec().InfraError("Unable to map resource kind %d", kind).Msg("")
		return ResourcePrefixUnspecified
	}
	return prefix
}

// GetResourceKindFromResource returns the actual resource kind set in the given Resource.
//
//nolint:cyclop,funlen // high cyclomatic complexity due to the switch and length due to the switch
func GetResourceKindFromResource(resource *inv_v1.Resource) inv_v1.ResourceKind {
	switch resource.GetResource().(type) {
	case *inv_v1.Resource_Instance:
		return inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE
	case *inv_v1.Resource_Host:
		return inv_v1.ResourceKind_RESOURCE_KIND_HOST
	case *inv_v1.Resource_Hoststorage:
		return inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE
	case *inv_v1.Resource_Hostnic:
		return inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC
	case *inv_v1.Resource_Hostusb:
		return inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB
	case *inv_v1.Resource_Hostgpu:
		return inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU
	case *inv_v1.Resource_NetworkSegment:
		return inv_v1.ResourceKind_RESOURCE_KIND_NETWORKSEGMENT
	case *inv_v1.Resource_Netlink:
		return inv_v1.ResourceKind_RESOURCE_KIND_NETLINK
	case *inv_v1.Resource_Endpoint:
		return inv_v1.ResourceKind_RESOURCE_KIND_ENDPOINT
	case *inv_v1.Resource_Site:
		return inv_v1.ResourceKind_RESOURCE_KIND_SITE
	case *inv_v1.Resource_Region:
		return inv_v1.ResourceKind_RESOURCE_KIND_REGION
	case *inv_v1.Resource_Ou:
		return inv_v1.ResourceKind_RESOURCE_KIND_OU
	case *inv_v1.Resource_Provider:
		return inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER
	case *inv_v1.Resource_Os:
		return inv_v1.ResourceKind_RESOURCE_KIND_OS
	case *inv_v1.Resource_Singleschedule:
		return inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE
	case *inv_v1.Resource_Repeatedschedule:
		return inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE
	case *inv_v1.Resource_TelemetryGroup:
		return inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_GROUP
	case *inv_v1.Resource_TelemetryProfile:
		return inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_PROFILE
	case *inv_v1.Resource_Workload:
		return inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD
	case *inv_v1.Resource_WorkloadMember:
		return inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD_MEMBER
	case *inv_v1.Resource_Ipaddress:
		return inv_v1.ResourceKind_RESOURCE_KIND_IPADDRESS
	case *inv_v1.Resource_RemoteAccess:
		return inv_v1.ResourceKind_RESOURCE_KIND_RMT_ACCESS_CONF
	case *inv_v1.Resource_Tenant:
		return inv_v1.ResourceKind_RESOURCE_KIND_TENANT
	case *inv_v1.Resource_LocalAccount:
		return inv_v1.ResourceKind_RESOURCE_KIND_LOCALACCOUNT
	case *inv_v1.Resource_OsUpdatePolicy:
		return inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATEPOLICY
	case *inv_v1.Resource_CustomConfig:
		return inv_v1.ResourceKind_RESOURCE_KIND_CUSTOMCONFIG
	case *inv_v1.Resource_OsUpdateRun:
		return inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATERUN
	}
	zlog.InfraSec().InfraError("Unable to map resource to its prefix: %s", resource).Msg("")
	return inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED
}

func PrefixToResourceKind(prefix ResourcePrefix) inv_v1.ResourceKind {
	mapPrefixToResourceKind := map[ResourcePrefix]inv_v1.ResourceKind{
		ResourcePrefixInstance:         inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE,
		ResourcePrefixHost:             inv_v1.ResourceKind_RESOURCE_KIND_HOST,
		ResourcePrefixHoststorage:      inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE,
		ResourcePrefixHostnic:          inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC,
		ResourcePrefixHostusb:          inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB,
		ResourcePrefixHostgpu:          inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU,
		ResourcePrefixNetworkSegment:   inv_v1.ResourceKind_RESOURCE_KIND_NETWORKSEGMENT,
		ResourcePrefixNetlink:          inv_v1.ResourceKind_RESOURCE_KIND_NETLINK,
		ResourcePrefixEndpoint:         inv_v1.ResourceKind_RESOURCE_KIND_ENDPOINT,
		ResourcePrefixSite:             inv_v1.ResourceKind_RESOURCE_KIND_SITE,
		ResourcePrefixRegion:           inv_v1.ResourceKind_RESOURCE_KIND_REGION,
		ResourcePrefixOu:               inv_v1.ResourceKind_RESOURCE_KIND_OU,
		ResourcePrefixProvider:         inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER,
		ResourcePrefixOs:               inv_v1.ResourceKind_RESOURCE_KIND_OS,
		ResourcePrefixSingleSchedule:   inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE,
		ResourcePrefixRepeatedSchedule: inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE,
		ResourcePrefixTelemetryGroup:   inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_GROUP,
		ResourcePrefixTelemetryProfile: inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_PROFILE,
		ResourcePrefixWorkload:         inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD,
		ResourcePrefixWorkloadMember:   inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD_MEMBER,
		ResourcePrefixIPAddress:        inv_v1.ResourceKind_RESOURCE_KIND_IPADDRESS,
		ResourcePrefixRemoteAccessConf: inv_v1.ResourceKind_RESOURCE_KIND_RMT_ACCESS_CONF,
		ResourcePrefixLocalAccount:     inv_v1.ResourceKind_RESOURCE_KIND_LOCALACCOUNT,
		ResourcePrefixOsUpdatePolicy:   inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATEPOLICY,
		ResourcePrefixTenant:           inv_v1.ResourceKind_RESOURCE_KIND_TENANT,
		ResourcePrefixCustomConfig:     inv_v1.ResourceKind_RESOURCE_KIND_CUSTOMCONFIG,
		ResourcePrefixOsUpdateRun:      inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATERUN,
	}

	resourceKind, ok := mapPrefixToResourceKind[prefix]
	if !ok {
		zlog.InfraSec().InfraError("Unable to map resource prefix %s", prefix).Msg("")
		return inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED
	}
	return resourceKind
}

func stringToPrefix(s string) (ResourcePrefix, error) {
	mapStringToPrefix := map[string]ResourcePrefix{
		string(ResourcePrefixUnspecified):      ResourcePrefixUnspecified,
		string(ResourcePrefixInstance):         ResourcePrefixInstance,
		string(ResourcePrefixHost):             ResourcePrefixHost,
		string(ResourcePrefixHoststorage):      ResourcePrefixHoststorage,
		string(ResourcePrefixHostnic):          ResourcePrefixHostnic,
		string(ResourcePrefixHostusb):          ResourcePrefixHostusb,
		string(ResourcePrefixHostgpu):          ResourcePrefixHostgpu,
		string(ResourcePrefixNetworkSegment):   ResourcePrefixNetworkSegment,
		string(ResourcePrefixNetlink):          ResourcePrefixNetlink,
		string(ResourcePrefixEndpoint):         ResourcePrefixEndpoint,
		string(ResourcePrefixRegion):           ResourcePrefixRegion,
		string(ResourcePrefixSite):             ResourcePrefixSite,
		string(ResourcePrefixOu):               ResourcePrefixOu,
		string(ResourcePrefixProject):          ResourcePrefixProject,
		string(ResourcePrefixUser):             ResourcePrefixUser,
		string(ResourcePrefixProvider):         ResourcePrefixProvider,
		string(ResourcePrefixOs):               ResourcePrefixOs,
		string(ResourcePrefixSingleSchedule):   ResourcePrefixSingleSchedule,
		string(ResourcePrefixRepeatedSchedule): ResourcePrefixRepeatedSchedule,
		string(ResourcePrefixTelemetryGroup):   ResourcePrefixTelemetryGroup,
		string(ResourcePrefixTelemetryProfile): ResourcePrefixTelemetryProfile,
		string(ResourcePrefixWorkload):         ResourcePrefixWorkload,
		string(ResourcePrefixWorkloadMember):   ResourcePrefixWorkloadMember,
		string(ResourcePrefixIPAddress):        ResourcePrefixIPAddress,
		string(ResourcePrefixRemoteAccessConf): ResourcePrefixRemoteAccessConf,
		string(ResourcePrefixLocalAccount):     ResourcePrefixLocalAccount,
		string(ResourcePrefixTenant):           ResourcePrefixTenant,
		string(ResourcePrefixOsUpdatePolicy):   ResourcePrefixOsUpdatePolicy,
		string(ResourcePrefixCustomConfig):     ResourcePrefixCustomConfig,
		string(ResourcePrefixOsUpdateRun):      ResourcePrefixOsUpdateRun,
	}

	prefix, ok := mapStringToPrefix[s]
	if !ok {
		zlog.InfraSec().InfraError("%s does not match any known ResourcePrefix", s).Msg("")
		return ResourcePrefixUnspecified, errors.Errorfc(codes.InvalidArgument,
			"%s does not match any known ResourcePrefix",
			s,
		)
	}
	return prefix, nil
}

// GetResourceKindFromResourceID extracts the resource kind from a resource ID.
func GetResourceKindFromResourceID(resID string) (inv_v1.ResourceKind, error) {
	prefix, _, found := strings.Cut(resID, "-")
	if !found {
		zlog.InfraSec().InfraError("invalid ResourceID").Msg("")
		return inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED,
			errors.Errorfc(codes.InvalidArgument, "invalid ResourceID")
	}
	typedPrefix, err := stringToPrefix(prefix)
	if err != nil {
		return inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED, err
	}

	return PrefixToResourceKind(typedPrefix), nil
}

// GetResourceIDFromResource extracts the resource ID from a wrapped resource.
//
//nolint:cyclop,funlen // high cyclomatic complexity due to the switch and length due to the switch
func GetResourceIDFromResource(resource *inv_v1.Resource) (string, error) {
	switch resource.GetResource().(type) {
	case *inv_v1.Resource_Region:
		return resource.GetRegion().GetResourceId(), nil
	case *inv_v1.Resource_Site:
		return resource.GetSite().GetResourceId(), nil
	case *inv_v1.Resource_Ou:
		return resource.GetOu().GetResourceId(), nil
	case *inv_v1.Resource_Instance:
		return resource.GetInstance().GetResourceId(), nil
	case *inv_v1.Resource_Host:
		return resource.GetHost().GetResourceId(), nil
	case *inv_v1.Resource_Hoststorage:
		return resource.GetHoststorage().GetResourceId(), nil
	case *inv_v1.Resource_Hostnic:
		return resource.GetHostnic().GetResourceId(), nil
	case *inv_v1.Resource_Hostusb:
		return resource.GetHostusb().GetResourceId(), nil
	case *inv_v1.Resource_Hostgpu:
		return resource.GetHostgpu().GetResourceId(), nil
	case *inv_v1.Resource_NetworkSegment:
		return resource.GetNetworkSegment().GetResourceId(), nil
	case *inv_v1.Resource_Netlink:
		return resource.GetNetlink().GetResourceId(), nil
	case *inv_v1.Resource_Endpoint:
		return resource.GetEndpoint().GetResourceId(), nil
	case *inv_v1.Resource_Ipaddress:
		return resource.GetIpaddress().GetResourceId(), nil
	case *inv_v1.Resource_Provider:
		return resource.GetProvider().GetResourceId(), nil
	case *inv_v1.Resource_Os:
		return resource.GetOs().GetResourceId(), nil
	case *inv_v1.Resource_Singleschedule:
		return resource.GetSingleschedule().GetResourceId(), nil
	case *inv_v1.Resource_Repeatedschedule:
		return resource.GetRepeatedschedule().GetResourceId(), nil
	case *inv_v1.Resource_TelemetryGroup:
		return resource.GetTelemetryGroup().GetResourceId(), nil
	case *inv_v1.Resource_TelemetryProfile:
		return resource.GetTelemetryProfile().GetResourceId(), nil
	case *inv_v1.Resource_Workload:
		return resource.GetWorkload().GetResourceId(), nil
	case *inv_v1.Resource_WorkloadMember:
		return resource.GetWorkloadMember().GetResourceId(), nil
	case *inv_v1.Resource_RemoteAccess:
		return resource.GetRemoteAccess().GetResourceId(), nil
	case *inv_v1.Resource_Tenant:
		return resource.GetTenant().GetResourceId(), nil
	case *inv_v1.Resource_LocalAccount:
		return resource.GetLocalAccount().GetResourceId(), nil
	case *inv_v1.Resource_OsUpdatePolicy:
		return resource.GetOsUpdatePolicy().GetResourceId(), nil
	case *inv_v1.Resource_CustomConfig:
		return resource.GetCustomConfig().GetResourceId(), nil
	case *inv_v1.Resource_OsUpdateRun:
		return resource.GetOsUpdateRun().GetResourceId(), nil
	default:
		zlog.InfraSec().InfraError("unknown Resource type: %T", resource.GetResource()).Msg("")
		return "", errors.Errorfc(codes.InvalidArgument, "unknown Resource type: %T", resource.GetResource())
	}
}

// WrapResource takes a resource and returns it in the generic form.
//
//nolint:cyclop,funlen // high cyclomatic complexity and length due to the switch
func WrapResource(resource proto.Message) (*inv_v1.Resource, error) {
	wrap := &inv_v1.Resource{}
	switch r := resource.(type) {
	case *location_v1.RegionResource:
		wrap.Resource = &inv_v1.Resource_Region{Region: r}
	case *location_v1.SiteResource:
		wrap.Resource = &inv_v1.Resource_Site{Site: r}
	case *ou_v1.OuResource:
		wrap.Resource = &inv_v1.Resource_Ou{Ou: r}
	case *compute_v1.InstanceResource:
		wrap.Resource = &inv_v1.Resource_Instance{Instance: r}
	case *compute_v1.HostResource:
		wrap.Resource = &inv_v1.Resource_Host{Host: r}
	case *compute_v1.HoststorageResource:
		wrap.Resource = &inv_v1.Resource_Hoststorage{Hoststorage: r}
	case *compute_v1.HostnicResource:
		wrap.Resource = &inv_v1.Resource_Hostnic{Hostnic: r}
	case *compute_v1.HostusbResource:
		wrap.Resource = &inv_v1.Resource_Hostusb{Hostusb: r}
	case *compute_v1.HostgpuResource:
		wrap.Resource = &inv_v1.Resource_Hostgpu{Hostgpu: r}
	case *compute_v1.WorkloadResource:
		wrap.Resource = &inv_v1.Resource_Workload{Workload: r}
	case *compute_v1.WorkloadMember:
		wrap.Resource = &inv_v1.Resource_WorkloadMember{WorkloadMember: r}
	case *network_v1.NetworkSegment:
		wrap.Resource = &inv_v1.Resource_NetworkSegment{NetworkSegment: r}
	case *network_v1.NetlinkResource:
		wrap.Resource = &inv_v1.Resource_Netlink{Netlink: r}
	case *network_v1.EndpointResource:
		wrap.Resource = &inv_v1.Resource_Endpoint{Endpoint: r}
	case *network_v1.IPAddressResource:
		wrap.Resource = &inv_v1.Resource_Ipaddress{Ipaddress: r}
	case *provider_v1.ProviderResource:
		wrap.Resource = &inv_v1.Resource_Provider{Provider: r}
	case *os_v1.OperatingSystemResource:
		wrap.Resource = &inv_v1.Resource_Os{Os: r}
	case *schedule_v1.SingleScheduleResource:
		wrap.Resource = &inv_v1.Resource_Singleschedule{Singleschedule: r}
	case *schedule_v1.RepeatedScheduleResource:
		wrap.Resource = &inv_v1.Resource_Repeatedschedule{Repeatedschedule: r}
	case *telemetry_v1.TelemetryGroupResource:
		wrap.Resource = &inv_v1.Resource_TelemetryGroup{TelemetryGroup: r}
	case *telemetry_v1.TelemetryProfile:
		wrap.Resource = &inv_v1.Resource_TelemetryProfile{TelemetryProfile: r}
	case *remoteaccessv1.RemoteAccessConfiguration:
		wrap.Resource = &inv_v1.Resource_RemoteAccess{
			RemoteAccess: r,
		}
	case *tenantv1.Tenant:
		wrap.Resource = &inv_v1.Resource_Tenant{
			Tenant: r,
		}
	case *localaccountv1.LocalAccountResource:
		wrap.Resource = &inv_v1.Resource_LocalAccount{
			LocalAccount: r,
		}
	case *compute_v1.OSUpdatePolicyResource:
		wrap.Resource = &inv_v1.Resource_OsUpdatePolicy{OsUpdatePolicy: r}
	case *compute_v1.CustomConfigResource:
		wrap.Resource = &inv_v1.Resource_CustomConfig{
			CustomConfig: r,
		}
	case *compute_v1.OSUpdateRunResource:
		wrap.Resource = &inv_v1.Resource_OsUpdateRun{OsUpdateRun: r}
	default:
		zlog.InfraSec().InfraError("unknown Resource type: %T", resource).Msg("")
		return nil, errors.Errorfc(codes.InvalidArgument, "unknown Resource type: %T", resource)
	}

	return wrap, nil
}

// UnwrapResource returns the underlying resource given a generic Resource.
// If you don't need a concrete resource as return type, you can still get the
// inner resource as a generic proto message:
//
//	msg, err := UnwrapResource[proto.Message](resource)
func UnwrapResource[T proto.Message](resource *inv_v1.Resource) (T, error) {
	var zero T // Used to return a 'nil' like default object on errors
	rk, err := getResourceProtoMessage(resource)
	if err != nil {
		// This should never happen
		return zero, err
	}
	r, ok := rk.(T)
	if !ok {
		// This should never happen
		zlog.InfraSec().Error().Msgf("error while extracting concrete type")
		return zero, errors.Errorfc(codes.Internal, "error while extracting concrete type")
	}
	return r, nil
}

func GetResourceKindFromMessage(message proto.Message) (inv_v1.ResourceKind, error) {
	mapStringToPrefix := map[string]inv_v1.ResourceKind{
		"unspecified":              inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED,
		"InstanceResource":         inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE,
		"HostResource":             inv_v1.ResourceKind_RESOURCE_KIND_HOST,
		"HoststorageResource":      inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE,
		"HostnicResource":          inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC,
		"HostusbResource":          inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB,
		"HostgpuResource":          inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU,
		"NetworkSegment":           inv_v1.ResourceKind_RESOURCE_KIND_NETWORKSEGMENT,
		"NetlinkResource":          inv_v1.ResourceKind_RESOURCE_KIND_NETLINK,
		"EndpointResource":         inv_v1.ResourceKind_RESOURCE_KIND_ENDPOINT,
		"RegionResource":           inv_v1.ResourceKind_RESOURCE_KIND_REGION,
		"SiteResource":             inv_v1.ResourceKind_RESOURCE_KIND_SITE,
		"OuResource":               inv_v1.ResourceKind_RESOURCE_KIND_OU,
		"ProviderResource":         inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER,
		"OperatingSystemResource":  inv_v1.ResourceKind_RESOURCE_KIND_OS,
		"SingleScheduleResource":   inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE,
		"RepeatedScheduleResource": inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE,
		"TelemetryGroupResource":   inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_GROUP,
		"TelemetryProfile":         inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_PROFILE,
		"WorkloadResource":         inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD,
		"WorkloadMember":           inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD_MEMBER,
		"IPAddressResource":        inv_v1.ResourceKind_RESOURCE_KIND_IPADDRESS,
		"RemoteAccessConfResource": inv_v1.ResourceKind_RESOURCE_KIND_RMT_ACCESS_CONF,
		"LocalAccountResource":     inv_v1.ResourceKind_RESOURCE_KIND_LOCALACCOUNT,
		"CustomConfigResource":     inv_v1.ResourceKind_RESOURCE_KIND_CUSTOMCONFIG,
		string(proto.MessageName(&tenantv1.Tenant{}).Name()):                   inv_v1.ResourceKind_RESOURCE_KIND_TENANT,
		string(proto.MessageName(&compute_v1.OSUpdatePolicyResource{}).Name()): inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATEPOLICY,
		string(proto.MessageName(&compute_v1.OSUpdateRunResource{}).Name()):    inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATERUN,
	}
	resname := string(proto.MessageName(message).Name())
	kind, ok := mapStringToPrefix[resname]
	if !ok {
		zlog.InfraSec().InfraError("%s does not match any known Resource", resname).Msg("")
		return inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED, errors.Errorfc(codes.InvalidArgument,
			"%s does not match any known Resource",
			resname,
		)
	}
	return kind, nil
}

func NewInvID(kind inv_v1.ResourceKind) string {
	// Idea for the future, do we want to recycle the ids instead of allocating new ones?
	n, err := rand.Int(rand.Reader, big.NewInt(math.MaxUint32))
	if err != nil {
		// if we get an error here - the system is in very bad state
		zlog.Fatal().Err(err).Msg("Failed to generate a new inventory id")
		return ""
	}
	return fmt.Sprintf("%s-%08x", ResourceKindToPrefix(kind), n)
}

func ValidateMaskAndFilterMessage(in proto.Message, fieldmask *fieldmaskpb.FieldMask, filter bool) error {
	if fieldmask != nil {
		fieldmask.Normalize()
		if valid := fieldmask.IsValid(in); !valid {
			zlog.InfraSec().InfraError("invalid FieldMask for the given %s",
				in.ProtoReflect().Descriptor().Fields(),
			).Msg("")
			return errors.Errorfc(codes.InvalidArgument, "invalid FieldMask for the given %s",
				in.ProtoReflect().Descriptor().Fields(),
			)
		}
		if filter {
			// Filter the input applying the fieldmask in order to be sure to update only required paths
			fmutils.Filter(in, fieldmask.GetPaths())
		}
	}
	return nil
}

// BuildFieldMaskFromMessage builds a field mask containing all fields set in the given message.
func BuildFieldMaskFromMessage(message proto.Message, skipFields ...string) (*fieldmaskpb.FieldMask, error) {
	mpr := message.ProtoReflect()
	var fields []string
	mpr.Range(func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
		if !slices.Contains(skipFields, string(fd.Name())) {
			fields = append(fields, string(fd.Name()))
		}
		return true
	})
	fm, err := fieldmaskpb.New(message, fields...)
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msg("")
		return nil, err
	}
	return fm, nil
}

// BuildAllFieldMaskFromProto builds a fieldmask containing all fields for the given proto Message type.
func BuildAllFieldMaskFromProto(message proto.Message, skipFields ...string) (*fieldmaskpb.FieldMask, error) {
	var fields []string

	for i := 0; i < message.ProtoReflect().Descriptor().Fields().Len(); i++ {
		fieldName := string(message.ProtoReflect().Descriptor().Fields().Get(i).Name())
		if !slices.Contains(skipFields, fieldName) {
			fields = append(fields, fieldName)
		}
	}
	fm, err := fieldmaskpb.New(message, fields...)
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msg("")
		return nil, err
	}
	return fm, nil
}

func UintToInt(u uint) (int, error) {
	if u > math.MaxInt {
		return 0, errors.Errorfc(codes.InvalidArgument, "uint value exceeds int range")
	}
	return int(u), nil
}

func Uint64ToInt(u uint64) (int, error) {
	if u > math.MaxInt64 {
		return 0, errors.Errorfc(codes.InvalidArgument, "uint64 value exceeds int range")
	}
	return int(u), nil
}

func IntToUint64(i int) (uint64, error) {
	if i < 0 {
		return 0, errors.Errorfc(codes.InvalidArgument, "int value is negative and cannot be converted to uint64")
	}
	return uint64(i), nil
}

// IntToUint32 safely converts int to uint32. Returns an error when the value is out of the range.
func IntToUint32(i int) (uint32, error) {
	if i < 0 {
		return 0, errors.Errorfc(codes.InvalidArgument, "int value is negative and cannot be converted to uint32")
	}
	if i > math.MaxUint32 {
		return 0, errors.Errorfc(codes.InvalidArgument, "int value exceeds uint32 range")
	}
	res := uint32(i)
	if int(res) != i {
		zlog.InfraSec().InfraError("%#v of type int is out of range for uint32", i).Msg("")
		return 0, errors.Errorfc(codes.InvalidArgument, "%#v of type int is out of range for uint32", i)
	}
	return res, nil
}

// Uint64ToUint32 safely converts uint64 to uint32. Returns an error when the value is out of the range.
func Uint64ToUint32(i uint64) (uint32, error) {
	if i > math.MaxUint32 {
		return 0, errors.Errorfc(codes.InvalidArgument, "uint64 value exceeds uint32 range")
	}
	res := uint32(i)
	if uint64(res) != i {
		zlog.InfraSec().InfraError("%#v of type uint64 is out of range for uint32", i).Msg("")
		return 0, errors.Errorfc(codes.InvalidArgument, "%#v of type uint64 is out of range for uint32", i)
	}
	return res, nil
}

func Int32ToInt(i int32) (int, error) {
	res := int(i)
	if res < math.MaxInt32 && int32(res) != i { //nolint:gosec // no risk of overflow
		zlog.InfraSec().InfraError("%#v of type int32 is out of range for int", i).Msg("")
		return 0, errors.Errorfc(codes.InvalidArgument, "%#v of type int32 is out of range for int", i)
	}
	return res, nil
}

// IntToInt32 safely converts int to int32. This is needed for 64bit systems where int is defined as a 64bit integer.
// Returns an error when the value is out of the range.
func IntToInt32(i int) (int32, error) {
	if i < math.MinInt32 || i > math.MaxInt32 {
		return 0, errors.Errorfc(codes.InvalidArgument, "int value exceeds int32 range")
	}
	res := int32(i)
	if int(res) != i {
		zlog.InfraSec().InfraError("%#v of type int is out of range for int32", i).Msg("")
		return 0, errors.Errorfc(codes.InvalidArgument, "%#v of type int is out of range for int32", i)
	}
	return res, nil
}

// Uint32ToInt safely converts uint32 to int. This is needed for 32bit systems where int is defined as a 32bit integer.
// Returns an error when the value is out of the range.
func Uint32ToInt(i uint32) (int, error) {
	if i > math.MaxInt32 {
		return 0, errors.Errorfc(codes.InvalidArgument, "uint32 value exceeds int range")
	}
	res := int(i)
	if uint32(res) != i { //nolint:gosec // no risk of overflow
		zlog.InfraSec().InfraError("%#v of type uint32 is out of range for int", i).Msg("")
		return 0, errors.Errorfc(codes.InvalidArgument, "%#v of type uint32 is out of range for int", i)
	}
	return res, nil
}

// Int64ToInt32 safely converts int64 to int32. Returns an error when the value is out of the range.
func Int64ToInt32(i int64) (int32, error) {
	if i < math.MinInt32 || i > math.MaxInt32 {
		return 0, errors.Errorfc(codes.InvalidArgument, "int64 value exceeds int32 range")
	}
	res := int32(i)
	if int64(res) != i {
		zlog.InfraSec().InfraError("%#v of type int64 is out of range for int32", i).Msg("")
		return 0, errors.Errorfc(codes.InvalidArgument, "%#v of type int64 is out of range for int32", i)
	}
	return res, nil
}

// Int64ToInt safely converts int64 to int. This is needed for 32bit systems where int is defined as a 32bit integer.
// Returns an error when the value is out of the range.
func Int64ToInt(i int64) (int, error) {
	if i < math.MinInt || i > math.MaxInt {
		return 0, errors.Errorfc(codes.InvalidArgument, "int64 value exceeds int range")
	}
	res := int(i)
	if int64(res) != i {
		zlog.InfraSec().InfraError("%#v of type int64 is out of range for int", i).Msg("")
		return 0, errors.Errorfc(codes.InvalidArgument, "%#v of type int64 is out of range for int", i)
	}
	return res, nil
}

func Uint64ToInt64(u uint64) (int64, error) {
	if u > math.MaxInt64 {
		return 0, errors.Errorfc(codes.InvalidArgument, "uint64 value exceeds int64 range")
	}
	return int64(u), nil
}

func Int64ToUint64(i int64) (uint64, error) {
	if i < 0 {
		return 0, errors.Errorfc(codes.InvalidArgument, "int64 value is negative and cannot be converted to uint64")
	}
	return uint64(i), nil
}

func signedMulOverflows(l, r int64) bool {
	if l == 0 || r == 0 || l == 1 || r == 1 {
		return false
	}
	if l == math.MinInt64 || r == math.MinInt64 {
		return true
	}
	c := l * r
	return c/r != l
}

func MulInt64(left, right int64) (int64, error) {
	if signedMulOverflows(left, right) {
		return 0, errors.Errorfc(codes.InvalidArgument, "overflow while multiplying int64 (%#v*%#v)", left, right)
	}
	return left * right, nil
}

type DBEnv struct {
	Host     string
	Port     string
	Database string
	User     string
	Pass     string
	SslMode  string
}

func (env DBEnv) String() string {
	t := env
	t.Pass = ""
	return fmt.Sprintf("%#v", t)
}

// LookupDBEnv fetches the database configuration provided via k8s secret in env.
// Configuration for the Reader can be provided via the "_RO" env variables.
// If the PGHOST_RO is missing, the DBEnv for the RO won't be populated. If any other _RO env variables are missing
// the function will return an error.
//
//nolint:cyclop // complexity is high due to high number of ifs to correctly parse env variables
func LookupDBEnv() (*DBEnv, *DBEnv, error) {
	const (
		databaseHost    = "PGHOST"
		databasePort    = "PGPORT"
		databaseName    = "PGDATABASE"
		databaseUser    = "PGUSER"
		databasePwd     = "PGPASSWORD"
		databaseSslMode = "PGSSLMODE"
		databaseHostRO  = "PGHOST_RO"
		databasePortRO  = "PGPORT_RO"
		databaseNameRO  = "PGDATABASE_RO"
		databaseUserRO  = "PGUSER_RO"
		databasePwdRO   = "PGPASSWORD_RO"
	)
	env := &DBEnv{}
	var ok bool
	if env.Host, ok = os.LookupEnv(databaseHost); !ok {
		zlog.InfraSec().InfraError("%s env var is not set", databaseHost).Msg("")
		return nil, nil, status.Errorf(codes.InvalidArgument, "%s env var is not set", databaseHost)
	}
	if env.Port, ok = os.LookupEnv(databasePort); !ok {
		zlog.InfraSec().InfraError("%s env var is not set", databasePort).Msg("")
		return nil, nil, status.Errorf(codes.InvalidArgument, "%s env var is not set", databasePort)
	}
	if env.Database, ok = os.LookupEnv(databaseName); !ok {
		zlog.InfraSec().InfraError("%s env var is not set", databaseName).Msg("")
		return nil, nil, status.Errorf(codes.InvalidArgument, "%s env var is not set", databaseName)
	}
	if env.User, ok = os.LookupEnv(databaseUser); !ok {
		zlog.InfraSec().InfraError("%s env var is not set", databaseUser).Msg("")
		return nil, nil, status.Errorf(codes.InvalidArgument, "%s env var is not set", databaseUser)
	}
	if env.Pass, ok = os.LookupEnv(databasePwd); !ok {
		zlog.InfraSec().InfraError("%s env var is not set", databasePwd).Msg("")
		return nil, nil, status.Errorf(codes.InvalidArgument, "%s env var is not set", databasePwd)
	}
	if env.SslMode, ok = os.LookupEnv(databaseSslMode); !ok {
		zlog.InfraSec().InfraError("%s env var is not set", databasePwd).Msg("")
		return nil, nil, status.Errorf(codes.InvalidArgument, "%s env var is not set", databaseSslMode)
	}

	envRO := &DBEnv{}
	envRO.SslMode = env.SslMode
	if envRO.Host, ok = os.LookupEnv(databaseHostRO); !ok {
		zlog.Warn().Msgf("%s env var is not set, no read-only replica set", databaseHostRO)
		envRO = nil
	} else {
		if envRO.Port, ok = os.LookupEnv(databasePortRO); !ok {
			zlog.InfraSec().InfraError("%s env var is not set", databasePortRO).Msg("")
			return nil, nil, status.Errorf(codes.InvalidArgument, "%s env var is not set", databasePortRO)
		}
		if envRO.Database, ok = os.LookupEnv(databaseNameRO); !ok {
			zlog.InfraSec().InfraError("%s env var is not set", databaseNameRO).Msg("")
			return nil, nil, status.Errorf(codes.InvalidArgument, "%s env var is not set", databaseNameRO)
		}
		if envRO.User, ok = os.LookupEnv(databaseUserRO); !ok {
			zlog.InfraSec().InfraError("%s env var is not set", databaseUserRO).Msg("")
			return nil, nil, status.Errorf(codes.InvalidArgument, "%s env var is not set", databaseUserRO)
		}
		if envRO.Pass, ok = os.LookupEnv(databasePwdRO); !ok {
			zlog.InfraSec().InfraError("%s env var is not set", databasePwdRO).Msg("")
			return nil, nil, status.Errorf(codes.InvalidArgument, "%s env var is not set", databasePwdRO)
		}
	}

	return env, envRO, nil
}

func LookupDBTestEnv() *DBEnv {
	return &DBEnv{
		Host:     "localhost",
		Port:     "5432",
		Database: "postgres",
		User:     "admin",
		Pass:     "pass",
		SslMode:  "disable",
	}
}

func GetDBURL(env *DBEnv) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?search_path=public&sslmode=%s",
		env.User, env.Pass, env.Host, env.Port, env.Database, env.SslMode)
}

// GetResourceFromKind Get a Resource with the given Resource kind set. Useful when filtering without any specified filter
// to get all resources of a given kind.
func GetResourceFromKind(resourceType inv_v1.ResourceKind) (*inv_v1.Resource, error) {
	invResMap := map[inv_v1.ResourceKind]*inv_v1.Resource{
		inv_v1.ResourceKind_RESOURCE_KIND_REGION: {Resource: &inv_v1.Resource_Region{}},
		inv_v1.ResourceKind_RESOURCE_KIND_SITE:   {Resource: &inv_v1.Resource_Site{}},

		inv_v1.ResourceKind_RESOURCE_KIND_OU: {Resource: &inv_v1.Resource_Ou{}},

		inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE: {Resource: &inv_v1.Resource_Instance{}},

		inv_v1.ResourceKind_RESOURCE_KIND_HOST:        {Resource: &inv_v1.Resource_Host{}},
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE: {Resource: &inv_v1.Resource_Hoststorage{}},
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC:     {Resource: &inv_v1.Resource_Hostnic{}},
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB:     {Resource: &inv_v1.Resource_Hostusb{}},
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU:     {Resource: &inv_v1.Resource_Hostgpu{}},

		inv_v1.ResourceKind_RESOURCE_KIND_NETWORKSEGMENT: {Resource: &inv_v1.Resource_NetworkSegment{}},
		inv_v1.ResourceKind_RESOURCE_KIND_NETLINK:        {Resource: &inv_v1.Resource_Netlink{}},
		inv_v1.ResourceKind_RESOURCE_KIND_ENDPOINT:       {Resource: &inv_v1.Resource_Endpoint{}},
		inv_v1.ResourceKind_RESOURCE_KIND_IPADDRESS:      {Resource: &inv_v1.Resource_Ipaddress{}},

		inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER: {Resource: &inv_v1.Resource_Provider{}},

		inv_v1.ResourceKind_RESOURCE_KIND_OS: {Resource: &inv_v1.Resource_Os{}},

		inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE:   {Resource: &inv_v1.Resource_Singleschedule{}},
		inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE: {Resource: &inv_v1.Resource_Repeatedschedule{}},

		inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_GROUP:   {Resource: &inv_v1.Resource_TelemetryGroup{}},
		inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_PROFILE: {Resource: &inv_v1.Resource_TelemetryProfile{}},

		inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD:        {Resource: &inv_v1.Resource_Workload{}},
		inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD_MEMBER: {Resource: &inv_v1.Resource_WorkloadMember{}},

		inv_v1.ResourceKind_RESOURCE_KIND_RMT_ACCESS_CONF: {Resource: &inv_v1.Resource_RemoteAccess{}},
		inv_v1.ResourceKind_RESOURCE_KIND_TENANT:          {Resource: &inv_v1.Resource_Tenant{}},
		inv_v1.ResourceKind_RESOURCE_KIND_LOCALACCOUNT:    {Resource: &inv_v1.Resource_LocalAccount{}},

		inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATEPOLICY: {Resource: &inv_v1.Resource_OsUpdatePolicy{}},
		inv_v1.ResourceKind_RESOURCE_KIND_CUSTOMCONFIG:   {Resource: &inv_v1.Resource_CustomConfig{}},
		inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATERUN:    {Resource: &inv_v1.Resource_OsUpdateRun{}},
	}
	if res, ok := invResMap[resourceType]; ok {
		return res, nil
	}
	err := errors.Errorfc(codes.InvalidArgument, "unsupported resource kind %s", resourceType)
	zlog.InfraSec().InfraErr(err).Msg("")
	return nil, err
}

// GetSetResource returns the set resource as proto message.
func GetSetResource(resource *inv_v1.Resource) (proto.Message, error) {
	kind := GetResourceKindFromResource(resource)
	var resProto proto.Message

	kindToResource := map[inv_v1.ResourceKind]func(*inv_v1.Resource) proto.Message{
		inv_v1.ResourceKind_RESOURCE_KIND_REGION: func(r *inv_v1.Resource) proto.Message { return r.GetRegion() },
		inv_v1.ResourceKind_RESOURCE_KIND_SITE:   func(r *inv_v1.Resource) proto.Message { return r.GetSite() },

		inv_v1.ResourceKind_RESOURCE_KIND_OU: func(r *inv_v1.Resource) proto.Message { return r.GetOu() },

		inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE: func(r *inv_v1.Resource) proto.Message { return r.GetInstance() },

		inv_v1.ResourceKind_RESOURCE_KIND_HOST: func(r *inv_v1.Resource) proto.Message {
			return r.GetHost()
		},
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE: func(r *inv_v1.Resource) proto.Message {
			return r.GetHoststorage()
		},
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC: func(r *inv_v1.Resource) proto.Message {
			return r.GetHostnic()
		},
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB: func(r *inv_v1.Resource) proto.Message {
			return r.GetHostusb()
		},
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU: func(r *inv_v1.Resource) proto.Message {
			return r.GetHostgpu()
		},

		inv_v1.ResourceKind_RESOURCE_KIND_NETWORKSEGMENT: func(r *inv_v1.Resource) proto.Message {
			return r.GetNetworkSegment()
		},
		inv_v1.ResourceKind_RESOURCE_KIND_NETLINK: func(r *inv_v1.Resource) proto.Message {
			return r.GetNetlink()
		},
		inv_v1.ResourceKind_RESOURCE_KIND_ENDPOINT: func(r *inv_v1.Resource) proto.Message {
			return r.GetEndpoint()
		},
		inv_v1.ResourceKind_RESOURCE_KIND_IPADDRESS: func(r *inv_v1.Resource) proto.Message {
			return r.GetIpaddress()
		},

		inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER: func(r *inv_v1.Resource) proto.Message { return r.GetProvider() },

		inv_v1.ResourceKind_RESOURCE_KIND_OS: func(r *inv_v1.Resource) proto.Message { return r.GetOs() },

		inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE: func(r *inv_v1.Resource) proto.Message {
			return r.GetSingleschedule()
		},
		inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE: func(r *inv_v1.Resource) proto.Message {
			return r.GetRepeatedschedule()
		},

		inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_GROUP: func(r *inv_v1.Resource) proto.Message {
			return r.GetTelemetryGroup()
		},
		inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_PROFILE: func(r *inv_v1.Resource) proto.Message {
			return r.GetTelemetryProfile()
		},

		inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD: func(r *inv_v1.Resource) proto.Message {
			return r.GetWorkload()
		},
		inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD_MEMBER: func(r *inv_v1.Resource) proto.Message {
			return r.GetWorkloadMember()
		},

		inv_v1.ResourceKind_RESOURCE_KIND_RMT_ACCESS_CONF: func(r *inv_v1.Resource) proto.Message {
			return r.GetRemoteAccess()
		},
		inv_v1.ResourceKind_RESOURCE_KIND_TENANT: func(r *inv_v1.Resource) proto.Message {
			return r.GetTenant()
		},
		inv_v1.ResourceKind_RESOURCE_KIND_LOCALACCOUNT: func(r *inv_v1.Resource) proto.Message {
			return r.GetLocalAccount()
		},

		inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATEPOLICY: func(r *inv_v1.Resource) proto.Message {
			return r.GetOsUpdatePolicy()
		},
		inv_v1.ResourceKind_RESOURCE_KIND_CUSTOMCONFIG: func(r *inv_v1.Resource) proto.Message {
			return r.GetCustomConfig()
		},
		inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATERUN: func(r *inv_v1.Resource) proto.Message {
			return r.GetOsUpdateRun()
		},
	}
	convert, ok := kindToResource[kind]
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument, "unsupported resource kind %s", kind)
		zlog.InfraSec().InfraErr(err).Msg("")
		return nil, err
	}
	resProto = convert(resource)
	return resProto, nil
}

// getResourceProtoMessage returns "oneof" the proto message given
// the generic resource message provided as input.
//
//nolint:cyclop,funlen // high cyclomatic complexity and length due to the switch
func getResourceProtoMessage(resource *inv_v1.Resource) (proto.Message, error) {
	var message proto.Message
	switch resource.GetResource().(type) {
	case *inv_v1.Resource_Region:
		message = resource.GetRegion()
	case *inv_v1.Resource_Site:
		message = resource.GetSite()
	case *inv_v1.Resource_Ou:
		message = resource.GetOu()
	case *inv_v1.Resource_Instance:
		message = resource.GetInstance()
	case *inv_v1.Resource_Host:
		message = resource.GetHost()
	case *inv_v1.Resource_Hoststorage:
		message = resource.GetHoststorage()
	case *inv_v1.Resource_Hostnic:
		message = resource.GetHostnic()
	case *inv_v1.Resource_Hostusb:
		message = resource.GetHostusb()
	case *inv_v1.Resource_Hostgpu:
		message = resource.GetHostgpu()
	case *inv_v1.Resource_NetworkSegment:
		message = resource.GetNetworkSegment()
	case *inv_v1.Resource_Netlink:
		message = resource.GetNetlink()
	case *inv_v1.Resource_Endpoint:
		message = resource.GetEndpoint()
	case *inv_v1.Resource_Ipaddress:
		message = resource.GetIpaddress()
	case *inv_v1.Resource_Provider:
		message = resource.GetProvider()
	case *inv_v1.Resource_Os:
		message = resource.GetOs()
	case *inv_v1.Resource_Singleschedule:
		message = resource.GetSingleschedule()
	case *inv_v1.Resource_Repeatedschedule:
		message = resource.GetRepeatedschedule()
	case *inv_v1.Resource_TelemetryGroup:
		message = resource.GetTelemetryGroup()
	case *inv_v1.Resource_TelemetryProfile:
		message = resource.GetTelemetryProfile()
	case *inv_v1.Resource_Workload:
		message = resource.GetWorkload()
	case *inv_v1.Resource_WorkloadMember:
		message = resource.GetWorkloadMember()
	case *inv_v1.Resource_RemoteAccess:
		message = resource.GetRemoteAccess()
	case *inv_v1.Resource_Tenant:
		message = resource.GetTenant()
	case *inv_v1.Resource_LocalAccount:
		message = resource.GetLocalAccount()
	case *inv_v1.Resource_OsUpdatePolicy:
		message = resource.GetOsUpdatePolicy()
	case *inv_v1.Resource_CustomConfig:
		message = resource.GetCustomConfig()
	case *inv_v1.Resource_OsUpdateRun:
		message = resource.GetOsUpdateRun()
	default:
		zlog.InfraSec().InfraError("unknown Resource type: %T", resource.GetResource()).Msg("")
		return nil, errors.Errorfc(codes.InvalidArgument, "unknown Resource type: %T", resource.GetResource())
	}

	return message, nil
}

// GetSpecificResourceList returns the underlying list of resources given a generic Resource.
func GetSpecificResourceList[T proto.Message](resources []*inv_v1.Resource) ([]T, error) {
	resRet := make([]T, 0, len(resources))
	for _, res := range resources {
		r, err := UnwrapResource[T](res)
		if err != nil {
			return nil, err
		}
		resRet = append(resRet, r)
	}
	return resRet, nil
}

// BuildNestedFieldMaskFromFields joins the given fields with "." to build a nested field mask.
func BuildNestedFieldMaskFromFields(fields ...string) string {
	return strings.Join(fields, ".")
}

// CheckListOutputIsSingular checks if result returned by list function contains single entry.
// If not, it generates appropriate error with correct code.
func CheckListOutputIsSingular[T any](res []T) error {
	if len(res) == 0 {
		return errors.Errorfc(codes.NotFound, "No Resources found")
	}
	if len(res) != 1 {
		return errors.Errorfc(codes.Internal, "Obtained multiple (%d) Resources, but expected a single one", len(res))
	}
	return nil
}

// ConvertRawUUIDToInventoryUUID function parses the raw UUID to the Inventory-compatible format.
func ConvertRawUUIDToInventoryUUID(rawUUID string) (string, error) {
	inventoryUUID, err := uuid.Parse(rawUUID)
	if err != nil {
		newErr := errors.Errorfc(codes.Internal, "Failed to parse UUID: %v", err)
		zlog.InfraErr(newErr).Msgf("")
		return "", newErr
	}
	return inventoryUUID.String(), nil
}

// ConvertInventoryUUIDToLenovoUUID function converts the Inventory UUID to the raw UUID by
// removing all hyphens and putting all characters in the uppercase.
func ConvertInventoryUUIDToLenovoUUID(uuidStr string) string {
	return strings.ToUpper(strings.ReplaceAll(uuidStr, "-", ""))
}

type resourceKeyCarrier interface {
	GetTenantId() string
	GetResourceId() string
}

func GetResourceKeyFromResource(resource *inv_v1.Resource) (tenantID, resourceID string, err error) {
	internalRes, err := getResourceProtoMessage(resource)
	if err != nil {
		return "", "", err
	}
	carrier, ok := internalRes.(resourceKeyCarrier)
	if !ok {
		// This error should never happen
		err = errors.Errorfc(codes.InvalidArgument, "Resource does not implement resourceKeyCarrier interface")
		zlog.Err(err).Msgf("This error should never happen: resource=%v", resource)
		return "", "", err
	}
	return carrier.GetTenantId(), carrier.GetResourceId(), nil
}

type Tuple[A, B any] struct {
	A A
	B B
}

func NewTuple[A, B any](a A, b B) *Tuple[A, B] {
	return &Tuple[A, B]{
		A: a,
		B: b,
	}
}

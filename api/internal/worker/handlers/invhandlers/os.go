// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package invhandlers

import (
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/api/internal/types"
	"github.com/open-edge-platform/infra-core/api/internal/worker/clients"
	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

// OpenAPIOSResourceToProto maps OpenAPI fields name to Proto fields name.
// The key is derived from the json property respectively of the
// structs OSResource defined in edge-infrastructure-manager-openapi-types.gen.go.
var OpenAPIOSResourceToProto = map[string]string{
	"name":              osv1.OperatingSystemResourceFieldName,
	"architecture":      osv1.OperatingSystemResourceFieldArchitecture,
	"installedPackages": osv1.OperatingSystemResourceFieldInstalledPackages,
	"description":       osv1.OperatingSystemResourceFieldDescription,
}

// OpenAPIOSToProtoExcluded defines exclusion rules as there are some fields
// defined in the OpenAPI spec that are not currently mapped to the proto
// fields.
var OpenAPIOSToProtoExcluded = map[string]struct{}{
	"osResourceID":    {}, // osResourceID must not be set from the API
	"resourceId":      {}, // resourceID must not be set from the API
	"securityFeature": {}, // immutable field
	"profileName":     {}, // immutable field
	"profileVersion":  {}, // immutable field
	"sha256":          {}, // immutable field
	"imageId":         {}, // immutable field
	"osType":          {}, // immutable field
	"repoUrl":         {}, // immutable field
	"imageUrl":        {}, // immutable field
	"osProvider":      {}, // immutable field
	"platformBundle":  {}, // read-only field
	"timestamps":      {}, // read-only field
}

func NewOSHandler(invClient *clients.InventoryClientHandler) InventoryResource {
	return &osHandler{invClient: invClient}
}

type osHandler struct {
	invClient *clients.InventoryClientHandler
}

func (h *osHandler) Create(job *types.Job) (*types.Payload, error) {
	body, err := castOSAPI(&job.Payload)
	if err != nil {
		return nil, err
	}

	os, err := openapiToGrpcOSResource(body)
	if err != nil {
		return nil, err
	}

	req := &inventory.Resource{
		Resource: &inventory.Resource_Os{
			Os: os,
		},
	}

	invResp, err := h.invClient.InvClient.Create(job.Context, req)
	if err != nil {
		return nil, err
	}
	createdOs := invResp.GetOs()

	obj := grpcToOpenAPIOSResource(createdOs, nil)

	return &types.Payload{Data: obj}, err
}

func (h *osHandler) Get(job *types.Job) (*types.Payload, error) {
	req, err := osResourceID(&job.Payload)
	if err != nil {
		return nil, err
	}

	invResp, err := h.invClient.InvClient.Get(job.Context, req)
	if err != nil {
		return nil, err
	}

	os, meta, err := castToOSResource(invResp)
	if err != nil {
		return nil, err
	}

	obj := grpcToOpenAPIOSResource(os, meta)

	return &types.Payload{Data: obj}, nil
}

func (h *osHandler) Update(job *types.Job) (*types.Payload, error) {
	resID, err := osResourceID(&job.Payload)
	if err != nil {
		return nil, err
	}

	fm, err := osFieldMask(&job.Payload, job.Operation)
	if err != nil {
		return nil, err
	}

	res, err := osResource(&job.Payload)
	if err != nil {
		return nil, err
	}

	invResp, err := h.invClient.InvClient.Update(job.Context, resID, fm, res)
	if err != nil {
		return nil, err
	}

	updatedOs := invResp.GetOs()
	obj := grpcToOpenAPIOSResource(updatedOs, nil)
	obj.OsResourceID = &resID // to be removed
	obj.ResourceId = &resID

	return &types.Payload{Data: obj}, nil
}

func (h *osHandler) Delete(job *types.Job) error {
	req, err := osResourceID(&job.Payload)
	if err != nil {
		return err
	}

	_, err = h.invClient.InvClient.Delete(job.Context, req)
	if err != nil {
		return err
	}

	return nil
}

func (h *osHandler) List(job *types.Job) (*types.Payload, error) {
	filter, err := osFilter(&job.Payload)
	if err != nil {
		return nil, err
	}

	resp, err := h.invClient.InvClient.List(job.Context, filter)
	if err != nil {
		return nil, err
	}

	osResource := make([]api.OperatingSystemResource, 0, len(resp.GetResources()))

	for _, res := range resp.GetResources() {
		os, meta, err := castToOSResource(res)
		if err != nil {
			return nil, err
		}
		obj := grpcToOpenAPIOSResource(os, meta)
		osResource = append(osResource, *obj)
	}

	hasNext := resp.GetHasNext()
	totalElems := int(resp.GetTotalElements())
	osResourceList := api.OperatingSystemResourceList{
		OperatingSystemResources: &osResource,
		HasNext:                  &hasNext,
		TotalElements:            &totalElems,
	}

	payload := &types.Payload{Data: osResourceList}
	return payload, nil
}

func castOSAPI(payload *types.Payload) (*api.OperatingSystemResource, error) {
	body, ok := payload.Data.(*api.OperatingSystemResource)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument,
			"body format is not OperatingSystemResource: %T",
			payload.Data,
		)
		log.InfraErr(err).Msgf("")
		return nil, err
	}
	return body, nil
}

func osResource(payload *types.Payload) (*inventory.Resource, error) {
	body, err := castOSAPI(payload)
	if err != nil {
		return nil, err
	}

	os, err := openapiToGrpcOSResource(body)
	if err != nil {
		return nil, err
	}

	req := &inventory.Resource{
		Resource: &inventory.Resource_Os{
			Os: os,
		},
	}
	return req, nil
}

func osFilter(payload *types.Payload) (*inventory.ResourceFilter, error) {
	req := &inventory.ResourceFilter{
		Resource: &inventory.Resource{Resource: &inventory.Resource_Os{Os: &osv1.OperatingSystemResource{}}},
	}
	if payload.Data != nil {
		query, ok := payload.Data.(api.GetOSResourcesParams)
		if !ok {
			err := errors.Errorfc(codes.InvalidArgument,
				"GetOSResourcesParams incorrectly formatted: %T",
				payload.Data,
			)
			log.InfraErr(err).Msg("list operation")
			return nil, err
		}
		err := castOSQueryList(&query, req)
		if err != nil {
			log.Debug().Msgf("error parsing query parameters in list operation: %s",
				err.Error())
			return nil, err
		}
	}

	if err := validator.ValidateMessage(req); err != nil {
		log.InfraSec().InfraErr(err).Msg("failed to validate query params")
		return nil, errors.Wrap(err)
	}
	return req, nil
}

func osResourceID(payload *types.Payload) (string, error) {
	params, ok := payload.Params.(OSResourceURLParams)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument, "OSResourceURLParams incorrectly formatted: %T",
			payload.Data)
		log.InfraErr(err).Msg("could not parse job payload params")
		return "", err
	}
	return params.OSResourceID, nil
}

func osFieldMask(payload *types.Payload, operation types.Operation) (*fieldmaskpb.FieldMask, error) {
	body, ok := payload.Data.(*api.OperatingSystemResource)
	if !ok {
		err := errors.Errorfc(codes.InvalidArgument,
			"body format is not OperatingSystemResource: %T",
			payload.Data,
		)
		log.InfraErr(err).Msgf("")
		return nil, err
	}

	osRes, err := osResource(payload)
	if err != nil {
		return nil, err
	}

	var fieldmask *fieldmaskpb.FieldMask
	if operation == types.Patch {
		fieldmask, err = getOSResourceFieldmask(body)
	} else {
		fieldmask, err = fieldmaskpb.New(osRes.GetOs(), maps.Values(OpenAPIOSResourceToProto)...)
	}
	if err != nil {
		log.InfraErr(err).Msgf("could not create fieldmask")
		return nil, errors.Wrap(err)
	}

	return fieldmask, nil
}

func castOSQueryList(
	query *api.GetOSResourcesParams,
	req *inventory.ResourceFilter,
) error {
	err := error(nil)
	req.Limit, req.Offset, err = parsePagination(
		query.PageSize,
		query.Offset,
	)
	if err != nil {
		return err
	}
	if query.Filter != nil {
		req.Filter = *query.Filter
	}
	if query.OrderBy != nil {
		req.OrderBy = *query.OrderBy
	}
	return nil
}

// helpers method to convert between API formats.
func castToOSResource(resp *inventory.GetResourceResponse) (
	*osv1.OperatingSystemResource, *inventory.GetResourceResponse_ResourceMetadata, error,
) {
	if resp.GetResource().GetOs() != nil {
		return resp.GetResource().GetOs(), resp.GetRenderedMetadata(), nil
	}
	err := errors.Errorfc(codes.Internal, "%s is not a OperatingSystemResourceResource", resp.GetResource())
	log.InfraErr(err).Msgf("could not cast inventory resource")
	return nil, nil, err
}

func getOSResourceFieldmask(body *api.OperatingSystemResource) (*fieldmaskpb.FieldMask, error) {
	fieldList := getProtoFieldListFromOpenapiValue(*body, OpenAPIOSResourceToProto)
	log.Debug().Msgf("Proto Valid Fields: %s", fieldList)
	return fieldmaskpb.New(&osv1.OperatingSystemResource{}, fieldList...)
}

func grpcSecurityFeatureToOpenAPISecurityFeature(s osv1.SecurityFeature) api.SecurityFeature {
	switch s {
	case osv1.SecurityFeature_SECURITY_FEATURE_NONE:
		return api.SECURITYFEATURENONE
	case osv1.SecurityFeature_SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION:
		return api.SECURITYFEATURESECUREBOOTANDFULLDISKENCRYPTION
	default:
		return api.SECURITYFEATUREUNSPECIFIED
	}
}

func grpcOsTypeToOpenAPIOSType(osType osv1.OsType) api.OperatingSystemType {
	switch osType {
	case osv1.OsType_OS_TYPE_MUTABLE:
		return api.OPERATINGSYSTEMTYPEMUTABLE
	case osv1.OsType_OS_TYPE_IMMUTABLE:
		return api.OPERATINGSYSTEMTYPEIMMUTABLE
	default:
		return api.OPERATINGSYSTEMTYPEUNSPECIFIED
	}
}

func grpcOsProviderToOpenAPIOsProvider(osProvider osv1.OsProviderKind) api.OperatingSystemProvider {
	switch osProvider {
	case osv1.OsProviderKind_OS_PROVIDER_KIND_INFRA:
		return api.OPERATINGSYSTEMPROVIDERINFRA
	case osv1.OsProviderKind_OS_PROVIDER_KIND_LENOVO:
		return api.OPERATINGSYSTEMPROVIDERLENOVO
	default:
		return api.OPERATINGSYSTEMPROVIDERUNSPECIFIED
	}
}

func openAPISecurityFeatureTogrpcSecurityFeature(s *api.SecurityFeature) osv1.SecurityFeature {
	if s != nil {
		switch *s {
		case api.SECURITYFEATURENONE:
			return osv1.SecurityFeature_SECURITY_FEATURE_NONE
		case api.SECURITYFEATURESECUREBOOTANDFULLDISKENCRYPTION:
			return osv1.SecurityFeature_SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION
		default:
			return osv1.SecurityFeature_SECURITY_FEATURE_UNSPECIFIED
		}
	}
	return osv1.SecurityFeature_SECURITY_FEATURE_UNSPECIFIED
}

func openAPIOperatingSystemTypeToGrpcOsType(t api.OperatingSystemType) osv1.OsType {
	switch t {
	case api.OPERATINGSYSTEMTYPEMUTABLE:
		return osv1.OsType_OS_TYPE_MUTABLE
	case api.OPERATINGSYSTEMTYPEIMMUTABLE:
		return osv1.OsType_OS_TYPE_IMMUTABLE
	default:
		return osv1.OsType_OS_TYPE_UNSPECIFIED
	}
}

func openAPIOperatingSystemProviderToGrpcOsProvider(p api.OperatingSystemProvider) osv1.OsProviderKind {
	switch p {
	case api.OPERATINGSYSTEMPROVIDERINFRA:
		return osv1.OsProviderKind_OS_PROVIDER_KIND_INFRA
	case api.OPERATINGSYSTEMPROVIDERLENOVO:
		return osv1.OsProviderKind_OS_PROVIDER_KIND_LENOVO
	default:
		return osv1.OsProviderKind_OS_PROVIDER_KIND_UNSPECIFIED
	}
}

//nolint:cyclop // cyclomatic complexity is 11
func openapiToGrpcOSResource(body *api.OperatingSystemResource) (*osv1.OperatingSystemResource, error) {
	os := &osv1.OperatingSystemResource{}

	os.Sha256 = body.Sha256

	if body.ProfileName != nil {
		os.ProfileName = *body.ProfileName
	}

	if body.ProfileVersion != nil {
		os.ProfileVersion = *body.ProfileVersion
	}

	if body.Name != nil {
		os.Name = *body.Name
	}

	if body.Description != nil {
		os.Description = *body.Description
	}

	if body.Architecture != nil {
		os.Architecture = *body.Architecture
	}

	if body.RepoUrl != nil {
		os.ImageUrl = *body.RepoUrl
	}

	if body.ImageUrl != nil {
		os.ImageUrl = *body.ImageUrl
	}

	if body.ImageId != nil {
		os.ImageId = *body.ImageId
	}

	if body.InstalledPackages != nil {
		os.InstalledPackages = *body.InstalledPackages
	}

	if body.SecurityFeature != nil {
		os.SecurityFeature = openAPISecurityFeatureTogrpcSecurityFeature(body.SecurityFeature)
	}

	if body.OsType != nil {
		os.OsType = openAPIOperatingSystemTypeToGrpcOsType(*body.OsType)
	}

	if body.OsProvider != nil {
		os.OsProvider = openAPIOperatingSystemProviderToGrpcOsProvider(*body.OsProvider)
	}

	err := validator.ValidateMessage(os)
	if err != nil {
		log.InfraSec().InfraErr(err).Msg("could not validate inventory resource")
		return nil, errors.Wrap(err)
	}

	return os, nil
}

func grpcToOpenAPIOSResource(
	os *osv1.OperatingSystemResource,
	_ *inventory.GetResourceResponse_ResourceMetadata,
) *api.OperatingSystemResource {
	resID := os.GetResourceId()
	resName := os.GetName()
	description := os.GetDescription()
	arch := os.GetArchitecture()
	repoURL := os.GetImageUrl()
	imageID := os.GetImageId()
	sha256 := os.GetSha256()
	profileName := os.GetProfileName()
	installedPackages := os.GetInstalledPackages()
	securityFeature := grpcSecurityFeatureToOpenAPISecurityFeature(os.GetSecurityFeature())
	osType := grpcOsTypeToOpenAPIOSType(os.GetOsType())
	osProvider := grpcOsProviderToOpenAPIOsProvider(os.GetOsProvider())
	profileVersion := os.GetProfileVersion()
	platformBundle := os.GetPlatformBundle()

	obj := api.OperatingSystemResource{
		OsResourceID:      &resID,
		Name:              &resName,
		Description:       &description,
		Architecture:      &arch,
		RepoUrl:           &repoURL,
		ImageUrl:          &repoURL,
		ImageId:           &imageID,
		Sha256:            sha256,
		ProfileName:       &profileName,
		ProfileVersion:    &profileVersion,
		InstalledPackages: &installedPackages,
		SecurityFeature:   &securityFeature,
		OsType:            &osType,
		OsProvider:        &osProvider,
		ResourceId:        &resID,
		PlatformBundle:    &platformBundle,
		Timestamps:        GrpcToOpenAPITimestamps(os),
	}

	return &obj
}

// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"bytes"
	"regexp"
	"strings"

	"entgo.io/ent"
	"github.com/goccy/go-json"
	"github.com/iancoleman/strcase"
	"google.golang.org/grpc/codes"

	internal_ent "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/customconfigresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/endpointresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostgpuresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostnicresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostresource"
	hoststorage "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hoststorageresource"
	hostusb "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostusbresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/instanceresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/ipaddressresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/localaccountresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/netlinkresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/networksegment"
	oss "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/operatingsystemresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/osupdatepolicyresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/osupdaterunresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/ouresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/providerresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/regionresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/remoteaccessconfiguration"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/repeatedscheduleresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/singlescheduleresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/siteresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/telemetrygroupresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/telemetryprofile"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/tenant"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/workloadmember"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/workloadresource"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

const (
	PatchString = "PATCH"
	PutString   = "PUT"
)

// MetadataPatternKey representing the metadata pattern for key.
var MetadataPatternKey = regexp.MustCompile(
	"^$|^[a-z.]+/$|^[a-z.]+/[a-z0-9][a-z0-9-_.]*[a-z0-9]$|^[a-z.]+/[a-z0-9]$|^[a-z]$|^[a-z0-9][a-z0-9-_.]*[a-z0-9]$")

// MetadataPatternValue representing the metadata pattern for value.
var MetadataPatternValue = regexp.MustCompile("^$|^[a-z0-9]$|^[a-z0-9][a-z0-9+._-]*[a-z0-9]$")

// Metadata struct representing the JSON metadata.
type Metadata struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// BuildResourceMeta builds the ResourceMetadata from the given map of physical and logical metadata.
// errors are not returned to the caller as they are skipped as of today.
func BuildResourceMeta(phy, logi map[string]string) *inv_v1.GetResourceResponse_ResourceMetadata {
	sPhyMeta := ""
	sLogiMeta := ""
	err := error(nil)
	if len(phy) != 0 {
		sPhyMeta, err = metaMapToJSONString(phy)
		if err != nil {
			return nil
		}
	}
	if len(logi) != 0 {
		sLogiMeta, err = metaMapToJSONString(logi)
		if err != nil {
			return nil
		}
	}
	zlog.Debug().Msgf("PHY rendered String: %s", sPhyMeta)
	zlog.Debug().Msgf("LOGI rendered String: %s", sLogiMeta)
	return &inv_v1.GetResourceResponse_ResourceMetadata{
		PhyMetadata:  sPhyMeta,
		LogiMetadata: sLogiMeta,
	}
}

// ParseMetadata parses the given metadata, we expect a JSON encoded metadata that follows the Metadata struct definition.
func ParseMetadata(metadata string) (map[string]string, error) {
	if metadata == "" {
		return make(map[string]string), nil
	}
	var cMeta []Metadata
	err := json.Unmarshal([]byte(metadata), &cMeta)
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msgf("Error while un-marshaling the metadata")
		return nil, errors.Wrap(err)
	}
	var metaMap map[string]string
	metaMap, err = MetadataToMetaMap(cMeta)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return metaMap, nil
}

//nolint:cyclop // calculated cyclomatic complexity for func is 11, max is 10
func validateKeyValue(meta []Metadata) error {
	maxMetaKeyNameLen, maxMetaValueLen, maxMetaKeyPrefixLen := 63, 63, 253

	for _, rmetadata := range meta {
		if rmetadata.Key != "" {
			if !MetadataPatternKey.MatchString(rmetadata.Key) {
				return errors.Errorfc(codes.InvalidArgument, "Invalid metadata key")
			}
			// meta data key pattern <prefix>/<name>
			// max prefix len: 253  max name len:63
			if strings.Contains(rmetadata.Key, `/`) {
				prefixname := strings.Split(rmetadata.Key, "/")
				if len(prefixname[0]) > maxMetaKeyPrefixLen ||
					len(prefixname[1]) > maxMetaKeyNameLen {
					return errors.Errorfc(codes.InvalidArgument, "Invalid length of metadata key")
				}
			} else if len(rmetadata.Key) > maxMetaKeyNameLen { // meta data key pattern with name
				return errors.Errorfc(codes.InvalidArgument, "Invalid length of metadata key")
			}
		}
		if rmetadata.Value != "" {
			if len(rmetadata.Value) > maxMetaValueLen {
				return errors.Errorfc(codes.InvalidArgument, "Invalid length of metadata value")
			}
			if !MetadataPatternValue.MatchString(rmetadata.Value) {
				return errors.Errorfc(codes.InvalidArgument, "Invalid metadata value")
			}
		}
	}
	return nil
}

// ValidateMetadata verifies that the given JSON encoded metadata contains the required fields, otherwise returns an error.
func ValidateMetadata(metadata string) (string, error) {
	if metadata == "" {
		return "", nil
	}
	var tmp []Metadata
	err := json.Unmarshal([]byte(metadata), &tmp)
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msgf("Error while un-marshaling the metadata")
		return "", errors.Wrap(err)
	}

	// validate if the metadata contains duplicate keys
	_, err = MetadataToMetaMap(tmp)
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msgf("Error while validating the metadata")
		return "", err
	}

	// validate the actual key and value fields in metadata
	metaerr := validateKeyValue(tmp)
	if metaerr != nil {
		zlog.InfraSec().InfraErr(metaerr).Msgf("Error while validating the metadata")
		return "", errors.Wrap(metaerr)
	}
	// Marshal the metadata, so what we store is always the same, not random spaces in between.
	val, err := json.Marshal(tmp)
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msgf("Error while re-marshaling the metadata")
		return "", errors.Wrap(err)
	}
	// TODO: How can we validate that actual key and value are present?
	return string(val), nil
}

// MetadataToMetaMap transforms the given slide of Metadata into a map of string {Metadata.key : Metadata.value}.
// Returns an error in case of duplicated keys exist in the metadata.
func MetadataToMetaMap(metadata []Metadata) (map[string]string, error) {
	metaMap := make(map[string]string, len(metadata))
	for _, m := range metadata {
		if _, hasKey := metaMap[m.Key]; hasKey {
			return nil, errors.Errorfc(codes.InvalidArgument, "Duplicate metadata keys are forbidden")
		}
		metaMap[m.Key] = m.Value
	}
	return metaMap, nil
}

// mapsDifference calculates the difference between the two maps, removing elements in metaMap2 from metaMap1.
func mapsDifference(metaMap1, metaMap2 map[string]string) {
	for k := range metaMap2 {
		delete(metaMap1, k)
	}
}

// metaMapToJSONString converts a map of string into a JSON encoding of type Metadata.
func metaMapToJSONString(metaMap map[string]string) (string, error) {
	// Pre-allocate the slice
	meta := make([]Metadata, len(metaMap))
	i := 0
	for k, v := range metaMap {
		meta[i] = Metadata{
			Key:   k,
			Value: v,
		}
		i++
	}
	metaString, err := json.Marshal(meta)
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msgf("Error while marshaling the metadata")
		return "", errors.Wrap(err)
	}
	return string(metaString), nil
}

// EmptyEnumStateMap enum state mapping.
func EmptyEnumStateMap(fname string, _ int32) (ent.Value, error) {
	zlog.InfraSec().InfraError("unknown Enum field %s", fname).Msg("")
	return nil, errors.Errorfc(codes.InvalidArgument, "unknown Enum field %s", fname)
}

func getOffsetAndLimit(filter *inv_v1.ResourceFilter) (offset, limit int, err error) {
	offset, err = util.Uint32ToInt(filter.Offset)
	if err != nil {
		return 0, 0, err
	}
	limit, err = util.Uint32ToInt(filter.Limit)
	if err != nil {
		return 0, 0, err
	}
	return offset, limit, err
}

type OrderOption interface {
	endpointresource.OrderOption |
		hostgpuresource.OrderOption |
		hostnicresource.OrderOption |
		hostresource.OrderOption |
		hoststorage.OrderOption |
		hostusb.OrderOption |
		instanceresource.OrderOption |
		ipaddressresource.OrderOption |
		netlinkresource.OrderOption |
		networksegment.OrderOption |
		oss.OrderOption |
		ouresource.OrderOption |
		providerresource.OrderOption |
		regionresource.OrderOption |
		repeatedscheduleresource.OrderOption |
		singlescheduleresource.OrderOption |
		siteresource.OrderOption |
		workloadresource.OrderOption |
		workloadmember.OrderOption |
		telemetryprofile.OrderOption |
		telemetrygroupresource.OrderOption |
		remoteaccessconfiguration.OrderOption |
		tenant.OrderOption |
		localaccountresource.OrderOption |
		osupdatepolicyresource.OrderOption |
		customconfigresource.OrderOption |
		osupdaterunresource.OrderOption
}

// GetOrderByOptions takes an AIP-132 compliant orderBy string and returns the
// corresponding ent OrderOption. columnValidator is used to ensure only valid
// fields are selected. If no order is chosen (empty string), this returns a
// selector sorting by resource ID in ascending order.
func GetOrderByOptions[T OrderOption](orderBy string, columnValidator func(string) bool) (opts []T, err error) {
	if orderBy == "" {
		opts = append(opts, internal_ent.Asc("resource_id"))
		return opts, nil
	}
	for _, p := range strings.Split(orderBy, ",") {
		p = strings.Trim(p, " ")
		op := internal_ent.Asc
		if strings.HasSuffix(p, " desc") {
			p = strings.TrimSuffix(p, " desc")
			op = internal_ent.Desc
		} else if strings.HasSuffix(p, " asc") {
			p = strings.TrimSuffix(p, " asc")
			op = internal_ent.Asc
		}
		if p == "" {
			return nil, errors.Errorfc(codes.InvalidArgument, "empty `order_by` field")
		}
		p = strcase.ToSnake(p)
		// We have some fields that require special treatment after snake casing them.
		if p == "sha_256" {
			p = "sha256"
		}
		if !columnValidator(p) {
			return nil, errors.Errorfc(codes.InvalidArgument, "unknown column `%v`", p)
		}
		opts = append(opts, op(p))
	}
	return opts, nil
}

// ValidateOSMetadata verifies that the given JSON encoded metadata contains the required fields, otherwise returns an error.
func ValidateOSMetadata(metadata string) (string, error) {
	if metadata == "" {
		return "", nil
	}

	if err := validateFlatJSONNoDuplicates([]byte(metadata)); err != nil {
		return "", err
	}
	var tmp map[string]string
	err := json.Unmarshal([]byte(metadata), &tmp)
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msgf("Error while un-marshaling the OS metadata")
		return "", errors.Wrap(err)
	}

	// From OSMetadata to list of Metadata for easier validation
	metaList := make([]Metadata, 0, len(tmp))
	for k, v := range tmp {
		metaList = append(metaList, Metadata{
			Key:   k,
			Value: v,
		})
	}

	// validate the actual key and value fields in metadata
	metaerr := validateKeyValue(metaList)
	if metaerr != nil {
		zlog.InfraSec().InfraErr(metaerr).Msgf("Error while validating the OS metadata")
		return "", errors.Wrap(metaerr)
	}
	// Marshal the metadata, so what we store is always the same, not random spaces in between.
	val, err := json.Marshal(tmp)
	if err != nil {
		zlog.InfraSec().InfraErr(err).Msgf("Error while re-marshaling the OS metadata")
		return "", errors.Wrap(err)
	}
	return string(val), nil
}

func validateFlatJSONNoDuplicates(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	// Expecting top-level object
	tok, err := dec.Token()
	if err != nil {
		return errors.Errorfc(codes.InvalidArgument, "invalid JSON: %v", err)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return errors.Errorfc(codes.InvalidArgument, "JSON must start with an object")
	}

	keys := make(map[string]struct{})
	for dec.More() {
		keyToken, err := dec.Token()
		if err != nil {
			return errors.Errorfc(codes.InvalidArgument, "error reading key: %v", err)
		}
		key, ok := keyToken.(string)
		if !ok {
			return errors.Errorfc(codes.InvalidArgument, "expected string for key, got %T", keyToken)
		}
		if _, exists := keys[key]; exists {
			return errors.Errorfc(codes.InvalidArgument, "duplicate key found: %s", key)
		}
		keys[key] = struct{}{}

		// Skip value (we don't care about contents)
		if err := skipValue(dec); err != nil {
			return errors.Errorfc(codes.InvalidArgument, "error reading value for key %s: %v", key, err)
		}
	}

	// Consume closing '}'
	if _, err := dec.Token(); err != nil {
		return errors.Errorfc(codes.InvalidArgument, "missing closing object brace: %v", err)
	}

	return nil
}

func skipValue(dec *json.Decoder) error {
	// Just decode the next value into empty interface to move the cursor forward
	var v interface{}
	return dec.Decode(&v)
}

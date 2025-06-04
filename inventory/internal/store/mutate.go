// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store

// mutate.go
// turns gRPC values into an ent mutation (used with Create/Update)

import (
	"fmt"
	compute_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	location_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	"golang.org/x/exp/slices"
	"strings"

	entpb "entgo.io/contrib/entproto/cmd/protoc-gen-ent/options/ent"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

func clearEntMutate(mut ent.Mutation, fieldlist []string) error {
	// clears the given fields in the mutation object
	// fieldlist must be already validated via the FieldMask.isValid(...)
	for _, fieldname := range fieldlist {
		err := mut.ClearField(fieldname)
		if err != nil {
			err := mut.ClearEdge(fieldname)
			if err != nil {
				zlog.InfraSec().InfraErr(err).Msgf("Unable to clear ent mutate %s", mut.Type())
				return errors.Wrap(err)
			}
		}
	}
	return nil
}

func buildEntMutate(
	in proto.Message,
	mut ent.Mutation,
	enummap func(string, int32) (ent.Value, error),
	fieldmask []string,
) error {
	// builds a mutation object (changes to fields)
	// enums are handled by the enummap function, defined for every resource kind
	err := error(nil)
	inpr := in.ProtoReflect()

	// will only iterate over fields that are set in the Protobuf
	inpr.Range(func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		if len(fieldmask) != 0 { // Fieldmask is empty on Create, where immutability is not applicable.
			var markedImmutable bool
			markedImmutable, err = fieldIsEntImmutable(fd)
			if err != nil {
				zlog.InfraSec().InfraErr(err).Msgf("Error getting 'immutable' option on field %s, value: %v", fd.TextName(), val)
				return false
			}
			if markedImmutable {
				err = errors.Errorfc(codes.InvalidArgument, "field %s is immutable", fd.TextName())
				return false
			}
		}
		err = mutateFromProtoField(fd, val, mut, enummap)
		if err != nil {
			zlog.InfraSec().InfraErr(err).Msgf("Error setting field %s, value: %v", fd.TextName(), val)
			return false
		}
		return true
	})
	if err != nil {
		return err
	}

	// Eventually clear fields and edges that have not been already set
	if fieldmask != nil {
		fieldmaskmap := make(map[string]bool)
		for _, field := range fieldmask {
			fieldmaskmap[field] = true
		}
		// Find fields to be cleared
		for _, mutatedfield := range mut.Fields() {
			delete(fieldmaskmap, mutatedfield)
		}
		for _, mutatedEdge := range mut.AddedEdges() {
			delete(fieldmaskmap, mutatedEdge)
		}
		err = clearEntMutate(mut, maps.Keys(fieldmaskmap))
	}

	return err
}

func mutateFromProtoField(
	fd protoreflect.FieldDescriptor, val protoreflect.Value, mut ent.Mutation, enummap func(string, int32) (ent.Value, error),
) error {
	fname := fd.TextName()
	err := error(nil)

	switch fkind := fd.Kind(); fkind {
	case protoreflect.StringKind:
		err = handleStringKind(fd, val, mut)

	case protoreflect.Uint32Kind:
		err = handleUint32Kind(fd, val, mut)

	case protoreflect.Uint64Kind:
		err = errors.Wrap(mut.SetField(fname, val.Uint()))

	case protoreflect.Int32Kind:
		err = handleInt32Kind(fd, val, mut)

	case protoreflect.Int64Kind:
		err = errors.Wrap(mut.SetField(fname, val.Int()))

	case protoreflect.BoolKind:
		err = errors.Wrap(mut.SetField(fname, val.Bool()))

	case protoreflect.MessageKind:
		// Embedded fields represent edges that are handled by the resource
		// specific calling code.
		err = nil

	case protoreflect.EnumKind:
		err = handleEnumKind(fd, val, mut, enummap)

	default: // if an unsupported field.
		zlog.InfraSec().InfraError("Unsupported field: %s, of kind: %s (%d), with value %s",
			fname, fkind, fkind, val,
		).Msg("")
		err = errors.Errorfc(codes.InvalidArgument, "Unsupported field: %s, of kind: %s (%d), with value %s",
			fname, fkind, fkind, val,
		)
	}
	return err
}

func handleInt32Kind(fd protoreflect.FieldDescriptor, val protoreflect.Value, mut ent.Mutation) error {
	v, err := util.Int64ToInt32(val.Int()) // val.Int() returns int64
	if err != nil {
		return err
	}
	return errors.Wrap(mut.SetField(fd.TextName(), v))
}

func handleUint32Kind(fd protoreflect.FieldDescriptor, val protoreflect.Value, mut ent.Mutation) error {
	v, err := util.Uint64ToUint32(val.Uint()) // val.Uint() returns uint64
	if err != nil {
		return err
	}
	return errors.Wrap(mut.SetField(fd.TextName(), v))
}

func handleStringKind(fd protoreflect.FieldDescriptor, val protoreflect.Value, mut ent.Mutation) error {
	fname := fd.TextName()
	fullName := string(fd.FullName())
	var stringValue string
	err := error(nil)
	if fd.IsList() { // handle lists of strings by combining them with a pipe delimiter
		strlist := val.List()
		strslice := []string{}
		for i := 0; i < strlist.Len(); i++ {
			strslice = append(strslice, strlist.Get(i).String())
		}
		return errors.Wrap(mut.SetField(fname, strings.Join(strslice, "|")))
	}

	// TODO: add further validation for JSON fields.
	resourceWithMetadataToValidate := []string{
		fmt.Sprintf("%s.%s", getProtoFullName(&compute_v1.HostResource{}), compute_v1.HostResourceFieldMetadata),
		fmt.Sprintf("%s.%s", getProtoFullName(&compute_v1.WorkloadResource{}), compute_v1.WorkloadResourceFieldMetadata),
		fmt.Sprintf("%s.%s", getProtoFullName(&location_v1.SiteResource{}), location_v1.SiteResourceFieldMetadata),
		fmt.Sprintf("%s.%s", getProtoFullName(&location_v1.RegionResource{}), location_v1.RegionResourceFieldMetadata),
	}
	if slices.Contains(resourceWithMetadataToValidate, fullName) {
		if stringValue, err = ValidateMetadata(val.String()); err != nil {
			return err
		}
	} else {
		stringValue = val.String()
	}

	return errors.Wrap(mut.SetField(fname, stringValue))
}

func getProtoFullName(msg proto.Message) string {
	return string(msg.ProtoReflect().Descriptor().FullName())
}

// fieldIsWhitelistedImmutable contains a list of proto fields that should
// be immutable, but for which components currently issues Updates. To
// reduce breakage, we do not reject such requests until the issue is
// addressed there.
func fieldIsWhitelistedImmutable(_ protoreflect.FieldDescriptor) bool {
	// Using the FullName prevents false positives from fields with
	// the same name but in different resources.
	// switch fd.FullName() {
	// // case "os.v1.OperatingSystemResource.security_feature":
	// //	return true
	// default:
	// 	return false
	// }
	return false
}

func fieldIsEntImmutable(fd protoreflect.FieldDescriptor) (bool, error) {
	if fieldIsWhitelistedImmutable(fd) {
		zlog.Warn().Msgf("Whitelisted write request to immutable field '%v'."+
			" Consider fixing the request in the offending application.", fd.FullName())
		return false, nil
	}
	// Special case for "updated_at" field, this field is immutable from API, but should be mutable within inventory.
	if fd.TextName() == "updated_at" {
		return true, nil
	}
	opts, ok := fd.Options().(*descriptorpb.FieldOptions)
	if !ok {
		return false, errors.Errorfc(codes.Internal, "can't get options on field descriptor")
	}
	if !proto.HasExtension(opts, entpb.E_Field) {
		return false, nil
	}
	field, ok := proto.GetExtension(opts, entpb.E_Field).(*entpb.Field)
	if !ok {
		return false, errors.Errorfc(codes.Internal, "unexpected option extension on field descriptor")
	}
	return field.GetImmutable(), nil
}

func handleEnumKind(
	fd protoreflect.FieldDescriptor, val protoreflect.Value, mut ent.Mutation, enummap func(string, int32) (ent.Value, error),
) error {
	err := error(nil)
	fname := fd.TextName()

	enumval, enumerr := enummap(fname, int32(val.Enum()))
	if enumerr == nil { // only set if enummap returns valid results
		err = errors.Wrap(mut.SetField(fname, enumval))
	}
	return err
}

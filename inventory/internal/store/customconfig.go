// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"context"

	"google.golang.org/grpc/codes"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent"
	instances "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/instanceresource"
	customconfigs "github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/customconfigresource"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	cl "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

var customConfigResourceCreationValidators = []resourceValidator[*computev1.CustomConfigResource]{
	protoValidator[*computev1.CustomConfigResource],
	doNotAcceptResourceID[*computev1.CustomConfigResource],
}

func (is *InvStore) CreateCustomConfig(ctx context.Context, in *computev1.CustomConfigResource) (*inv_v1.Resource, error) {
	if err := validate(in, customConfigResourceCreationValidators...); err != nil {
		return nil, err
	}

	res, err := ExecuteInTxAndReturnSingle[inv_v1.Resource](is)(ctx, customConfigResourceCreator(in))
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("CustomConfig Created: %s, %s", res.GetCustomConfig().GetResourceId(), res)
	return res, nil
}


func customConfigResourceCreator(in *computev1.CustomConfigResource) func(context.Context, *ent.Tx) (
	*inv_v1.Resource, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inv_v1.Resource, error) {
		id := util.NewInvID(inv_v1.ResourceKind_RESOURCE_KIND_CUSTOMCONFIG)
		zlog.Debug().Msgf("CustomConfig: %s", id)

		newEntity := tx.CustomConfigResource.Create()
		mut := newEntity.Mutation()

		if err := buildEntMutate(in, mut, EmptyEnumStateMap, nil); err != nil {
			return nil, err
		}

		if err := mut.SetField(customconfigs.FieldResourceID, id); err != nil {
			return nil, errors.Wrap(err)
		}

		_, err := newEntity.Save(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		res, err := getCustomConfigQuery(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		return util.WrapResource(entCustomConfigResourceToProtoCustomConfigResource(res))
	}
}

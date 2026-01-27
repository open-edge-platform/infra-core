// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package invclient

import (
	"context"

	"github.com/mitchellh/hashstructure/v2"

	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
)

type resourceLister interface {
	ListAll(context.Context, *inv_v1.ResourceFilter) ([]*inv_v1.Resource, error)
}

func NewInventoryClientCache(ic resourceLister) resourceLister {
	return &inventoryClientNaiveCache{
		resourceLister:   ic,
		cacheByOperation: cacheByOperation{},
	}
}

type respByQueryHash map[uint64]interface{}

type cacheByOperation map[string]respByQueryHash

func (c cacheByOperation) get(op string, hash uint64) (interface{}, bool) {
	opCache, opCacheOk := c[op]
	if !opCacheOk {
		return nil, false
	}
	rsp, ok := opCache[hash]
	if !ok {
		return nil, false
	}

	return rsp, true
}

// inventoryClientNaiveCache - this is naive implementation of inventory client cache.
// It caches results of ListAll operation, and ignores all other operations exposed through client.TenantAwareInventoryClient
// It never expires and never updates cached responses, so intentionally it shall be used as short-lived cache.
type inventoryClientNaiveCache struct {
	resourceLister
	cacheByOperation
}

func (i *inventoryClientNaiveCache) ListAll(ctx context.Context, filter *inv_v1.ResourceFilter) ([]*inv_v1.Resource, error) {
	hash, herr := hash(filter)
	if herr != nil {
		return nil, herr
	}

	cached, ok := i.get("ListAll", hash)
	if ok {
		resource, ok := cached.([]*inv_v1.Resource)
		if !ok {
			return nil, errors.Errorf("unexpected type for []*inv_v1.Resource: %T", cached)
		}
		return resource, nil
	}

	rsp, err := i.resourceLister.ListAll(ctx, filter)
	if err != nil {
		return nil, err
	}
	if _, ok := i.cacheByOperation["ListAll"]; !ok {
		i.cacheByOperation["ListAll"] = map[uint64]interface{}{}
	}
	i.cacheByOperation["ListAll"][hash] = rsp
	return rsp, err
}

func hash(filter interface{}) (uint64, error) {
	return hashstructure.Hash(filter, hashstructure.FormatV2, nil)
}

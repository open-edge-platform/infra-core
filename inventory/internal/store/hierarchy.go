// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"google.golang.org/grpc/codes"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/ouresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/regionresource"
	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/siteresource"
	inventoryv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

type treeNode struct {
	ID             int
	name           string
	nodeType       inventoryv1.ResourceKind
	regionParentID *int
	ouParentID     *int
	siteParentID   *int
	metadata       string
	depth          int
}

type sqlSortOrder string

// NEVER change these constants, they drive ordering of results from SQL queries.
const (
	ascending     sqlSortOrder = "" // By default ordering is ascending
	descending    sqlSortOrder = "DESC"
	whereTenantID              = "WHERE tenant_id=$1"
	tenantIDAnd                = "tenant_id=$1 AND"
)

// getSQLPlaceholdersAndArgs returns a generic slice of objects ([]interface{}) and a string of placeholders one for each
// of the arguments in the provided slice ("$1,$2"). The offset, if provided, is used to offset the initial number in the
// placeholder string.
// The result is used to parametrize the SQL query.
func getSQLPlaceholdersAndArgs[T any](args []T, offsets ...int) ([]interface{}, string) {
	offset := 0
	if len(offsets) > 0 {
		offset = offsets[0]
	}
	placeholders := make([]string, len(args))
	rArgs := make([]interface{}, len(args))
	for i, id := range args {
		rArgs[i] = id
		placeholders[i] = fmt.Sprintf("$%d", offset+i+1)
	}
	return rArgs, strings.Join(placeholders, ",")
}

func logAndSanitizeErrorRawSQLf(err error, msg string) error {
	zlog.InfraSec().Err(err).Msg(msg)
	return errors.Errorfc(codes.Internal, "%s", msg)
}

func getHierarchyRegions(ctx context.Context, client *ent.Client, order sqlSortOrder, tenantID *string, regionIDs []int) (
	[]*treeNode,
	error,
) {
	var args []interface{}
	offset := 0
	// Optional tenantID matching clause. Needs to be optional because this function is used also in case of filters and lists.
	tenantIDAndClause := ""
	whereTenantIDClause := ""
	if tenantID != nil {
		// The first arg is always tenantID.
		offset = 1
		args = append(args, *tenantID)
		tenantIDAndClause = tenantIDAnd
		whereTenantIDClause = whereTenantID
	}
	// TODO: how much expensive is this clause? If we have to plot all trees, it would be more efficient maybe to traverse
	//  trees root to leaves.
	// By default start from any of the leaf regions (a leaf region is a region that has no parent that links to it).
	whereClause := "WHERE " + tenantIDAndClause + " ID NOT IN (" +
		"SELECT DISTINCT region_resource_parent_region " +
		"FROM region_resources " +
		"WHERE " + tenantIDAndClause + "region_resource_parent_region IS NOT NULL)"
	if len(regionIDs) > 0 {
		otherArgs, placeholders := getSQLPlaceholdersAndArgs(regionIDs, offset)
		args = append(args, otherArgs...)
		whereClause = "WHERE " + tenantIDAndClause + " ID IN (" + placeholders + ")"
	}
	// The query starts from the target leaf edges (the given region IDs) or from any leaf regions.
	// From there it builds the whole hierarchy recursively increasing the depth one level by one level.
	// We also collect metadata while traversing the trees.
	// Duplicates (if any) are removed, and the result is ordered by depth.
	// The output of the query is driven by the requirement of the executeHierarchyQueryAndParse function
	// and will contain one per line, root to leaf (ordered desc by depth) all the nodes of the required subtrees.
	query := `
	WITH RECURSIVE region_hierarchy AS (
		SELECT ID AS id, region_resource_parent_region AS parent_id, 0 AS depth, metadata AS meta
		FROM region_resources
		` + whereClause + `
		UNION ALL
		SELECT r.ID AS id, r.region_resource_parent_region AS parent_id, rh.depth+1 AS depth, metadata AS meta
		FROM region_resources AS r JOIN region_hierarchy AS rh ON rh.parent_id=r.ID
		` + whereTenantIDClause + `
	)
	SELECT DISTINCT id, NULL AS name, parent_id AS reg_parent_id, NULL AS ou_parent_id, NULL AS site_parent_id, meta, depth, ` +
		`'` + string(util.ResourcePrefixRegion) + `' AS type
	FROM region_hierarchy
	ORDER BY depth ` + string(order) + `, id`
	return executeHierarchyQueryAndParse(ctx, client, query, args)
}

func getHierarchyOus(ctx context.Context, client *ent.Client, order sqlSortOrder, tenantID *string, ouIDs []int) (
	[]*treeNode,
	error,
) {
	var args []interface{}
	offset := 0
	// Optional tenantID matching clause. Needs to be optional because this function is used also in case of filters and lists.
	tenantIDAndClause := ""
	whereTenantIDClause := ""
	if tenantID != nil {
		// The first arg is always tenantID.
		offset = 1
		args = append(args, *tenantID)
		tenantIDAndClause = tenantIDAnd
		whereTenantIDClause = whereTenantID
	}
	// TODO: how much expensive is this clause? If we have to plot all trees, it would be more efficient maybe to traverse
	//  trees root to leaves.
	// By default start from any of the leaf OUs (a leaf OU is a OU that has no parent that links to it).
	whereClause := "WHERE " + tenantIDAndClause + " ID NOT IN (" +
		"SELECT DISTINCT ou_resource_parent_ou FROM ou_resources " +
		"WHERE " + tenantIDAndClause + "ou_resource_parent_ou IS NOT NULL)"
	if len(ouIDs) > 0 {
		otherArgs, placeholders := getSQLPlaceholdersAndArgs(ouIDs, offset)
		args = append(args, otherArgs...)
		whereClause = "WHERE " + tenantIDAndClause + " ID IN (" + placeholders + ")"
	}
	// The query starts from the target leaf edges (the given OU IDs) or from any leaf OUs.
	// From there it builds the whole hierarchy recursively increasing the depth one level by one level.
	// We also collect metadata while traversing the trees.
	// Duplicates (if any) are removed, and the result is ordered by depth.
	// The output of the query is driven by the requirement of the executeHierarchyQueryAndParse function
	// and will contain one per line, root to leaf (ordered desc by depth) all the nodes of the required subtrees.
	query := `
	WITH RECURSIVE ou_hierarchy AS (
		SELECT ID AS id, ou_resource_parent_ou AS parent_id, 0 AS depth, metadata AS meta
		FROM ou_resources
		` + whereClause + `
		UNION ALL
		SELECT o.ID AS id, o.ou_resource_parent_ou AS parent_id, oh.depth+1 AS depth, o.metadata AS meta
		FROM ou_resources AS o JOIN ou_hierarchy AS oh ON oh.parent_id=o.ID
		` + whereTenantIDClause + `
	)
	SELECT DISTINCT id, NULL AS name, NULL AS reg_parent_id, parent_id AS ou_parent_id, NULL AS site_parent_id, meta, depth, ` +
		`'` + string(util.ResourcePrefixOu) + `' AS type
	FROM ou_hierarchy
	ORDER BY depth ` + string(order) + `, id`
	return executeHierarchyQueryAndParse(ctx, client, query, args)
}

func getHierarchySites(ctx context.Context, client *ent.Client, order sqlSortOrder, tenantID *string, siteIDs []int) (
	[]*treeNode,
	error,
) {
	var args []interface{}
	offset := 0
	// Optional tenantID matching clause. Needs to be optional because this function is used also in case of filters and lists.
	tenantIDAndClause := ""
	whereTenantIDClause := ""
	if tenantID != nil {
		// The first arg is always tenantID.
		offset = 1
		args = append(args, *tenantID)
		tenantIDAndClause = tenantIDAnd
		whereTenantIDClause = whereTenantID
	}
	// TODO: how much expensive is this clause? We maybe be returning the whole tree!
	// By default start from any of the sites (all sites are leaf!)
	whereClause := ""
	if len(siteIDs) > 0 {
		otherArgs, placeholders := getSQLPlaceholdersAndArgs(siteIDs, offset)
		args = append(args, otherArgs...)
		whereClause = "WHERE " + tenantIDAndClause + " ID IN (" + placeholders + ")"
	}
	// Here we have two recursive queries, one to build the region hierarchy starting from sites, the other to
	// build the OU hierarchy starting from sites.
	// The first query, starts from the target leaf edges (the given site IDs) or from any site.
	// Sites are the first level of the hierarchy (always for this query).
	// From sites, it recursively builds the hierarchy of regions, increasing the depth one level by one level.
	// The same is done for the OU hierarchy, following the same logic.
	// We also collect metadata while traversing the trees.
	// Finally, the results from the two recursive queries is merged, duplicates (if any) are removed, and the result
	// is ordered by depth.
	// The output of the query is driven by the requirement of the executeHierarchyQueryAndParse function
	// and will contain one per line, root to leaf (ordered desc by depth) all the nodes of the required subtrees.
	query := `
	WITH RECURSIVE region_hierarchy AS (
		SELECT ID AS id, site_resource_region AS reg_parent_id, site_resource_ou AS ou_parent_id, ` +
		` 0 AS depth, metadata AS meta, '` + string(util.ResourcePrefixSite) + `' AS type
		FROM site_resources
		` + whereClause + `
		UNION ALL
		SELECT r.ID AS id, r.region_resource_parent_region AS reg_parent_id, NULL AS ou_parent_id, ` +
		` rh.depth+1 AS depth, r.metadata AS meta, '` + string(util.ResourcePrefixRegion) + `' AS type
		FROM region_resources AS r JOIN region_hierarchy AS rh ON rh.reg_parent_id=r.ID
		` + whereTenantIDClause + `
	),
	ou_hierarchy AS (
		SELECT ID AS id, site_resource_region AS reg_parent_id, site_resource_ou AS ou_parent_id, ` +
		` 0 AS depth, metadata AS meta, '` + string(util.ResourcePrefixSite) + `' AS type
		FROM site_resources
		` + whereClause + `
		UNION ALL
		SELECT o.ID AS id, NULL AS reg_parent_id, o.ou_resource_parent_ou AS ou_parent_id, ` +
		` oh.depth+1 AS depth, o.metadata AS meta, '` + string(util.ResourcePrefixOu) + `' AS type
		FROM ou_resources AS o JOIN ou_hierarchy AS oh ON oh.ou_parent_id=o.ID
		` + whereTenantIDClause + `
	)
	SELECT DISTINCT id, NULL AS name, reg_parent_id, ou_parent_id, NULL AS site_parent_id, meta, depth, type
	FROM
	(
		(SELECT id, reg_parent_id, ou_parent_id, meta, depth, type FROM region_hierarchy)
		UNION ALL
	 	(SELECT id, reg_parent_id, ou_parent_id, meta, depth, type FROM ou_hierarchy)
	) AS all_hierarchy
	ORDER BY depth ` + string(order) + `, type, id`
	return executeHierarchyQueryAndParse(ctx, client, query, args)
}

func getHierarchyHosts(ctx context.Context, client *ent.Client, order sqlSortOrder, tenantID *string, hostIDs []int) (
	[]*treeNode,
	error,
) {
	var args []interface{}
	offset := 0
	// optional tenantID matching clause. Needs to be optional because this function is used also in case of filters and lists.
	tenantIDAndClause := ""
	whereTenantIDClause := ""
	if tenantID != nil {
		// The first arg is always tenantID.
		offset = 1
		args = append(args, *tenantID)
		tenantIDAndClause = tenantIDAnd
		whereTenantIDClause = whereTenantID
	}
	// TODO: how much expensive is this clause? We maybe be returning the whole tree!
	// By default start from all the hosts (all hosts are leaf). For sites, instead we want to only take sites that
	// are actually linked to a host.
	whereHostClause := ""
	whereSiteClause := "WHERE ID IN (SELECT host_resource_site FROM host_resources " +
		"WHERE " + tenantIDAndClause + "host_resource_site IS NOT NULL)"
	if len(hostIDs) > 0 {
		otherArgs, placeholders := getSQLPlaceholdersAndArgs(hostIDs, offset)
		args = append(args, otherArgs...)
		whereHostClause = "WHERE " + tenantIDAndClause + " ID IN (" + placeholders + ")"
		// We only consider sites that are attached to one of the hosts we are filtering.
		whereSiteClause = "WHERE " + tenantIDAndClause + " ID IN (" +
			"SELECT DISTINCT host_resource_site FROM host_resources " + whereHostClause + ")"
	}
	// Here we have two recursive queries, one to build the region hierarchy the other to build the OU hierarchy
	// starting from sites where the required hosts sit (either the given one, or all hosts).
	// The first query, starts from the target leaf edges (site IDs).
	// Sites are the second level of the hierarchy for the scope of this query (the first level are hosts).
	// From sites, it recursively builds the hierarchy of regions, increasing the depth one level by one level.
	// The same is done for the OU hierarchy, following the same logic.
	// We also collect metadata while traversing the trees.
	// Finally, the results from the two recursive queries is merged with the result from a query of the hosts,
	// duplicates (if any) are removed, and the result is ordered by depth.
	// The output of the query is driven by the requirement of the executeHierarchyQueryAndParse function
	// and will contain one per line, root to leaf (ordered desc by depth) all the nodes of the required subtrees.
	query := `
	WITH RECURSIVE region_hierarchy AS (
		SELECT ID AS id, site_resource_region AS reg_parent_id, site_resource_ou AS ou_parent_id, ` +
		` 1 AS depth, metadata AS meta, '` +
		string(util.ResourcePrefixSite) + `' AS type
		FROM site_resources
		` + whereSiteClause + `
		UNION ALL
		SELECT r.ID AS id, r.region_resource_parent_region AS reg_parent_id, NULL AS ou_parent_id, rh.depth+1 AS depth, ` +
		` r.metadata AS meta, '` + string(util.ResourcePrefixRegion) + `' AS type
		FROM region_resources AS r JOIN region_hierarchy AS rh ON rh.reg_parent_id=r.ID
		` + whereTenantIDClause + `
	),
	ou_hierarchy AS (
		SELECT ID AS id, site_resource_region AS reg_parent_id, site_resource_ou AS ou_parent_id, ` +
		` 1 AS depth, metadata AS meta, '` + string(util.ResourcePrefixSite) + `' AS type
		FROM site_resources
		` + whereSiteClause + `
		UNION ALL
		SELECT o.ID AS id, NULL AS reg_parent_id, o.ou_resource_parent_ou AS ou_parent_id, ` +
		` oh.depth+1 AS depth, o.metadata AS meta, '` + string(util.ResourcePrefixOu) + `' AS type
		FROM ou_resources AS o JOIN ou_hierarchy AS oh ON oh.ou_parent_id=o.ID
		` + whereTenantIDClause + `
	)
	SELECT DISTINCT id, NULL AS name, reg_parent_id, ou_parent_id, site_parent_id, meta, depth, type
	FROM (
		(
			SELECT ID AS id, NULL AS reg_parent_id, NULL AS ou_parent_id, host_resource_site AS site_parent_id, ` +
		`metadata AS meta, 0 AS depth, '` + string(util.ResourcePrefixHost) + `' AS type
			FROM host_resources
			` + whereHostClause + `
		)
		UNION ALL
		(SELECT id, reg_parent_id, ou_parent_id, NULL AS site_parent_id, meta, depth, type FROM region_hierarchy)
		UNION ALL
	 	(SELECT id, reg_parent_id, ou_parent_id,  NULL AS site_parent_id, meta, depth, type FROM ou_hierarchy)
	) AS all_hierarchy
	ORDER BY depth ` + string(order) + `, type, id`
	return executeHierarchyQueryAndParse(ctx, client, query, args)
}

func filterHierarchies(ctx context.Context, client *ent.Client, order sqlSortOrder, tenantID string, resourceIDs []string) (
	[]*treeNode,
	error,
) {
	// TODO: how much expensive is this clause? We maybe be returning the whole tree!
	// Host and site clause require to filter on tenantID because we filter is applied to the Host and Site table directly.
	// By default, hosts and site are all leaves
	whereHostClause := whereTenantID
	whereSiteClause := whereTenantID
	// OU and region clause do not require filtering by tenantID because the reg_tab_depth and ou_tab_depth provide
	// results already filtered by the given tenant.
	// By default, start from any of the leaf regions (a leaf region is a region that has no parent that links to it).
	whereRegionClause := "WHERE ID NOT IN (" +
		"SELECT DISTINCT region_resource_parent_region " +
		"FROM region_resources WHERE region_resource_parent_region IS NOT NULL AND tenant_id=$1)"
	// By default, start from any of the leaf OUs (a leaf OU is a OU that has no parent that links to it).
	whereOuClause := "WHERE ID NOT IN (" +
		"SELECT DISTINCT ou_resource_parent_ou FROM ou_resources WHERE ou_resource_parent_ou IS NOT NULL AND tenant_id=$1)"
	var args []interface{}
	// The first arg is always tenantID.
	args = append(args, tenantID)
	if len(resourceIDs) > 0 {
		otherArgs, placeholders := getSQLPlaceholdersAndArgs(resourceIDs, 1)
		args = append(args, otherArgs...)
		resourceIDClause := " resource_id IN (" + placeholders + ")"
		// Host and site clause require to filter on tenantID because we filter on the Host and Site table directly.
		whereHostClause = "WHERE " + tenantIDAnd + resourceIDClause
		// Sites are all the one linked to one of the matched hosts OR sites that are actually matching the filter
		whereSiteClause = `WHERE ` + tenantIDAnd + ` (ID IN (SELECT DISTINCT host_resource_site FROM host_resources ` +
			whereHostClause + `) OR ` + resourceIDClause + `)`
		// OU and region clause do not require filtering by tenantID because the reg_tab_depth and ou_tab_depth provide
		// results already filtered by the given tenant.
		// Region are all the one linked to one of the matched sites OR sites that are actually matching the filter
		whereRegionClause = "WHERE " + resourceIDClause +
			" OR ID IN (SELECT DISTINCT site_resource_region FROM site_resources " + whereSiteClause + ")"
		whereOuClause = "WHERE " + resourceIDClause +
			" OR ID IN (SELECT DISTINCT site_resource_ou FROM site_resources " + whereSiteClause + ")"
	}
	depthSites := fmt.Sprintf("%d", util.MaxResourceNestingLevel)
	depthHosts := fmt.Sprintf("%d", util.MaxResourceNestingLevel+1)
	// TODO: how much expensive is this query? Need a pass in the query planner
	//  Ideas to optimize: calculate subtables first, and then re-use them across the recursive queries
	//  For example: gather all the Hosts, from hosts gather the sites, from sites gather the leaf region and leaf ous
	//  use the temp result to feed the recursive queries.
	// First we calculate the depth in the tree for every region and OUs.
	// Then we render the tree based on the provided filters, starting from region and OUs. Finally, we merge the result
	// with matching sites and hosts.
	query := `
	WITH RECURSIVE 
	--- intermediate table with all region with their depth in the region tree (depth=0 is the root of the region tree)
	reg_tab_depth AS (
		SELECT ID AS id, resource_id, name, region_resource_parent_region AS parent_id, 0 AS depth
		FROM region_resources ` +
		whereTenantID + ` AND region_resource_parent_region IS NULL
		UNION ALL
		SELECT r.ID AS id, r.resource_id AS resource_id, r.name AS name, r.region_resource_parent_region AS parent_id, 
		rh.depth+1 AS depth
		FROM region_resources AS r JOIN reg_tab_depth AS rh ON r.region_resource_parent_region=rh.id
		` + whereTenantID + `
	),
	--- intermediate table with all OUs with their depth in the OU tree (depth=0 is the root of the OU tree)
	ou_tab_depth AS (
		SELECT ID AS id, resource_id, name, ou_resource_parent_ou AS parent_id, 0 AS depth
		FROM ou_resources ` +
		whereTenantID + ` AND ou_resource_parent_ou IS NULL
		UNION ALL
		SELECT o.ID AS id, o.resource_id AS resource_id, o.name AS name, o.ou_resource_parent_ou AS parent_id, 
		oh.depth+1 AS depth
		FROM ou_resources AS o JOIN ou_tab_depth AS oh ON o.ou_resource_parent_ou=oh.id
		` + whereTenantID + `
	),
	region_hierarchy AS (
		SELECT id, name, parent_id, depth
		FROM reg_tab_depth
		` + whereRegionClause + `
		UNION ALL
		SELECT r.id AS id, r.name AS name, r.parent_id AS parent_id, r.depth AS depth
		FROM reg_tab_depth AS r JOIN region_hierarchy AS rh ON rh.parent_id=r.ID
	),
	ou_hierarchy AS (
		SELECT id, name, parent_id, depth
		FROM ou_tab_depth
		` + whereOuClause + `
		UNION ALL
		SELECT o.id AS id, o.name AS name, o.parent_id AS parent_id, o.depth AS depth
		FROM ou_tab_depth AS o JOIN ou_hierarchy AS oh ON oh.parent_id=o.ID
	)
	-- for this query we do not care about metadata
	SELECT DISTINCT id, name, reg_parent_id, ou_parent_id, site_parent_id, NULL AS meta, depth, type
	FROM (
		SELECT ID AS id, name, NULL AS reg_parent_id, NULL AS ou_parent_id, host_resource_site AS site_parent_id, ` +
		depthHosts + ` AS depth, '` + string(util.ResourcePrefixHost) + `' AS type
		FROM host_resources
		` + whereHostClause + `
		UNION ALL
		SELECT ID AS id, name, site_resource_region AS reg_parent_id, site_resource_ou AS ou_parent_id, ` +
		`NULL AS site_parent_id,` + depthSites + ` AS depth, '` + string(util.ResourcePrefixSite) + `' AS type
		FROM site_resources
		` + whereSiteClause + `
		UNION ALL
		SELECT id, name, parent_id AS reg_parent_id, NULL AS ou_parent_id, NULL AS site_parent_id, depth, '` +
		string(util.ResourcePrefixRegion) + `' AS type 
		FROM region_hierarchy
		UNION ALL
		SELECT id, name, NULL AS reg_parent_id, parent_id AS ou_parent_id, NULL AS site_parent_id, depth, '` +
		string(util.ResourcePrefixOu) + `' AS type 
		FROM ou_hierarchy
	) AS ah
	ORDER BY depth ` + string(order) + `, type, id`
	return executeHierarchyQueryAndParse(ctx, client, query, args)
}

// getResourceIDs returns the resource IDs given the PK ID of the host, site, region and OU.
func getResourceIDs(ctx context.Context, client *ent.Client, hostIDs, siteIDs, regionIDs, ouIDs []int, tenantID string) (
	map[inventoryv1.ResourceKind]map[int]string, error,
) {
	if len(hostIDs)+len(siteIDs)+len(regionIDs)+len(ouIDs) == 0 {
		return make(map[inventoryv1.ResourceKind]map[int]string), nil
	}
	// Pre-allocate the id to res ID map for all the expected resources kind
	idToResIDs := make(map[inventoryv1.ResourceKind]map[int]string)
	idToResIDs[inventoryv1.ResourceKind_RESOURCE_KIND_HOST] = make(map[int]string)
	idToResIDs[inventoryv1.ResourceKind_RESOURCE_KIND_SITE] = make(map[int]string)
	idToResIDs[inventoryv1.ResourceKind_RESOURCE_KIND_REGION] = make(map[int]string)
	idToResIDs[inventoryv1.ResourceKind_RESOURCE_KIND_OU] = make(map[int]string)

	query, args := buildGetResourceIDQuery(hostIDs, siteIDs, regionIDs, ouIDs, tenantID)
	rows, err := client.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, logAndSanitizeErrorRawSQLf(err, "error querying resource IDs")
	}
	for rows.Next() {
		var id int
		var resourceID string
		if err := rows.Scan(&resourceID, &id); err != nil {
			return nil, logAndSanitizeErrorRawSQLf(err, "error parsing results while querying resource IDs")
		}
		resKind, err := util.GetResourceKindFromResourceID(resourceID)
		if err != nil {
			zlog.InfraSec().Err(err).Msgf("this error should never happen")
			return nil, err
		}
		idMap, ok := idToResIDs[resKind]
		if !ok {
			idMap = make(map[int]string)
		}
		idMap[id] = resourceID
	}
	return idToResIDs, nil
}

func buildGetResourceIDQuery(hostIDs, siteIDs, regionIDs, ouIDs []int, tenantID string) (string, []interface{}) {
	// Base format for the query, where we are interested in getting the ID->resource_id mapping.
	queryFormat := "SELECT ID, resource_id FROM %s WHERE ID IN (%s) AND tenant_id=$1"
	queries := make([]string, 0)
	args := make([]interface{}, 0, 1+len(hostIDs)+len(siteIDs)+len(regionIDs)+len(ouIDs))
	args = append(args, tenantID)
	if len(hostIDs) > 0 {
		hArgs, hPlaceholders := getSQLPlaceholdersAndArgs(hostIDs, len(args))
		queries = append(queries, fmt.Sprintf(queryFormat, hostresource.Table, hPlaceholders))
		args = append(args, hArgs...)
	}
	if len(siteIDs) > 0 {
		sArgs, sPlaceholders := getSQLPlaceholdersAndArgs(siteIDs, len(args))
		queries = append(queries, fmt.Sprintf(queryFormat, siteresource.Table, sPlaceholders))
		args = append(args, sArgs...)
	}
	if len(regionIDs) > 0 {
		rArgs, rPlaceholders := getSQLPlaceholdersAndArgs(regionIDs, len(args))
		queries = append(queries, fmt.Sprintf(queryFormat, regionresource.Table, rPlaceholders))
		args = append(args, rArgs...)
	}
	if len(ouIDs) > 0 {
		oArgs, oPlaceholders := getSQLPlaceholdersAndArgs(ouIDs, len(args))
		queries = append(queries, fmt.Sprintf(queryFormat, ouresource.Table, oPlaceholders))
		args = append(args, oArgs...)
	}
	finalQuery := "SELECT DISTINCT resource_id, ID  FROM (" + strings.Join(queries, " UNION ALL ") + ") AS all_res"

	return finalQuery, args
}

// executeHierarchyQueryAndParse execute the given query with given args and parse the result.
// The query is expected to have placeholders for the given args.
// The result is expected to contain the following fields (with types) as column:
// ID (int64), regionParentID (int64), ouParentID (int64), siteParentID (int64),
// metadata (string), name (string), depth (int), nodeType (string).
// The function returns the representation of all the subtrees as a list of adjacent nodes (ordered according to the query).
func executeHierarchyQueryAndParse(ctx context.Context, client *ent.Client, query string, args []interface{}) (
	[]*treeNode, error,
) {
	rows, err := client.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, logAndSanitizeErrorRawSQLf(err, "error querying hierarchy tree")
	}
	tree := make([]*treeNode, 0)
	for rows.Next() {
		var depth int
		var id int
		var regParentID, ouParentID, siteParentID sql.NullInt64
		var nodeType string
		var metadata, name sql.NullString
		if err := rows.Scan(&id, &name, &regParentID, &ouParentID, &siteParentID, &metadata, &depth, &nodeType); err != nil {
			return nil, logAndSanitizeErrorRawSQLf(err, "error parsing results while rendering the hierarchy tree")
		}
		var parentRegionIDPtr *int
		if regParentID.Valid {
			regParentIDInt, err := util.Int64ToInt(regParentID.Int64)
			if err != nil {
				return nil, err
			}
			parentRegionIDPtr = &regParentIDInt
		}
		var parentOUIDPtr *int
		if ouParentID.Valid {
			ouParentIDInt, err := util.Int64ToInt(ouParentID.Int64)
			if err != nil {
				return nil, err
			}
			parentOUIDPtr = &ouParentIDInt
		}
		var parentSiteIDPtr *int
		if siteParentID.Valid {
			siteParentIDInt, err := util.Int64ToInt(siteParentID.Int64)
			if err != nil {
				return nil, err
			}
			parentSiteIDPtr = &siteParentIDInt
		}
		tree = append(tree, &treeNode{
			id,
			name.String,
			util.PrefixToResourceKind(util.ResourcePrefix(nodeType)),
			parentRegionIDPtr,
			parentOUIDPtr,
			parentSiteIDPtr,
			metadata.String,
			depth,
		})
	}
	return tree, nil
}

func (is *InvStore) GetTreeHierarchy(ctx context.Context, filter *inventoryv1.GetTreeHierarchyRequest) (
	*inventoryv1.GetTreeHierarchyResponse, error,
) {
	if err := validator.ValidateMessage(filter); err != nil {
		zlog.InfraSec().InfraErr(err).Send()
		return nil, errors.Wrap(err)
	}
	res, err := ExecuteInRoTxAndReturnSingle[inventoryv1.GetTreeHierarchyResponse](is)(ctx, getTreeHierarchyCreator(filter))
	if err != nil {
		zlog.InfraSec().InfraErr(err).Send()
		return nil, err
	}
	if err := validator.ValidateMessage(res); err != nil {
		zlog.InfraSec().InfraErr(err).Send()
		return nil, errors.Wrap(err)
	}
	if err := ValidateTreeHierarchyResponse(res); err != nil {
		zlog.InfraSec().InfraErr(err).Send()
		return nil, errors.Wrap(err)
	}
	return res, nil
}

func getTreeHierarchyCreator(filter *inventoryv1.GetTreeHierarchyRequest) func(context.Context, *ent.Tx) (
	*inventoryv1.GetTreeHierarchyResponse, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inventoryv1.GetTreeHierarchyResponse, error) {
		resourceIDs := filter.GetFilter()
		order := ascending
		if filter.GetDescending() {
			order = descending
		}

		tree, err := filterHierarchies(ctx, tx.Client(), order, filter.GetTenantId(), resourceIDs)
		if err != nil {
			return nil, err
		}
		kindToIDs := make(map[inventoryv1.ResourceKind][]int)
		for _, node := range tree {
			kindToIDs[node.nodeType] = append(kindToIDs[node.nodeType], node.ID)
		}
		idToResIDMap, err := getResourceIDs(ctx, tx.Client(),
			kindToIDs[inventoryv1.ResourceKind_RESOURCE_KIND_HOST],
			kindToIDs[inventoryv1.ResourceKind_RESOURCE_KIND_SITE],
			kindToIDs[inventoryv1.ResourceKind_RESOURCE_KIND_REGION],
			kindToIDs[inventoryv1.ResourceKind_RESOURCE_KIND_OU],
			filter.GetTenantId(),
		)
		if err != nil {
			return nil, err
		}
		responseNodes := collections.MapSlice[*treeNode, *inventoryv1.GetTreeHierarchyResponse_TreeNode](
			tree,
			func(t *treeNode) *inventoryv1.GetTreeHierarchyResponse_TreeNode {
				depthInt, err := util.IntToInt32(t.depth)
				if err != nil {
					zlog.InfraSec().Err(err).Msgf("while converting tree nodes, continuing, returning depth = -1")
					depthInt = -1
				}
				resID := idToResIDMap[t.nodeType][t.ID]
				node := inventoryv1.GetTreeHierarchyResponse_TreeNode{
					Name:  t.name,
					Depth: depthInt,
					CurrentNode: &inventoryv1.GetTreeHierarchyResponse_Node{
						ResourceId:   resID,
						ResourceKind: t.nodeType,
					},
				}
				applyFuncToParents(t, func(parentID *int, kind inventoryv1.ResourceKind) {
					resID := idToResIDMap[kind][*parentID]
					node.ParentNodes = append(node.ParentNodes,
						&inventoryv1.GetTreeHierarchyResponse_Node{
							ResourceId:   resID,
							ResourceKind: kind,
						})
				})
				return &node
			})
		return &inventoryv1.GetTreeHierarchyResponse{
			Tree: responseNodes,
		}, nil
	}
}

// Apply the given function to all parents of the given Tree Node.
func applyFuncToParents(node *treeNode, f func(*int, inventoryv1.ResourceKind)) {
	parents := map[inventoryv1.ResourceKind]*int{
		inventoryv1.ResourceKind_RESOURCE_KIND_SITE:   node.siteParentID,
		inventoryv1.ResourceKind_RESOURCE_KIND_OU:     node.ouParentID,
		inventoryv1.ResourceKind_RESOURCE_KIND_REGION: node.regionParentID,
	}
	for kind, parentID := range parents {
		if parentID != nil {
			f(parentID, kind)
		}
	}
}

// Common error to return when failing to validate the tree hierarchy.
var errInvalidTreeHierarchyResourceID = errors.Errorfc(
	codes.InvalidArgument,
	"invalid resource ID in the Tree hierarchy response",
)

// ValidateTreeHierarchyResponse returns an error if the tree hierarchy contains resource with unexpected parents.
// For example, a site containing a host as a parent.
func ValidateTreeHierarchyResponse(in *inventoryv1.GetTreeHierarchyResponse) error {
	return collections.FirstError[*inventoryv1.GetTreeHierarchyResponse_TreeNode](in.GetTree(), validateTreeNodeParents)
}

func validateTreeNodeParents(node *inventoryv1.GetTreeHierarchyResponse_TreeNode) error {
	parentKinds := mapTreeNodeToParentKinds(node)
	// validate parent kinds wrt current node kind
	switch node.CurrentNode.ResourceKind {
	case inventoryv1.ResourceKind_RESOURCE_KIND_HOST:
		if !validateResourceKinds(parentKinds, []inventoryv1.ResourceKind{inventoryv1.ResourceKind_RESOURCE_KIND_SITE}) {
			return errInvalidTreeHierarchyResourceID
		}
	case inventoryv1.ResourceKind_RESOURCE_KIND_SITE:
		if !validateResourceKinds(
			parentKinds,
			[]inventoryv1.ResourceKind{
				inventoryv1.ResourceKind_RESOURCE_KIND_REGION,
				inventoryv1.ResourceKind_RESOURCE_KIND_OU,
			}) {
			return errInvalidTreeHierarchyResourceID
		}
	case inventoryv1.ResourceKind_RESOURCE_KIND_REGION:
		if !validateResourceKinds(
			parentKinds,
			[]inventoryv1.ResourceKind{inventoryv1.ResourceKind_RESOURCE_KIND_REGION}) {
			return errInvalidTreeHierarchyResourceID
		}
	case inventoryv1.ResourceKind_RESOURCE_KIND_OU:
		if !validateResourceKinds(
			parentKinds,
			[]inventoryv1.ResourceKind{inventoryv1.ResourceKind_RESOURCE_KIND_OU}) {
			return errInvalidTreeHierarchyResourceID
		}
	default:
		return errInvalidTreeHierarchyResourceID
	}
	return nil
}

func validateResourceKinds(resourceKinds, expectedKinds []inventoryv1.ResourceKind) bool {
	err := collections.FirstError[inventoryv1.ResourceKind](resourceKinds, func(kind inventoryv1.ResourceKind) error {
		for _, expKind := range expectedKinds {
			if expKind == kind {
				return nil
			}
		}
		// Return any error, we don't care
		return errors.Errorfc(codes.InvalidArgument, "")
	})
	return err == nil
}

func mapTreeNodeToParentKinds(node *inventoryv1.GetTreeHierarchyResponse_TreeNode) []inventoryv1.ResourceKind {
	return collections.MapSlice[*inventoryv1.GetTreeHierarchyResponse_Node, inventoryv1.ResourceKind](
		node.GetParentNodes(),
		func(node *inventoryv1.GetTreeHierarchyResponse_Node) inventoryv1.ResourceKind {
			return node.GetResourceKind()
		})
}

func getSitesPerRegionQuery(tenantID string, resourceIDs []string) (string, []interface{}) {
	args := make([]interface{}, 1, 1+len(resourceIDs))
	args[0] = tenantID
	otherArgs, placeholders := getSQLPlaceholdersAndArgs(resourceIDs, 1)
	args = append(args, otherArgs...)
	whereClause := "WHERE " + tenantIDAnd + " resource_id IN (" + placeholders + ")"
	query := `
	WITH RECURSIVE RegionHierarchy AS (
		SELECT ID AS id, resource_id, resource_id AS root_id
			FROM region_resources
		` + whereClause + `
			UNION ALL
		SELECT r.ID AS id, r.resource_id AS resource_id, rh.root_id
			FROM region_resources r
			JOIN RegionHierarchy rh ON rh.id = r.region_resource_parent_region
			` + whereTenantID + `
		)
	SELECT rh.root_id, COALESCE(COUNT(o.id), 0) AS total_sites
	FROM RegionHierarchy rh
	LEFT JOIN site_resources o ON o.site_resource_region = rh.id
	GROUP BY rh.root_id;`
	return query, args
}

func getSitesPerRegion(ctx context.Context, client *ent.Client, tenantID string, resourceIDs []string) (map[string]int, error) {
	query, args := getSitesPerRegionQuery(tenantID, resourceIDs)
	rows, err := client.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, logAndSanitizeErrorRawSQLf(err, "error querying getSitesPerRegion")
	}

	results := make(map[string]int)
	for rows.Next() {
		var resourceID string
		var totalSites int
		if err := rows.Scan(&resourceID, &totalSites); err != nil {
			return nil, logAndSanitizeErrorRawSQLf(err, "error parsing results while querying getSitesPerRegion")
		}
		results[resourceID] = totalSites
	}
	return results, nil
}

func getSitesPerRegionCreator(filter *inventoryv1.GetSitesPerRegionRequest) func(context.Context, *ent.Tx) (
	*inventoryv1.GetSitesPerRegionResponse, error) {
	return func(ctx context.Context, tx *ent.Tx) (*inventoryv1.GetSitesPerRegionResponse, error) {
		resourceIDs := filter.GetFilter()
		siterPerReg, err := getSitesPerRegion(ctx, tx.Client(), filter.GetTenantId(), resourceIDs)
		if err != nil {
			return nil, err
		}

		responseNodes := []*inventoryv1.GetSitesPerRegionResponse_Node{}
		for regionID, amountSites := range siterPerReg {
			int32AmountSites, err := util.IntToInt32(amountSites)
			if err != nil {
				zlog.InfraSec().Err(err).Msgf("failed converting sites per region")
				return nil, err
			}
			node := &inventoryv1.GetSitesPerRegionResponse_Node{
				ResourceId: regionID,
				ChildSites: int32AmountSites,
			}
			responseNodes = append(responseNodes, node)
		}

		return &inventoryv1.GetSitesPerRegionResponse{
			Regions: responseNodes,
		}, nil
	}
}

func (is *InvStore) GetSitesPerRegion(ctx context.Context, filter *inventoryv1.GetSitesPerRegionRequest) (
	*inventoryv1.GetSitesPerRegionResponse, error,
) {
	res, err := ExecuteInRoTxAndReturnSingle[inventoryv1.GetSitesPerRegionResponse](is)(ctx, getSitesPerRegionCreator(filter))
	if err != nil {
		zlog.InfraSec().InfraErr(err).Send()
		return nil, err
	}
	if err := validator.ValidateMessage(res); err != nil {
		zlog.InfraSec().InfraErr(err).Send()
		return nil, errors.Wrap(err)
	}
	return res, nil
}

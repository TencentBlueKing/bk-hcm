/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

// Package routetable defines route table service.
package routetable

import (
	"fmt"

	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	adcore "hcm/pkg/adaptor/types/core"
	adroutetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/api/core"
	coreroutetable "hcm/pkg/api/core/cloud/route-table"
	dataservice "hcm/pkg/api/data-service"
	routetable "hcm/pkg/api/data-service/cloud/route-table"
	hcroutetable "hcm/pkg/api/hc-service/route-table"
	hcservice "hcm/pkg/api/hc-service/vpc"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/uuid"
)

// GcpRouteTableSync sync gcp cloud route table.
func GcpRouteTableSync(kt *kit.Kit, req *hcroutetable.GcpRouteTableSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	// syncs route table list from cloudapi.
	allCloudIDMap, err := SyncGcpRouteTableList(kt, req, adaptor, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable request cloudapi response failed. accountID: %s, cloudIDs: %v, err: %v",
			enumor.Gcp, req.AccountID, req.CloudIDs, err)
		return nil, err
	}

	// compare and delete route table idmap from db.
	err = compareDeleteGcpRouteTableList(kt, req, allCloudIDMap, adaptor, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable compare delete and dblist failed. accountID: %s, cloudIDs: %v, err: %v",
			enumor.Gcp, req.AccountID, req.CloudIDs, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// SyncGcpRouteTableList sync route table list from cloudapi.
func SyncGcpRouteTableList(kt *kit.Kit, req *hcroutetable.GcpRouteTableSyncReq, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client) (map[string]bool, error) {

	cli, err := adaptor.Gcp(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	allCloudIDMap := make(map[string]bool, 0)
	for {
		opt := &adcore.GcpListOption{
			Page: &adcore.GcpPage{
				PageSize: int64(adcore.GcpQueryLimit / 10),
			},
		}

		if nextToken != "" {
			opt.Page.PageToken = nextToken
		}
		if len(req.CloudIDs) > 0 {
			opt.CloudIDs = req.CloudIDs
		}
		if len(req.SelfLinks) > 0 {
			opt.SelfLinks = req.SelfLinks
		}

		tmpList, tmpErr := cli.ListRoute(kt, opt)
		if tmpErr != nil {
			logs.Errorf("%s-routetable batch get cloudapi failed. accountID: %s, cloudIDs: %v, err: %v",
				enumor.Gcp, req.AccountID, req.CloudIDs, tmpErr)
			return allCloudIDMap, tmpErr
		}
		if len(tmpList.Details) == 0 {
			break
		}

		cloudIDs := make([]string, 0)
		for _, item := range tmpList.Details {
			cloudIDs = append(cloudIDs, item.CloudID)
			allCloudIDMap[item.CloudID] = true
		}

		// get route table info from db.
		resourceDBMap, err := GetGcpRouteTableInfoFromDB(kt, cloudIDs, dataCli)
		if err != nil {
			logs.Errorf("%s-routetable get routetabledblist failed. accountID: %s, cloudIDs: %v, err: %v",
				enumor.Gcp, req.AccountID, req.CloudIDs, err)
			return allCloudIDMap, err
		}

		// compare and update route table list.
		err = compareUpdateGcpRouteTableList(kt, req, tmpList, resourceDBMap, dataCli)
		if err != nil {
			logs.Errorf("%s-routetable compare and update routetabledblist failed. accountID: %s, "+
				"cloudIDs: %v, err: %v", enumor.Gcp, req.AccountID, req.CloudIDs, err)
			return allCloudIDMap, err
		}

		if len(tmpList.NextPageToken) == 0 {
			break
		}
		nextToken = tmpList.NextPageToken
	}

	return allCloudIDMap, nil
}

// GetGcpRouteTableInfoFromDB get gcp route table info from db.
func GetGcpRouteTableInfoFromDB(kt *kit.Kit, cloudIDs []string, dataCli *dataclient.Client) (
	map[string]coreroutetable.GcpRoute, error) {
	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "cloud_id",
				Op:    filter.In.Factory(),
				Value: cloudIDs,
			},
		},
	}

	dbQueryReq := &routetable.GcpRouteListReq{
		ListReq: &core.ListReq{
			Filter: expr,
			Page:   &core.BasePage{Count: false, Start: 0, Limit: core.DefaultMaxPageLimit},
		},
	}
	dbList, err := dataCli.Gcp.RouteTable.ListRoute(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("%s-routetable batch get routetablelist db error. limit: %d, err: %v",
			enumor.Gcp, core.DefaultMaxPageLimit, err)
		return nil, err
	}

	resourceMap := make(map[string]coreroutetable.GcpRoute, 0)
	if len(dbList.Details) == 0 {
		return resourceMap, nil
	}

	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// compareUpdateGcpRouteTableList compare and update route table list.
func compareUpdateGcpRouteTableList(kt *kit.Kit, req *hcroutetable.GcpRouteTableSyncReq,
	list *adroutetable.GcpRouteListResult, resourceDBMap map[string]coreroutetable.GcpRoute,
	dataCli *dataclient.Client) error {

	createResources, err := filterGcpRouteTableList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &routetable.GcpRouteBatchCreateReq{
			GcpRoutes: createResources,
		}
		if _, err = dataCli.Gcp.RouteTable.BatchCreateRoute(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("%s-routetable batch compare db create failed. accountID: %s, cloudIDs: %v, err: %v",
				enumor.Gcp, req.AccountID, req.CloudIDs, err)
			return err
		}
	}

	return nil
}

// filterGcpRouteTableList filter gcp route table list
func filterGcpRouteTableList(req *hcroutetable.GcpRouteTableSyncReq, list *adroutetable.GcpRouteListResult,
	resourceDBMap map[string]coreroutetable.GcpRoute) (createResources []routetable.GcpRouteCreateReq, err error) {

	if list == nil || len(list.Details) == 0 {
		return createResources,
			fmt.Errorf("cloudapi routetablelist is empty, accountID: %s, cloudIDs: %v",
				req.AccountID, req.CloudIDs)
	}

	for _, item := range list.Details {
		// need compare and update resource data
		if _, ok := resourceDBMap[item.CloudID]; ok {
			logs.Errorf("%s-routetable batch compare do not update, but now changed. accountID: %s, "+
				"cloudIDs: %v, cloudID: %s", enumor.Gcp, req.AccountID, req.CloudIDs, item.CloudID)
			continue
		}

		// need add resource data
		tmpRes := routetable.GcpRouteCreateReq{
			CloudID:          item.CloudID,
			SelfLink:         item.SelfLink,
			Network:          item.Network,
			Name:             item.Name,
			DestRange:        item.DestRange,
			NextHopGateway:   item.NextHopGateway,
			NextHopIlb:       item.NextHopIlb,
			NextHopInstance:  item.NextHopInstance,
			NextHopIp:        item.NextHopIp,
			NextHopNetwork:   item.NextHopNetwork,
			NextHopPeering:   item.NextHopPeering,
			NextHopVpnTunnel: item.NextHopVpnTunnel,
			Priority:         item.Priority,
			RouteStatus:      item.RouteStatus,
			RouteType:        item.RouteType,
			Tags:             item.Tags,
			Memo:             item.Memo,
		}
		createResources = append(createResources, tmpRes)
	}

	return createResources, nil
}

// compareDeleteGcpRouteTableList compare and delete route table list from db.
func compareDeleteGcpRouteTableList(kt *kit.Kit, req *hcroutetable.GcpRouteTableSyncReq, allCloudIDMap map[string]bool,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) error {

	if len(req.CloudIDs) > 0 || len(req.SelfLinks) > 0 {
		return nil
	}
	page := uint32(0)
	for {
		count := core.DefaultMaxPageLimit
		offset := page * uint32(count)
		expr := &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "route_table_id",
					Op:    filter.NotEqual.Factory(),
					Value: "",
				},
			},
		}
		dbQueryReq := &routetable.GcpRouteListReq{
			ListReq: &core.ListReq{
				Filter: expr,
				Page:   &core.BasePage{Count: false, Start: offset, Limit: count},
			},
		}
		dbList, err := dataCli.Gcp.RouteTable.ListRoute(kt.Ctx, kt.Header(), dbQueryReq)
		if err != nil {
			logs.Errorf("%s-routetable batch get vpclist db error. offset: %d, limit: %d, "+
				"err: %v", enumor.Gcp, offset, count, err)
			return err
		}

		if len(dbList.Details) == 0 {
			return nil
		}

		deleteCloudIDMap := make(map[string]hcroutetable.RouteDeleteReq, 0)
		for _, item := range dbList.Details {
			if _, ok := allCloudIDMap[item.CloudID]; !ok {
				deleteCloudIDMap[item.CloudID] = hcroutetable.RouteDeleteReq{
					RouteID:      item.ID,
					RouteTableID: item.RouteTableID,
				}
			}
		}

		// batch query need delete route table list
		deleteIDs, deleteIDMap := GetNeedDeleteGcpRouteTableList(kt, req, deleteCloudIDMap, adaptor)
		if len(deleteIDs) > 0 {
			err = cancelRouteTableSubnetRel(kt, dataCli, enumor.Gcp, deleteIDs)
			if err != nil {
				logs.Errorf("%s-routetable batch cancel subnet rel failed. deleteIDs: %v, err: %v",
					enumor.Gcp, deleteIDs, err)
				return err
			}

			err = BatchDeleteGcpRouteByIDs(kt, deleteIDMap, dataCli)
			if err != nil {
				logs.Errorf("%s-routetable batch compare db delete failed. deleteIDs: %v, err: %v",
					enumor.Gcp, deleteIDs, err)
				return err
			}
		}
		deleteIDs = nil

		if len(dbList.Details) < int(count) {
			break
		}
		page++
	}
	allCloudIDMap = nil

	return nil
}

// GetNeedDeleteGcpRouteTableList get need delete gcp route table list
func GetNeedDeleteGcpRouteTableList(kt *kit.Kit, req *hcroutetable.GcpRouteTableSyncReq,
	deleteCloudIDMap map[string]hcroutetable.RouteDeleteReq,
	adaptor *cloudclient.CloudAdaptorClient) ([]string, map[string][]string) {

	deleteRouteTableIDs := make([]string, 0, len(deleteCloudIDMap))
	deleteIDMap := make(map[string][]string, len(deleteCloudIDMap))
	if len(deleteCloudIDMap) == 0 {
		return deleteRouteTableIDs, deleteIDMap
	}

	cli, err := adaptor.Gcp(kt, req.AccountID)
	if err != nil {
		logs.Errorf("%s-routetable get account failed. accountID: %s, cloudIDs: %v, err: %v",
			enumor.Gcp, req.AccountID, req.CloudIDs, err)
		return deleteRouteTableIDs, deleteIDMap
	}

	var tmpResourceIDs []string
	for tmpCloudID, tmpItem := range deleteCloudIDMap {
		tmpResourceIDs = append(tmpResourceIDs, tmpCloudID)
		deleteIDMap[tmpItem.RouteTableID] = append(deleteIDMap[tmpItem.RouteTableID], tmpItem.RouteID)
		deleteRouteTableIDs = append(deleteRouteTableIDs, tmpItem.RouteTableID)
	}
	deleteRouteTableIDs = slice.Unique(deleteRouteTableIDs)

	opt := &adcore.GcpListOption{
		CloudIDs: tmpResourceIDs,
	}
	tmpList, tmpErr := cli.ListRoute(kt, opt)
	if tmpErr != nil {
		logs.Errorf("%s-routetable batch get cloudapi failed. accountID: %s, cloudIDs: %v, err: %v",
			enumor.Gcp, req.AccountID, req.CloudIDs, tmpErr)
		return deleteRouteTableIDs, deleteIDMap
	}

	if len(tmpList.Details) == 0 {
		return deleteRouteTableIDs, deleteIDMap
	}

	for _, item := range tmpList.Details {
		if _, ok := deleteCloudIDMap[item.CloudID]; ok {
			delete(deleteCloudIDMap, item.CloudID)
		}
	}

	deleteRouteTableIDs = make([]string, 0, len(deleteCloudIDMap))
	deleteIDMap = make(map[string][]string, len(deleteCloudIDMap))
	for _, tmpItem := range deleteCloudIDMap {
		deleteIDMap[tmpItem.RouteTableID] = append(deleteIDMap[tmpItem.RouteTableID], tmpItem.RouteID)
		deleteRouteTableIDs = append(deleteRouteTableIDs, tmpItem.RouteTableID)
	}
	deleteRouteTableIDs = slice.Unique(deleteRouteTableIDs)

	return deleteRouteTableIDs, deleteIDMap
}

// BatchDeleteGcpRouteByIDs batch delete route table ids
func BatchDeleteGcpRouteByIDs(kt *kit.Kit, deleteIDMap map[string][]string, dataCli *dataclient.Client) error {
	if len(deleteIDMap) == 0 {
		return nil
	}

	for tmpRouteTableID, deleteIDs := range deleteIDMap {
		querySize := int(filter.DefaultMaxInLimit)
		times := len(deleteIDs) / querySize
		if len(deleteIDs)%querySize != 0 {
			times++
		}

		for i := 0; i < times; i++ {
			var newDeleteIDs []string
			if i == times-1 {
				newDeleteIDs = append(newDeleteIDs, deleteIDs[i*querySize:]...)
			} else {
				newDeleteIDs = append(newDeleteIDs, deleteIDs[i*querySize:(i+1)*querySize]...)
			}

			deleteReq := &dataservice.BatchDeleteReq{
				Filter: tools.ContainersExpression("id", newDeleteIDs),
			}
			if err := dataCli.Gcp.RouteTable.BatchDeleteRoute(kt.Ctx, kt.Header(), tmpRouteTableID, deleteReq); err != nil {
				return err
			}
		}
	}

	return nil
}

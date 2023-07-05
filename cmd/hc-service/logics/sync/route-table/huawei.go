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
	"hcm/pkg/adaptor/huawei"
	adcore "hcm/pkg/adaptor/types/core"
	routetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/api/core"
	cloudRouteTable "hcm/pkg/api/core/cloud/route-table"
	dataservice "hcm/pkg/api/data-service"
	dataproto "hcm/pkg/api/data-service/cloud/route-table"
	hcroutetable "hcm/pkg/api/hc-service/route-table"
	hcservice "hcm/pkg/api/hc-service/vpc"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// HuaWeiRouteTableSync sync huawei cloud route table.
func HuaWeiRouteTableSync(kt *kit.Kit, req *hcroutetable.HuaWeiRouteTableSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	// syncs route table list from cloudapi.
	allCloudIDMap, err := SyncHuaWeiRouteTableList(kt, req, adaptor, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	// compare and delete route table idmap from db.
	err = compareDeleteHuaWeiRouteTableList(kt, req, allCloudIDMap, adaptor, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable compare delete and dblist failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// SyncHuaWeiRouteTableList sync route table list from cloudapi.
func SyncHuaWeiRouteTableList(kt *kit.Kit, req *hcroutetable.HuaWeiRouteTableSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (map[string]bool, error) {

	cli, err := adaptor.HuaWei(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	var nextMarker string
	var allCloudIDMap = make(map[string]bool, 0)
	for {
		tmpIDs, err := BatchGetHuaWeiRouteTableList(kt, req, cli, nextMarker)
		if err != nil {
			logs.Errorf("%s-routetable batch get cloudapi failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, err)
			return nil, err
		}
		if len(tmpIDs) == 0 {
			break
		}

		cloudIDs := make([]string, 0)
		for _, tmpCloudID := range tmpIDs {
			cloudIDs = append(cloudIDs, tmpCloudID)
			allCloudIDMap[tmpCloudID] = true
		}

		// get route table info from db.
		resourceDBMap, err := GetHuaWeiRouteTableInfoFromDB(kt, cloudIDs, dataCli)
		if err != nil {
			logs.Errorf("%s-routetable get routetabledblist failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, err)
			return nil, err
		}

		tmpDetails := make([]routetable.HuaWeiRouteTable, 0, len(tmpIDs))
		for _, tmpCloudID := range tmpIDs {
			routeOpt := &routetable.HuaWeiRouteTableGetOption{
				Region: req.Region,
				ID:     tmpCloudID,
			}
			tmpRouteTable, err := cli.GetRouteTable(kt, routeOpt)
			if err == nil {
				tmpDetails = append(tmpDetails, converter.PtrToVal(tmpRouteTable))
			}
		}
		tmpList := &routetable.HuaWeiRouteTableListResult{
			Details: tmpDetails,
		}
		tmpList.Details = tmpDetails

		// compare and update route table list.
		err = compareUpdateHuaWeiRouteTableList(kt, req, tmpList, resourceDBMap, dataCli)
		if err != nil {
			logs.Errorf("%s-routetable compare and update routetabledblist failed. accountID: %s, "+
				"region: %s, err: %v", enumor.HuaWei, req.AccountID, req.Region, err)
			return nil, err
		}

		if len(req.CloudIDs) > 0 || len(tmpIDs) < adcore.HuaWeiQueryLimit {
			break
		}

		nextMarker = tmpIDs[len(tmpIDs)-1]
	}
	return allCloudIDMap, nil
}

// BatchGetHuaWeiRouteTableList batch get route table list from cloudapi.
func BatchGetHuaWeiRouteTableList(kt *kit.Kit, req *hcroutetable.HuaWeiRouteTableSyncReq, cli *huawei.HuaWei,
	nextMarker string) ([]string, error) {

	if len(req.CloudIDs) > 0 {
		return BatchGetHuaWeiRouteTableListByCloudIDs(kt, req, cli)
	}

	return BatchGetHuaWeiRouteTableAllList(kt, req, cli, nextMarker)
}

// BatchGetHuaWeiRouteTableListByCloudIDs batch get route table list from cloudapi by cloud_ids.
func BatchGetHuaWeiRouteTableListByCloudIDs(kt *kit.Kit, req *hcroutetable.HuaWeiRouteTableSyncReq,
	cli *huawei.HuaWei) ([]string, error) {

	var idArr []string
	for _, tmpCloudID := range req.CloudIDs {
		opt := &routetable.HuaWeiRouteTableListOption{
			Region: req.Region,
			ID:     tmpCloudID,
		}
		tmpIDs, err := cli.ListRouteTableIDs(kt, opt)
		if err != nil {
			logs.Errorf("%s-routetable batch get cloud api by id failed, req: %+v, err: %v, rid: %s",
				enumor.HuaWei, req, err, kt.Rid)
			return nil, err
		}
		idArr = append(idArr, tmpIDs...)
	}

	return idArr, nil
}

// BatchGetHuaWeiRouteTableAllList batch get route table list from cloudapi.
func BatchGetHuaWeiRouteTableAllList(kt *kit.Kit, req *hcroutetable.HuaWeiRouteTableSyncReq, cli *huawei.HuaWei,
	nextMarker string) ([]string, error) {

	opt := &routetable.HuaWeiRouteTableListOption{
		Region: req.Region,
		Page: &adcore.HuaWeiPage{
			Limit: converter.ValToPtr(int32(adcore.HuaWeiQueryLimit)),
		},
	}
	if len(nextMarker) != 0 {
		opt.Page.Marker = converter.ValToPtr(nextMarker)
	}

	tmpIDs, tmpErr := cli.ListRouteTableIDs(kt, opt)
	if tmpErr != nil {
		logs.Errorf("%s-routetable batch get cloud api failed. req: %+v, err: %v, rid: %s",
			enumor.HuaWei, req, tmpErr, kt.Rid)
		return nil, tmpErr
	}

	return tmpIDs, nil
}

// GetHuaWeiRouteTableInfoFromDB get route table info from db.
func GetHuaWeiRouteTableInfoFromDB(kt *kit.Kit, cloudIDs []string, dataCli *dataclient.Client) (
	map[string]*cloudRouteTable.RouteTable[cloudRouteTable.HuaWeiRouteTableExtension], error) {

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "vendor",
				Op:    filter.Equal.Factory(),
				Value: enumor.HuaWei,
			},
			&filter.AtomRule{
				Field: "cloud_id",
				Op:    filter.In.Factory(),
				Value: cloudIDs,
			},
		},
	}
	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   &core.BasePage{Count: false, Start: 0, Limit: core.DefaultMaxPageLimit},
	}
	dbList, err := dataCli.HuaWei.RouteTable.ListRouteTableWithExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("%s-routetable batch get routetablelist db error. limit: %d, err: %v",
			enumor.HuaWei, core.DefaultMaxPageLimit, err)
		return nil, err
	}

	resourceMap := make(map[string]*cloudRouteTable.RouteTable[cloudRouteTable.HuaWeiRouteTableExtension], 0)
	if len(dbList) == 0 {
		return resourceMap, nil
	}

	for _, item := range dbList {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

func checkHuaWeiIsUpdate(item routetable.HuaWeiRouteTable,
	dbInfo *cloudRouteTable.RouteTable[cloudRouteTable.HuaWeiRouteTableExtension]) bool {

	if dbInfo.Name != item.Name {
		return true
	}
	if !assert.IsPtrStringEqual(item.Memo, dbInfo.Memo) {
		return true
	}

	return false
}

// compareUpdateHuaWeiRouteTableList compare and update route table list.
func compareUpdateHuaWeiRouteTableList(kt *kit.Kit, req *hcroutetable.HuaWeiRouteTableSyncReq,
	list *routetable.HuaWeiRouteTableListResult,
	resourceDBMap map[string]*cloudRouteTable.RouteTable[cloudRouteTable.HuaWeiRouteTableExtension],
	dataCli *dataclient.Client) error {

	createResources, updateResources, subnetMap, err := filterHuaWeiRouteTableList(kt, req, list, resourceDBMap,
		dataCli)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &dataproto.RouteTableBaseInfoBatchUpdateReq{
			RouteTables: updateResources,
		}
		if err = dataCli.Global.RouteTable.BatchUpdateBaseInfo(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-routetable batch compare db update failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &dataproto.RouteTableBatchCreateReq[dataproto.HuaWeiRouteTableCreateExt]{
			RouteTables: createResources,
		}
		createIDs, err := dataCli.HuaWei.RouteTable.BatchCreate(kt.Ctx, kt.Header(), createReq)
		if err != nil {
			logs.Errorf("%s-routetable batch compare db create failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, err)
			return err
		}

		existCloudIDMap := make(map[string]string, 0)
		existRouteTableIDMap := make(map[string]string, 0)
		for _, newID := range createIDs.IDs {
			for _, item := range list.Details {
				if _, ok := existRouteTableIDMap[newID]; ok {
					break
				}
				if _, ok := existCloudIDMap[item.CloudID]; ok {
					continue
				}

				err = BatchCreateHuaWeiRoute(kt, newID, &item, dataCli)
				if err != nil {
					logs.Errorf("%s-routetable sync create route failed. accountID: %s, region: %s, err: %v",
						enumor.HuaWei, req.AccountID, req.Region, err)
					continue
				}
				existRouteTableIDMap[newID] = item.CloudID
				existCloudIDMap[item.CloudID] = newID
			}
		}
	}
	if len(subnetMap) > 0 {
		UpdateSubnetRouteTableByIDs(kt, enumor.HuaWei, subnetMap, dataCli)
	}

	return nil
}

// filterHuaWeiRouteTableList filter huawei route table list
func filterHuaWeiRouteTableList(kt *kit.Kit, req *hcroutetable.HuaWeiRouteTableSyncReq,
	list *routetable.HuaWeiRouteTableListResult,
	resourceDBMap map[string]*cloudRouteTable.RouteTable[cloudRouteTable.HuaWeiRouteTableExtension],
	dataCli *dataclient.Client) (createResources []dataproto.RouteTableCreateReq[dataproto.HuaWeiRouteTableCreateExt],
	updateResources []dataproto.RouteTableBaseInfoUpdateReq, subnetMap map[string]dataproto.RouteTableSubnetReq,
	err error) {

	if list == nil || len(list.Details) == 0 {
		return createResources, updateResources, nil,
			fmt.Errorf("cloudapi routetablelist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	subnetMap = make(map[string]dataproto.RouteTableSubnetReq, 0)
	for _, item := range list.Details {
		// need compare and update resource data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if !checkHuaWeiIsUpdate(item, resourceInfo) {
				continue
			}

			tmpRes := dataproto.RouteTableBaseInfoUpdateReq{
				IDs: []string{resourceInfo.ID},
			}
			tmpRes.Data = &dataproto.RouteTableUpdateBaseInfo{
				Name: converter.ValToPtr(item.Name),
				Memo: item.Memo,
			}
			if item.Extension != nil && len(item.Extension.CloudSubnetIDs) > 0 {
				for _, tmpSubnetID := range item.Extension.CloudSubnetIDs {
					subnetMap[tmpSubnetID] = dataproto.RouteTableSubnetReq{
						RouteTableID:      resourceInfo.ID,
						CloudRouteTableID: item.CloudID,
					}
				}
			}

			updateResources = append(updateResources, tmpRes)

			// huawei route sync.
			err = HuaWeiRouteSync(kt, req, resourceInfo.ID, &item, dataCli)
			if err != nil {
				logs.Errorf("%s-routetable sync update route failed. accountID: %s, region: %s, err: %v",
					enumor.HuaWei, req.AccountID, req.Region, err)
			}
			continue
		} else {
			// need add resource data
			tmpRes := dataproto.RouteTableCreateReq[dataproto.HuaWeiRouteTableCreateExt]{
				AccountID:  req.AccountID,
				CloudID:    item.CloudID,
				Name:       converter.ValToPtr(item.Name),
				Region:     item.Region,
				CloudVpcID: item.CloudVpcID,
				Memo:       item.Memo,
				BkBizID:    constant.UnassignedBiz,
			}
			if item.Extension != nil {
				tmpRes.Extension = &dataproto.HuaWeiRouteTableCreateExt{
					Default:  item.Extension.Default,
					TenantID: item.Extension.TenantID,
				}
				for _, tmpSubnetID := range item.Extension.CloudSubnetIDs {
					subnetMap[tmpSubnetID] = dataproto.RouteTableSubnetReq{
						CloudRouteTableID: item.CloudID,
					}
				}
			}

			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, subnetMap, nil
}

// GetNeedDeleteHuaWeiRouteTableList get need delete huawei route table list
func GetNeedDeleteHuaWeiRouteTableList(kt *kit.Kit, req *hcroutetable.HuaWeiRouteTableSyncReq,
	deleteCloudIDMap map[string]string, adaptor *cloudclient.CloudAdaptorClient) []string {

	deleteIDs := make([]string, 0, len(deleteCloudIDMap))
	if len(deleteCloudIDMap) == 0 {
		return deleteIDs
	}

	cli, err := adaptor.HuaWei(kt, req.AccountID)
	if err != nil {
		logs.Errorf("%s-routetable get account failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return deleteIDs
	}

	for tmpCloudID, tmpID := range deleteCloudIDMap {
		opt := &routetable.HuaWeiRouteTableListOption{
			Region: req.Region,
			ID:     tmpCloudID,
		}

		tmpList, tmpErr := cli.ListRouteTableIDs(kt, opt)
		if tmpErr != nil || len(tmpList) == 0 {
			logs.Errorf("%s-routetable batch get cloudapi failed. accountID: %s, region: %s, cloudID: %s, "+
				"err: %v", enumor.HuaWei, req.AccountID, req.Region, tmpCloudID, tmpErr)
			deleteIDs = append(deleteIDs, tmpID)
			continue
		}
	}

	return deleteIDs
}

// HuaWeiRouteSync huawei route sync.
func HuaWeiRouteSync(kt *kit.Kit, req *hcroutetable.HuaWeiRouteTableSyncReq, routeTableID string,
	routeList *routetable.HuaWeiRouteTable, dataCli *dataclient.Client) error {

	if routeList == nil {
		return nil
	}

	// batch get route map from db.
	resourceDBMap, err := BatchGetHuaWeiRouteMapFromDB(kt, req, routeTableID, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable-route batch get routedblist failed. accountID: %s, routeTableID: %s, err: %v",
			enumor.HuaWei, req.AccountID, routeTableID, err)
		return err
	}

	// batch sync vendor route list.
	err = BatchSyncHuaWeiRoute(kt, req, routeTableID, routeList, resourceDBMap, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable-route compare api and dblist failed. accountID: %s, routeTableID: %s, "+
			"err: %v", enumor.HuaWei, req.AccountID, routeTableID, err)
		return err
	}

	return nil
}

// BatchGetHuaWeiRouteMapFromDB batch get route map from db.
func BatchGetHuaWeiRouteMapFromDB(kt *kit.Kit, req *hcroutetable.HuaWeiRouteTableSyncReq,
	routeTableID string, dataCli *dataclient.Client) (map[string]cloudRouteTable.HuaWeiRoute, error) {

	dbList, err := dataCli.HuaWei.RouteTable.ListRoute(kt.Ctx, kt.Header(), routeTableID,
		&core.ListReq{
			Filter: tools.EqualExpression("route_table_id", routeTableID),
			Page: &core.BasePage{
				Limit: core.DefaultMaxPageLimit,
			},
		})
	if err != nil {
		logs.Errorf("%s-routetable-route batch list db error. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	resourceMap := make(map[string]cloudRouteTable.HuaWeiRoute, 0)
	if len(dbList.Details) == 0 {
		return resourceMap, nil
	}

	for _, item := range dbList.Details {
		resourceMap[item.CloudRouteTableID+item.Destination] = item
	}

	return resourceMap, nil
}

// BatchSyncHuaWeiRoute batch sync huawei route.
func BatchSyncHuaWeiRoute(kt *kit.Kit, req *hcroutetable.HuaWeiRouteTableSyncReq, routeTableID string,
	list *routetable.HuaWeiRouteTable, resourceDBMap map[string]cloudRouteTable.HuaWeiRoute,
	dataCli *dataclient.Client) error {

	createResources, updateResources, existIDMap, err := filterHuaWeiRouteList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &dataproto.HuaWeiRouteBatchUpdateReq{
			HuaWeiRoutes: updateResources,
		}
		if err = dataCli.HuaWei.RouteTable.BatchUpdateRoute(kt.Ctx, kt.Header(), routeTableID, updateReq); err != nil {
			logs.Errorf("%s-routetable-route batch compare db update failed. accountID: %s, region: %s, "+
				"routeTableID: %s, err: %v", enumor.HuaWei, req.AccountID, req.Region, routeTableID, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &dataproto.HuaWeiRouteBatchCreateReq{
			HuaWeiRoutes: createResources,
		}
		if _, err = dataCli.HuaWei.RouteTable.BatchCreateRoute(kt.Ctx, kt.Header(), routeTableID, createReq); err != nil {
			logs.Errorf("%s-routetable-route batch compare db create failed. accountID: %s, region: %s, "+
				"routeTableID: %s, err: %v", enumor.HuaWei, req.AccountID, req.Region, routeTableID, err)
			return err
		}
	}

	// delete resource data
	deleteIDs := make([]string, 0)
	for _, resItem := range resourceDBMap {
		if _, ok := existIDMap[resItem.ID]; !ok {
			deleteIDs = append(deleteIDs, resItem.ID)
		}
	}

	if len(deleteIDs) > 0 {
		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("id", deleteIDs),
		}
		if err = dataCli.HuaWei.RouteTable.BatchDeleteRoute(kt.Ctx, kt.Header(), routeTableID, deleteReq); err != nil {
			logs.Errorf("%s-routetable-route batch compare db delete failed. accountID: %s, region: %s, "+
				"routeTableID: %s, delIDs: %v, err: %v", enumor.HuaWei, req.AccountID, req.Region, routeTableID,
				deleteIDs, err)
			return err
		}
	}

	return nil
}

func filterHuaWeiRouteList(req *hcroutetable.HuaWeiRouteTableSyncReq,
	list *routetable.HuaWeiRouteTable, resourceDBMap map[string]cloudRouteTable.HuaWeiRoute) (
	createResources []dataproto.HuaWeiRouteCreateReq, updateResources []dataproto.HuaWeiRouteUpdateReq,
	existIDMap map[string]bool, err error) {

	if list == nil || list.Extension == nil {
		return nil, nil, nil,
			fmt.Errorf("cloudapi huaweiroutelist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	existIDMap = make(map[string]bool, 0)
	for _, routeItem := range list.Extension.Routes {
		// need compare and update huawei route data
		if resourceInfo, ok := resourceDBMap[routeItem.CloudRouteTableID+routeItem.Destination]; ok {
			if resourceInfo.Type == routeItem.Type && resourceInfo.Destination == routeItem.Destination &&
				resourceInfo.NextHop == routeItem.NextHop &&
				converter.PtrToVal(resourceInfo.Memo) == converter.PtrToVal(routeItem.Memo) {
				existIDMap[resourceInfo.ID] = true
				continue
			}

			tmpRes := dataproto.HuaWeiRouteUpdateReq{
				ID: resourceInfo.ID,
			}
			tmpRes.Type = routeItem.Type
			tmpRes.Destination = routeItem.Destination
			tmpRes.NextHop = routeItem.NextHop
			tmpRes.Memo = routeItem.Memo

			updateResources = append(updateResources, tmpRes)
			existIDMap[resourceInfo.ID] = true
		} else {
			// need add huawei route data
			tmpRes := dataproto.HuaWeiRouteCreateReq{
				CloudRouteTableID: routeItem.CloudRouteTableID,
				Type:              routeItem.Type,
				Destination:       routeItem.Destination,
				NextHop:           routeItem.NextHop,
				Memo:              routeItem.Memo,
			}

			createResources = append(createResources, tmpRes)
		}
	}
	return createResources, updateResources, existIDMap, nil
}

// BatchCreateHuaWeiRoute batch create huawei route
func BatchCreateHuaWeiRoute(kt *kit.Kit, newID string, list *routetable.HuaWeiRouteTable,
	dataCli *dataclient.Client) error {

	if list.Extension == nil || len(list.Extension.Routes) == 0 {
		return nil
	}

	createRes := make([]dataproto.HuaWeiRouteCreateReq, 0, len(list.Extension.Routes))
	for _, routeItem := range list.Extension.Routes {
		tmpRes := dataproto.HuaWeiRouteCreateReq{
			CloudRouteTableID: routeItem.CloudRouteTableID,
			Type:              routeItem.Type,
			Destination:       routeItem.Destination,
			NextHop:           routeItem.NextHop,
			Memo:              routeItem.Memo,
		}
		createRes = append(createRes, tmpRes)
	}

	createReq := &dataproto.HuaWeiRouteBatchCreateReq{
		HuaWeiRoutes: createRes,
	}
	if _, err := dataCli.HuaWei.RouteTable.BatchCreateRoute(kt.Ctx, kt.Header(), newID, createReq); err != nil {
		return err
	}

	return nil
}

// compareDeleteHuaWeiRouteTableList compare and delete route table list from db.
func compareDeleteHuaWeiRouteTableList(kt *kit.Kit, req *hcroutetable.HuaWeiRouteTableSyncReq,
	allCloudIDMap map[string]bool, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) error {

	if len(req.CloudIDs) > 0 {
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
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: enumor.HuaWei,
				},
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: req.AccountID,
				},
				&filter.AtomRule{
					Field: "region",
					Op:    filter.Equal.Factory(),
					Value: req.Region,
				},
			},
		}
		dbQueryReq := &core.ListReq{
			Filter: expr,
			Page:   &core.BasePage{Count: false, Start: offset, Limit: count},
		}
		dbList, err := dataCli.Global.RouteTable.List(kt.Ctx, kt.Header(), dbQueryReq)
		if err != nil {
			logs.Errorf("%s-routetable batch get vpclist db error. offset: %d, limit: %d, "+
				"err: %v", enumor.HuaWei, offset, count, err)
			return err
		}

		if len(dbList.Details) == 0 {
			return nil
		}

		deleteCloudIDMap := make(map[string]string, 0)
		for _, item := range dbList.Details {
			if _, ok := allCloudIDMap[item.CloudID]; !ok {
				deleteCloudIDMap[item.CloudID] = item.ID
			}
		}

		// batch query need delete route table list
		deleteIDs := GetNeedDeleteHuaWeiRouteTableList(kt, req, deleteCloudIDMap, adaptor)
		if len(deleteIDs) > 0 {
			err = cancelRouteTableSubnetRel(kt, dataCli, enumor.HuaWei, deleteIDs)
			if err != nil {
				logs.Errorf("%s-routetable batch cancel subnet rel failed. deleteIDs: %v, err: %v",
					enumor.HuaWei, deleteIDs, err)
				return err
			}

			err = BatchDeleteRouteTableByIDs(kt, deleteIDs, dataCli)
			if err != nil {
				logs.Errorf("%s-routetable batch compare db delete failed. deleteIDs: %v, err: %v",
					enumor.HuaWei, deleteIDs, err)
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

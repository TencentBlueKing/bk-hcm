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

// Package routetable defines routetable service.
package routetable

import (
	"fmt"

	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	adcore "hcm/pkg/adaptor/types/core"
	routetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/api/core"
	cloudRouteTable "hcm/pkg/api/core/cloud/route-table"
	dataservice "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
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

// TCloudRouteTableSync sync tencent cloud routetable.
func TCloudRouteTableSync(kt *kit.Kit, req *hcroutetable.TCloudRouteTableSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	// syncs route table list from cloudapi.
	allCloudIDMap, err := SyncTCloudRouteTableList(kt, req, adaptor, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.TCloud, req.AccountID, req.Region, err)
		return nil, err
	}

	// compare and delete route table idmap from db.
	err = compareDeleteTCloudRouteTableList(kt, req, allCloudIDMap, adaptor, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable compare delete and dblist failed. accountID: %s, region: %s, err: %v",
			enumor.TCloud, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// SyncTCloudRouteTableList sync route table list from cloudapi.
func SyncTCloudRouteTableList(kt *kit.Kit, req *hcroutetable.TCloudRouteTableSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (map[string]bool, error) {

	cli, err := adaptor.TCloud(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	page := uint64(0)
	allCloudIDMap := make(map[string]bool, 0)
	for {
		count := uint64(adcore.TCloudQueryLimit)
		offset := page * count
		opt := &adcore.TCloudListOption{
			Region: req.Region,
		}

		// 查询指定CloudIDs
		if len(req.CloudIDs) > 0 {
			opt.CloudIDs = req.CloudIDs
		} else {
			opt.Page = &adcore.TCloudPage{
				Offset: offset,
				Limit:  count,
			}
		}

		tmpList, tmpErr := cli.ListRouteTable(kt, opt)
		if tmpErr != nil {
			logs.Errorf("%s-routetable batch get cloudapi failed. accountID: %s, region: %s, offset: %d, "+
				"count: %d, err: %v", enumor.TCloud, req.AccountID, req.Region, offset, count, tmpErr)
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
		resourceDBMap, err := GetTCloudRouteTableInfoFromDB(kt, cloudIDs, dataCli)
		if err != nil {
			logs.Errorf("%s-routetable get routetabledblist failed. accountID: %s, region: %s, err: %v",
				enumor.TCloud, req.AccountID, req.Region, err)
			return allCloudIDMap, err
		}

		// compare and update route table list.
		err = compareUpdateTCloudRouteTableList(kt, req, tmpList, resourceDBMap, dataCli)
		if err != nil {
			logs.Errorf("%s-routetable compare and update routetabledblist failed. accountID: %s, "+
				"region: %s, err: %v", enumor.TCloud, req.AccountID, req.Region, err)
			return allCloudIDMap, err
		}

		if len(tmpList.Details) < int(count) {
			break
		}

		page++
	}

	return allCloudIDMap, nil
}

// GetTCloudRouteTableInfoFromDB get route table info from db.
func GetTCloudRouteTableInfoFromDB(kt *kit.Kit, cloudIDs []string, dataCli *dataclient.Client) (
	map[string]*cloudRouteTable.RouteTable[cloudRouteTable.TCloudRouteTableExtension], error) {

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "vendor",
				Op:    filter.Equal.Factory(),
				Value: enumor.TCloud,
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
	dbList, err := dataCli.TCloud.RouteTable.ListRouteTableWithExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("%s-routetable batch get routetablelist db error. limit: %d, err: %v",
			enumor.TCloud, core.DefaultMaxPageLimit, err)
		return nil, err
	}

	resourceMap := make(map[string]*cloudRouteTable.RouteTable[cloudRouteTable.TCloudRouteTableExtension], 0)
	if len(dbList) == 0 {
		return resourceMap, nil
	}

	for _, item := range dbList {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// compareUpdateTCloudRouteTableList compare and update route table list.
func compareUpdateTCloudRouteTableList(kt *kit.Kit, req *hcroutetable.TCloudRouteTableSyncReq,
	list *routetable.TCloudRouteTableListResult,
	resourceDBMap map[string]*cloudRouteTable.RouteTable[cloudRouteTable.TCloudRouteTableExtension],
	dataCli *dataclient.Client) error {

	createResources, updateResources, subnetMap, err := filterTCloudRouteTableList(kt, req, list,
		resourceDBMap, dataCli)
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
				enumor.TCloud, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &dataproto.RouteTableBatchCreateReq[dataproto.TCloudRouteTableCreateExt]{
			RouteTables: createResources,
		}
		createIDs, err := dataCli.TCloud.RouteTable.BatchCreate(kt.Ctx, kt.Header(), createReq)
		if err != nil {
			logs.Errorf("%s-routetable batch compare db create failed. accountID: %s, region: %s, err: %v",
				enumor.TCloud, req.AccountID, req.Region, err)
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

				err = BatchCreateTCloudRoute(kt, newID, &item, dataCli)
				if err != nil {
					logs.Errorf("%s-routetable sync create route failed. accountID: %s, region: %s, err: %v",
						enumor.TCloud, req.AccountID, req.Region, err)
					continue
				}
				existRouteTableIDMap[newID] = item.CloudID
				existCloudIDMap[item.CloudID] = newID
			}
		}
	}
	if len(subnetMap) > 0 {
		UpdateSubnetRouteTableByIDs(kt, enumor.TCloud, subnetMap, dataCli)
	}

	return nil
}

func UpdateSubnetRouteTableByIDs(kt *kit.Kit, vendor enumor.Vendor, subnetMap map[string]dataproto.RouteTableSubnetReq,
	dataCli *dataclient.Client) {

	tmpCloudIDs := make([]string, 0)
	tmpCloudSubnetIDs := make([]string, 0)
	for tmpSubnetID, tmpRouteItem := range subnetMap {
		tmpCloudIDs = append(tmpCloudIDs, tmpRouteItem.CloudRouteTableID)
		tmpCloudSubnetIDs = append(tmpCloudSubnetIDs, tmpSubnetID)
	}
	subnetListReq := &core.ListReq{
		Fields: []string{"id", "cloud_id", "cloud_route_table_id", "route_table_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: tmpCloudSubnetIDs,
				},
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: vendor,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	subnetList, err := dataCli.Global.Subnet.List(kt.Ctx, kt.Header(), subnetListReq)
	if err != nil {
		logs.Errorf("%s-routetable update subnet route_table_id failed. subnetMap: %+v, err: %v",
			vendor, subnetMap, err)
		return
	}
	if len(subnetList.Details) == 0 {
		return
	}

	rtListReq := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: tmpCloudIDs,
				},
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: vendor,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	routeTableList, err := dataCli.Global.RouteTable.List(kt.Ctx, kt.Header(), rtListReq)
	if err != nil {
		return
	}
	routeTableInfoMap := make(map[string]dataproto.RouteTableSubnetReq, 0)
	for _, rtItem := range routeTableList.Details {
		routeTableInfoMap[rtItem.CloudID] = dataproto.RouteTableSubnetReq{
			RouteTableID:      rtItem.ID,
			CloudRouteTableID: rtItem.CloudID,
		}
	}

	tmpSubnetArr := make([]protocloud.SubnetBaseInfoUpdateReq, 0)
	for _, tmpItem := range subnetList.Details {
		rtSubnetInfo, ok := subnetMap[tmpItem.CloudID]
		if !ok {
			continue
		}
		tmpSubnetReq := protocloud.SubnetBaseInfoUpdateReq{
			IDs: []string{tmpItem.ID},
			Data: &protocloud.SubnetUpdateBaseInfo{
				CloudRouteTableID: converter.ValToPtr(rtSubnetInfo.CloudRouteTableID),
			},
		}
		// 检查routeTable表的cloud_id是否存在
		if rtInfo, ok := routeTableInfoMap[rtSubnetInfo.CloudRouteTableID]; ok {
			tmpSubnetReq.Data.RouteTableID = converter.ValToPtr(rtInfo.RouteTableID)
		}
		tmpSubnetArr = append(tmpSubnetArr, tmpSubnetReq)
	}

	if len(tmpSubnetArr) > 0 {
		subnetReq := &protocloud.SubnetBaseInfoBatchUpdateReq{
			Subnets: tmpSubnetArr,
		}
		err = dataCli.Global.Subnet.BatchUpdateBaseInfo(kt.Ctx, kt.Header(), subnetReq)
	}
}

// filterTCloudRouteTableList filter tcloud route table list
func filterTCloudRouteTableList(kt *kit.Kit, req *hcroutetable.TCloudRouteTableSyncReq,
	list *routetable.TCloudRouteTableListResult,
	resourceDBMap map[string]*cloudRouteTable.RouteTable[cloudRouteTable.TCloudRouteTableExtension],
	dataCli *dataclient.Client) (createResources []dataproto.RouteTableCreateReq[dataproto.TCloudRouteTableCreateExt],
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
			if !checkTCloudIsUpdate(item, resourceInfo) {
				continue
			}

			tmpRes := dataproto.RouteTableBaseInfoUpdateReq{
				IDs: []string{resourceInfo.ID},
			}
			tmpRes.Data = &dataproto.RouteTableUpdateBaseInfo{
				Name: converter.ValToPtr(item.Name),
				Memo: item.Memo,
			}
			if item.Extension != nil && len(item.Extension.Associations) > 0 {
				for _, subnetItem := range item.Extension.Associations {
					subnetMap[subnetItem.CloudSubnetID] = dataproto.RouteTableSubnetReq{
						RouteTableID:      resourceInfo.ID,
						CloudRouteTableID: item.CloudID,
					}
				}
			}

			updateResources = append(updateResources, tmpRes)

			// tcloud route sync.
			err = TCloudRouteSync(kt, req, resourceInfo.ID, &item, dataCli)
			if err != nil {
				logs.Errorf("%s-routetable sync update route failed. accountID: %s, region: %s, err: %v",
					enumor.TCloud, req.AccountID, req.Region, err)
			}
			continue
		} else {
			// need add resource data
			tmpRes := dataproto.RouteTableCreateReq[dataproto.TCloudRouteTableCreateExt]{
				AccountID:  req.AccountID,
				CloudID:    item.CloudID,
				Name:       converter.ValToPtr(item.Name),
				Region:     item.Region,
				CloudVpcID: item.CloudVpcID,
				Memo:       item.Memo,
				BkBizID:    constant.UnassignedBiz,
			}
			if item.Extension != nil {
				tmpRes.Extension = &dataproto.TCloudRouteTableCreateExt{
					Main: item.Extension.Main,
				}
				for _, subnetItem := range item.Extension.Associations {
					subnetMap[subnetItem.CloudSubnetID] = dataproto.RouteTableSubnetReq{
						CloudRouteTableID: item.CloudID,
					}
				}
			}

			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, subnetMap, nil
}

func checkTCloudIsUpdate(item routetable.TCloudRouteTable,
	dbInfo *cloudRouteTable.RouteTable[cloudRouteTable.TCloudRouteTableExtension]) bool {

	if dbInfo.Name != item.Name {
		return true
	}
	if !assert.IsPtrStringEqual(item.Memo, dbInfo.Memo) {
		return true
	}

	return false
}

// compareDeleteTCloudRouteTableList compare and delete route table list from db.
func compareDeleteTCloudRouteTableList(kt *kit.Kit, req *hcroutetable.TCloudRouteTableSyncReq,
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
					Value: string(enumor.TCloud),
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
				"err: %v", enumor.TCloud, offset, count, err)
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
		deleteIDs := GetNeedDeleteTCloudRouteTableList(kt, req, deleteCloudIDMap, adaptor)
		if len(deleteIDs) > 0 {
			err = cancelRouteTableSubnetRel(kt, dataCli, enumor.TCloud, deleteIDs)
			if err != nil {
				logs.Errorf("%s-routetable batch cancel subnet rel failed. deleteIDs: %v, err: %v",
					enumor.TCloud, deleteIDs, err)
				return err
			}

			err = BatchDeleteRouteTableByIDs(kt, deleteIDs, dataCli)
			if err != nil {
				logs.Errorf("%s-routetable batch compare db delete failed. deleteIDs: %v, err: %v",
					enumor.TCloud, deleteIDs, err)
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

// BatchDeleteRouteTableByIDs batch delete route table ids
func BatchDeleteRouteTableByIDs(kt *kit.Kit, deleteIDs []string, dataCli *dataclient.Client) error {
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
		if err := dataCli.Global.RouteTable.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
			return err
		}
	}

	return nil
}

// GetNeedDeleteTCloudRouteTableList get need delete tcloud route table list
func GetNeedDeleteTCloudRouteTableList(kt *kit.Kit, req *hcroutetable.TCloudRouteTableSyncReq,
	deleteCloudIDMap map[string]string, adaptor *cloudclient.CloudAdaptorClient) []string {

	deleteIDs := make([]string, 0, len(deleteCloudIDMap))
	if len(deleteCloudIDMap) == 0 {
		return deleteIDs
	}

	cli, err := adaptor.TCloud(kt, req.AccountID)
	if err != nil {
		logs.Errorf("%s-routetable get account failed. accountID: %s, region: %s, err: %v",
			enumor.TCloud, req.AccountID, req.Region, err)
		return deleteIDs
	}

	tmpResourceIDs := make([]string, len(deleteCloudIDMap))
	for tmpCloudID, tmpID := range deleteCloudIDMap {
		tmpResourceIDs = append(tmpResourceIDs, tmpCloudID)
		deleteIDs = append(deleteIDs, tmpID)
	}

	opt := &adcore.TCloudListOption{
		Region:   req.Region,
		CloudIDs: tmpResourceIDs,
		Page: &adcore.TCloudPage{
			Offset: 0,
			Limit:  adcore.TCloudQueryLimit,
		},
	}

	tmpList, tmpErr := cli.ListRouteTable(kt, opt)
	if tmpErr != nil {
		logs.Errorf("%s-routetable batch get cloudapi failed. accountID: %s, opt: %+v, err: %v",
			enumor.TCloud, req.AccountID, opt, tmpErr)
		return deleteIDs
	}

	if len(tmpList.Details) == 0 {
		return deleteIDs
	}

	for _, item := range tmpList.Details {
		if _, ok := deleteCloudIDMap[item.CloudID]; ok {
			delete(deleteCloudIDMap, item.CloudID)
		}
	}

	deleteIDs = make([]string, 0, len(deleteCloudIDMap))
	for _, tmpID := range deleteCloudIDMap {
		deleteIDs = append(deleteIDs, tmpID)
	}

	return deleteIDs
}

// TCloudRouteSync tcloud route sync.
func TCloudRouteSync(kt *kit.Kit, req *hcroutetable.TCloudRouteTableSyncReq, routeTableID string,
	routeList *routetable.TCloudRouteTable, dataCli *dataclient.Client) error {

	if routeList == nil {
		return nil
	}

	// batch get route map from db.
	resourceDBMap, err := BatchGetTCloudRouteMapFromDB(kt, req, routeTableID, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable-route batch get routedblist failed. accountID: %s, routeTableID: %s, "+
			"err: %v", enumor.TCloud, req.AccountID, routeTableID, err)
		return err
	}

	// batch sync vendor route list.
	err = BatchSyncTCloudRoute(kt, req, routeTableID, routeList, resourceDBMap, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable-route compare api and dblist failed. accountID: %s, routeTableID: %s, "+
			"err: %v", enumor.TCloud, req.AccountID, routeTableID, err)
		return err
	}

	return nil
}

// BatchGetTCloudRouteMapFromDB batch get route map from db.
func BatchGetTCloudRouteMapFromDB(kt *kit.Kit, req *hcroutetable.TCloudRouteTableSyncReq,
	routeTableID string, dataCli *dataclient.Client) (map[string]cloudRouteTable.TCloudRoute, error) {

	dbList, err := dataCli.TCloud.RouteTable.ListRoute(kt.Ctx, kt.Header(), routeTableID, &core.ListReq{
		Filter: tools.EqualExpression("route_table_id", routeTableID),
		Page: &core.BasePage{
			Limit: core.DefaultMaxPageLimit,
		},
	})
	if err != nil {
		logs.Errorf("%s-routetable-route batch list db error. accountID: %s, region: %s, routeTableID: %s, "+
			"err: %v", enumor.TCloud, req.AccountID, req.Region, routeTableID, err)
		return nil, err
	}

	resourceMap := make(map[string]cloudRouteTable.TCloudRoute, 0)
	if len(dbList.Details) == 0 {
		return resourceMap, nil
	}

	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// BatchSyncTCloudRoute batch sync tcloud route.
func BatchSyncTCloudRoute(kt *kit.Kit, req *hcroutetable.TCloudRouteTableSyncReq, routeTableID string,
	list *routetable.TCloudRouteTable, resourceDBMap map[string]cloudRouteTable.TCloudRoute,
	dataCli *dataclient.Client) error {

	createResources, updateResources, existIDMap, err := filterTCloudRouteList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &dataproto.TCloudRouteBatchUpdateReq{
			TCloudRoutes: updateResources,
		}
		if err = dataCli.TCloud.RouteTable.BatchUpdateRoute(kt, routeTableID, updateReq); err != nil {
			logs.Errorf("%s-routetable-route batch compare db update failed. accountID: %s, region: %s, "+
				"routeTableID: %s, err: %v", enumor.TCloud, req.AccountID, req.Region, routeTableID, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		if newCreateRes, ok := createResources[routeTableID]; ok {
			createReq := &dataproto.TCloudRouteBatchCreateReq{
				TCloudRoutes: newCreateRes,
			}

			if _, err = dataCli.TCloud.RouteTable.BatchCreateRoute(kt, routeTableID, createReq); err != nil {
				logs.Errorf("%s-routetable-route batch compare db create failed. accountID: %s, region: %s, "+
					"routeTableID: %s, err: %v", enumor.TCloud, req.AccountID, req.Region, routeTableID, err)
				return err
			}
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
		if err = dataCli.TCloud.RouteTable.BatchDeleteRoute(kt, routeTableID, deleteReq); err != nil {
			logs.Errorf("%s-routetable-route batch compare db delete failed. accountID: %s, region: %s, "+
				"routeTableID: %s, delIDs: %v, err: %v", enumor.TCloud, req.AccountID, req.Region, routeTableID,
				deleteIDs, err)
			return err
		}
	}

	return nil
}

func filterTCloudRouteList(req *hcroutetable.TCloudRouteTableSyncReq,
	list *routetable.TCloudRouteTable, resourceDBMap map[string]cloudRouteTable.TCloudRoute) (
	createResMap map[string][]dataproto.TCloudRouteCreateReq, updateResources []dataproto.TCloudRouteUpdateReq,
	existIDMap map[string]bool, err error) {

	if list == nil || list.Extension == nil {
		return nil, nil, nil,
			fmt.Errorf("cloudapi tcloudroutelist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	existIDMap = make(map[string]bool, 0)
	createResMap = make(map[string][]dataproto.TCloudRouteCreateReq, 0)
	for _, routeItem := range list.Extension.Routes {
		// need compare and update tcloud route data
		if resourceInfo, ok := resourceDBMap[routeItem.CloudID]; ok {
			if resourceInfo.DestinationCidrBlock == routeItem.DestinationCidrBlock &&
				converter.PtrToVal(resourceInfo.DestinationIpv6CidrBlock) ==
					converter.PtrToVal(routeItem.DestinationIpv6CidrBlock) &&
				resourceInfo.GatewayType == routeItem.GatewayType &&
				resourceInfo.CloudGatewayID == routeItem.CloudGatewayID && resourceInfo.Enabled == routeItem.Enabled &&
				resourceInfo.RouteType == routeItem.RouteType &&
				resourceInfo.PublishedToVbc == routeItem.PublishedToVbc &&
				converter.PtrToVal(resourceInfo.Memo) == converter.PtrToVal(routeItem.Memo) {
				existIDMap[resourceInfo.ID] = true
				continue
			}

			tmpRes := dataproto.TCloudRouteUpdateReq{
				ID: resourceInfo.ID,
			}
			tmpRes.DestinationCidrBlock = routeItem.DestinationCidrBlock
			tmpRes.DestinationIpv6CidrBlock = routeItem.DestinationIpv6CidrBlock
			tmpRes.GatewayType = routeItem.GatewayType
			tmpRes.CloudGatewayID = routeItem.CloudGatewayID
			tmpRes.Enabled = converter.ValToPtr(routeItem.Enabled)
			tmpRes.RouteType = routeItem.RouteType
			tmpRes.PublishedToVbc = converter.ValToPtr(routeItem.PublishedToVbc)
			tmpRes.Memo = routeItem.Memo

			updateResources = append(updateResources, tmpRes)
			existIDMap[resourceInfo.ID] = true
		} else {
			// need add tcloud route data
			tmpRes := dataproto.TCloudRouteCreateReq{
				CloudID:                  routeItem.CloudID,
				CloudRouteTableID:        routeItem.CloudRouteTableID,
				DestinationCidrBlock:     routeItem.DestinationCidrBlock,
				DestinationIpv6CidrBlock: routeItem.DestinationIpv6CidrBlock,
				GatewayType:              routeItem.GatewayType,
				CloudGatewayID:           routeItem.CloudGatewayID,
				Enabled:                  routeItem.Enabled,
				RouteType:                routeItem.RouteType,
				PublishedToVbc:           routeItem.PublishedToVbc,
				Memo:                     routeItem.Memo,
			}

			createResMap[routeItem.CloudRouteTableID] = append(createResMap[routeItem.CloudRouteTableID], tmpRes)
		}
	}
	return createResMap, updateResources, existIDMap, nil
}

// BatchCreateTCloudRoute batch create tcloud route
func BatchCreateTCloudRoute(kt *kit.Kit, newID string, list *routetable.TCloudRouteTable,
	dataCli *dataclient.Client) error {

	if list.Extension == nil || len(list.Extension.Routes) == 0 {
		return nil
	}

	createRes := make([]dataproto.TCloudRouteCreateReq, 0, len(list.Extension.Routes))
	for _, routeItem := range list.Extension.Routes {
		tmpRes := dataproto.TCloudRouteCreateReq{
			CloudID:                  routeItem.CloudID,
			CloudRouteTableID:        routeItem.CloudRouteTableID,
			DestinationCidrBlock:     routeItem.DestinationCidrBlock,
			DestinationIpv6CidrBlock: routeItem.DestinationIpv6CidrBlock,
			GatewayType:              routeItem.GatewayType,
			CloudGatewayID:           routeItem.CloudGatewayID,
			Enabled:                  routeItem.Enabled,
			RouteType:                routeItem.RouteType,
			PublishedToVbc:           routeItem.PublishedToVbc,
			Memo:                     routeItem.Memo,
		}
		createRes = append(createRes, tmpRes)
	}

	createReq := &dataproto.TCloudRouteBatchCreateReq{
		TCloudRoutes: createRes,
	}
	if _, err := dataCli.TCloud.RouteTable.BatchCreateRoute(kt, newID, createReq); err != nil {
		return err
	}

	return nil
}

// cancelRouteTableSubnetRel cancel route table and subnet rel.
func cancelRouteTableSubnetRel(kt *kit.Kit, dataCli *dataclient.Client, vendor enumor.Vendor, delIDs []string) error {
	if len(delIDs) == 0 {
		return nil
	}

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "vendor",
				Op:    filter.Equal.Factory(),
				Value: vendor,
			},
			&filter.AtomRule{
				Field: "route_table_id",
				Op:    filter.In.Factory(),
				Value: delIDs,
			},
		},
	}
	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   &core.BasePage{Count: false, Start: 0, Limit: core.DefaultMaxPageLimit},
	}
	dbList, err := dataCli.Global.Subnet.List(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("%s-routetable-route batch cancel route table and subnet rel failed. delIDs: %v, err: %v, "+
			"rid: %s", vendor, delIDs, err, kt.Rid)
		return err
	}

	if len(dbList.Details) == 0 {
		return nil
	}

	var subnetUpdateReq = &protocloud.SubnetBaseInfoBatchUpdateReq{}
	for _, item := range dbList.Details {
		tmpSubnet := protocloud.SubnetBaseInfoUpdateReq{
			IDs: []string{item.ID},
			Data: &protocloud.SubnetUpdateBaseInfo{
				Name:              converter.ValToPtr(item.Name),
				Ipv4Cidr:          item.Ipv4Cidr,
				Ipv6Cidr:          item.Ipv6Cidr,
				Memo:              item.Memo,
				BkBizID:           item.BkBizID,
				CloudRouteTableID: converter.ValToPtr(""),
				RouteTableID:      converter.ValToPtr(""),
			},
		}
		subnetUpdateReq.Subnets = append(subnetUpdateReq.Subnets, tmpSubnet)
	}
	return dataCli.Global.Subnet.BatchUpdateBaseInfo(kt.Ctx, kt.Header(), subnetUpdateReq)
}

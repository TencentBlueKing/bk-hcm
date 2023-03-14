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
	routetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/api/core"
	cloudRouteTable "hcm/pkg/api/core/cloud/route-table"
	dataservice "hcm/pkg/api/data-service"
	dataproto "hcm/pkg/api/data-service/cloud/route-table"
	hcservice "hcm/pkg/api/hc-service"
	hcroutetable "hcm/pkg/api/hc-service/route-table"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// AzureRouteTableSync sync azure cloud route table.
func AzureRouteTableSync(kt *kit.Kit, req *hcroutetable.AzureRouteTableSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	// syncs route table list from cloudapi.
	allCloudIDMap, err := SyncAzureRouteTableList(kt, req, adaptor, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable request cloudapi response failed. accountID: %s, resGroupName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	// compare and delete route table idmap from db.
	err = compareDeleteAzureRouteTableList(kt, req, allCloudIDMap, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable compare delete and dblist failed. accountID: %s, resGroupName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// SyncAzureRouteTableList sync route table list from cloudapi.
func SyncAzureRouteTableList(kt *kit.Kit, req *hcroutetable.AzureRouteTableSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (map[string]bool, error) {

	cli, err := adaptor.Azure(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adcore.AzureListByIDOption{
		ResourceGroupName: req.ResourceGroupName,
	}
	pager, subscriptionID, err := cli.ListRouteTablePage(opt)
	if err != nil {
		logs.Errorf("%s-routetable batch get cloudapi failed. accountID: %s, resGroupName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	details := make([]routetable.AzureRouteTable, 0)
	idMap := make(map[string]struct{}, 0)
	if len(req.CloudIDs) == 0 {
		idMap = converter.StringSliceToMap(req.CloudIDs)
		details = make([]routetable.AzureRouteTable, 0, len(idMap))
	}

	allCloudIDMap := make(map[string]bool, 0)
	for pager.More() {
		page, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("list azure route table but get next page failed, err: %v", err)
		}

		tmpList := &routetable.AzureRouteTableListResult{}
		if len(req.CloudIDs) == 0 {
			for _, routeTable := range page.Value {
				details = append(details, converter.PtrToVal(cli.ConvertRouteTable(routeTable,
					opt.ResourceGroupName, subscriptionID)))
			}
			tmpList.Details = details
		} else {
			for _, routeTable := range page.Value {
				if _, exist := idMap[*routeTable.ID]; !exist {
					continue
				}

				details = append(details, converter.PtrToVal(cli.ConvertRouteTable(routeTable,
					opt.ResourceGroupName, subscriptionID)))
				delete(idMap, *routeTable.ID)

				if len(idMap) == 0 {
					tmpList.Details = details
					break
				}
			}
		}

		cloudIDs := make([]string, 0)
		for _, item := range tmpList.Details {
			cloudIDs = append(cloudIDs, item.CloudID)
			allCloudIDMap[item.CloudID] = true
		}

		// get route table info from db.
		resourceDBMap, err := GetAzureRouteTableInfoFromDB(kt, cloudIDs, dataCli)
		if err != nil {
			logs.Errorf("%s-routetable get routetabledblist failed. accountID: %s, resGroupName: %s, err: %v",
				enumor.Azure, req.AccountID, req.ResourceGroupName, err)
			return nil, err
		}

		// compare and update route table list.
		err = compareUpdateAzureRouteTableList(kt, req, tmpList, resourceDBMap, dataCli)
		if err != nil {
			logs.Errorf("%s-routetable compare and update routetabledblist failed. accountID: %s, "+
				"resGroupName: %s, err: %v", enumor.Azure, req.AccountID, req.ResourceGroupName, err)
			return nil, err
		}
	}

	return allCloudIDMap, nil
}

// GetAzureRouteTableInfoFromDB get route table info from db.
func GetAzureRouteTableInfoFromDB(kt *kit.Kit, cloudIDs []string, dataCli *dataclient.Client) (
	map[string]*cloudRouteTable.RouteTable[cloudRouteTable.AzureRouteTableExtension], error) {

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "vendor",
				Op:    filter.Equal.Factory(),
				Value: enumor.Azure,
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
	dbList, err := dataCli.Azure.RouteTable.ListRouteTableWithExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("%s-routetable batch get routetablelist db error. limit: %d, err: %v",
			enumor.HuaWei, core.DefaultMaxPageLimit, err)
		return nil, err
	}

	resourceMap := make(map[string]*cloudRouteTable.RouteTable[cloudRouteTable.AzureRouteTableExtension], 0)
	if len(dbList) == 0 {
		return resourceMap, nil
	}

	for _, item := range dbList {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// compareUpdateAzureRouteTableList compare and update route table list.
func compareUpdateAzureRouteTableList(kt *kit.Kit, req *hcroutetable.AzureRouteTableSyncReq,
	list *routetable.AzureRouteTableListResult,
	resourceDBMap map[string]*cloudRouteTable.RouteTable[cloudRouteTable.AzureRouteTableExtension],
	dataCli *dataclient.Client) error {

	createResources, updateResources, subnetMap, err := filterAzureRouteTableList(kt, req, list, resourceDBMap, dataCli)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &dataproto.RouteTableBaseInfoBatchUpdateReq{
			RouteTables: updateResources,
		}
		if err = dataCli.Global.RouteTable.BatchUpdateBaseInfo(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-routetable batch compare db update failed. accountID: %s, resGroupName: %s, err: %v",
				enumor.Azure, req.AccountID, req.ResourceGroupName, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &dataproto.RouteTableBatchCreateReq[dataproto.AzureRouteTableCreateExt]{
			RouteTables: createResources,
		}
		createIDs, err := dataCli.Azure.RouteTable.BatchCreate(kt.Ctx, kt.Header(), createReq)
		if err != nil {
			logs.Errorf("%s-routetable batch compare db create failed. accountID: %s, resGroupName: %s, err: %v",
				enumor.Azure, req.AccountID, req.ResourceGroupName, err)
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

				err = BatchCreateAzureRoute(kt, newID, &item, dataCli)
				if err != nil {
					logs.Errorf("%s-routetable sync create route failed. accountID: %s, err: %v", enumor.Azure,
						req.AccountID, err)
					continue
				}
				existRouteTableIDMap[newID] = item.CloudID
				existCloudIDMap[item.CloudID] = newID
			}
		}
	}
	if len(subnetMap) > 0 {
		UpdateSubnetRouteTableByIDs(kt, enumor.Azure, subnetMap, dataCli)
	}

	return nil
}

// filterAzureRouteTableList filter azure route table list
func filterAzureRouteTableList(kt *kit.Kit, req *hcroutetable.AzureRouteTableSyncReq,
	list *routetable.AzureRouteTableListResult,
	resourceDBMap map[string]*cloudRouteTable.RouteTable[cloudRouteTable.AzureRouteTableExtension],
	dataCli *dataclient.Client) (createResources []dataproto.RouteTableCreateReq[dataproto.AzureRouteTableCreateExt],
	updateResources []dataproto.RouteTableBaseInfoUpdateReq, subnetMap map[string]dataproto.RouteTableSubnetReq,
	err error) {

	if list == nil || len(list.Details) == 0 {
		return createResources, updateResources, nil,
			fmt.Errorf("cloudapi routetablelist is empty, accountID: %s, resGroupName: %s",
				req.AccountID, req.ResourceGroupName)
	}

	subnetMap = make(map[string]dataproto.RouteTableSubnetReq, 0)
	for _, item := range list.Details {
		// need compare and update resource data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if !checkAzureIsUpdate(item, resourceInfo) {
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

			// azure route sync.
			err = AzureRouteSync(kt, req, resourceInfo.ID, &item, dataCli)
			if err != nil {
				logs.Errorf("%s-routetable sync update route failed. accountID: %s, resGroupName: %s, err: %v",
					enumor.Azure, req.AccountID, req.ResourceGroupName, err)
			}
			continue
		} else {
			// need add resource data
			tmpRes := dataproto.RouteTableCreateReq[dataproto.AzureRouteTableCreateExt]{
				AccountID:  req.AccountID,
				CloudID:    item.CloudID,
				Name:       converter.ValToPtr(item.Name),
				Region:     item.Region,
				CloudVpcID: item.CloudVpcID,
				Memo:       item.Memo,
			}
			if item.Extension != nil {
				tmpRes.Extension = &dataproto.AzureRouteTableCreateExt{
					ResourceGroupName: item.Extension.ResourceGroupName,
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

// AzureRouteSync azure route sync.
func AzureRouteSync(kt *kit.Kit, req *hcroutetable.AzureRouteTableSyncReq, routeTableID string,
	routeList *routetable.AzureRouteTable, dataCli *dataclient.Client) error {

	if routeList == nil {
		return nil
	}

	// batch get route map from db.
	resourceDBMap, err := BatchGetAzureRouteMapFromDB(kt, req, routeTableID, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable-route batch get routedblist failed. accountID: %s, routeTableID: %s, "+
			"err: %v", enumor.Azure, req.AccountID, routeTableID, err)
		return err
	}

	// batch sync vendor route list.
	err = BatchSyncAzureRoute(kt, req, routeTableID, routeList, resourceDBMap, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable-route compare api and dblist failed. accountID: %s, routeTableID: %s, "+
			"err: %v", enumor.Azure, req.AccountID, routeTableID, err)
		return err
	}

	return nil
}

// BatchGetAzureRouteMapFromDB batch get route map from db.
func BatchGetAzureRouteMapFromDB(kt *kit.Kit, req *hcroutetable.AzureRouteTableSyncReq,
	routeTableID string, dataCli *dataclient.Client) (map[string]cloudRouteTable.AzureRoute, error) {

	dbList, err := dataCli.Azure.RouteTable.ListRoute(kt.Ctx, kt.Header(), routeTableID,
		&core.ListReq{
			Filter: tools.EqualExpression("route_table_id", routeTableID),
			Page: &core.BasePage{
				Limit: core.DefaultMaxPageLimit,
			},
		})
	if err != nil {
		logs.Errorf("%s-routetable-route batch list db error. accountID: %s, rgName: %s, routeTableID: %s, "+
			"err: %v", enumor.Azure, req.AccountID, req.ResourceGroupName, routeTableID, err)
		return nil, err
	}

	resourceMap := make(map[string]cloudRouteTable.AzureRoute, 0)
	if len(dbList.Details) == 0 {
		return resourceMap, nil
	}

	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// BatchSyncAzureRoute batch sync azure route.
func BatchSyncAzureRoute(kt *kit.Kit, req *hcroutetable.AzureRouteTableSyncReq, routeTableID string,
	list *routetable.AzureRouteTable, resourceDBMap map[string]cloudRouteTable.AzureRoute,
	dataCli *dataclient.Client) error {

	createResources, updateResources, existIDMap, err := filterAzureRouteList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &dataproto.AzureRouteBatchUpdateReq{
			AzureRoutes: updateResources,
		}
		if err = dataCli.Azure.RouteTable.BatchUpdateRoute(kt.Ctx, kt.Header(), routeTableID, updateReq); err != nil {
			logs.Errorf("%s-routetable-route batch compare db update failed. accountID: %s, routeTableID: %s, err: %v",
				enumor.Azure, req.AccountID, routeTableID, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &dataproto.AzureRouteBatchCreateReq{
			AzureRoutes: createResources,
		}

		if _, err = dataCli.Azure.RouteTable.BatchCreateRoute(kt.Ctx, kt.Header(), routeTableID, createReq); err != nil {
			logs.Errorf("%s-routetable-route batch compare db create failed. accountID: %s, routeTableID: %s, err: %v",
				enumor.Azure, req.AccountID, routeTableID, err)
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
		if err = dataCli.Azure.RouteTable.BatchDeleteRoute(kt.Ctx, kt.Header(), routeTableID, deleteReq); err != nil {
			logs.Errorf("%s-routetable-route batch compare db delete failed. accountID: %s, routeTableID: %s, "+
				"delIDs: %v, err: %v", enumor.Azure, req.AccountID, routeTableID,
				deleteIDs, err)
			return err
		}
	}

	return nil
}

func filterAzureRouteList(req *hcroutetable.AzureRouteTableSyncReq,
	list *routetable.AzureRouteTable, resourceDBMap map[string]cloudRouteTable.AzureRoute) (
	createResources []dataproto.AzureRouteCreateReq, updateResources []dataproto.AzureRouteUpdateReq,
	existIDMap map[string]bool, err error) {

	if list == nil || list.Extension == nil {
		return nil, nil, nil,
			fmt.Errorf("cloudapi azureroutelist is empty, accountID: %s", req.AccountID)
	}

	existIDMap = make(map[string]bool, 0)
	for _, routeItem := range list.Extension.Routes {
		// need compare and update azure route data
		if resourceInfo, ok := resourceDBMap[routeItem.CloudID]; ok {
			if resourceInfo.AddressPrefix == routeItem.AddressPrefix &&
				resourceInfo.NextHopType == routeItem.NextHopType &&
				converter.PtrToVal(resourceInfo.NextHopIPAddress) == converter.PtrToVal(routeItem.NextHopIPAddress) &&
				resourceInfo.ProvisioningState == routeItem.ProvisioningState {
				existIDMap[resourceInfo.ID] = true
				continue
			}

			tmpRes := dataproto.AzureRouteUpdateReq{
				ID: resourceInfo.ID,
			}
			tmpRes.AddressPrefix = routeItem.AddressPrefix
			tmpRes.NextHopType = routeItem.NextHopType
			tmpRes.NextHopIPAddress = routeItem.NextHopIPAddress
			tmpRes.ProvisioningState = routeItem.ProvisioningState

			updateResources = append(updateResources, tmpRes)
			existIDMap[resourceInfo.ID] = true
		} else {
			// need add azure route data
			tmpRes := dataproto.AzureRouteCreateReq{
				CloudID:           routeItem.CloudID,
				CloudRouteTableID: routeItem.CloudRouteTableID,
				Name:              routeItem.Name,
				AddressPrefix:     routeItem.AddressPrefix,
				NextHopType:       routeItem.NextHopType,
				NextHopIPAddress:  routeItem.NextHopIPAddress,
				ProvisioningState: routeItem.ProvisioningState,
			}

			createResources = append(createResources, tmpRes)
		}
	}
	return createResources, updateResources, existIDMap, nil
}

// BatchCreateAzureRoute batch create azure route
func BatchCreateAzureRoute(kt *kit.Kit, newID string, list *routetable.AzureRouteTable,
	dataCli *dataclient.Client) error {

	if list.Extension == nil || len(list.Extension.Routes) == 0 {
		return nil
	}

	createRes := make([]dataproto.AzureRouteCreateReq, 0, len(list.Extension.Routes))
	for _, routeItem := range list.Extension.Routes {
		tmpRes := dataproto.AzureRouteCreateReq{
			CloudID:           routeItem.CloudID,
			CloudRouteTableID: routeItem.CloudRouteTableID,
			Name:              routeItem.Name,
			AddressPrefix:     routeItem.AddressPrefix,
			NextHopType:       routeItem.NextHopType,
			NextHopIPAddress:  routeItem.NextHopIPAddress,
			ProvisioningState: routeItem.ProvisioningState,
		}
		createRes = append(createRes, tmpRes)
	}

	createReq := &dataproto.AzureRouteBatchCreateReq{
		AzureRoutes: createRes,
	}
	if _, err := dataCli.Azure.RouteTable.BatchCreateRoute(kt.Ctx, kt.Header(), newID, createReq); err != nil {
		return err
	}

	return nil
}

func checkAzureIsUpdate(item routetable.AzureRouteTable,
	dbInfo *cloudRouteTable.RouteTable[cloudRouteTable.AzureRouteTableExtension]) bool {

	if dbInfo.Name != item.Name {
		return true
	}
	if !assert.IsPtrStringEqual(item.Memo, dbInfo.Memo) {
		return true
	}

	return false
}

// compareDeleteAzureRouteTableList compare and delete route table list from db.
func compareDeleteAzureRouteTableList(kt *kit.Kit, req *hcroutetable.AzureRouteTableSyncReq,
	allCloudIDMap map[string]bool, dataCli *dataclient.Client) error {

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
					Value: string(enumor.Azure),
				},
			},
		}
		dbQueryReq := &core.ListReq{
			Filter: expr,
			Page:   &core.BasePage{Count: false, Start: offset, Limit: count},
		}
		dbList, err := dataCli.Global.RouteTable.List(kt.Ctx, kt.Header(), dbQueryReq)
		if err != nil {
			logs.Errorf("%s-routetable-delete batch get vpclist db error. offset: %d, limit: %d, err: %v",
				enumor.Azure, offset, count, err)
			return err
		}

		if len(dbList.Details) == 0 {
			return nil
		}

		deleteIDs := make([]string, 0)
		for _, item := range dbList.Details {
			if _, ok := allCloudIDMap[item.CloudID]; !ok {
				deleteIDs = append(deleteIDs, item.ID)
			}
		}

		// batch query need delete route table list
		err = BatchDeleteRouteTableByIDs(kt, deleteIDs, dataCli)
		if err != nil {
			logs.Errorf("%s-routetable batch compare db delete failed. deleteIDs: %v, err: %v",
				enumor.Azure, deleteIDs, err)
			return err
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

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
	hcproto "hcm/pkg/api/hc-service/route-table"
	hcroutetable "hcm/pkg/api/hc-service/route-table"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// AwsRouteTableSync sync aws cloud route table.
func AwsRouteTableSync(kt *kit.Kit, req *hcroutetable.AwsRouteTableSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	// syncs route table list from cloudapi.
	allCloudIDMap, err := SyncAwsRouteTableList(kt, req, adaptor, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.Aws, req.AccountID, req.Region, err)
		return nil, err
	}

	// compare and delete route table idmap from db.
	err = compareDeleteAwsRouteTableList(kt, req, allCloudIDMap, adaptor, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable compare delete and dblist failed. accountID: %s, region: %s, err: %v",
			enumor.Aws, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// SyncAwsRouteTableList sync route table list from cloudapi.
func SyncAwsRouteTableList(kt *kit.Kit, req *hcproto.AwsRouteTableSyncReq, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client) (map[string]bool, error) {

	cli, err := adaptor.Aws(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	allCloudIDMap := make(map[string]bool, 0)
	for {
		opt := &routetable.AwsRouteTableListOption{
			AwsListOption: &adcore.AwsListOption{
				Region: req.Region,
				Page: &adcore.AwsPage{
					MaxResults: converter.ValToPtr(int64(adcore.AwsQueryLimit / 10)),
				},
			},
		}

		if nextToken != "" {
			opt.Page.NextToken = converter.ValToPtr(nextToken)
		}

		tmpList, tmpErr := cli.ListRouteTable(kt, opt)
		if tmpErr != nil {
			logs.Errorf("%s-routetable batch get cloudapi failed. accountID: %s, region: %s, err: %v",
				enumor.Aws, req.AccountID, req.Region, tmpErr)
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
		resourceDBMap, err := GetAwsRouteTableInfoFromDB(kt, cloudIDs, dataCli)
		if err != nil {
			logs.Errorf("%s-routetable get routetabledblist failed. accountID: %s, region: %s, err: %v",
				enumor.Aws, req.AccountID, req.Region, err)
			return allCloudIDMap, err
		}

		// compare and update route table list.
		err = compareUpdateAwsRouteTableList(kt, req, tmpList, resourceDBMap, dataCli)
		if err != nil {
			logs.Errorf("%s-routetable compare and update routetabledblist failed. accountID: %s, "+
				"region: %s, err: %v", enumor.Aws, req.AccountID, req.Region, err)
			return allCloudIDMap, err
		}

		if tmpList.NextToken == nil {
			break
		}
		nextToken = *tmpList.NextToken
	}

	return allCloudIDMap, nil
}

// GetAwsRouteTableInfoFromDB get route table info from db.
func GetAwsRouteTableInfoFromDB(kt *kit.Kit, cloudIDs []string, dataCli *dataclient.Client) (
	map[string]*cloudRouteTable.RouteTable[cloudRouteTable.AwsRouteTableExtension], error) {

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "vendor",
				Op:    filter.Equal.Factory(),
				Value: enumor.Aws,
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
	dbList, err := dataCli.Aws.RouteTable.ListRouteTableWithExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("%s-routetable batch get routetablelist db error. limit: %d, err: %v",
			enumor.Aws, core.DefaultMaxPageLimit, err)
		return nil, err
	}

	resourceMap := make(map[string]*cloudRouteTable.RouteTable[cloudRouteTable.AwsRouteTableExtension], 0)
	if len(dbList) == 0 {
		return resourceMap, nil
	}

	for _, item := range dbList {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

func checkAwsIsUpdate(item routetable.AwsRouteTable,
	resourceInfo *cloudRouteTable.RouteTable[cloudRouteTable.AwsRouteTableExtension]) bool {

	if resourceInfo.Name == item.Name && converter.PtrToVal(resourceInfo.Memo) == converter.PtrToVal(item.Memo) {
		return false
	}

	return true
}

func checkAwsRouteIsUpdate(routeItem routetable.AwsRoute, resourceInfo cloudRouteTable.AwsRoute) bool {
	if converter.PtrToVal(resourceInfo.CloudCarrierGatewayID) == converter.PtrToVal(routeItem.CloudCarrierGatewayID) &&
		converter.PtrToVal(resourceInfo.CoreNetworkArn) == converter.PtrToVal(routeItem.CoreNetworkArn) &&
		converter.PtrToVal(resourceInfo.CloudEgressOnlyInternetGatewayID) ==
			converter.PtrToVal(routeItem.CloudEgressOnlyInternetGatewayID) &&
		resourceInfo.CloudGatewayID == routeItem.CloudGatewayID &&
		resourceInfo.CloudInstanceID == routeItem.CloudInstanceID &&
		resourceInfo.CloudInstanceOwnerID == routeItem.CloudInstanceOwnerID &&
		converter.PtrToVal(resourceInfo.CloudLocalGatewayID) == converter.PtrToVal(routeItem.CloudLocalGatewayID) &&
		converter.PtrToVal(resourceInfo.CloudNatGatewayID) == converter.PtrToVal(routeItem.CloudNatGatewayID) &&
		converter.PtrToVal(resourceInfo.CloudNetworkInterfaceID) ==
			converter.PtrToVal(routeItem.CloudNetworkInterfaceID) &&
		converter.PtrToVal(resourceInfo.CloudTransitGatewayID) == converter.PtrToVal(routeItem.CloudTransitGatewayID) &&
		converter.PtrToVal(resourceInfo.CloudVpcPeeringConnectionID) ==
			converter.PtrToVal(routeItem.CloudVpcPeeringConnectionID) &&
		resourceInfo.State == routeItem.State &&
		resourceInfo.Propagated == routeItem.Propagated {
		return false
	}

	return true
}

// compareUpdateAwsRouteTableList compare and update route table list.
func compareUpdateAwsRouteTableList(kt *kit.Kit, req *hcproto.AwsRouteTableSyncReq,
	list *routetable.AwsRouteTableListResult,
	resourceDBMap map[string]*cloudRouteTable.RouteTable[cloudRouteTable.AwsRouteTableExtension],
	dataCli *dataclient.Client) error {

	createResources, updateResources, err := filterAwsRouteTableList(kt, req, list, resourceDBMap, dataCli)
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
				enumor.Aws, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &dataproto.RouteTableBatchCreateReq[dataproto.AwsRouteTableCreateExt]{
			RouteTables: createResources,
		}
		createIDs, err := dataCli.Aws.RouteTable.BatchCreate(kt.Ctx, kt.Header(), createReq)
		if err != nil {
			logs.Errorf("%s-routetable batch compare db create failed. accountID: %s, region: %s, err: %v",
				enumor.Aws, req.AccountID, req.Region, err)
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

				err = BatchCreateAwsRoute(kt, newID, &item, dataCli)
				if err != nil {
					logs.Errorf("%s-routetable-create sync create route failed. accountID: %s, region: %s, "+
						"err: %v", enumor.Aws, req.AccountID, req.Region, err)
					continue
				}
				existRouteTableIDMap[newID] = item.CloudID
				existCloudIDMap[item.CloudID] = newID
			}
		}
	}

	return nil
}

// filterAwsRouteTableList filter aws route table list
func filterAwsRouteTableList(kt *kit.Kit, req *hcproto.AwsRouteTableSyncReq,
	list *routetable.AwsRouteTableListResult,
	resourceDBMap map[string]*cloudRouteTable.RouteTable[cloudRouteTable.AwsRouteTableExtension],
	dataCli *dataclient.Client) (createResources []dataproto.RouteTableCreateReq[dataproto.AwsRouteTableCreateExt],
	updateResources []dataproto.RouteTableBaseInfoUpdateReq, err error) {

	if list == nil || len(list.Details) == 0 {
		return createResources, updateResources,
			fmt.Errorf("cloudapi routetablelist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	for _, item := range list.Details {
		// need compare and update resource data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if !checkAwsIsUpdate(item, resourceInfo) {
				continue
			}

			tmpRes := dataproto.RouteTableBaseInfoUpdateReq{
				IDs: []string{resourceInfo.ID},
			}
			tmpRes.Data = &dataproto.RouteTableUpdateBaseInfo{
				Name: converter.ValToPtr(item.Name),
				Memo: item.Memo,
			}

			updateResources = append(updateResources, tmpRes)

			// aws route sync.
			err = AwsRouteSync(kt, req, resourceInfo.ID, &item, dataCli)
			if err != nil {
				logs.Errorf("%s-routetable sync update route failed. accountID: %s, region: %s, err: %v",
					enumor.Aws, req.AccountID, req.Region, err)
			}
			continue
		} else {
			// need add resource data
			tmpRes := dataproto.RouteTableCreateReq[dataproto.AwsRouteTableCreateExt]{
				AccountID:  req.AccountID,
				CloudID:    item.CloudID,
				Name:       converter.ValToPtr(item.Name),
				Region:     item.Region,
				CloudVpcID: item.CloudVpcID,
				Memo:       item.Memo,
			}
			if item.Extension != nil {
				tmpRes.Extension = &dataproto.AwsRouteTableCreateExt{
					Main: item.Extension.Main,
				}
			}

			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, nil
}

// GetNeedDeleteAwsRouteTableList get need delete aws route table list
func GetNeedDeleteAwsRouteTableList(kt *kit.Kit, req *hcproto.AwsRouteTableSyncReq,
	deleteCloudIDMap map[string]string, adaptor *cloudclient.CloudAdaptorClient) []string {

	deleteIDs := make([]string, 0, len(deleteCloudIDMap))
	if len(deleteCloudIDMap) == 0 {
		return deleteIDs
	}

	cli, err := adaptor.Aws(kt, req.AccountID)
	if err != nil {
		logs.Errorf("%s-routetable get account failed. accountID: %s, region: %s, err: %v",
			enumor.Aws, req.AccountID, req.Region, err)
		return deleteIDs
	}

	var tmpResourceIDs []string
	for tmpCloudID, tmpID := range deleteCloudIDMap {
		tmpResourceIDs = append(tmpResourceIDs, tmpCloudID)
		deleteIDs = append(deleteIDs, tmpID)
	}

	opt := &routetable.AwsRouteTableListOption{
		AwsListOption: &adcore.AwsListOption{
			Region:   req.Region,
			CloudIDs: tmpResourceIDs,
		},
	}

	tmpList, tmpErr := cli.ListRouteTable(kt, opt)
	if tmpErr != nil {
		logs.Errorf("%s-routetable batch get cloudapi failed. accountID: %s, region: %s, err: %v",
			enumor.Aws, req.AccountID, req.Region, tmpErr)
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

// AwsRouteSync aws route sync.
func AwsRouteSync(kt *kit.Kit, req *hcproto.AwsRouteTableSyncReq, routeTableID string,
	routeList *routetable.AwsRouteTable, dataCli *dataclient.Client) error {

	// batch get route map from db.
	resourceDBMap, err := BatchGetAwsRouteMapFromDB(kt, req, routeTableID, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable-route batch get routedblist failed. accountID: %s, routeTableID: %s, err: %v",
			enumor.Aws, req.AccountID, routeTableID, err)
		return err
	}

	// batch sync vendor route list.
	err = BatchSyncAwsRoute(kt, req, routeTableID, routeList, resourceDBMap, dataCli)
	if err != nil {
		logs.Errorf("%s-routetable-route compare api and dblist failed. accountID: %s, routeTableID: %s, "+
			"err: %v", enumor.Aws, req.AccountID, routeTableID, err)
		return err
	}

	return nil
}

// BatchGetAwsRouteMapFromDB batch get route map from db.
func BatchGetAwsRouteMapFromDB(kt *kit.Kit, req *hcproto.AwsRouteTableSyncReq,
	routeTableID string, dataCli *dataclient.Client) (map[string]cloudRouteTable.AwsRoute, error) {

	dbList, err := dataCli.Aws.RouteTable.ListRoute(kt.Ctx, kt.Header(), routeTableID,
		&core.ListReq{
			Filter: tools.EqualExpression("route_table_id", routeTableID),
			Page: &core.BasePage{
				Limit: core.DefaultMaxPageLimit,
			},
		})
	if err != nil {
		logs.Errorf("%s-routetable-route batch list db error. accountID: %s, region: %s, err: %v",
			enumor.Aws, req.AccountID, req.Region, err)
		return nil, err
	}

	resourceMap := make(map[string]cloudRouteTable.AwsRoute, 0)
	if len(dbList.Details) == 0 {
		return resourceMap, nil
	}

	for _, item := range dbList.Details {
		resourceMap[item.CloudRouteTableID] = item
	}

	return resourceMap, nil
}

// BatchSyncAwsRoute batch sync aws route.
func BatchSyncAwsRoute(kt *kit.Kit, req *hcproto.AwsRouteTableSyncReq, routeTableID string,
	list *routetable.AwsRouteTable, resourceDBMap map[string]cloudRouteTable.AwsRoute,
	dataCli *dataclient.Client) error {

	createResources, updateResources, existIDMap, err := filterAwsRouteList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &dataproto.AwsRouteBatchUpdateReq{
			AwsRoutes: updateResources,
		}
		if err = dataCli.Aws.RouteTable.BatchUpdateRoute(kt.Ctx, kt.Header(), routeTableID, updateReq); err != nil {
			logs.Errorf("%s-routetable-route batch compare db update failed. accountID: %s, region: %s, "+
				"routeTableID: %s, err: %v", enumor.Aws, req.AccountID, req.Region, routeTableID, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &dataproto.AwsRouteBatchCreateReq{
			AwsRoutes: createResources,
		}
		if _, err = dataCli.Aws.RouteTable.BatchCreateRoute(kt.Ctx, kt.Header(), routeTableID, createReq); err != nil {
			logs.Errorf("%s-routetable-route batch compare db create failed. accountID: %s, region: %s, "+
				"routeTableID: %s, err: %v", enumor.Aws, req.AccountID, req.Region, routeTableID, err)
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
		if err = dataCli.Aws.RouteTable.BatchDeleteRoute(kt.Ctx, kt.Header(), routeTableID, deleteReq); err != nil {
			logs.Errorf("%s-routetable-route batch compare db delete failed. accountID: %s, region: %s, "+
				"routeTableID: %s, delIDs: %v, err: %v", enumor.Aws, req.AccountID, req.Region, routeTableID,
				deleteIDs, err)
			return err
		}
	}

	return nil
}

func filterAwsRouteList(req *hcproto.AwsRouteTableSyncReq, list *routetable.AwsRouteTable,
	resourceDBMap map[string]cloudRouteTable.AwsRoute) (createResources []dataproto.AwsRouteCreateReq,
	updateResources []dataproto.AwsRouteUpdateReq, existIDMap map[string]bool, err error) {
	if list == nil || list.Extension == nil {
		return nil, nil, nil,
			fmt.Errorf("cloudapi awsroutelist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	existIDMap = make(map[string]bool, 0)
	for _, routeItem := range list.Extension.Routes {
		// need compare and update aws route data
		if resourceInfo, ok := resourceDBMap[routeItem.CloudRouteTableID]; ok {
			if !checkAwsRouteIsUpdate(routeItem, resourceInfo) {
				existIDMap[resourceInfo.ID] = true
				continue
			}

			tmpRes := dataproto.AwsRouteUpdateReq{
				ID: resourceInfo.ID,
			}
			tmpRes.CloudCarrierGatewayID = routeItem.CloudCarrierGatewayID
			tmpRes.CoreNetworkArn = routeItem.CoreNetworkArn
			tmpRes.CloudEgressOnlyInternetGatewayID = routeItem.CloudEgressOnlyInternetGatewayID
			tmpRes.CloudGatewayID = routeItem.CloudGatewayID
			tmpRes.CloudInstanceID = routeItem.CloudInstanceID
			tmpRes.CloudInstanceOwnerID = routeItem.CloudInstanceOwnerID
			tmpRes.CloudLocalGatewayID = routeItem.CloudLocalGatewayID
			tmpRes.CloudNatGatewayID = routeItem.CloudNatGatewayID
			tmpRes.CloudNetworkInterfaceID = routeItem.CloudNetworkInterfaceID
			tmpRes.CloudTransitGatewayID = routeItem.CloudTransitGatewayID
			tmpRes.CloudVpcPeeringConnectionID = routeItem.CloudVpcPeeringConnectionID
			tmpRes.State = routeItem.State
			tmpRes.Propagated = converter.ValToPtr(routeItem.Propagated)
			updateResources = append(updateResources, tmpRes)

			existIDMap[resourceInfo.ID] = true
		} else {
			// need add aws route data
			tmpRes := dataproto.AwsRouteCreateReq{
				CloudRouteTableID:                routeItem.CloudRouteTableID,
				DestinationCidrBlock:             routeItem.DestinationCidrBlock,
				DestinationIpv6CidrBlock:         routeItem.DestinationIpv6CidrBlock,
				CloudDestinationPrefixListID:     routeItem.CloudDestinationPrefixListID,
				CloudCarrierGatewayID:            routeItem.CloudCarrierGatewayID,
				CoreNetworkArn:                   routeItem.CoreNetworkArn,
				CloudEgressOnlyInternetGatewayID: routeItem.CloudEgressOnlyInternetGatewayID,
				CloudGatewayID:                   routeItem.CloudGatewayID,
				CloudInstanceID:                  routeItem.CloudInstanceID,
				CloudInstanceOwnerID:             routeItem.CloudInstanceOwnerID,
				CloudLocalGatewayID:              routeItem.CloudLocalGatewayID,
				CloudNatGatewayID:                routeItem.CloudNatGatewayID,
				CloudNetworkInterfaceID:          routeItem.CloudNetworkInterfaceID,
				CloudTransitGatewayID:            routeItem.CloudTransitGatewayID,
				CloudVpcPeeringConnectionID:      routeItem.CloudVpcPeeringConnectionID,
				State:                            routeItem.State,
				Propagated:                       routeItem.Propagated,
			}
			createResources = append(createResources, tmpRes)
		}
	}
	return createResources, updateResources, existIDMap, nil
}

// BatchCreateAwsRoute batch create aws route
func BatchCreateAwsRoute(kt *kit.Kit, newID string, list *routetable.AwsRouteTable, dataCli *dataclient.Client) error {
	if list.Extension == nil || len(list.Extension.Routes) == 0 {
		return nil
	}

	createRes := make([]dataproto.AwsRouteCreateReq, 0, len(list.Extension.Routes))
	for _, routeItem := range list.Extension.Routes {
		tmpRes := dataproto.AwsRouteCreateReq{
			CloudRouteTableID:                routeItem.CloudRouteTableID,
			DestinationCidrBlock:             routeItem.DestinationCidrBlock,
			DestinationIpv6CidrBlock:         routeItem.DestinationIpv6CidrBlock,
			CloudDestinationPrefixListID:     routeItem.CloudDestinationPrefixListID,
			CloudCarrierGatewayID:            routeItem.CloudCarrierGatewayID,
			CoreNetworkArn:                   routeItem.CoreNetworkArn,
			CloudEgressOnlyInternetGatewayID: routeItem.CloudEgressOnlyInternetGatewayID,
			CloudGatewayID:                   routeItem.CloudGatewayID,
			CloudInstanceID:                  routeItem.CloudInstanceID,
			CloudInstanceOwnerID:             routeItem.CloudInstanceOwnerID,
			CloudLocalGatewayID:              routeItem.CloudLocalGatewayID,
			CloudNatGatewayID:                routeItem.CloudNatGatewayID,
			CloudNetworkInterfaceID:          routeItem.CloudNetworkInterfaceID,
			CloudTransitGatewayID:            routeItem.CloudTransitGatewayID,
			CloudVpcPeeringConnectionID:      routeItem.CloudVpcPeeringConnectionID,
			State:                            routeItem.State,
			Propagated:                       routeItem.Propagated,
		}
		createRes = append(createRes, tmpRes)
	}

	createReq := &dataproto.AwsRouteBatchCreateReq{
		AwsRoutes: createRes,
	}

	if _, err := dataCli.Aws.RouteTable.BatchCreateRoute(kt.Ctx, kt.Header(), newID, createReq); err != nil {
		return err
	}

	return nil
}

// compareDeleteAwsRouteTableList compare and delete route table list from db.
func compareDeleteAwsRouteTableList(kt *kit.Kit, req *hcroutetable.AwsRouteTableSyncReq, allCloudIDMap map[string]bool,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) error {

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
					Value: string(enumor.Aws),
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
				"err: %v", enumor.Aws, offset, count, err)
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
		deleteIDs := GetNeedDeleteAwsRouteTableList(kt, req, deleteCloudIDMap, adaptor)
		if len(deleteIDs) > 0 {
			err = BatchDeleteRouteTableByIDs(kt, deleteIDs, dataCli)
			if err != nil {
				logs.Errorf("%s-routetable batch compare db delete failed. deleteIDs: %v, err: %v",
					enumor.Aws, deleteIDs, err)
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

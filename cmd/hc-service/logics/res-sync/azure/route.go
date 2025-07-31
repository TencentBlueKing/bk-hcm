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

package azure

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	adcore "hcm/pkg/adaptor/types/core"
	typesroutetable "hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/api/core"
	routetable "hcm/pkg/api/core/cloud/route-table"
	dataservice "hcm/pkg/api/data-service"
	dataproto "hcm/pkg/api/data-service/cloud/route-table"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/concurrence"
	"hcm/pkg/tools/converter"
)

// SyncRouteOption ...
type SyncRouteOption struct {
}

// Validate ...
func (opt SyncRouteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Route ...
func (cli *client) Route(kt *kit.Kit, params *SyncBaseParams, opt *SyncRouteOption) (*SyncResult, error) {

	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := concurrence.BaseExec(constant.SyncConcurrencyDefaultMaxLimit, params.CloudIDs, func(param string) error {
		syncOpt := &syncRouteOption{
			AccountID:         params.AccountID,
			ResourceGroupName: params.ResourceGroupName,
			CloudRouteTableID: param,
		}
		if _, err := cli.route(kt, syncOpt); err != nil {
			logs.ErrorDepthf(1, "[%s] account: %s route_table: %s sync route failed, err: %v, rid: %s",
				enumor.Azure, params.AccountID, param, err, kt.Rid)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return new(SyncResult), nil
}

type syncRouteOption struct {
	AccountID         string `json:"account_id" validate:"required"`
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
	CloudRouteTableID string `json:"cloud_route_table_id" validate:"required"`
}

// Validate ...
func (opt syncRouteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

func (cli *client) route(kt *kit.Kit, opt *syncRouteOption) (*SyncResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	params := &SyncBaseParams{
		AccountID:         opt.AccountID,
		ResourceGroupName: opt.ResourceGroupName,
		CloudIDs:          []string{opt.CloudRouteTableID},
	}
	rts, err := cli.listRouteTableFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(rts) == 0 {
		return nil, fmt.Errorf("route table: %s not found from db", opt.CloudRouteTableID)
	}
	routeTable := rts[0]

	routeFromDB, err := cli.listRouteFromDB(kt, opt, routeTable.ID)
	if err != nil {
		return nil, err
	}

	routeFromCloud, err := cli.listRouteFromCloud(kt, opt)
	if err != nil {
		return nil, err
	}

	if len(routeFromCloud) == 0 && len(routeFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typesroutetable.AzureRoute,
		routetable.AzureRoute](routeFromCloud, routeFromDB, isRouteChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteRoute(kt, opt.AccountID, opt.ResourceGroupName, opt.CloudRouteTableID, routeTable.ID,
			delCloudIDs); err != nil {

			return nil, err
		}
	}

	if len(addSlice) > 0 {
		err := cli.createRoute(kt, opt.AccountID, opt.ResourceGroupName, routeTable.ID, addSlice)
		if err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		err := cli.updateRoute(kt, opt.AccountID, opt.ResourceGroupName, routeTable.ID, updateMap)
		if err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

// createRoute create route in db
func (cli *client) createRoute(kt *kit.Kit, accountID string, resGroupName string, routeTableID string,
	addSlice []typesroutetable.AzureRoute) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("route addSlice is <= 0, not create")
	}

	createResources := make([]dataproto.AzureRouteCreateReq, 0, len(addSlice))
	for _, one := range addSlice {
		tmpRes := dataproto.AzureRouteCreateReq{
			CloudID:           one.CloudID,
			CloudRouteTableID: one.CloudRouteTableID,
			Name:              one.Name,
			AddressPrefix:     one.AddressPrefix,
			NextHopType:       one.NextHopType,
			NextHopIPAddress:  one.NextHopIPAddress,
			ProvisioningState: one.ProvisioningState,
		}
		createResources = append(createResources, tmpRes)
	}

	createReq := &dataproto.AzureRouteBatchCreateReq{
		AzureRoutes: createResources,
	}
	if _, err := cli.dbCli.Azure.RouteTable.BatchCreateRoute(kt.Ctx, kt.Header(),
		routeTableID, createReq); err != nil {
		logs.Errorf("[%s] batch create route failed. err: %v, accountID: %s, resGroupName: %s, "+
			"routeTableID: %s, rid: %s", enumor.Azure, err, accountID, resGroupName, routeTableID, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync route to create route success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(addSlice), kt.Rid)

	return nil
}

// updateRoute update route in db
func (cli *client) updateRoute(kt *kit.Kit, accountID, resGroupName, routeTableID string,
	updateMap map[string]typesroutetable.AzureRoute) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("route updateMap is <= 0, not update")
	}

	updateResources := make([]dataproto.AzureRouteUpdateReq, 0, len(updateMap))
	for id, one := range updateMap {
		tmpRes := dataproto.AzureRouteUpdateReq{
			ID: id,
		}
		tmpRes.AddressPrefix = one.AddressPrefix
		tmpRes.NextHopType = one.NextHopType
		tmpRes.NextHopIPAddress = one.NextHopIPAddress
		tmpRes.ProvisioningState = one.ProvisioningState

		updateResources = append(updateResources, tmpRes)
	}

	updateReq := &dataproto.AzureRouteBatchUpdateReq{
		AzureRoutes: updateResources,
	}
	if err := cli.dbCli.Azure.RouteTable.BatchUpdateRoute(kt.Ctx, kt.Header(), routeTableID,
		updateReq); err != nil {
		logs.Errorf("[%s] batch update route failed. err: %v, accountID: %s, resGroupName: %s, "+
			"routeTableID: %s, rid: %s", enumor.Azure, err, accountID, resGroupName, routeTableID, kt.Rid)
		return err
	}
	logs.Infof("[%s] sync route to update route success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(updateMap), kt.Rid)

	return nil
}

// deleteRoute delete route in db
func (cli *client) deleteRoute(kt *kit.Kit, accountID, resGroupName, cloudRTID, rtID string,
	delCloudIDs []string) error {

	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("route delCloudIDs is <= 0, not delete")
	}

	checkParams := &syncRouteOption{
		AccountID:         accountID,
		ResourceGroupName: resGroupName,
		CloudRouteTableID: cloudRTID,
	}
	delRouteFromCloud, err := cli.listRouteFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	cloudIDMap := converter.StringSliceToMap(delCloudIDs)
	for _, route := range delRouteFromCloud {
		if _, exist := cloudIDMap[route.GetCloudID()]; exist {
			logs.Errorf("[%s] validate route not exist failed, before delete, opt: %v, cloud_id: %s, rid: %s",
				enumor.Azure, checkParams, route.GetCloudID(), kt.Rid)
			return fmt.Errorf("validate route not exist failed, before delete")
		}
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Azure.RouteTable.BatchDeleteRoute(kt.Ctx, kt.Header(), rtID, deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete route failed, err: %v, rid: %s", enumor.Azure,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync route to delete route success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

// listRouteTableFromCloud list route table from cloud
func (cli *client) listRouteFromCloud(kt *kit.Kit, opt *syncRouteOption) ([]typesroutetable.AzureRoute, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	routeOpt := &adcore.AzureListByIDOption{
		ResourceGroupName: opt.ResourceGroupName,
		CloudIDs:          []string{opt.CloudRouteTableID},
	}
	routeTables, err := cli.cloudCli.ListRouteTableByID(kt, routeOpt)
	if err != nil {
		logs.Errorf("[%s] list routeTable from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure, err,
			opt.AccountID, opt, kt.Rid)
		return nil, err
	}

	results := make([]typesroutetable.AzureRoute, 0)
	for _, routeTable := range routeTables.Details {
		results = append(results, routeTable.Extension.Routes...)
	}

	return results, nil
}

// listRouteTableFromDB list route table from db
func (cli *client) listRouteFromDB(kt *kit.Kit, opt *syncRouteOption, rtID string) ([]routetable.AzureRoute, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: tools.EqualExpression("cloud_route_table_id", opt.CloudRouteTableID),
		Page: &core.BasePage{
			Limit: core.DefaultMaxPageLimit,
		},
	}
	results, err := cli.dbCli.Azure.RouteTable.ListRoute(kt.Ctx, kt.Header(), rtID, req)
	if err != nil {
		logs.Errorf("[%s] batch list route failed. err: %v, accountID: %s, resGroupName: %s, "+
			"routeTableID: %s, rid: %s", enumor.Azure, err, opt.AccountID, opt.ResourceGroupName, rtID, kt.Rid)
		return nil, err
	}

	routes := make([]routetable.AzureRoute, 0)
	for _, item := range results.Details {
		routes = append(routes, item)
	}

	return routes, nil
}

// isRouteChange checks if the route has changed between cloud and db
func isRouteChange(cloud typesroutetable.AzureRoute,
	db routetable.AzureRoute) bool {

	if cloud.AddressPrefix != db.AddressPrefix {
		return true
	}

	if cloud.NextHopType != db.NextHopType {
		return true
	}

	if converter.PtrToVal(cloud.NextHopIPAddress) != converter.PtrToVal(db.NextHopIPAddress) {
		return true
	}

	if cloud.ProvisioningState != db.ProvisioningState {
		return true
	}

	return false
}

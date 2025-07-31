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

package aws

import (
	"fmt"
	"strings"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/adaptor/aws"
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
	"hcm/pkg/tools/assert"
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
			Region:            params.Region,
			CloudRouteTableID: param,
		}
		if _, err := cli.route(kt, syncOpt); err != nil {
			logs.ErrorDepthf(1, "[%s] account: %s route_table: %s sync route failed, err: %v, rid: %s",
				enumor.Aws, params.AccountID, param, err, kt.Rid)
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
	Region            string `json:"region" validate:"required"`
	CloudRouteTableID string `json:"cloud_route_table_id" validate:"required"`
}

// Validate ...
func (opt syncRouteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// route 同步路由表
func (cli *client) route(kt *kit.Kit, opt *syncRouteOption) (*SyncResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	params := &SyncBaseParams{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		CloudIDs:  []string{opt.CloudRouteTableID},
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

	addSlice, updateMap, delCloudIDs := common.Diff[typesroutetable.AwsRoute,
		routetable.AwsRoute](routeFromCloud, routeFromDB, isRouteChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteRoute(kt, opt.AccountID, opt.Region, opt.CloudRouteTableID, routeTable.ID,
			delCloudIDs, routeFromDB); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		err := cli.createRoute(kt, opt.AccountID, opt.Region, routeTable.ID, addSlice)
		if err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		err := cli.updateRoute(kt, opt.AccountID, opt.Region, routeTable.ID, updateMap)
		if err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

// createRoute 批量创建路由
func (cli *client) createRoute(kt *kit.Kit, accountID string, region string, routeTableID string,
	addSlice []typesroutetable.AwsRoute) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("route addSlice is <= 0, not create")
	}

	createResources := make([]dataproto.AwsRouteCreateReq, 0, len(addSlice))
	for _, one := range addSlice {
		tmpRes := dataproto.AwsRouteCreateReq{
			CloudRouteTableID:                one.CloudRouteTableID,
			DestinationCidrBlock:             one.DestinationCidrBlock,
			DestinationIpv6CidrBlock:         one.DestinationIpv6CidrBlock,
			CloudDestinationPrefixListID:     one.CloudDestinationPrefixListID,
			CloudCarrierGatewayID:            one.CloudCarrierGatewayID,
			CoreNetworkArn:                   one.CoreNetworkArn,
			CloudEgressOnlyInternetGatewayID: one.CloudEgressOnlyInternetGatewayID,
			CloudGatewayID:                   one.CloudGatewayID,
			CloudInstanceID:                  one.CloudInstanceID,
			CloudInstanceOwnerID:             one.CloudInstanceOwnerID,
			CloudLocalGatewayID:              one.CloudLocalGatewayID,
			CloudNatGatewayID:                one.CloudNatGatewayID,
			CloudNetworkInterfaceID:          one.CloudNetworkInterfaceID,
			CloudTransitGatewayID:            one.CloudTransitGatewayID,
			CloudVpcPeeringConnectionID:      one.CloudVpcPeeringConnectionID,
			State:                            one.State,
			Propagated:                       one.Propagated,
		}
		createResources = append(createResources, tmpRes)
	}

	createReq := &dataproto.AwsRouteBatchCreateReq{
		AwsRoutes: createResources,
	}
	if _, err := cli.dbCli.Aws.RouteTable.BatchCreateRoute(kt.Ctx, kt.Header(),
		routeTableID, createReq); err != nil {
		logs.Errorf("[%s] batch create route failed. err: %v, accountID: %s, region: %s, "+
			"routeTableID: %s, rid: %s", enumor.Aws, err, accountID, region, routeTableID, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync route to create route success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(addSlice), kt.Rid)

	return nil
}

// updateRoute 批量更新路由
func (cli *client) updateRoute(kt *kit.Kit, accountID, region, routeTableID string,
	updateMap map[string]typesroutetable.AwsRoute) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("route updateMap is <= 0, not update")
	}

	updateResources := make([]dataproto.AwsRouteUpdateReq, 0, len(updateMap))
	for id, one := range updateMap {
		tmpRes := dataproto.AwsRouteUpdateReq{
			ID: id,
		}
		tmpRes.CloudCarrierGatewayID = one.CloudCarrierGatewayID
		tmpRes.CoreNetworkArn = one.CoreNetworkArn
		tmpRes.CloudEgressOnlyInternetGatewayID = one.CloudEgressOnlyInternetGatewayID
		tmpRes.CloudGatewayID = one.CloudGatewayID
		tmpRes.CloudInstanceID = one.CloudInstanceID
		tmpRes.CloudInstanceOwnerID = one.CloudInstanceOwnerID
		tmpRes.CloudLocalGatewayID = one.CloudLocalGatewayID
		tmpRes.CloudNatGatewayID = one.CloudNatGatewayID
		tmpRes.CloudNetworkInterfaceID = one.CloudNetworkInterfaceID
		tmpRes.CloudTransitGatewayID = one.CloudTransitGatewayID
		tmpRes.CloudVpcPeeringConnectionID = one.CloudVpcPeeringConnectionID
		tmpRes.State = one.State
		tmpRes.Propagated = converter.ValToPtr(one.Propagated)

		updateResources = append(updateResources, tmpRes)
	}

	updateReq := &dataproto.AwsRouteBatchUpdateReq{
		AwsRoutes: updateResources,
	}
	if err := cli.dbCli.Aws.RouteTable.BatchUpdateRoute(kt.Ctx, kt.Header(), routeTableID,
		updateReq); err != nil {
		logs.Errorf("[%s] batch update route failed. err: %v, accountID: %s, region: %s, "+
			"routeTableID: %s, rid: %s", enumor.Aws, err, accountID, region, routeTableID, kt.Rid)
		return err
	}
	logs.Infof("[%s] sync route to update route success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(updateMap), kt.Rid)

	return nil
}

// deleteRoute 批量删除路由
func (cli *client) deleteRoute(kt *kit.Kit, accountID, region, cloudRTID, rtID string,
	delCloudIDs []string, routeFromDB []routetable.AwsRoute) error {

	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("route delCloudIDs is <= 0, not delete")
	}

	checkParams := &syncRouteOption{
		AccountID:         accountID,
		Region:            region,
		CloudRouteTableID: cloudRTID,
	}
	delRouteFromCloud, err := cli.listRouteFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	cloudIDMap := converter.StringSliceToMap(delCloudIDs)
	for _, route := range delRouteFromCloud {
		if _, exist := cloudIDMap[route.GetCloudID()]; exist {
			logs.Errorf("[%s] validate route not exist failed, before delete, opt: %v, cloudID: %s, rid: %s",
				enumor.Aws, checkParams, route.GetCloudID(), kt.Rid)
			return fmt.Errorf("validate route not exist failed, before delete")
		}
	}

	deleteIDs := make([]string, 0)
	for _, one := range routeFromDB {
		if _, exist := cloudIDMap[one.GetCloudID()]; exist {
			deleteIDs = append(deleteIDs, one.ID)
		}
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.ContainersExpression("id", deleteIDs),
	}
	if err := cli.dbCli.Aws.RouteTable.BatchDeleteRoute(kt.Ctx, kt.Header(), rtID, deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete route failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync route to delete route success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

// listRouteTableFromDB 从数据库中获取路由表
func (cli *client) listRouteFromCloud(kt *kit.Kit, opt *syncRouteOption) ([]typesroutetable.AwsRoute, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	routeOpt := &typesroutetable.AwsRouteTableListOption{
		AwsListOption: &adcore.AwsListOption{
			Region:   opt.Region,
			CloudIDs: []string{opt.CloudRouteTableID},
		},
	}
	routeTables, err := cli.cloudCli.ListRouteTable(kt, routeOpt)
	if err != nil {
		if strings.Contains(err.Error(), aws.ErrRouteTableNotFound) {
			return nil, nil
		}

		logs.Errorf("[%s] list routeTable from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Aws, err,
			opt.AccountID, opt, kt.Rid)
		return nil, err
	}

	results := make([]typesroutetable.AwsRoute, 0)
	for _, routeTable := range routeTables.Details {
		results = append(results, routeTable.Extension.Routes...)
	}

	return results, nil
}

// listRouteTableFromDB 从数据库中获取路由表
func (cli *client) listRouteFromDB(kt *kit.Kit, opt *syncRouteOption, rtID string) ([]routetable.AwsRoute, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: tools.EqualExpression("cloud_route_table_id", opt.CloudRouteTableID),
		Page: &core.BasePage{
			Limit: core.DefaultMaxPageLimit,
		},
	}
	results, err := cli.dbCli.Aws.RouteTable.ListRoute(kt.Ctx, kt.Header(), rtID, req)
	if err != nil {
		logs.Errorf("[%s] batch list route failed. err: %v, accountID: %s, region: %s, "+
			"routeTableID: %s, rid: %s", enumor.Aws, err, opt.AccountID, opt.Region, rtID, kt.Rid)
		return nil, err
	}

	routes := make([]routetable.AwsRoute, 0)
	for _, item := range results.Details {
		routes = append(routes, item)
	}

	return routes, nil
}

// isRouteChange 判断云端路由和数据库路由是否有变化
func isRouteChange(cloud typesroutetable.AwsRoute,
	db routetable.AwsRoute) bool {

	if !assert.IsPtrStringEqual(cloud.CloudCarrierGatewayID, db.CloudCarrierGatewayID) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.CoreNetworkArn, db.CoreNetworkArn) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.CloudEgressOnlyInternetGatewayID, db.CloudEgressOnlyInternetGatewayID) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.CloudGatewayID, db.CloudGatewayID) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.CloudInstanceID, db.CloudInstanceID) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.CloudInstanceOwnerID, db.CloudInstanceOwnerID) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.CloudLocalGatewayID, db.CloudLocalGatewayID) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.CloudNatGatewayID, db.CloudNatGatewayID) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.CloudNetworkInterfaceID, db.CloudNetworkInterfaceID) {
		return true
	}
	if !assert.IsPtrStringEqual(cloud.CloudTransitGatewayID, db.CloudTransitGatewayID) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.CloudVpcPeeringConnectionID, db.CloudVpcPeeringConnectionID) {
		return true
	}

	if cloud.State != db.State {
		return true
	}

	if cloud.Propagated != db.Propagated {
		return true
	}

	return false
}

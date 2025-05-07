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

package huawei

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
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
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
)

// SyncRouteTableOption ...
type SyncRouteTableOption struct {
}

// Validate ...
func (opt SyncRouteTableOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// RouteTable ...
func (cli *client) RouteTable(kt *kit.Kit, params *SyncBaseParams, opt *SyncRouteTableOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	routeTableFromCloud, err := cli.listRouteTableFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	routeTableFromDB, err := cli.listRouteTableFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(routeTableFromCloud) == 0 && len(routeTableFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typesroutetable.HuaWeiRouteTable,
		routetable.HuaWeiRouteTable](routeTableFromCloud, routeTableFromDB, isRouteTableChange)

	subnetMap := make(map[string]dataproto.RouteTableSubnetReq, 0)

	if len(delCloudIDs) > 0 {
		err = common.CancelRouteTableSubnetRel(kt, cli.dbCli, enumor.HuaWei, delCloudIDs)
		if err != nil {
			logs.Errorf("[%s] routetable batch cancel subnet rel failed. deleteIDs: %v, err: %v, rid: %s",
				enumor.HuaWei, delCloudIDs, err, kt.Rid)
			return nil, err
		}
		if err = cli.deleteRouteTable(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
			logs.Errorf("delete huawei routeTable failed, err: %v, account: %s, region: %s, delCloudIDs: %v, rid: %s",
				err, params.AccountID, params.Region, delCloudIDs, kt.Rid)
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		addSubnetMap, err := cli.createRouteTable(kt, params.AccountID, params.Region, addSlice)
		if err != nil {
			return nil, err
		}
		for k, v := range addSubnetMap {
			subnetMap[k] = v
		}
	}

	if len(updateMap) > 0 {
		updateSubnetMap, err := cli.updateRouteTalbe(kt, params.AccountID, params.Region, updateMap)
		if err != nil {
			return nil, err
		}
		for k, v := range updateSubnetMap {
			subnetMap[k] = v
		}
	}

	// 更新子网的路由表信息
	if len(subnetMap) > 0 {
		err = common.UpdateSubnetRouteTableByIDs(kt, enumor.HuaWei, subnetMap, cli.dbCli)
		if err != nil {
			logs.Errorf("[%s] routetable update subnet's route_table failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, params.AccountID, params.Region, err)
			return nil, err
		}
	}

	// 同步db中路由表的路由规则
	if err = cli.syncRoute(kt, params); err != nil {
		return nil, err
	}

	return new(SyncResult), nil
}

func (cli *client) syncRoute(kt *kit.Kit, params *SyncBaseParams) error {
	existRT, err := cli.listRouteTableFromDB(kt, params)
	if err != nil {
		return err
	}

	if len(existRT) != 0 {
		rtCloudIDs := make([]string, 0, len(existRT))
		for _, table := range existRT {
			rtCloudIDs = append(rtCloudIDs, table.CloudID)
		}

		ruleParams := &SyncBaseParams{
			AccountID: params.AccountID,
			Region:    params.Region,
			CloudIDs:  rtCloudIDs,
		}
		if _, err = cli.Route(kt, ruleParams, new(SyncRouteOption)); err != nil {
			return err
		}
	}

	return nil
}

func (cli *client) createRouteTable(kt *kit.Kit, accountID string, resGroupName string,
	addSlice []typesroutetable.HuaWeiRouteTable) (map[string]dataproto.RouteTableSubnetReq, error) {

	if len(addSlice) <= 0 {
		return nil, fmt.Errorf("routeTable addSlice is <= 0, not create")
	}

	subnetMap := make(map[string]dataproto.RouteTableSubnetReq, 0)
	createResources := make([]dataproto.RouteTableCreateReq[dataproto.HuaWeiRouteTableCreateExt], 0, len(addSlice))

	for _, one := range addSlice {
		tmpRes := dataproto.RouteTableCreateReq[dataproto.HuaWeiRouteTableCreateExt]{
			AccountID:  accountID,
			CloudID:    one.CloudID,
			Name:       converter.ValToPtr(one.Name),
			Region:     one.Region,
			CloudVpcID: one.CloudVpcID,
			Memo:       one.Memo,
			BkBizID:    constant.UnassignedBiz,
		}
		if one.Extension != nil {
			tmpRes.Extension = &dataproto.HuaWeiRouteTableCreateExt{
				Default:  one.Extension.Default,
				TenantID: one.Extension.TenantID,
			}
			for _, tmpSubnetID := range one.Extension.CloudSubnetIDs {
				subnetMap[tmpSubnetID] = dataproto.RouteTableSubnetReq{
					CloudRouteTableID: one.CloudID,
				}
			}
		}

		createResources = append(createResources, tmpRes)
	}

	createReq := &dataproto.RouteTableBatchCreateReq[dataproto.HuaWeiRouteTableCreateExt]{
		RouteTables: createResources,
	}
	_, err := cli.dbCli.HuaWei.RouteTable.BatchCreate(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("[%s] routetable batch compare db create failed. accountID: %s, resGroupName: %s, err: %v",
			enumor.HuaWei, accountID, resGroupName, err)
		return subnetMap, err
	}

	logs.Infof("[%s] sync routeTable to create routeTable success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(addSlice), kt.Rid)

	return subnetMap, nil
}

func (cli *client) updateRouteTalbe(kt *kit.Kit, accountID string, resGroupName string,
	updateMap map[string]typesroutetable.HuaWeiRouteTable) (map[string]dataproto.RouteTableSubnetReq, error) {

	if len(updateMap) <= 0 {
		return nil, fmt.Errorf("routeTable updateMap is <= 0, not update")
	}

	subnetMap := make(map[string]dataproto.RouteTableSubnetReq, 0)
	updateResources := make([]dataproto.RouteTableBaseInfoUpdateReq, 0, len(updateMap))

	for id, one := range updateMap {
		tmpRes := dataproto.RouteTableBaseInfoUpdateReq{
			IDs: []string{id},
		}
		tmpRes.Data = &dataproto.RouteTableUpdateBaseInfo{
			Name: converter.ValToPtr(one.Name),
			Memo: one.Memo,
		}
		if one.Extension != nil && len(one.Extension.CloudSubnetIDs) > 0 {
			for _, tmpSubnetID := range one.Extension.CloudSubnetIDs {
				subnetMap[tmpSubnetID] = dataproto.RouteTableSubnetReq{
					RouteTableID:      id,
					CloudRouteTableID: one.CloudID,
				}
			}
		}
		updateResources = append(updateResources, tmpRes)
	}

	updateReq := &dataproto.RouteTableBaseInfoBatchUpdateReq{
		RouteTables: updateResources,
	}
	if err := cli.dbCli.Global.RouteTable.BatchUpdateBaseInfo(kt.Ctx, kt.Header(), updateReq); err != nil {
		logs.Errorf("[%s] routetable batch compare db update failed. accountID: %s, resGroupName: %s, err: %v",
			enumor.HuaWei, accountID, resGroupName, err)
		return subnetMap, err
	}

	logs.Infof("[%s] sync routeTable to update routeTable success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(updateMap), kt.Rid)

	return subnetMap, nil
}

func (cli *client) deleteRouteTable(kt *kit.Kit, accountID string, region string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("routeTable delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  delCloudIDs,
	}
	delRouteTableFromCloud, err := cli.listRouteTableFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delRouteTableFromCloud) > 0 {
		logs.Errorf("[%s] validate routeTable not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.HuaWei, checkParams, len(delRouteTableFromCloud), kt.Rid)
		return fmt.Errorf("validate routeTable not exist failed, before delete")
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.RouteTable.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete routeTable failed, err: %v, rid: %s", enumor.HuaWei, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync routeTable to delete routeTable success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listRouteTableFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]typesroutetable.HuaWeiRouteTable, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	routeTables := make([]typesroutetable.HuaWeiRouteTable, 0)
	for _, id := range params.CloudIDs {
		opt := &typesroutetable.HuaWeiRouteTableListOption{
			Region: params.Region,
			ID:     id,
		}
		results, err := cli.cloudCli.ListRouteTables(kt, opt)
		if err != nil {
			logs.Errorf("[%s] list routeTable from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.HuaWei, err,
				params.AccountID, opt, kt.Rid)
			return nil, err
		}
		routeTables = append(routeTables, results...)
	}

	return routeTables, nil
}

func (cli *client) listRouteTableFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]routetable.HuaWeiRouteTable, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: params.AccountID,
				},
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: params.CloudIDs,
				},
				&filter.AtomRule{
					Field: "region",
					Op:    filter.Equal.Factory(),
					Value: params.Region,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	results, err := cli.dbCli.HuaWei.RouteTable.ListRouteTableWithExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list routeTable from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.HuaWei, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	routeTables := make([]routetable.HuaWeiRouteTable, 0)
	for _, one := range results {
		routeTables = append(routeTables, routetable.HuaWeiRouteTable(*one))
	}

	return routeTables, nil
}

// RemoveRouteTableDeleteFromCloud ...
func (cli *client) RemoveRouteTableDeleteFromCloud(kt *kit.Kit, accountID string, region string) error {
	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: region},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.HuaWei.RouteTable.ListRouteTableWithExt(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list routeTable failed, err: %v, req: %v, rid: %s", enumor.HuaWei,
				err, req, kt.Rid)
			return err
		}

		cloudIDs := make([]string, 0)
		for _, one := range resultFromDB {
			cloudIDs = append(cloudIDs, one.CloudID)
		}

		if len(cloudIDs) == 0 {
			break
		}

		var resultFromCloud []typesroutetable.HuaWeiRouteTable
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID: accountID,
				Region:    region,
				CloudIDs:  cloudIDs,
			}
			resultFromCloud, err = cli.listRouteTableFromCloud(kt, params)
			if err != nil {
				return err
			}
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, one.CloudID)
			}

			delCloudIDs := converter.MapKeyToStringSlice(cloudIDMap)
			err = common.CancelRouteTableSubnetRel(kt, cli.dbCli, enumor.HuaWei, delCloudIDs)
			if err != nil {
				logs.Errorf("[%s] routetable batch cancel subnet rel failed. deleteIDs: %v, err: %v, rid: %s",
					enumor.HuaWei, delCloudIDs, err, kt.Rid)
				return err
			}
			if len(delCloudIDs) > 0 {
				if err = cli.deleteRouteTable(kt, accountID, region, delCloudIDs); err != nil {
					logs.Errorf("delete huawei routeTable failed, err: %v, account: %s, region: %s, delCloudIDs: %v, rid: %s",
						err, accountID, region, delCloudIDs, kt.Rid)
					return err
				}
			}
		}

		if len(resultFromDB) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

func isRouteTableChange(cloud typesroutetable.HuaWeiRouteTable,
	db routetable.HuaWeiRouteTable) bool {

	if cloud.Name != db.Name {
		return true
	}
	if !assert.IsPtrStringEqual(cloud.Memo, db.Memo) {
		return true
	}

	return false
}

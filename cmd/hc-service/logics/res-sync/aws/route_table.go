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

	addSlice, updateMap, delCloudIDs := common.Diff[typesroutetable.AwsRouteTable,
		routetable.AwsRouteTable](routeTableFromCloud, routeTableFromDB, isRouteTableChange)

	subnetMap := make(map[string]dataproto.RouteTableSubnetReq, 0)

	if len(delCloudIDs) > 0 {
		err = common.CancelRouteTableSubnetRel(kt, cli.dbCli, enumor.Aws, delCloudIDs)
		if err != nil {
			logs.Errorf("[%s] routetable batch cancel subnet rel failed. deleteIDs: %v, err: %v, rid: %s",
				enumor.Aws, delCloudIDs, err, kt.Rid)
			return nil, err
		}
		if err = cli.deleteRouteTable(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
			logs.Errorf("delete aws routeTable failed, err: %v, account: %s, region: %s, delCloudIDs: %v, rid: %s",
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
		err = common.UpdateSubnetRouteTableByIDs(kt, enumor.Aws, subnetMap, cli.dbCli)
		if err != nil {
			logs.Errorf("[%s] routetable update subnet's route_table failed. accountID: %s, region: %s, err: %v",
				enumor.Aws, params.AccountID, params.Region, err)
			return nil, err
		}
	}

	if err = cli.syncCloud(kt, params); err != nil {
		return nil, err
	}

	return new(SyncResult), nil
}

func (cli *client) syncCloud(kt *kit.Kit, params *SyncBaseParams) error {
	existRT, err := cli.listRouteTableFromDB(kt, params)
	if err != nil {
		return err
	}

	// 同步db中路由表的路由规则
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
	addSlice []typesroutetable.AwsRouteTable) (map[string]dataproto.RouteTableSubnetReq, error) {

	if len(addSlice) <= 0 {
		return nil, fmt.Errorf("routeTable addSlice is <= 0, not create")
	}

	subnetMap := make(map[string]dataproto.RouteTableSubnetReq, 0)
	createResources := make([]dataproto.RouteTableCreateReq[dataproto.AwsRouteTableCreateExt], 0, len(addSlice))

	for _, one := range addSlice {
		tmpRes := dataproto.RouteTableCreateReq[dataproto.AwsRouteTableCreateExt]{
			AccountID:  accountID,
			CloudID:    one.CloudID,
			Name:       converter.ValToPtr(one.Name),
			Region:     one.Region,
			CloudVpcID: one.CloudVpcID,
			Memo:       one.Memo,
			BkBizID:    constant.UnassignedBiz,
		}
		if one.Extension != nil {
			tmpRes.Extension = &dataproto.AwsRouteTableCreateExt{
				Main: one.Extension.Main,
			}
			for _, subnetItem := range one.Extension.Associations {
				subnetMap[converter.PtrToVal(subnetItem.CloudSubnetID)] = dataproto.RouteTableSubnetReq{
					CloudRouteTableID: one.CloudID,
				}
			}
		}

		createResources = append(createResources, tmpRes)
	}

	createReq := &dataproto.RouteTableBatchCreateReq[dataproto.AwsRouteTableCreateExt]{
		RouteTables: createResources,
	}
	_, err := cli.dbCli.Aws.RouteTable.BatchCreate(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("[%s] routetable batch compare db create failed. accountID: %s, resGroupName: %s, err: %v",
			enumor.Aws, accountID, resGroupName, err)
		return subnetMap, err
	}

	logs.Infof("[%s] sync routeTable to create routeTable success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(addSlice), kt.Rid)

	return subnetMap, nil
}

func (cli *client) updateRouteTalbe(kt *kit.Kit, accountID string, resGroupName string,
	updateMap map[string]typesroutetable.AwsRouteTable) (map[string]dataproto.RouteTableSubnetReq, error) {

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
		if one.Extension != nil && len(one.Extension.Associations) > 0 {
			for _, subnetItem := range one.Extension.Associations {
				subnetMap[converter.PtrToVal(subnetItem.CloudSubnetID)] = dataproto.RouteTableSubnetReq{
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
			enumor.Aws, accountID, resGroupName, err)
		return subnetMap, err
	}

	logs.Infof("[%s] sync routeTable to update routeTable success, accountID: %s, count: %d, rid: %s", enumor.Aws,
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
			enumor.Aws, checkParams, len(delRouteTableFromCloud), kt.Rid)
		return fmt.Errorf("validate routeTable not exist failed, before delete")
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.RouteTable.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete routeTable failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync routeTable to delete routeTable success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) listRouteTableFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]typesroutetable.AwsRouteTable, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &typesroutetable.AwsRouteTableListOption{
		AwsListOption: &adcore.AwsListOption{
			Region:   params.Region,
			CloudIDs: params.CloudIDs,
		},
	}
	result, err := cli.cloudCli.ListRouteTable(kt, opt)
	if err != nil {
		if strings.Contains(err.Error(), aws.ErrRouteTableNotFound) {
			return nil, nil
		}

		logs.Errorf("[%s] list routeTable from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Aws, err,
			params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) listRouteTableFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]routetable.AwsRouteTable, error) {

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
	results, err := cli.dbCli.Aws.RouteTable.ListRouteTableWithExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list routeTable from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Aws, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	routeTables := make([]routetable.AwsRouteTable, 0)
	for _, one := range results {
		routeTables = append(routeTables, routetable.AwsRouteTable(*one))
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
		resultFromDB, err := cli.dbCli.Aws.RouteTable.ListRouteTableWithExt(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list routeTable failed, err: %v, req: %v, rid: %s", enumor.Aws,
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

		var delCloudIDs []string
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID: accountID,
				Region:    region,
				CloudIDs:  cloudIDs,
			}
			delCloudIDs, err = cli.listRemoveRouteTableID(kt, params)
			if err != nil {
				logs.Errorf("list remove routeTableID failed, err: %v, opt: %v, rid: %s", err, params, kt.Rid)
				return err
			}
		}

		err = common.CancelRouteTableSubnetRel(kt, cli.dbCli, enumor.Aws, delCloudIDs)
		if err != nil {
			logs.Errorf("[%s] routetable batch cancel subnet rel failed. deleteIDs: %v, err: %v, rid: %s",
				enumor.Aws, delCloudIDs, err, kt.Rid)
			return err
		}

		if len(delCloudIDs) != 0 {
			if err = cli.deleteRouteTable(kt, accountID, region, delCloudIDs); err != nil {
				logs.Errorf("delete aws routeTable failed, err: %v, account: %s, region: %s, delCloudIDs: %v, rid: %s",
					err, accountID, region, delCloudIDs, kt.Rid)
				return err
			}
		}

		if len(resultFromDB) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

func (cli *client) listRemoveRouteTableID(kt *kit.Kit, params *SyncBaseParams) ([]string, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	delCloudIDs := make([]string, 0)
	for _, one := range params.CloudIDs {
		opt := &typesroutetable.AwsRouteTableListOption{
			AwsListOption: &adcore.AwsListOption{
				Region:   params.Region,
				CloudIDs: []string{one},
			},
		}
		_, err := cli.cloudCli.ListRouteTable(kt, opt)
		if err != nil {
			if strings.Contains(err.Error(), aws.ErrRouteTableNotFound) {
				delCloudIDs = append(delCloudIDs, one)
			}
		}
	}

	return delCloudIDs, nil
}

func isRouteTableChange(cloud typesroutetable.AwsRouteTable,
	db routetable.AwsRouteTable) bool {

	if cloud.Name != db.Name {
		return true
	}
	if !assert.IsPtrStringEqual(cloud.Memo, db.Memo) {
		return true
	}

	return false
}

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
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
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

// SyncVpcOption ...
type SyncVpcOption struct {
}

// Validate ...
func (opt SyncVpcOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Vpc ...
func (cli *client) Vpc(kt *kit.Kit, params *SyncBaseParams, opt *SyncVpcOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vpcFromCloud, err := cli.listVpcFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	vpcFromDB, err := cli.listVpcFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(vpcFromCloud) == 0 && len(vpcFromDB) == 0 {
		return new(SyncResult), nil
	}

	addVpc, updateMap, delCloudIDs := common.Diff[types.AzureVpc, cloudcore.Vpc[cloudcore.AzureVpcExtension]](
		vpcFromCloud, vpcFromDB, isVpcChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteVpc(kt, params.AccountID, params.ResourceGroupName, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addVpc) > 0 {
		if err = cli.createVpc(kt, params.AccountID, addVpc); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateVpc(kt, params.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

// RemoveVpcDeleteFromCloud ...
func (cli *client) RemoveVpcDeleteFromCloud(kt *kit.Kit, accountID string, resGroupName string) error {

	req := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "extension.resource_group_name", Op: filter.JSONEqual.Factory(),
					Value: resGroupName},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Global.Vpc.List(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list vpc failed, err: %v, req: %v, rid: %s", enumor.Azure,
				err, req, kt.Rid)
			return err
		}

		cloudIDs := make([]string, 0)
		for _, one := range resultFromDB.Details {
			cloudIDs = append(cloudIDs, one.CloudID)
		}

		if len(cloudIDs) == 0 {
			break
		}

		var resultFromCloud []types.AzureVpc
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID:         accountID,
				ResourceGroupName: resGroupName,
				CloudIDs:          cloudIDs,
			}
			resultFromCloud, err = cli.listVpcFromCloud(kt, params)
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
			if err = cli.deleteVpc(kt, accountID, resGroupName, delCloudIDs); err != nil {
				return err
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

func (cli *client) deleteVpc(kt *kit.Kit, accountID string, resGroupName string, delCloudIDs []string) error {

	if len(delCloudIDs) == 0 {
		return fmt.Errorf("delete vpc, cloudIDs is required")
	}

	checkParams := &SyncBaseParams{
		AccountID:         accountID,
		ResourceGroupName: resGroupName,
		CloudIDs:          delCloudIDs,
	}
	delVpcFromCloud, err := cli.listVpcFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delVpcFromCloud) > 0 {
		logs.Errorf("[%s] validate vpc not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Azure, checkParams, len(delVpcFromCloud), kt.Rid)
		return fmt.Errorf("validate vpc not exist failed, before delete")
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.Vpc.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete vpc failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync vpc to delete vpc success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) updateVpc(kt *kit.Kit, accountID string, updateMap map[string]types.AzureVpc) error {
	if len(updateMap) == 0 {
		return fmt.Errorf("update vpc, vpcs is required")
	}

	vpcs := make([]cloud.VpcUpdateReq[cloud.AzureVpcUpdateExt], 0)
	for id, one := range updateMap {
		tmpRes := cloud.VpcUpdateReq[cloud.AzureVpcUpdateExt]{
			ID: id,
			VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
				Name: converter.ValToPtr(one.Name),
				Memo: one.Memo,
			},
			Extension: &cloud.AzureVpcUpdateExt{
				DNSServers: one.Extension.DNSServers,
			},
		}

		if one.Extension.Cidr != nil {
			tmpCidrs := make([]cloud.AzureCidr, 0, len(one.Extension.Cidr))
			for _, cidrItem := range one.Extension.Cidr {
				tmpCidrs = append(tmpCidrs, cloud.AzureCidr{
					Type: cidrItem.Type,
					Cidr: cidrItem.Cidr,
				})
			}
			tmpRes.Extension.Cidr = tmpCidrs
		}

		vpcs = append(vpcs, tmpRes)
	}

	updateReq := &cloud.VpcBatchUpdateReq[cloud.AzureVpcUpdateExt]{
		Vpcs: vpcs,
	}
	if err := cli.dbCli.Azure.Vpc.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch update db vpc failed, err: %v, rid: %s", enumor.Azure,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync vpc to update vpc success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createVpc(kt *kit.Kit, accountID string, addVpc []types.AzureVpc) error {

	if len(addVpc) == 0 {
		return fmt.Errorf("create vpc, vpcs is required")
	}

	vpcs := make([]cloud.VpcCreateReq[cloud.AzureVpcCreateExt], 0, len(addVpc))
	for _, one := range addVpc {
		tmpRes := cloud.VpcCreateReq[cloud.AzureVpcCreateExt]{
			AccountID: accountID,
			CloudID:   one.CloudID,
			Name:      converter.ValToPtr(one.Name),
			BkBizID:   constant.UnassignedBiz,
			BkCloudID: constant.UnbindBkCloudID,
			Region:    one.Region,
			Category:  enumor.BizVpcCategory,
			Memo:      one.Memo,
			Extension: &cloud.AzureVpcCreateExt{
				ResourceGroupName: one.Extension.ResourceGroupName,
				DNSServers:        one.Extension.DNSServers,
			},
		}

		if one.Extension.Cidr != nil {
			tmpCidrs := make([]cloud.AzureCidr, 0, len(one.Extension.Cidr))
			for _, cidrItem := range one.Extension.Cidr {
				tmpCidrs = append(tmpCidrs, cloud.AzureCidr{
					Type: cidrItem.Type,
					Cidr: cidrItem.Cidr,
				})
			}
			tmpRes.Extension.Cidr = tmpCidrs
		}

		vpcs = append(vpcs, tmpRes)
	}

	createReq := &cloud.VpcBatchCreateReq[cloud.AzureVpcCreateExt]{
		Vpcs: vpcs,
	}
	if _, err := cli.dbCli.Azure.Vpc.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch create vpc failed, err: %v, rid: %s", enumor.Azure, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync vpc to create vpc success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(addVpc), kt.Rid)

	return nil
}

func (cli *client) listVpcFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]types.AzureVpc, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &adcore.AzureListByIDOption{
		ResourceGroupName: params.ResourceGroupName,
		CloudIDs:          params.CloudIDs,
	}
	result, err := cli.cloudCli.ListVpcByID(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list vpc from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure, err,
			params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) listVpcFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]cloudcore.Vpc[cloudcore.AzureVpcExtension], error) {

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
					Field: "extension.resource_group_name",
					Op:    filter.JSONEqual.Factory(),
					Value: params.ResourceGroupName,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Azure.Vpc.ListVpcExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list vpc from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Azure, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func isVpcChange(item types.AzureVpc, info cloudcore.Vpc[cloudcore.AzureVpcExtension]) bool {
	if info.Name != item.Name {
		return true
	}

	if info.Region != item.Region {
		return true
	}

	if !assert.IsPtrStringEqual(info.Memo, item.Memo) {
		return true
	}

	for _, db := range info.Extension.Cidr {
		for _, cloud := range item.Extension.Cidr {
			if db.Cidr != cloud.Cidr {
				return true
			}

			if db.Type != cloud.Type {
				return true
			}
		}
	}

	if info.Extension.ResourceGroupName != item.Extension.ResourceGroupName {
		return true
	}

	if !assert.IsStringSliceEqual(info.Extension.DNSServers, item.Extension.DNSServers) {
		return true
	}

	return false
}

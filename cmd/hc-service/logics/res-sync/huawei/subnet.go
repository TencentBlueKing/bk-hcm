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
	"hcm/pkg/adaptor/types/subnet"
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

// SyncSubnetOption ...
type SyncSubnetOption struct {
	CloudVpcID string `json:"cloud_vpc_id" validate:"required"`
}

// Validate ...
func (opt SyncSubnetOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Subnet ...
func (cli *client) Subnet(kt *kit.Kit, params *SyncBaseParams, opt *SyncSubnetOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	subnetFromCloud, err := cli.listSubnetFromCloud(kt, params, opt.CloudVpcID)
	if err != nil {
		return nil, err
	}

	subnetFromDB, err := cli.listSubnetFromDB(kt, params, opt.CloudVpcID)
	if err != nil {
		return nil, err
	}

	if len(subnetFromCloud) == 0 && len(subnetFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSubnet, updateMap, delCloudIDs := common.Diff[adtysubnet.HuaWeiSubnet,
		cloudcore.Subnet[cloudcore.HuaWeiSubnetExtension]](subnetFromCloud, subnetFromDB, isHuaWeiSubnetChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteSubnet(kt, params.AccountID, params.Region, opt.CloudVpcID, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSubnet) > 0 {
		if err = cli.createSubnet(kt, params.AccountID, params.Region, opt.CloudVpcID, addSubnet); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateSubnet(kt, params.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

// RemoveSubnetDeleteFromCloud ...
func (cli *client) RemoveSubnetDeleteFromCloud(kt *kit.Kit, accountID, region, cloudVpcID string) error {

	req := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: region},
				&filter.AtomRule{Field: "cloud_vpc_id", Op: filter.Equal.Factory(), Value: cloudVpcID},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Global.Subnet.List(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list subnet failed, err: %v, req: %v, rid: %s", enumor.HuaWei,
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

		var resultFromCloud []adtysubnet.HuaWeiSubnet
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID: accountID,
				Region:    region,
				CloudIDs:  cloudIDs,
			}
			resultFromCloud, err = cli.listSubnetFromCloud(kt, params, cloudVpcID)
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
			if err = cli.deleteSubnet(kt, accountID, region, cloudVpcID, delCloudIDs); err != nil {
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

func (cli *client) deleteSubnet(kt *kit.Kit, accountID, region, cloudVpcID string, delCloudIDs []string) error {
	if len(delCloudIDs) == 0 {
		return fmt.Errorf("delete subnet, cloudIDs is required")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  delCloudIDs,
	}
	delSubnetFromCloud, err := cli.listSubnetFromCloud(kt, checkParams, cloudVpcID)
	if err != nil {
		return err
	}

	if len(delSubnetFromCloud) > 0 {
		logs.Errorf("[%s] validate subnet not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.HuaWei, checkParams, len(delSubnetFromCloud), kt.Rid)
		return fmt.Errorf("validate subnet not exist failed, before delete")
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.Subnet.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete subnet failed, err: %v, rid: %s", enumor.HuaWei, err,
			kt.Rid)
		return err
	}

	logs.Infof("[%s] sync subnet to delete subnet success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) updateSubnet(kt *kit.Kit, accountID string, updateMap map[string]adtysubnet.HuaWeiSubnet) error {
	if len(updateMap) == 0 {
		return fmt.Errorf("update subnet, subnets is required")
	}

	subnets := make([]cloud.SubnetUpdateReq[cloud.HuaWeiSubnetUpdateExt], 0)
	for id, item := range updateMap {
		tmpRes := cloud.SubnetUpdateReq[cloud.HuaWeiSubnetUpdateExt]{
			ID: id,
			SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
				Region:   item.Region,
				Name:     converter.ValToPtr(item.Name),
				Ipv4Cidr: item.Ipv4Cidr,
				Ipv6Cidr: item.Ipv6Cidr,
				Memo:     item.Memo,
			},
			Extension: &cloud.HuaWeiSubnetUpdateExt{
				Status:       item.Extension.Status,
				DhcpEnable:   converter.ValToPtr(item.Extension.DhcpEnable),
				GatewayIp:    item.Extension.GatewayIp,
				DnsList:      item.Extension.DnsList,
				NtpAddresses: item.Extension.NtpAddresses,
			},
		}

		subnets = append(subnets, tmpRes)
	}

	updateReq := &cloud.SubnetBatchUpdateReq[cloud.HuaWeiSubnetUpdateExt]{
		Subnets: subnets,
	}
	if err := cli.dbCli.HuaWei.Subnet.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch update db subnet failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync subnet to update subnet success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createSubnet(kt *kit.Kit, accountID, region, cloudVpcID string,
	addSubnet []adtysubnet.HuaWeiSubnet) error {
	if len(addSubnet) == 0 {
		return fmt.Errorf("create subnet, subnets is required")
	}

	params := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  []string{cloudVpcID},
	}
	vpcs, err := cli.listVpcFromDB(kt, params)
	if err != nil {
		return err
	}

	if len(vpcs) == 0 {
		return fmt.Errorf("vpc: %s not found", cloudVpcID)
	}

	subnets := make([]cloud.SubnetCreateReq[cloud.HuaWeiSubnetCreateExt], 0, len(addSubnet))
	for _, item := range addSubnet {
		tmpRes := cloud.SubnetCreateReq[cloud.HuaWeiSubnetCreateExt]{
			AccountID:  accountID,
			CloudVpcID: item.CloudVpcID,
			VpcID:      vpcs[0].ID,
			BkBizID:    constant.UnassignedBiz,
			CloudID:    item.CloudID,
			Name:       converter.ValToPtr(item.Name),
			Region:     item.Region,
			Zone:       "",
			Ipv4Cidr:   item.Ipv4Cidr,
			Ipv6Cidr:   item.Ipv6Cidr,
			Memo:       item.Memo,
			Extension: &cloud.HuaWeiSubnetCreateExt{
				Status:       item.Extension.Status,
				DhcpEnable:   item.Extension.DhcpEnable,
				GatewayIp:    item.Extension.GatewayIp,
				DnsList:      item.Extension.DnsList,
				NtpAddresses: item.Extension.NtpAddresses,
			},
		}

		subnets = append(subnets, tmpRes)
	}

	createReq := &cloud.SubnetBatchCreateReq[cloud.HuaWeiSubnetCreateExt]{
		Subnets: subnets,
	}
	if _, err := cli.dbCli.HuaWei.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch create subnet failed, err: %v, rid: %s", enumor.HuaWei, err,
			kt.Rid)
		return err
	}

	logs.Infof("[%s] sync subnet to create subnet success, accountID: %s, count: %d, rid: %s", enumor.HuaWei,
		accountID, len(addSubnet), kt.Rid)

	return nil
}

func (cli *client) listSubnetFromCloud(kt *kit.Kit, params *SyncBaseParams, cloudVpcID string) (
	[]adtysubnet.HuaWeiSubnet, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &adtysubnet.HuaWeiSubnetListByIDOption{
		Region:     params.Region,
		CloudIDs:   params.CloudIDs,
		CloudVpcID: cloudVpcID,
	}
	result, err := cli.cloudCli.ListSubnetByID(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list subnet from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.HuaWei, err,
			params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) listSubnetFromDB(kt *kit.Kit, params *SyncBaseParams, cloudVpcID string) (
	[]cloudcore.Subnet[cloudcore.HuaWeiSubnetExtension], error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: params.AccountID},
				&filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: params.CloudIDs},
				&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: params.Region},
				&filter.AtomRule{Field: "cloud_vpc_id", Op: filter.Equal.Factory(), Value: cloudVpcID},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.HuaWei.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list subnet from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.HuaWei, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func isHuaWeiSubnetChange(item adtysubnet.HuaWeiSubnet, info cloudcore.Subnet[cloudcore.HuaWeiSubnetExtension]) bool {
	if info.Region != item.Region {
		return true
	}

	if info.CloudVpcID != item.CloudVpcID {
		return true
	}

	if info.Name != item.Name {
		return true
	}

	if !assert.IsStringSliceEqual(info.Ipv4Cidr, item.Ipv4Cidr) {
		return true
	}

	if !assert.IsStringSliceEqual(info.Ipv6Cidr, item.Ipv6Cidr) {
		return true
	}

	if !assert.IsPtrStringEqual(item.Memo, info.Memo) {
		return true
	}

	if info.Extension.Status != item.Extension.Status {
		return true
	}

	if info.Extension.DhcpEnable != item.Extension.DhcpEnable {
		return true
	}

	if info.Extension.GatewayIp != item.Extension.GatewayIp {
		return true
	}

	if !assert.IsStringSliceEqual(info.Extension.DnsList, item.Extension.DnsList) {
		return true
	}

	if !assert.IsStringSliceEqual(info.Extension.NtpAddresses, item.Extension.NtpAddresses) {
		return true
	}

	return false
}

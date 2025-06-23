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

package tcloud

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	adcore "hcm/pkg/adaptor/types/core"
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

	subnetFromCloud, err := cli.listSubnetFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	subnetFromDB, err := cli.listSubnetFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(subnetFromCloud) == 0 && len(subnetFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSubnet, updateMap, delCloudIDs := common.Diff[adtysubnet.TCloudSubnet,
		cloudcore.Subnet[cloudcore.TCloudSubnetExtension]](subnetFromCloud, subnetFromDB, isTCloudSubnetChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteSubnet(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSubnet) > 0 {
		if err = cli.createSubnet(kt, params.AccountID, params.Region, addSubnet); err != nil {
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

// deleteSubnet delete subnet from db
func (cli *client) deleteSubnet(kt *kit.Kit, accountID string, region string, delCloudIDs []string) error {
	if len(delCloudIDs) == 0 {
		return fmt.Errorf("delete subnet, cloudIDs is required")
	}

	// check subnet not exist in cloud, before delete
	checkParams := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  delCloudIDs,
	}
	delFromCloud, err := cli.listSubnetFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delFromCloud) > 0 {
		logs.Errorf("[%s] validate subnet not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.TCloud, checkParams, len(delFromCloud), kt.Rid)
		return fmt.Errorf("validate subnet not exist failed, before delete")
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.Subnet.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete subnet failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync subnet to delete subnet success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

// updateSubnet update subnet in db
func (cli *client) updateSubnet(kt *kit.Kit, accountID string, updateMap map[string]adtysubnet.TCloudSubnet) error {
	if len(updateMap) == 0 {
		return fmt.Errorf("update subnet, subnets is required")
	}

	subnets := make([]cloud.SubnetUpdateReq[cloud.TCloudSubnetUpdateExt], 0)
	for id, item := range updateMap {
		tmpRes := cloud.SubnetUpdateReq[cloud.TCloudSubnetUpdateExt]{
			ID: id,
			SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
				Region:   item.Region,
				Name:     converter.ValToPtr(item.Name),
				Ipv4Cidr: item.Ipv4Cidr,
				Ipv6Cidr: item.Ipv6Cidr,
				Memo:     item.Memo,
			},
			Extension: &cloud.TCloudSubnetUpdateExt{
				IsDefault:         item.Extension.IsDefault,
				Region:            item.Region,
				Zone:              item.Extension.Zone,
				CloudNetworkAclID: item.Extension.CloudNetworkAclID,
			},
		}

		subnets = append(subnets, tmpRes)
	}

	updateReq := &cloud.SubnetBatchUpdateReq[cloud.TCloudSubnetUpdateExt]{
		Subnets: subnets,
	}
	if err := cli.dbCli.TCloud.Subnet.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch update db subnet failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync subnet to update subnet success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		accountID, len(updateMap), kt.Rid)

	return nil
}

// createSubnet create subnet in db
func (cli *client) createSubnet(kt *kit.Kit, accountID string, region string,
	addSubnet []adtysubnet.TCloudSubnet) error {
	if len(addSubnet) == 0 {
		return fmt.Errorf("create subnet, subnets is required")
	}

	vpcCloudIDMap := make(map[string]struct{})
	for _, one := range addSubnet {
		vpcCloudIDMap[one.CloudVpcID] = struct{}{}
	}

	params := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  converter.MapKeyToStringSlice(vpcCloudIDMap),
	}
	vpcs, err := cli.listVpcFromDB(kt, params)
	if err != nil {
		return err
	}

	cloudIDMap := make(map[string]string)
	for _, vpc := range vpcs {
		cloudIDMap[vpc.CloudID] = vpc.ID
	}

	subnets := make([]cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt], 0, len(addSubnet))
	for _, item := range addSubnet {
		vpcID, exist := cloudIDMap[item.CloudVpcID]
		if !exist {
			logs.Errorf("create subnet to get vpc id not found, subnet: %v, cloudVpcID: %s, rid: %s",
				item, item.CloudVpcID, kt.Rid)
			return fmt.Errorf("create subnet to get vpc id not found")
		}

		tmpRes := cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt]{
			AccountID:  accountID,
			CloudVpcID: item.CloudVpcID,
			VpcID:      vpcID,
			BkBizID:    constant.UnassignedBiz,
			CloudID:    item.CloudID,
			Name:       converter.ValToPtr(item.Name),
			Region:     item.Region,
			Zone:       item.Extension.Zone,
			Ipv4Cidr:   item.Ipv4Cidr,
			Ipv6Cidr:   item.Ipv6Cidr,
			Memo:       item.Memo,
			Extension: &cloud.TCloudSubnetCreateExt{
				IsDefault:         item.Extension.IsDefault,
				CloudNetworkAclID: item.Extension.CloudNetworkAclID,
			},
		}

		subnets = append(subnets, tmpRes)
	}

	createReq := &cloud.SubnetBatchCreateReq[cloud.TCloudSubnetCreateExt]{
		Subnets: subnets,
	}
	if _, err := cli.dbCli.TCloud.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch create subnet failed, err: %v, rid: %s", enumor.TCloud, err,
			kt.Rid)
		return err
	}

	logs.Infof("[%s] sync subnet to create subnet success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		accountID, len(addSubnet), kt.Rid)

	return nil
}

// isTCloudSubnetChange check if the subnet has changed
func isTCloudSubnetChange(item adtysubnet.TCloudSubnet, info cloudcore.Subnet[cloudcore.TCloudSubnetExtension]) bool {
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

	if info.Extension.IsDefault != item.Extension.IsDefault {
		return true
	}

	if !assert.IsPtrStringEqual(info.Extension.CloudNetworkAclId, item.Extension.CloudNetworkAclID) {
		return true
	}

	return false
}

func (cli *client) listSubnetFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]cloudcore.Subnet[cloudcore.TCloudSubnetExtension], error) {

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
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.TCloud.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list subnet from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.TCloud, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) listSubnetFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]adtysubnet.TCloudSubnet, error) {

	opt := &adcore.TCloudListOption{
		Region:   params.Region,
		CloudIDs: params.CloudIDs,
		Page: &adcore.TCloudPage{
			Offset: 0,
			Limit:  adcore.TCloudQueryLimit,
		},
	}
	result, err := cli.cloudCli.ListSubnet(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list subnet from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.TCloud, err,
			params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

// RemoveSubnetDeleteFromCloud ...
func (cli *client) RemoveSubnetDeleteFromCloud(kt *kit.Kit, accountID string, region string) error {

	req := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
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
		resultFromDB, err := cli.dbCli.Global.Subnet.List(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list subnet failed, err: %v, req: %v, rid: %s", enumor.TCloud,
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

		var resultFromCloud []adtysubnet.TCloudSubnet
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID: accountID,
				Region:    region,
				CloudIDs:  cloudIDs,
			}
			resultFromCloud, err = cli.listSubnetFromCloud(kt, params)
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
			if err = cli.deleteSubnet(kt, accountID, region, delCloudIDs); err != nil {
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

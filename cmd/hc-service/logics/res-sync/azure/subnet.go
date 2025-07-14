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

	addSubnet, updateMap, delCloudIDs := common.Diff[adtysubnet.AzureSubnet,
		cloudcore.Subnet[cloudcore.AzureSubnetExtension]](subnetFromCloud, subnetFromDB, isSubnetChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteSubnet(kt, params.AccountID, params.ResourceGroupName, opt.CloudVpcID,
			delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSubnet) > 0 {
		if err = cli.createSubnet(kt, params.AccountID, params.ResourceGroupName, opt.CloudVpcID,
			addSubnet); err != nil {
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
func (cli *client) RemoveSubnetDeleteFromCloud(kt *kit.Kit, accountID, resGroupName, cloudVpcID string) error {

	req := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "cloud_vpc_id", Op: filter.Equal.Factory(), Value: cloudVpcID},
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
		resultFromDB, err := cli.dbCli.Global.Subnet.List(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list subnet failed, err: %v, req: %v, rid: %s", enumor.Azure,
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

		var resultFromCloud []adtysubnet.AzureSubnet
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID:         accountID,
				ResourceGroupName: resGroupName,
				CloudIDs:          cloudIDs,
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
			if err = cli.deleteSubnet(kt, accountID, resGroupName, cloudVpcID, delCloudIDs); err != nil {
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

func (cli *client) deleteSubnet(kt *kit.Kit, accountID, resGroupName, cloudVpcID string, delCloudIDs []string) error {

	if len(delCloudIDs) == 0 {
		return fmt.Errorf("delete subnet, cloudIDs is required")
	}

	checkParams := &SyncBaseParams{
		AccountID:         accountID,
		ResourceGroupName: resGroupName,
		CloudIDs:          delCloudIDs,
	}
	delSubnetFromCloud, err := cli.listSubnetFromCloud(kt, checkParams, cloudVpcID)
	if err != nil {
		return err
	}

	if len(delSubnetFromCloud) > 0 {
		logs.Errorf("[%s] validate subnet not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Azure, checkParams, len(delSubnetFromCloud), kt.Rid)
		return fmt.Errorf("validate subnet not exist failed, before delete")
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.Subnet.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete subnet failed, err: %v, rid: %s",
			enumor.Azure, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync subnet to delete subnet success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) updateSubnet(kt *kit.Kit, accountID string, updateMap map[string]adtysubnet.AzureSubnet) error {

	if len(updateMap) == 0 {
		return fmt.Errorf("update subnet, subnets is required")
	}

	subnets := make([]cloud.SubnetUpdateReq[cloud.AzureSubnetUpdateExt], 0)
	for id, one := range updateMap {
		tmpRes := cloud.SubnetUpdateReq[cloud.AzureSubnetUpdateExt]{
			ID: id,
			SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
				Name:     converter.ValToPtr(one.Name),
				Ipv4Cidr: one.Ipv4Cidr,
				Ipv6Cidr: one.Ipv6Cidr,
				Region:   one.Region,
				Memo:     one.Memo,
			},
			Extension: &cloud.AzureSubnetUpdateExt{
				NatGateway:           converter.ValToPtr(one.Extension.NatGateway),
				CloudSecurityGroupID: converter.ValToPtr(one.Extension.NetworkSecurityGroup),
				SecurityGroupID:      nil,
			},
		}

		subnets = append(subnets, tmpRes)
	}

	updateReq := &cloud.SubnetBatchUpdateReq[cloud.AzureSubnetUpdateExt]{
		Subnets: subnets,
	}
	if err := cli.dbCli.Azure.Subnet.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch update db subnet failed, err: %v, rid: %s", enumor.Azure,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync subnet to update subnet success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createSubnet(kt *kit.Kit, accountID, resGroupName, cloudVpcID string,
	addSubnet []adtysubnet.AzureSubnet) error {

	if len(addSubnet) == 0 {
		return fmt.Errorf("create subnet, subnets is required")
	}

	params := &SyncBaseParams{
		AccountID:         accountID,
		ResourceGroupName: resGroupName,
		CloudIDs:          []string{cloudVpcID},
	}
	vpcs, err := cli.listVpcFromDB(kt, params)
	if err != nil {
		return err
	}

	if len(vpcs) == 0 {
		return fmt.Errorf("vpc: %s not found", cloudVpcID)
	}

	subnets := make([]cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt], 0, len(addSubnet))
	for _, one := range addSubnet {
		// need add subnet data
		tmpRes := cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt]{
			AccountID:  accountID,
			CloudVpcID: one.CloudVpcID,
			VpcID:      vpcs[0].ID,
			BkBizID:    constant.UnassignedBiz,
			CloudID:    one.CloudID,
			Name:       converter.ValToPtr(one.Name),
			Region:     one.Region,
			Zone:       "",
			Ipv4Cidr:   one.Ipv4Cidr,
			Ipv6Cidr:   one.Ipv6Cidr,
			Memo:       one.Memo,
			Extension: &cloud.AzureSubnetCreateExt{
				ResourceGroupName:    one.Extension.ResourceGroupName,
				NatGateway:           one.Extension.NatGateway,
				CloudSecurityGroupID: one.Extension.NetworkSecurityGroup,
				SecurityGroupID:      "",
			},
		}

		subnets = append(subnets, tmpRes)
	}

	createReq := &cloud.SubnetBatchCreateReq[cloud.AzureSubnetCreateExt]{
		Subnets: subnets,
	}
	if _, err := cli.dbCli.Azure.Subnet.BatchCreate(kt, createReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch create subnet failed, err: %v, rid: %s", enumor.Azure, err,
			kt.Rid)
		return err
	}

	logs.Infof("[%s] sync subnet to create subnet success, accountID: %s, count: %d, rid: %s", enumor.Azure,
		accountID, len(addSubnet), kt.Rid)

	return nil
}

func (cli *client) listSubnetFromCloud(kt *kit.Kit, params *SyncBaseParams, cloudVpcId string) (
	[]adtysubnet.AzureSubnet, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &adtysubnet.AzureSubnetListByIDOption{
		AzureListByIDOption: adcore.AzureListByIDOption{
			ResourceGroupName: params.ResourceGroupName,
			CloudIDs:          params.CloudIDs,
		},
		CloudVpcID: cloudVpcId,
	}
	result, err := cli.cloudCli.ListSubnetByID(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list subnet from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Azure, err,
			params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) listSubnetFromDB(kt *kit.Kit, params *SyncBaseParams, cloudVpcID string) (
	[]cloudcore.Subnet[cloudcore.AzureSubnetExtension], error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: params.AccountID},
				&filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: params.CloudIDs},
				&filter.AtomRule{Field: "cloud_vpc_id", Op: filter.Equal.Factory(), Value: cloudVpcID},
				&filter.AtomRule{Field: "extension.resource_group_name", Op: filter.JSONEqual.Factory(),
					Value: params.ResourceGroupName},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Azure.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list subnet from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Azure, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) listSubnetFromDBForCvm(kt *kit.Kit, params *SyncBaseParams) (
	[]cloudcore.Subnet[cloudcore.AzureSubnetExtension], error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: params.AccountID},
				&filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: params.CloudIDs},
				&filter.AtomRule{Field: "extension.resource_group_name", Op: filter.JSONEqual.Factory(),
					Value: params.ResourceGroupName},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Azure.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list subnet from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Azure, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func isSubnetChange(item adtysubnet.AzureSubnet, info cloudcore.Subnet[cloudcore.AzureSubnetExtension]) bool {

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

	if info.Extension.ResourceGroupName != item.Extension.ResourceGroupName {
		return true
	}

	if info.Extension.NatGateway != item.Extension.NatGateway {
		return true
	}

	if info.Extension.SecurityGroupID != item.Extension.NetworkSecurityGroup {
		return true
	}

	if info.Region != item.Region {
		return true
	}

	return false
}

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

package gcp

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
	Region string `json:"region" validate:"required"`
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

	subnetFromCloud, err := cli.listSubnetFromCloud(kt, params, opt.Region)
	if err != nil {
		return nil, err
	}

	subnetFromDB, err := cli.listSubnetFromDB(kt, params, opt.Region)
	if err != nil {
		return nil, err
	}

	if len(subnetFromCloud) == 0 && len(subnetFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSubnet, updateMap, delCloudIDs := common.Diff[adtysubnet.GcpSubnet, cloudcore.Subnet[cloudcore.GcpSubnetExtension]](
		subnetFromCloud, subnetFromDB, isGcpSubnetChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteSubnet(kt, params.AccountID, opt.Region, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSubnet) > 0 {
		if err = cli.createSubnet(kt, params.AccountID, addSubnet); err != nil {
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
func (cli *client) RemoveSubnetDeleteFromCloud(kt *kit.Kit, accountID, region string) error {

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
			logs.Errorf("[%s] request dataservice to list subnet failed, err: %v, req: %v, rid: %s", enumor.Gcp,
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

		var resultFromCloud []adtysubnet.GcpSubnet
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID: accountID,
				CloudIDs:  cloudIDs,
			}
			resultFromCloud, err = cli.listSubnetFromCloud(kt, params, region)
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

func (cli *client) deleteSubnet(kt *kit.Kit, accountID, region string, delCloudIDs []string) error {
	if len(delCloudIDs) == 0 {
		return fmt.Errorf("delete subnet, cloudIDs is required")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		CloudIDs:  delCloudIDs,
	}
	delSubnetFromCloud, err := cli.listSubnetFromCloud(kt, checkParams, region)
	if err != nil {
		return err
	}

	if len(delSubnetFromCloud) > 0 {
		logs.Errorf("[%s] validate subnet not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Gcp, checkParams, len(delSubnetFromCloud), kt.Rid)
		return fmt.Errorf("validate subnet not exist failed, before delete")
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.Subnet.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete subnet failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync subnet to delete subnet success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) updateSubnet(kt *kit.Kit, accountID string, updateMap map[string]adtysubnet.GcpSubnet) error {
	if len(updateMap) == 0 {
		return fmt.Errorf("update subnet, subnets is required")
	}

	subnets := make([]cloud.SubnetUpdateReq[cloud.GcpSubnetUpdateExt], 0)
	for id, item := range updateMap {
		tmpRes := cloud.SubnetUpdateReq[cloud.GcpSubnetUpdateExt]{
			ID: id,
			SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
				Region:   item.Region,
				Name:     converter.ValToPtr(item.Name),
				Ipv4Cidr: item.Ipv4Cidr,
				Ipv6Cidr: item.Ipv6Cidr,
				Memo:     item.Memo,
			},
			Extension: &cloud.GcpSubnetUpdateExt{
				StackType:             item.Extension.StackType,
				Ipv6AccessType:        item.Extension.Ipv6AccessType,
				GatewayAddress:        item.Extension.GatewayAddress,
				PrivateIpGoogleAccess: converter.ValToPtr(item.Extension.PrivateIpGoogleAccess),
				EnableFlowLogs:        converter.ValToPtr(item.Extension.EnableFlowLogs),
			},
		}

		subnets = append(subnets, tmpRes)
	}

	updateReq := &cloud.SubnetBatchUpdateReq[cloud.GcpSubnetUpdateExt]{
		Subnets: subnets,
	}
	if err := cli.dbCli.Gcp.Subnet.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch update db subnet failed, err: %v, rid: %s", enumor.Gcp,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync subnet to update subnet success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createSubnet(kt *kit.Kit, accountID string, addSubnet []adtysubnet.GcpSubnet) error {
	if len(addSubnet) == 0 {
		return fmt.Errorf("create subnet, subnets is required")
	}

	selfLinkMap := make(map[string]struct{})
	for _, one := range addSubnet {
		selfLinkMap[one.CloudVpcID] = struct{}{}
	}

	opt := &ListBySelfLinkOption{
		AccountID: accountID,
		SelfLink:  converter.MapKeyToStringSlice(selfLinkMap),
	}
	vpcs, err := cli.listVpcFromDBBySelfLink(kt, opt)
	if err != nil {
		return err
	}

	slVpcMap := make(map[string]cloudcore.Vpc[cloudcore.GcpVpcExtension], len(vpcs))
	for _, vpc := range vpcs {
		slVpcMap[vpc.Extension.SelfLink] = vpc
	}

	subnets := make([]cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt], 0, len(addSubnet))
	for _, item := range addSubnet {
		vpc, exist := slVpcMap[item.CloudVpcID]
		if !exist {
			logs.Errorf("create subnet to get vpc id not found, subnet: %v, cloudVpcID: %s, rid: %s",
				item, item.CloudVpcID, kt.Rid)
			return fmt.Errorf("create subnet to get vpc id not found")
		}

		tmpRes := cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt]{
			AccountID:  accountID,
			CloudVpcID: vpc.CloudID,
			VpcID:      vpc.ID,
			BkBizID:    constant.UnassignedBiz,
			CloudID:    item.CloudID,
			Name:       converter.ValToPtr(item.Name),
			Region:     item.Region,
			Zone:       "",
			Ipv4Cidr:   item.Ipv4Cidr,
			Ipv6Cidr:   item.Ipv6Cidr,
			Memo:       item.Memo,
			Extension: &cloud.GcpSubnetCreateExt{
				VpcSelfLink:           item.CloudVpcID,
				SelfLink:              item.Extension.SelfLink,
				StackType:             item.Extension.StackType,
				Ipv6AccessType:        item.Extension.Ipv6AccessType,
				GatewayAddress:        item.Extension.GatewayAddress,
				PrivateIpGoogleAccess: item.Extension.PrivateIpGoogleAccess,
				EnableFlowLogs:        item.Extension.EnableFlowLogs,
			},
		}

		subnets = append(subnets, tmpRes)
	}

	createReq := &cloud.SubnetBatchCreateReq[cloud.GcpSubnetCreateExt]{
		Subnets: subnets,
	}
	if _, err := cli.dbCli.Gcp.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch create subnet failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync subnet to create subnet success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(addSubnet), kt.Rid)

	return nil
}

func (cli *client) listSubnetFromCloud(kt *kit.Kit, params *SyncBaseParams, region string) ([]adtysubnet.GcpSubnet,
	error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &adtysubnet.GcpSubnetListOption{
		GcpListOption: adcore.GcpListOption{
			Page: &adcore.GcpPage{
				PageSize: adcore.GcpQueryLimit,
			},
			CloudIDs: params.CloudIDs,
		},
		Region: region,
	}
	result, err := cli.cloudCli.ListSubnet(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list subnet from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Gcp, err,
			params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) listSubnetFromDBBySelfLink(kt *kit.Kit, params *ListSubnetBySelfLinkOption) (
	[]cloudcore.Subnet[cloudcore.GcpSubnetExtension], error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: params.AccountID},
				&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: params.Region},
				&filter.AtomRule{Field: "extension.self_link", Op: filter.JSONIn.Factory(), Value: params.SelfLink},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Gcp.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list subnet from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Gcp, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) listSubnetFromDB(kt *kit.Kit, params *SyncBaseParams, region string) (
	[]cloudcore.Subnet[cloudcore.GcpSubnetExtension], error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: params.AccountID},
				&filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: params.CloudIDs},
				&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: region},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Gcp.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list subnet from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Gcp, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}
func isGcpSubnetChange(item adtysubnet.GcpSubnet, info cloudcore.Subnet[cloudcore.GcpSubnetExtension]) bool {
	if info.Region != item.Region {
		return true
	}

	if info.Extension.VpcSelfLink != item.CloudVpcID {
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

	if info.Extension.SelfLink != item.Extension.SelfLink {
		return true
	}

	if info.Extension.StackType != item.Extension.StackType {
		return true
	}

	if info.Extension.Ipv6AccessType != item.Extension.Ipv6AccessType {
		return true
	}

	if info.Extension.GatewayAddress != item.Extension.GatewayAddress {
		return true
	}

	if info.Extension.PrivateIpGoogleAccess != item.Extension.PrivateIpGoogleAccess {
		return true
	}

	if info.Extension.EnableFlowLogs != item.Extension.EnableFlowLogs {
		return true
	}

	return false
}

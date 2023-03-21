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

// Package subnet defines subnet service.
package subnet

import (
	"errors"
	"fmt"

	"hcm/cmd/hc-service/logics/sync/logics"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/data-service/cloud"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
)

// SyncAwsOption define aws sync option.
type SyncAwsOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"required"`
}

// Validate SyncAwsOption.
func (opt SyncAwsOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) == 0 {
		return errors.New("cloudIDs is required")
	}

	if len(opt.CloudIDs) > int(core.DefaultMaxPageLimit) {
		return fmt.Errorf("cloudIDs should <= %d", core.DefaultMaxPageLimit)
	}

	return nil
}

// AwsSubnetSync sync aws cloud subnet.
func AwsSubnetSync(kt *kit.Kit, opt *SyncAwsOption,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	list, err := listAwsSubnetFromCloud(kt, opt, adaptor)
	if err != nil {
		logs.Errorf("list aws subnet from cloud failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	resourceDBMap, err := listAwsSubnetMapFromDB(kt, opt, dataCli)
	if err != nil {
		logs.Errorf("list aws subnet from db failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	if len(list.Details) == 0 && len(resourceDBMap) == 0 {
		return nil, nil
	}

	err = diffAwsSubnetAndSync(kt, opt, list, resourceDBMap, dataCli, adaptor)
	if err != nil {
		logs.Errorf("diff aws subnet and sync failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	return nil, nil
}

func listAwsSubnetMapFromDB(kt *kit.Kit, opt *SyncAwsOption, dataCli *dataclient.Client) (
	map[string]cloudcore.Subnet[cloudcore.AwsSubnetExtension], error) {

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "account_id",
				Op:    filter.Equal.Factory(),
				Value: opt.AccountID,
			},
			&filter.AtomRule{
				Field: "region",
				Op:    filter.Equal.Factory(),
				Value: opt.Region,
			},
			&filter.AtomRule{
				Field: "cloud_id",
				Op:    filter.In.Factory(),
				Value: opt.CloudIDs,
			},
		},
	}
	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   core.DefaultBasePage,
	}
	dbList, err := dataCli.Aws.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		return nil, err
	}

	resourceMap := make(map[string]cloudcore.Subnet[cloudcore.AwsSubnetExtension], len(dbList.Details))
	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// listAwsSubnetFromCloud batch get subnet list from cloudapi.
func listAwsSubnetFromCloud(kt *kit.Kit, req *SyncAwsOption, adaptor *cloudclient.CloudAdaptorClient) (
	*types.AwsSubnetListResult, error) {

	cli, err := adaptor.Aws(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adcore.AwsListOption{
		Region:   req.Region,
		CloudIDs: req.CloudIDs,
	}

	result, err := cli.ListSubnet(kt, opt)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// diffAwsSubnetAndSync batch sync vendor subnet list.
func diffAwsSubnetAndSync(kt *kit.Kit, req *SyncAwsOption, list *types.AwsSubnetListResult,
	resourceDBMap map[string]cloudcore.Subnet[cloudcore.AwsSubnetExtension], dataCli *dataclient.Client,
	adaptor *cloudclient.CloudAdaptorClient) error {

	createResources, updateResources, delCloudIDs, err := diffAwsSubnet(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.SubnetBatchUpdateReq[cloud.AwsSubnetUpdateExt]{
			Subnets: updateResources,
		}
		if err = dataCli.Aws.Subnet.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			return err
		}
	}

	if len(createResources) > 0 {
		_, err = BatchCreateAwsSubnet(kt, createResources, dataCli, adaptor, req)
		if err != nil {
			return err
		}
	}

	if len(delCloudIDs) > 0 {
		delListOpt := &SyncAwsOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  delCloudIDs,
		}
		delResult, err := listAwsSubnetFromCloud(kt, delListOpt, adaptor)
		if err != nil {
			return err
		}

		if len(delResult.Details) > 0 {
			logs.Errorf("validate subnet not exist failed, before delete, opt: %v, rid: %s", delListOpt, kt.Rid)
			return fmt.Errorf("validate subnet not exist failed, before delete")
		}

		if err = batchDeleteSubnetByIDs(kt, delCloudIDs, dataCli); err != nil {
			return err
		}
	}

	return nil
}

// filterAwsVpcList filter aws subnet list
func diffAwsSubnet(req *SyncAwsOption, list *types.AwsSubnetListResult,
	resourceDBMap map[string]cloudcore.Subnet[cloudcore.AwsSubnetExtension]) (
	[]cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt],
	[]cloud.SubnetUpdateReq[cloud.AwsSubnetUpdateExt], []string, error) {

	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpclist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	createResources := make([]cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt], 0)
	updateResources := make([]cloud.SubnetUpdateReq[cloud.AwsSubnetUpdateExt], 0)
	for _, item := range list.Details {
		// need compare and update subnet data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if isAwsSubnetChange(resourceInfo, item) {
				tmpRes := cloud.SubnetUpdateReq[cloud.AwsSubnetUpdateExt]{
					ID: resourceInfo.ID,
					SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
						Name:              converter.ValToPtr(item.Name),
						Ipv4Cidr:          item.Ipv4Cidr,
						Ipv6Cidr:          item.Ipv6Cidr,
						Memo:              item.Memo,
						CloudRouteTableID: nil,
					},
					Extension: &cloud.AwsSubnetUpdateExt{
						State:                       item.Extension.State,
						Region:                      item.Extension.Region,
						Zone:                        item.Extension.Zone,
						IsDefault:                   converter.ValToPtr(item.Extension.IsDefault),
						MapPublicIpOnLaunch:         converter.ValToPtr(item.Extension.MapPublicIpOnLaunch),
						AssignIpv6AddressOnCreation: converter.ValToPtr(item.Extension.AssignIpv6AddressOnCreation),
						HostnameType:                item.Extension.HostnameType,
					},
				}

				updateResources = append(updateResources, tmpRes)
			}

			delete(resourceDBMap, item.CloudID)
		} else {
			// need add subnet data
			tmpRes := cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt]{
				AccountID:  req.AccountID,
				CloudVpcID: item.CloudVpcID,
				VpcID:      "",
				CloudID:    item.CloudID,
				BkBizID:    constant.UnassignedBiz,
				Name:       converter.ValToPtr(item.Name),
				Region:     item.Extension.Region,
				Zone:       item.Extension.Zone,
				Ipv4Cidr:   item.Ipv4Cidr,
				Memo:       item.Memo,
				Ipv6Cidr:   item.Ipv6Cidr,
				Extension: &cloud.AwsSubnetCreateExt{
					State:                       item.Extension.State,
					IsDefault:                   item.Extension.IsDefault,
					MapPublicIpOnLaunch:         item.Extension.MapPublicIpOnLaunch,
					AssignIpv6AddressOnCreation: item.Extension.AssignIpv6AddressOnCreation,
					HostnameType:                item.Extension.HostnameType,
				},
			}

			createResources = append(createResources, tmpRes)
		}
	}

	deleteCloudIDs := make([]string, 0, len(resourceDBMap))
	for _, vpc := range resourceDBMap {
		deleteCloudIDs = append(deleteCloudIDs, vpc.CloudID)
	}

	return createResources, updateResources, deleteCloudIDs, nil
}

func isAwsSubnetChange(info cloudcore.Subnet[cloudcore.AwsSubnetExtension], item types.AwsSubnet) bool {
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

	if info.Extension.State != item.Extension.State {
		return true
	}

	if info.Extension.IsDefault != item.Extension.IsDefault {
		return true
	}

	if info.Extension.MapPublicIpOnLaunch != item.Extension.MapPublicIpOnLaunch {
		return true
	}

	if info.Extension.AssignIpv6AddressOnCreation != item.Extension.AssignIpv6AddressOnCreation {
		return true
	}

	if info.Extension.HostnameType != item.Extension.HostnameType {
		return true
	}

	return false
}

// BatchCreateAwsSubnet ...
// TODO right now this method is used by create subnet api to get created result, because sync method do not return it.
// TODO modify sync logics to return crud infos, then change this method to 'batchCreateAwsSubnet'.
func BatchCreateAwsSubnet(kt *kit.Kit, createResources []cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt],
	dataCli *dataclient.Client, adaptor *cloudclient.CloudAdaptorClient, req *SyncAwsOption) (
	*core.BatchCreateResult, error) {

	cloudVpcIDs := make([]string, 0, len(createResources))
	for _, one := range createResources {
		cloudVpcIDs = append(cloudVpcIDs, one.CloudVpcID)
	}

	opt := &logics.QueryVpcIDsAndSyncOption{
		Vendor:      enumor.Aws,
		AccountID:   req.AccountID,
		CloudVpcIDs: cloudVpcIDs,
		Region:      req.Region,
	}
	vpcMap, err := logics.QueryVpcIDsAndSync(kt, adaptor, dataCli, opt)
	if err != nil {
		logs.Errorf("query vpcIDs and sync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for index, resource := range createResources {
		one, exist := vpcMap[resource.CloudVpcID]
		if !exist {
			return nil, fmt.Errorf("vpc: %s not sync from cloud", resource.CloudVpcID)
		}

		createResources[index].VpcID = one
	}

	createReq := &cloud.SubnetBatchCreateReq[cloud.AwsSubnetCreateExt]{
		Subnets: createResources,
	}

	return dataCli.Aws.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq)
}

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
	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/data-service/cloud"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
)

// SyncGcpOption define gcp sync option.
type SyncGcpOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
	SelfLinks []string `json:"self_links" validate:"omitempty"`
}

// Validate SyncGcpOption.
func (opt SyncGcpOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.SelfLinks) == 0 && len(opt.CloudIDs) == 0 {
		return errors.New("self_links or cloud_ids is required")
	}

	if len(opt.SelfLinks) != 0 && len(opt.CloudIDs) != 0 {
		return errors.New("self_links or cloud_ids only one can be set")
	}

	if len(opt.CloudIDs) > int(core.DefaultMaxPageLimit) {
		return fmt.Errorf("cloudIDs should <= %d", core.DefaultMaxPageLimit)
	}

	if len(opt.SelfLinks) > int(core.DefaultMaxPageLimit) {
		return fmt.Errorf("selfLinks should <= %d", core.DefaultMaxPageLimit)
	}

	return nil
}

// GcpSubnetSync sync gcp cloud subnet.
func GcpSubnetSync(kt *kit.Kit, opt *SyncGcpOption, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client) (interface{}, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	list, err := listGcpSubnetFromCloud(kt, opt, adaptor)
	if err != nil {
		logs.Errorf("list gcp subnet from cloud failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	resourceDBMap, err := listGcpSubnetMapFromDB(kt, opt, dataCli)
	if err != nil {
		logs.Errorf("list gcp subnet from db failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	if len(list.Details) == 0 && len(resourceDBMap) == 0 {
		return nil, nil
	}

	err = diffGcpSubnetAndSync(kt, opt, list, resourceDBMap, dataCli, adaptor)
	if err != nil {
		logs.Errorf("diff gcp subnet and sync failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	return nil, nil
}

func listGcpSubnetMapFromDB(kt *kit.Kit, opt *SyncGcpOption, dataCli *dataclient.Client) (
	map[string]cloudcore.Subnet[cloudcore.GcpSubnetExtension], error) {

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
		},
	}

	if len(opt.CloudIDs) != 0 {
		expr.Rules = append(expr.Rules, &filter.AtomRule{
			Field: "cloud_id",
			Op:    filter.In.Factory(),
			Value: opt.CloudIDs,
		})
	}

	if len(opt.SelfLinks) != 0 {
		expr.Rules = append(expr.Rules, &filter.AtomRule{
			Field: "extension.self_link",
			Op:    filter.JSONIn.Factory(),
			Value: opt.SelfLinks,
		})
	}

	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   core.DefaultBasePage,
	}
	dbList, err := dataCli.Gcp.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		return nil, err
	}

	resourceMap := make(map[string]cloudcore.Subnet[cloudcore.GcpSubnetExtension], len(dbList.Details))
	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// listGcpSubnetFromCloud batch get subnet list from cloudapi.
func listGcpSubnetFromCloud(kt *kit.Kit, req *SyncGcpOption, adaptor *cloudclient.CloudAdaptorClient) (
	*types.GcpSubnetListResult, error) {

	cli, err := adaptor.Gcp(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.GcpSubnetListOption{
		Region: req.Region,
	}

	if len(req.CloudIDs) > 0 {
		opt.CloudIDs = req.CloudIDs
	}

	if len(req.SelfLinks) > 0 {
		opt.SelfLinks = req.SelfLinks
	}

	result, err := cli.ListSubnet(kt, opt)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// diffGcpSubnetAndSync batch sync vendor subnet list.
func diffGcpSubnetAndSync(kt *kit.Kit, req *SyncGcpOption, list *types.GcpSubnetListResult,
	resourceDBMap map[string]cloudcore.Subnet[cloudcore.GcpSubnetExtension], dataCli *dataclient.Client,
	adaptor *cloudclient.CloudAdaptorClient) error {

	createResources, updateResources, delCloudIDs, err := diffGcpSubnet(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	if len(updateResources) > 0 {
		updateReq := &cloud.SubnetBatchUpdateReq[cloud.GcpSubnetUpdateExt]{
			Subnets: updateResources,
		}
		if err = dataCli.Gcp.Subnet.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			return err
		}
	}

	if len(createResources) > 0 {
		_, err = BatchCreateGcpSubnet(kt, createResources, dataCli, adaptor, req)
		if err != nil {
			return err
		}
	}

	// delete resource data
	if len(delCloudIDs) > 0 {
		delListOpt := &SyncGcpOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  delCloudIDs,
		}
		delResult, err := listGcpSubnetFromCloud(kt, delListOpt, adaptor)
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

// diffGcpSubnet filter gcp subnet list
func diffGcpSubnet(req *SyncGcpOption, list *types.GcpSubnetListResult,
	resourceDBMap map[string]cloudcore.Subnet[cloudcore.GcpSubnetExtension]) (
	[]cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt],
	[]cloud.SubnetUpdateReq[cloud.GcpSubnetUpdateExt], []string, error) {

	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi subnetlist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	createResources := make([]cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt], 0)
	updateResources := make([]cloud.SubnetUpdateReq[cloud.GcpSubnetUpdateExt], 0)
	for _, item := range list.Details {
		// need compare and update subnet data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if isGcpSubnetChange(resourceInfo, item) {
				tmpRes := cloud.SubnetUpdateReq[cloud.GcpSubnetUpdateExt]{
					ID: resourceInfo.ID,
					SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
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

				updateResources = append(updateResources, tmpRes)
			}

			delete(resourceDBMap, item.CloudID)
		} else {
			// need add subnet data
			tmpRes := cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt]{
				AccountID:  req.AccountID,
				CloudVpcID: "",
				VpcID:      "",
				BkBizID:    constant.UnassignedBiz,
				CloudID:    item.CloudID,
				Name:       converter.ValToPtr(item.Name),
				Region:     item.Extension.Region,
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

			createResources = append(createResources, tmpRes)
		}
	}

	deleteCloudIDs := make([]string, 0, len(resourceDBMap))
	for _, vpc := range resourceDBMap {
		deleteCloudIDs = append(deleteCloudIDs, vpc.CloudID)
	}

	return createResources, updateResources, deleteCloudIDs, nil
}

func isGcpSubnetChange(info cloudcore.Subnet[cloudcore.GcpSubnetExtension], item types.GcpSubnet) bool {
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

// BatchCreateGcpSubnet ...
// TODO right now this method is used by create subnet api to get created result, because sync method do not return it.
// TODO modify sync logics to return crud infos, then change this method to 'batchCreateGcpSubnet'.
func BatchCreateGcpSubnet(kt *kit.Kit, createResources []cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt],
	dataCli *dataclient.Client, adaptor *cloudclient.CloudAdaptorClient, req *SyncGcpOption) (
	*core.BatchCreateResult, error) {

	selfLinks := make([]string, 0, len(createResources))
	for _, one := range createResources {
		selfLinks = append(selfLinks, one.Extension.VpcSelfLink)
	}

	vpcMap, err := logics.QueryVpcIDsAndSyncForGcp(kt, adaptor, dataCli, req.AccountID, selfLinks)
	if err != nil {
		logs.Errorf("query vpcIDs and sync for gcp failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for index, resource := range createResources {
		one, exist := vpcMap[resource.Extension.VpcSelfLink]
		if !exist {
			return nil, fmt.Errorf("vpc: %s not sync from cloud", resource.CloudVpcID)
		}

		createResources[index].VpcID = one.ID
		createResources[index].CloudVpcID = one.CloudID
	}

	createReq := &cloud.SubnetBatchCreateReq[cloud.GcpSubnetCreateExt]{
		Subnets: createResources,
	}
	return dataCli.Gcp.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq)
}

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
	hcservice "hcm/pkg/api/hc-service"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// AzureSubnetSync sync azure cloud subnet.
func AzureSubnetSync(kt *kit.Kit, req *hcservice.AzureResourceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	// batch get subnet list from cloudapi.
	list, err := BatchGetAzureSubnetList(kt, req, adaptor)
	if err != nil {
		logs.Errorf("%s-subnet request cloudapi response failed. accountID: %s, rgName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	// get vpc info from db for azure
	if len(req.CloudVpcID) == 0 {
		return nil, errors.New("cloud_vpc_id is required")
	}

	// batch get subnet map from db.
	resourceDBMap, err := listAzureSubnetMapFromDB(kt, req.CloudIDs, req.CloudVpcID, dataCli)
	if err != nil {
		logs.Errorf("%s-subnet batch get subnetdblist failed. accountID: %s, rgName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	// batch sync vendor subnet list.
	err = BatchSyncAzureSubnetList(kt, req, list, resourceDBMap, dataCli, adaptor)
	if err != nil {
		logs.Errorf("%s-subnet compare api and dblist failed. accountID: %s, rgName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

func listAzureSubnetMapFromDB(kt *kit.Kit, cloudIDs []string, cloudVpcID string,
	dataCli *dataclient.Client) (map[string]cloudcore.Subnet[cloudcore.AzureSubnetExtension], error) {

	rulesCommon := make([]filter.RuleFactory, 0)

	if cloudVpcID != "" {
		rulesCommon = append(rulesCommon, &filter.AtomRule{
			Field: "cloud_vpc_id",
			Op:    filter.Equal.Factory(),
			Value: cloudVpcID,
		})
	}

	if len(cloudIDs) > 0 {
		rulesCommon = append(rulesCommon, &filter.AtomRule{
			Field: "cloud_id",
			Op:    filter.In.Factory(),
			Value: cloudIDs,
		})
	}

	page := uint32(0)
	count := core.DefaultMaxPageLimit
	resourceMap := make(map[string]cloudcore.Subnet[cloudcore.AzureSubnetExtension], 0)
	for {
		offset := page * uint32(count)
		expr := &filter.Expression{
			Op:    filter.And,
			Rules: rulesCommon,
		}
		dbQueryReq := &core.ListReq{
			Filter: expr,
			Page:   &core.BasePage{Count: false, Start: offset, Limit: count},
		}
		dbList, err := dataCli.Azure.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), dbQueryReq)
		if err != nil {
			logs.Errorf("azure-subnet batch list db error. offset: %d, limit: %d, err: %v", offset, count, err)
			return nil, err
		}

		if len(dbList.Details) == 0 {
			return resourceMap, nil
		}

		for _, item := range dbList.Details {
			resourceMap[item.CloudID] = item
		}

		if len(dbList.Details) < int(count) {
			break
		}
		page++
	}

	return resourceMap, nil
}

// BatchGetAzureSubnetList batch get subnet list from cloudapi.
func BatchGetAzureSubnetList(kt *kit.Kit, req *hcservice.AzureResourceSyncReq, adaptor *cloudclient.CloudAdaptorClient) (
	*types.AzureSubnetListResult, error) {

	// 查询指定CloudIDs
	if len(req.CloudIDs) > 0 {
		return BatchGetAzureSubnetListByCloudIDs(kt, req, adaptor)
	}

	return BatchGetAzureSubnetAllList(kt, req, adaptor)
}

// BatchGetAzureSubnetListByCloudIDs batch get subnet list from cloudapi.
func BatchGetAzureSubnetListByCloudIDs(kt *kit.Kit, req *hcservice.AzureResourceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient) (*types.AzureSubnetListResult, error) {

	cli, err := adaptor.Azure(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.AzureSubnetListByIDOption{
		VpcID: req.CloudVpcID,
	}
	opt.ResourceGroupName = req.ResourceGroupName
	opt.CloudIDs = req.CloudIDs

	list, err := cli.ListSubnetByID(kt, opt)
	if err != nil {
		logs.Errorf("%s-subnet batch get cloud api failed. accountID: %s, rgName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	return list, nil
}

// BatchGetAzureSubnetAllList batch get subnet list from cloudapi.
func BatchGetAzureSubnetAllList(kt *kit.Kit, req *hcservice.AzureResourceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient) (*types.AzureSubnetListResult, error) {

	cli, err := adaptor.Azure(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.AzureSubnetListOption{
		VpcID: req.CloudVpcID,
	}
	opt.ResourceGroupName = req.ResourceGroupName

	list, err := cli.ListSubnet(kt, opt)
	if err != nil {
		logs.Errorf("%s-subnet batch get cloud api failed. accountID: %s, rgName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	return list, nil
}

// BatchSyncAzureSubnetList batch sync vendor subnet list.
func BatchSyncAzureSubnetList(kt *kit.Kit, req *hcservice.AzureResourceSyncReq, list *types.AzureSubnetListResult,
	resourceDBMap map[string]cloudcore.Subnet[cloudcore.AzureSubnetExtension], dataCli *dataclient.Client,
	adaptor *cloudclient.CloudAdaptorClient) error {

	createResources, updateResources, existIDMap, err := filterAzureSubnetList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.SubnetBatchUpdateReq[cloud.AzureSubnetUpdateExt]{
			Subnets: updateResources,
		}
		if err = dataCli.Azure.Subnet.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-subnet batch compare db update failed. accountID: %s, rgName: %s, err: %v",
				enumor.Azure, req.AccountID, req.ResourceGroupName, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		err = batchCreateAzureSubnet(kt, createResources, dataCli, adaptor, req)
		if err != nil {
			logs.Errorf("%s-subnet batch compare db create failed. accountID: %s, rgName: %s, err: %v",
				enumor.Azure, req.AccountID, req.ResourceGroupName, err)
			return err
		}
	}

	// delete resource data
	deleteIDs := make([]string, 0)
	for _, resItem := range resourceDBMap {
		if _, ok := existIDMap[resItem.ID]; !ok {
			deleteIDs = append(deleteIDs, resItem.ID)
		}
	}

	if len(deleteIDs) > 0 {
		err = BatchDeleteSubnetByIDs(kt, deleteIDs, dataCli)
		if err != nil {
			logs.Errorf("%s-subnet batch compare db delete failed. accountID: %s, rgName: %s, delIDs: %v, "+
				"err: %v", enumor.Azure, req.AccountID, req.ResourceGroupName, deleteIDs, err)
			return err
		}
	}

	return nil
}

// filterAzureSubnetList filter azure subnet list
func filterAzureSubnetList(req *hcservice.AzureResourceSyncReq, list *types.AzureSubnetListResult,
	resourceDBMap map[string]cloudcore.Subnet[cloudcore.AzureSubnetExtension]) (
	createResources []cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt],
	updateResources []cloud.SubnetUpdateReq[cloud.AzureSubnetUpdateExt], existIDMap map[string]bool, err error) {

	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi subnetlist is empty, accountID: %s, rgName: %s",
				req.AccountID, req.ResourceGroupName)
	}

	existIDMap = make(map[string]bool, 0)
	for _, item := range list.Details {
		// need compare and update subnet data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if isAzureSubnetChange(resourceInfo, item) {
				tmpRes := cloud.SubnetUpdateReq[cloud.AzureSubnetUpdateExt]{
					ID: resourceInfo.ID,
					SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
						Name:              converter.ValToPtr(item.Name),
						Ipv4Cidr:          item.Ipv4Cidr,
						Ipv6Cidr:          item.Ipv6Cidr,
						Memo:              item.Memo,
						CloudRouteTableID: nil,
						RouteTableID:      nil,
					},
					Extension: &cloud.AzureSubnetUpdateExt{
						NatGateway:           converter.ValToPtr(item.Extension.NatGateway),
						CloudSecurityGroupID: converter.ValToPtr(item.Extension.NetworkSecurityGroup),
						SecurityGroupID:      nil,
					},
				}

				updateResources = append(updateResources, tmpRes)
			}

			existIDMap[resourceInfo.ID] = true

		} else {
			// need add subnet data
			tmpRes := cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt]{
				AccountID:         req.AccountID,
				CloudVpcID:        item.CloudVpcID,
				VpcID:             "",
				BkBizID:           constant.UnassignedBiz,
				CloudRouteTableID: "",
				RouteTableID:      "",
				CloudID:           item.CloudID,
				Name:              converter.ValToPtr(item.Name),
				Region:            "",
				Zone:              "",
				Ipv4Cidr:          item.Ipv4Cidr,
				Ipv6Cidr:          item.Ipv6Cidr,
				Memo:              item.Memo,
				Extension: &cloud.AzureSubnetCreateExt{
					ResourceGroupName:    item.Extension.ResourceGroupName,
					NatGateway:           item.Extension.NatGateway,
					CloudSecurityGroupID: item.Extension.NetworkSecurityGroup,
					SecurityGroupID:      "",
				},
			}

			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, existIDMap, nil
}

func isAzureSubnetChange(info cloudcore.Subnet[cloudcore.AzureSubnetExtension], item types.AzureSubnet) bool {
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

	if info.Extension.NetworkSecurityGroup != item.Extension.NetworkSecurityGroup {
		return true
	}

	return false
}

func batchCreateAzureSubnet(kt *kit.Kit, createResources []cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt],
	dataCli *dataclient.Client, adaptor *cloudclient.CloudAdaptorClient, req *hcservice.AzureResourceSyncReq) error {

	querySize := int(filter.DefaultMaxInLimit)
	times := len(createResources) / querySize
	if len(createResources)%querySize != 0 {
		times++
	}

	for i := 0; i < times; i++ {
		var newResources []cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt]
		if i == times-1 {
			newResources = append(newResources, createResources[i*querySize:]...)
		} else {
			newResources = append(newResources, createResources[i*querySize:(i+1)*querySize]...)
		}

		opt := &logics.QueryVpcIDsAndSyncOption{
			Vendor:            enumor.Azure,
			AccountID:         req.AccountID,
			CloudVpcIDs:       []string{req.CloudVpcID},
			ResourceGroupName: req.ResourceGroupName,
		}
		vpcMap, err := logics.QueryVpcIDsAndSync(kt, adaptor, dataCli, opt)
		if err != nil {
			logs.Errorf("query vpcIDs and sync failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		for index, resource := range newResources {
			one, exist := vpcMap[resource.CloudVpcID]
			if !exist {
				return fmt.Errorf("vpc: %s not sync from cloud", resource.CloudVpcID)
			}

			newResources[index].VpcID = one
		}

		createReq := &cloud.SubnetBatchCreateReq[cloud.AzureSubnetCreateExt]{
			Subnets: newResources,
		}
		if _, err := dataCli.Azure.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			return err
		}
	}

	return nil
}

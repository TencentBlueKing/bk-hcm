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
	securitygrouplogics "hcm/cmd/hc-service/logics/sync/logics/security-group"
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
)

// AzureSubnetSync sync azure cloud subnet.
func AzureSubnetSync(kt *kit.Kit, opt *hcservice.AzureResourceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	if len(opt.CloudVpcID) == 0 {
		return nil, errors.New("cloud_vpc_id is required")
	}

	list, err := listAzureSubnetFromCloud(kt, opt, adaptor)
	if err != nil {
		logs.Errorf("list azure subnet from cloud failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	resourceDBMap, err := listAzureSubnetMapFromDB(kt, opt, dataCli)
	if err != nil {
		logs.Errorf("list azure subnet from db failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	if len(list.Details) == 0 && len(resourceDBMap) == 0 {
		return nil, nil
	}

	err = diffAzureSubnetAndSync(kt, opt, list, resourceDBMap, dataCli, adaptor)
	if err != nil {
		logs.Errorf("diff azure subnet and sync failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	return nil, nil
}

func listAzureSubnetMapFromDB(kt *kit.Kit, opt *hcservice.AzureResourceSyncReq,
	dataCli *dataclient.Client) (map[string]cloudcore.Subnet[cloudcore.AzureSubnetExtension], error) {

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "account_id",
				Op:    filter.Equal.Factory(),
				Value: opt.AccountID,
			},
			&filter.AtomRule{
				Field: "cloud_vpc_id",
				Op:    filter.Equal.Factory(),
				Value: opt.CloudVpcID,
			},
			&filter.AtomRule{
				Field: "extension.resource_group_name",
				Op:    filter.JSONEqual.Factory(),
				Value: opt.ResourceGroupName,
			},
		},
	}

	if len(opt.CloudIDs) > 0 {
		expr.Rules = append(expr.Rules, &filter.AtomRule{
			Field: "cloud_id",
			Op:    filter.In.Factory(),
			Value: opt.CloudIDs,
		})
	}

	page := uint32(0)
	count := core.DefaultMaxPageLimit
	resourceMap := make(map[string]cloudcore.Subnet[cloudcore.AzureSubnetExtension], 0)
	for {
		offset := page * uint32(count)
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

// listAzureSubnetFromCloud batch get subnet list from cloudapi.
func listAzureSubnetFromCloud(kt *kit.Kit, req *hcservice.AzureResourceSyncReq, adaptor *cloudclient.CloudAdaptorClient) (
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

// diffAzureSubnetAndSync batch sync vendor subnet list.
func diffAzureSubnetAndSync(kt *kit.Kit, req *hcservice.AzureResourceSyncReq, list *types.AzureSubnetListResult,
	resourceDBMap map[string]cloudcore.Subnet[cloudcore.AzureSubnetExtension], dataCli *dataclient.Client,
	adaptor *cloudclient.CloudAdaptorClient) error {

	createResources, updateResources, existIDMap, err := diffAzureSubnet(kt, dataCli, req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// delete resource data
	deleteIDs := make([]string, 0)
	for _, resItem := range resourceDBMap {
		if _, ok := existIDMap[resItem.ID]; !ok {
			deleteIDs = append(deleteIDs, resItem.ID)
		}
	}

	if len(deleteIDs) > 0 {
		delListOpt := &hcservice.AzureResourceSyncReq{
			AccountID:         req.AccountID,
			ResourceGroupName: req.ResourceGroupName,
			CloudVpcID:        req.CloudVpcID,
			CloudIDs:          deleteIDs,
		}
		delResult, err := listAzureSubnetFromCloud(kt, delListOpt, adaptor)
		if err != nil {
			return err
		}

		if len(delResult.Details) > 0 {
			logs.Errorf("validate subnet not exist failed, before delete, opt: %v, rid: %s", delListOpt, kt.Rid)
			return fmt.Errorf("validate subnet not exist failed, before delete")
		}

		if err = batchDeleteSubnetByIDs(kt, deleteIDs, dataCli); err != nil {
			return err
		}
	}

	// update resource data
	if len(updateResources) > 0 {
		if err := batchUpdateAzureSubnet(kt, updateResources, dataCli, adaptor, req); err != nil {
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		_, err = BatchCreateAzureSubnet(kt, createResources, dataCli, adaptor, req)
		if err != nil {
			return err
		}
	}

	return nil
}

func batchUpdateAzureSubnet(kt *kit.Kit, updateResources []cloud.SubnetUpdateReq[cloud.AzureSubnetUpdateExt],
	dataCli *dataclient.Client, adaptor *cloudclient.CloudAdaptorClient, req *hcservice.AzureResourceSyncReq) error {

	cloudSGIDs := make([]string, 0)
	for _, one := range updateResources {
		if len(converter.PtrToVal(one.Extension.CloudSecurityGroupID)) != 0 {
			cloudSGIDs = append(cloudSGIDs, *one.Extension.CloudSecurityGroupID)
		}
	}

	listSGOpt := &securitygrouplogics.QuerySecurityGroupIDsAndSyncOption{
		Vendor:                enumor.Azure,
		AccountID:             req.AccountID,
		ResourceGroupName:     req.ResourceGroupName,
		CloudSecurityGroupIDs: cloudSGIDs,
	}
	securityGroupMap, err := securitygrouplogics.QuerySecurityGroupIDsAndSync(kt, adaptor, dataCli, listSGOpt)
	if err != nil {
		return err
	}

	for index, resource := range updateResources {
		if len(converter.PtrToVal(resource.Extension.CloudSecurityGroupID)) != 0 {
			sgID, exist := securityGroupMap[*resource.Extension.CloudSecurityGroupID]
			if !exist {
				return fmt.Errorf("security group: %s not sync from cloud", *resource.Extension.CloudSecurityGroupID)
			}

			updateResources[index].Extension.SecurityGroupID = converter.ValToPtr(sgID)
		}
	}

	updateReq := &cloud.SubnetBatchUpdateReq[cloud.AzureSubnetUpdateExt]{
		Subnets: updateResources,
	}
	if err := dataCli.Azure.Subnet.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
		return err
	}

	return nil
}

// diffAzureSubnet filter azure subnet list
func diffAzureSubnet(kt *kit.Kit, dataCli *dataclient.Client, req *hcservice.AzureResourceSyncReq,
	list *types.AzureSubnetListResult, resourceDBMap map[string]cloudcore.Subnet[cloudcore.AzureSubnetExtension]) (
	createResources []cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt],
	updateResources []cloud.SubnetUpdateReq[cloud.AzureSubnetUpdateExt], existIDMap map[string]bool, err error) {

	existIDMap = make(map[string]bool, 0)
	for _, item := range list.Details {
		vpcData, err := getAzureVpcDataFromDB(kt, dataCli, req, item.CloudVpcID)
		if err != nil {
			continue
		}
		// need compare and update subnet data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if isAzureSubnetChange(resourceInfo, item, vpcData.Region) {
				tmpRes := cloud.SubnetUpdateReq[cloud.AzureSubnetUpdateExt]{
					ID: resourceInfo.ID,
					SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
						Region:   vpcData.Region,
						Name:     converter.ValToPtr(item.Name),
						Ipv4Cidr: item.Ipv4Cidr,
						Ipv6Cidr: item.Ipv6Cidr,
						Memo:     item.Memo,
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
				AccountID:  req.AccountID,
				CloudVpcID: item.CloudVpcID,
				VpcID:      vpcData.ID,
				BkBizID:    constant.UnassignedBiz,
				CloudID:    item.CloudID,
				Name:       converter.ValToPtr(item.Name),
				Region:     vpcData.Region,
				Zone:       "",
				Ipv4Cidr:   item.Ipv4Cidr,
				Ipv6Cidr:   item.Ipv6Cidr,
				Memo:       item.Memo,
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

func isAzureSubnetChange(info cloudcore.Subnet[cloudcore.AzureSubnetExtension], item types.AzureSubnet,
	region string) bool {
	if info.Region != region {
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

	if info.Extension.ResourceGroupName != item.Extension.ResourceGroupName {
		return true
	}

	if info.Extension.NatGateway != item.Extension.NatGateway {
		return true
	}

	if info.Extension.SecurityGroupID != item.Extension.NetworkSecurityGroup {
		return true
	}

	return false
}

// BatchCreateAzureSubnet ...
// TODO right now this method is used by create subnet api to get created result, because sync method do not return it.
// TODO modify sync logics to return crud infos, then change this method to 'batchCreateAzureSubnet'.
func BatchCreateAzureSubnet(kt *kit.Kit, createResources []cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt],
	dataCli *dataclient.Client, adaptor *cloudclient.CloudAdaptorClient, req *hcservice.AzureResourceSyncReq) (
	*core.BatchCreateResult, error) {

	querySize := int(filter.DefaultMaxInLimit)
	times := len(createResources) / querySize
	if len(createResources)%querySize != 0 {
		times++
	}

	listVpcOpt := &logics.QueryVpcIDsAndSyncOption{
		Vendor:            enumor.Azure,
		AccountID:         req.AccountID,
		CloudVpcIDs:       []string{req.CloudVpcID},
		ResourceGroupName: req.ResourceGroupName,
	}
	vpcMap, err := logics.QueryVpcIDsAndSync(kt, adaptor, dataCli, listVpcOpt)
	if err != nil {
		logs.Errorf("query vpcIDs and sync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	createRes := &core.BatchCreateResult{IDs: make([]string, 0)}
	for i := 0; i < times; i++ {
		var newResources []cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt]
		if i == times-1 {
			newResources = append(newResources, createResources[i*querySize:]...)
		} else {
			newResources = append(newResources, createResources[i*querySize:(i+1)*querySize]...)
		}

		cloudSGIDs := make([]string, 0)
		for _, one := range newResources {
			if len(one.Extension.CloudSecurityGroupID) != 0 {
				cloudSGIDs = append(cloudSGIDs, one.Extension.CloudSecurityGroupID)
			}
		}

		listSGOpt := &securitygrouplogics.QuerySecurityGroupIDsAndSyncOption{
			Vendor:                enumor.Azure,
			AccountID:             req.AccountID,
			ResourceGroupName:     req.ResourceGroupName,
			CloudSecurityGroupIDs: cloudSGIDs,
		}
		securityGroupMap, err := securitygrouplogics.QuerySecurityGroupIDsAndSync(kt, adaptor, dataCli, listSGOpt)
		if err != nil {
			return nil, err
		}

		for index, resource := range newResources {
			vpcID, exist := vpcMap[resource.CloudVpcID]
			if !exist {
				return nil, fmt.Errorf("vpc: %s not sync from cloud", resource.CloudVpcID)
			}
			newResources[index].VpcID = vpcID

			if len(resource.Extension.CloudSecurityGroupID) != 0 {
				sgID, exist := securityGroupMap[resource.Extension.CloudSecurityGroupID]
				if !exist {
					return nil, fmt.Errorf("security group: %s not sync from cloud", resource.Extension.CloudSecurityGroupID)
				}
				newResources[index].Extension.SecurityGroupID = sgID
			}
		}

		createReq := &cloud.SubnetBatchCreateReq[cloud.AzureSubnetCreateExt]{
			Subnets: newResources,
		}
		res, err := dataCli.Azure.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq)
		if err != nil {
			return nil, err
		}
		createRes.IDs = append(createRes.IDs, res.IDs...)
	}

	return createRes, nil
}

func getAzureVpcDataFromDB(kt *kit.Kit, dataCli *dataclient.Client, req *hcservice.AzureResourceSyncReq,
	cloudVpcID string) (cloudcore.Vpc[cloudcore.AzureVpcExtension], error) {

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "account_id",
				Op:    filter.Equal.Factory(),
				Value: req.AccountID,
			},
			&filter.AtomRule{
				Field: "cloud_id",
				Op:    filter.Equal.Factory(),
				Value: cloudVpcID,
			},
		},
	}

	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   core.DefaultBasePage,
	}
	dbList, err := dataCli.Azure.Vpc.ListVpcExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		return cloudcore.Vpc[cloudcore.AzureVpcExtension]{}, err
	}

	if len(dbList.Details) <= 0 {
		return cloudcore.Vpc[cloudcore.AzureVpcExtension]{}, fmt.Errorf("get vpc data from db is <= 0")
	}

	return dbList.Details[0], nil
}

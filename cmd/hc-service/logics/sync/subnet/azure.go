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
	"fmt"

	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// AzureSubnetSync sync azure cloud subnet.
func AzureSubnetSync(kt *kit.Kit, req *hcservice.ResourceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	if len(req.ResourceGroupName) == 0 {
		return nil, errf.New(errf.InvalidParameter, "resource_group_name is required")
	}

	if len(req.VpcID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vpc_id is required")
	}

	// batch get subnet list from cloudapi.
	list, err := BatchGetAzureSubnetList(kt, req, adaptor)
	if err != nil {
		logs.Errorf("%s-subnet request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.Azure, req.AccountID, req.Region, err)
		return nil, err
	}

	// get vpc info from db for azure
	vpcInfo, isVpcExist, err := GetVpcInfoFromDBForAzure(kt, req, enumor.Azure, dataCli)
	if err != nil {
		return nil, err
	}

	if !isVpcExist {
		return nil, errf.New(errf.InvalidParameter, "vpc info is not found")
	}

	if vpcInfo.CloudID == "" {
		return nil, errf.New(errf.InvalidParameter, "cloud_id is empty")
	}

	// batch get subnet map from db.
	resourceDBMap, err := BatchGetSubnetMapFromDB(kt, req, enumor.Azure, vpcInfo.CloudID, dataCli)
	if err != nil {
		logs.Errorf("%s-subnet batch get subnetdblist failed. accountID: %s, region: %s, err: %v",
			enumor.Azure, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch sync vendor subnet list.
	err = BatchSyncAzureSubnetList(kt, req, list, resourceDBMap, dataCli)
	if err != nil {
		logs.Errorf("%s-subnet compare api and dblist failed. accountID: %s, region: %s, err: %v",
			enumor.Azure, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// BatchGetAzureSubnetList batch get subnet list from cloudapi.
func BatchGetAzureSubnetList(kt *kit.Kit, req *hcservice.ResourceSyncReq, adaptor *cloudclient.CloudAdaptorClient) (
	*types.AzureSubnetListResult, error) {

	// 查询指定CloudIDs
	if len(req.CloudIDs) > 0 {
		return BatchGetAzureSubnetListByCloudIDs(kt, req, adaptor)
	}

	return BatchGetAzureSubnetAllList(kt, req, adaptor)
}

// BatchGetAzureSubnetListByCloudIDs batch get subnet list from cloudapi.
func BatchGetAzureSubnetListByCloudIDs(kt *kit.Kit, req *hcservice.ResourceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient) (*types.AzureSubnetListResult, error) {

	cli, err := adaptor.Azure(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.AzureSubnetListByIDOption{
		VpcID: req.VpcID,
	}
	opt.ResourceGroupName = req.ResourceGroupName
	opt.CloudIDs = req.CloudIDs

	list, err := cli.ListSubnetByID(kt, opt)
	if err != nil {
		logs.Errorf("%s-subnet batch get cloud api failed. accountID: %s, region: %s, err: %v",
			enumor.Azure, req.AccountID, req.Region, err)
		return nil, err
	}

	return list, nil
}

// BatchGetAzureSubnetAllList batch get subnet list from cloudapi.
func BatchGetAzureSubnetAllList(kt *kit.Kit, req *hcservice.ResourceSyncReq, adaptor *cloudclient.CloudAdaptorClient) (
	*types.AzureSubnetListResult, error) {
	cli, err := adaptor.Azure(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.AzureSubnetListOption{
		VpcID: req.VpcID,
	}
	opt.ResourceGroupName = req.ResourceGroupName

	list, err := cli.ListSubnet(kt, opt)
	if err != nil {
		logs.Errorf("%s-subnet batch get cloud api failed. accountID: %s, region: %s, err: %v",
			enumor.Azure, req.AccountID, req.Region, err)
		return nil, err
	}

	return list, nil
}

// GetVpcInfoFromDBForAzure get vpc info from db for azure.
func GetVpcInfoFromDBForAzure(kt *kit.Kit, req *hcservice.ResourceSyncReq, vendor enumor.Vendor,
	dataCli *dataclient.Client) (cloudcore.BaseVpc, bool, error) {

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "vendor",
				Op:    filter.Equal.Factory(),
				Value: vendor,
			},
			&filter.AtomRule{
				Field: "account_id",
				Op:    filter.Equal.Factory(),
				Value: req.AccountID,
			},
			&filter.AtomRule{
				Field: "cloud_id",
				Op:    filter.Equal.Factory(),
				Value: req.VpcID,
			},
		},
	}
	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}

	dbInfo, err := dataCli.Global.Vpc.List(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("%s-vpc batch get vpclist db error. accountID: %s, region: %s, err: %v",
			vendor, req.AccountID, req.Region, err)
		return cloudcore.BaseVpc{}, false, err
	}

	if len(dbInfo.Details) == 0 {
		return cloudcore.BaseVpc{}, false, nil
	}

	return dbInfo.Details[0], true, nil
}

// BatchSyncAzureSubnetList batch sync vendor subnet list.
func BatchSyncAzureSubnetList(kt *kit.Kit, req *hcservice.ResourceSyncReq, list *types.AzureSubnetListResult,
	resourceDBMap map[string]cloudcore.BaseSubnet, dataCli *dataclient.Client) error {

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
			logs.Errorf("%s-subnet batch compare db update failed. accountID: %s, region: %s, err: %v",
				enumor.Azure, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		err = batchCreateAzureSubnet(kt, createResources, dataCli)
		if err != nil {
			logs.Errorf("%s-subnet batch compare db create failed. accountID: %s, region: %s, err: %v",
				enumor.Azure, req.AccountID, req.Region, err)
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
			logs.Errorf("%s-subnet batch compare db delete failed. accountID: %s, region: %s, delIDs: %v, "+
				"err: %v", enumor.Azure, req.AccountID, req.Region, deleteIDs, err)
			return err
		}
	}

	return nil
}

// filterAzureSubnetList filter azure subnet list
func filterAzureSubnetList(req *hcservice.ResourceSyncReq, list *types.AzureSubnetListResult,
	resourceDBMap map[string]cloudcore.BaseSubnet) (createResources []cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt],
	updateResources []cloud.SubnetUpdateReq[cloud.AzureSubnetUpdateExt], existIDMap map[string]bool, err error) {

	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi subnetlist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	existIDMap = make(map[string]bool, 0)
	for _, item := range list.Details {
		// need compare and update subnet data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if resourceInfo.Name == item.Name && resourceInfo.CloudVpcID == item.CloudVpcID &&
				converter.PtrToVal(resourceInfo.Memo) == converter.PtrToVal(item.Memo) {
				existIDMap[resourceInfo.ID] = true
				continue
			}

			tmpRes := cloud.SubnetUpdateReq[cloud.AzureSubnetUpdateExt]{
				ID: resourceInfo.ID,
				Extension: &cloud.AzureSubnetUpdateExt{
					NatGateway:           converter.ValToPtr(item.Extension.NatGateway),
					CloudSecurityGroupID: converter.ValToPtr(item.Extension.NetworkSecurityGroup),
				},
			}
			tmpRes.Name = converter.ValToPtr(item.Name)
			tmpRes.Ipv4Cidr = item.Ipv4Cidr

			if len(item.Ipv6Cidr) > 0 {
				tmpRes.Ipv6Cidr = item.Ipv6Cidr
			} else {
				tmpRes.Ipv6Cidr = []string{""}
			}

			tmpRes.Memo = item.Memo
			updateResources = append(updateResources, tmpRes)
			existIDMap[resourceInfo.ID] = true
		} else {
			// need add subnet data
			tmpRes := cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt]{
				AccountID:  req.AccountID,
				CloudVpcID: item.CloudVpcID,
				CloudID:    item.CloudID,
				Name:       converter.ValToPtr(item.Name),
				Ipv4Cidr:   item.Ipv4Cidr,
				Memo:       item.Memo,
				Extension: &cloud.AzureSubnetCreateExt{
					ResourceGroup:        item.Extension.ResourceGroup,
					NatGateway:           item.Extension.NatGateway,
					NetworkSecurityGroup: item.Extension.NetworkSecurityGroup,
				},
			}

			if len(item.Ipv6Cidr) > 0 {
				tmpRes.Ipv6Cidr = item.Ipv6Cidr
			} else {
				tmpRes.Ipv6Cidr = []string{""}
			}

			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, existIDMap, nil
}

func batchCreateAzureSubnet(kt *kit.Kit, createResources []cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt],
	dataCli *dataclient.Client) error {

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

		createReq := &cloud.SubnetBatchCreateReq[cloud.AzureSubnetCreateExt]{
			Subnets: newResources,
		}
		if _, err := dataCli.Azure.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			return err
		}
	}

	return nil
}

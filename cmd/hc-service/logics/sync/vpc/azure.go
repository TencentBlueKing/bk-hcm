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

// Package vpc defines vpc service.
package vpc

import (
	"fmt"

	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// AzureVpcSync sync azure cloud vpc.
func AzureVpcSync(kt *kit.Kit, req *hcservice.AzureResourceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	// batch get vpc list from cloudapi.
	list, err := BatchGetAzureVpcList(kt, req, adaptor)
	if err != nil {
		logs.Errorf("%s-vpc request cloudapi response failed. accountID: %s, rgName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	// batch get vpc map from db.
	resourceDBMap, err := BatchGetVpcMapFromDB(kt, enumor.Azure, dataCli)
	if err != nil {
		logs.Errorf("%s-vpc batch get vpcdblist failed. accountID: %s, rgName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	// batch sync vendor vpc list.
	err = BatchSyncAzureVpcList(kt, req, list, resourceDBMap, adaptor, dataCli)
	if err != nil {
		logs.Errorf("%s-vpc compare api and dblist failed. accountID: %s, rgName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// BatchGetAzureVpcList batch get vpc list from cloudapi.
func BatchGetAzureVpcList(kt *kit.Kit, req *hcservice.AzureResourceSyncReq, adaptor *cloudclient.CloudAdaptorClient) (
	*types.AzureVpcListResult, error) {
	cli, err := adaptor.Azure(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adcore.AzureListOption{
		ResourceGroupName: req.ResourceGroupName,
	}
	list, err := cli.ListVpc(kt, opt)
	if err != nil {
		logs.Errorf("%s-vpc batch get cloud api failed. accountID: %s, rgName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	return list, nil
}

// BatchSyncAzureVpcList batch sync vendor vpc list.
func BatchSyncAzureVpcList(kt *kit.Kit, req *hcservice.AzureResourceSyncReq,
	list *types.AzureVpcListResult, resourceDBMap map[string]cloudcore.BaseVpc, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client) error {

	createResources, updateResources, existIDMap, err := filterAzureVpcList(kt, req, list, resourceDBMap,
		adaptor, dataCli)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.VpcBatchUpdateReq[cloud.AzureVpcUpdateExt]{
			Vpcs: updateResources,
		}
		if err = dataCli.Azure.Vpc.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-vpc batch compare db update failed. accountID: %s, rgName: %s, err: %v",
				enumor.Azure, req.AccountID, req.ResourceGroupName, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &cloud.VpcBatchCreateReq[cloud.AzureVpcCreateExt]{
			Vpcs: createResources,
		}
		if _, err = dataCli.Azure.Vpc.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("%s-vpc batch compare db create failed. accountID: %s, rgName: %s, err: %v",
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
		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("id", deleteIDs),
		}
		if err = dataCli.Global.Vpc.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
			logs.Errorf("%s-vpc batch compare db delete failed. accountID: %s, rgName: %s, delIDs: %v, err: %v",
				enumor.Azure, req.AccountID, req.ResourceGroupName, deleteIDs, err)
			return err
		}
	}

	return nil
}

// filterAzureVpcList filter azure vpc list
func filterAzureVpcList(kt *kit.Kit, req *hcservice.AzureResourceSyncReq, list *types.AzureVpcListResult,
	resourceDBMap map[string]cloudcore.BaseVpc, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client) (createResources []cloud.VpcCreateReq[cloud.AzureVpcCreateExt],
	updateResources []cloud.VpcUpdateReq[cloud.AzureVpcUpdateExt], existIDMap map[string]bool, err error) {

	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpclist is empty, accountID: %s, rgName: %s",
				req.AccountID, req.ResourceGroupName)
	}

	existIDMap = make(map[string]bool, 0)
	for _, item := range list.Details {
		// need compare and update vpc data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if resourceInfo.Name == item.Name && resourceInfo.Region == item.Region &&
				converter.PtrToVal(resourceInfo.Memo) == converter.PtrToVal(item.Memo) {
				existIDMap[resourceInfo.ID] = true
				continue
			}

			tmpRes := cloud.VpcUpdateReq[cloud.AzureVpcUpdateExt]{
				ID: resourceInfo.ID,
				Extension: &cloud.AzureVpcUpdateExt{
					DNSServers: item.Extension.DNSServers,
				},
			}
			tmpRes.Name = converter.ValToPtr(item.Name)
			tmpRes.Memo = item.Memo

			if item.Extension.Cidr != nil {
				tmpCidrs := []cloud.AzureCidr{}
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.AzureCidr{
						Type: cidrItem.Type,
						Cidr: cidrItem.Cidr,
					})
				}
				tmpRes.Extension.Cidr = tmpCidrs
			}

			updateResources = append(updateResources, tmpRes)
			existIDMap[resourceInfo.ID] = true

			// sync azure cloud subnet.
			req.VpcID = item.Name
			err = AzureSubnetSync(kt, req, item.CloudID, adaptor, dataCli)
			if err != nil {
				return nil, nil, nil, err
			}
		} else {
			// need add vpc data
			tmpRes := cloud.VpcCreateReq[cloud.AzureVpcCreateExt]{
				AccountID: req.AccountID,
				CloudID:   item.CloudID,
				Name:      converter.ValToPtr(item.Name),
				Region:    item.Region,
				Category:  enumor.BizVpcCategory,
				Memo:      item.Memo,
				Extension: &cloud.AzureVpcCreateExt{
					ResourceGroup: item.Extension.ResourceGroup,
					DNSServers:    item.Extension.DNSServers,
				},
			}

			if item.Extension.Cidr != nil {
				tmpCidrs := []cloud.AzureCidr{}
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.AzureCidr{
						Type: cidrItem.Type,
						Cidr: cidrItem.Cidr,
					})
				}
				tmpRes.Extension.Cidr = tmpCidrs
			}

			createResources = append(createResources, tmpRes)

			// sync azure cloud subnet.
			req.VpcID = item.Name
			err = AzureSubnetSync(kt, req, item.CloudID, adaptor, dataCli)
			if err != nil {
				return nil, nil, nil, err
			}
		}
	}

	return createResources, updateResources, existIDMap, nil
}

// AzureSubnetSync sync azure cloud subnet.
func AzureSubnetSync(kt *kit.Kit, req *hcservice.AzureResourceSyncReq, cloudVpcID string,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) error {

	// batch get subnet list from cloudapi.
	list, err := BatchGetAzureSubnetList(kt, req, adaptor)
	if err != nil {
		logs.Errorf("%s-vpc-subnet request cloudapi response failed. accountID: %s, cloudVpcID: %s, err: %v",
			enumor.Azure, req.AccountID, cloudVpcID, err)
		return err
	}

	if len(list.Details) == 0 {
		return nil
	}

	// batch get subnet map from db.
	resourceDBMap, err := BatchGetSubnetMapFromDB(kt, req, enumor.Azure, cloudVpcID, dataCli)
	if err != nil {
		logs.Errorf("%s-vpc-subnet batch get subnetdblist failed. accountID: %s, cloudVpcID: %s, err: %v",
			enumor.Azure, req.AccountID, cloudVpcID, err)
		return err
	}

	// batch sync vendor subnet list.
	err = BatchSyncAzureSubnetList(kt, req, list, resourceDBMap, dataCli)
	if err != nil {
		logs.Errorf("%s-vpc-subnet compare api and dblist failed. accountID: %s, cloudVpcID: %s, err: %v",
			enumor.Azure, req.AccountID, cloudVpcID, err)
		return err
	}

	return nil
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
		VpcID: req.VpcID,
	}
	opt.ResourceGroupName = req.ResourceGroupName
	opt.CloudIDs = req.CloudIDs

	list, err := cli.ListSubnetByID(kt, opt)
	if err != nil {
		logs.Errorf("%s-vpc-subnet batch get cloud api failed. accountID: %s, rgName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	return list, nil
}

// BatchGetAzureSubnetAllList batch get subnet list from cloudapi.
func BatchGetAzureSubnetAllList(kt *kit.Kit, req *hcservice.AzureResourceSyncReq, adaptor *cloudclient.CloudAdaptorClient) (
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
		logs.Errorf("%s-vpc-subnet batch get cloud api failed. accountID: %s, rgName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	return list, nil
}

// BatchGetSubnetMapFromDB batch get subnet map from db.
func BatchGetSubnetMapFromDB(kt *kit.Kit, req *hcservice.AzureResourceSyncReq, vendor enumor.Vendor, cloudVpcID string,
	dataCli *dataclient.Client) (map[string]cloudcore.BaseSubnet, error) {

	page := uint32(0)
	resourceMap := make(map[string]cloudcore.BaseSubnet, 0)
	for {
		count := core.DefaultMaxPageLimit
		offset := page * uint32(count)
		expr := &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: vendor,
				},
				&filter.AtomRule{
					Field: "cloud_vpc_id",
					Op:    filter.Equal.Factory(),
					Value: cloudVpcID,
				},
			},
		}
		dbQueryReq := &core.ListReq{
			Filter: expr,
			Page:   &core.BasePage{Count: false, Start: offset, Limit: count},
		}

		dbList, err := dataCli.Global.Subnet.List(kt.Ctx, kt.Header(), dbQueryReq)
		if err != nil {
			logs.Errorf("%s-vpc-subnet batch list db error. accountID: %s, rgName: %s, offset: %d, limit: %d, "+
				"cloudVpcID: %s, err: %v", vendor, req.AccountID, req.ResourceGroupName, offset, count, cloudVpcID, err)
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

// BatchSyncAzureSubnetList batch sync vendor subnet list.
func BatchSyncAzureSubnetList(kt *kit.Kit, req *hcservice.AzureResourceSyncReq, list *types.AzureSubnetListResult,
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
			logs.Errorf("%s-vpc-subnet batch compare db update failed. accountID: %s, rgName: %s, err: %v",
				enumor.Azure, req.AccountID, req.ResourceGroupName, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		err = batchCreateAzureSubnet(kt, createResources, dataCli)
		if err != nil {
			logs.Errorf("%s-vpc-subnet batch compare db create failed. accountID: %s, rgName: %s, err: %v",
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
		err = BatchDeleteVpcByIDs(kt, deleteIDs, dataCli)
		if err != nil {
			logs.Errorf("%s-vpc-subnet batch compare db delete failed. accountID: %s, rgName: %s, delIDs: %v, "+
				"err: %v", enumor.Azure, req.AccountID, req.ResourceGroupName, deleteIDs, err)
			return err
		}
	}
	return nil
}

func filterAzureSubnetList(req *hcservice.AzureResourceSyncReq, list *types.AzureSubnetListResult,
	resourceDBMap map[string]cloudcore.BaseSubnet) (createResources []cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt],
	updateResources []cloud.SubnetUpdateReq[cloud.AzureSubnetUpdateExt], existIDMap map[string]bool, err error) {
	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpcsubnetlist is empty, accountID: %s, rgName: %s",
				req.AccountID, req.ResourceGroupName)
	}

	existIDMap = make(map[string]bool, 0)
	for _, item := range list.Details {
		// need compare and update subnet data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
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

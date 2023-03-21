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
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
)

// AzureVpcSync sync azure cloud vpc.
func AzureVpcSync(kt *kit.Kit, opt *hcservice.AzureResourceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	list, err := listAzureVpcFromCloud(kt, opt, adaptor)
	if err != nil {
		logs.Errorf("list azure vpc from cloud failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	// batch get vpc map from db.
	resourceDBMap, err := listAzureVpcMapFromDB(kt, dataCli, &BatchGetVpcMapOption{
		AccountID:         opt.AccountID,
		ResourceGroupName: opt.ResourceGroupName,
		CloudIDs:          opt.CloudIDs,
	})
	if err != nil {
		logs.Errorf("list azure vpc from db failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	if len(list.Details) == 0 && len(resourceDBMap) == 0 {
		return nil, nil
	}

	createResources, updateResources, delCloudIDs, err := diffAzureVpc(opt, list, resourceDBMap)
	if err != nil {
		logs.Errorf("diff azure vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(delCloudIDs) > 0 {
		delListOpt := &hcservice.AzureResourceSyncReq{
			AccountID:         opt.AccountID,
			ResourceGroupName: opt.ResourceGroupName,
			CloudIDs:          delCloudIDs,
		}
		delResult, err := listAzureVpcFromCloud(kt, delListOpt, adaptor)
		if err != nil {
			logs.Errorf("list azure vpc failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
			return nil, err
		}

		if len(delResult.Details) > 0 {
			logs.Errorf("validate vpc not exist failed, before delete, opt: %v, rid: %s", opt, kt.Rid)
			return nil, fmt.Errorf("validate vpc not exist failed, before delete")
		}

		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
		}
		if err = dataCli.Global.Vpc.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
			logs.Errorf("batch delete db vpc failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(updateResources) > 0 {
		updateReq := &cloud.VpcBatchUpdateReq[cloud.AzureVpcUpdateExt]{
			Vpcs: updateResources,
		}
		if err = dataCli.Azure.Vpc.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("batch update db vpc failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(createResources) > 0 {
		createReq := &cloud.VpcBatchCreateReq[cloud.AzureVpcCreateExt]{
			Vpcs: createResources,
		}
		if _, err = dataCli.Azure.Vpc.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("batch create db vpc failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func listAzureVpcMapFromDB(kt *kit.Kit, dataCli *dataclient.Client, opt *BatchGetVpcMapOption) (
	map[string]cloudcore.Vpc[cloudcore.AzureVpcExtension], error) {

	resourceMap := make(map[string]cloudcore.Vpc[cloudcore.AzureVpcExtension], 0)
	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "account_id",
				Op:    filter.Equal.Factory(),
				Value: opt.AccountID,
			},
			&filter.AtomRule{
				Field: "extension.resource_group_name",
				Op:    filter.JSONEqual.Factory(),
				Value: opt.ResourceGroupName,
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

	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   core.DefaultBasePage,
	}
	dbList, err := dataCli.Azure.Vpc.ListVpcExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		return nil, err
	}

	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

func listAzureVpcFromCloud(kt *kit.Kit, req *hcservice.AzureResourceSyncReq, adaptor *cloudclient.CloudAdaptorClient) (
	*types.AzureVpcListResult, error) {
	cli, err := adaptor.Azure(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	if len(req.CloudIDs) == 0 {
		opt := &adcore.AzureListOption{
			ResourceGroupName: req.ResourceGroupName,
		}
		list, err := cli.ListVpc(kt, opt)
		if err != nil {
			return nil, err
		}

		return list, nil
	}

	opt := &adcore.AzureListByIDOption{
		ResourceGroupName: req.ResourceGroupName,
		CloudIDs:          req.CloudIDs,
	}
	list, err := cli.ListVpcByID(kt, opt)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// diffAzureVpc filter azure vpc list
func diffAzureVpc(req *hcservice.AzureResourceSyncReq, list *types.AzureVpcListResult,
	resourceDBMap map[string]cloudcore.Vpc[cloudcore.AzureVpcExtension]) ([]cloud.VpcCreateReq[cloud.AzureVpcCreateExt],
	[]cloud.VpcUpdateReq[cloud.AzureVpcUpdateExt], []string, error) {

	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpclist is empty, accountID: %s, rgName: %s",
				req.AccountID, req.ResourceGroupName)
	}

	createResources := make([]cloud.VpcCreateReq[cloud.AzureVpcCreateExt], 0)
	updateResources := make([]cloud.VpcUpdateReq[cloud.AzureVpcUpdateExt], 0)
	for _, item := range list.Details {
		// need compare and update vpc data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if isAzureVpcChange(resourceInfo, item) {
				tmpRes := cloud.VpcUpdateReq[cloud.AzureVpcUpdateExt]{
					ID: resourceInfo.ID,
					VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
						Name:      converter.ValToPtr(item.Name),
						Category:  "",
						Memo:      item.Memo,
						BkCloudID: 0,
						BkBizID:   0,
					},
					Extension: &cloud.AzureVpcUpdateExt{
						DNSServers: item.Extension.DNSServers,
					},
				}

				if item.Extension.Cidr != nil {
					tmpCidrs := make([]cloud.AzureCidr, 0, len(item.Extension.Cidr))
					for _, cidrItem := range item.Extension.Cidr {
						tmpCidrs = append(tmpCidrs, cloud.AzureCidr{
							Type: cidrItem.Type,
							Cidr: cidrItem.Cidr,
						})
					}
					tmpRes.Extension.Cidr = tmpCidrs
				}

				updateResources = append(updateResources, tmpRes)
			}

			delete(resourceDBMap, item.CloudID)
		} else {
			// need add vpc data
			tmpRes := cloud.VpcCreateReq[cloud.AzureVpcCreateExt]{
				AccountID: req.AccountID,
				CloudID:   item.CloudID,
				BkBizID:   constant.UnassignedBiz,
				BkCloudID: constant.UnbindBkCloudID,
				Name:      converter.ValToPtr(item.Name),
				Region:    item.Region,
				Category:  enumor.BizVpcCategory,
				Memo:      item.Memo,
				Extension: &cloud.AzureVpcCreateExt{
					ResourceGroupName: item.Extension.ResourceGroupName,
					DNSServers:        item.Extension.DNSServers,
				},
			}

			if item.Extension.Cidr != nil {
				tmpCidrs := make([]cloud.AzureCidr, 0, len(item.Extension.Cidr))
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.AzureCidr{
						Type: cidrItem.Type,
						Cidr: cidrItem.Cidr,
					})
				}
				tmpRes.Extension.Cidr = tmpCidrs
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

func isAzureVpcChange(info cloudcore.Vpc[cloudcore.AzureVpcExtension], item types.AzureVpc) bool {
	if info.Name != item.Name {
		return true
	}

	if info.Region != item.Region {
		return true
	}

	if !assert.IsPtrStringEqual(info.Memo, item.Memo) {
		return true
	}

	for _, db := range info.Extension.Cidr {
		for _, cloud := range item.Extension.Cidr {
			if db.Cidr != cloud.Cidr {
				return true
			}

			if db.Type != cloud.Type {
				return true
			}
		}
	}

	if info.Extension.ResourceGroupName != item.Extension.ResourceGroupName {
		return true
	}

	if !assert.IsStringSliceEqual(info.Extension.DNSServers, item.Extension.DNSServers) {
		return true
	}

	return false
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
		VpcID: req.CloudVpcID,
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
		VpcID: req.CloudVpcID,
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
					ResourceGroupName:    item.Extension.ResourceGroupName,
					NatGateway:           item.Extension.NatGateway,
					CloudSecurityGroupID: item.Extension.NetworkSecurityGroup,
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

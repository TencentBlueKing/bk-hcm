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
	adcore "hcm/pkg/adaptor/types/core"
	cloudcore "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// AwsSubnetSync sync aws cloud subnet.
func AwsSubnetSync(kt *kit.Kit, req *hcservice.AwsResourceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	// batch get subnet list from cloudapi.
	list, err := BatchGetAwsSubnetList(kt, req, adaptor)
	if err != nil {
		logs.Errorf("%s-subnet request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.Aws, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch get subnet map from db.
	resourceDBMap, err := BatchGetSubnetMapFromDB(kt, enumor.Aws, req.CloudIDs, "", dataCli)
	if err != nil {
		logs.Errorf("%s-subnet batch get vpcdblist failed. accountID: %s, region: %s, err: %v",
			enumor.Aws, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch sync vendor subnet list.
	err = BatchSyncAwsSubnetList(kt, req, list, resourceDBMap, dataCli)
	if err != nil {
		logs.Errorf("%s-subnet compare api and dblist failed. accountID: %s, region: %s, err: %v",
			enumor.Aws, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// BatchGetAwsSubnetList batch get subnet list from cloudapi.
func BatchGetAwsSubnetList(kt *kit.Kit, req *hcservice.AwsResourceSyncReq, adaptor *cloudclient.CloudAdaptorClient) (
	*types.AwsSubnetListResult, error) {

	cli, err := adaptor.Aws(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	list := new(types.AwsSubnetListResult)
	for {
		opt := new(adcore.AwsListOption)
		opt.Region = req.Region

		// 查询指定CloudIDs
		if len(req.CloudIDs) > 0 {
			opt.CloudIDs = req.CloudIDs
		} else {
			count := int64(adcore.AwsQueryLimit)
			opt.Page = &adcore.AwsPage{
				MaxResults: converter.ValToPtr(count),
			}

			if nextToken != "" {
				opt.Page.NextToken = converter.ValToPtr(nextToken)
			}
		}

		tmpList, tmpErr := cli.ListSubnet(kt, opt)
		if tmpErr != nil {
			logs.Errorf("%s-subnet batch get cloud api failed. accountID: %s, region: %s, nextToken: %s, "+
				"err: %v", enumor.Aws, req.AccountID, req.Region, nextToken, tmpErr)
			return nil, tmpErr
		}

		if len(tmpList.Details) == 0 {
			break
		}

		list.Details = append(list.Details, tmpList.Details...)
		if len(req.CloudIDs) > 0 || tmpList.NextToken == nil {
			break
		}

		nextToken = *tmpList.NextToken
	}

	return list, nil
}

// BatchSyncAwsSubnetList batch sync vendor subnet list.
func BatchSyncAwsSubnetList(kt *kit.Kit, req *hcservice.AwsResourceSyncReq, list *types.AwsSubnetListResult,
	resourceDBMap map[string]cloudcore.BaseSubnet, dataCli *dataclient.Client) error {

	createResources, updateResources, existIDMap, err := filterAwsSubnetList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.SubnetBatchUpdateReq[cloud.AwsSubnetUpdateExt]{
			Subnets: updateResources,
		}
		if err = dataCli.Aws.Subnet.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-subnet batch compare db update failed. accountID: %s, region: %s, err: %v",
				enumor.Aws, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		err = batchCreateAwsSubnet(kt, createResources, dataCli)
		if err != nil {
			logs.Errorf("%s-subnet batch compare db create failed. accountID: %s, region: %s, err: %v",
				enumor.Aws, req.AccountID, req.Region, err)
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
				"err: %v", enumor.Aws, req.AccountID, req.Region, deleteIDs, err)
			return err
		}
	}

	return nil
}

// filterAwsVpcList filter aws subnet list
func filterAwsSubnetList(req *hcservice.AwsResourceSyncReq, list *types.AwsSubnetListResult,
	resourceDBMap map[string]cloudcore.BaseSubnet) (createResources []cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt],
	updateResources []cloud.SubnetUpdateReq[cloud.AwsSubnetUpdateExt], existIDMap map[string]bool, err error) {

	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpclist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
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

			tmpRes := cloud.SubnetUpdateReq[cloud.AwsSubnetUpdateExt]{
				ID: resourceInfo.ID,
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
			tmpRes := cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt]{
				AccountID:  req.AccountID,
				CloudVpcID: item.CloudVpcID,
				CloudID:    item.CloudID,
				Name:       converter.ValToPtr(item.Name),
				Region:     item.Extension.Region,
				Zone:       item.Extension.Zone,
				Ipv4Cidr:   item.Ipv4Cidr,
				Memo:       item.Memo,
				Extension: &cloud.AwsSubnetCreateExt{
					State:                       item.Extension.State,
					IsDefault:                   item.Extension.IsDefault,
					MapPublicIpOnLaunch:         item.Extension.MapPublicIpOnLaunch,
					AssignIpv6AddressOnCreation: item.Extension.AssignIpv6AddressOnCreation,
					HostnameType:                item.Extension.HostnameType,
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

func batchCreateAwsSubnet(kt *kit.Kit, createResources []cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt],
	dataCli *dataclient.Client) error {

	querySize := int(filter.DefaultMaxInLimit)
	times := len(createResources) / querySize
	if len(createResources)%querySize != 0 {
		times++
	}
	for i := 0; i < times; i++ {
		var newResources []cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt]

		if i == times-1 {
			newResources = append(newResources, createResources[i*querySize:]...)
		} else {
			newResources = append(newResources, createResources[i*querySize:(i+1)*querySize]...)
		}

		createReq := &cloud.SubnetBatchCreateReq[cloud.AwsSubnetCreateExt]{
			Subnets: newResources,
		}

		if _, err := dataCli.Aws.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			return err
		}
	}

	return nil
}

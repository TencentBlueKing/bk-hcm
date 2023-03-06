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
	cloudcore "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// AwsVpcSync sync aws cloud vpc.
func AwsVpcSync(kt *kit.Kit, req *hcservice.ResourceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	if len(req.Region) == 0 {
		return nil, errf.New(errf.InvalidParameter, "region is required")
	}

	// batch get vpc list from cloudapi.
	list, err := BatchGetAwsVpcList(kt, req, adaptor)
	if err != nil {
		logs.Errorf("%s-vpc request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.Aws, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch get vpc map from db.
	resourceDBMap, err := BatchGetVpcMapFromDB(kt, req, enumor.Aws, dataCli)
	if err != nil {
		logs.Errorf("%s-vpc batch get vpcdblist failed. accountID: %s, region: %s, err: %v",
			enumor.Aws, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch sync vendor vpc list.
	err = BatchSyncAwsVpcList(kt, req, list, resourceDBMap, dataCli)
	if err != nil {
		logs.Errorf("%s-vpc compare api and dblist failed. accountID: %s, region: %s, err: %v",
			enumor.Aws, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// BatchGetAwsVpcList batch get vpc list from cloudapi.
func BatchGetAwsVpcList(kt *kit.Kit, req *hcservice.ResourceSyncReq, adaptor *cloudclient.CloudAdaptorClient) (
	*types.AwsVpcListResult, error) {

	cli, err := adaptor.Aws(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	list := new(types.AwsVpcListResult)
	for {
		opt := new(adcore.AwsListOption)
		opt.Region = req.Region

		// 查询指定CloudIDs
		if len(req.CloudIDs) > 0 {
			opt.CloudIDs = req.CloudIDs
		} else {
			opt.Page = &adcore.AwsPage{
				MaxResults: converter.ValToPtr(int64(adcore.AwsQueryLimit)),
			}

			if nextToken != "" {
				opt.Page.NextToken = converter.ValToPtr(nextToken)
			}
		}

		tmpList, tmpErr := cli.ListVpc(kt, opt)
		if tmpErr != nil {
			logs.Errorf("%s-vpc batch get cloud api failed. accountID: %s, region: %s, nextToken: %s, err: %v",
				enumor.Aws, req.AccountID, req.Region, nextToken, tmpErr)
			return nil, tmpErr
		}

		// traversal vpclist supply fields
		for _, item := range tmpList.Details {
			dnsHostnames, dnsSupport, dnsErr := cli.GetVpcAttribute(kt, item.CloudID, item.Region)
			if dnsErr == nil {
				item.Extension.EnableDnsHostnames = dnsHostnames
				item.Extension.EnableDnsSupport = dnsSupport
			}
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

// BatchSyncAwsVpcList batch sync vendor vpc list.
func BatchSyncAwsVpcList(kt *kit.Kit, req *hcservice.ResourceSyncReq, list *types.AwsVpcListResult,
	resourceDBMap map[string]cloudcore.BaseVpc, dataCli *dataclient.Client) error {

	createResources, updateResources, existIDMap, err := filterAwsVpcList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.VpcBatchUpdateReq[cloud.AwsVpcUpdateExt]{
			Vpcs: updateResources,
		}
		if err = dataCli.Aws.Vpc.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-vpc batch compare db update failed. accountID: %s, region: %s, err: %v",
				enumor.Aws, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &cloud.VpcBatchCreateReq[cloud.AwsVpcCreateExt]{
			Vpcs: createResources,
		}
		if _, err = dataCli.Aws.Vpc.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("%s-vpc batch compare db create failed. accountID: %s, region: %s, err: %v",
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
		err = BatchDeleteVpcByIDs(kt, deleteIDs, dataCli)
		if err != nil {
			logs.Errorf("%s-vpc batch compare db delete failed. accountID: %s, region: %s, delIDs: %v, err: %v",
				enumor.Aws, req.AccountID, req.Region, deleteIDs, err)
			return err
		}
	}

	return nil
}

// filterAwsVpcList filter aws vpc list
func filterAwsVpcList(req *hcservice.ResourceSyncReq, list *types.AwsVpcListResult,
	resourceDBMap map[string]cloudcore.BaseVpc) (createResources []cloud.VpcCreateReq[cloud.AwsVpcCreateExt],
	updateResources []cloud.VpcUpdateReq[cloud.AwsVpcUpdateExt], existIDMap map[string]bool, err error) {
	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpclist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
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

			tmpRes := cloud.VpcUpdateReq[cloud.AwsVpcUpdateExt]{
				ID: resourceInfo.ID,
				Extension: &cloud.AwsVpcUpdateExt{
					State:              item.Extension.State,
					InstanceTenancy:    converter.ValToPtr(item.Extension.InstanceTenancy),
					IsDefault:          converter.ValToPtr(item.Extension.IsDefault),
					EnableDnsHostnames: converter.ValToPtr(item.Extension.EnableDnsHostnames),
					EnableDnsSupport:   converter.ValToPtr(item.Extension.EnableDnsSupport),
				},
			}
			tmpRes.Name = converter.ValToPtr(item.Name)
			tmpRes.Memo = item.Memo

			if item.Extension.Cidr != nil {
				tmpCidrs := []cloud.AwsCidr{}
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.AwsCidr{
						Type:        cidrItem.Type,
						Cidr:        cidrItem.Cidr,
						AddressPool: cidrItem.AddressPool,
						State:       cidrItem.State,
					})
				}
				tmpRes.Extension.Cidr = tmpCidrs
			}

			updateResources = append(updateResources, tmpRes)
			existIDMap[resourceInfo.ID] = true
		} else {
			// need add vpc data
			tmpRes := cloud.VpcCreateReq[cloud.AwsVpcCreateExt]{
				AccountID: req.AccountID,
				CloudID:   item.CloudID,
				Name:      converter.ValToPtr(item.Name),
				Region:    item.Region,
				Category:  enumor.BizVpcCategory,
				Memo:      item.Memo,
				Extension: &cloud.AwsVpcCreateExt{
					State:              item.Extension.State,
					InstanceTenancy:    item.Extension.InstanceTenancy,
					IsDefault:          item.Extension.IsDefault,
					EnableDnsHostnames: item.Extension.EnableDnsHostnames,
					EnableDnsSupport:   item.Extension.EnableDnsSupport,
				},
			}

			if item.Extension.Cidr != nil {
				tmpCidrs := []cloud.AwsCidr{}
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.AwsCidr{
						Type:        cidrItem.Type,
						Cidr:        cidrItem.Cidr,
						AddressPool: cidrItem.AddressPool,
						State:       cidrItem.State,
					})
				}
				tmpRes.Extension.Cidr = tmpCidrs
			}

			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, existIDMap, nil
}

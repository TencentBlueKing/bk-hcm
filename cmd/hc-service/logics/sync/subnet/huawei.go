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
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// HuaWeiSubnetSync sync huawei cloud subnet.
func HuaWeiSubnetSync(kt *kit.Kit, req *hcservice.ResourceSyncReq, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client) (interface{}, error) {

	if len(req.CloudIDs) > 0 && len(req.VpcID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vpc_id is required")
	}

	if len(req.Region) == 0 {
		return nil, errf.New(errf.InvalidParameter, "region is required")
	}

	// batch get subnet list from cloudapi.
	list, err := BatchGetHuaWeiSubnetList(kt, req, adaptor)
	if err != nil {
		logs.Errorf("%s-subnet request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	if list.Details == nil {
		return nil, nil
	}

	// batch get subnet map from db.
	resourceDBMap, err := BatchGetSubnetMapFromDB(kt, req, enumor.HuaWei, "", dataCli)
	if err != nil {
		logs.Errorf("%s-subnet batch get subnetdblist failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch sync vendor subnet list.
	err = BatchSyncHuaWeiSubnetList(kt, req, list, resourceDBMap, dataCli)
	if err != nil {
		logs.Errorf("%s-subnet compare api and subnetdblist failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// BatchGetHuaWeiSubnetList batch get subnet list from cloudapi.
func BatchGetHuaWeiSubnetList(kt *kit.Kit, req *hcservice.ResourceSyncReq, adaptor *cloudclient.CloudAdaptorClient) (
	*types.HuaWeiSubnetListResult, error) {

	if len(req.CloudIDs) > 0 {
		return BatchGetHuaWeiSubnetListByCloudIDs(kt, req, adaptor)
	}

	return BatchGetHuaWeiSubnetAllList(kt, req, adaptor)
}

// BatchGetHuaWeiSubnetListByCloudIDs batch get subnet list from cloudapi.
func BatchGetHuaWeiSubnetListByCloudIDs(kt *kit.Kit, req *hcservice.ResourceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient) (*types.HuaWeiSubnetListResult, error) {

	cli, err := adaptor.HuaWei(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.HuaWeiSubnetListByIDOption{
		Region:   req.Region,
		VpcID:    req.VpcID,
		CloudIDs: req.CloudIDs,
	}
	list, err := cli.ListSubnetByID(kt, opt)
	if err != nil {
		logs.Errorf("%s-subnet batch get cloud api failed, err: %v, opt: %v, rid: %s", enumor.HuaWei, err, opt, kt.Rid)
		return nil, err
	}

	return list, nil
}

// BatchGetHuaWeiSubnetAllList batch get subnet list from cloudapi.
func BatchGetHuaWeiSubnetAllList(kt *kit.Kit, req *hcservice.ResourceSyncReq, adaptor *cloudclient.CloudAdaptorClient) (
	*types.HuaWeiSubnetListResult, error) {
	cli, err := adaptor.HuaWei(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	var nextMarker *string
	list := new(types.HuaWeiSubnetListResult)
	for {
		opt := &types.HuaWeiSubnetListOption{
			Region: req.Region,
			Page: &adcore.HuaWeiPage{
				Limit: converter.ValToPtr(int32(adcore.HuaWeiQueryLimit)),
			},
			VpcID: req.VpcID,
		}

		// 分页查询的起始资源ID，表示从指定资源的下一条记录开始查询。
		if nextMarker != nil {
			opt.Page.Marker = nextMarker
		}

		tmpList, tmpErr := cli.ListSubnet(kt, opt)
		if tmpErr != nil {
			logs.Errorf("%s-subnet batch get cloud api failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, tmpErr)
			return nil, tmpErr
		}

		list.Details = append(list.Details, tmpList.Details...)
		nextMarker = converter.ValToPtr(tmpList.Details[len(tmpList.Details)-1].CloudID)

		if len(tmpList.Details) < adcore.HuaWeiQueryLimit {
			break
		}
	}

	return list, nil
}

// BatchSyncHuaWeiSubnetList batch sync vendor subnet list.
func BatchSyncHuaWeiSubnetList(kt *kit.Kit, req *hcservice.ResourceSyncReq, list *types.HuaWeiSubnetListResult,
	resourceDBMap map[string]cloudcore.BaseSubnet, dataCli *dataclient.Client) error {

	createResources, updateResources, existIDMap, err := filterHuaWeiSubnetList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.SubnetBatchUpdateReq[cloud.HuaWeiSubnetUpdateExt]{
			Subnets: updateResources,
		}
		if err = dataCli.HuaWei.Subnet.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-subnet batch compare db update failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		err = batchCreateHuaWeiSubnet(kt, createResources, dataCli)
		if err != nil {
			logs.Errorf("%s-subnet batch compare db create failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, err)
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
				"err: %v", enumor.HuaWei, req.AccountID, req.Region, deleteIDs, err)
			return err
		}
	}

	return nil
}

// filterHuaWeiSubnetList filter huawei subnet list
func filterHuaWeiSubnetList(req *hcservice.ResourceSyncReq, list *types.HuaWeiSubnetListResult,
	resourceDBMap map[string]cloudcore.BaseSubnet) (
	createResources []cloud.SubnetCreateReq[cloud.HuaWeiSubnetCreateExt],
	updateResources []cloud.SubnetUpdateReq[cloud.HuaWeiSubnetUpdateExt], existIDMap map[string]bool, err error) {
	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi subnetlist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	existIDMap = make(map[string]bool, 0)
	for _, item := range list.Details {
		// need compare and update subnet data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if resourceInfo.Name == item.Name && resourceInfo.CloudVpcID == item.CloudVpcID &&
				resourceInfo.Memo == item.Memo {
				existIDMap[resourceInfo.ID] = true
				continue
			}

			tmpRes := cloud.SubnetUpdateReq[cloud.HuaWeiSubnetUpdateExt]{
				ID: resourceInfo.ID,
				Extension: &cloud.HuaWeiSubnetUpdateExt{
					Status:       item.Extension.Status,
					DhcpEnable:   converter.ValToPtr(item.Extension.DhcpEnable),
					GatewayIp:    item.Extension.GatewayIp,
					DnsList:      item.Extension.DnsList,
					NtpAddresses: item.Extension.NtpAddresses,
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
			tmpRes := cloud.SubnetCreateReq[cloud.HuaWeiSubnetCreateExt]{
				AccountID:  req.AccountID,
				CloudVpcID: item.CloudVpcID,
				CloudID:    item.CloudID,
				Name:       converter.ValToPtr(item.Name),
				Region:     item.Extension.Region,
				Ipv4Cidr:   item.Ipv4Cidr,
				Memo:       item.Memo,
				Extension: &cloud.HuaWeiSubnetCreateExt{
					Status:       item.Extension.Status,
					DhcpEnable:   item.Extension.DhcpEnable,
					GatewayIp:    item.Extension.GatewayIp,
					DnsList:      item.Extension.DnsList,
					NtpAddresses: item.Extension.NtpAddresses,
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

func batchCreateHuaWeiSubnet(kt *kit.Kit, createResources []cloud.SubnetCreateReq[cloud.HuaWeiSubnetCreateExt],
	dataCli *dataclient.Client) error {

	querySize := int(filter.DefaultMaxInLimit)
	times := len(createResources) / querySize
	if len(createResources)%querySize != 0 {
		times++
	}

	for i := 0; i < times; i++ {
		var newResources []cloud.SubnetCreateReq[cloud.HuaWeiSubnetCreateExt]
		if i == times-1 {
			newResources = append(newResources, createResources[i*querySize:]...)
		} else {
			newResources = append(newResources, createResources[i*querySize:(i+1)*querySize]...)
		}

		createReq := &cloud.SubnetBatchCreateReq[cloud.HuaWeiSubnetCreateExt]{
			Subnets: newResources,
		}
		if _, err := dataCli.HuaWei.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			return err
		}
	}

	return nil
}

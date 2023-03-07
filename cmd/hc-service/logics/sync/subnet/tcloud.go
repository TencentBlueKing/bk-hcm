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

// TCloudSubnetSync sync tencent cloud subnet.
func TCloudSubnetSync(kt *kit.Kit, req *hcservice.TCloudResourceSyncReq, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client) (interface{}, error) {

	// batch get subnet list from cloudapi.
	list, err := BatchGetTCloudSubnetList(kt, req, adaptor)
	if err != nil {
		logs.Errorf("%s-subnet request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.TCloud, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch get subnet map from db.
	resourceDBMap, err := BatchGetSubnetMapFromDB(kt, enumor.TCloud, req.CloudIDs, "", dataCli)
	if err != nil {
		logs.Errorf("%s-subnet batch get subnetdblist failed. accountID: %s, region: %s, err: %v",
			enumor.TCloud, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch sync vendor subnet list.
	err = BatchSyncTcloudSubnetList(kt, req, list, resourceDBMap, dataCli)
	if err != nil {
		logs.Errorf("%s-subnet compare api and subnetdblist failed. accountID: %s, region: %s, err: %v",
			enumor.TCloud, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// BatchGetTCloudSubnetList batch get subnet list from cloudapi.
func BatchGetTCloudSubnetList(kt *kit.Kit, req *hcservice.TCloudResourceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient) (*types.TCloudSubnetListResult, error) {

	cli, err := adaptor.TCloud(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	page := uint64(0)
	list := new(types.TCloudSubnetListResult)
	for {
		count := uint64(adcore.TCloudQueryLimit)
		offset := page * count
		opt := &adcore.TCloudListOption{
			Region: req.Region,
		}

		// 查询指定CloudIDs
		if len(req.CloudIDs) > 0 {
			opt.CloudIDs = req.CloudIDs
		} else {
			opt.Page = &adcore.TCloudPage{
				Offset: offset,
				Limit:  count,
			}
		}

		tmpList, tmpErr := cli.ListSubnet(kt, opt)
		if tmpErr != nil {
			logs.Errorf("%s-subnet batch get cloudapi failed. accountID: %s, region: %s, offset: %d, "+
				"count: %d, err: %v", enumor.TCloud, req.AccountID, req.Region, offset, count, tmpErr)
			return nil, tmpErr
		}

		list.Details = append(list.Details, tmpList.Details...)

		if len(req.CloudIDs) > 0 || len(tmpList.Details) < int(count) {
			break
		}

		page++
	}

	return list, nil
}

// BatchGetSubnetMapFromDB batch get subnet map from db.
func BatchGetSubnetMapFromDB(kt *kit.Kit, vendor enumor.Vendor, cloudIDs []string, cloudVpcID string,
	dataCli *dataclient.Client) (map[string]cloudcore.BaseSubnet, error) {

	rulesCommon := []filter.RuleFactory{
		&filter.AtomRule{
			Field: "vendor",
			Op:    filter.Equal.Factory(),
			Value: vendor,
		},
	}

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
	resourceMap := make(map[string]cloudcore.BaseSubnet, 0)
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
		dbList, err := dataCli.Global.Subnet.List(kt.Ctx, kt.Header(), dbQueryReq)
		if err != nil {
			logs.Errorf("%s-subnet batch list db error. offset: %d, limit: %d, err: %v",
				vendor, offset, count, err)
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

// BatchSyncTcloudSubnetList batch sync vendor subnet list.
func BatchSyncTcloudSubnetList(kt *kit.Kit, req *hcservice.TCloudResourceSyncReq, list *types.TCloudSubnetListResult,
	resourceDBMap map[string]cloudcore.BaseSubnet, dataCli *dataclient.Client) error {

	createResources, updateResources, existIDMap, err := filterTcloudSubnetList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.SubnetBatchUpdateReq[cloud.TCloudSubnetUpdateExt]{
			Subnets: updateResources,
		}
		if err = dataCli.TCloud.Subnet.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-subnet batch compare db update failed. accountID: %s, region: %s, err: %v",
				enumor.TCloud, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		err = batchCreateTcloudSubnet(kt, createResources, dataCli)
		if err != nil {
			logs.Errorf("%s-subnet batch compare db create failed. accountID: %s, region: %s, err: %v",
				enumor.TCloud, req.AccountID, req.Region, err)
			return err
		}
	}

	// delete resource data
	deleteIDs := make([]string, 0)
	for _, resourceItem := range resourceDBMap {
		if _, ok := existIDMap[resourceItem.ID]; !ok {
			deleteIDs = append(deleteIDs, resourceItem.ID)
		}
	}

	if len(deleteIDs) > 0 {
		err = BatchDeleteSubnetByIDs(kt, deleteIDs, dataCli)
		if err != nil {
			logs.Errorf("%s-subnet batch compare db delete failed. accountID: %s, region: %s, delIDs: %v, "+
				"err: %v", enumor.TCloud, req.AccountID, req.Region, deleteIDs, err)
			return err
		}
	}

	return nil
}

// filterTcloudSubnetList filter tcloud subnet list
func filterTcloudSubnetList(req *hcservice.TCloudResourceSyncReq, list *types.TCloudSubnetListResult,
	resourceDBMap map[string]cloudcore.BaseSubnet) (
	createResources []cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt],
	updateResources []cloud.SubnetUpdateReq[cloud.TCloudSubnetUpdateExt], existIDMap map[string]bool, err error) {

	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi subnetlist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	existIDMap = make(map[string]bool, 0)
	for _, item := range list.Details {
		// need compare and update resource data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if resourceInfo.Name == item.Name && resourceInfo.CloudVpcID == item.CloudVpcID &&
				converter.PtrToVal(resourceInfo.Memo) == converter.PtrToVal(item.Memo) {
				existIDMap[resourceInfo.ID] = true
				continue
			}

			tmpRes := cloud.SubnetUpdateReq[cloud.TCloudSubnetUpdateExt]{
				ID: resourceInfo.ID,
				Extension: &cloud.TCloudSubnetUpdateExt{
					IsDefault:         item.Extension.IsDefault,
					Region:            item.Extension.Region,
					Zone:              item.Extension.Zone,
					CloudNetworkAclID: item.Extension.CloudNetworkAclID,
				},
			}
			tmpRes.Name = converter.ValToPtr(item.Name)
			tmpRes.Ipv4Cidr = item.Ipv4Cidr
			tmpRes.Ipv6Cidr = item.Ipv6Cidr
			tmpRes.Memo = item.Memo

			updateResources = append(updateResources, tmpRes)
			existIDMap[resourceInfo.ID] = true
		} else {
			// need add resource data
			tmpRes := cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt]{
				AccountID:  req.AccountID,
				CloudVpcID: item.CloudVpcID,
				CloudID:    item.CloudID,
				Name:       converter.ValToPtr(item.Name),
				Region:     item.Extension.Region,
				Zone:       item.Extension.Zone,
				Ipv4Cidr:   item.Ipv4Cidr,
				Ipv6Cidr:   item.Ipv6Cidr,
				Memo:       item.Memo,
				Extension: &cloud.TCloudSubnetCreateExt{
					IsDefault:         item.Extension.IsDefault,
					CloudNetworkAclID: item.Extension.CloudNetworkAclID,
				},
			}
			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, existIDMap, nil
}

func batchCreateTcloudSubnet(kt *kit.Kit, createResources []cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt],
	dataCli *dataclient.Client) error {

	querySize := int(filter.DefaultMaxInLimit)
	times := len(createResources) / querySize
	if len(createResources)%querySize != 0 {
		times++
	}

	for i := 0; i < times; i++ {
		var newResources []cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt]
		if i == times-1 {
			newResources = append(newResources, createResources[i*querySize:]...)
		} else {
			newResources = append(newResources, createResources[i*querySize:(i+1)*querySize]...)
		}

		createReq := &cloud.SubnetBatchCreateReq[cloud.TCloudSubnetCreateExt]{
			Subnets: newResources,
		}

		if _, err := dataCli.TCloud.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			return err
		}
	}

	return nil
}

// BatchDeleteSubnetByIDs batch delete subnet ids
func BatchDeleteSubnetByIDs(kt *kit.Kit, deleteIDs []string, dataCli *dataclient.Client) error {
	querySize := int(filter.DefaultMaxInLimit)
	times := len(deleteIDs) / querySize
	if len(deleteIDs)%querySize != 0 {
		times++
	}

	for i := 0; i < times; i++ {
		var newDeleteIDs []string
		if i == times-1 {
			newDeleteIDs = append(newDeleteIDs, deleteIDs[i*querySize:]...)
		} else {
			newDeleteIDs = append(newDeleteIDs, deleteIDs[i*querySize:(i+1)*querySize]...)
		}

		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("id", newDeleteIDs),
		}
		if err := dataCli.Global.Subnet.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
			return err
		}
	}

	return nil
}

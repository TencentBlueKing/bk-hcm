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

// TCloudVpcSync sync tencent cloud vpc.
func TCloudVpcSync(kt *kit.Kit, req *hcservice.TCloudResourceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	// batch get vpc list from cloudapi.
	list, err := BatchGetTCloudVpcList(kt, req, adaptor)
	if err != nil {
		logs.Errorf("%s-vpc request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.TCloud, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch get vpc map from db.
	resourceDBMap, err := BatchGetVpcMapFromDB(kt, enumor.TCloud, dataCli)
	if err != nil {
		logs.Errorf("%s-vpc batch get vpcdblist failed. accountID: %s, region: %s, err: %v",
			enumor.TCloud, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch sync vendor vpc list.
	err = BatchSyncTcloudVpcList(kt, req, list, resourceDBMap, dataCli)
	if err != nil {
		logs.Errorf("%s-vpc compare api and dblist failed. accountID: %s, region: %s, err: %v",
			enumor.TCloud, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// BatchGetTCloudVpcList batch get vpc list from cloudapi.
func BatchGetTCloudVpcList(kt *kit.Kit, req *hcservice.TCloudResourceSyncReq, adaptor *cloudclient.CloudAdaptorClient) (
	*types.TCloudVpcListResult, error) {
	cli, err := adaptor.TCloud(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	page := uint64(0)
	list := new(types.TCloudVpcListResult)
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

		tmpList, tmpErr := cli.ListVpc(kt, opt)
		if tmpErr != nil {
			logs.Errorf("%s-vpc batch get cloudapi failed. accountID: %s, region: %s, offset: %d, count: %d, "+
				"err: %v", enumor.TCloud, req.AccountID, req.Region, offset, count, tmpErr)
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

// BatchGetVpcMapFromDB batch get vpc map from db.
func BatchGetVpcMapFromDB(kt *kit.Kit, vendor enumor.Vendor, dataCli *dataclient.Client) (
	map[string]cloudcore.BaseVpc, error) {

	page := uint32(0)
	resourceMap := make(map[string]cloudcore.BaseVpc, 0)
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
			},
		}
		dbQueryReq := &core.ListReq{
			Filter: expr,
			Page:   &core.BasePage{Count: false, Start: offset, Limit: count},
		}
		dbList, err := dataCli.Global.Vpc.List(kt.Ctx, kt.Header(), dbQueryReq)
		if err != nil {
			logs.Errorf("%s-vpc batch get vpclist db error. offset: %d, limit: %d, err: %v",
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

// BatchSyncTcloudVpcList batch sync vendor vpc list.
func BatchSyncTcloudVpcList(kt *kit.Kit, req *hcservice.TCloudResourceSyncReq, list *types.TCloudVpcListResult,
	resourceDBMap map[string]cloudcore.BaseVpc, dataCli *dataclient.Client) error {

	createResources, updateResources, existIDMap, err := filterTcloudVpcList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.VpcBatchUpdateReq[cloud.TCloudVpcUpdateExt]{
			Vpcs: updateResources,
		}
		if err = dataCli.TCloud.Vpc.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-vpc batch compare db update failed. accountID: %s, region: %s, err: %v",
				enumor.TCloud, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &cloud.VpcBatchCreateReq[cloud.TCloudVpcCreateExt]{
			Vpcs: createResources,
		}
		if _, err = dataCli.TCloud.Vpc.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("%s-vpc batch compare db create failed. accountID: %s, region: %s, err: %v",
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
		err = BatchDeleteVpcByIDs(kt, deleteIDs, dataCli)
		if err != nil {
			logs.Errorf("%s-vpc batch compare db delete failed. accountID: %s, region: %s, deleteIDs: %v, "+
				"err: %v", enumor.TCloud, req.AccountID, req.Region, deleteIDs, err)
			return err
		}
	}

	return nil
}

// filterTcloudVpcList filter tcloud vpc list
func filterTcloudVpcList(req *hcservice.TCloudResourceSyncReq, list *types.TCloudVpcListResult,
	resourceDBMap map[string]cloudcore.BaseVpc) (createResources []cloud.VpcCreateReq[cloud.TCloudVpcCreateExt],
	updateResources []cloud.VpcUpdateReq[cloud.TCloudVpcUpdateExt], existIDMap map[string]bool, err error) {
	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpclist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	existIDMap = make(map[string]bool, 0)
	for _, item := range list.Details {
		// need compare and update resource data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if resourceInfo.Name == item.Name && resourceInfo.Region == item.Region &&
				converter.PtrToVal(resourceInfo.Memo) == converter.PtrToVal(item.Memo) {
				existIDMap[resourceInfo.ID] = true
				continue
			}

			tmpRes := cloud.VpcUpdateReq[cloud.TCloudVpcUpdateExt]{
				ID: resourceInfo.ID,
				Extension: &cloud.TCloudVpcUpdateExt{
					IsDefault:       converter.ValToPtr(item.Extension.IsDefault),
					EnableMulticast: converter.ValToPtr(item.Extension.EnableMulticast),
					DnsServerSet:    item.Extension.DnsServerSet,
					DomainName:      converter.ValToPtr(item.Extension.DomainName),
				},
			}
			tmpRes.Name = converter.ValToPtr(item.Name)
			tmpRes.Memo = item.Memo

			if item.Extension.Cidr != nil {
				tmpCidrs := []cloud.TCloudCidr{}
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.TCloudCidr{
						Type:     cidrItem.Type,
						Cidr:     cidrItem.Cidr,
						Category: cidrItem.Category,
					})
				}
				tmpRes.Extension.Cidr = tmpCidrs
			}

			updateResources = append(updateResources, tmpRes)
			existIDMap[resourceInfo.ID] = true
		} else {
			// need add resource data
			tmpRes := cloud.VpcCreateReq[cloud.TCloudVpcCreateExt]{
				AccountID: req.AccountID,
				CloudID:   item.CloudID,
				Name:      converter.ValToPtr(item.Name),
				Region:    item.Region,
				Category:  enumor.BizVpcCategory,
				Memo:      item.Memo,
				Extension: &cloud.TCloudVpcCreateExt{
					IsDefault:       item.Extension.IsDefault,
					EnableMulticast: item.Extension.EnableMulticast,
					DnsServerSet:    item.Extension.DnsServerSet,
					DomainName:      item.Extension.DomainName,
				},
			}

			if item.Extension.Cidr != nil {
				tmpCidrs := []cloud.TCloudCidr{}
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.TCloudCidr{
						Type:     cidrItem.Type,
						Cidr:     cidrItem.Cidr,
						Category: cidrItem.Category,
					})
				}
				tmpRes.Extension.Cidr = tmpCidrs
			}

			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, existIDMap, nil
}

// BatchDeleteVpcByIDs batch delete vpc ids
func BatchDeleteVpcByIDs(kt *kit.Kit, deleteIDs []string, dataCli *dataclient.Client) error {
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
		if err := dataCli.Global.Vpc.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
			return err
		}
	}

	return nil
}

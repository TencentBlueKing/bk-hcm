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

	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// TCloudSubnetUpdate update tencent cloud subnet.
func (s subnet) TCloudSubnetUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.SubnetUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := s.cs.DataService().TCloud.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.TCloud(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := new(types.TCloudSubnetUpdateOption)
	err = cli.UpdateSubnet(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.SubnetBatchUpdateReq[cloud.TCloudSubnetUpdateExt]{
		Subnets: []cloud.SubnetUpdateReq[cloud.TCloudSubnetUpdateExt]{{
			ID: id,
			SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = s.cs.DataService().TCloud.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// TCloudSubnetDelete delete tencent cloud subnet.
func (s subnet) TCloudSubnetDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := s.cs.DataService().TCloud.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.TCloud(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	delOpt := &adcore.BaseRegionalDeleteOption{
		BaseDeleteOption: adcore.BaseDeleteOption{ResourceID: getRes.CloudID},
		Region:           getRes.Extension.Region,
	}
	err = cli.DeleteSubnet(cts.Kit, delOpt)
	if err != nil {
		return nil, err
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	err = s.cs.DataService().Global.Subnet.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// TCloudSubnetSync sync tencent cloud subnet.
func (s subnet) TCloudSubnetSync(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.ResourceSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if len(req.Region) == 0 {
		return nil, errf.New(errf.InvalidParameter, "region is required")
	}

	// batch get subnet list from cloudapi.
	list, err := s.BatchGetTCloudSubnetList(cts, req)
	if err != nil {
		logs.Errorf("[%s-subnet] request cloudapi response failed. accountID:%s, region:%s, err: %v",
			enumor.TCloud, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch get subnet map from db.
	resourceDBMap, err := s.BatchGetSubnetMapFromDB(cts, req, enumor.TCloud, "")
	if err != nil {
		logs.Errorf("[%s-subnet] batch get subnetdblist failed. accountID:%s, region:%s, err: %v",
			enumor.TCloud, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch sync vendor subnet list.
	err = s.BatchSyncTcloudSubnetList(cts, req, list, resourceDBMap)
	if err != nil {
		logs.Errorf("[%s-subnet] compare api and subnetdblist failed. accountID:%s, region:%s, err: %v",
			enumor.TCloud, req.AccountID, req.Region, err)
		return nil, err
	}

	return hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// BatchGetTCloudSubnetList batch get subnet list from cloudapi.
func (s subnet) BatchGetTCloudSubnetList(cts *rest.Contexts, req *hcservice.ResourceSyncReq) (
	*types.TCloudSubnetListResult, error) {
	var (
		page  uint64
		count uint64 = adcore.TCloudQueryLimit
		list         = new(types.TCloudSubnetListResult)
	)

	cli, err := s.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	for {
		offset := page * count
		opt := &adcore.TCloudListOption{
			Region: req.Region,
			Page: &adcore.TCloudPage{
				Offset: offset,
				Limit:  count,
			},
		}
		tmpList, tmpErr := cli.ListSubnet(cts.Kit, opt)
		if tmpErr != nil {
			logs.Errorf("[%s-subnet]batch get cloudapi failed. accountID:%s, region:%s, offset:%d, count:%d, "+
				"err: %v", enumor.TCloud, req.AccountID, req.Region, offset, count, tmpErr)
			return nil, tmpErr
		}

		list.Details = append(list.Details, tmpList.Details...)
		if len(tmpList.Details) < int(count) {
			break
		}
		page++
	}
	return list, nil
}

// BatchGetSubnetMapFromDB batch get subnet map from db.
func (s subnet) BatchGetSubnetMapFromDB(cts *rest.Contexts, req *hcservice.ResourceSyncReq, vendor enumor.Vendor,
	cloudVpcID string) (map[string]cloudcore.BaseSubnet, error) {
	var (
		page        uint32
		count       = core.DefaultMaxPageLimit
		resourceMap = make(map[string]cloudcore.BaseSubnet, 0)
		rulesCommon = []filter.RuleFactory{
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
		}
	)
	if cloudVpcID != "" {
		rulesCommon = append(rulesCommon, &filter.AtomRule{
			Field: "cloud_vpc_id",
			Op:    filter.Equal.Factory(),
			Value: cloudVpcID,
		})
	}

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
		dbList, err := s.cs.DataService().Global.Subnet.List(cts.Kit.Ctx, cts.Kit.Header(), dbQueryReq)
		if err != nil {
			logs.Errorf("[%s-subnet]batch list db error. accountID:%s, region:%s, offset:%d, limit:%d, err: %v",
				vendor, req.AccountID, req.Region, offset, count, err)
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
func (s subnet) BatchSyncTcloudSubnetList(cts *rest.Contexts, req *hcservice.ResourceSyncReq,
	list *types.TCloudSubnetListResult, resourceDBMap map[string]cloudcore.BaseSubnet) error {
	createResources, updateResources, existIDMap, err := s.filterTcloudSubnetList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.SubnetBatchUpdateReq[cloud.TCloudSubnetUpdateExt]{
			Subnets: updateResources,
		}
		if err = s.cs.DataService().TCloud.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
			logs.Errorf("[%s-subnet]batch compare db update failed. accountID:%s, region:%s, err: %v",
				enumor.TCloud, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		err = s.batchCreateTcloudSubnet(cts, createResources)
		if err != nil {
			logs.Errorf("[%s-subnet]batch compare db create failed. accountID:%s, region:%s, err: %v",
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
		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("id", deleteIDs),
		}
		if err = s.cs.DataService().Global.Subnet.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq); err != nil {
			logs.Errorf("[%s-subnet]batch compare db delete failed. accountID:%s, region:%s, delIDs:%v, err: %v",
				enumor.TCloud, req.AccountID, req.Region, deleteIDs, err)
			return err
		}
	}
	return nil
}

// filterTcloudSubnetList filter tcloud subnet list
func (s subnet) filterTcloudSubnetList(req *hcservice.ResourceSyncReq, list *types.TCloudSubnetListResult,
	resourceDBMap map[string]cloudcore.BaseSubnet) (
	createResources []cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt],
	updateResources []cloud.SubnetUpdateReq[cloud.TCloudSubnetUpdateExt], existIDMap map[string]bool, err error) {
	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi subnetlist is empty, accountID:%s, region:%s", req.AccountID, req.Region)
	}

	existIDMap = make(map[string]bool, 0)
	for _, item := range list.Details {
		// need compare and update resource data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			tmpRes := cloud.SubnetUpdateReq[cloud.TCloudSubnetUpdateExt]{
				ID: resourceInfo.ID,
				Extension: &cloud.TCloudSubnetUpdateExt{
					IsDefault:    item.Extension.IsDefault,
					Region:       item.Extension.Region,
					Zone:         item.Extension.Zone,
					NetworkAclId: item.Extension.NetworkAclId,
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
				Ipv4Cidr:   item.Ipv4Cidr,
				Ipv6Cidr:   item.Ipv6Cidr,
				Memo:       item.Memo,
				Extension: &cloud.TCloudSubnetCreateExt{
					IsDefault:    item.Extension.IsDefault,
					Region:       item.Extension.Region,
					Zone:         item.Extension.Zone,
					NetworkAclId: item.Extension.NetworkAclId,
				},
			}
			createResources = append(createResources, tmpRes)
		}
	}
	return createResources, updateResources, existIDMap, nil
}

func (s subnet) batchCreateTcloudSubnet(cts *rest.Contexts,
	createResources []cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt]) error {
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
		if _, err := s.cs.DataService().TCloud.Subnet.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq); err != nil {
			return err
		}
	}
	return nil
}

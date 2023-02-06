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

// HuaWeiSubnetUpdate update huawei subnet.
func (s subnet) HuaWeiSubnetUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.SubnetUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := s.cs.DataService().HuaWei.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.HuaWei(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := &types.HuaweiSubnetUpdateOption{
		SubnetUpdateOption: types.SubnetUpdateOption{
			ResourceID: getRes.CloudID,
			Data:       &types.BaseSubnetUpdateData{Memo: req.Memo},
		},
		Name:   getRes.Name,
		Region: getRes.Extension.Region,
		VpcID:  getRes.CloudVpcID,
	}
	err = cli.UpdateSubnet(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.SubnetBatchUpdateReq[cloud.HuaWeiSubnetUpdateExt]{
		Subnets: []cloud.SubnetUpdateReq[cloud.HuaWeiSubnetUpdateExt]{{
			ID: id,
			SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = s.cs.DataService().HuaWei.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// HuaWeiSubnetDelete delete huawei subnet.
func (s subnet) HuaWeiSubnetDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := s.cs.DataService().HuaWei.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.HuaWei(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	delOpt := &types.HuaweiSubnetDeleteOption{
		BaseRegionalDeleteOption: adcore.BaseRegionalDeleteOption{
			BaseDeleteOption: adcore.BaseDeleteOption{ResourceID: getRes.CloudID},
			Region:           getRes.Extension.Region,
		},
		VpcID: getRes.CloudVpcID,
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

// HuaweiSubnetSync sync huawei cloud subnet.
func (s subnet) HuaweiSubnetSync(cts *rest.Contexts) (interface{}, error) {
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
	list, err := s.BatchGetHuaWeiSubnetList(cts, req)
	if err != nil {
		logs.Errorf("[%s-subnet] request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch get subnet map from db.
	resourceDBMap, err := s.BatchGetSubnetMapFromDB(cts, req, enumor.HuaWei, "")
	if err != nil {
		logs.Errorf("[%s-subnet] batch get subnetdblist failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch sync vendor subnet list.
	err = s.BatchSyncHuaWeiSubnetList(cts, req, list, resourceDBMap)
	if err != nil {
		logs.Errorf("[%s-subnet] compare api and subnetdblist failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// BatchGetHuaWeiSubnetList batch get subnet list from cloudapi.
func (s subnet) BatchGetHuaWeiSubnetList(cts *rest.Contexts, req *hcservice.ResourceSyncReq) (
	*types.HuaweiSubnetListResult, error) {
	cli, err := s.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	list := new(types.HuaweiSubnetListResult)
	for {
		opt := new(types.HuaweiSubnetListOption)
		opt.Region = req.Region
		count := int32(adcore.HuaweiQueryLimit)
		opt.Page = &adcore.HuaweiPage{
			Limit: converter.ValToPtr(count),
		}

		tmpList, tmpErr := cli.ListSubnet(cts.Kit, opt)
		if tmpErr != nil {
			logs.Errorf("[%s-subnet]batch get cloud api failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, tmpErr)
			return nil, tmpErr
		}

		list.Details = append(list.Details, tmpList.Details...)
		if len(tmpList.Details) < int(count) {
			break
		}
	}

	return list, nil
}

// BatchSyncHuaWeiSubnetList batch sync vendor subnet list.
func (s subnet) BatchSyncHuaWeiSubnetList(cts *rest.Contexts, req *hcservice.ResourceSyncReq,
	list *types.HuaweiSubnetListResult, resourceDBMap map[string]cloudcore.BaseSubnet) error {
	createResources, updateResources, existIDMap, err := s.filterHuaWeiSubnetList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.SubnetBatchUpdateReq[cloud.HuaWeiSubnetUpdateExt]{
			Subnets: updateResources,
		}
		if err = s.cs.DataService().HuaWei.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
			logs.Errorf("[%s-subnet]batch compare db update failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		err = s.batchCreateHuaWeiSubnet(cts, createResources)
		if err != nil {
			logs.Errorf("[%s-subnet]batch compare db create failed. accountID: %s, region: %s, err: %v",
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
		err = s.BatchDeleteSubnetByIDs(cts, deleteIDs)
		if err != nil {
			logs.Errorf("[%s-subnet]batch compare db delete failed. accountID: %s, region: %s, delIDs: %v, "+
				"err: %v", enumor.HuaWei, req.AccountID, req.Region, deleteIDs, err)
			return err
		}
	}

	return nil
}

// filterHuaWeiSubnetList filter huawei subnet list
func (s subnet) filterHuaWeiSubnetList(req *hcservice.ResourceSyncReq, list *types.HuaweiSubnetListResult,
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
				Ipv4Cidr:   item.Ipv4Cidr,
				Memo:       item.Memo,
				Extension: &cloud.HuaWeiSubnetCreateExt{
					Region:       item.Extension.Region,
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

func (s subnet) batchCreateHuaWeiSubnet(cts *rest.Contexts,
	createResources []cloud.SubnetCreateReq[cloud.HuaWeiSubnetCreateExt]) error {
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
		if _, err := s.cs.DataService().HuaWei.Subnet.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq); err != nil {
			return err
		}
	}

	return nil
}

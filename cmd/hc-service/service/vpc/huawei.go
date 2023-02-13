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
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// HuaWeiVpcUpdate update huawei vpc.
func (v vpc) HuaWeiVpcUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.VpcUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := v.cs.DataService().HuaWei.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.HuaWei(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := &types.HuaWeiVpcUpdateOption{
		VpcUpdateOption: types.VpcUpdateOption{
			ResourceID: getRes.CloudID,
			Data:       &types.BaseVpcUpdateData{Memo: req.Memo},
		},
		Region: getRes.Region,
	}
	err = cli.UpdateVpc(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.VpcBatchUpdateReq[cloud.HuaWeiVpcUpdateExt]{
		Vpcs: []cloud.VpcUpdateReq[cloud.HuaWeiVpcUpdateExt]{{
			ID: id,
			VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = v.cs.DataService().HuaWei.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// HuaWeiVpcDelete delete huawei vpc.
func (v vpc) HuaWeiVpcDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := v.cs.DataService().HuaWei.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.HuaWei(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	delOpt := &adcore.BaseRegionalDeleteOption{
		BaseDeleteOption: adcore.BaseDeleteOption{ResourceID: getRes.CloudID},
		Region:           getRes.Region,
	}
	err = cli.DeleteVpc(cts.Kit, delOpt)
	if err != nil {
		return nil, err
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	err = v.cs.DataService().Global.Vpc.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// HuaWeiVpcSync sync huawei vpc
func (v vpc) HuaWeiVpcSync(cts *rest.Contexts) (interface{}, error) {
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

	// batch get vpc list from cloudapi
	list, err := v.BatchGetHuaWeiVpcList(cts, req)
	if err != nil {
		logs.Errorf("%s-vpc request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch get vpc map from db.
	resourceDBMap, err := v.BatchGetVpcMapFromDB(cts, req, enumor.HuaWei)
	if err != nil {
		logs.Errorf("%s-vpc batch get vpcdblist failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch sync vendor vpc list.
	err = v.BatchSyncHuaWeiVpcList(cts, req, list, resourceDBMap)
	if err != nil {
		logs.Errorf("%s-vpc compare api and dblist failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// BatchGetHuaWeiVpcList batch get vpc list from cloudapi.
func (v vpc) BatchGetHuaWeiVpcList(cts *rest.Contexts, req *hcservice.ResourceSyncReq) (
	*types.HuaWeiVpcListResult, error) {
	cli, err := v.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	nextMarker := ""
	list := new(types.HuaWeiVpcListResult)
	for {
		opt := new(types.HuaWeiVpcListOption)
		opt.Region = req.Region
		opt.Page = &adcore.HuaWeiPage{
			Limit: converter.ValToPtr(int32(adcore.HuaWeiQueryLimit)),
		}
		if nextMarker != "" {
			opt.Page.Marker = converter.ValToPtr(nextMarker)
		}
		tmpList, tmpErr := cli.ListVpc(cts.Kit, opt)
		if tmpErr != nil {
			logs.Errorf("%s-vpc batch get cloud api failed. accountID: %s, region: %s, marker: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, nextMarker, tmpErr)
			return nil, tmpErr
		}

		if len(tmpList.Details) == 0 {
			break
		}

		list.Details = append(list.Details, tmpList.Details...)

		if tmpList.NextMarker == nil {
			break
		}

		nextMarker = *tmpList.NextMarker
	}
	return list, nil
}

// BatchSyncHuaWeiVpcList batch sync vendor vpc list.
func (v vpc) BatchSyncHuaWeiVpcList(cts *rest.Contexts, req *hcservice.ResourceSyncReq,
	list *types.HuaWeiVpcListResult, resourceDBMap map[string]cloudcore.BaseVpc) error {
	createResources, updateResources, existIDMap, err := v.filterHuaWeiVpcList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.VpcBatchUpdateReq[cloud.HuaWeiVpcUpdateExt]{
			Vpcs: updateResources,
		}
		if err = v.cs.DataService().HuaWei.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
			logs.Errorf("%s-vpc batch compare db update failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &cloud.VpcBatchCreateReq[cloud.HuaWeiVpcCreateExt]{
			Vpcs: createResources,
		}
		if _, err = v.cs.DataService().HuaWei.Vpc.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq); err != nil {
			logs.Errorf("%s-vpc batch compare db create failed. accountID: %s, region: %s, err: %v",
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
		err = v.BatchDeleteVpcByIDs(cts, deleteIDs)
		if err != nil {
			logs.Errorf("%s-vpc batch compare db delete failed. accountID: %s, region: %s, delIDs: %v, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, deleteIDs, err)
			return err
		}
	}

	return nil
}

// filterHuaWeiVpcList filter huawei vpc list
func (v vpc) filterHuaWeiVpcList(req *hcservice.ResourceSyncReq, list *types.HuaWeiVpcListResult,
	resourceDBMap map[string]cloudcore.BaseVpc) (createResources []cloud.VpcCreateReq[cloud.HuaWeiVpcCreateExt],
	updateResources []cloud.VpcUpdateReq[cloud.HuaWeiVpcUpdateExt], existIDMap map[string]bool, err error) {
	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpclist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	existIDMap = make(map[string]bool, 0)
	for _, item := range list.Details {
		// need compare and update vpc data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			tmpRes := cloud.VpcUpdateReq[cloud.HuaWeiVpcUpdateExt]{
				ID: resourceInfo.ID,
				Extension: &cloud.HuaWeiVpcUpdateExt{
					Status:              item.Extension.Status,
					EnterpriseProjectId: converter.ValToPtr(item.Extension.EnterpriseProjectId),
				},
			}
			tmpRes.Name = converter.ValToPtr(item.Name)
			tmpRes.Memo = item.Memo

			if item.Extension.Cidr != nil {
				tmpCidrs := []cloud.HuaWeiCidr{}
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.HuaWeiCidr{
						Type: cidrItem.Type,
						Cidr: cidrItem.Cidr,
					})
				}
				tmpRes.Extension.Cidr = tmpCidrs
			}

			updateResources = append(updateResources, tmpRes)
			existIDMap[resourceInfo.ID] = true
		} else {
			// need add vpc data
			tmpRes := cloud.VpcCreateReq[cloud.HuaWeiVpcCreateExt]{
				AccountID: req.AccountID,
				CloudID:   item.CloudID,
				Name:      converter.ValToPtr(item.Name),
				Region:    item.Region,
				Category:  enumor.BizVpcCategory,
				Memo:      item.Memo,
				Extension: &cloud.HuaWeiVpcCreateExt{
					Status:              item.Extension.Status,
					EnterpriseProjectId: item.Extension.EnterpriseProjectId,
				},
			}

			if item.Extension.Cidr != nil {
				tmpCidrs := []cloud.HuaWeiCidr{}
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.HuaWeiCidr{
						Type: cidrItem.Type,
						Cidr: cidrItem.Cidr,
					})
				}
				tmpRes.Extension.Cidr = tmpCidrs
			}

			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, existIDMap, nil
}

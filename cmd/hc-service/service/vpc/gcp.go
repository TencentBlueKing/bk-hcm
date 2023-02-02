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

// GcpVpcUpdate update gcp vpc.
func (v vpc) GcpVpcUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.VpcUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := v.cs.DataService().Gcp.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.Gcp(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := &types.GcpVpcUpdateOption{
		ResourceID: getRes.CloudID,
		Data:       &types.BaseVpcUpdateData{Memo: req.Memo},
	}
	err = cli.UpdateVpc(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.VpcBatchUpdateReq[cloud.GcpVpcUpdateExt]{
		Vpcs: []cloud.VpcUpdateReq[cloud.GcpVpcUpdateExt]{{
			ID: id,
			VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = v.cs.DataService().Gcp.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// GcpVpcDelete delete gcp vpc.
func (v vpc) GcpVpcDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := v.cs.DataService().Gcp.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.Gcp(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	delOpt := &adcore.BaseDeleteOption{
		ResourceID: getRes.CloudID,
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

// GcpVpcSync sync gcp cloud vpc.
func (v vpc) GcpVpcSync(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.ResourceSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// batch get vpc list from cloudapi.
	list, err := v.BatchGetGcpVpcList(cts, req)
	if err != nil {
		logs.Errorf("[%s-vpc] request cloudapi response failed. accountID:%s, region:%s, err: %v",
			enumor.Gcp, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch get vpc map from db.
	resourceDBMap, err := v.BatchGetVpcMapFromDB(cts, req, enumor.Gcp)
	if err != nil {
		logs.Errorf("[%s-vpc] batch get vpcdblist failed. accountID:%s, region:%s, err: %v",
			enumor.Gcp, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch sync vendor vpc list.
	err = v.BatchSyncGcpVpcList(cts, req, list, resourceDBMap)
	if err != nil {
		logs.Errorf("[%s-vpc] compare api and dblist failed. accountID:%s, region:%s, err: %v",
			enumor.Gcp, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// BatchGetGcpVpcList batch get vpc list from cloudapi.
func (v vpc) BatchGetGcpVpcList(cts *rest.Contexts, req *hcservice.ResourceSyncReq) (*types.GcpVpcListResult, error) {
	cli, err := v.ad.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	list := new(types.GcpVpcListResult)
	for {
		opt := new(adcore.GcpListOption)
		opt.Page = &adcore.GcpPage{
			PageSize: int64(adcore.GcpQueryLimit),
		}

		if nextToken != "" {
			opt.Page.PageToken = nextToken
		}

		tmpList, tmpErr := cli.ListVpc(cts.Kit, opt)
		if tmpErr != nil {
			logs.Errorf("[%s-vpc]batch get cloud api failed. accountID:%s, region:%s, nextToken:%s, err: %v",
				enumor.Gcp, req.AccountID, req.Region, nextToken, tmpErr)
			return nil, tmpErr
		}

		if len(tmpList.Details) == 0 {
			break
		}

		list.Details = append(list.Details, tmpList.Details...)
		if len(tmpList.NextPageToken) == 0 {
			break
		}

		nextToken = tmpList.NextPageToken
	}

	return list, nil
}

// BatchSyncGcpVpcList batch sync vendor vpc list.
func (v vpc) BatchSyncGcpVpcList(cts *rest.Contexts, req *hcservice.ResourceSyncReq, list *types.GcpVpcListResult,
	resourceDBMap map[string]cloudcore.BaseVpc) error {
	createResources, updateResources, existIDMap, err := v.filterGcpVpcList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.VpcBatchUpdateReq[cloud.GcpVpcUpdateExt]{
			Vpcs: updateResources,
		}
		if err = v.cs.DataService().Gcp.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
			logs.Errorf("[%s-vpc]batch compare db update failed. accountID:%s, region:%s, err: %v",
				enumor.Gcp, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &cloud.VpcBatchCreateReq[cloud.GcpVpcCreateExt]{
			Vpcs: createResources,
		}
		if _, err = v.cs.DataService().Gcp.Vpc.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq); err != nil {
			logs.Errorf("[%s-vpc]batch compare db create failed. accountID:%s, region:%s, err: %v",
				enumor.Gcp, req.AccountID, req.Region, err)
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
			logs.Errorf("[%s-vpc]batch compare db delete failed. accountID:%s, region:%s, delIDs:%v, err: %v",
				enumor.Gcp, req.AccountID, req.Region, deleteIDs, err)
			return err
		}
	}

	return nil
}

// filterGcpVpcList filter gcp vpc list
func (v vpc) filterGcpVpcList(req *hcservice.ResourceSyncReq, list *types.GcpVpcListResult,
	resourceDBMap map[string]cloudcore.BaseVpc) (createResources []cloud.VpcCreateReq[cloud.GcpVpcCreateExt],
	updateResources []cloud.VpcUpdateReq[cloud.GcpVpcUpdateExt], existIDMap map[string]bool, err error) {
	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpclist is empty, accountID:%s, region:%s", req.AccountID, req.Region)
	}

	existIDMap = make(map[string]bool, 0)
	for _, item := range list.Details {
		// need compare and update vpc data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			tmpRes := cloud.VpcUpdateReq[cloud.GcpVpcUpdateExt]{
				ID: resourceInfo.ID,
				Extension: &cloud.GcpVpcUpdateExt{
					EnableUlaInternalIpv6: converter.ValToPtr(item.Extension.EnableUlaInternalIpv6),
					Mtu:                   item.Extension.Mtu,
					RoutingMode:           item.Extension.RoutingMode,
				},
			}
			tmpRes.Name = converter.ValToPtr(item.Name)
			tmpRes.Memo = item.Memo

			updateResources = append(updateResources, tmpRes)
			existIDMap[resourceInfo.ID] = true
		} else {
			// need add vpc data
			tmpRes := cloud.VpcCreateReq[cloud.GcpVpcCreateExt]{
				AccountID: req.AccountID,
				CloudID:   item.CloudID,
				Name:      converter.ValToPtr(item.Name),
				Category:  enumor.BizVpcCategory,
				Memo:      item.Memo,
				Extension: &cloud.GcpVpcCreateExt{
					AutoCreateSubnetworks: item.Extension.AutoCreateSubnetworks,
					EnableUlaInternalIpv6: item.Extension.EnableUlaInternalIpv6,
					Mtu:                   item.Extension.Mtu,
					RoutingMode:           item.Extension.RoutingMode,
				},
			}
			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, existIDMap, nil
}

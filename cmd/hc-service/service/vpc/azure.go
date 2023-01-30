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

// AzureVpcUpdate update azure vpc.
func (v vpc) AzureVpcUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.VpcUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := v.cs.DataService().Azure.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.Azure(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := new(types.AzureVpcUpdateOption)
	err = cli.UpdateVpc(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.VpcBatchUpdateReq[cloud.AzureVpcUpdateExt]{
		Vpcs: []cloud.VpcUpdateReq[cloud.AzureVpcUpdateExt]{{
			ID: id,
			VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = v.cs.DataService().Azure.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// AzureVpcDelete delete azure vpc.
func (v vpc) AzureVpcDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := v.cs.DataService().Azure.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.Azure(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	delOpt := &adcore.AzureDeleteOption{
		BaseDeleteOption:  adcore.BaseDeleteOption{ResourceID: getRes.CloudID},
		ResourceGroupName: getRes.Extension.ResourceGroup,
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

// AzureVpcSync sync azure cloud vpc.
func (v vpc) AzureVpcSync(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.ResourceSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if len(req.ResourceGroupName) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("resource_group_name is required"))
	}

	var (
		vendorName = enumor.Azure
		rsp        = hcservice.ResourceSyncResult{
			TaskID: uuid.UUID(),
		}
	)

	// batch get vpc list from cloudapi.
	list, err := v.BatchGetAzureVpcList(cts, req)
	if err != nil || list == nil {
		logs.Errorf("[%s-vpc] request cloudapi response failed. accountID:%s, region:%s, err:%v",
			vendorName, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch get vpc map from db.
	resourceDBMap, err := v.BatchGetVpcMapFromDB(cts, req, vendorName)
	if err != nil {
		logs.Errorf("[%s-vpc] batch get vpcdblist failed. accountID:%s, region:%s, err:%v",
			vendorName, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch compare vendor vpc list.
	_, err = v.BatchCompareAzureVpcList(cts, req, list, resourceDBMap)
	if err != nil {
		logs.Errorf("[%s-vpc] compare api and dblist failed. accountID:%s, region:%s, err:%v",
			vendorName, req.AccountID, req.Region, err)
		return nil, err
	}

	return rsp, nil
}

// BatchGetAzureVpcList batch get vpc list from cloudapi.
func (v vpc) BatchGetAzureVpcList(cts *rest.Contexts, req *hcservice.ResourceSyncReq) (
	*types.AzureVpcListResult, error) {
	cli, err := v.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adcore.AzureListOption{
		ResourceGroupName: req.ResourceGroupName,
	}
	list, err := cli.ListVpc(cts.Kit, opt)
	if err != nil || list == nil {
		logs.Errorf("[%s-vpc]batch get cloud api failed. accountID:%s, region:%s, err:%v",
			enumor.Azure, req.AccountID, req.Region, err)
		return nil, err
	}
	return list, nil
}

// BatchCompareAzureVpcList batch compare vendor vpc list.
func (v vpc) BatchCompareAzureVpcList(cts *rest.Contexts, req *hcservice.ResourceSyncReq,
	list *types.AzureVpcListResult, resourceDBMap map[string]cloudcore.BaseVpc) (interface{}, error) {
	var (
		createResources []cloud.VpcCreateReq[cloud.AzureVpcCreateExt]
		updateResources []cloud.VpcUpdateReq[cloud.AzureVpcUpdateExt]
		existIDMap      = map[string]bool{}
		deleteIDs       []string
	)

	err := v.filterAzureVpcList(req, list, resourceDBMap, &createResources, &updateResources, existIDMap)
	if err != nil {
		return nil, err
	}

	// update resource data
	if len(updateResources) > 0 {
		if err = v.cs.DataService().Azure.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(),
			&cloud.VpcBatchUpdateReq[cloud.AzureVpcUpdateExt]{
				Vpcs: updateResources,
			}); err != nil {
			logs.Errorf("[%s-vpc]batch compare db update failed. accountID:%s, region:%s, err:%v",
				enumor.Azure, req.AccountID, req.Region, err)
			return nil, err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		if _, err = v.cs.DataService().Azure.Vpc.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(),
			&cloud.VpcBatchCreateReq[cloud.AzureVpcCreateExt]{
				Vpcs: createResources,
			}); err != nil {
			logs.Errorf("[%s-vpc]batch compare db create failed. accountID:%s, region:%s, err:%v",
				enumor.Azure, req.AccountID, req.Region, err)
			return nil, err
		}
	}

	// delete resource data
	for _, resItem := range resourceDBMap {
		if _, ok := existIDMap[resItem.ID]; !ok {
			deleteIDs = append(deleteIDs, resItem.ID)
		}
	}
	if len(deleteIDs) > 0 {
		if err = v.cs.DataService().Global.Vpc.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("id", deleteIDs),
		}); err != nil {
			logs.Errorf("[%s-vpc]batch compare db delete failed. accountID:%s, region:%s, delIDs:%v, err:%v",
				enumor.Azure, req.AccountID, req.Region, deleteIDs, err)
			return nil, err
		}
	}
	return nil, nil
}

func (v vpc) filterAzureVpcList(req *hcservice.ResourceSyncReq, list *types.AzureVpcListResult,
	resourceDBMap map[string]cloudcore.BaseVpc, createResources *[]cloud.VpcCreateReq[cloud.AzureVpcCreateExt],
	updateResources *[]cloud.VpcUpdateReq[cloud.AzureVpcUpdateExt], existIDMap map[string]bool) error {
	if list == nil || len(list.Details) == 0 {
		return fmt.Errorf("cloudapi vpclist is empty, accountID:%s, region:%s", req.AccountID, req.Region)
	}

	for _, item := range list.Details {
		// need compare and update vpc data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			tmpRes := cloud.VpcUpdateReq[cloud.AzureVpcUpdateExt]{
				ID: resourceInfo.ID,
				Extension: &cloud.AzureVpcUpdateExt{
					DNSServers: item.Extension.DNSServers,
				},
			}
			tmpRes.Name = converter.ValToPtr(item.Name)
			tmpRes.Memo = item.Memo

			if item.Extension.Cidr != nil {
				tmpCidrs := []cloud.AzureCidr{}
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.AzureCidr{
						Type: cidrItem.Type,
						Cidr: cidrItem.Cidr,
					})
				}
				tmpRes.Extension.Cidr = tmpCidrs
			}
			*updateResources = append(*updateResources, tmpRes)
			existIDMap[resourceInfo.ID] = true
		} else {
			// need add vpc data
			tmpRes := cloud.VpcCreateReq[cloud.AzureVpcCreateExt]{
				AccountID: req.AccountID,
				CloudID:   item.CloudID,
				Name:      converter.ValToPtr(item.Name),
				Category:  enumor.BizVpcCategory,
				Memo:      item.Memo,
				Extension: &cloud.AzureVpcCreateExt{
					ResourceGroup: item.Extension.ResourceGroup,
					Region:        item.Extension.Region,
					DNSServers:    item.Extension.DNSServers,
				},
			}

			if item.Extension.Cidr != nil {
				tmpCidrs := []cloud.AzureCidr{}
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.AzureCidr{
						Type: cidrItem.Type,
						Cidr: cidrItem.Cidr,
					})
				}
				tmpRes.Extension.Cidr = tmpCidrs
			}
			*createResources = append(*createResources, tmpRes)
		}
	}
	return nil
}

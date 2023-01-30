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
<<<<<<< HEAD
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

// HuaWeiSubnetUpdate update huawei subnet.
func (v subnet) HuaWeiSubnetUpdate(cts *rest.Contexts) (interface{}, error) {
=======
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/rest"
)

// HuaWeiSubnetUpdate update huawei subnet.
func (s subnet) HuaWeiSubnetUpdate(cts *rest.Contexts) (interface{}, error) {
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41
	id := cts.PathParameter("id").String()

	req := new(hcservice.SubnetUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

<<<<<<< HEAD
	getRes, err := v.cs.DataService().HuaWei.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
=======
	getRes, err := s.cs.DataService().HuaWei.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41
	if err != nil {
		return nil, err
	}

<<<<<<< HEAD
	cli, err := v.ad.HuaWei(cts.Kit, getRes.AccountID)
=======
	cli, err := s.ad.HuaWei(cts.Kit, getRes.AccountID)
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41
	if err != nil {
		return nil, err
	}

	updateOpt := &types.HuaweiSubnetUpdateOption{
		SubnetUpdateOption: types.SubnetUpdateOption{
			ResourceID: id,
			Data:       &types.BaseSubnetUpdateData{Memo: req.Memo},
		},
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
<<<<<<< HEAD
	err = v.cs.DataService().HuaWei.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
=======
	err = s.cs.DataService().HuaWei.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// HuaWeiSubnetDelete delete huawei subnet.
<<<<<<< HEAD
func (v subnet) HuaWeiSubnetDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := v.cs.DataService().HuaWei.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
=======
func (s subnet) HuaWeiSubnetDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := s.cs.DataService().HuaWei.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41
	if err != nil {
		return nil, err
	}

<<<<<<< HEAD
	cli, err := v.ad.HuaWei(cts.Kit, getRes.AccountID)
=======
	cli, err := s.ad.HuaWei(cts.Kit, getRes.AccountID)
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41
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
<<<<<<< HEAD
	err = v.cs.DataService().Global.Subnet.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
=======
	err = s.cs.DataService().Global.Subnet.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41
	if err != nil {
		return nil, err
	}

	return nil, nil
}
<<<<<<< HEAD

// HuaweiSubnetSync sync huawei cloud subnet.
func (v subnet) HuaweiSubnetSync(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.ResourceSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if len(req.Region) == 0 {
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("region is required"))
	}

	var (
		vendorName = enumor.HuaWei
		rsp        = hcservice.ResourceSyncResult{
			TaskID: uuid.UUID(),
		}
	)

	// batch get subnet list from cloudapi.
	list, err := v.BatchGetHuaWeiSubnetList(cts, req)
	if err != nil || list == nil {
		logs.Errorf("[%s-subnet] request cloudapi response failed. accountID:%s, region:%s, err:%v",
			vendorName, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch get subnet map from db.
	resourceDBMap, err := v.BatchGetSubnetMapFromDB(cts, req, vendorName)
	if err != nil {
		logs.Errorf("[%s-subnet] batch get subnetdblist failed. accountID:%s, region:%s, err:%v",
			vendorName, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch compare vendor subnet list.
	_, err = v.BatchCompareHuaWeiSubnetList(cts, req, list, resourceDBMap)
	if err != nil {
		logs.Errorf("[%s-subnet] compare api and subnetdblist failed. accountID:%s, region:%s, err:%v",
			vendorName, req.AccountID, req.Region, err)
		return nil, err
	}

	return rsp, nil
}

// BatchGetHuaWeiSubnetList batch get subnet list from cloudapi.
func (v subnet) BatchGetHuaWeiSubnetList(cts *rest.Contexts, req *hcservice.ResourceSyncReq) (
	*types.HuaweiSubnetListResult, error) {
	var (
		count int32 = adcore.HuaweiQueryLimit
		list        = &types.HuaweiSubnetListResult{}
	)

	cli, err := v.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	for {
		opt := &types.HuaweiSubnetListOption{}
		opt.Region = req.Region
		opt.Page = &adcore.HuaweiPage{
			Limit: converter.ValToPtr(count),
		}
		tmpList, tmpErr := cli.ListSubnet(cts.Kit, opt)
		if tmpErr != nil || tmpList == nil {
			logs.Errorf("[%s-subnet]batch get cloud api failed. accountID:%s, region:%s, err:%v",
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

// BatchCompareHuaWeiSubnetList batch compare vendor subnet list.
func (v subnet) BatchCompareHuaWeiSubnetList(cts *rest.Contexts, req *hcservice.ResourceSyncReq,
	list *types.HuaweiSubnetListResult, resourceDBMap map[string]cloudcore.BaseSubnet) (interface{}, error) {
	var (
		createResources []cloud.SubnetCreateReq[cloudcore.HuaWeiSubnetExtension]
		updateResources []cloud.SubnetUpdateReq[cloud.HuaWeiSubnetUpdateExt]
		existIDMap      = map[string]bool{}
		deleteIDs       []string
	)

	err := v.filterHuaWeiSubnetList(req, list, resourceDBMap, &createResources, &updateResources, existIDMap)
	if err != nil {
		return nil, err
	}

	// update resource data
	if len(updateResources) > 0 {
		if err = v.cs.DataService().HuaWei.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(),
			&cloud.SubnetBatchUpdateReq[cloud.HuaWeiSubnetUpdateExt]{
				Subnets: updateResources,
			}); err != nil {
			logs.Errorf("[%s-subnet]batch compare db update failed. accountID:%s, region:%s, err:%v",
				enumor.HuaWei, req.AccountID, req.Region, err)
			return nil, err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		if _, err = v.cs.DataService().HuaWei.Subnet.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(),
			&cloud.SubnetBatchCreateReq[cloudcore.HuaWeiSubnetExtension]{
				Subnets: createResources,
			}); err != nil {
			logs.Errorf("[%s-subnet]batch compare db create failed. accountID:%s, region:%s, err:%v",
				enumor.HuaWei, req.AccountID, req.Region, err)
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
		if err = v.cs.DataService().Global.Subnet.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(),
			&dataservice.BatchDeleteReq{
				Filter: tools.ContainersExpression("id", deleteIDs),
			}); err != nil {
			logs.Errorf("[%s-subnet]batch compare db delete failed. accountID:%s, region:%s, delIDs:%v, err:%v",
				enumor.HuaWei, req.AccountID, req.Region, deleteIDs, err)
			return nil, err
		}
	}
	return nil, nil
}

func (v subnet) filterHuaWeiSubnetList(req *hcservice.ResourceSyncReq, list *types.HuaweiSubnetListResult,
	resourceDBMap map[string]cloudcore.BaseSubnet,
	createResources *[]cloud.SubnetCreateReq[cloudcore.HuaWeiSubnetExtension],
	updateResources *[]cloud.SubnetUpdateReq[cloud.HuaWeiSubnetUpdateExt], existIDMap map[string]bool) error {
	if list == nil || len(list.Details) == 0 {
		return fmt.Errorf("cloudapi subnetlist is empty, accountID:%s, region:%s", req.AccountID, req.Region)
	}

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

			*updateResources = append(*updateResources, tmpRes)
			existIDMap[resourceInfo.ID] = true
		} else {
			// need add subnet data
			tmpRes := cloud.SubnetCreateReq[cloudcore.HuaWeiSubnetExtension]{
				AccountID:  req.AccountID,
				CloudVpcID: item.CloudVpcID,
				CloudID:    item.CloudID,
				Name:       converter.ValToPtr(item.Name),
				Ipv4Cidr:   item.Ipv4Cidr,
				Memo:       item.Memo,
				Extension: &cloudcore.HuaWeiSubnetExtension{
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
			*createResources = append(*createResources, tmpRes)
		}
	}
	return nil
}
=======
>>>>>>> 304144ec282c951c6c2127f39ca83cb7f1c70b41

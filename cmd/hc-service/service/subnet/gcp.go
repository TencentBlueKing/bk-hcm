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

// GcpSubnetUpdate update gcp subnet.
func (s subnet) GcpSubnetUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.SubnetUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := s.cs.DataService().Gcp.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.Gcp(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := &types.GcpSubnetUpdateOption{
		SubnetUpdateOption: types.SubnetUpdateOption{
			ResourceID: getRes.CloudID,
			Data:       &types.BaseSubnetUpdateData{Memo: req.Memo},
		},
		Region: getRes.Extension.Region,
	}
	err = cli.UpdateSubnet(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.SubnetBatchUpdateReq[cloud.GcpSubnetUpdateExt]{
		Subnets: []cloud.SubnetUpdateReq[cloud.GcpSubnetUpdateExt]{{
			ID: id,
			SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = s.cs.DataService().Gcp.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// GcpSubnetDelete delete gcp subnet.
func (s subnet) GcpSubnetDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := s.cs.DataService().Gcp.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.Gcp(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	if getRes.Extension == nil {
		return nil, errf.New(errf.InvalidParameter, "subnet extension is empty")
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

// GcpSubnetSync sync gcp cloud subnet.
func (s subnet) GcpSubnetSync(cts *rest.Contexts) (interface{}, error) {
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
		vendorName = enumor.Gcp
		rsp        = hcservice.ResourceSyncResult{
			TaskID: uuid.UUID(),
		}
	)

	// batch get subnet list from cloudapi.
	list, err := s.BatchGetGcpSubnetList(cts, req)
	if err != nil || list == nil {
		logs.Errorf("[%s-subnet] request cloudapi response failed. accountID:%s, region:%s, err:%v",
			vendorName, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch get subnet map from db.
	resourceDBMap, err := s.BatchGetSubnetMapFromDB(cts, req, vendorName, "")
	if err != nil {
		logs.Errorf("[%s-subnet] batch get subnetdblist failed. accountID:%s, region:%s, err:%v",
			vendorName, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch compare vendor subnet list.
	_, err = s.BatchCompareGcpSubnetList(cts, req, list, resourceDBMap)
	if err != nil {
		logs.Errorf("[%s-subnet] compare api and dblist failed. accountID:%s, region:%s, err:%v",
			vendorName, req.AccountID, req.Region, err)
		return nil, err
	}

	return rsp, nil
}

// BatchGetGcpSubnetList batch get subnet list from cloudapi.
func (s subnet) BatchGetGcpSubnetList(cts *rest.Contexts, req *hcservice.ResourceSyncReq) (
	*types.GcpSubnetListResult, error) {
	var (
		nextToken string
		count     int64 = adcore.GcpQueryLimit
		list            = &types.GcpSubnetListResult{}
	)

	cli, err := s.ad.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	for {
		opt := &types.GcpSubnetListOption{
			Region: req.Region,
		}
		opt.Page = &adcore.GcpPage{
			PageSize: count,
		}
		if nextToken != "" {
			opt.Page.PageToken = nextToken
		}
		tmpList, tmpErr := cli.ListSubnet(cts.Kit, opt)
		if tmpErr != nil || tmpList == nil {
			logs.Errorf("[%s-subnet]batch get cloud api failed. accountID:%s, region:%s, nextToken:%s, err:%v",
				enumor.Gcp, req.AccountID, req.Region, nextToken, tmpErr)
			return nil, tmpErr
		}

		list.Details = append(list.Details, tmpList.Details...)
		if len(tmpList.Details) == 0 || len(tmpList.NextPageToken) == 0 {
			break
		}
		nextToken = tmpList.NextPageToken
	}
	return list, nil
}

// BatchCompareGcpSubnetList batch compare vendor subnet list.
func (s subnet) BatchCompareGcpSubnetList(cts *rest.Contexts, req *hcservice.ResourceSyncReq,
	list *types.GcpSubnetListResult, resourceDBMap map[string]cloudcore.BaseSubnet) (interface{}, error) {
	var (
		createResources []cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt]
		updateResources []cloud.SubnetUpdateReq[cloud.GcpSubnetUpdateExt]
		existIDMap      = map[string]bool{}
		deleteIDs       []string
	)

	err := s.filterGcpSubnetList(req, list, resourceDBMap, &createResources, &updateResources, existIDMap)
	if err != nil {
		return nil, err
	}

	// update resource data
	if len(updateResources) > 0 {
		if err = s.cs.DataService().Gcp.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(),
			&cloud.SubnetBatchUpdateReq[cloud.GcpSubnetUpdateExt]{
				Subnets: updateResources,
			}); err != nil {
			logs.Errorf("[%s-subnet]batch compare db update failed. accountID:%s, region:%s, err:%v",
				enumor.Gcp, req.AccountID, req.Region, err)
			return nil, err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		err = s.batchCreateGcpSubnet(cts, createResources)
		if err != nil {
			logs.Errorf("[%s-subnet]batch compare db create failed. accountID:%s, region:%s, err:%v",
				enumor.Gcp, req.AccountID, req.Region, err)
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
		if err = s.cs.DataService().Global.Subnet.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(),
			&dataservice.BatchDeleteReq{
				Filter: tools.ContainersExpression("id", deleteIDs),
			}); err != nil {
			logs.Errorf("[%s-subnet]batch compare db delete failed. accountID:%s, region:%s, delIDs:%v, err:%v",
				enumor.Gcp, req.AccountID, req.Region, deleteIDs, err)
			return nil, err
		}
	}
	return nil, nil
}

// filterGcpSubnetList filter gcp subnet
func (s subnet) filterGcpSubnetList(req *hcservice.ResourceSyncReq, list *types.GcpSubnetListResult,
	resourceDBMap map[string]cloudcore.BaseSubnet,
	createResources *[]cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt],
	updateResources *[]cloud.SubnetUpdateReq[cloud.GcpSubnetUpdateExt], existIDMap map[string]bool) error {
	if list == nil || len(list.Details) == 0 {
		return fmt.Errorf("cloudapi subnetlist is empty, accountID:%s, region:%s", req.AccountID, req.Region)
	}

	for _, item := range list.Details {
		// need compare and update subnet data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			tmpRes := cloud.SubnetUpdateReq[cloud.GcpSubnetUpdateExt]{
				ID: resourceInfo.ID,
				Extension: &cloud.GcpSubnetUpdateExt{
					StackType:             item.Extension.StackType,
					Ipv6AccessType:        item.Extension.Ipv6AccessType,
					GatewayAddress:        item.Extension.GatewayAddress,
					PrivateIpGoogleAccess: converter.ValToPtr(item.Extension.PrivateIpGoogleAccess),
					EnableFlowLogs:        converter.ValToPtr(item.Extension.EnableFlowLogs),
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
			tmpRes := cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt]{
				AccountID:  req.AccountID,
				CloudVpcID: item.CloudVpcID,
				CloudID:    item.CloudID,
				Name:       converter.ValToPtr(item.Name),
				Ipv4Cidr:   item.Ipv4Cidr,
				Memo:       item.Memo,
				Extension: &cloud.GcpSubnetCreateExt{
					Region:                item.Extension.Region,
					StackType:             item.Extension.StackType,
					Ipv6AccessType:        item.Extension.Ipv6AccessType,
					GatewayAddress:        item.Extension.GatewayAddress,
					PrivateIpGoogleAccess: item.Extension.PrivateIpGoogleAccess,
					EnableFlowLogs:        item.Extension.EnableFlowLogs,
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

func (s subnet) batchCreateGcpSubnet(cts *rest.Contexts,
	createResources []cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt]) error {
	querySize := int(filter.DefaultMaxInLimit)
	times := len(createResources) / querySize
	if len(createResources)%querySize != 0 {
		times++
	}
	for i := 0; i < times; i++ {
		var newResources []cloud.SubnetCreateReq[cloud.GcpSubnetCreateExt]
		if i == times-1 {
			newResources = append(newResources, createResources[i*querySize:]...)
		} else {
			newResources = append(newResources, createResources[i*querySize:(i+1)*querySize]...)
		}

		if _, err := s.cs.DataService().Gcp.Subnet.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(),
			&cloud.SubnetBatchCreateReq[cloud.GcpSubnetCreateExt]{
				Subnets: newResources,
			}); err != nil {
			return err
		}
	}
	return nil
}

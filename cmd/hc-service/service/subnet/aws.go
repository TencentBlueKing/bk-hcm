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
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// AwsSubnetUpdate update aws subnet.
func (s subnet) AwsSubnetUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.SubnetUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := s.cs.DataService().Aws.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.Aws(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := new(types.AwsSubnetUpdateOption)
	err = cli.UpdateSubnet(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.SubnetBatchUpdateReq[cloud.AwsSubnetUpdateExt]{
		Subnets: []cloud.SubnetUpdateReq[cloud.AwsSubnetUpdateExt]{{
			ID: id,
			SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = s.cs.DataService().Aws.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// AwsSubnetDelete delete aws subnet.
func (s subnet) AwsSubnetDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := s.cs.DataService().Aws.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := s.ad.Aws(cts.Kit, getRes.AccountID)
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

// AwsSubnetSync sync aws cloud subnet.
func (s subnet) AwsSubnetSync(cts *rest.Contexts) (interface{}, error) {
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
		vendorName = enumor.Aws
		rsp        = hcservice.ResourceSyncResult{
			TaskID: uuid.UUID(),
		}
	)

	// batch get subnet list from cloudapi.
	list, err := s.BatchGetAwsSubnetList(cts, req)
	if err != nil || list == nil {
		logs.Errorf("[%s-subnet] request cloudapi response failed. accountID:%s, region:%s, err:%v",
			vendorName, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch get subnet map from db.
	resourceDBMap, err := s.BatchGetSubnetMapFromDB(cts, req, vendorName)
	if err != nil {
		logs.Errorf("[%s-subnet] batch get vpcdblist failed. accountID:%s, region:%s, err:%v",
			vendorName, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch compare vendor subnet list.
	_, err = s.BatchCompareAwsSubnetList(cts, req, list, resourceDBMap)
	if err != nil {
		logs.Errorf("[%s-subnet] compare api and dblist failed. accountID:%s, region:%s, err:%v",
			vendorName, req.AccountID, req.Region, err)
		return nil, err
	}

	return rsp, nil
}

// BatchGetAwsSubnetList batch get subnet list from cloudapi.
func (s subnet) BatchGetAwsSubnetList(cts *rest.Contexts, req *hcservice.ResourceSyncReq) (
	*types.AwsSubnetListResult, error) {
	var (
		nextToken string
		count     int64 = adcore.AwsQueryLimit
		list            = &types.AwsSubnetListResult{}
	)

	cli, err := s.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	for {
		opt := &adcore.AwsListOption{}
		opt.Region = req.Region
		opt.Page = &adcore.AwsPage{
			MaxResults: converter.ValToPtr(count),
		}
		if nextToken != "" {
			opt.Page.NextToken = converter.ValToPtr(nextToken)
		}
		tmpList, tmpErr := cli.ListSubnet(cts.Kit, opt)
		if tmpErr != nil || tmpList == nil {
			logs.Errorf("[%s-subnet]batch get cloud api failed. accountID:%s, region:%s, nextToken:%s, err:%v",
				enumor.Aws, req.AccountID, req.Region, nextToken, tmpErr)
			return nil, tmpErr
		}

		list.Details = append(list.Details, tmpList.Details...)
		if len(tmpList.Details) == 0 || tmpList.NextToken == nil {
			break
		}
		nextToken = *tmpList.NextToken
	}
	return list, nil
}

// BatchCompareAwsSubnetList batch compare vendor subnet list.
func (s subnet) BatchCompareAwsSubnetList(cts *rest.Contexts, req *hcservice.ResourceSyncReq,
	list *types.AwsSubnetListResult, resourceDBMap map[string]cloudcore.BaseSubnet) (interface{}, error) {
	var (
		createResources []cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt]
		updateResources []cloud.SubnetUpdateReq[cloud.AwsSubnetUpdateExt]
		existIDMap      = map[string]bool{}
		deleteIDs       []string
	)

	err := s.filterAwsSubnetList(req, list, resourceDBMap, &createResources, &updateResources, existIDMap)
	if err != nil {
		return nil, err
	}

	// update resource data
	if len(updateResources) > 0 {
		if err = s.cs.DataService().Aws.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(),
			&cloud.SubnetBatchUpdateReq[cloud.AwsSubnetUpdateExt]{
				Subnets: updateResources,
			}); err != nil {
			logs.Errorf("[%s-subnet]batch compare db update failed. accountID:%s, region:%s, err:%v",
				enumor.Aws, req.AccountID, req.Region, err)
			return nil, err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		if _, err = s.cs.DataService().Aws.Subnet.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(),
			&cloud.SubnetBatchCreateReq[cloud.AwsSubnetCreateExt]{
				Subnets: createResources,
			}); err != nil {
			logs.Errorf("[%s-subnet]batch compare db create failed. accountID:%s, region:%s, err:%v",
				enumor.Aws, req.AccountID, req.Region, err)
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
				enumor.Aws, req.AccountID, req.Region, deleteIDs, err)
			return nil, err
		}
	}
	return nil, nil
}

func (s subnet) filterAwsSubnetList(req *hcservice.ResourceSyncReq, list *types.AwsSubnetListResult,
	resourceDBMap map[string]cloudcore.BaseSubnet,
	createResources *[]cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt],
	updateResources *[]cloud.SubnetUpdateReq[cloud.AwsSubnetUpdateExt], existIDMap map[string]bool) error {
	if list == nil || len(list.Details) == 0 {
		return fmt.Errorf("cloudapi vpclist is empty, accountID:%s, region:%s", req.AccountID, req.Region)
	}

	for _, item := range list.Details {
		// need compare and update subnet data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			tmpRes := cloud.SubnetUpdateReq[cloud.AwsSubnetUpdateExt]{
				ID: resourceInfo.ID,
				Extension: &cloud.AwsSubnetUpdateExt{
					State:                       item.Extension.State,
					Region:                      item.Extension.Region,
					Zone:                        item.Extension.Zone,
					IsDefault:                   converter.ValToPtr(item.Extension.IsDefault),
					MapPublicIpOnLaunch:         converter.ValToPtr(item.Extension.MapPublicIpOnLaunch),
					AssignIpv6AddressOnCreation: converter.ValToPtr(item.Extension.AssignIpv6AddressOnCreation),
					HostnameType:                item.Extension.HostnameType,
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
			tmpRes := cloud.SubnetCreateReq[cloud.AwsSubnetCreateExt]{
				AccountID:  req.AccountID,
				CloudVpcID: item.CloudVpcID,
				CloudID:    item.CloudID,
				Name:       converter.ValToPtr(item.Name),
				Ipv4Cidr:   item.Ipv4Cidr,
				Memo:       item.Memo,
				Extension: &cloud.AwsSubnetCreateExt{
					State:                       item.Extension.State,
					Region:                      item.Extension.Region,
					Zone:                        item.Extension.Zone,
					IsDefault:                   item.Extension.IsDefault,
					MapPublicIpOnLaunch:         item.Extension.MapPublicIpOnLaunch,
					AssignIpv6AddressOnCreation: item.Extension.AssignIpv6AddressOnCreation,
					HostnameType:                item.Extension.HostnameType,
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

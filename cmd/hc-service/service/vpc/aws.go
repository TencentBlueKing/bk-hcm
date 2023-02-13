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

// AwsVpcUpdate update aws vpc.
func (v vpc) AwsVpcUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.VpcUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := v.cs.DataService().Aws.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.Aws(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := new(types.AwsVpcUpdateOption)
	err = cli.UpdateVpc(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.VpcBatchUpdateReq[cloud.AwsVpcUpdateExt]{
		Vpcs: []cloud.VpcUpdateReq[cloud.AwsVpcUpdateExt]{{
			ID: id,
			VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = v.cs.DataService().Aws.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// AwsVpcDelete delete aws vpc.
func (v vpc) AwsVpcDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := v.cs.DataService().Aws.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.Aws(cts.Kit, getRes.AccountID)
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

// AwsVpcSync sync aws cloud vpc.
func (v vpc) AwsVpcSync(cts *rest.Contexts) (interface{}, error) {
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

	// batch get vpc list from cloudapi.
	list, err := v.BatchGetAwsVpcList(cts, req)
	if err != nil {
		logs.Errorf("%s-vpc request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.Aws, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch get vpc map from db.
	resourceDBMap, err := v.BatchGetVpcMapFromDB(cts, req, enumor.Aws)
	if err != nil {
		logs.Errorf("%s-vpc batch get vpcdblist failed. accountID: %s, region: %s, err: %v",
			enumor.Aws, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch sync vendor vpc list.
	err = v.BatchSyncAwsVpcList(cts, req, list, resourceDBMap)
	if err != nil {
		logs.Errorf("%s-vpc compare api and dblist failed. accountID: %s, region: %s, err: %v",
			enumor.Aws, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// BatchGetAwsVpcList batch get vpc list from cloudapi.
func (v vpc) BatchGetAwsVpcList(cts *rest.Contexts, req *hcservice.ResourceSyncReq) (*types.AwsVpcListResult, error) {
	cli, err := v.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	list := new(types.AwsVpcListResult)
	for {
		opt := new(adcore.AwsListOption)
		opt.Region = req.Region
		opt.Page = &adcore.AwsPage{
			MaxResults: converter.ValToPtr(int64(adcore.AwsQueryLimit)),
		}

		if nextToken != "" {
			opt.Page.NextToken = converter.ValToPtr(nextToken)
		}

		tmpList, tmpErr := cli.ListVpc(cts.Kit, opt)
		if tmpErr != nil {
			logs.Errorf("%s-vpc batch get cloud api failed. accountID: %s, region: %s, nextToken: %s, err: %v",
				enumor.Aws, req.AccountID, req.Region, nextToken, tmpErr)
			return nil, tmpErr
		}

		// traversal vpclist supply fields
		for _, item := range tmpList.Details {
			dnsHostnames, dnsSupport, dnsErr := cli.GetVpcAttribute(cts.Kit, item.CloudID, item.Region)
			if dnsErr == nil {
				item.Extension.EnableDnsHostnames = dnsHostnames
				item.Extension.EnableDnsSupport = dnsSupport
			}
		}

		if len(tmpList.Details) == 0 {
			break
		}

		list.Details = append(list.Details, tmpList.Details...)
		if tmpList.NextToken == nil {
			break
		}
		nextToken = *tmpList.NextToken
	}

	return list, nil
}

// BatchSyncAwsVpcList batch sync vendor vpc list.
func (v vpc) BatchSyncAwsVpcList(cts *rest.Contexts, req *hcservice.ResourceSyncReq, list *types.AwsVpcListResult,
	resourceDBMap map[string]cloudcore.BaseVpc) error {
	createResources, updateResources, existIDMap, err := v.filterAwsVpcList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.VpcBatchUpdateReq[cloud.AwsVpcUpdateExt]{
			Vpcs: updateResources,
		}
		if err = v.cs.DataService().Aws.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
			logs.Errorf("%s-vpc batch compare db update failed. accountID: %s, region: %s, err: %v",
				enumor.Aws, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &cloud.VpcBatchCreateReq[cloud.AwsVpcCreateExt]{
			Vpcs: createResources,
		}
		if _, err = v.cs.DataService().Aws.Vpc.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq); err != nil {
			logs.Errorf("%s-vpc batch compare db create failed. accountID: %s, region: %s, err: %v",
				enumor.Aws, req.AccountID, req.Region, err)
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
				enumor.Aws, req.AccountID, req.Region, deleteIDs, err)
			return err
		}
	}

	return nil
}

// filterAwsVpcList filter aws vpc list
func (v vpc) filterAwsVpcList(req *hcservice.ResourceSyncReq, list *types.AwsVpcListResult,
	resourceDBMap map[string]cloudcore.BaseVpc) (createResources []cloud.VpcCreateReq[cloud.AwsVpcCreateExt],
	updateResources []cloud.VpcUpdateReq[cloud.AwsVpcUpdateExt], existIDMap map[string]bool, err error) {
	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpclist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	existIDMap = make(map[string]bool, 0)
	for _, item := range list.Details {
		// need compare and update vpc data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			tmpRes := cloud.VpcUpdateReq[cloud.AwsVpcUpdateExt]{
				ID: resourceInfo.ID,
				Extension: &cloud.AwsVpcUpdateExt{
					State:              item.Extension.State,
					InstanceTenancy:    converter.ValToPtr(item.Extension.InstanceTenancy),
					IsDefault:          converter.ValToPtr(item.Extension.IsDefault),
					EnableDnsHostnames: converter.ValToPtr(item.Extension.EnableDnsHostnames),
					EnableDnsSupport:   converter.ValToPtr(item.Extension.EnableDnsSupport),
				},
			}
			tmpRes.Name = converter.ValToPtr(item.Name)
			tmpRes.Memo = item.Memo

			if item.Extension.Cidr != nil {
				tmpCidrs := []cloud.AwsCidr{}
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.AwsCidr{
						Type:        cidrItem.Type,
						Cidr:        cidrItem.Cidr,
						AddressPool: cidrItem.AddressPool,
						State:       cidrItem.State,
					})
				}
				tmpRes.Extension.Cidr = tmpCidrs
			}

			updateResources = append(updateResources, tmpRes)
			existIDMap[resourceInfo.ID] = true
		} else {
			// need add vpc data
			tmpRes := cloud.VpcCreateReq[cloud.AwsVpcCreateExt]{
				AccountID: req.AccountID,
				CloudID:   item.CloudID,
				Name:      converter.ValToPtr(item.Name),
				Region:    item.Region,
				Category:  enumor.BizVpcCategory,
				Memo:      item.Memo,
				Extension: &cloud.AwsVpcCreateExt{
					State:              item.Extension.State,
					InstanceTenancy:    item.Extension.InstanceTenancy,
					IsDefault:          item.Extension.IsDefault,
					EnableDnsHostnames: item.Extension.EnableDnsHostnames,
					EnableDnsSupport:   item.Extension.EnableDnsSupport,
				},
			}

			if item.Extension.Cidr != nil {
				tmpCidrs := []cloud.AwsCidr{}
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.AwsCidr{
						Type:        cidrItem.Type,
						Cidr:        cidrItem.Cidr,
						AddressPool: cidrItem.AddressPool,
						State:       cidrItem.State,
					})
				}
				tmpRes.Extension.Cidr = tmpCidrs
			}

			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, existIDMap, nil
}

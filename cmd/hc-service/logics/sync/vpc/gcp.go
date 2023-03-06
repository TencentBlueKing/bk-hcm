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
	cloudcore "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// GcpVpcSync sync gcp cloud vpc.
func GcpVpcSync(kt *kit.Kit, req *hcservice.ResourceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	// batch get vpc list from cloudapi.
	list, err := BatchGetGcpVpcList(kt, req, adaptor)
	if err != nil {
		logs.Errorf("%s-vpc request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.Gcp, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch get vpc map from db.
	resourceDBMap, err := BatchGetVpcMapFromDB(kt, req, enumor.Gcp, dataCli)
	if err != nil {
		logs.Errorf("%s-vpc batch get vpcdblist failed. accountID: %s, region: %s, err: %v",
			enumor.Gcp, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch sync vendor vpc list.
	err = BatchSyncGcpVpcList(kt, req, list, resourceDBMap, dataCli)
	if err != nil {
		logs.Errorf("%s-vpc compare api and dblist failed. accountID: %s, region: %s, err: %v",
			enumor.Gcp, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// BatchGetGcpVpcList batch get vpc list from cloudapi.
func BatchGetGcpVpcList(kt *kit.Kit, req *hcservice.ResourceSyncReq, adaptor *cloudclient.CloudAdaptorClient) (
	*types.GcpVpcListResult, error) {

	cli, err := adaptor.Gcp(kt, req.AccountID)
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

		// 查询指定CloudIDs
		if len(req.CloudIDs) > 0 {
			opt.Page = nil
			opt.SelfLinks = req.CloudIDs
		}

		tmpList, tmpErr := cli.ListVpc(kt, opt)
		if tmpErr != nil {
			logs.Errorf("%s-vpc batch get cloud api failed. accountID: %s, region: %s, nextToken: %s, err: %v",
				enumor.Gcp, req.AccountID, req.Region, nextToken, tmpErr)
			return nil, tmpErr
		}

		if len(tmpList.Details) == 0 {
			break
		}

		list.Details = append(list.Details, tmpList.Details...)
		if len(req.CloudIDs) > 0 || len(tmpList.NextPageToken) == 0 {
			break
		}

		nextToken = tmpList.NextPageToken
	}

	return list, nil
}

// BatchSyncGcpVpcList batch sync vendor vpc list.
func BatchSyncGcpVpcList(kt *kit.Kit, req *hcservice.ResourceSyncReq, list *types.GcpVpcListResult,
	resourceDBMap map[string]cloudcore.BaseVpc, dataCli *dataclient.Client) error {

	createResources, updateResources, existIDMap, err := filterGcpVpcList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.VpcBatchUpdateReq[cloud.GcpVpcUpdateExt]{
			Vpcs: updateResources,
		}
		if err = dataCli.Gcp.Vpc.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-vpc batch compare db update failed. accountID: %s, region: %s, err: %v",
				enumor.Gcp, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &cloud.VpcBatchCreateReq[cloud.GcpVpcCreateExt]{
			Vpcs: createResources,
		}
		if _, err = dataCli.Gcp.Vpc.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("%s-vpc batch compare db create failed. accountID: %s, region: %s, err: %v",
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
		err = BatchDeleteVpcByIDs(kt, deleteIDs, dataCli)
		if err != nil {
			logs.Errorf("%s-vpc batch compare db delete failed. accountID: %s, region: %s, delIDs: %v, err: %v",
				enumor.Gcp, req.AccountID, req.Region, deleteIDs, err)
			return err
		}
	}

	return nil
}

// filterGcpVpcList filter gcp vpc list
func filterGcpVpcList(req *hcservice.ResourceSyncReq, list *types.GcpVpcListResult,
	resourceDBMap map[string]cloudcore.BaseVpc) (createResources []cloud.VpcCreateReq[cloud.GcpVpcCreateExt],
	updateResources []cloud.VpcUpdateReq[cloud.GcpVpcUpdateExt], existIDMap map[string]bool, err error) {
	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpclist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	existIDMap = make(map[string]bool, 0)
	for _, item := range list.Details {
		// need compare and update vpc data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if resourceInfo.Name == item.Name && resourceInfo.Region == item.Region &&
				converter.PtrToVal(resourceInfo.Memo) == converter.PtrToVal(item.Memo) {
				existIDMap[resourceInfo.ID] = true
				continue
			}

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

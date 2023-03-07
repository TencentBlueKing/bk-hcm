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
	"errors"
	"fmt"

	"hcm/cmd/hc-service/logics/sync/logics"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// SyncHuaWeiOption define huawei sync option.
type SyncHuaWeiOption struct {
	AccountID  string   `json:"account_id" validate:"required"`
	Region     string   `json:"region" validate:"required"`
	CloudVpcID string   `json:"vpc_id" validate:"required"`
	CloudIDs   []string `json:"cloud_ids" validate:"required"`
}

// Validate SyncHuaWeiOption.
func (opt SyncHuaWeiOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) == 0 {
		return errors.New("cloudIDs is required")
	}

	if len(opt.CloudIDs) > int(core.DefaultMaxPageLimit) {
		return fmt.Errorf("cloudIDs should <= %d", core.DefaultMaxPageLimit)
	}

	return nil
}

// SyncHuaWeiSubnet sync huawei cloud subnet.
func SyncHuaWeiSubnet(kt *kit.Kit, req *SyncHuaWeiOption, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client) (interface{}, error) {

	if len(req.CloudIDs) > 0 && len(req.CloudVpcID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vpc_id is required")
	}

	// batch get subnet list from cloudapi.
	list, err := BatchGetHuaWeiSubnetList(kt, req, adaptor)
	if err != nil {
		logs.Errorf("%s-subnet request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	if list.Details == nil {
		return nil, nil
	}

	// batch get subnet map from db.
	resourceDBMap, err := listHuaWeiSubnetMapFromDB(kt, req.CloudIDs, dataCli)
	if err != nil {
		logs.Errorf("%s-subnet batch get subnetdblist failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch sync vendor subnet list.
	err = BatchSyncHuaWeiSubnetList(kt, req, list, resourceDBMap, dataCli, adaptor)
	if err != nil {
		logs.Errorf("%s-subnet compare api and subnetdblist failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

func listHuaWeiSubnetMapFromDB(kt *kit.Kit, cloudIDs []string, dataCli *dataclient.Client) (
	map[string]cloudcore.Subnet[cloudcore.HuaWeiSubnetExtension], error) {

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "cloud_id",
				Op:    filter.In.Factory(),
				Value: cloudIDs,
			},
		},
	}
	resourceMap := make(map[string]cloudcore.Subnet[cloudcore.HuaWeiSubnetExtension], 0)
	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   core.DefaultBasePage,
	}
	dbList, err := dataCli.HuaWei.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("huawei-subnet list ext db failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// BatchGetHuaWeiSubnetList batch get subnet list from cloudapi.
func BatchGetHuaWeiSubnetList(kt *kit.Kit, req *SyncHuaWeiOption,
	adaptor *cloudclient.CloudAdaptorClient) (*types.HuaWeiSubnetListResult, error) {

	if len(req.CloudIDs) > 0 {
		return BatchGetHuaWeiSubnetListByCloudIDs(kt, req, adaptor)
	}

	return BatchGetHuaWeiSubnetAllList(kt, req, adaptor)
}

// BatchGetHuaWeiSubnetListByCloudIDs batch get subnet list from cloudapi.
func BatchGetHuaWeiSubnetListByCloudIDs(kt *kit.Kit, req *SyncHuaWeiOption,
	adaptor *cloudclient.CloudAdaptorClient) (*types.HuaWeiSubnetListResult, error) {

	cli, err := adaptor.HuaWei(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.HuaWeiSubnetListByIDOption{
		Region:   req.Region,
		VpcID:    req.CloudVpcID,
		CloudIDs: req.CloudIDs,
	}
	list, err := cli.ListSubnetByID(kt, opt)
	if err != nil {
		logs.Errorf("%s-subnet batch get cloud api failed, err: %v, opt: %v, rid: %s",
			enumor.HuaWei, err, opt, kt.Rid)
		return nil, err
	}

	return list, nil
}

// BatchGetHuaWeiSubnetAllList batch get subnet list from cloudapi.
func BatchGetHuaWeiSubnetAllList(kt *kit.Kit, req *SyncHuaWeiOption,
	adaptor *cloudclient.CloudAdaptorClient) (*types.HuaWeiSubnetListResult, error) {

	cli, err := adaptor.HuaWei(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	var nextMarker *string
	list := new(types.HuaWeiSubnetListResult)
	for {
		opt := &types.HuaWeiSubnetListOption{
			Region: req.Region,
			Page: &adcore.HuaWeiPage{
				Limit: converter.ValToPtr(int32(adcore.HuaWeiQueryLimit)),
			},
			CloudVpcID: req.CloudVpcID,
		}

		// 分页查询的起始资源ID，表示从指定资源的下一条记录开始查询。
		if nextMarker != nil {
			opt.Page.Marker = nextMarker
		}

		tmpList, tmpErr := cli.ListSubnet(kt, opt)
		if tmpErr != nil {
			logs.Errorf("%s-subnet batch get cloud api failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, tmpErr)
			return nil, tmpErr
		}

		list.Details = append(list.Details, tmpList.Details...)
		nextMarker = converter.ValToPtr(tmpList.Details[len(tmpList.Details)-1].CloudID)

		if len(tmpList.Details) < adcore.HuaWeiQueryLimit {
			break
		}
	}

	return list, nil
}

// BatchSyncHuaWeiSubnetList batch sync vendor subnet list.
func BatchSyncHuaWeiSubnetList(kt *kit.Kit, req *SyncHuaWeiOption, list *types.HuaWeiSubnetListResult,
	resourceDBMap map[string]cloudcore.Subnet[cloudcore.HuaWeiSubnetExtension], dataCli *dataclient.Client,
	adaptor *cloudclient.CloudAdaptorClient) error {

	createResources, updateResources, delCloudIDs, err := filterHuaWeiSubnetList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.SubnetBatchUpdateReq[cloud.HuaWeiSubnetUpdateExt]{
			Subnets: updateResources,
		}
		if err = dataCli.HuaWei.Subnet.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-subnet batch compare db update failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		err = batchCreateHuaWeiSubnet(kt, createResources, dataCli, adaptor, req)
		if err != nil {
			logs.Errorf("%s-subnet batch compare db create failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, err)
			return err
		}
	}

	// delete resource data
	if len(delCloudIDs) > 0 {
		if err = BatchDeleteSubnetByIDs(kt, delCloudIDs, dataCli); err != nil {
			logs.Errorf("%s-subnet batch compare db delete failed. accountID: %s, region: %s, delIDs: %v, "+
				"err: %v", enumor.HuaWei, req.AccountID, req.Region, delCloudIDs, err)
			return err
		}
	}

	return nil
}

// filterHuaWeiSubnetList filter huawei subnet list
func filterHuaWeiSubnetList(req *SyncHuaWeiOption, list *types.HuaWeiSubnetListResult,
	resourceDBMap map[string]cloudcore.Subnet[cloudcore.HuaWeiSubnetExtension]) (
	[]cloud.SubnetCreateReq[cloud.HuaWeiSubnetCreateExt],
	[]cloud.SubnetUpdateReq[cloud.HuaWeiSubnetUpdateExt], []string, error) {
	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi subnetlist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	createResources := make([]cloud.SubnetCreateReq[cloud.HuaWeiSubnetCreateExt], 0)
	updateResources := make([]cloud.SubnetUpdateReq[cloud.HuaWeiSubnetUpdateExt], 0)
	for _, item := range list.Details {
		// need compare and update subnet data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if isHuaWeiSubnetChange(resourceInfo, item) {
				tmpRes := cloud.SubnetUpdateReq[cloud.HuaWeiSubnetUpdateExt]{
					ID: resourceInfo.ID,
					SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
						Name:              converter.ValToPtr(item.Name),
						Ipv4Cidr:          item.Ipv4Cidr,
						Ipv6Cidr:          item.Ipv6Cidr,
						Memo:              item.Memo,
						CloudRouteTableID: nil,
						RouteTableID:      nil,
					},
					Extension: &cloud.HuaWeiSubnetUpdateExt{
						Status:       item.Extension.Status,
						DhcpEnable:   converter.ValToPtr(item.Extension.DhcpEnable),
						GatewayIp:    item.Extension.GatewayIp,
						DnsList:      item.Extension.DnsList,
						NtpAddresses: item.Extension.NtpAddresses,
					},
				}

				updateResources = append(updateResources, tmpRes)
			}

			delete(resourceDBMap, item.CloudID)
		} else {
			// need add subnet data
			tmpRes := cloud.SubnetCreateReq[cloud.HuaWeiSubnetCreateExt]{
				AccountID:         req.AccountID,
				CloudVpcID:        item.CloudVpcID,
				VpcID:             "",
				BkBizID:           constant.UnassignedBiz,
				CloudRouteTableID: "",
				RouteTableID:      "",
				CloudID:           item.CloudID,
				Name:              converter.ValToPtr(item.Name),
				Region:            item.Extension.Region,
				Zone:              "",
				Ipv4Cidr:          item.Ipv4Cidr,
				Ipv6Cidr:          item.Ipv6Cidr,
				Memo:              item.Memo,
				Extension: &cloud.HuaWeiSubnetCreateExt{
					Status:       item.Extension.Status,
					DhcpEnable:   item.Extension.DhcpEnable,
					GatewayIp:    item.Extension.GatewayIp,
					DnsList:      item.Extension.DnsList,
					NtpAddresses: item.Extension.NtpAddresses,
				},
			}

			createResources = append(createResources, tmpRes)
		}
	}

	deleteCloudIDs := make([]string, 0, len(resourceDBMap))
	for _, vpc := range resourceDBMap {
		deleteCloudIDs = append(deleteCloudIDs, vpc.CloudID)
	}

	return createResources, updateResources, deleteCloudIDs, nil
}

func isHuaWeiSubnetChange(info cloudcore.Subnet[cloudcore.HuaWeiSubnetExtension], item types.HuaWeiSubnet) bool {
	if info.CloudVpcID != item.CloudVpcID {
		return true
	}

	if info.Name != item.Name {
		return true
	}

	if !assert.IsStringSliceEqual(info.Ipv4Cidr, item.Ipv4Cidr) {
		return true
	}

	if !assert.IsStringSliceEqual(info.Ipv6Cidr, item.Ipv6Cidr) {
		return true
	}

	if !assert.IsPtrStringEqual(item.Memo, info.Memo) {
		return true
	}

	if info.Extension.Status != item.Extension.Status {
		return true
	}

	if info.Extension.DhcpEnable != item.Extension.DhcpEnable {
		return true
	}

	if info.Extension.GatewayIp != item.Extension.GatewayIp {
		return true
	}

	if !assert.IsStringSliceEqual(info.Extension.DnsList, item.Extension.DnsList) {
		return true
	}

	if !assert.IsStringSliceEqual(info.Extension.NtpAddresses, item.Extension.NtpAddresses) {
		return true
	}

	return false
}

func batchCreateHuaWeiSubnet(kt *kit.Kit, createResources []cloud.SubnetCreateReq[cloud.HuaWeiSubnetCreateExt],
	dataCli *dataclient.Client, adaptor *cloudclient.CloudAdaptorClient, req *SyncHuaWeiOption) error {

	opt := &logics.QueryVpcIDsAndSyncOption{
		Vendor:      enumor.HuaWei,
		AccountID:   req.AccountID,
		CloudVpcIDs: []string{req.CloudVpcID},
		Region:      req.Region,
	}
	vpcMap, err := logics.QueryVpcIDsAndSync(kt, adaptor, dataCli, opt)
	if err != nil {
		logs.Errorf("query vpcIDs and sync failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	for index, resource := range createResources {
		one, exist := vpcMap[resource.CloudVpcID]
		if !exist {
			return fmt.Errorf("vpc: %s not sync from cloud", resource.CloudVpcID)
		}

		createResources[index].VpcID = one
	}

	createReq := &cloud.SubnetBatchCreateReq[cloud.HuaWeiSubnetCreateExt]{
		Subnets: createResources,
	}
	if _, err := dataCli.HuaWei.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
		return err
	}

	return nil
}

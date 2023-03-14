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
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// SyncTCloudOption define tcloud sync option.
type SyncTCloudOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"required"`
}

// Validate SyncTCloudOption.
func (opt SyncTCloudOption) Validate() error {
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

// TCloudSubnetSync sync tencent cloud subnet.
func TCloudSubnetSync(kt *kit.Kit, req *SyncTCloudOption, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client) (interface{}, error) {

	// batch get subnet list from cloudapi.
	list, err := BatchGetTCloudSubnetList(kt, req, adaptor)
	if err != nil {
		logs.Errorf("%s-subnet request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.TCloud, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch get subnet map from db.
	resourceDBMap, err := listTcloudSubnetMapFromDB(kt, req.CloudIDs, dataCli)
	if err != nil {
		logs.Errorf("%s-subnet batch get subnetdblist failed. accountID: %s, region: %s, err: %v",
			enumor.TCloud, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch sync vendor subnet list.
	err = BatchSyncTcloudSubnetList(kt, req, list, resourceDBMap, dataCli, adaptor)
	if err != nil {
		logs.Errorf("%s-subnet compare api and subnetdblist failed. accountID: %s, region: %s, err: %v",
			enumor.TCloud, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

func listTcloudSubnetMapFromDB(kt *kit.Kit, cloudIDs []string, dataCli *dataclient.Client) (
	map[string]cloudcore.Subnet[cloudcore.TCloudSubnetExtension], error) {

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
	resourceMap := make(map[string]cloudcore.Subnet[cloudcore.TCloudSubnetExtension], 0)
	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   core.DefaultBasePage,
	}
	dbList, err := dataCli.TCloud.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("tcloud-subnet list ext db failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// BatchGetTCloudSubnetList batch get subnet list from cloudapi.
func BatchGetTCloudSubnetList(kt *kit.Kit, req *SyncTCloudOption,
	adaptor *cloudclient.CloudAdaptorClient) (*types.TCloudSubnetListResult, error) {

	cli, err := adaptor.TCloud(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	page := uint64(0)
	list := new(types.TCloudSubnetListResult)
	for {
		count := uint64(adcore.TCloudQueryLimit)
		offset := page * count
		opt := &adcore.TCloudListOption{
			Region: req.Region,
		}

		// 查询指定CloudIDs
		if len(req.CloudIDs) > 0 {
			opt.CloudIDs = req.CloudIDs
		} else {
			opt.Page = &adcore.TCloudPage{
				Offset: offset,
				Limit:  count,
			}
		}

		tmpList, tmpErr := cli.ListSubnet(kt, opt)
		if tmpErr != nil {
			logs.Errorf("%s-subnet batch get cloudapi failed. accountID: %s, region: %s, offset: %d, "+
				"count: %d, err: %v", enumor.TCloud, req.AccountID, req.Region, offset, count, tmpErr)
			return nil, tmpErr
		}

		list.Details = append(list.Details, tmpList.Details...)

		if len(req.CloudIDs) > 0 || len(tmpList.Details) < int(count) {
			break
		}

		page++
	}

	return list, nil
}

// BatchSyncTcloudSubnetList batch sync vendor subnet list.
func BatchSyncTcloudSubnetList(kt *kit.Kit, req *SyncTCloudOption, list *types.TCloudSubnetListResult,
	resourceDBMap map[string]cloudcore.Subnet[cloudcore.TCloudSubnetExtension], dataCli *dataclient.Client,
	adaptor *cloudclient.CloudAdaptorClient) error {

	createResources, updateResources, delCloudIDs, err := filterTcloudSubnetList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.SubnetBatchUpdateReq[cloud.TCloudSubnetUpdateExt]{
			Subnets: updateResources,
		}
		if err = dataCli.TCloud.Subnet.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-subnet batch compare db update failed. accountID: %s, region: %s, err: %v",
				enumor.TCloud, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		_, err = BatchCreateTcloudSubnet(kt, createResources, dataCli, adaptor, req)
		if err != nil {
			logs.Errorf("%s-subnet batch compare db create failed. accountID: %s, region: %s, err: %v",
				enumor.TCloud, req.AccountID, req.Region, err)
			return err
		}
	}

	// delete resource data
	if len(delCloudIDs) > 0 {
		if err = BatchDeleteSubnetByIDs(kt, delCloudIDs, dataCli); err != nil {
			logs.Errorf("%s-subnet batch compare db delete failed. accountID: %s, region: %s, delIDs: %v, "+
				"err: %v", enumor.TCloud, req.AccountID, req.Region, delCloudIDs, err)
			return err
		}
	}

	return nil
}

// filterTcloudSubnetList filter tcloud subnet list
func filterTcloudSubnetList(req *SyncTCloudOption, list *types.TCloudSubnetListResult,
	resourceDBMap map[string]cloudcore.Subnet[cloudcore.TCloudSubnetExtension]) (
	[]cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt],
	[]cloud.SubnetUpdateReq[cloud.TCloudSubnetUpdateExt], []string, error) {

	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi subnetlist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	createResources := make([]cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt], 0)
	updateResources := make([]cloud.SubnetUpdateReq[cloud.TCloudSubnetUpdateExt], 0)
	for _, item := range list.Details {
		// need compare and update resource data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if isTCloudSubnetChange(resourceInfo, item) {
				tmpRes := cloud.SubnetUpdateReq[cloud.TCloudSubnetUpdateExt]{
					ID: resourceInfo.ID,
					SubnetUpdateBaseInfo: cloud.SubnetUpdateBaseInfo{
						Name:              converter.ValToPtr(item.Name),
						Ipv4Cidr:          item.Ipv4Cidr,
						Ipv6Cidr:          item.Ipv6Cidr,
						Memo:              item.Memo,
						CloudRouteTableID: item.Extension.CloudRouteTableID,
						RouteTableID:      nil,
					},
					Extension: &cloud.TCloudSubnetUpdateExt{
						IsDefault:         item.Extension.IsDefault,
						Region:            item.Extension.Region,
						Zone:              item.Extension.Zone,
						CloudNetworkAclID: item.Extension.CloudNetworkAclID,
					},
				}

				updateResources = append(updateResources, tmpRes)
			}

			delete(resourceDBMap, item.CloudID)
		} else {
			// need add resource data
			tmpRes := cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt]{
				AccountID:  req.AccountID,
				CloudVpcID: item.CloudVpcID,
				VpcID:      "",
				BkBizID:    constant.UnbindBkCloudID,
				// 该字段不支持
				CloudRouteTableID: converter.PtrToVal(item.Extension.CloudRouteTableID),
				RouteTableID:      "",
				CloudID:           item.CloudID,
				Name:              converter.ValToPtr(item.Name),
				Region:            item.Extension.Region,
				Zone:              item.Extension.Zone,
				Ipv4Cidr:          item.Ipv4Cidr,
				Ipv6Cidr:          item.Ipv6Cidr,
				Memo:              item.Memo,
				Extension: &cloud.TCloudSubnetCreateExt{
					IsDefault:         item.Extension.IsDefault,
					CloudNetworkAclID: item.Extension.CloudNetworkAclID,
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

func isTCloudSubnetChange(info cloudcore.Subnet[cloudcore.TCloudSubnetExtension], item types.TCloudSubnet) bool {
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

	if info.Extension.IsDefault != item.Extension.IsDefault {
		return true
	}

	if info.CloudRouteTableID != converter.PtrToVal(item.Extension.CloudRouteTableID) {
		return true
	}

	if info.Extension.CloudNetworkAclId != item.Extension.CloudNetworkAclID {
		return true
	}

	return false
}

func BatchCreateTcloudSubnet(kt *kit.Kit, createResources []cloud.SubnetCreateReq[cloud.TCloudSubnetCreateExt],
	dataCli *dataclient.Client, adaptor *cloudclient.CloudAdaptorClient, req *SyncTCloudOption) (
	*core.BatchCreateResult, error) {

	cloudVpcIDs := make([]string, 0, len(createResources))
	for _, one := range createResources {
		cloudVpcIDs = append(cloudVpcIDs, one.CloudVpcID)
	}

	opt := &logics.QueryVpcIDsAndSyncOption{
		Vendor:      enumor.TCloud,
		AccountID:   req.AccountID,
		CloudVpcIDs: cloudVpcIDs,
		Region:      req.Region,
	}
	vpcMap, err := logics.QueryVpcIDsAndSync(kt, adaptor, dataCli, opt)
	if err != nil {
		logs.Errorf("query vpcIDs and sync failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for index, resource := range createResources {
		one, exist := vpcMap[resource.CloudVpcID]
		if !exist {
			return nil, fmt.Errorf("vpc: %s not sync from cloud", resource.CloudVpcID)
		}

		createResources[index].VpcID = one
	}

	createReq := &cloud.SubnetBatchCreateReq[cloud.TCloudSubnetCreateExt]{
		Subnets: createResources,
	}

	res, err := dataCli.TCloud.Subnet.BatchCreate(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// BatchDeleteSubnetByIDs batch delete subnet ids
func BatchDeleteSubnetByIDs(kt *kit.Kit, deleteIDs []string, dataCli *dataclient.Client) error {
	querySize := int(filter.DefaultMaxInLimit)
	times := len(deleteIDs) / querySize
	if len(deleteIDs)%querySize != 0 {
		times++
	}

	for i := 0; i < times; i++ {
		var newDeleteIDs []string
		if i == times-1 {
			newDeleteIDs = append(newDeleteIDs, deleteIDs[i*querySize:]...)
		} else {
			newDeleteIDs = append(newDeleteIDs, deleteIDs[i*querySize:(i+1)*querySize]...)
		}

		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("id", newDeleteIDs),
		}
		if err := dataCli.Global.Subnet.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
			return err
		}
	}

	return nil
}

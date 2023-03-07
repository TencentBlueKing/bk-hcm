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
	"errors"
	"fmt"

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

// TCloudVpcSync sync tencent cloud vpc.
func TCloudVpcSync(kt *kit.Kit, opt *SyncTCloudOption,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	// batch get vpc list from cloudapi.
	list, err := BatchGetTCloudVpcList(kt, opt, adaptor)
	if err != nil {
		logs.Errorf("%s-vpc request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.TCloud, opt.AccountID, opt.Region, err)
		return nil, err
	}

	// batch get vpc map from db.
	resourceDBMap, err := listTcloudVpcMapFromDB(kt, dataCli, &BatchGetVpcMapOption{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		CloudIDs:  opt.CloudIDs,
	})
	if err != nil {
		logs.Errorf("%s-vpc batch get vpcdblist failed. accountID: %s, region: %s, err: %v",
			enumor.TCloud, opt.AccountID, opt.Region, err)
		return nil, err
	}

	// batch sync vendor vpc list.
	err = BatchSyncTcloudVpcList(kt, opt, list, resourceDBMap, dataCli)
	if err != nil {
		logs.Errorf("%s-vpc compare api and dblist failed. accountID: %s, region: %s, err: %v",
			enumor.TCloud, opt.AccountID, opt.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

func listTcloudVpcMapFromDB(kt *kit.Kit, dataCli *dataclient.Client, opt *BatchGetVpcMapOption) (
	map[string]cloudcore.Vpc[cloudcore.TCloudVpcExtension], error) {

	resourceMap := make(map[string]cloudcore.Vpc[cloudcore.TCloudVpcExtension], 0)
	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "account_id",
				Op:    filter.Equal.Factory(),
				Value: opt.AccountID,
			},
		},
	}

	if len(opt.Region) != 0 {
		expr.Rules = append(expr.Rules, &filter.AtomRule{
			Field: "region",
			Op:    filter.Equal.Factory(),
			Value: opt.Region,
		})
	}

	if len(opt.CloudIDs) != 0 {
		expr.Rules = append(expr.Rules, &filter.AtomRule{
			Field: "cloud_id",
			Op:    filter.In.Factory(),
			Value: opt.CloudIDs,
		})
	}

	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   core.DefaultBasePage,
	}
	dbList, err := dataCli.TCloud.Vpc.ListVpcExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("list tcloud vpc from db failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// BatchGetTCloudVpcList batch get vpc list from cloudapi.
func BatchGetTCloudVpcList(kt *kit.Kit, req *SyncTCloudOption, adaptor *cloudclient.CloudAdaptorClient) (
	*types.TCloudVpcListResult, error) {
	cli, err := adaptor.TCloud(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	page := uint64(0)
	list := new(types.TCloudVpcListResult)
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

		tmpList, tmpErr := cli.ListVpc(kt, opt)
		if tmpErr != nil {
			logs.Errorf("%s-vpc batch get cloudapi failed. accountID: %s, region: %s, offset: %d, count: %d, "+
				"err: %v", enumor.TCloud, req.AccountID, req.Region, offset, count, tmpErr)
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

// BatchGetVpcMapOption ...
type BatchGetVpcMapOption struct {
	AccountID         string   `json:"account_id" validate:"required"`
	Region            string   `json:"region" validate:"omitempty"`
	ResourceGroupName string   `json:"resource_group_name" validate:"omitempty"`
	CloudIDs          []string `json:"cloud_ids" validate:"omitempty"`
}

// BatchSyncTcloudVpcList batch sync vendor vpc list.
func BatchSyncTcloudVpcList(kt *kit.Kit, req *SyncTCloudOption, list *types.TCloudVpcListResult,
	resourceDBMap map[string]cloudcore.Vpc[cloudcore.TCloudVpcExtension], dataCli *dataclient.Client) error {

	createResources, updateResources, delCloudIDs, err := filterTcloudVpcList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.VpcBatchUpdateReq[cloud.TCloudVpcUpdateExt]{
			Vpcs: updateResources,
		}
		if err = dataCli.TCloud.Vpc.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-vpc batch compare db update failed. accountID: %s, region: %s, err: %v",
				enumor.TCloud, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &cloud.VpcBatchCreateReq[cloud.TCloudVpcCreateExt]{
			Vpcs: createResources,
		}
		if _, err = dataCli.TCloud.Vpc.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("%s-vpc batch compare db create failed. accountID: %s, region: %s, err: %v",
				enumor.TCloud, req.AccountID, req.Region, err)
			return err
		}
	}

	// delete resource data
	if len(delCloudIDs) > 0 {
		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
		}
		if err := dataCli.Global.Vpc.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
			logs.Errorf("%s-vpc batch compare db delete failed. accountID: %s, region: %s, delIDs: %v, err: %v",
				enumor.TCloud, req.AccountID, req.Region, delCloudIDs, err)
			return err
		}
	}

	return nil
}

// filterTcloudVpcList filter tcloud vpc list
func filterTcloudVpcList(req *SyncTCloudOption, list *types.TCloudVpcListResult,
	resourceDBMap map[string]cloudcore.Vpc[cloudcore.TCloudVpcExtension]) (
	[]cloud.VpcCreateReq[cloud.TCloudVpcCreateExt], []cloud.VpcUpdateReq[cloud.TCloudVpcUpdateExt], []string, error) {
	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpclist is empty, accountID: %s, region: %s", req.AccountID, req.Region)
	}

	createResources := make([]cloud.VpcCreateReq[cloud.TCloudVpcCreateExt], 0)
	updateResources := make([]cloud.VpcUpdateReq[cloud.TCloudVpcUpdateExt], 0)
	for _, item := range list.Details {
		// need compare and update resource data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if isTCloudVpcChange(resourceInfo, item) {
				tmpRes := cloud.VpcUpdateReq[cloud.TCloudVpcUpdateExt]{
					ID: resourceInfo.ID,
					VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
						Name: converter.ValToPtr(item.Name),
						Memo: item.Memo,
					},
					Extension: &cloud.TCloudVpcUpdateExt{
						IsDefault:       converter.ValToPtr(item.Extension.IsDefault),
						EnableMulticast: converter.ValToPtr(item.Extension.EnableMulticast),
						DnsServerSet:    item.Extension.DnsServerSet,
						DomainName:      converter.ValToPtr(item.Extension.DomainName),
					},
				}

				if item.Extension.Cidr != nil {
					tmpCidrs := make([]cloud.TCloudCidr, 0, len(item.Extension.Cidr))
					for _, cidrItem := range item.Extension.Cidr {
						tmpCidrs = append(tmpCidrs, cloud.TCloudCidr{
							Type:     cidrItem.Type,
							Cidr:     cidrItem.Cidr,
							Category: cidrItem.Category,
						})
					}
					tmpRes.Extension.Cidr = tmpCidrs
				}

				updateResources = append(updateResources, tmpRes)
			}

			delete(resourceDBMap, item.CloudID)
		} else {
			// need add resource data
			tmpRes := cloud.VpcCreateReq[cloud.TCloudVpcCreateExt]{
				AccountID: req.AccountID,
				CloudID:   item.CloudID,
				Name:      converter.ValToPtr(item.Name),
				BkBizID:   constant.UnassignedBiz,
				BkCloudID: constant.UnbindBkCloudID,
				Region:    item.Region,
				Category:  enumor.BizVpcCategory,
				Memo:      item.Memo,
				Extension: &cloud.TCloudVpcCreateExt{
					IsDefault:       item.Extension.IsDefault,
					EnableMulticast: item.Extension.EnableMulticast,
					DnsServerSet:    item.Extension.DnsServerSet,
					DomainName:      item.Extension.DomainName,
				},
			}

			if item.Extension.Cidr != nil {
				tmpCidrs := make([]cloud.TCloudCidr, 0, len(item.Extension.Cidr))
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.TCloudCidr{
						Type:     cidrItem.Type,
						Cidr:     cidrItem.Cidr,
						Category: cidrItem.Category,
					})
				}
				tmpRes.Extension.Cidr = tmpCidrs
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

func isTCloudVpcChange(info cloudcore.Vpc[cloudcore.TCloudVpcExtension], item types.TCloudVpc) bool {
	if info.Name != item.Name {
		return true
	}

	if info.Region != item.Region {
		return true
	}

	if !assert.IsPtrStringEqual(info.Memo, item.Memo) {
		return true
	}

	for _, db := range info.Extension.Cidr {
		for _, cloud := range item.Extension.Cidr {
			if db.Cidr != cloud.Cidr {
				return true
			}

			if db.Category != cloud.Category {
				return true
			}

			if db.Type != cloud.Type {
				return true
			}
		}
	}

	if info.Extension.IsDefault != item.Extension.IsDefault {
		return true
	}

	if info.Extension.EnableMulticast != item.Extension.EnableMulticast {
		return true
	}

	if !assert.IsStringSliceEqual(info.Extension.DnsServerSet, item.Extension.DnsServerSet) {
		return true
	}

	if info.Extension.DomainName != item.Extension.DomainName {
		return true
	}

	return false
}

// BatchDeleteVpcByIDs batch delete vpc ids
func BatchDeleteVpcByIDs(kt *kit.Kit, deleteCloudIDs []string, dataCli *dataclient.Client) error {
	querySize := int(filter.DefaultMaxInLimit)
	times := len(deleteCloudIDs) / querySize
	if len(deleteCloudIDs)%querySize != 0 {
		times++
	}

	for i := 0; i < times; i++ {
		var newDeleteIDs []string
		if i == times-1 {
			newDeleteIDs = append(newDeleteIDs, deleteCloudIDs[i*querySize:]...)
		} else {
			newDeleteIDs = append(newDeleteIDs, deleteCloudIDs[i*querySize:(i+1)*querySize]...)
		}

	}

	return nil
}

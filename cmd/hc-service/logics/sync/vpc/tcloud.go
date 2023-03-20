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
	"hcm/pkg/tools/slice"
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
func TCloudVpcSync(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client,
	opt *SyncTCloudOption) (interface{}, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	list, err := listTCloudVpcFromCloud(kt, opt, adaptor)
	if err != nil {
		logs.Errorf("list tcloud vpc from cloud failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	resourceDBMap, err := listTCloudVpcMapFromDB(kt, dataCli, opt)
	if err != nil {
		logs.Errorf("list tcloud vpc from db failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	if len(list.Details) == 0 && len(resourceDBMap) == 0 {
		return nil, nil
	}

	createResources, updateResources, delCloudIDs, err := diffTCloudVpc(opt, list, resourceDBMap)
	if err != nil {
		logs.Errorf("diff tcloud vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(updateResources) > 0 {
		updateReq := &cloud.VpcBatchUpdateReq[cloud.TCloudVpcUpdateExt]{
			Vpcs: updateResources,
		}
		if err = dataCli.TCloud.Vpc.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("batch update db vpc failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(createResources) > 0 {
		createReq := &cloud.VpcBatchCreateReq[cloud.TCloudVpcCreateExt]{
			Vpcs: createResources,
		}
		if _, err = dataCli.TCloud.Vpc.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("batch create db vpc failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(delCloudIDs) > 0 {
		delListOpt := &SyncTCloudOption{
			AccountID: opt.AccountID,
			Region:    opt.Region,
			CloudIDs:  delCloudIDs,
		}
		delResult, err := listTCloudVpcFromCloud(kt, delListOpt, adaptor)
		if err != nil {
			logs.Errorf("list tcloud vpc failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
			return nil, err
		}

		if len(delResult.Details) > 0 {
			logs.Errorf("validate vpc not exist failed, before delete, opt: %v, rid: %s", opt, kt.Rid)
			return nil, fmt.Errorf("validate vpc not exist failed, before delete")
		}

		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
		}
		if err = dataCli.Global.Vpc.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
			logs.Errorf("batch delete db vpc failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func listTCloudVpcMapFromDB(kt *kit.Kit, dataCli *dataclient.Client, opt *SyncTCloudOption) (
	map[string]cloudcore.Vpc[cloudcore.TCloudVpcExtension], error) {

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "account_id",
				Op:    filter.Equal.Factory(),
				Value: opt.AccountID,
			},
			&filter.AtomRule{
				Field: "cloud_id",
				Op:    filter.In.Factory(),
				Value: opt.CloudIDs,
			},
			&filter.AtomRule{
				Field: "region",
				Op:    filter.Equal.Factory(),
				Value: opt.Region,
			},
		},
	}

	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   core.DefaultBasePage,
	}
	dbList, err := dataCli.TCloud.Vpc.ListVpcExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		return nil, err
	}

	resourceMap := make(map[string]cloudcore.Vpc[cloudcore.TCloudVpcExtension], len(dbList.Details))
	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

func listTCloudVpcFromCloud(kt *kit.Kit, req *SyncTCloudOption, adaptor *cloudclient.CloudAdaptorClient) (
	*types.TCloudVpcListResult, error) {

	cli, err := adaptor.TCloud(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	list := &types.TCloudVpcListResult{
		Details: make([]types.TCloudVpc, 0),
	}
	elems := slice.Split(req.CloudIDs, adcore.TCloudQueryLimit)
	for _, partIDs := range elems {
		opt := &adcore.TCloudListOption{
			Region:   req.Region,
			CloudIDs: partIDs,
		}

		tmpList, err := cli.ListVpc(kt, opt)
		if err != nil {
			return nil, err
		}

		list.Details = append(list.Details, tmpList.Details...)
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

// diffTCloudVpc filter tcloud vpc list
func diffTCloudVpc(req *SyncTCloudOption, list *types.TCloudVpcListResult,
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

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
)

// SyncHuaWeiOption define huawei sync option.
type SyncHuaWeiOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"required"`
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

// HuaWeiVpcSync sync huawei vpc
func HuaWeiVpcSync(kt *kit.Kit, opt *SyncHuaWeiOption,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	list, err := listHuaWeiVpcFromCloud(kt, opt, adaptor)
	if err != nil {
		logs.Errorf("list huawei vpc from cloud failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	resourceDBMap, err := listHuaWeiVpcMapFromDB(kt, dataCli, opt)
	if err != nil {
		logs.Errorf("list huawei vpc from db failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	if len(list.Details) == 0 && len(resourceDBMap) == 0 {
		return nil, nil
	}

	createResources, updateResources, delCloudIDs, err := diffHuaWeiVpc(list, resourceDBMap, opt)
	if err != nil {
		logs.Errorf("diff huawei vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.VpcBatchUpdateReq[cloud.HuaWeiVpcUpdateExt]{
			Vpcs: updateResources,
		}
		if err = dataCli.HuaWei.Vpc.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("batch update vpc from db failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &cloud.VpcBatchCreateReq[cloud.HuaWeiVpcCreateExt]{
			Vpcs: createResources,
		}
		if _, err = dataCli.HuaWei.Vpc.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("batch create vpc to db failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	// delete resource data
	if len(delCloudIDs) > 0 {
		delListOpt := &SyncHuaWeiOption{
			AccountID: opt.AccountID,
			Region:    opt.Region,
			CloudIDs:  delCloudIDs,
		}
		delResult, err := listHuaWeiVpcFromCloud(kt, delListOpt, adaptor)
		if err != nil {
			logs.Errorf("list huawei vpc failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
			return nil, err
		}

		if len(delResult.Details) > 0 {
			logs.Errorf("validate vpc not exist failed, before delete, opt: %v, rid: %s", opt, kt.Rid)
			return nil, fmt.Errorf("validate vpc not exist failed, before delete")
		}

		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
		}
		if err := dataCli.Global.Vpc.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
			logs.Errorf("batch delete vpc from db failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func listHuaWeiVpcMapFromDB(kt *kit.Kit, dataCli *dataclient.Client, opt *SyncHuaWeiOption) (
	map[string]cloudcore.Vpc[cloudcore.HuaWeiVpcExtension], error) {

	dbQueryReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: opt.AccountID,
				}, &filter.AtomRule{
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
		},
		Page: core.DefaultBasePage,
	}
	dbList, err := dataCli.HuaWei.Vpc.ListVpcExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("list vpc from db failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resourceMap := make(map[string]cloudcore.Vpc[cloudcore.HuaWeiVpcExtension], len(dbList.Details))
	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// listHuaWeiVpcFromCloud batch get vpc list from cloudapi.
func listHuaWeiVpcFromCloud(kt *kit.Kit, req *SyncHuaWeiOption, adaptor *cloudclient.CloudAdaptorClient) (
	*types.HuaWeiVpcListResult, error) {

	cli, err := adaptor.HuaWei(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.HuaWeiVpcListOption{
		HuaWeiListOption: adcore.HuaWeiListOption{
			Region:   req.Region,
			CloudIDs: req.CloudIDs,
		},
	}

	result, err := cli.ListVpc(kt, opt)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// diffHuaWeiVpc filter huawei vpc list
func diffHuaWeiVpc(list *types.HuaWeiVpcListResult, resourceDBMap map[string]cloudcore.
	Vpc[cloudcore.HuaWeiVpcExtension], opt *SyncHuaWeiOption) ([]cloud.VpcCreateReq[cloud.HuaWeiVpcCreateExt],
	[]cloud.VpcUpdateReq[cloud.HuaWeiVpcUpdateExt], []string, error) {

	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpclist is empty, accountID: %s, region: %s", opt.AccountID, opt.Region)
	}

	createResources := make([]cloud.VpcCreateReq[cloud.HuaWeiVpcCreateExt], 0)
	updateResources := make([]cloud.VpcUpdateReq[cloud.HuaWeiVpcUpdateExt], 0)
	for _, item := range list.Details {
		// need compare and update vpc data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if isHuaWeiVpcChange(resourceInfo, item) {
				tmpRes := cloud.VpcUpdateReq[cloud.HuaWeiVpcUpdateExt]{
					ID: resourceInfo.ID,
					VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
						Name: converter.ValToPtr(item.Name),
						Memo: item.Memo,
					},
					Extension: &cloud.HuaWeiVpcUpdateExt{
						Status:              item.Extension.Status,
						EnterpriseProjectId: converter.ValToPtr(item.Extension.EnterpriseProjectId),
					},
				}

				if item.Extension.Cidr != nil {
					tmpCidrs := make([]cloud.HuaWeiCidr, 0, len(item.Extension.Cidr))
					for _, cidrItem := range item.Extension.Cidr {
						tmpCidrs = append(tmpCidrs, cloud.HuaWeiCidr{
							Type: cidrItem.Type,
							Cidr: cidrItem.Cidr,
						})
					}
					tmpRes.Extension.Cidr = tmpCidrs
				}

				updateResources = append(updateResources, tmpRes)
			}

			delete(resourceDBMap, item.CloudID)
		} else {
			// need add vpc data
			tmpRes := cloud.VpcCreateReq[cloud.HuaWeiVpcCreateExt]{
				AccountID: opt.AccountID,
				CloudID:   item.CloudID,
				BkBizID:   constant.UnassignedBiz,
				BkCloudID: constant.UnbindBkCloudID,
				Name:      converter.ValToPtr(item.Name),
				Region:    item.Region,
				Category:  enumor.BizVpcCategory,
				Memo:      item.Memo,
				Extension: &cloud.HuaWeiVpcCreateExt{
					Status:              item.Extension.Status,
					EnterpriseProjectID: item.Extension.EnterpriseProjectId,
				},
			}

			if item.Extension.Cidr != nil {
				tmpCidrs := make([]cloud.HuaWeiCidr, 0, len(item.Extension.Cidr))
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.HuaWeiCidr{
						Type: cidrItem.Type,
						Cidr: cidrItem.Cidr,
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

func isHuaWeiVpcChange(info cloudcore.Vpc[cloudcore.HuaWeiVpcExtension], item types.HuaWeiVpc) bool {
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

			if db.Type != cloud.Type {
				return true
			}
		}
	}

	if info.Extension.Status != item.Extension.Status {
		return true
	}

	if info.Extension.EnterpriseProjectId != item.Extension.EnterpriseProjectId {
		return true
	}

	return false
}

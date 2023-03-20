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

// SyncGcpOption define gcp sync option.
type SyncGcpOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
	SelfLinks []string `json:"self_links" validate:"omitempty"`
}

// Validate SyncGcpOption.
func (opt SyncGcpOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.SelfLinks) == 0 && len(opt.CloudIDs) == 0 {
		return errors.New("self_links or cloud_ids is required")
	}

	if len(opt.SelfLinks) != 0 && len(opt.CloudIDs) != 0 {
		return errors.New("self_links or cloud_ids only one can be set")
	}

	if len(opt.CloudIDs) > int(core.DefaultMaxPageLimit) {
		return fmt.Errorf("cloudIDs should <= %d", core.DefaultMaxPageLimit)
	}

	if len(opt.SelfLinks) > int(core.DefaultMaxPageLimit) {
		return fmt.Errorf("selfLinks should <= %d", core.DefaultMaxPageLimit)
	}

	return nil
}

// GcpVpcSync sync gcp cloud vpc.
func GcpVpcSync(kt *kit.Kit, opt *SyncGcpOption, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client) (interface{}, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	list, err := listGcpVpcFromCloud(kt, opt, adaptor)
	if err != nil {
		logs.Errorf("list gcp vpc from cloud failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	resourceDBMap, err := listGcpVpcMapFromDB(kt, dataCli, opt)
	if err != nil {
		logs.Errorf("list gcp vpc from db failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	if len(list.Details) == 0 && len(resourceDBMap) == 0 {
		return nil, nil
	}

	createResources, updateResources, delCloudIDs, err := diffGcpVpc(opt, list, resourceDBMap)
	if err != nil {
		logs.Errorf("diff gcp vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(updateResources) > 0 {
		updateReq := &cloud.VpcBatchUpdateReq[cloud.GcpVpcUpdateExt]{
			Vpcs: updateResources,
		}
		if err = dataCli.Gcp.Vpc.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("batch update db vpc failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(createResources) > 0 {
		createReq := &cloud.VpcBatchCreateReq[cloud.GcpVpcCreateExt]{
			Vpcs: createResources,
		}
		if _, err = dataCli.Gcp.Vpc.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("batch create db vpc failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(delCloudIDs) > 0 {
		delListOpt := &SyncGcpOption{
			AccountID: opt.AccountID,
			CloudIDs:  delCloudIDs,
		}
		delResult, err := listGcpVpcFromCloud(kt, delListOpt, adaptor)
		if err != nil {
			logs.Errorf("list gcp vpc failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
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

func listGcpVpcMapFromDB(kt *kit.Kit, dataCli *dataclient.Client, opt *SyncGcpOption) (
	map[string]cloudcore.Vpc[cloudcore.GcpVpcExtension], error) {

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

	if len(opt.CloudIDs) != 0 {
		expr.Rules = append(expr.Rules, &filter.AtomRule{
			Field: "cloud_id",
			Op:    filter.In.Factory(),
			Value: opt.CloudIDs,
		})
	}

	if len(opt.SelfLinks) != 0 {
		expr.Rules = append(expr.Rules, &filter.AtomRule{
			Field: "extension.self_link",
			Op:    filter.JSONIn.Factory(),
			Value: opt.SelfLinks,
		})
	}

	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   core.DefaultBasePage,
	}
	dbList, err := dataCli.Gcp.Vpc.ListVpcExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("list gcp vpc from db failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resourceMap := make(map[string]cloudcore.Vpc[cloudcore.GcpVpcExtension], len(dbList.Details))
	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// listGcpVpcFromCloud batch get vpc list from cloudapi.
func listGcpVpcFromCloud(kt *kit.Kit, req *SyncGcpOption, adaptor *cloudclient.CloudAdaptorClient) (
	*types.GcpVpcListResult, error) {

	cli, err := adaptor.Gcp(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := new(types.GcpListOption)

	if len(req.CloudIDs) > 0 {
		opt.CloudIDs = req.CloudIDs
	}

	if len(req.SelfLinks) > 0 {
		opt.SelfLinks = req.SelfLinks
	}

	result, err := cli.ListVpc(kt, opt)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// diffGcpVpc filter gcp vpc list
func diffGcpVpc(req *SyncGcpOption, list *types.GcpVpcListResult,
	resourceDBMap map[string]cloudcore.Vpc[cloudcore.GcpVpcExtension]) ([]cloud.VpcCreateReq[cloud.GcpVpcCreateExt],
	[]cloud.VpcUpdateReq[cloud.GcpVpcUpdateExt], []string, error) {
	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpclist is empty, accountID: %s", req.AccountID)
	}

	createResources := make([]cloud.VpcCreateReq[cloud.GcpVpcCreateExt], 0)
	updateResources := make([]cloud.VpcUpdateReq[cloud.GcpVpcUpdateExt], 0)
	for _, item := range list.Details {
		// need compare and update vpc data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if isGcpVpcChange(resourceInfo, item) {
				tmpRes := cloud.VpcUpdateReq[cloud.GcpVpcUpdateExt]{
					ID: resourceInfo.ID,
					VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
						Name:      converter.ValToPtr(item.Name),
						Category:  "",
						Memo:      item.Memo,
						BkCloudID: 0,
						BkBizID:   0,
					},
					Extension: &cloud.GcpVpcUpdateExt{
						EnableUlaInternalIpv6: converter.ValToPtr(item.Extension.EnableUlaInternalIpv6),
						InternalIpv6Range:     &item.Extension.InternalIpv6Range,
						Mtu:                   item.Extension.Mtu,
						RoutingMode:           &item.Extension.RoutingMode,
					},
				}

				updateResources = append(updateResources, tmpRes)
			}

			delete(resourceDBMap, item.CloudID)
		} else {
			// need add vpc data
			tmpRes := cloud.VpcCreateReq[cloud.GcpVpcCreateExt]{
				Region:    item.Region,
				AccountID: req.AccountID,
				CloudID:   item.CloudID,
				BkBizID:   constant.UnassignedBiz,
				BkCloudID: constant.UnbindBkCloudID,
				Name:      converter.ValToPtr(item.Name),
				Category:  enumor.BizVpcCategory,
				Memo:      item.Memo,
				Extension: &cloud.GcpVpcCreateExt{
					SelfLink:              item.Extension.SelfLink,
					AutoCreateSubnetworks: item.Extension.AutoCreateSubnetworks,
					EnableUlaInternalIpv6: item.Extension.EnableUlaInternalIpv6,
					InternalIpv6Range:     item.Extension.InternalIpv6Range,
					Mtu:                   item.Extension.Mtu,
					RoutingMode:           item.Extension.RoutingMode,
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

func isGcpVpcChange(info cloudcore.Vpc[cloudcore.GcpVpcExtension], item types.GcpVpc) bool {
	if info.Name != item.Name {
		return true
	}

	if info.Region != item.Region {
		return true
	}

	if !assert.IsPtrStringEqual(info.Memo, item.Memo) {
		return true
	}

	if info.Extension.SelfLink != item.Extension.SelfLink {
		return true
	}

	if info.Extension.AutoCreateSubnetworks != item.Extension.AutoCreateSubnetworks {
		return true
	}

	if info.Extension.EnableUlaInternalIpv6 != item.Extension.EnableUlaInternalIpv6 {
		return true
	}

	if info.Extension.InternalIpv6Range != item.Extension.InternalIpv6Range {
		return true
	}

	if info.Extension.Mtu != item.Extension.Mtu {
		return true
	}

	if info.Extension.RoutingMode != item.Extension.RoutingMode {
		return true
	}

	return false
}

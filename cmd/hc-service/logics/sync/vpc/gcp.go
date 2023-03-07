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
func GcpVpcSync(kt *kit.Kit, opt *SyncGcpOption,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	// batch get vpc list from cloudapi.
	list, err := BatchGetGcpVpcList(kt, opt, adaptor)
	if err != nil {
		logs.Errorf("%s-vpc request cloudapi response failed. accountID: %s, err: %v", enumor.Gcp, opt.AccountID, err)
		return nil, err
	}

	// batch get vpc map from db.
	resourceDBMap, err := listGcpVpcMapFromDB(kt, dataCli, &BatchGetVpcMapOption{
		AccountID: opt.AccountID,
		CloudIDs:  opt.CloudIDs,
	})
	if err != nil {
		logs.Errorf("%s-vpc batch get vpcdblist failed. accountID: %s, err: %v", enumor.Gcp, opt.AccountID, err)
		return nil, err
	}

	// batch sync vendor vpc list.
	err = BatchSyncGcpVpcList(kt, opt, list, resourceDBMap, dataCli)
	if err != nil {
		logs.Errorf("%s-vpc compare api and dblist failed. accountID: %s, err: %v", enumor.Gcp, opt.AccountID, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

func listGcpVpcMapFromDB(kt *kit.Kit, dataCli *dataclient.Client, opt *BatchGetVpcMapOption) (
	map[string]cloudcore.Vpc[cloudcore.GcpVpcExtension], error) {

	resourceMap := make(map[string]cloudcore.Vpc[cloudcore.GcpVpcExtension], 0)
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

	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   core.DefaultBasePage,
	}
	dbList, err := dataCli.Gcp.Vpc.ListVpcExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("list gcp vpc from db failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// BatchGetGcpVpcList batch get vpc list from cloudapi.
func BatchGetGcpVpcList(kt *kit.Kit, req *SyncGcpOption, adaptor *cloudclient.CloudAdaptorClient) (
	*types.GcpVpcListResult, error) {

	cli, err := adaptor.Gcp(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	list := new(types.GcpVpcListResult)
	for {
		opt := new(types.GcpListOption)
		opt.Page = &adcore.GcpPage{
			PageSize: int64(adcore.GcpQueryLimit),
		}

		if nextToken != "" {
			opt.Page.PageToken = nextToken
		}

		// 查询指定CloudIDs
		if len(req.CloudIDs) > 0 {
			opt.Page = nil
			opt.CloudIDs = req.CloudIDs
		} else if len(req.SelfLinks) > 0 {
			opt.Page = nil
			opt.SelfLinks = req.SelfLinks
		}

		tmpList, tmpErr := cli.ListVpc(kt, opt)
		if tmpErr != nil {
			logs.Errorf("%s-vpc batch get cloud api failed. accountID: %s, nextToken: %s, err: %v",
				enumor.Gcp, req.AccountID, nextToken, tmpErr)
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
func BatchSyncGcpVpcList(kt *kit.Kit, req *SyncGcpOption, list *types.GcpVpcListResult,
	resourceDBMap map[string]cloudcore.Vpc[cloudcore.GcpVpcExtension], dataCli *dataclient.Client) error {

	createResources, updateResources, delCloudIDs, err := filterGcpVpcList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.VpcBatchUpdateReq[cloud.GcpVpcUpdateExt]{
			Vpcs: updateResources,
		}
		if err = dataCli.Gcp.Vpc.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-vpc batch compare db update failed. accountID: %s, err: %v",
				enumor.Gcp, req.AccountID, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &cloud.VpcBatchCreateReq[cloud.GcpVpcCreateExt]{
			Vpcs: createResources,
		}
		if _, err = dataCli.Gcp.Vpc.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("%s-vpc batch compare db create failed. accountID: %s, err: %v",
				enumor.Gcp, req.AccountID, err)
			return err
		}
	}

	// delete resource data
	if len(delCloudIDs) > 0 {
		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
		}
		if err := dataCli.Global.Vpc.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
			logs.Errorf("%s-vpc batch compare db delete failed. accountID: %s, delIDs: %v, err: %v",
				enumor.Gcp, req.AccountID, delCloudIDs, err)
			return err
		}
	}

	return nil
}

// filterGcpVpcList filter gcp vpc list
func filterGcpVpcList(req *SyncGcpOption, list *types.GcpVpcListResult,
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
						Mtu:                   item.Extension.Mtu,
						RoutingMode:           item.Extension.RoutingMode,
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

	if info.Extension.Mtu != item.Extension.Mtu {
		return true
	}

	if info.Extension.RoutingMode != item.Extension.RoutingMode {
		return true
	}

	return false
}

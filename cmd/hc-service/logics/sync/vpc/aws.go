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

// SyncAwsOption define aws sync option.
type SyncAwsOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"required"`
}

// Validate SyncAwsOption.
func (opt SyncAwsOption) Validate() error {
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

// AwsVpcSync sync aws cloud vpc.
func AwsVpcSync(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client,
	opt *SyncAwsOption) (interface{}, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	// batch get vpc list from cloudapi.
	list, err := BatchGetAwsVpcList(kt, adaptor, opt)
	if err != nil {
		logs.Errorf("%s-vpc request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.Aws, opt.AccountID, opt.Region, err)
		return nil, err
	}

	// batch get vpc map from db.
	resourceDBMap, err := listAwsVpcMapFromDB(kt, dataCli, &BatchGetVpcMapOption{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		CloudIDs:  opt.CloudIDs,
	})
	if err != nil {
		logs.Errorf("%s-vpc batch get vpcdblist failed. accountID: %s, region: %s, err: %v",
			enumor.Aws, opt.AccountID, opt.Region, err)
		return nil, err
	}

	// batch sync vendor vpc list.
	err = BatchSyncAwsVpcList(kt, list, resourceDBMap, dataCli, opt)
	if err != nil {
		logs.Errorf("%s-vpc compare api and dblist failed. accountID: %s, region: %s, err: %v",
			enumor.Aws, opt.AccountID, opt.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

func listAwsVpcMapFromDB(kt *kit.Kit, dataCli *dataclient.Client, opt *BatchGetVpcMapOption) (
	map[string]cloudcore.Vpc[cloudcore.AwsVpcExtension], error) {

	resourceMap := make(map[string]cloudcore.Vpc[cloudcore.AwsVpcExtension], 0)
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
	dbList, err := dataCli.Aws.Vpc.ListVpcExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("list aws vpc from db failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// BatchGetAwsVpcList batch get vpc list from cloudapi.
func BatchGetAwsVpcList(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient, opt *SyncAwsOption) (
	*types.AwsVpcListResult, error) {

	cli, err := adaptor.Aws(kt, opt.AccountID)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	list := new(types.AwsVpcListResult)
	for {
		listOpt := new(adcore.AwsListOption)
		listOpt.Region = opt.Region

		// 查询指定CloudIDs
		if len(opt.CloudIDs) > 0 {
			listOpt.CloudIDs = opt.CloudIDs
		} else {
			listOpt.Page = &adcore.AwsPage{
				MaxResults: converter.ValToPtr(int64(adcore.AwsQueryLimit)),
			}

			if nextToken != "" {
				listOpt.Page.NextToken = converter.ValToPtr(nextToken)
			}
		}

		tmpList, tmpErr := cli.ListVpc(kt, listOpt)
		if tmpErr != nil {
			logs.Errorf("%s-vpc batch get cloud api failed. accountID: %s, region: %s, nextToken: %s, err: %v",
				enumor.Aws, opt.AccountID, opt.Region, nextToken, tmpErr)
			return nil, tmpErr
		}

		// traversal vpclist supply fields
		for _, item := range tmpList.Details {
			dnsHostnames, dnsSupport, dnsErr := cli.GetVpcAttribute(kt, item.CloudID, item.Region)
			if dnsErr == nil {
				item.Extension.EnableDnsHostnames = dnsHostnames
				item.Extension.EnableDnsSupport = dnsSupport
			}
		}

		if len(tmpList.Details) == 0 {
			break
		}

		list.Details = append(list.Details, tmpList.Details...)
		if len(opt.CloudIDs) > 0 || tmpList.NextToken == nil {
			break
		}
		nextToken = *tmpList.NextToken
	}

	return list, nil
}

// BatchSyncAwsVpcList batch sync vendor vpc list.
func BatchSyncAwsVpcList(kt *kit.Kit, list *types.AwsVpcListResult,
	resourceDBMap map[string]cloudcore.Vpc[cloudcore.AwsVpcExtension],
	dataCli *dataclient.Client, opt *SyncAwsOption) error {

	createResources, updateResources, deleteCloudIDs, err := filterAwsVpcList(list, resourceDBMap, opt)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.VpcBatchUpdateReq[cloud.AwsVpcUpdateExt]{
			Vpcs: updateResources,
		}
		if err = dataCli.Aws.Vpc.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-vpc batch compare db update failed. accountID: %s, region: %s, err: %v",
				enumor.Aws, opt.AccountID, opt.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &cloud.VpcBatchCreateReq[cloud.AwsVpcCreateExt]{
			Vpcs: createResources,
		}
		if _, err = dataCli.Aws.Vpc.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("%s-vpc batch compare db create failed. accountID: %s, region: %s, err: %v",
				enumor.Aws, opt.AccountID, opt.Region, err)
			return err
		}
	}

	// delete resource data
	if len(deleteCloudIDs) > 0 {
		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", deleteCloudIDs),
		}
		if err := dataCli.Global.Vpc.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
			logs.Errorf("%s-vpc batch compare db delete failed. accountID: %s, region: %s, delIDs: %v, err: %v",
				enumor.Aws, opt.AccountID, opt.Region, deleteCloudIDs, err)
			return err
		}
	}

	return nil
}

// filterAwsVpcList filter aws vpc list
func filterAwsVpcList(list *types.AwsVpcListResult, resourceDBMap map[string]cloudcore.Vpc[cloudcore.AwsVpcExtension],
	opt *SyncAwsOption) ([]cloud.VpcCreateReq[cloud.AwsVpcCreateExt],
	[]cloud.VpcUpdateReq[cloud.AwsVpcUpdateExt], []string, error) {
	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpclist is empty, accountID: %s, region: %s", opt.AccountID, opt.Region)
	}

	createResources := make([]cloud.VpcCreateReq[cloud.AwsVpcCreateExt], 0)
	updateResources := make([]cloud.VpcUpdateReq[cloud.AwsVpcUpdateExt], 0)
	for _, item := range list.Details {
		// db存在，判断是否需要更新
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			if isAwsVpcChange(resourceInfo, item) {
				tmpRes := cloud.VpcUpdateReq[cloud.AwsVpcUpdateExt]{
					ID: resourceInfo.ID,
					VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
						Name: converter.ValToPtr(item.Name),
						Memo: item.Memo,
					},
					Extension: &cloud.AwsVpcUpdateExt{
						State:              item.Extension.State,
						InstanceTenancy:    converter.ValToPtr(item.Extension.InstanceTenancy),
						IsDefault:          converter.ValToPtr(item.Extension.IsDefault),
						EnableDnsHostnames: converter.ValToPtr(item.Extension.EnableDnsHostnames),
						EnableDnsSupport:   converter.ValToPtr(item.Extension.EnableDnsSupport),
					},
				}

				if item.Extension.Cidr != nil {
					tmpCidrs := make([]cloud.AwsCidr, 0, len(item.Extension.Cidr))
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
			}

			delete(resourceDBMap, item.CloudID)
		} else {
			// need add vpc data
			tmpRes := cloud.VpcCreateReq[cloud.AwsVpcCreateExt]{
				AccountID: opt.AccountID,
				CloudID:   item.CloudID,
				BkBizID:   constant.UnassignedBiz,
				BkCloudID: constant.UnbindBkCloudID,
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
				tmpCidrs := make([]cloud.AwsCidr, 0, len(item.Extension.Cidr))
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

	deleteCloudIDs := make([]string, 0, len(resourceDBMap))
	for _, vpc := range resourceDBMap {
		deleteCloudIDs = append(deleteCloudIDs, vpc.CloudID)
	}

	return createResources, updateResources, deleteCloudIDs, nil
}

func isAwsVpcChange(info cloudcore.Vpc[cloudcore.AwsVpcExtension], item types.AwsVpc) bool {
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

			if db.AddressPool != cloud.AddressPool {
				return true
			}

			if db.Type != cloud.Type {
				return true
			}

			if db.State != cloud.State {
				return true
			}

		}
	}

	if info.Extension.IsDefault != item.Extension.IsDefault {
		return true
	}

	if info.Extension.InstanceTenancy != item.Extension.InstanceTenancy {
		return true
	}

	if info.Extension.EnableDnsHostnames != item.Extension.EnableDnsHostnames {
		return true
	}

	if info.Extension.EnableDnsSupport != item.Extension.EnableDnsSupport {
		return true
	}

	return false
}

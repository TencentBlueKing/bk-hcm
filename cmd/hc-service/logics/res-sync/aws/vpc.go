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

package aws

import (
	"fmt"
	"strings"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/adaptor/aws"
	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
)

// SyncVpcOption ...
type SyncVpcOption struct {
}

// Validate ...
func (opt SyncVpcOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Vpc ...
func (cli *client) Vpc(kt *kit.Kit, params *SyncBaseParams, opt *SyncVpcOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vpcFromCloud, err := cli.listVpcFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	vpcFromDB, err := cli.listVpcFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(vpcFromCloud) == 0 && len(vpcFromDB) == 0 {
		return new(SyncResult), nil
	}

	addVpc, updateMap, delCloudIDs := common.Diff[types.AwsVpc, cloudcore.Vpc[cloudcore.AwsVpcExtension]](
		vpcFromCloud, vpcFromDB, isAwsVpcChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteVpc(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addVpc) > 0 {
		if err = cli.createVpc(kt, params.AccountID, addVpc); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateVpc(kt, params.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (cli *client) deleteVpc(kt *kit.Kit, accountID string, region string, delCloudIDs []string) error {
	if len(delCloudIDs) == 0 {
		return fmt.Errorf("delete vpc, cloudIDs is required")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  delCloudIDs,
	}
	delVpcFromCloud, err := cli.listVpcFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delVpcFromCloud) > 0 {
		logs.Errorf("[%s] validate vpc not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Aws, checkParams, len(delVpcFromCloud), kt.Rid)
		return fmt.Errorf("validate vpc not exist failed, before delete")
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err := cli.dbCli.Global.Vpc.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete vpc failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync vpc to delete vpc success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) updateVpc(kt *kit.Kit, accountID string, updateMap map[string]types.AwsVpc) error {
	if len(updateMap) == 0 {
		return fmt.Errorf("update vpc, vpcs is required")
	}

	vpcs := make([]cloud.VpcUpdateReq[cloud.AwsVpcUpdateExt], 0)
	for id, item := range updateMap {
		tmpRes := cloud.VpcUpdateReq[cloud.AwsVpcUpdateExt]{
			ID: id,
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

		vpcs = append(vpcs, tmpRes)
	}

	updateReq := &cloud.VpcBatchUpdateReq[cloud.AwsVpcUpdateExt]{
		Vpcs: vpcs,
	}
	if err := cli.dbCli.Aws.Vpc.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch update db vpc failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync vpc to update vpc success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createVpc(kt *kit.Kit, accountID string, addVpcs []types.AwsVpc) error {
	if len(addVpcs) == 0 {
		return fmt.Errorf("create vpc, vpcs is required")
	}

	vpcs := make([]cloud.VpcCreateReq[cloud.AwsVpcCreateExt], 0)
	for _, item := range addVpcs {
		// need add vpc data
		tmpRes := cloud.VpcCreateReq[cloud.AwsVpcCreateExt]{
			AccountID: accountID,
			CloudID:   item.CloudID,
			BkBizID:   constant.UnassignedBiz,
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

		vpcs = append(vpcs, tmpRes)
	}

	createReq := &cloud.VpcBatchCreateReq[cloud.AwsVpcCreateExt]{
		Vpcs: vpcs,
	}
	if _, err := cli.dbCli.Aws.Vpc.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch create vpc failed, err: %v, rid: %s", enumor.Aws, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync vpc to create vpc success, accountID: %s, count: %d, rid: %s", enumor.Aws,
		accountID, len(addVpcs), kt.Rid)

	return nil
}

func isAwsVpcChange(item types.AwsVpc, info cloudcore.Vpc[cloudcore.AwsVpcExtension]) bool {
	if info.Name != item.Name {
		return true
	}

	if info.Region != item.Region {
		return true
	}

	if !assert.IsPtrStringEqual(info.Memo, item.Memo) {
		return true
	}

	cidrMap := make(map[string]cloudcore.AwsCidr)
	for _, one := range item.Extension.Cidr {
		cidrMap[one.Cidr] = one
	}
	for _, db := range info.Extension.Cidr {
		cloud, exist := cidrMap[db.Cidr]
		if !exist {
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

// RemoveVpcDeleteFromCloud ...
func (cli *client) RemoveVpcDeleteFromCloud(kt *kit.Kit, accountID string, region string) error {

	req := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
				&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: region},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Global.Vpc.List(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list vpc failed, err: %v, req: %v, rid: %s", enumor.Aws,
				err, req, kt.Rid)
			return err
		}

		cloudIDs := make([]string, 0)
		for _, one := range resultFromDB.Details {
			cloudIDs = append(cloudIDs, one.CloudID)
		}

		if len(cloudIDs) == 0 {
			break
		}

		var delCloudIDs []string
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID: accountID,
				Region:    region,
				CloudIDs:  cloudIDs,
			}
			delCloudIDs, err = cli.listRemoveVpcID(kt, params)
			if err != nil {
				return err
			}
		}

		if len(delCloudIDs) != 0 {
			if err = cli.deleteVpc(kt, accountID, region, delCloudIDs); err != nil {
				return err
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

func (cli *client) listRemoveVpcID(kt *kit.Kit, params *SyncBaseParams) ([]string, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	delCloudIDs := make([]string, 0)
	cloudIDs := params.CloudIDs
	for {
		opt := &adcore.AwsListOption{
			Region:   params.Region,
			CloudIDs: cloudIDs,
		}
		_, err := cli.cloudCli.ListVpc(kt, opt)
		if err != nil {
			if strings.Contains(err.Error(), aws.ErrVpcNotFound) {
				var delCloudID string
				cloudIDs, delCloudID = removeNotFoundCloudID(cloudIDs, err)
				delCloudIDs = append(delCloudIDs, delCloudID)

				if len(cloudIDs) <= 0 {
					break
				}

				continue
			}

			logs.Errorf("[%s] list vpc from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Aws, err,
				params.AccountID, opt, kt.Rid)
			return nil, err
		}

		break
	}

	return delCloudIDs, nil
}

func (cli *client) listVpcFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]types.AwsVpc, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &adcore.AwsListOption{
		Region:   params.Region,
		CloudIDs: params.CloudIDs,
	}
	result, err := cli.cloudCli.ListVpc(kt, opt)
	if err != nil {
		if strings.Contains(err.Error(), aws.ErrVpcNotFound) {
			return make([]types.AwsVpc, 0), nil
		}

		logs.Errorf("[%s] list vpc from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Aws, err,
			params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) listVpcFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]cloudcore.Vpc[cloudcore.AwsVpcExtension], error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: params.AccountID,
				},
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: params.CloudIDs,
				},
				&filter.AtomRule{
					Field: "region",
					Op:    filter.Equal.Factory(),
					Value: params.Region,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Aws.Vpc.ListVpcExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list vpc from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Aws, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

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

package gcp

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
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

	addVpc, updateMap, delCloudIDs := common.Diff[types.GcpVpc, cloudcore.Vpc[cloudcore.GcpVpcExtension]](
		vpcFromCloud, vpcFromDB, isGcpVpcChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteVpc(kt, params.AccountID, delCloudIDs); err != nil {
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

	return new(SyncResult), nil
}

// RemoveVpcDeleteFromCloud ...
func (cli *client) RemoveVpcDeleteFromCloud(kt *kit.Kit, accountID string) error {

	req := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
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
			logs.Errorf("[%s] request dataservice to list vpc failed, err: %v, req: %v, rid: %s", enumor.Gcp,
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

		var resultFromCloud []types.GcpVpc
		if len(cloudIDs) != 0 {
			params := &SyncBaseParams{
				AccountID: accountID,
				CloudIDs:  cloudIDs,
			}
			resultFromCloud, err = cli.listVpcFromCloud(kt, params)
			if err != nil {
				return err
			}
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, one.CloudID)
			}

			delCloudIDs := converter.MapKeyToStringSlice(cloudIDMap)
			if err = cli.deleteVpc(kt, accountID, delCloudIDs); err != nil {
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

// deleteVpc delete vpc from db, and validate if the vpc exist in cloud
func (cli *client) deleteVpc(kt *kit.Kit, accountID string, delCloudIDs []string) error {
	if len(delCloudIDs) == 0 {
		return fmt.Errorf("delete vpc, cloudIDs is required")
	}

	// 检查云上是否存在vpc
	checkParams := &SyncBaseParams{
		AccountID: accountID,
		CloudIDs:  delCloudIDs,
	}
	delVpcFromCloud, err := cli.listVpcFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delVpcFromCloud) > 0 {
		// 如果云上还有vpc，说明数据不一致，不能删除
		logs.Errorf("[%s] validate vpc not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.Gcp, checkParams, len(delVpcFromCloud), kt.Rid)
		return fmt.Errorf("validate vpc not exist failed, before delete")
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.Vpc.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete vpc failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync vpc to delete vpc success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

// updateVpc update vpc in db
func (cli *client) updateVpc(kt *kit.Kit, accountID string, updateMap map[string]types.GcpVpc) error {
	if len(updateMap) == 0 {
		return fmt.Errorf("update vpc, vpcs is required")
	}

	vpcs := make([]cloud.VpcUpdateReq[cloud.GcpVpcUpdateExt], 0)
	for id, item := range updateMap {
		tmpRes := cloud.VpcUpdateReq[cloud.GcpVpcUpdateExt]{
			ID: id,
			VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
				Name:     converter.ValToPtr(item.Name),
				Category: enumor.BizVpcCategory,
				Memo:     item.Memo,
				BkBizID:  0,
			},
			Extension: &cloud.GcpVpcUpdateExt{
				EnableUlaInternalIpv6: converter.ValToPtr(item.Extension.EnableUlaInternalIpv6),
				InternalIpv6Range:     &item.Extension.InternalIpv6Range,
				Mtu:                   item.Extension.Mtu,
				RoutingMode:           &item.Extension.RoutingMode,
			},
		}

		vpcs = append(vpcs, tmpRes)
	}

	updateReq := &cloud.VpcBatchUpdateReq[cloud.GcpVpcUpdateExt]{
		Vpcs: vpcs,
	}
	if err := cli.dbCli.Gcp.Vpc.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch update db vpc failed, err: %v, rid: %s", enumor.Gcp,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync vpc to update vpc success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(updateMap), kt.Rid)

	return nil
}

// createVpc create vpc in db
func (cli *client) createVpc(kt *kit.Kit, accountID string, addVpc []types.GcpVpc) error {
	if len(addVpc) == 0 {
		return fmt.Errorf("create vpc, vpcs is required")
	}

	vpcs := make([]cloud.VpcCreateReq[cloud.GcpVpcCreateExt], 0, len(addVpc))
	for _, item := range addVpc {
		tmpRes := cloud.VpcCreateReq[cloud.GcpVpcCreateExt]{
			Region:    item.Region,
			AccountID: accountID,
			CloudID:   item.CloudID,
			BkBizID:   constant.UnassignedBiz,
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

		vpcs = append(vpcs, tmpRes)
	}

	createReq := &cloud.VpcBatchCreateReq[cloud.GcpVpcCreateExt]{
		Vpcs: vpcs,
	}
	if _, err := cli.dbCli.Gcp.Vpc.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch create vpc failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync vpc to create vpc success, accountID: %s, count: %d, rid: %s", enumor.Gcp,
		accountID, len(addVpc), kt.Rid)

	return nil
}

// listVpcFromCloud lists vpc from cloud
func (cli *client) listVpcFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]types.GcpVpc, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.GcpListOption{
		CloudIDs: params.CloudIDs,
		Page: &adcore.GcpPage{
			PageSize: adcore.GcpQueryLimit,
		},
	}
	result, err := cli.cloudCli.ListVpc(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list vpc from cloud failed, err: %v, account: %s, opt: %v, rid: %s", enumor.Gcp, err,
			params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

// listVpcFromCloudBySelfLink lists vpc from cloud by self link
func (cli *client) listVpcFromCloudBySelfLink(kt *kit.Kit, params *ListBySelfLinkOption) ([]types.GcpVpc, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.GcpListOption{
		SelfLinks: params.SelfLink,
		Page: &adcore.GcpPage{
			PageSize: adcore.GcpQueryLimit,
		},
	}
	result, err := cli.cloudCli.ListVpc(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list vpc from cloud by self link failed, err: %v, account: %s, opt: %v, rid: %s",
			enumor.Gcp, err, params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

// listVpcFromDB lists vpc from db
func (cli *client) listVpcFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]cloudcore.Vpc[cloudcore.GcpVpcExtension], error) {

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
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Gcp.Vpc.ListVpcExt(kt, req)
	if err != nil {
		logs.Errorf("[%s] list vpc from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Gcp, err,
			params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

// listVpcFromDBBySelfLink lists vpc from db by self link
func (cli *client) listVpcFromDBBySelfLink(kt *kit.Kit, opt *ListBySelfLinkOption) (
	[]cloudcore.Vpc[cloudcore.GcpVpcExtension], error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: opt.AccountID,
				},
				&filter.AtomRule{
					Field: "extension.self_link",
					Op:    filter.JSONIn.Factory(),
					Value: opt.SelfLink,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Gcp.Vpc.ListVpcExt(kt, req)
	if err != nil {
		logs.Errorf("[%s] list vpc from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Gcp, err,
			opt.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

// isGcpVpcChange checks if the GCP VPC has changed
func isGcpVpcChange(item types.GcpVpc, info cloudcore.Vpc[cloudcore.GcpVpcExtension]) bool {
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

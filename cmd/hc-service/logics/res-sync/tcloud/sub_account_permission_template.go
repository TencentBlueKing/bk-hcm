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

package tcloud

import (
	"fmt"

	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	dssubaccount "hcm/pkg/api/data-service/cloud/sub-account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SyncSubAccountPermissionTmplOption defines options for syncing subaccount permission templates.
type SyncSubAccountPermissionTmplOption struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate SyncSubAccountPermissionTmplOption.
func (opt SyncSubAccountPermissionTmplOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// SubAccountPermissionTemplate 同步指定账号下所有子账号绑定的权限模板信息。
func (cli *client) SubAccountPermissionTemplate(kt *kit.Kit, opt *SyncSubAccountPermissionTmplOption) (
	*SyncResult, error) {

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	subAccounts, err := cli.listSubAccountFromDB(kt, &SyncSubAccountOption{AccountID: opt.AccountID})
	if err != nil {
		return nil, err
	}

	if len(subAccounts) == 0 {
		logs.Infof("[%s] sync sub account permission template: no sub accounts found, accountID: %s, rid: %s",
			enumor.TCloud, opt.AccountID, kt.Rid)
		return new(SyncResult), nil
	}

	cloudIDToLocalID, err := cli.buildPermissionTmplCloudIDMap(kt, opt.AccountID)
	if err != nil {
		return nil, err
	}

	updateItems := make([]dssubaccount.UpdateField, 0, len(subAccounts))
	for _, subAccount := range subAccounts {
		if subAccount.AccountType == string(enumor.MainAccount) {
			continue
		}

		if subAccount.Extension == nil || subAccount.Extension.Uin == nil {
			logs.Errorf("[%s] sync sub account(%s) failed, extension has no uin, err: %v, rid: %s",
				enumor.TCloud, subAccount.ID, err, kt.Rid)
			return nil, errf.NewFromErr(errf.InvalidParameter,
				fmt.Errorf("sub account %s has no uin", subAccount.ID))
		}

		templateIDs, err := cli.listSubAccountPermissionTemplateIDs(
			kt, converter.PtrToVal(subAccount.Extension.Uin), cloudIDToLocalID)
		if err != nil {
			logs.Errorf("[%s] list sub account(%s) permission template ids failed, err: %v, rid: %s",
				enumor.TCloud, subAccount.ID, err, kt.Rid)
			return nil, err
		}

		if assert.IsStringSliceEqual(subAccount.PermissionTemplateIDs, templateIDs) {
			continue
		}

		updateItems = append(updateItems, dssubaccount.UpdateField{
			ID:                    subAccount.ID,
			PermissionTemplateIDs: templateIDs,
		})
	}

	if len(updateItems) > 0 {
		updateReq := &dssubaccount.UpdateReq{Items: updateItems}
		if err = cli.dbCli.Global.SubAccount.BatchUpdate(kt, updateReq); err != nil {
			logs.Errorf("[%s] batch update sub account permission template ids failed, err: %v, rid: %s",
				enumor.TCloud, err, kt.Rid)
			return nil, err
		}
	}

	logs.Infof("[%s] sync sub account permission template done, accountID: %s, total: %d, rid: %s",
		enumor.TCloud, opt.AccountID, len(subAccounts), kt.Rid)

	return new(SyncResult), nil
}

// listSubAccountPermissionTemplateIDs 获取单个子账号绑定的所有本地权限模板 ID 列表。
func (cli *client) listSubAccountPermissionTemplateIDs(kt *kit.Kit, uin uint64, cloudIDToLocalID map[string]string) (
	[]string, error) {

	const pageSize = uint64(100)
	templateIDs := make([]string, 0)

	// page是1开始的
	for page := uint64(1); ; page++ {
		result, err := cli.cloudCli.ListAttachedUserAllPolicies(kt,
			&typeaccount.TCloudListAttachedUserAllPoliciesOption{
				TargetUin: uin,
				Page:      page,
				Rp:        pageSize,
				// 直接关联和组关联都返回，所以 attach_type 传 0
				AttachType: converter.ValToPtr(uint64(0)),
			})
		if err != nil {
			return nil, err
		}

		for _, policy := range result.PolicyList {
			localID, ok := cloudIDToLocalID[policy.PolicyID]
			if !ok {
				logs.Errorf("[%s] policy cloud_id %s of account(cloud_id: %d) not found, rid: %s",
					enumor.TCloud, policy.PolicyID, uin, kt.Rid)
				return nil, errf.NewFromErr(errf.InvalidParameter,
					fmt.Errorf("[%s] policy cloud_id %s account(cloud_id: %d) not found",
						enumor.TCloud, policy.PolicyID, uin))
			}
			templateIDs = append(templateIDs, localID)
		}

		if uint64(len(result.PolicyList)) < pageSize || uint64(len(templateIDs)) >= result.TotalNum {
			break
		}
	}

	return slice.Unique(templateIDs), nil
}

// buildPermissionTmplCloudIDMap 构建 cloud_id → local_id 的映射表。
func (cli *client) buildPermissionTmplCloudIDMap(kt *kit.Kit, accountID string) (map[string]string, error) {
	req := &protocloud.PermissionTemplateExtListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", enumor.TCloud),
			tools.RuleEqual("account_id", accountID)),
		Page: core.NewDefaultBasePage(),
	}

	cloudIDToLocalID := make(map[string]string)
	start := uint32(0)
	for {
		req.Page.Start = start
		resp, err := cli.dbCli.TCloud.PermissionTemplate.ListPermissionTemplateExt(kt, req)
		if err != nil {
			logs.Errorf("[%s] list permission template from db failed, accountID: %s, err: %v, rid: %s",
				enumor.TCloud, accountID, err, kt.Rid)
			return nil, err
		}

		for _, tmpl := range resp.Details {
			cloudIDToLocalID[tmpl.CloudID] = tmpl.ID
		}

		if len(resp.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
		start += uint32(core.DefaultMaxPageLimit)
	}

	return cloudIDToLocalID, nil
}

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
	"crypto/sha256"
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SyncPermissionTemplateOption defines options for syncing permission templates.
type SyncPermissionTemplateOption struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate SyncPermissionTemplateOption.
func (opt SyncPermissionTemplateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// PermissionTemplate sync permission templates for the given account.
func (cli *client) PermissionTemplate(kt *kit.Kit, opt *SyncPermissionTemplateOption) (*SyncResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	fromCloud, err := cli.listPermissionTemplateFromCloud(kt, opt)
	if err != nil {
		return nil, err
	}

	fromDB, err := cli.listPermissionTemplateFromDB(kt, opt)
	if err != nil {
		return nil, err
	}

	if len(fromCloud) == 0 && len(fromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typeaccount.TCloudPolicyDetail,
		corecloud.PermissionTemplate[corecloud.TCloudPermissionTemplateExtension]](
		fromCloud, fromDB, isPermissionTemplateChanged)

	if len(delCloudIDs) > 0 {
		if err = cli.deletePermissionTemplate(kt, opt, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createPermissionTemplate(kt, opt, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updatePermissionTemplate(kt, opt, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

// listPoliciesFromCloud 分页拉取云上所有策略列表。
func (cli *client) listPoliciesFromCloud(kt *kit.Kit, opt *SyncPermissionTemplateOption) (
	[]typeaccount.TCloudPolicyItem, error) {

	// TCloud拉取策略列表的接口最大分页大小为200
	const pageSize = uint64(200)
	results := make([]typeaccount.TCloudPolicyItem, 0)

	for page := uint64(1); ; page++ {
		listOpt := &typeaccount.TCloudListPoliciesOption{
			Page: page,
			Rp:   pageSize,
		}

		listItems, total, err := cli.cloudCli.ListPolicies(kt, listOpt)
		if err != nil {
			logs.Errorf("[%s] list policies from cloud failed, account: %s, err: %v, rid: %s",
				enumor.TCloud, opt.AccountID, err, kt.Rid)
			return nil, err
		}

		results = append(results, listItems...)

		// 已拉取完毕
		if uint64(len(results)) >= total || uint64(len(listItems)) < pageSize {
			break
		}
	}

	return results, nil
}

// getPolicyDetails 批量获取策略详情。
func (cli *client) getPolicyDetails(kt *kit.Kit, opt *SyncPermissionTemplateOption,
	items []typeaccount.TCloudPolicyItem) ([]typeaccount.TCloudPolicyDetail, error) {

	results := make([]typeaccount.TCloudPolicyDetail, 0, len(items))

	for _, item := range items {
		detail, err := cli.cloudCli.GetPolicyDetail(kt, &typeaccount.TCloudGetPolicyDetailOption{
			PolicyID: item.PolicyID,
		})
		if err != nil {
			logs.Errorf("[%s] get policy detail failed, policyID: %d, account: %s, err: %v, rid: %s",
				enumor.TCloud, item.PolicyID, opt.AccountID, err, kt.Rid)
			return nil, err
		}
		results = append(results, converter.PtrToVal(detail))
	}

	return results, nil
}

// listPermissionTemplateFromCloud 分页拉取云上所有策略，并逐个获取 PolicyDocument。
func (cli *client) listPermissionTemplateFromCloud(kt *kit.Kit, opt *SyncPermissionTemplateOption) (
	[]typeaccount.TCloudPolicyDetail, error) {

	items, err := cli.listPoliciesFromCloud(kt, opt)
	if err != nil {
		return nil, err
	}

	details, err := cli.getPolicyDetails(kt, opt, items)
	if err != nil {
		return nil, err
	}

	return details, nil
}

// listPermissionTemplateFromDB 分页拉取本地 permission_template 表中该账号的所有记录。
func (cli *client) listPermissionTemplateFromDB(kt *kit.Kit, opt *SyncPermissionTemplateOption) (
	[]corecloud.PermissionTemplate[corecloud.TCloudPermissionTemplateExtension], error) {

	req := &protocloud.PermissionTemplateExtListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", enumor.TCloud),
			tools.RuleEqual("account_id", opt.AccountID)),
		Page: core.NewDefaultBasePage(),
	}

	results := make([]corecloud.PermissionTemplate[corecloud.TCloudPermissionTemplateExtension], 0)
	start := uint32(0)
	for {
		req.Page.Start = start
		resp, err := cli.dbCli.TCloud.PermissionTemplate.ListPermissionTemplateExt(kt, req)
		if err != nil {
			logs.Errorf("[%s] list permission template from db failed, account: %s, err: %v, rid: %s",
				enumor.TCloud, opt.AccountID, err, kt.Rid)
			return nil, err
		}

		results = append(results, resp.Details...)

		if len(resp.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
		start += uint32(core.DefaultMaxPageLimit)
	}

	return results, nil
}

// computePolicyHash computes SHA256 hash of policy document.
func computePolicyHash(policyDocument string) string {
	h := sha256.New()
	h.Write([]byte(policyDocument))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// isPermissionTemplateChanged 判断云上策略是否与本地记录有变化。
func isPermissionTemplateChanged(cloud typeaccount.TCloudPolicyDetail,
	db corecloud.PermissionTemplate[corecloud.TCloudPermissionTemplateExtension]) bool {

	cloudPolicyHash := computePolicyHash(cloud.PolicyDocument)
	if cloudPolicyHash != db.PolicyHash {
		return true
	}

	if cloud.PolicyName != db.Name {
		return true
	}

	if cloud.Description != converter.PtrToVal(db.Memo) {
		return true
	}

	return false
}

// createPermissionTemplate 批量创建云上新增的策略到本地。
func (cli *client) createPermissionTemplate(kt *kit.Kit, opt *SyncPermissionTemplateOption,
	addSlice []typeaccount.TCloudPolicyDetail) error {

	if len(addSlice) == 0 {
		return nil
	}

	batches := slice.Split(addSlice, constant.CloudResourceSyncMaxLimit)
	for _, batch := range batches {
		items := make([]protocloud.PermissionTemplateCreate[corecloud.TCloudPermissionTemplateExtension], 0, len(batch))
		for _, one := range batch {
			memo := one.Description
			items = append(items, protocloud.PermissionTemplateCreate[corecloud.TCloudPermissionTemplateExtension]{
				CloudID:        one.GetCloudID(),
				Name:           one.PolicyName,
				AccountID:      opt.AccountID,
				PolicyDocument: one.PolicyDocument,
				Memo:           converter.ValToPtr(memo),
				Extension: &corecloud.TCloudPermissionTemplateExtension{
					CloudType: one.PolicyType,
				},
			})
		}

		createReq := &protocloud.PermissionTemplateBatchCreateReq[corecloud.TCloudPermissionTemplateExtension]{
			PermissionTemplates: items,
		}
		if _, err := cli.dbCli.TCloud.PermissionTemplate.BatchCreate(kt, createReq); err != nil {
			logs.Errorf("[%s] create permission template failed, account: %s, err: %v, rid: %s",
				enumor.TCloud, opt.AccountID, err, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync permission template to create success, account: %s, count: %d, rid: %s",
		enumor.TCloud, opt.AccountID, len(addSlice), kt.Rid)

	return nil
}

// updatePermissionTemplate 批量更新 policy_hash 有变化的本地记录。
func (cli *client) updatePermissionTemplate(kt *kit.Kit, opt *SyncPermissionTemplateOption,
	updateMap map[string]typeaccount.TCloudPolicyDetail) error {

	if len(updateMap) == 0 {
		return nil
	}

	items := make([]protocloud.PermissionTemplateUpdate[corecloud.TCloudPermissionTemplateExtension], 0, len(updateMap))
	for id, one := range updateMap {
		memo := one.Description
		items = append(items, protocloud.PermissionTemplateUpdate[corecloud.TCloudPermissionTemplateExtension]{
			ID:             id,
			Name:           one.PolicyName,
			PolicyDocument: one.PolicyDocument,
			Memo:           converter.ValToPtr(memo),
			Extension: &corecloud.TCloudPermissionTemplateExtension{
				CloudType: one.PolicyType,
			},
		})
	}

	// 分批更新
	idBatches := slice.Split(items, constant.CloudResourceSyncMaxLimit)
	for _, batch := range idBatches {
		updateReq := &protocloud.PermissionTemplateBatchUpdateReq[corecloud.TCloudPermissionTemplateExtension]{
			PermissionTemplates: batch,
		}
		if err := cli.dbCli.TCloud.PermissionTemplate.BatchUpdate(kt, updateReq); err != nil {
			logs.Errorf("[%s] update permission template failed, account: %s, err: %v, rid: %s",
				enumor.TCloud, opt.AccountID, err, kt.Rid)
			return err
		}
	}

	logs.Infof("[%s] sync permission template to update success, account: %s, count: %d, rid: %s",
		enumor.TCloud, opt.AccountID, len(updateMap), kt.Rid)
	return nil
}

// deletePermissionTemplate 删除云上已不存在的本地记录。
func (cli *client) deletePermissionTemplate(kt *kit.Kit, opt *SyncPermissionTemplateOption,
	delCloudIDs []string) error {

	if len(delCloudIDs) == 0 {
		return nil
	}

	elems := slice.Split(delCloudIDs, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		deleteReq := &protocloud.PermissionTemplateBatchDeleteReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("cloud_id", parts),
				tools.RuleEqual("vendor", enumor.TCloud)),
		}
		if err := cli.dbCli.Global.PermissionTemplate.BatchDelete(kt, deleteReq); err != nil {
			logs.Errorf("[%s] delete permission template failed, account: %s, err: %v, rid: %s",
				enumor.TCloud, opt.AccountID, err, kt.Rid)
			return fmt.Errorf("delete permission template failed, err: %v", err)
		}
	}

	logs.Infof("[%s] sync permission template to delete success, account: %s, count: %d, rid: %s",
		enumor.TCloud, opt.AccountID, len(delCloudIDs), kt.Rid)

	return nil
}

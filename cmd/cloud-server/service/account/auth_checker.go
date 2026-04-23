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

package account

import (
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// accountResTypeAuthChecker 账号资源类型权限校验器接口，用于批量过滤有权限的账号ID
// 新增资源类型校验只需实现该接口并注册到 typeCheckerMap 即可
type accountResTypeAuthChecker interface {
	// filterAuthorizedIDs 过滤出有权限的账号ID列表
	filterAuthorizedIDs(kt *kit.Kit, accountIDs []string, bizID int64, vendor enumor.Vendor) ([]string, error)
}

// newAuthChecker 根据资源类型创建对应的权限校验器
func newAuthChecker(client *client.ClientSet, resType meta.ResourceType) (accountResTypeAuthChecker, error) {
	switch resType {
	case meta.SubAccount:
		return &subAccountAuthChecker{client: client}, nil
	case meta.SubAccountSecret:
		return &subAccountSecretAuthChecker{client: client}, nil
	case meta.PermissionTemplate:
		return &permissionTemplateAuthChecker{client: client}, nil
	default:
		return nil, errf.Newf(errf.InvalidParameter, "the checker not support res_type: %s", resType)
	}
}

// subAccountAuthChecker 三级账号资源权限校验器
type subAccountAuthChecker struct {
	client *client.ClientSet
}

// filterAuthorizedIDs 查询 sub_account 表，筛选条件为 account_id IN ids AND bk_biz_ids JSON_CONTAINS bizID AND vendor=vendor
func (c *subAccountAuthChecker) filterAuthorizedIDs(kt *kit.Kit, accountIDs []string, bizID int64,
	vendor enumor.Vendor) ([]string, error) {

	var authorizedIDs []string
	for _, batch := range slice.Split(accountIDs, int(core.DefaultMaxPageLimit)) {
		req := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("account_id", batch),
				tools.RuleJSONContains("bk_biz_ids", bizID),
				tools.RuleEqual("vendor", string(vendor)),
			),
			Page:   core.NewDefaultBasePage(),
			Fields: []string{"account_id"},
		}

		result, err := c.client.DataService().Global.SubAccount.List(kt, req)
		if err != nil {
			logs.Errorf("list sub_account failed, biz_id: %d, vendor: %s, err: %v, rid: %s",
				bizID, vendor, err, kt.Rid)
			return nil, err
		}

		for _, detail := range result.Details {
			authorizedIDs = append(authorizedIDs, detail.AccountID)
		}
	}

	return slice.Unique(authorizedIDs), nil
}

// subAccountSecretAuthChecker 三级账号密钥资源权限校验器
type subAccountSecretAuthChecker struct {
	client *client.ClientSet
}

// filterAuthorizedIDs 先查询 sub_account 满足条件的记录，再查询 sub_account_secret 确认存在密钥
func (c *subAccountSecretAuthChecker) filterAuthorizedIDs(kt *kit.Kit, accountIDs []string, bizID int64,
	vendor enumor.Vendor) ([]string, error) {

	// 1. 分页查询满足条件的 sub_account 记录
	var subAccountIDs []string
	for _, batch := range slice.Split(accountIDs, int(core.DefaultMaxPageLimit)) {
		subAccountReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("account_id", batch),
				tools.RuleJSONContains("bk_biz_ids", bizID),
				tools.RuleEqual("vendor", string(vendor)),
			),
			Page:   core.NewDefaultBasePage(),
			Fields: []string{"id"},
		}

		subAccountResult, err := c.client.DataService().Global.SubAccount.List(kt, subAccountReq)
		if err != nil {
			logs.Errorf("list sub_account for secret check failed, biz_id: %d, vendor: %s, err: %v, rid: %s",
				bizID, vendor, err, kt.Rid)
			return nil, err
		}

		for _, detail := range subAccountResult.Details {
			subAccountIDs = append(subAccountIDs, detail.ID)
		}
	}

	if len(subAccountIDs) == 0 {
		return []string{}, nil
	}

	// 2. 查询 sub_account_secret 表，获取有密钥记录的 account_id
	return c.listAccountIDsBySubAccountIDs(kt, subAccountIDs)
}

// listAccountIDsBySubAccountIDs 根据有密钥的 sub_account ID 列表查询对应的 account ID
func (c *subAccountSecretAuthChecker) listAccountIDsBySubAccountIDs(kt *kit.Kit,
	subAccountIDs []string) ([]string, error) {

	var authorizedIDs []string
	for _, batch := range slice.Split(subAccountIDs, int(core.DefaultMaxPageLimit)) {
		secretReq := &protocloud.SubAccountSecretListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("sub_account_id", batch),
			),
			Page: core.NewDefaultBasePage(),
		}

		secretResult, err := c.client.DataService().Global.SubAccountSecret.ListSubAccountSecret(kt, secretReq)
		if err != nil {
			logs.Errorf("list sub_account_secret details failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, detail := range secretResult.Details {
			authorizedIDs = append(authorizedIDs, detail.AccountID)
		}
	}

	return slice.Unique(authorizedIDs), nil
}

// permissionTemplateAuthChecker 权限模版资源权限校验器
type permissionTemplateAuthChecker struct {
	client *client.ClientSet
}

// filterAuthorizedIDs 查询 account 表，校验当前业务是否属于该账号的使用业务
func (c *permissionTemplateAuthChecker) filterAuthorizedIDs(kt *kit.Kit, accountIDs []string, bizID int64,
	vendor enumor.Vendor) ([]string, error) {

	var authorizedIDs []string
	for _, batch := range slice.Split(accountIDs, int(core.DefaultMaxPageLimit)) {
		req := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("id", batch),
				tools.RuleEqual("vendor", vendor),
			),
			Page:   core.NewDefaultBasePage(),
			Fields: []string{"id", "usage_biz_ids"},
		}

		result, err := c.client.DataService().Global.Account.List(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("list account for permission_template check failed, biz_id: %d, err: %v, rid: %s",
				bizID, err, kt.Rid)
			return nil, err
		}

		for _, detail := range result.Details {
			if slice.IsItemInSlice(detail.UsageBizIDs, bizID) {
				authorizedIDs = append(authorizedIDs, detail.ID)
			}
		}
	}

	return slice.Unique(authorizedIDs), nil
}

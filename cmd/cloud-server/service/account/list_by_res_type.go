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
	proto "hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListAccountByResType 根据资源类型批量查询二级账号元数据信息
func (a *accountSvc) ListAccountByResType(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountListByResTypeReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 1. 校验业务访问权限
	if err := a.checkBizAccessPermission(cts, bizID); err != nil {
		return nil, err
	}

	// 2. 根据资源类型过滤有权限的账号ID
	authorizedIDs, err := a.bizFilterAuthorizedAccountIDs(cts, req, bizID, vendor)
	if err != nil {
		return nil, err
	}
	if len(authorizedIDs) == 0 {
		return &proto.AccountListByResTypeResp{Details: []proto.AccountInfoByTypeDetail{}}, nil
	}

	// 3. 批量查询账号详情（基本信息 + 扩展字段）
	details, err := a.getAccountDetails(cts, authorizedIDs, vendor)
	if err != nil {
		return nil, err
	}

	return &proto.AccountListByResTypeResp{Details: details}, nil
}

// checkBizAccessPermission 校验用户是否有业务访问权限
func (a *accountSvc) checkBizAccessPermission(cts *rest.Contexts, bizID int64) error {
	attribute := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID}
	_, authorized, err := a.authorizer.Authorize(cts.Kit, attribute)
	if err != nil {
		return err
	}
	if !authorized {
		return errf.New(errf.PermissionDenied, "biz permission denied")
	}
	return nil
}

// bizFilterAuthorizedAccountIDs 根据资源类型调用对应校验器，过滤出有权限的账号ID
func (a *accountSvc) bizFilterAuthorizedAccountIDs(cts *rest.Contexts, req *proto.AccountListByResTypeReq,
	bizID int64, vendor enumor.Vendor) ([]string, error) {

	checker, err := newAuthChecker(a.client, req.ResType)
	if err != nil {
		return nil, err
	}

	authorizedIDs, err := checker.filterAuthorizedIDs(cts.Kit, req.IDs, bizID, vendor)
	if err != nil {
		logs.Errorf("filter authorized account ids failed, res_type: %s, biz_id: %d, err: %v, rid: %s",
			req.ResType, bizID, err, cts.Kit.Rid)
		return nil, err
	}

	return authorizedIDs, nil
}

// getAccountDetails 批量查询账号详情，包含基本信息和扩展字段，并组装为响应结构体
func (a *accountSvc) getAccountDetails(cts *rest.Contexts, authorizedIDs []string,
	vendor enumor.Vendor) ([]proto.AccountInfoByTypeDetail, error) {

	// 1. 批量查询账号基本信息
	accountMap, err := a.batchGetAccountBaseInfo(cts, authorizedIDs)
	if err != nil {
		return nil, err
	}

	// 2. 获取云厂商扩展字段
	extensionMap, err := a.getAccountExtensions(cts, authorizedIDs, vendor)
	if err != nil {
		return nil, err
	}

	// 3. 组装响应
	return buildAccountInfoByTypeDetails(accountMap, extensionMap, authorizedIDs), nil
}

// batchGetAccountBaseInfo 批量查询账号基本信息
func (a *accountSvc) batchGetAccountBaseInfo(cts *rest.Contexts, accountIDs []string) (
	map[string]*corecloud.BaseAccount, error) {

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("id", accountIDs)),
		Page:   core.NewDefaultBasePage(),
	}
	resp, err := a.client.DataService().Global.Account.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("list account base info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	accountMap := make(map[string]*corecloud.BaseAccount, len(resp.Details))
	for i := range resp.Details {
		accountMap[resp.Details[i].ID] = resp.Details[i]
	}

	return accountMap, nil
}

// getAccountExtensions 获取云厂商扩展字段
func (a *accountSvc) getAccountExtensions(cts *rest.Contexts, accountIDs []string,
	vendor enumor.Vendor) (map[string]map[string]interface{}, error) {

	listReq := &dataproto.AccountListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("id", accountIDs)),
		Page:   core.NewDefaultBasePage(),
	}

	resp, err := a.client.DataService().Global.Account.ListWithExtension(
		cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("list account with extension failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 去除 SecretKey，仅保留主账号ID等扩展字段
	secretKeyField := vendor.GetSecretField()
	extensionMap := make(map[string]map[string]interface{}, len(resp.Details))
	for _, detail := range resp.Details {
		if _, ok := detail.Extension[secretKeyField]; ok {
			delete(detail.Extension, secretKeyField)
		}
		extensionMap[detail.ID] = detail.Extension
	}

	return extensionMap, nil
}

// buildAccountInfoByTypeDetails 将账号基本信息和扩展字段组装为响应结构体
func buildAccountInfoByTypeDetails(accountMap map[string]*corecloud.BaseAccount,
	extensionMap map[string]map[string]interface{}, authorizedIDs []string) []proto.AccountInfoByTypeDetail {

	details := make([]proto.AccountInfoByTypeDetail, 0, len(authorizedIDs))
	for _, id := range authorizedIDs {
		account, ok := accountMap[id]
		if !ok {
			continue
		}

		extension := extensionMap[id]
		if extension == nil {
			extension = make(map[string]interface{})
		}

		details = append(details, proto.AccountInfoByTypeDetail{
			ID:          account.ID,
			Name:        account.Name,
			BkBizID:     account.BkBizID,
			UsageBizIDs: account.UsageBizIDs,
			Managers:    account.Managers,
			Extension:   extension,
		})
	}

	return details
}

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
	"fmt"

	proto "hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListAccountByResType 根据资源类型批量查询二级账号元数据信息
func (a *accountSvc) ListAccountByResType(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountListByResTypeReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := a.checkAdminPermission(cts.Kit, req.ResType); err != nil {
		return nil, err
	}

	details, err := a.getAccountDetails(cts.Kit, req.IDs, vendor)
	if err != nil {
		return nil, err
	}

	return &proto.AccountListByResTypeResp{Details: details}, nil
}

func (a *accountSvc) checkAdminPermission(kt *kit.Kit, resType meta.ResourceType) error {
	var authType meta.ResourceType
	var authAction meta.Action
	switch resType {
	case meta.PermissionPolicyLibrary:
		authType = meta.PermissionPolicyLibrary
		authAction = meta.Find
	default:
		return fmt.Errorf("invalid resource type: %s", resType)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: authType, Action: authAction}}
	if err := a.authorizer.AuthorizeWithPerm(kt, authRes); err != nil {
		return errf.NewFromErr(errf.PermissionDenied, err)
	}

	return nil
}

// ListBizAccountByResType 业务下根据资源类型批量查询二级账号元数据信息
func (a *accountSvc) ListBizAccountByResType(cts *rest.Contexts) (interface{}, error) {
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

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 1. 校验业务访问权限
	if err := a.checkBizAccessPermission(cts.Kit, bizID); err != nil {
		return nil, err
	}

	// 2. 根据资源类型过滤有权限的账号ID
	authorizedIDs, err := a.bizFilterAuthorizedAccountIDs(cts.Kit, req, bizID, vendor)
	if err != nil {
		return nil, err
	}
	if len(authorizedIDs) == 0 {
		return &proto.AccountListByResTypeResp{Details: []proto.AccountInfoByResTypeDetail{}}, nil
	}

	// 3. 批量查询账号详情（基本信息 + 扩展字段）
	details, err := a.getAccountDetails(cts.Kit, authorizedIDs, vendor)
	if err != nil {
		return nil, err
	}

	return &proto.AccountListByResTypeResp{Details: details}, nil
}

// checkBizAccessPermission 校验用户是否有业务访问权限
func (a *accountSvc) checkBizAccessPermission(kt *kit.Kit, bizID int64) error {
	attribute := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID}
	_, authorized, err := a.authorizer.Authorize(kt, attribute)
	if err != nil {
		return err
	}
	if !authorized {
		return errf.New(errf.PermissionDenied, "biz permission denied")
	}
	return nil
}

// bizFilterAuthorizedAccountIDs 根据资源类型调用对应校验器，过滤出有权限的账号ID
func (a *accountSvc) bizFilterAuthorizedAccountIDs(kt *kit.Kit, req *proto.AccountListByResTypeReq,
	bizID int64, vendor enumor.Vendor) ([]string, error) {

	checker, err := newAuthChecker(a.client, req.ResType)
	if err != nil {
		return nil, err
	}

	authorizedIDs, err := checker.filterAuthorizedIDs(kt, req.IDs, bizID, vendor)
	if err != nil {
		logs.Errorf("filter authorized account ids failed, res_type: %s, biz_id: %d, err: %v, rid: %s",
			req.ResType, bizID, err, kt.Rid)
		return nil, err
	}

	return authorizedIDs, nil
}

// getAccountDetails 批量查询账号详情，包含基本信息和扩展字段，并组装为响应结构体
func (a *accountSvc) getAccountDetails(kt *kit.Kit, authorizedIDs []string,
	vendor enumor.Vendor) ([]proto.AccountInfoByResTypeDetail, error) {

	// 1. 批量查询账号基本信息
	accounts, err := a.batchGetAccountBaseInfo(kt, vendor, authorizedIDs)
	if err != nil {
		return nil, err
	}

	// 2. 转换响应
	return convAccountInfoByResTypeDetails(kt, accounts)
}

// batchGetAccountBaseInfo 批量查询账号基本信息
func (a *accountSvc) batchGetAccountBaseInfo(kt *kit.Kit, vendor enumor.Vendor, accountIDs []string) (
	[]*dataproto.BaseAccountWithExtensionListResp, error) {

	accounts := make([]*dataproto.BaseAccountWithExtensionListResp, 0)
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleIn("id", accountIDs),
			tools.RuleEqual("vendor", string(vendor))),
		Page: core.NewDefaultBasePage(),
	}

	for {
		resp, err := a.client.DataService().Global.Account.ListWithExtension(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("list account base info failed, start: %d, err: %v, rid: %s", listReq.Page.Start, err, kt.Rid)
			return nil, err
		}

		for _, detail := range resp.Details {
			accounts = append(accounts, detail)
		}

		if uint(len(resp.Details)) < listReq.Page.Limit {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return accounts, nil
}

// convAccountInfoByResTypeDetails 将账号基本信息和扩展字段转换为响应结构体
func convAccountInfoByResTypeDetails(kt *kit.Kit, accounts []*dataproto.BaseAccountWithExtensionListResp,
) ([]proto.AccountInfoByResTypeDetail, error) {

	details := make([]proto.AccountInfoByResTypeDetail, 0, len(accounts))
	for _, account := range accounts {
		ext, err := convExtensionByVendor(account)
		if err != nil {
			logs.Errorf("conv extension failed, account_id: %s, err: %v, rid: %s", account.ID, err, kt.Rid)
			return nil, err
		}
		details = append(details, proto.AccountInfoByResTypeDetail{
			ID:          account.ID,
			Name:        account.Name,
			BkBizID:     account.BkBizID,
			Vendor:      account.Vendor,
			UsageBizIDs: account.UsageBizIDs,
			Managers:    account.Managers,
			Extension:   ext,
		})
	}

	return details, nil
}

func convExtensionByVendor(account *dataproto.BaseAccountWithExtensionListResp) (map[string]interface{}, error) {
	if account == nil {
		return nil, errf.Newf(errf.InvalidParameter, "account is nil")
	}

	ext := make(map[string]interface{})
	switch account.Vendor {
	// 过滤敏感信息
	case enumor.TCloud:
		ext["cloud_main_account_id"] = account.Extension["cloud_main_account_id"]
		return ext, nil
	default:
		return nil, errf.Newf(errf.InvalidParameter, "invalid vendor: %s", account.Vendor)
	}
}

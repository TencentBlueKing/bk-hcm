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

package subaccount

import (
	"fmt"

	logicaccount "hcm/cmd/cloud-server/logics/account"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	coresubaccount "hcm/pkg/api/core/cloud/sub-account"
	protocloud "hcm/pkg/api/data-service/cloud"
	dssubaccount "hcm/pkg/api/data-service/cloud/sub-account"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// ListSubAccount list sub account.
func (svc *service) ListSubAccount(cts *rest.Contexts) (interface{}, error) {
	return svc.listSubAccount(cts, handler.ListResourceAuthRes)
}

func (svc *service) listSubAccount(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.SubAccount, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		logs.Errorf("list sub account auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	req.Filter = expr

	result, err := svc.client.DataService().Global.SubAccount.List(cts.Kit, req)
	if err != nil {
		logs.Errorf("request ds to list sub account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// ListSubAccountExt list sub account.
func (svc *service) ListSubAccountExt(cts *rest.Contexts) (interface{}, error) {
	return svc.listSubAccountExt(cts, handler.ListResourceAuthRes)
}

func (svc *service) listSubAccountExt(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.SubAccount, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		logs.Errorf("list sub account auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	expr, err = tools.And(expr, tools.EqualExpression("vendor", vendor))
	if err != nil {
		logs.Errorf("expression append vendor rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	req.Filter = expr
	switch vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.SubAccount.ListExt(cts.Kit, req)
	case enumor.Aws:
		return svc.client.DataService().Aws.SubAccount.ListExt(cts.Kit, req)
	case enumor.HuaWei:
		return svc.client.DataService().HuaWei.SubAccount.ListExt(cts.Kit, req)
	case enumor.Azure:
		return svc.client.DataService().Azure.SubAccount.ListExt(cts.Kit, req)
	case enumor.Gcp:
		return svc.client.DataService().Gcp.SubAccount.ListExt(cts.Kit, req)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "vendor: %s not support", vendor)
	}
}

// ListBizSubAccountExt list biz sub account with extension.
func (svc *service) ListBizSubAccountExt(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	return svc.listBizSubAccountExt(cts, bkBizID)
}

func (svc *service) listBizSubAccountExt(cts *rest.Contexts, bkBizID int64) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	expr, noPermFlag, err := svc.listBizSubAccountAuthRes(cts, req.Filter)
	if err != nil {
		logs.Errorf("list sub account auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	expr, err = tools.And(expr, tools.EqualExpression("vendor", vendor))
	if err != nil {
		logs.Errorf("expression append vendor rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	req.Filter = expr

	return svc.listBizSubAccountExtByVendor(cts, vendor, bkBizID, req)
}

func convertBizSubAccountExtList[Ext coresubaccount.Extension](svc *service, kt *kit.Kit, bkBizID int64,
	listResult *dssubaccount.ListExtResult[Ext]) (*coresubaccount.BizSubAccountExtListResult[Ext], error) {

	accountIDs := extractAccountIDsFromSubAccountList(listResult.Details)
	accountMap, operableMap, err := logicaccount.BatchBuildOperableAndNameMap(
		kt, svc.client.DataService(), bkBizID, accountIDs,
	)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(listResult.Details))
	for _, item := range listResult.Details {
		if item.ID != "" {
			ids = append(ids, item.ID)
		}
	}

	secretCountMap, err := buildSubAccountSecretCountMap(kt, svc, ids)
	if err != nil {
		logs.Errorf("build secret count map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	permissionTemplateMap, err := buildPermissionTemplateMap(svc.client, kt, listResult.Details)
	if err != nil {
		logs.Errorf("build permission template map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	accountNameMap := logicaccount.BuildAccountNameMapByAccountMap(accountIDs, accountMap)
	return &coresubaccount.BizSubAccountExtListResult[Ext]{
		Count: listResult.Count,
		Details: buildBizSubAccountExtDetails(listResult.Details, operableMap, accountNameMap,
			secretCountMap, permissionTemplateMap),
	}, nil
}

func extractAccountIDsFromSubAccountList[Ext coresubaccount.Extension](
	details []coresubaccount.SubAccount[Ext]) []string {

	result := make([]string, 0, len(details))
	for _, item := range details {
		if item.AccountID == "" {
			continue
		}

		result = append(result, item.AccountID)
	}

	return result
}

// buildSubAccountSecretCountMap queries all secrets for the given sub-account IDs
// with pagination and returns a map of sub-account ID to secret count.
func buildSubAccountSecretCountMap(kt *kit.Kit, svc *service, subAccountIDs []string) (map[string]uint64, error) {
	countMap := make(map[string]uint64, len(subAccountIDs))
	if len(subAccountIDs) == 0 {
		return countMap, nil
	}

	req := &protocloud.SubAccountSecretListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("sub_account_id", subAccountIDs)),
		Page:   &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
	}
	for {
		result, err := svc.client.DataService().Global.SubAccountSecret.ListSubAccountSecret(kt, req)
		if err != nil {
			logs.Errorf("list sub account secrets failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("list sub account secrets failed, err: %v", err)
		}

		for _, item := range result.Details {
			countMap[item.SubAccountID]++
		}

		if uint(len(result.Details)) < core.DefaultMaxPageLimit {
			break
		}
		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return countMap, nil
}

func buildBizSubAccountExtDetails[Ext coresubaccount.Extension](details []coresubaccount.SubAccount[Ext],
	operableMap map[string]bool, accountNameMap map[string]string, secretCountMap map[string]uint64,
	permTmplMap map[string][]corecloud.PermissionTmplBasicInfo) []coresubaccount.BizSubAccountItem[Ext] {

	result := make([]coresubaccount.BizSubAccountItem[Ext], 0, len(details))
	for _, item := range details {
		result = append(result, coresubaccount.BizSubAccountItem[Ext]{
			SubAccount:            item,
			Operable:              operableMap[item.AccountID],
			AccountName:           accountNameMap[item.AccountID],
			PermissionTemplates:   permTmplMap[item.ID],
			SubAccountSecretCount: secretCountMap[item.ID],
		})
	}

	return result
}

func (svc *service) listBizSubAccountExtByVendor(cts *rest.Contexts, vendor enumor.Vendor, bkBizID int64,
	req *core.ListReq) (interface{}, error) {

	switch vendor {
	case enumor.TCloud:
		result, err := svc.client.DataService().TCloud.SubAccount.ListExt(cts.Kit, req)
		if err != nil {
			return nil, err
		}
		return convertBizSubAccountExtList(svc, cts.Kit, bkBizID, result)
	case enumor.Aws:
		result, err := svc.client.DataService().Aws.SubAccount.ListExt(cts.Kit, req)
		if err != nil {
			return nil, err
		}
		return convertBizSubAccountExtList(svc, cts.Kit, bkBizID, result)
	case enumor.HuaWei:
		result, err := svc.client.DataService().HuaWei.SubAccount.ListExt(cts.Kit, req)
		if err != nil {
			return nil, err
		}
		return convertBizSubAccountExtList(svc, cts.Kit, bkBizID, result)
	case enumor.Azure:
		result, err := svc.client.DataService().Azure.SubAccount.ListExt(cts.Kit, req)
		if err != nil {
			return nil, err
		}
		return convertBizSubAccountExtList(svc, cts.Kit, bkBizID, result)
	case enumor.Gcp:
		result, err := svc.client.DataService().Gcp.SubAccount.ListExt(cts.Kit, req)
		if err != nil {
			return nil, err
		}
		return convertBizSubAccountExtList(svc, cts.Kit, bkBizID, result)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "vendor: %s not support", vendor)
	}
}

func (svc *service) listBizSubAccountAuthRes(cts *rest.Contexts, reqFilter *filter.Expression,
) (*filter.Expression, bool, error) {

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, false, err
	}

	if bizID <= 0 {
		return nil, false, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.SubAccount, Action: meta.Find}, BizID: bizID}
	_, authorized, err := svc.authorizer.Authorize(cts.Kit, authRes)
	if err != nil {
		return nil, false, err
	}

	if !authorized {
		return nil, true, nil
	}

	// 查询管理业务为biz_id的资源类型二级账号ID列表
	accountIDs, err := logicaccount.ListAccountIDsByBizID(cts.Kit, svc.client.DataService(), bizID)
	if err != nil {
		return nil, false, err
	}

	scopeFilter := tools.ExpressionOr(
		tools.RuleJSONContains[int64]("bk_biz_ids", bizID),
		tools.RuleIn("account_id", accountIDs),
	)

	// 过滤主账号
	scopeFilter, err = tools.And(scopeFilter, tools.RuleNotEqual("account_type", string(enumor.MainAccount)))
	if err != nil {
		return nil, false, err
	}

	if reqFilter == nil {
		return scopeFilter, false, nil
	}

	finalFilter, err := tools.And(scopeFilter, reqFilter)
	if err != nil {
		return nil, false, err
	}

	return finalFilter, false, err
}

// buildPermissionTemplateMap 批量查询权限模版，构建 subAccountID -> []PermissionTmplBasicInfo 的映射。
func buildPermissionTemplateMap[Ext coresubaccount.Extension](cli *client.ClientSet, kt *kit.Kit,
	details []coresubaccount.SubAccount[Ext]) (map[string][]corecloud.PermissionTmplBasicInfo, error) {

	// 1. 收集所有 permission_template_id 并去重
	tmplIDs := make([]string, 0)
	for _, detail := range details {
		tmplIDs = append(tmplIDs, detail.PermissionTemplateIDs...)
	}
	tmplIDs = slice.Unique(tmplIDs)

	if len(tmplIDs) == 0 {
		return make(map[string][]corecloud.PermissionTmplBasicInfo), nil
	}

	// 2. 批量查询权限模版
	tmplMap := make(map[string]corecloud.PermissionTmplBasicInfo, len(tmplIDs))
	for _, batch := range slice.Split(tmplIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &protocloud.PermissionTemplateListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", batch)),
			Page:   core.NewDefaultBasePage(),
		}
		
		result, err := cli.DataService().Global.PermissionTemplate.ListPermissionTemplate(kt, listReq)
		if err != nil {
			logs.Errorf("list permission template failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, template := range result.Details {
			tmplMap[template.ID] = corecloud.PermissionTmplBasicInfo{
				ID:   template.ID,
				Name: template.Name,
			}
		}
	}

	// 3. 构建 subAccountID -> []PermissionTmplBasicInfo 的映射
	result := make(map[string][]corecloud.PermissionTmplBasicInfo, len(details))
	for _, detail := range details {
		templates := make([]corecloud.PermissionTmplBasicInfo, 0, len(detail.PermissionTemplateIDs))
		for _, tmplID := range detail.PermissionTemplateIDs {
			info, ok := tmplMap[tmplID]
			if !ok {
				logs.Errorf("permission template not found, tmpl_id: %s, rid: %s", tmplID, kt.Rid)
				return nil, errf.New(errf.Aborted, fmt.Sprintf("permission template not found, tmpl_id: %s", tmplID))
			}
			templates = append(templates, info)
		}
		result[detail.ID] = templates
	}

	return result, nil
}

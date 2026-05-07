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

package permissiontemplates

import (
	"encoding/json"
	"fmt"

	logicaccount "hcm/cmd/cloud-server/logics/account"
	cloudserver "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
)

// ListPermissionTemplate lists business-scoped cloud permission templates.
func (svc *service) ListPermissionTemplate(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if bizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "bk_biz_id is invalid")
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(cloudserver.ListBizPermissionTemplateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bizID,
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, errf.NewFromErr(errf.PermissionDenied, err)
	}

	switch vendor {
	case enumor.TCloud:
		return svc.listBizPermissionTmplForTCloud(cts.Kit, bizID, req)
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}
}

func (svc *service) listBizPermissionTmplForTCloud(kt *kit.Kit, bizID int64,
	req *cloudserver.ListBizPermissionTemplateReq) (*cloudserver.BizPermissionTemplateListResult, error) {

	ext := new(corecloud.TCloudPermissionTemplateListExt)
	if !req.Extension.IsEmpty() {
		if err := json.Unmarshal([]byte(req.Extension), ext); err != nil {
			return nil, fmt.Errorf("unmarshal tcloud extension failed: %w", err)
		}
	}

	accountIDs, err := getAccountIDs(kt, svc.client, enumor.TCloud, bizID, ext.CloudMainAccountIDs)
	if err != nil {
		return nil, err
	}

	if len(accountIDs) == 0 {
		return &cloudserver.BizPermissionTemplateListResult{
			Count: 0, Details: []cloudserver.BizPermissionTemplateDetail{},
		}, nil
	}

	dsReq := &protocloud.PermissionTmplJoinExtListReq{
		BkBizID: bizID,
		PermissionTemplateFilters: protocloud.PermissionTemplateFilters{
			IDs:        req.IDs,
			CloudIDs:   req.CloudIDs,
			Names:      req.Names,
			AccountIDs: accountIDs,
			Creator:    req.Creator,
			Reviser:    req.Reviser,
			Extension:  req.Extension,
		},
		Page: req.Page,
	}

	results, err := svc.client.DataService().TCloud.PermissionTemplate.ListPermissionTmplJoinExt(kt, dsReq)
	if err != nil {
		logs.Errorf("list tcloud permission template failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &cloudserver.BizPermissionTemplateListResult{Count: results.Count}, nil
	}

	return svc.convBizPermTmplListResult(kt, results)
}

// convBizPermTmplListResult assembles the cloud-server response from data-service result.
func (svc *service) convBizPermTmplListResult(kt *kit.Kit, dsResults *protocloud.PermissionTmplJoinExtListResult) (
	*cloudserver.BizPermissionTemplateListResult, error) {

	if dsResults == nil {
		return &cloudserver.BizPermissionTemplateListResult{Details: []cloudserver.BizPermissionTemplateDetail{}}, nil
	}

	// 补充二级账号cloud_id
	cloudIDMap, err := svc.batchLoadAccountCloudIDs(kt, dsResults.Details)
	if err != nil {
		return nil, err
	}

	// 补充权限策略库名称
	policyLibraryNameMap, err := svc.batchLoadPolicyLibraryNames(kt, dsResults.Details)
	if err != nil {
		return nil, err
	}

	details := make([]cloudserver.BizPermissionTemplateDetail, 0, len(dsResults.Details))
	for _, d := range dsResults.Details {
		details = append(details, cloudserver.BizPermissionTemplateDetail{
			BasePermissionTemplate:    d.BasePermissionTemplate,
			CloudAccountID:            cloudIDMap[d.AccountID],
			PolicyLibraryName:         policyLibraryNameMap[converter.PtrToVal(d.PolicyLibraryID)],
			AssociatedSubAccountCount: d.AssociatedSubAccountCount,
			Extension:                 d.Extension,
		})
	}

	return &cloudserver.BizPermissionTemplateListResult{Details: details}, nil
}

// batchLoadAccountCloudIDs collects account IDs from details, batch-queries the account table,
func (svc *service) batchLoadAccountCloudIDs(kt *kit.Kit, details []protocloud.PermissionTmplJoinExtDetail) (
	map[string]string, error) {

	accountIDs := make([]string, 0)
	for _, d := range details {
		accountIDs = append(accountIDs, d.AccountID)
	}
	result := make(map[string]string, len(accountIDs))

	for _, batch := range slice.Split(accountIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &protocloud.AccountListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", batch)),
			Page:   core.NewDefaultBasePage(),
		}

		accounts, err := svc.client.DataService().Global.Account.ListWithExtension(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("list account failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, item := range accounts.Details {
			cloudAccountID := ""
			if item.Extension != nil {
				if v, ok := item.Extension["cloud_main_account_id"]; ok {
					if s, ok := v.(string); ok {
						cloudAccountID = s
					}
				}
			}
			result[item.ID] = cloudAccountID
		}
	}

	return result, nil
}

// batchLoadPolicyLibraryNames queries permission_policy_library by the IDs collected from details,
func (svc *service) batchLoadPolicyLibraryNames(kt *kit.Kit, details []protocloud.PermissionTmplJoinExtDetail,
) (map[string]string, error) {

	libIDSet := make(map[string]struct{})
	for _, d := range details {
		if converter.PtrToVal(d.PolicyLibraryID) != "" {
			libIDSet[converter.PtrToVal(d.PolicyLibraryID)] = struct{}{}
		}
	}

	result := make(map[string]string, len(libIDSet))
	if len(libIDSet) == 0 {
		return result, nil
	}

	libIDs := make([]string, 0, len(libIDSet))
	for id := range libIDSet {
		libIDs = append(libIDs, id)
	}

	for _, batch := range slice.Split(libIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &protocloud.PermissionPolicyLibraryListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", batch)),
			Page:   core.NewDefaultBasePage(),
		}

		libs, err := svc.client.DataService().Global.PermissionPolicyLibrary.ListPermissionPolicyLibrary(
			kt, listReq)
		if err != nil {
			logs.Errorf("list perm policy library failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("list permission policy library failed, err: %v", err)
		}

		for _, lib := range libs.Details {
			result[lib.ID] = lib.Name
		}
	}

	return result, nil
}

// getAccountIDs resolves cloud-side main account IDs (cloud_id) to local
func getAccountIDs(kt *kit.Kit, cli *client.ClientSet, vendor enumor.Vendor, bkBizID int64, cloudMainAccountIDs []string) (
	[]string, error) {

	listReq := &protocloud.AccountListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bkBizID),
			tools.RuleEqual("vendor", vendor),
			tools.RuleEqual("type", enumor.ResourceAccount),
		),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"id"},
	}

	if len(cloudMainAccountIDs) != 0 {
		listReq.Filter.Rules = append(listReq.Filter.Rules,
			tools.RuleJsonIn("extension.cloud_main_account_id", cloudMainAccountIDs))
	}

	idSet := make(map[string]struct{})
	for {
		accounts, err := cli.DataService().Global.Account.List(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			return nil, err
		}

		for _, item := range accounts.Details {
			// ID不为空，且二级账号的使用业务中有当前业务的二级账号
			if item.ID != "" && slice.IsItemInSlice(item.UsageBizIDs, bkBizID) {
				idSet[item.ID] = struct{}{}
			}
		}

		if len(accounts.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return maps.Keys(idSet), nil
}

// ListPermTmplSubAccountIDs lists sub account ids associated with a permission template.
func (svc *service) ListPermTmplSubAccountIDs(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if bizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "bk_biz_id is invalid")
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	templateID := cts.PathParameter("id").String()
	if templateID == "" {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	// 业务访问鉴权
	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bizID,
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		return nil, errf.NewFromErr(errf.PermissionDenied, err)
	}

	switch vendor {
	case enumor.TCloud:
		return svc.listPermTmplSubAccountIDsForTCloud(cts.Kit, bizID, templateID)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "vendor: %s not support", vendor)
	}
}

func (svc *service) listPermTmplSubAccountIDsForTCloud(kt *kit.Kit, bizID int64,
	templateID string) (*cloudserver.PermTmplSubAccountIDsResult, error) {

	tmplListReq := &protocloud.PermissionTemplateExtListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("id", templateID)),
		Page: core.NewCountPage(),
	}
	tmplResult, err := svc.client.DataService().TCloud.PermissionTemplate.ListPermissionTemplateExt(kt, tmplListReq)
	if err != nil {
		return nil, err
	}
	if tmplResult == nil || tmplResult.Count == 0 {
		return nil, errf.New(errf.InvalidParameter, "permission template not exist")
	}

	// 查询管理业务为biz_id的资源类型二级账号ID列表
	accountIDs, err := logicaccount.ListAccountIDsByBizID(kt, svc.client.DataService(), bizID)
	if err != nil {
		logs.Errorf("list account ids by biz id failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	filterExpr := tools.ExpressionAnd(
		tools.RuleJSONContains("permission_template_ids", templateID),
	)
	if len(accountIDs) > 0 {
		filterExpr.Rules = append(filterExpr.Rules,
			tools.ExpressionOr(
				tools.RuleIn("account_id", accountIDs),
				tools.RuleJSONContains("bk_biz_ids", bizID),
			),
		)
	} else {
		filterExpr.Rules = append(filterExpr.Rules, tools.RuleJSONContains("bk_biz_ids", bizID))
	}

	subAccounts := make([]cloudserver.PermRelateSubAccountInfo, 0)
	listReq := &core.ListReq{Filter: filterExpr, Page: core.NewDefaultBasePage(), Fields: []string{"id", "cloud_id"}}
	for {
		results, err := svc.client.DataService().TCloud.SubAccount.ListExt(kt, listReq)
		if err != nil {
			return nil, err
		}

		for _, d := range results.Details {
			subAccounts = append(subAccounts, cloudserver.PermRelateSubAccountInfo{
				ID:      d.ID,
				CloudID: d.CloudID,
			})
		}

		if len(results.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	return &cloudserver.PermTmplSubAccountIDsResult{SubAccounts: subAccounts}, nil
}

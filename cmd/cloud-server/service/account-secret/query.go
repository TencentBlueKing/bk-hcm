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

package accountsecret

import (
	proto "hcm/pkg/api/cloud-server/account-secret"
	"hcm/pkg/api/core"
	coreas "hcm/pkg/api/core/cloud/account-secret"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// getAccountSecretByID gets account secret by id.
func (s *service) getAccountSecretByID(kt *kit.Kit, id string) (*coreas.BaseAccountSecret, error) {
	result, err := s.client.DataService().Global.AccountSecret.ListAccountSecret(kt, &protocloud.AccountSecretListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		logs.Errorf("list account secret failed, id: %s, err: %v, rid: %s", id, err, kt.Rid)
		return nil, err
	}

	if len(result.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "account secret %s not found", id)
	}

	return &result.Details[0], nil
}

func (s *service) getTCloudAccountSecretByID(kt *kit.Kit, id string) (
	*coreas.AccountSecret[coreas.TCloudAccountSecretExtension], error) {

	req := &protocloud.AccountSecretExtListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := s.client.DataService().TCloud.AccountSecret.ListAccountSecretWithExtension(kt, req)
	if err != nil {
		logs.Errorf("list account secret failed, id: %s, err: %v, rid: %s", id, err, kt.Rid)
		return nil, err
	}

	if len(result.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "account secret %s not found", id)
	}

	return &result.Details[0], nil
}

// ListBizAccountSecret list biz account secret.
func (s *service) ListBizAccountSecret(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountSecretListReq)
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
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	// 权限校验
	attribute := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID}
	_, authorized, err := s.authorizer.Authorize(cts.Kit, attribute)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "biz permission denied")
	}

	// 查询业务下的账号IDs
	accountIDs := make([]string, 0)
	accountFilter := tools.ExpressionAnd(
		tools.RuleEqual("type", string(enumor.ResourceAccount)),
		tools.RuleEqual("bk_biz_id", bizID),
		tools.RuleEqual("vendor", string(vendor)),
	)
	accountListReq := &core.ListReq{Filter: accountFilter, Page: core.NewDefaultBasePage(), Fields: []string{"id"}}
	for {
		accountResp, err := s.client.DataService().Global.Account.List(cts.Kit.Ctx, cts.Kit.Header(), accountListReq)
		if err != nil {
			logs.Errorf("list account failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		for _, account := range accountResp.Details {
			accountIDs = append(accountIDs, account.ID)
		}
		if len(accountResp.Details) < int(accountListReq.Page.Limit) {
			break
		}
		accountListReq.Page.Start += uint32(accountListReq.Page.Limit)
	}
	// 如果没有账号，返回空结果
	if len(accountIDs) == 0 {
		if req.Page.Count {
			return &core.ListResult{Count: 0}, nil
		}
		return &core.ListResult{Details: make([]interface{}, 0)}, nil
	}

	secretFilter, err := tools.And(buildAccountIDsFilter(accountIDs), req.Filter)
	if err != nil {
		logs.Errorf("merge filter failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	switch vendor {
	case enumor.TCloud:
		return s.listTCloudAccountSecret(cts.Kit, secretFilter, req.Page)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "vendor: %s not support", vendor)
	}
}

// buildAccountIDsFilter builds a filter rule for account_id field, splitting into batches of 500
// to respect the IN clause limit, and OR-ing them together.
func buildAccountIDsFilter(accountIDs []string) filter.RuleFactory {
	if len(accountIDs) <= int(filter.DefaultMaxInLimit) {
		return tools.RuleIn("account_id", accountIDs)
	}

	inRules := make([]*filter.AtomRule, 0)
	for i := 0; i < len(accountIDs); i += int(filter.DefaultMaxInLimit) {
		end := i + int(filter.DefaultMaxInLimit)
		if end > len(accountIDs) {
			end = len(accountIDs)
		}
		inRules = append(inRules, tools.RuleIn("account_id", accountIDs[i:end]))
	}
	return tools.ExpressionOr(inRules...)
}

// listTCloudAccountSecret list tcloud account secret.
func (s *service) listTCloudAccountSecret(kt *kit.Kit, filter *filter.Expression, page *core.BasePage) (
	interface{}, error) {

	listReq := &protocloud.AccountSecretExtListReq{Filter: filter, Page: page}
	secretResp, err := s.client.DataService().TCloud.AccountSecret.ListAccountSecretWithExtension(kt, listReq)
	if err != nil {
		logs.Errorf("list tcloud account secret failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if page.Count {
		return secretResp, nil
	}

	// 密钥脱敏处理
	for i := range secretResp.Details {
		if secretResp.Details[i].Extension != nil {
			secretResp.Details[i].Extension.CloudSecretKey = ""
		}
	}

	return secretResp, nil
}

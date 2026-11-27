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

package image

import (
	csimage "hcm/pkg/api/cloud-server/image"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// ListBizImage lists images visible to a business: all public/shared images plus the biz's private images.
func (svc *imageSvc) ListBizImage(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if bizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "bk_biz_id is invalid")
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err = vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(csimage.BizImageListReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("decode list biz image request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err = req.Validate(); err != nil {
		logs.Errorf("validate list biz image request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bizID,
	}
	if err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("authorize list biz image failed, err: %v, biz_id: %d, rid: %s", err, bizID, cts.Kit.Rid)
		return nil, err
	}

	listFilter, err := buildBizImageListFilter(bizID, vendor, req)
	if err != nil {
		logs.Errorf("build list biz image filter failed, err: %v, biz_id: %d, vendor: %s, rid: %s",
			err, bizID, vendor, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Filter: listFilter,
		Page:   req.Page,
	}
	result, err := svc.client.DataService().Global.ListImage(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("list biz image from data-service failed, err: %v, biz_id: %d, vendor: %s, rid: %s",
			err, bizID, vendor, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// buildBizImageListFilter assembles the filter for biz-dimension image query.
func buildBizImageListFilter(bizID int64, vendor enumor.Vendor, req *csimage.BizImageListReq) (
	*filter.Expression, error) {

	rules := make([]filter.RuleFactory, 0)

	rules = append(rules, buildBizImageVisibilityFilter(bizID))
	rules = append(rules, tools.RuleEqual("vendor", vendor))

	if len(req.Platform) > 0 {
		rules = append(rules, tools.RuleEqual("platform", req.Platform))
	}
	if len(req.Region) > 0 {
		rules = append(rules, tools.RuleEqual("region", req.Region))
	}
	if len(req.Name) > 0 {
		rules = append(rules, tools.RuleCis("name", req.Name))
	}
	if len(req.Type) > 0 {
		rules = append(rules, tools.RuleEqual("type", req.Type))
	}

	return tools.And(rules...)
}

// buildBizImageVisibilityFilter returns:
// type in (public, shared) OR (type == private AND bk_biz_id == bizID)
func buildBizImageVisibilityFilter(bizID int64) *filter.Expression {
	privateAndBiz := tools.ExpressionAnd(
		tools.RuleEqual("type", enumor.ImageTypePrivate),
		tools.RuleEqual("bk_biz_id", bizID),
	)

	return &filter.Expression{
		Op: filter.Or,
		Rules: []filter.RuleFactory{
			tools.RuleIn("type", []enumor.ImageTypeNormalized{enumor.ImageTypePublic, enumor.ImageTypeShared}),
			privateAndBiz,
		},
	}
}

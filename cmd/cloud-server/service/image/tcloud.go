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
	"hcm/pkg/api/core"
	"hcm/pkg/api/hc-service/image"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// TCloudQueryImage ...
func (svc *imageSvc) TCloudQueryImage(cts *rest.Contexts) (interface{}, error) {

	req, err := svc.decodeAndValidateTCloudImageListOption(cts)
	if err != nil {
		logs.Errorf("decode and validate tcloud image list option failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return svc.tcloudQueryImage(cts, req, handler.ResOperateAuth)
}

// TCLoudBizQueryImage ...
func (svc *imageSvc) TCLoudBizQueryImage(cts *rest.Contexts) (interface{}, error) {

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	req, err := svc.decodeAndValidateTCloudImageListOption(cts)
	if err != nil {
		logs.Errorf("decode and validate tcloud image list option failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	accountBizReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", bizID),
			tools.RuleEqual("account_id", req.AccountID),
		),
		Page: core.NewCountPage(),
	}
	result, err := svc.client.DataService().Global.Account.ListAccountBizRel(cts.Kit.Ctx, cts.Kit.Header(), accountBizReq)
	if err != nil {
		logs.Errorf("list account biz rel failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if result.Count == 0 {
		return nil, errf.New(errf.PermissionDenied, "no permission")
	}

	return svc.tcloudQueryImage(cts, req, handler.BizOperateAuth)
}

func (svc *imageSvc) decodeAndValidateTCloudImageListOption(cts *rest.Contexts) (*image.TCloudImageListOption, error) {
	req := new(image.TCloudImageListOption)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return req, nil
}

func (svc *imageSvc) tcloudQueryImage(cts *rest.Contexts, req *image.TCloudImageListOption,
	validHandler handler.ValidWithAuthHandler) (interface{}, error) {

	// validate biz and authorize
	err := validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Image,
		Action: meta.Find})
	if err != nil {
		return nil, err
	}

	return svc.client.HCService().TCloud.Image.ListImage(cts.Kit, req)
}

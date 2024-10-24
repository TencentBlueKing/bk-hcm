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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// TCLoudListImage ...
func (svc *imageSvc) TCLoudListImage(cts *rest.Contexts) (interface{}, error) {

	req, err := svc.decodeAndValidateTCloudImageListOption(cts)
	if err != nil {
		logs.Errorf("decode and validate tcloud image list option failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Account, Action: meta.Find,
		ResourceID: req.AccountID}}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("describe resources auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	_, err = svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		// 这里校验账号是否存在，出现错误大概率是账号不存在
		logs.V(3).Errorf("fail to get account info, err: %s, account id: %s, rid: %s",
			err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}

	return svc.tcloudListImage(cts.Kit, req)
}

// TCLoudBizListImage ...
func (svc *imageSvc) TCLoudBizListImage(cts *rest.Contexts) (interface{}, error) {

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

	return svc.tcloudListImage(cts.Kit, req)
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

func (svc *imageSvc) tcloudListImage(kt *kit.Kit, req *image.TCloudImageListOption) (interface{}, error) {
	return svc.client.HCService().TCloud.Image.ListImage(kt, req)
}

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
	"encoding/json"
	"fmt"

	"hcm/cmd/cloud-server/service/common"
	proto "hcm/pkg/api/cloud-server/account-secret"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// CheckBizAccountSecret checks account secret validity before creating or updating.
func (s *service) CheckBizAccountSecret(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountSecretCheckReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验业务访问权限
	attribute := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID}
	_, authorized, err := s.authorizer.Authorize(cts.Kit, attribute)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "biz permission denied")
	}

	// 查询账号基本信息
	listReq := &core.ListReq{
		Filter: tools.EqualExpression("id", req.AccountID),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"id", "bk_biz_id", "vendor"},
	}
	resp, err := s.client.DataService().Global.Account.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("get account basic info failed, account_id: %s, err: %v, rid: %s", req.AccountID, err, cts.Kit.Rid)
		return nil, err
	}
	if len(resp.Details) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "account %s not found", req.AccountID)
	}
	account := resp.Details[0]

	// 校验账号是否属于该业务
	if account.BkBizID != bizID {
		return nil, errf.Newf(errf.PermissionDenied, "account %s does not belong to business %d", req.AccountID, bizID)
	}

	// 调用密钥校验
	return s.checkAccountSecretByVendor(cts.Kit, account.Vendor, account.ID, req.Extension)
}

// checkAccountSecretByVendor checks account secret by vendor.
func (s *service) checkAccountSecretByVendor(kt *kit.Kit, vendor enumor.Vendor, accountID string,
	extension json.RawMessage) (interface{}, error) {

	switch vendor {
	case enumor.TCloud:
		return s.checkTCloudAccountSecret(kt, accountID, extension)
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

// checkTCloudAccountSecret checks tcloud account secret.
func (s *service) checkTCloudAccountSecret(kt *kit.Kit, accountID string, extension json.RawMessage) (
	*proto.TCloudAccountSecretCheckResult, error) {

	// 解析Extension
	ext := new(proto.TCloudAccountSecretExtension)
	if err := common.DecodeExtension(kt, extension, ext); err != nil {
		logs.Errorf("decode tcloud extension failed, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验参数
	if err := ext.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 调用HCService校验密钥并获取账号信息
	req := &corecloud.TCloudSecret{CloudSecretID: ext.CloudSecretID, CloudSecretKey: ext.CloudSecretKey}
	info, err := s.client.HCService().TCloud.Account.GetBySecret(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("check tcloud account secret failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 校验与账号云上主账号id是否匹配
	account, err := s.client.DataService().TCloud.Account.Get(kt.Ctx, kt.Header(), accountID)
	if err != nil {
		logs.Errorf("get tcloud account failed, err: %v, account id: %s, rid: %s", err, accountID, kt.Rid)
		return nil, err
	}
	if account.Extension.CloudMainAccountID != info.CloudMainAccountID {
		logs.Errorf("tcloud account secret mismatch, account id: %s, cur cloud main account id: %s, target cloud "+
			"account id: %s, rid: %s", accountID, account.Extension.CloudMainAccountID, info.CloudMainAccountID, kt.Rid)
		return nil, errf.Newf(errf.InvalidParameter, "tcloud account secret mismatch, account id: %s, cur cloud main "+
			"account id: %s, target cloud account id: %s", accountID, account.Extension.CloudMainAccountID,
			info.CloudMainAccountID)
	}

	// 返回结果
	return &proto.TCloudAccountSecretCheckResult{
		CloudMainAccountID: info.CloudMainAccountID,
		CloudSubAccountID:  info.CloudSubAccountID,
	}, nil
}

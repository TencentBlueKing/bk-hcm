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
	"fmt"

	"hcm/cmd/cloud-server/service/common"
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
	cvt "hcm/pkg/tools/converter"
)

// CreateBizAccountSecret creates a biz account secret.
func (s *service) CreateBizAccountSecret(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountSecretCreateReq)
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

	// 校验二级账号操作权限
	attribute := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Account, Action: meta.Update}, BizID: bizID}
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

	// 创建密钥
	secretID, err := s.createAccountSecretByVendor(cts.Kit, req, account.Vendor)
	if err != nil {
		return nil, err
	}

	return &proto.AccountSecretCreateResp{ID: secretID}, nil
}

// checkResourceSecretUniqueness checks if the account already has a resource management secret.
func (s *service) checkResourceSecretUniqueness(kt *kit.Kit, accountID string) error {
	// 查询该账号的资源管理密钥
	expr := tools.ExpressionAnd(
		tools.RuleEqual("account_id", accountID),
		tools.RuleEqual("type", enumor.ResourceSecretType),
	)

	result, err := s.client.DataService().Global.AccountSecret.ListAccountSecret(kt, &protocloud.AccountSecretListReq{
		Filter: expr,
		Page:   core.NewCountPage(),
	})
	if err != nil {
		logs.Errorf("list account secret failed, account_id: %s, err: %v, rid: %s", accountID, err, kt.Rid)
		return err
	}

	if result.Count > 0 {
		return errf.New(errf.Aborted, "account already has a resource management secret, cannot create duplicate")
	}

	return nil
}

// createAccountSecretByVendor creates account secret by vendor.
func (s *service) createAccountSecretByVendor(kt *kit.Kit, req *proto.AccountSecretCreateReq, vendor enumor.Vendor) (
	string, error) {

	// 资源管理密钥唯一性校验
	if req.Type == enumor.ResourceSecretType {
		if err := s.checkResourceSecretUniqueness(kt, req.AccountID); err != nil {
			return "", err
		}
	}

	// 密钥有效性校验
	checkResult, err := s.checkAccountSecretByVendor(kt, vendor, req.AccountID, req.Extension)
	if err != nil {
		return "", err
	}

	var secretID string
	switch vendor {
	case enumor.TCloud:
		secretID, err = s.createTCloudAccountSecret(kt, req, checkResult)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported vendor: %s", vendor)
	}

	return secretID, nil

}

// createTCloudAccountSecret creates tcloud account secret.
func (s *service) createTCloudAccountSecret(kt *kit.Kit, req *proto.AccountSecretCreateReq, checkResult interface{}) (
	string, error) {

	ext := new(proto.TCloudAccountSecretExtension)
	if err := common.DecodeExtension(kt, req.Extension, ext); err != nil {
		logs.Errorf("decode tcloud extension failed, err: %v, rid: %s", err, kt.Rid)
		return "", errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := ext.Validate(); err != nil {
		return "", errf.NewFromErr(errf.InvalidParameter, err)
	}
	tcheckResult, ok := checkResult.(*proto.TCloudAccountSecretCheckResult)
	if !ok {
		return "", errf.New(errf.Aborted, "invalid check result type")
	}
	// 创建密钥
	dsExt := &protocloud.TCloudAccountSecretExtension{
		TCloudAccountSecretExtension: coreas.TCloudAccountSecretExtension{
			CloudSecretID:      ext.CloudSecretID,
			CloudSecretKey:     ext.CloudSecretKey,
			CloudSubAccountID:  tcheckResult.CloudSubAccountID,
			CloudMainAccountID: tcheckResult.CloudMainAccountID,
		},
	}
	createReq := &protocloud.AccountSecretBatchCreateReq[coreas.TCloudAccountSecretExtension]{
		AccountSecrets: []protocloud.AccountSecretCreate[coreas.TCloudAccountSecretExtension]{
			{
				AccountID: req.AccountID,
				Type:      req.Type,
				Status:    enumor.NormalSecretStatus,
				Extension: &dsExt.TCloudAccountSecretExtension,
			},
		},
	}
	result, err := s.client.DataService().TCloud.AccountSecret.BatchCreateAccountSecret(kt, createReq)
	if err != nil {
		logs.Errorf("create account secret failed, account_id: %s, type: %s, err: %v, rid: %s", req.AccountID, req.Type,
			err, kt.Rid)
		return "", err
	}
	if len(result.IDs) == 0 {
		return "", errf.New(errf.Aborted, "create account secret failed, no id returned")
	}

	// 如果是资源管理密钥，更新账号 extension
	if req.Type != enumor.ResourceSecretType {
		return result.IDs[0], nil
	}
	extReq := &protocloud.TCloudAccountExtensionUpdateReq{
		CloudSecretID:      cvt.ValToPtr(ext.CloudSecretID),
		CloudSecretKey:     cvt.ValToPtr(ext.CloudSecretKey),
		CloudSubAccountID:  cvt.ValToPtr(tcheckResult.CloudSubAccountID),
		CloudMainAccountID: tcheckResult.CloudMainAccountID,
	}
	if err = s.updateTCloudAccountExt(kt, req.AccountID, extReq); err != nil {
		logs.Errorf("update tcloud account extension failed, account_id: %s, err: %v, rid: %s", req.AccountID, err,
			kt.Rid)
		return "", err
	}

	return result.IDs[0], nil
}

func (s *service) updateTCloudAccountExt(kt *kit.Kit, accountID string,
	ext *protocloud.TCloudAccountExtensionUpdateReq) error {

	updateReq := &protocloud.AccountUpdateReq[protocloud.TCloudAccountExtensionUpdateReq]{Extension: ext}
	if _, err := s.client.DataService().TCloud.Account.Update(kt.Ctx, kt.Header(), accountID, updateReq); err != nil {
		logs.Errorf("update tcloud account extension failed, account_id: %s, err: %v, rid: %s", accountID, err, kt.Rid)
		return err
	}

	return nil
}

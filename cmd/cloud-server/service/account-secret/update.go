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

// UpdateBizAccountSecret updates a biz account secret.
func (s *service) UpdateBizAccountSecret(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountSecretUpdateReq)
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

	// 查询当前密钥信息
	secretID := cts.PathParameter("id").String()
	if secretID == "" {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}
	currentSecret, err := s.getAccountSecretByID(cts.Kit, secretID)
	if err != nil {
		logs.Errorf("get account secret failed, secret_id: %s, err: %v, rid: %s", secretID, err, cts.Kit.Rid)
		return nil, err
	}

	// 查询账号基本信息
	listReq := &core.ListReq{
		Filter: tools.EqualExpression("id", currentSecret.AccountID),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"id", "bk_biz_id", "vendor"},
	}
	resp, err := s.client.DataService().Global.Account.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("get account basic info failed, account_id: %s, err: %v, rid: %s", currentSecret.AccountID, err,
			cts.Kit.Rid)
		return nil, err
	}
	if len(resp.Details) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "account %s not found", currentSecret.AccountID)
	}
	account := resp.Details[0]

	if account.BkBizID != bizID {
		return nil, errf.Newf(errf.PermissionDenied, "secret %s does not belong to business %d", secretID, bizID)
	}

	// 记录更新审计
	updateFields, err := cvt.StructToMap(req)
	if err != nil {
		logs.Errorf("convert request to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err = s.audit.ResUpdateAudit(cts.Kit, enumor.AccountSecretAuditResType, secretID, updateFields); err != nil {
		logs.Errorf("create update audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := s.updateAccountSecretByType(cts.Kit, account.Vendor, currentSecret, req); err != nil {
		logs.Errorf("update account secret by type failed, secret_id: %s, err: %v, rid: %s", secretID, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (s *service) updateAccountSecretByType(kt *kit.Kit, vendor enumor.Vendor, secret *coreas.BaseAccountSecret,
	req *proto.AccountSecretUpdateReq) error {

	finalType := secret.Type
	if req.Type != nil {
		finalType = cvt.PtrToVal(req.Type)
	}

	// non-resource type or resource type → resource type
	if finalType == enumor.ResourceSecretType {
		if err := s.updateSecretToRes(kt, vendor, secret, req); err != nil {
			return err
		}
		return nil
	}

	// resource type → non-resource type
	if secret.Type == enumor.ResourceSecretType {
		if err := s.updateSecretResToNonRes(kt, vendor, secret, req); err != nil {
			return err
		}
		return nil
	}

	// non-resource type -> non-resource type
	return s.updateSecretNonResToNonRes(kt, vendor, secret, req)
}

func (s *service) updateSecretToRes(kt *kit.Kit, vendor enumor.Vendor, secret *coreas.BaseAccountSecret,
	req *proto.AccountSecretUpdateReq) error {

	// 如果密钥原来不是资源管理类型，需要检验资源密钥的唯一性
	if secret.Type != enumor.ResourceSecretType {
		if err := s.checkResourceSecretUniqueness(kt, secret.AccountID); err != nil {
			logs.Errorf("check resource secret uniqueness failed, err: %v, account_id: %s, rid: %s", err,
				secret.AccountID, kt.Rid)
			return err
		}
	}

	switch vendor {
	case enumor.TCloud:
		return s.updateTCloudSecretToRes(kt, secret, req)
	default:
		return fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

func (s *service) updateTCloudSecretToRes(kt *kit.Kit, secret *coreas.BaseAccountSecret,
	req *proto.AccountSecretUpdateReq) error {

	// 更新密钥
	secretExt, err := s.updateTCloudSecret(kt, secret, req)
	if err != nil {
		logs.Errorf("update tcloud secret failed, err: %v, secret_id: %s, rid: %s", err, secret.ID, kt.Rid)
		return err
	}

	// 更新账号extension
	extReq := &protocloud.TCloudAccountExtensionUpdateReq{
		CloudSecretID:      cvt.ValToPtr(secretExt.CloudSecretID),
		CloudSecretKey:     cvt.ValToPtr(secretExt.CloudSecretKey),
		CloudSubAccountID:  cvt.ValToPtr(secretExt.CloudSubAccountID),
		CloudMainAccountID: secretExt.CloudMainAccountID,
	}
	if err = s.updateTCloudAccountExt(kt, secret.AccountID, extReq); err != nil {
		logs.Errorf("update account extension failed, err: %v, account_id: %s, rid: %s", err, secret.AccountID, kt.Rid)
		return err
	}

	return nil
}

func (s *service) updateSecretResToNonRes(kt *kit.Kit, vendor enumor.Vendor, secret *coreas.BaseAccountSecret,
	req *proto.AccountSecretUpdateReq) error {

	switch vendor {
	case enumor.TCloud:
		return s.updateTCloudSecretResToNonRes(kt, secret, req)
	default:
		return fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

func (s *service) updateTCloudSecretResToNonRes(kt *kit.Kit, secret *coreas.BaseAccountSecret,
	req *proto.AccountSecretUpdateReq) error {

	// 更新密钥信息
	if _, err := s.updateTCloudSecret(kt, secret, req); err != nil {
		logs.Errorf("update tcloud secret failed, err: %v, secret_id: %s, rid: %s", err, secret.ID, kt.Rid)
		return err
	}

	// 清空账号extension
	if err := s.clearTCloudAccountExtension(kt, secret.AccountID); err != nil {
		logs.Errorf("clear account extension failed, account_id: %s, err: %v, rid: %s", secret.AccountID, err, kt.Rid)
		return err
	}

	return nil
}

// clearTCloudAccountExtension clears the tcloud account extension.
func (s *service) clearTCloudAccountExtension(kt *kit.Kit, accountID string) error {
	updateReq := &protocloud.AccountUpdateReq[protocloud.TCloudAccountExtensionUpdateReq]{
		Extension: &protocloud.TCloudAccountExtensionUpdateReq{
			CloudSubAccountID: cvt.ValToPtr(""),
			CloudSecretID:     cvt.ValToPtr(""),
			CloudSecretKey:    cvt.ValToPtr(""),
		},
	}
	if _, err := s.client.DataService().TCloud.Account.Update(kt.Ctx, kt.Header(), accountID, updateReq); err != nil {
		logs.Errorf("clear tcloud account extension failed, err: %v, account_id: %s, rid: %s", err, accountID, kt.Rid)
		return err
	}

	return nil
}

func (s *service) updateSecretNonResToNonRes(kt *kit.Kit, vendor enumor.Vendor, secret *coreas.BaseAccountSecret,
	req *proto.AccountSecretUpdateReq) error {

	switch vendor {
	case enumor.TCloud:
		if _, err := s.updateTCloudSecret(kt, secret, req); err != nil {
			logs.Errorf("update tcloud secret failed, err: %v, secret_id: %s, rid: %s", err, secret.ID, kt.Rid)
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

func (s *service) updateTCloudSecret(kt *kit.Kit, secret *coreas.BaseAccountSecret,
	req *proto.AccountSecretUpdateReq) (*coreas.TCloudAccountSecretExtension, error) {

	secretExt, err := s.getTCloudSecretExt(kt, secret.ID, req)
	if err != nil {
		logs.Errorf("get tcloud secret extension failed, err: %v, secret_id: %s, rid: %s", err, secret.ID, kt.Rid)
		return nil, err
	}
	if secretExt == nil {
		return nil, errf.Newf(errf.RecordNotFound, "account secret %s not found", secret.ID)
	}

	secretType := secret.Type
	if req.Type != nil {
		secretType = cvt.PtrToVal(req.Type)
	}

	updateReq := &protocloud.AccountSecretBatchUpdateReq[coreas.TCloudAccountSecretExtension]{
		AccountSecrets: []protocloud.AccountSecretUpdate[coreas.TCloudAccountSecretExtension]{
			{ID: secret.ID, Type: &secretType, Extension: secretExt},
		},
	}
	if err = s.client.DataService().TCloud.AccountSecret.BatchUpdateAccountSecret(kt, updateReq); err != nil {
		logs.Errorf("update account secret failed, secret_id: %s, err: %v, rid: %s", secret.ID, err, kt.Rid)
		return nil, err
	}

	return secretExt, nil
}

func (s *service) getTCloudSecretExt(kt *kit.Kit, secretID string, req *proto.AccountSecretUpdateReq) (
	*coreas.TCloudAccountSecretExtension, error) {

	secret, err := s.getTCloudAccountSecretByID(kt, secretID)
	if err != nil {
		logs.Errorf("list account secret failed, err: %v, secret_id: %s, rid: %s", err, secretID, kt.Rid)
		return nil, err
	}
	if secret == nil {
		return nil, errf.Newf(errf.RecordNotFound, "account secret %s not found", secretID)
	}

	if req.Extension != nil {
		checkResult, err := s.checkTCloudAccountSecret(kt, secret.AccountID, cvt.PtrToVal(req.Extension))
		if err != nil {
			logs.Errorf("check account secret failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		extParams := new(proto.TCloudAccountSecretExtension)
		if err := common.DecodeExtension(kt, cvt.PtrToVal(req.Extension), extParams); err != nil {
			logs.Errorf("decode tcloud extension failed, err: %v, rid: %s", err, kt.Rid)
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		return &coreas.TCloudAccountSecretExtension{
			CloudSecretID:      extParams.CloudSecretID,
			CloudSecretKey:     extParams.CloudSecretKey,
			CloudSubAccountID:  checkResult.CloudSubAccountID,
			CloudMainAccountID: checkResult.CloudMainAccountID,
		}, nil
	}

	return secret.Extension, nil
}

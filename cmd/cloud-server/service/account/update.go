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

	"hcm/cmd/cloud-server/service/common"
	proto "hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// UpdateAccount ...
func (a *accountSvc) UpdateAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	// 根据accountID拿到账户类型然后才进行请求参数校验，因为资源账号需要额外的校验
	accountID := cts.PathParameter("account_id").String()
	listReq := &dataproto.AccountListReq{
		Filter: tools.EqualExpression("id", accountID),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := a.client.DataService().Global.Account.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("list account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if len(result.Details) == 0 {
		logs.Errorf("account: %s not found, rid: %s", accountID, cts.Kit.Rid)
		return nil, errf.Newf(errf.RecordNotFound, "account: %s not found", accountID)
	}

	accountInfo := result.Details[0]
	if err := req.Validate(accountInfo); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 内置账号不允许通过UpdateAccount接口修改
	if accountInfo.Vendor == enumor.Other {
		return nil, fmt.Errorf("built-in account is not allowed to be updated")
	}

	action := meta.Update
	if req.RecycleReserveTime != 0 {
		action = meta.UpdateRRT
	}
	// 校验用户有该账号的更新权限
	if err := a.checkPermission(cts, action, accountID); err != nil {
		return nil, err
	}

	// 查询该账号对应的Vendor
	baseInfo, err := a.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.AccountCloudResType,
		accountID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// create update audit.
	updateFields, err := converter.StructToMap(req)
	if err != nil {
		logs.Errorf("convert request to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err = a.audit.ResUpdateAudit(cts.Kit, enumor.AccountAuditResType, accountID, updateFields); err != nil {
		logs.Errorf("create update audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 更新账号与业务关系
	if len(req.UsageBizIDs) > 0 {
		_, err = a.client.DataService().Global.Account.UpdateBizRel(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			accountID,
			&dataproto.AccountBizRelUpdateReq{
				UsageBizIDs: req.UsageBizIDs,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	switch baseInfo.Vendor {
	case enumor.TCloud:
		return a.updateForTCloud(cts, req, accountID)
	case enumor.Aws:
		return a.updateForAws(cts, req, accountID)
	case enumor.HuaWei:
		return a.updateForHuaWei(cts, req, accountID)
	case enumor.Gcp:
		return a.updateForGcp(cts, req, accountID)
	case enumor.Azure:
		return a.updateForAzure(cts, req, accountID)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", baseInfo.Vendor))
	}
}

func (a *accountSvc) updateForTCloud(
	cts *rest.Contexts, req *proto.AccountUpdateReq, accountID string,
) (
	interface{}, error,
) {
	// 解析Extension
	var (
		extension *proto.TCloudAccountExtensionUpdateReq
		err       error
	)
	if req.Extension != nil {
		extension, err = a.parseAndCheckTCloudExtensionByID(cts, accountID, req.Extension)
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}

	var shouldUpdatedExtension *dataproto.TCloudAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.TCloudAccountExtensionUpdateReq{
			CloudSubAccountID: extension.CloudSubAccountID,
			CloudSecretID:     &extension.CloudSecretID,
			CloudSecretKey:    &extension.CloudSecretKey,
		}
	}

	// 更新
	_, err = a.client.DataService().TCloud.Account.Update(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		accountID,
		&dataproto.AccountUpdateReq[dataproto.TCloudAccountExtensionUpdateReq]{
			Name:               req.Name,
			Managers:           req.Managers,
			RecycleReserveTime: req.RecycleReserveTime,
			Memo:               req.Memo,
			BkBizID:            req.BkBizID,
			Extension:          shouldUpdatedExtension,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil

}

func (a *accountSvc) updateForAws(
	cts *rest.Contexts, req *proto.AccountUpdateReq, accountID string,
) (
	interface{}, error,
) {
	// 解析Extension
	var (
		extension *proto.AwsAccountExtensionUpdateReq
		err       error
	)
	if req.Extension != nil {
		extension, err = a.parseAndCheckAwsExtensionByID(cts, accountID, req.Extension)
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}

	var shouldUpdatedExtension *dataproto.AwsAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.AwsAccountExtensionUpdateReq{
			CloudIamUsername: extension.CloudIamUsername,
			CloudSecretID:    &extension.CloudSecretID,
			CloudSecretKey:   &extension.CloudSecretKey,
		}
	}

	// 更新
	_, err = a.client.DataService().Aws.Account.Update(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		accountID,
		&dataproto.AccountUpdateReq[dataproto.AwsAccountExtensionUpdateReq]{
			Name:               req.Name,
			Managers:           req.Managers,
			Memo:               req.Memo,
			RecycleReserveTime: req.RecycleReserveTime,
			BkBizID:            req.BkBizID,
			Extension:          shouldUpdatedExtension,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil

}

func (a *accountSvc) updateForHuaWei(
	cts *rest.Contexts, req *proto.AccountUpdateReq, accountID string,
) (
	interface{}, error,
) {
	// 解析Extension
	var (
		extension *proto.HuaWeiAccountExtensionUpdateReq
		err       error
	)
	if req.Extension != nil {
		extension, err = a.parseAndCheckHuaWeiExtensionByID(cts, accountID, req.Extension)
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}

	var shouldUpdatedExtension *dataproto.HuaWeiAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.HuaWeiAccountExtensionUpdateReq{
			CloudSubAccountName: extension.CloudSubAccountName,
			CloudIamUserID:      extension.CloudIamUserID,
			CloudIamUsername:    extension.CloudIamUsername,
			CloudSecretID:       &extension.CloudSecretID,
			CloudSecretKey:      &extension.CloudSecretKey,
		}
	}

	// 更新
	_, err = a.client.DataService().HuaWei.Account.Update(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		accountID,
		&dataproto.AccountUpdateReq[dataproto.HuaWeiAccountExtensionUpdateReq]{
			Name:               req.Name,
			Managers:           req.Managers,
			Memo:               req.Memo,
			RecycleReserveTime: req.RecycleReserveTime,
			BkBizID:            req.BkBizID,
			Extension:          shouldUpdatedExtension,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil

}

func (a *accountSvc) updateForGcp(
	cts *rest.Contexts, req *proto.AccountUpdateReq, accountID string,
) (
	interface{}, error,
) {
	// 解析Extension
	var (
		extension *proto.GcpAccountExtensionUpdateReq
		err       error
	)
	if req.Extension != nil {
		extension, err = a.parseAndCheckGcpExtensionByID(cts, accountID, req.Extension)
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}

	var shouldUpdatedExtension *dataproto.GcpAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.GcpAccountExtensionUpdateReq{
			CloudProjectName:        extension.CloudProjectName,
			CloudServiceAccountID:   &extension.CloudServiceAccountID,
			CloudServiceAccountName: &extension.CloudServiceAccountName,
			CloudServiceSecretID:    &extension.CloudServiceSecretID,
			CloudServiceSecretKey:   &extension.CloudServiceSecretKey,
		}
	}

	// 更新
	_, err = a.client.DataService().Gcp.Account.Update(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		accountID,
		&dataproto.AccountUpdateReq[dataproto.GcpAccountExtensionUpdateReq]{
			Name:               req.Name,
			Managers:           req.Managers,
			Memo:               req.Memo,
			RecycleReserveTime: req.RecycleReserveTime,
			BkBizID:            req.BkBizID,
			Extension:          shouldUpdatedExtension,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil

}

func (a *accountSvc) updateForAzure(
	cts *rest.Contexts, req *proto.AccountUpdateReq, accountID string,
) (
	interface{}, error,
) {
	// 解析Extension
	var (
		extension *proto.AzureAccountExtensionUpdateReq
		err       error
	)
	if req.Extension != nil {
		extension, err = a.parseAndCheckAzureExtensionByID(cts, accountID, req.Extension)
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}

	var shouldUpdatedExtension *dataproto.AzureAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.AzureAccountExtensionUpdateReq{
			CloudTenantID:         extension.CloudTenantID,
			CloudSubscriptionName: extension.CloudSubscriptionName,
			CloudApplicationID:    &extension.CloudApplicationID,
			CloudApplicationName:  &extension.CloudApplicationName,
			CloudClientSecretKey:  &extension.CloudClientSecretKey,
		}
	}

	// 更新
	_, err = a.client.DataService().Azure.Account.Update(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		accountID,
		&dataproto.AccountUpdateReq[dataproto.AzureAccountExtensionUpdateReq]{
			Name:               req.Name,
			Managers:           req.Managers,
			Memo:               req.Memo,
			RecycleReserveTime: req.RecycleReserveTime,
			BkBizID:            req.BkBizID,
			Extension:          shouldUpdatedExtension,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil

}

// UpdateBuiltInAccount ...
func (a *accountSvc) UpdateBuiltInAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AccountUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	listReq := &dataproto.AccountListReq{
		Filter: tools.EqualExpression("vendor", enumor.Other),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := a.client.DataService().Global.Account.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("list built-in account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if len(result.Details) == 0 {
		logs.Errorf("built-in account not found, rid: %s", cts.Kit.Rid)
		return nil, errf.Newf(errf.RecordNotFound, "built-in account not found")
	}
	if err := req.Validate(result.Details[0]); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 获取内置账号ID
	accountID := result.Details[0].ID

	action := meta.Update
	if req.RecycleReserveTime != 0 {
		action = meta.UpdateRRT
	}
	// 校验用户有该账号的更新权限
	if err := a.checkPermission(cts, action, accountID); err != nil {
		return nil, err
	}

	// create update audit.
	updateFields, err := converter.StructToMap(req)
	if err != nil {
		logs.Errorf("convert request to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err = a.audit.ResUpdateAudit(cts.Kit, enumor.AccountAuditResType, accountID, updateFields); err != nil {
		logs.Errorf("create update audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 更新账号与业务关系
	if len(req.UsageBizIDs) > 0 {
		_, err = a.client.DataService().Global.Account.UpdateBizRel(
			cts.Kit.Ctx,
			cts.Kit.Header(),
			accountID,
			&dataproto.AccountBizRelUpdateReq{
				UsageBizIDs: req.UsageBizIDs,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return a.updateForOther(cts, req, accountID)
}

func (a *accountSvc) updateForOther(
	cts *rest.Contexts, req *proto.AccountUpdateReq, accountID string,
) (
	interface{}, error,
) {
	// 解析Extension
	var (
		extension *proto.OtherAccountExtensionUpdateReq
		err       error
	)
	if req.Extension != nil {
		// 解析Extension
		extension = new(proto.OtherAccountExtensionUpdateReq)
		if err = common.DecodeExtension(cts.Kit, req.Extension, extension); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}

	var shouldUpdatedExtension *dataproto.OtherAccountExtensionUpdateReq = nil
	if req.Extension != nil {
		shouldUpdatedExtension = &dataproto.OtherAccountExtensionUpdateReq{
			CloudID:     extension.CloudID,
			CloudSecKey: extension.CloudSecKey,
		}
	}

	// 更新
	_, err = a.client.DataService().Other.Account.Update(
		cts.Kit,
		accountID,
		&dataproto.AccountUpdateReq[dataproto.OtherAccountExtensionUpdateReq]{
			Name:               req.Name,
			Managers:           req.Managers,
			Memo:               req.Memo,
			RecycleReserveTime: req.RecycleReserveTime,
			BkBizID:            req.BkBizID,
			Extension:          shouldUpdatedExtension,
		},
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, nil
}

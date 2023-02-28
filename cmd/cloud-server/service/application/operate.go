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

package application

import (
	"fmt"

	accountsvc "hcm/cmd/cloud-server/service/account"
	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service"
	dataprotocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/json"
)

func (a *applicationSvc) updateStatus(cts *rest.Contexts, applicationID string, status enumor.ApplicationStatus) error {
	_, err := a.client.DataService().Global.Application.Update(
		cts.Kit.Ctx, cts.Kit.Header(),
		applicationID,
		&dataproto.ApplicationUpdateReq{Status: status},
	)
	return err
}

// Cancel ...
func (a *applicationSvc) Cancel(cts *rest.Contexts) (interface{}, error) {
	applicationID := cts.PathParameter("application_id").String()
	application, err := a.client.DataService().Global.Application.Get(cts.Kit.Ctx, cts.Kit.Header(), applicationID)
	if err != nil {
		return nil, err
	}

	// 只能查询自己的申请单
	if application.Applicant != cts.Kit.User {
		return nil, errf.NewFromErr(
			errf.PermissionDenied, fmt.Errorf("you can not operate other people's application"),
		)
	}

	// 根据SN调用ITSM接口撤销单据
	err = a.esbClient.Itsm().WithdrawTicket(cts.Kit.Ctx, application.SN, cts.Kit.User)
	if err != nil {
		return nil, fmt.Errorf("call itsm cancel ticket api failed, err: %v", err)
	}

	// 更新状态
	err = a.updateStatus(cts, applicationID, enumor.Cancelled)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (a *applicationSvc) getApplicationBySN(cts *rest.Contexts, sn string) (*dataproto.ApplicationResp, error) {
	// 构造过滤条件，只能查询自己的单据
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "sn", Op: filter.Equal.Factory(), Value: sn},
		},
	}
	// 查询
	resp, err := a.client.DataService().Global.Application.List(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.ApplicationListReq{
			Filter: reqFilter,
			Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
		},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("not found application by sn(%s)", sn)
	}

	return resp.Details[0], nil
}

func (a *applicationSvc) convertToStatus(ticketStatus string, approveResult bool) enumor.ApplicationStatus {
	if ticketStatus == "FINISHED" {
		if approveResult {
			return enumor.Pass
		}
		return enumor.Rejected
	}

	if ticketStatus == "TERMINATED" || ticketStatus == "REVOKED" {
		return enumor.Cancelled
	}

	return enumor.Pending
}

// Approve ...
func (a *applicationSvc) Approve(cts *rest.Contexts) (interface{}, error) {
	// Note: 由于该接口时给ITSM回调的，一般没什么反馈，这将任何错误到记录到日志里
	_, err := a.approve(cts)
	if err != nil {
		logs.Errorf("itsm approve callback failed, error: %v", err)
	}
	return nil, err
}

func (a *applicationSvc) approve(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.ItsmApproveResult)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验ITSM Token
	ok, err := a.esbClient.Itsm().VerifyToken(cts.Kit.Ctx, req.Token)
	if err != nil {
		return nil, fmt.Errorf("call itsm verify token api failed, err: %v", err)
	}
	if !ok {
		return nil, errf.NewFromErr(
			errf.PermissionDenied, fmt.Errorf("verify of token not paas"),
		)
	}

	// 查询单据
	application, err := a.getApplicationBySN(cts, req.SN)
	if err != nil {
		return nil, err
	}

	// 将ITSM单据状态转为hcm定义的单据状态
	status := a.convertToStatus(req.CurrentStatus, *req.ApproveResult)
	nextStatus := status
	// 对于审批通过，则下个状态为交付中，其他状态则保持原样
	if status == enumor.Pass {
		nextStatus = enumor.Delivering
	}
	// 更新状态
	err = a.updateStatus(cts, application.ID, nextStatus)
	if err != nil {
		return nil, err
	}

	// TODO: 需要引入异步任务，这里先暂时同步执行
	if status == enumor.Pass {
		err = a.addAccount(cts, application)
		deliverStatus := enumor.Completed
		if err != nil {
			logs.Errorf(
				"execute application[id=%s] of add account failed, err: %s, rid: %s",
				application.ID, err, cts.Kit.Rid,
			)
			deliverStatus = enumor.DeliverError
		}
		err = a.updateStatus(cts, application.ID, deliverStatus)
		if err != nil {
			return nil, err
		}

	}

	return nil, nil
}

func (a *applicationSvc) addAccount(cts *rest.Contexts, application *dataproto.ApplicationResp) error {
	// 将执行人设置为申请人
	cts.Kit.User = application.Applicant

	// 解析申请单内容
	req := new(proto.AccountAddReq)
	err := json.UnmarshalFromString(application.Content, req)
	if err != nil {
		return fmt.Errorf("json unmarshal content error: %w", err)
	}

	// 解密密钥
	secretKeyField := accountsvc.VendorSecretKeyFieldMap[req.Vendor]
	secretKey, err := a.cipher.DecryptFromBase64(req.Extension[secretKeyField])
	if err != nil {
		return fmt.Errorf("decrypt secret key failed, err: %w", err)
	}
	req.Extension[secretKeyField] = secretKey

	// 再次检查数据正确性
	err = a.checkForAddAccount(cts, req)
	if err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 执行创建账号
	switch req.Vendor {
	case enumor.TCloud:
		return a.createForTCloud(cts, req)
	case enumor.Aws:
		return a.createForAws(cts, req)
	case enumor.HuaWei:
		return a.createForHuaWei(cts, req)
	case enumor.Gcp:
		return a.createForGcp(cts, req)
	case enumor.Azure:
		return a.createForAzure(cts, req)
	}

	return nil
}

func (a *applicationSvc) createForTCloud(cts *rest.Contexts, req *proto.AccountAddReq) error {
	_, err := a.client.DataService().TCloud.Account.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataprotocloud.AccountCreateReq[dataprotocloud.TCloudAccountExtensionCreateReq]{
			Name:          req.Name,
			Managers:      req.Managers,
			DepartmentIDs: req.DepartmentIDs,
			Type:          req.Type,
			Site:          req.Site,
			Memo:          req.Memo,
			BkBizIDs:      req.BkBizIDs,
			Extension: &dataprotocloud.TCloudAccountExtensionCreateReq{
				CloudMainAccountID: req.Extension["cloud_main_account_id"],
				CloudSubAccountID:  req.Extension["cloud_sub_account_id"],
				CloudSecretID:      req.Extension["cloud_secret_id"],
				CloudSecretKey:     req.Extension["cloud_secret_key"],
			},
		},
	)
	return err
}

func (a *applicationSvc) createForAws(cts *rest.Contexts, req *proto.AccountAddReq) error {
	_, err := a.client.DataService().Aws.Account.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataprotocloud.AccountCreateReq[dataprotocloud.AwsAccountExtensionCreateReq]{
			Name:          req.Name,
			Managers:      req.Managers,
			DepartmentIDs: req.DepartmentIDs,
			Type:          req.Type,
			Site:          req.Site,
			Memo:          req.Memo,
			BkBizIDs:      req.BkBizIDs,
			Extension: &dataprotocloud.AwsAccountExtensionCreateReq{
				CloudAccountID:   req.Extension["cloud_account_id"],
				CloudIamUsername: req.Extension["cloud_iam_username"],
				CloudSecretID:    req.Extension["cloud_secret_id"],
				CloudSecretKey:   req.Extension["cloud_secret_key"],
			},
		},
	)
	return err
}

func (a *applicationSvc) createForHuaWei(cts *rest.Contexts, req *proto.AccountAddReq) error {
	_, err := a.client.DataService().HuaWei.Account.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataprotocloud.AccountCreateReq[dataprotocloud.HuaWeiAccountExtensionCreateReq]{
			Name:          req.Name,
			Managers:      req.Managers,
			DepartmentIDs: req.DepartmentIDs,
			Type:          req.Type,
			Site:          req.Site,
			Memo:          req.Memo,
			BkBizIDs:      req.BkBizIDs,
			Extension: &dataprotocloud.HuaWeiAccountExtensionCreateReq{
				CloudMainAccountName: req.Extension["cloud_main_account_name"],
				CloudSubAccountID:    req.Extension["cloud_sub_account_id"],
				CloudSubAccountName:  req.Extension["cloud_sub_account_name"],
				CloudSecretID:        req.Extension["cloud_secret_id"],
				CloudSecretKey:       req.Extension["cloud_secret_key"],
				CloudIamUserID:       req.Extension["cloud_iam_user_id"],
				CloudIamUsername:     req.Extension["cloud_iam_username"],
			},
		},
	)
	return err
}

func (a *applicationSvc) createForGcp(cts *rest.Contexts, req *proto.AccountAddReq) error {
	_, err := a.client.DataService().Gcp.Account.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataprotocloud.AccountCreateReq[dataprotocloud.GcpAccountExtensionCreateReq]{
			Name:          req.Name,
			Managers:      req.Managers,
			DepartmentIDs: req.DepartmentIDs,
			Type:          req.Type,
			Site:          req.Site,
			Memo:          req.Memo,
			BkBizIDs:      req.BkBizIDs,
			Extension: &dataprotocloud.GcpAccountExtensionCreateReq{
				CloudProjectID:          req.Extension["cloud_project_id"],
				CloudProjectName:        req.Extension["cloud_project_name"],
				CloudServiceAccountID:   req.Extension["cloud_service_account_id"],
				CloudServiceAccountName: req.Extension["cloud_service_account_name"],
				CloudServiceSecretID:    req.Extension["cloud_service_secret_id"],
				CloudServiceSecretKey:   req.Extension["cloud_service_secret_key"],
			},
		},
	)
	return err
}

func (a *applicationSvc) createForAzure(cts *rest.Contexts, req *proto.AccountAddReq) error {
	_, err := a.client.DataService().Azure.Account.Create(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataprotocloud.AccountCreateReq[dataprotocloud.AzureAccountExtensionCreateReq]{
			Name:          req.Name,
			Managers:      req.Managers,
			DepartmentIDs: req.DepartmentIDs,
			Type:          req.Type,
			Site:          req.Site,
			Memo:          req.Memo,
			BkBizIDs:      req.BkBizIDs,
			Extension: &dataprotocloud.AzureAccountExtensionCreateReq{
				CloudTenantID:         req.Extension["cloud_tenant_id"],
				CloudSubscriptionID:   req.Extension["cloud_subscription_id"],
				CloudSubscriptionName: req.Extension["cloud_subscription_name"],
				CloudApplicationID:    req.Extension["cloud_application_id"],
				CloudApplicationName:  req.Extension["cloud_application_name"],
				CloudClientSecretID:   req.Extension["cloud_client_secret_id"],
				CloudClientSecretKey:  req.Extension["cloud_client_secret_key"],
			},
		},
	)
	return err
}

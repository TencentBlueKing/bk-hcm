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

	"hcm/cmd/cloud-server/service/application/handlers"
	accounthandler "hcm/cmd/cloud-server/service/application/handlers/account"
	awscvmhandler "hcm/cmd/cloud-server/service/application/handlers/cvm/aws"
	azurecvmhandler "hcm/cmd/cloud-server/service/application/handlers/cvm/azure"
	gcpcvmhandler "hcm/cmd/cloud-server/service/application/handlers/cvm/gcp"
	huaweicvmhandler "hcm/cmd/cloud-server/service/application/handlers/cvm/huawei"
	tcloudcvmhandler "hcm/cmd/cloud-server/service/application/handlers/cvm/tcloud"
	awsdiskhandler "hcm/cmd/cloud-server/service/application/handlers/disk/aws"
	azurediskhandler "hcm/cmd/cloud-server/service/application/handlers/disk/azure"
	gcpdiskhandler "hcm/cmd/cloud-server/service/application/handlers/disk/gcp"
	huaweidiskhandler "hcm/cmd/cloud-server/service/application/handlers/disk/huawei"
	tclouddiskhandler "hcm/cmd/cloud-server/service/application/handlers/disk/tcloud"
	lbtcloud "hcm/cmd/cloud-server/service/application/handlers/load_balancer/tcloud"
	createmainaccount "hcm/cmd/cloud-server/service/application/handlers/main-account/create-main-account"
	updatemainaccount "hcm/cmd/cloud-server/service/application/handlers/main-account/update-main-account"
	subaccount "hcm/cmd/cloud-server/service/application/handlers/sub-account"
	createsubaccount "hcm/cmd/cloud-server/service/application/handlers/sub-account/create-sub-account"
	deletesubaccount "hcm/cmd/cloud-server/service/application/handlers/sub-account/delete-sub-account"
	updatesubaccount "hcm/cmd/cloud-server/service/application/handlers/sub-account/update-sub-account"
	awsvpchandler "hcm/cmd/cloud-server/service/application/handlers/vpc/aws"
	azurevpchandler "hcm/cmd/cloud-server/service/application/handlers/vpc/azure"
	gcpvpchandler "hcm/cmd/cloud-server/service/application/handlers/vpc/gcp"
	huaweivpchandler "hcm/cmd/cloud-server/service/application/handlers/vpc/huawei"
	tcloudvpchandler "hcm/cmd/cloud-server/service/application/handlers/vpc/tcloud"
	proto "hcm/pkg/api/cloud-server/application"
	cscvm "hcm/pkg/api/cloud-server/cvm"
	csdisk "hcm/pkg/api/cloud-server/disk"
	csvpc "hcm/pkg/api/cloud-server/vpc"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

func decodeCommonReqAndValidate(cts *rest.Contexts) (*proto.CreateCommonReq, error) {
	bytes, err := cts.RequestBody()
	if err != nil {
		logs.Errorf("get request body failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	req := new(proto.CreateCommonReq)
	if err = json.Unmarshal(bytes, req); err != nil {
		logs.Errorf("unmarshal create common req failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("create common request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return req, nil
}

func decodeSysCommonReqAndValidate(cts *rest.Contexts) (*proto.SysCreateCommonReq, error) {
	bytes, err := cts.RequestBody()
	if err != nil {
		logs.Errorf("get request body failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	req := new(proto.SysCreateCommonReq)
	if err = json.Unmarshal(bytes, req); err != nil {
		logs.Errorf("unmarshal sys create common req failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err = req.Validate(); err != nil {
		logs.Errorf("sys create common request validate failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return req, nil
}

// create 创建申请单的通用逻辑
func (a *applicationSvc) create(cts *rest.Contexts, req *proto.CreateCommonReq,
	handler handlers.ApplicationHandler) (interface{}, error) {

	// 校验数据正确性
	if err := handler.CheckReq(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	// 预处理数据
	if err := handler.PrepareReq(); err != nil {
		return nil, err
	}
	// 查询审批流程服务ID
	applicationType := handler.GetType()

	// 调用ITSM创建单据
	sn, err := a.createItsmTicket(cts, handler, applicationType)
	if err != nil {
		return nil, fmt.Errorf("call itsm create ticket api failed, err: %w", err)
	}

	result, err := a.createApplication(cts, req, handler, sn, applicationType)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// createApplicationRequest ...
func (a *applicationSvc) createApplication(cts *rest.Contexts, req *proto.CreateCommonReq,
	handler handlers.ApplicationHandler, sn string, applicationType enumor.ApplicationType) (
	*core.CreateResult, error) {

	// 调用DB创建单据
	content, err := json.MarshalToString(handler.GenerateApplicationContent())
	if err != nil {
		return nil, errf.NewFromErr(
			errf.InvalidParameter,
			fmt.Errorf("json marshal request data failed, err: %w", err),
		)
	}

	// 主机、硬盘、VPC、负载均衡需要记录业务ID
	var bkBizIDs = make([]int64, 0)
	if applicationType == enumor.CreateCvm || applicationType == enumor.CreateDisk ||
		applicationType == enumor.CreateVpc || applicationType == enumor.CreateLoadBalancer ||
		applicationType == enumor.AddAccount || applicationType == enumor.OperateSubAccount {
		bkBizIDs = handler.GetBkBizIDs()
	}
	return a.client.DataService().Global.Application.CreateApplication(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&dataproto.ApplicationCreateReq{
			SN:             sn,
			Source:         enumor.ApplicationSourceITSM,
			Type:           applicationType,
			Status:         enumor.Pending,
			BkBizIDs:       bkBizIDs,
			Applicant:      cts.Kit.User,
			Content:        content,
			DeliveryDetail: "{}",
			Memo:           req.Remark,
		},
	)
}

// createItsmTicket 调用ITSM创建单据
func (a *applicationSvc) createItsmTicket(cts *rest.Contexts, handler handlers.ApplicationHandler,
	applicationType enumor.ApplicationType) (string, error) {
	serviceID, managers, err := a.getApprovalProcessInfo(cts, applicationType)
	if err != nil {
		return "", fmt.Errorf("get approval process service id and managers failed, err: %v", err)
	}

	// 生成ITSM的回调地址
	callbackUrl := a.getCallbackUrl()

	// 渲染ITSM单据标题
	itsmTitle, err := handler.RenderItsmTitle()
	if err != nil {
		return "", fmt.Errorf("render itsm ticket title error: %w", err)
	}

	// 渲染ITSM单据申请内容
	itsmForm, err := handler.RenderItsmForm()
	if err != nil {
		return "", fmt.Errorf("render itsm ticket form error: %w", err)
	}

	// 获取ITSM单据涉及到的各个节点审批人
	approvers := handler.GetItsmApprover(managers)

	sn, err := a.itsmCli.CreateTicket(
		cts.Kit,
		&itsm.CreateTicketParams{
			ServiceID:      serviceID,
			Creator:        cts.Kit.User,
			CallbackURL:    callbackUrl,
			Title:          itsmTitle,
			ContentDisplay: itsmForm,
			// ITSM流程里使用变量引用的方式设置各个节点审批人
			VariableApprovers: approvers,
		},
	)
	if err != nil {
		return "", fmt.Errorf("call itsm create ticket api failed, err: %w", err)
	}
	return sn, nil
}

func parseReqFromRequestBody[T any](cts *rest.Contexts) (*T, error) {
	req := new(T)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	return req, nil
}

// CreateForAddAccount ...
func (a *applicationSvc) CreateForAddAccount(cts *rest.Contexts) (interface{}, error) {
	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Account, Action: meta.Import}}
	err := a.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	commReq, err := decodeCommonReqAndValidate(cts)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req, err := parseReqFromRequestBody[proto.AccountAddReq](cts)
	if err != nil {
		return nil, err
	}
	handler := accounthandler.NewApplicationOfAddAccount(a.getHandlerOption(cts), a.authorizer, req)

	return a.create(cts, commReq, handler)
}

// CreateBizForAddAccount create biz for add account
func (a *applicationSvc) CreateBizForAddAccount(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	attribute := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID}
	_, authorized, err := a.authorizer.Authorize(cts.Kit, attribute)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "biz permission denied")
	}

	commReq, err := decodeCommonReqAndValidate(cts)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req, err := parseReqFromRequestBody[proto.AccountAddReq](cts)
	if err != nil {
		return nil, err
	}
	if req.BkBizID != bizID {
		return nil, errf.Newf(errf.InvalidParameter,
			"path bk_biz_id(%d) does not match request body bk_biz_id(%d)", bizID, req.BkBizID)
	}
	handler := accounthandler.NewApplicationOfAddAccount(a.getHandlerOption(cts), a.authorizer, req)

	return a.create(cts, commReq, handler)
}

// CreateForCreateCvm ...
func (a *applicationSvc) CreateForCreateCvm(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	commReq, err := decodeCommonReqAndValidate(cts)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := a.checkApplyResPermission(cts, meta.Cvm); err != nil {
		return nil, err
	}

	opt := a.getHandlerOption(cts)

	switch vendor {
	case enumor.TCloud:
		req, err := parseReqFromRequestBody[cscvm.TCloudCvmCreateReq](cts)
		if err != nil {
			return nil, err
		}
		handler := tcloudcvmhandler.NewApplicationOfCreateTCloudCvm(opt, req)
		return a.create(cts, commReq, handler)
	case enumor.Aws:
		req, err := parseReqFromRequestBody[cscvm.AwsCvmCreateReq](cts)
		if err != nil {
			return nil, err
		}
		handler := awscvmhandler.NewApplicationOfCreateAwsCvm(opt, req)
		return a.create(cts, commReq, handler)
	case enumor.HuaWei:
		req, err := parseReqFromRequestBody[cscvm.HuaWeiCvmCreateReq](cts)
		if err != nil {
			return nil, err
		}
		handler := huaweicvmhandler.NewApplicationOfCreateHuaWeiCvm(opt, req)
		return a.create(cts, commReq, handler)
	case enumor.Gcp:
		req, err := parseReqFromRequestBody[cscvm.GcpCvmCreateReq](cts)
		if err != nil {
			return nil, err
		}
		handler := gcpcvmhandler.NewApplicationOfCreateGcpCvm(opt, req)
		return a.create(cts, commReq, handler)
	case enumor.Azure:
		req, err := parseReqFromRequestBody[cscvm.AzureCvmCreateReq](cts)
		if err != nil {
			return nil, err
		}
		handler := azurecvmhandler.NewApplicationOfCreateAzureCvm(opt, req)
		return a.create(cts, commReq, handler)
	}

	return nil, nil
}

// CreateForCreateVpc ...
func (a *applicationSvc) CreateForCreateVpc(cts *rest.Contexts) (interface{}, error) {

	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	commReq, err := decodeCommonReqAndValidate(cts)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := a.checkApplyResPermission(cts, meta.Vpc); err != nil {
		return nil, err
	}

	opt := a.getHandlerOption(cts)

	switch vendor {
	case enumor.TCloud:
		req, err := parseReqFromRequestBody[csvpc.TCloudVpcCreateReq](cts)
		if err != nil {
			return nil, err
		}
		handler := tcloudvpchandler.NewApplicationOfCreateTCloudVpc(opt, req)
		return a.create(cts, commReq, handler)
	case enumor.Aws:
		req, err := parseReqFromRequestBody[csvpc.AwsVpcCreateReq](cts)
		if err != nil {
			return nil, err
		}
		handler := awsvpchandler.NewApplicationOfCreateAwsVpc(opt, req)
		return a.create(cts, commReq, handler)
	case enumor.HuaWei:
		req, err := parseReqFromRequestBody[csvpc.HuaWeiVpcCreateReq](cts)
		if err != nil {
			return nil, err
		}
		handler := huaweivpchandler.NewApplicationOfCreateHuaWeiVpc(opt, req)
		return a.create(cts, commReq, handler)
	case enumor.Gcp:
		req, err := parseReqFromRequestBody[csvpc.GcpVpcCreateReq](cts)
		if err != nil {
			return nil, err
		}
		handler := gcpvpchandler.NewApplicationOfCreateGcpVpc(opt, req)
		return a.create(cts, commReq, handler)
	case enumor.Azure:
		req, err := parseReqFromRequestBody[csvpc.AzureVpcCreateReq](cts)
		if err != nil {
			return nil, err
		}
		handler := azurevpchandler.NewApplicationOfCreateAzureVpc(opt, req)
		return a.create(cts, commReq, handler)
	}

	return nil, nil
}

// CreateForCreateDisk ...
func (a *applicationSvc) CreateForCreateDisk(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	commReq, err := decodeCommonReqAndValidate(cts)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := a.checkApplyResPermission(cts, meta.Disk); err != nil {
		return nil, err
	}

	opt := a.getHandlerOption(cts)

	switch vendor {
	case enumor.TCloud:
		req, err := parseReqFromRequestBody[csdisk.TCloudDiskCreateReq](cts)
		if err != nil {
			return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
		}
		handler := tclouddiskhandler.NewApplicationOfCreateTCloudDisk(opt, req)
		return a.create(cts, commReq, handler)
	case enumor.Gcp:
		req, err := parseReqFromRequestBody[csdisk.GcpDiskCreateReq](cts)
		if err != nil {
			return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
		}
		handler := gcpdiskhandler.NewApplicationOfCreateGcpDisk(opt, req)
		return a.create(cts, commReq, handler)
	case enumor.Aws:
		req, err := parseReqFromRequestBody[csdisk.AwsDiskCreateReq](cts)
		if err != nil {
			return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
		}
		handler := awsdiskhandler.NewApplicationOfCreateAwsDisk(opt, req)
		return a.create(cts, commReq, handler)
	case enumor.HuaWei:
		req, err := parseReqFromRequestBody[csdisk.HuaWeiDiskCreateReq](cts)
		if err != nil {
			return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
		}
		handler := huaweidiskhandler.NewApplicationOfCreateHuaWeiDisk(opt, req)
		return a.create(cts, commReq, handler)
	case enumor.Azure:
		req, err := parseReqFromRequestBody[csdisk.AzureDiskCreateReq](cts)
		if err != nil {
			return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
		}
		handler := azurediskhandler.NewApplicationOfCreateAzureDisk(opt, req)
		return a.create(cts, commReq, handler)
	}

	return nil, nil
}

// CreateForCreateLB 创建负载均衡申请单
func (a *applicationSvc) CreateForCreateLB(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	commReq, err := decodeCommonReqAndValidate(cts)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := a.checkApplyResPermission(cts, meta.LoadBalancer); err != nil {
		return nil, err
	}

	opt := a.getHandlerOption(cts)

	switch vendor {
	case enumor.TCloud:
		req, err := parseReqFromRequestBody[hclb.TCloudLoadBalancerCreateReq](cts)
		if err != nil {
			return nil, err
		}
		handler := lbtcloud.NewApplicationOfCreateTCloudLB(opt, req)
		return a.create(cts, commReq, handler)
	}

	return nil, nil
}

// SysCreateForCreateLB creates a CLB application on behalf of a specified applicant for system-level callers.
func (a *applicationSvc) SysCreateForCreateLB(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	commReq, err := decodeSysCommonReqAndValidate(cts)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := a.checkApplyResPermission(cts, meta.LoadBalancer); err != nil {
		return nil, err
	}

	cts.Kit.User = commReq.Applicant
	opt := a.getHandlerOption(cts)

	switch vendor {
	case enumor.TCloud:
		req, err := parseReqFromRequestBody[hclb.TCloudLoadBalancerCreateReq](cts)
		if err != nil {
			return nil, err
		}
		handler := lbtcloud.NewApplicationOfCreateTCloudLB(opt, req)
		return a.create(cts, &commReq.CreateCommonReq, handler)
	}

	return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
}

// CreateForCreateMainAccount ...
func (a *applicationSvc) CreateForCreateMainAccount(cts *rest.Contexts) (interface{}, error) {
	req, err := parseReqFromRequestBody[proto.MainAccountCreateReq](cts)
	if err != nil {
		return nil, err
	}
	commReq := new(proto.CreateCommonReq)
	commReq.Remark = req.Memo

	// 组织架构信息暂时不需要用户填写，待需要这部分功能后，再删除组织架构的特殊设置
	req.DeptID = -1

	handler := createmainaccount.NewApplicationOfCreateMainAccount(a.getHandlerOption(cts), a.authorizer, req, nil)

	// 申请创建账号无需鉴权，由审批流程确认是否可以完成创建，如需对创建账号进行鉴权，可放开以下注释
	// authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.MainAccount, Action: meta.Create}}
	// err = a.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	// if err != nil {
	// 	return nil, err
	// }

	return a.create(cts, commReq, handler)
}

// CreateForUpdateMainAccount ...
func (a *applicationSvc) CreateForUpdateMainAccount(cts *rest.Contexts) (interface{}, error) {
	// 固定remark，该接口没有备注字段，为了保持接口一致，这里固定
	remark := "申请变更"
	commReq := new(proto.CreateCommonReq)
	commReq.Remark = &remark

	req, err := parseReqFromRequestBody[proto.MainAccountUpdateReq](cts)
	if err != nil {
		return nil, err
	}

	handler := updatemainaccount.NewApplicationOfUpdateMainAccount(a.getHandlerOption(cts), a.authorizer, req)

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.MainAccount, Action: meta.Update}}
	err = a.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	return a.create(cts, commReq, handler)
}

// CreateBizForAddSubAccount create application for adding subaccount.
func (a *applicationSvc) CreateBizForAddSubAccount(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	attribute := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID,
	}
	_, authorized, err := a.authorizer.Authorize(cts.Kit, attribute)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "biz permission denied")
	}

	req, err := parseReqFromRequestBody[proto.SubAccountBatchAddReq](cts)
	if err != nil {
		logs.Errorf("parse req from request body failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	req.Vendor = vendor
	req.BkBizID = bizID
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return a.batchCreateBizForAddSubAccount(cts, req)
}

func (a *applicationSvc) batchCreateBizForAddSubAccount(cts *rest.Contexts, req *proto.SubAccountBatchAddReq,
) (interface{}, error) {

	opt := a.getHandlerOption(cts)

	ids := make([]string, 0, len(req.SubAccounts))
	for i, subAccount := range req.SubAccounts {
		base := &subaccount.BaseSubAccountContent{
			Action:    enumor.SubAccountActionCreate,
			Vendor:    req.Vendor,
			BkBizID:   req.BkBizID,
			AccountID: subAccount.AccountID,
		}

		handler := createsubaccount.NewApplicationOfCreateSubAccount(opt, base, &req.SubAccounts[i])
		commReq := &proto.CreateCommonReq{Remark: subAccount.Memo}

		result, err := a.create(cts, commReq, handler)
		if err != nil {
			return nil, errf.NewFromErr(errf.Aborted,
				fmt.Errorf("create application for sub_account[%d](%s) failed, err: %w", i, subAccount.Name, err))
		}

		if createResult, ok := result.(*core.CreateResult); ok {
			ids = append(ids, createResult.ID)
		}
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// CreateBizForUpdateSubAccount create application for updating subaccount.
func (a *applicationSvc) CreateBizForUpdateSubAccount(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	attribute := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID,
	}
	_, authorized, err := a.authorizer.Authorize(cts.Kit, attribute)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "biz permission denied")
	}

	req, err := parseReqFromRequestBody[proto.SubAccountBatchUpdateReq](cts)
	if err != nil {
		logs.Errorf("parse req from request body failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	req.Vendor = vendor
	req.BkBizID = bizID
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return a.batchCreateBizForUpdateSubAccount(cts, req)
}

func (a *applicationSvc) batchCreateBizForUpdateSubAccount(cts *rest.Contexts, req *proto.SubAccountBatchUpdateReq,
) (interface{}, error) {

	subAccountIDs := make([]string, 0, len(req.SubAccounts))
	for _, item := range req.SubAccounts {
		subAccountIDs = append(subAccountIDs, item.ID)
	}

	subAccountMap, err := a.listSubAccountBasicInfo(cts, subAccountIDs)
	if err != nil {
		return nil, err
	}

	opt := a.getHandlerOption(cts)

	ids := make([]string, 0, len(req.SubAccounts))
	for i := range req.SubAccounts {
		info, ok := subAccountMap[req.SubAccounts[i].ID]
		if !ok {
			return nil, errf.Newf(errf.InvalidParameter, "sub account(%s) not found", req.SubAccounts[i].ID)
		}

		base := &subaccount.BaseSubAccountContent{
			Action:    enumor.SubAccountActionUpdate,
			Vendor:    req.Vendor,
			BkBizID:   req.BkBizID,
			AccountID: info.AccountID,
		}
		handler := updatesubaccount.NewApplicationOfUpdateSubAccount(
			opt, base, info.Name, &req.SubAccounts[i],
		)

		commReq := &proto.CreateCommonReq{Remark: req.SubAccounts[i].Memo}
		result, err := a.create(cts, commReq, handler)
		if err != nil {
			return nil, errf.NewFromErr(errf.Aborted,
				fmt.Errorf("create application for update sub_account[%d](%s) failed, err: %w",
					i, req.SubAccounts[i].ID, err))
		}

		if createResult, ok := result.(*core.CreateResult); ok {
			ids = append(ids, createResult.ID)
		}
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// CreateBizForDeleteSubAccount create application for deleting subaccount.
func (a *applicationSvc) CreateBizForDeleteSubAccount(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	if bizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is invalid")
	}

	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	attribute := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID,
	}
	_, authorized, err := a.authorizer.Authorize(cts.Kit, attribute)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "biz permission denied")
	}

	req, err := parseReqFromRequestBody[proto.SubAccountBatchDeleteReq](cts)
	if err != nil {
		logs.Errorf("parse req from request body failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	req.Vendor = vendor
	req.BkBizID = bizID
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return a.batchCreateBizForDeleteSubAccount(cts, req)
}

func (a *applicationSvc) batchCreateBizForDeleteSubAccount(cts *rest.Contexts, req *proto.SubAccountBatchDeleteReq,
) (interface{}, error) {

	infoMap, err := a.listSubAccountBasicInfo(cts, req.IDs)
	if err != nil {
		return nil, err
	}

	opt := a.getHandlerOption(cts)

	ids := make([]string, 0, len(req.IDs))
	for _, subAccountID := range req.IDs {
		info, ok := infoMap[subAccountID]
		if !ok {
			return nil, errf.Newf(errf.InvalidParameter, "sub account(%s) not found", subAccountID)
		}

		base := &subaccount.BaseSubAccountContent{
			Action:    enumor.SubAccountActionDelete,
			Vendor:    req.Vendor,
			BkBizID:   req.BkBizID,
			AccountID: info.AccountID,
		}
		handler := deletesubaccount.NewApplicationOfDeleteSubAccount(opt, base,
			&proto.SubAccountDeleteReq{SubAccountBasicInfo: converter.PtrToVal(info)},
		)

		result, err := a.create(cts, &proto.CreateCommonReq{}, handler)
		if err != nil {
			return nil, errf.NewFromErr(errf.Aborted,
				fmt.Errorf("create application for delete sub_account(%s) failed, err: %w",
					subAccountID, err))
		}

		if createResult, ok := result.(*core.CreateResult); ok {
			ids = append(ids, createResult.ID)
		}
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// listSubAccountBasicInfo batch queries subaccounts by IDs and returns a map keyed by sub-account ID.
func (a *applicationSvc) listSubAccountBasicInfo(cts *rest.Contexts, subAccountIDs []string,
) (map[string]*proto.SubAccountBasicInfo, error) {

	result, err := a.client.DataService().Global.SubAccount.List(
		cts.Kit,
		&core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", subAccountIDs)),
			Page:   &core.BasePage{Start: 0, Limit: uint(len(subAccountIDs))},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list sub accounts failed, err: %w", err)
	}

	if len(result.Details) != len(subAccountIDs) {
		return nil, fmt.Errorf("some sub accounts not found, expected %d but got %d",
			len(subAccountIDs), len(result.Details))
	}

	infoMap := make(map[string]*proto.SubAccountBasicInfo, len(result.Details))
	for _, sa := range result.Details {
		infoMap[sa.ID] = &proto.SubAccountBasicInfo{
			ID:        sa.ID,
			AccountID: sa.AccountID,
			Name:      sa.Name,
			CloudID:   sa.CloudID,
		}
	}

	return infoMap, nil
}

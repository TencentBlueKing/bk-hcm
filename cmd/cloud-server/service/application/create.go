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
	awsvpchandler "hcm/cmd/cloud-server/service/application/handlers/vpc/aws"
	azurevpchandler "hcm/cmd/cloud-server/service/application/handlers/vpc/azure"
	gcpvpchandler "hcm/cmd/cloud-server/service/application/handlers/vpc/gcp"
	huaweivpchandler "hcm/cmd/cloud-server/service/application/handlers/vpc/huawei"
	tcloudvpchandler "hcm/cmd/cloud-server/service/application/handlers/vpc/tcloud"
	proto "hcm/pkg/api/cloud-server/application"
	cscvm "hcm/pkg/api/cloud-server/cvm"
	csdisk "hcm/pkg/api/cloud-server/disk"
	csvpc "hcm/pkg/api/cloud-server/vpc"
	dataproto "hcm/pkg/api/data-service"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/itsm"
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
	serviceID, managers, err := a.getApprovalProcessInfo(cts, applicationType)
	if err != nil {
		return nil, fmt.Errorf("get approval process service id and managers failed, err: %v", err)
	}
	// 生成ITSM的回调地址
	callbackUrl := a.getCallbackUrl()
	// 渲染ITSM单据标题
	itsmTitle, err := handler.RenderItsmTitle()
	if err != nil {
		return nil, fmt.Errorf("render itsm ticket title error: %w", err)
	}
	// 渲染ITSM单据申请内容
	itsmForm, err := handler.RenderItsmForm()
	if err != nil {
		return nil, fmt.Errorf("render itsm ticket form error: %w", err)
	}
	// 获取ITSM单据涉及到的各个节点审批人
	approvers := handler.GetItsmApprover(managers)
	// 调用ITSM创建单据
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
		return nil, fmt.Errorf("call itsm create ticket api failed, err: %w", err)
	}
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
		applicationType == enumor.CreateVpc || applicationType == enumor.CreateLoadBalancer {
		bkBizIDs = handler.GetBkBizIDs()
	}

	result, err := a.client.DataService().Global.Application.CreateApplication(
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
	if err != nil {
		return nil, err
	}
	return result, nil
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
	commReq, err := decodeCommonReqAndValidate(cts)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req, err := parseReqFromRequestBody[proto.AccountAddReq](cts)
	if err != nil {
		return nil, err
	}
	handler := accounthandler.NewApplicationOfAddAccount(a.getHandlerOption(cts), a.authorizer, req)

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Account, Action: meta.Import}}
	err = a.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

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

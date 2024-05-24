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
	"errors"
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
	"hcm/cmd/cloud-server/service/application/handlers/load_balancer/tcloud"
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
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// Approve ...
func (a *applicationSvc) Approve(cts *rest.Contexts) (interface{}, error) {
	// Note: 由于该接口时给ITSM回调的，一般没什么反馈，这将任何错误到记录到日志里
	_, err := a.approve(cts)
	if err != nil {
		logs.Errorf("itsm approve callback failed, error: %v", err)
	}
	return nil, err
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

func (a *applicationSvc) approve(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.ItsmApproveResult)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验ITSM Token
	ok, err := a.itsmCli.VerifyToken(cts.Kit, req.Token)
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

	// 计算下个状态，实际上除了通过外，其他状态都是不需要变化了，要么是终结态，要么是持续中
	nextStatus := status
	// 对于审批通过，则下个状态为交付中，其他状态则保持原样
	if status == enumor.Pass {
		nextStatus = enumor.Delivering
	}

	// 更新状态
	err = a.updateStatusWithDetail(cts, application.ID, nextStatus, "")
	if err != nil {
		return nil, err
	}

	// 通过后需要进行资源交付
	if status == enumor.Pass {
		// TODO: 需要引入异步任务框架，这里先暂时用goroutine异步执行，无法记录状态等的，包括可能被kill等异常情况都无法处理和记录
		go a.deliver(cts, application)
	}

	return nil, nil
}

func parseReqFromApplicationContent[T any](content string) (*T, error) {
	// 解析申请单内容
	req := new(T)
	err := json.UnmarshalFromString(content, req)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal content error: %w", err)
	}

	return req, nil
}

func (a *applicationSvc) getHandlerOfCreateCvm(
	opt *handlers.HandlerOption, vendor enumor.Vendor, application *dataproto.ApplicationResp,
) (handlers.ApplicationHandler, error) {
	switch vendor {
	case enumor.TCloud:
		req, err := parseReqFromApplicationContent[cscvm.TCloudCvmCreateReq](application.Content)
		if err != nil {
			return nil, err
		}
		return tcloudcvmhandler.NewApplicationOfCreateTCloudCvm(opt, req), nil
	case enumor.Aws:
		req, err := parseReqFromApplicationContent[cscvm.AwsCvmCreateReq](application.Content)
		if err != nil {
			return nil, err
		}
		return awscvmhandler.NewApplicationOfCreateAwsCvm(opt, req), nil
	case enumor.HuaWei:
		req, err := parseReqFromApplicationContent[cscvm.HuaWeiCvmCreateReq](application.Content)
		if err != nil {
			return nil, err
		}
		return huaweicvmhandler.NewApplicationOfCreateHuaWeiCvm(opt, req), nil
	case enumor.Gcp:
		req, err := parseReqFromApplicationContent[cscvm.GcpCvmCreateReq](application.Content)
		if err != nil {
			return nil, err
		}
		return gcpcvmhandler.NewApplicationOfCreateGcpCvm(opt, req), nil
	case enumor.Azure:
		req, err := parseReqFromApplicationContent[cscvm.AzureCvmCreateReq](application.Content)
		if err != nil {
			return nil, err
		}
		return azurecvmhandler.NewApplicationOfCreateAzureCvm(opt, req), nil
	}

	return nil, fmt.Errorf("not support handler of create %s cvm", vendor)
}

func (a *applicationSvc) getHandlerOfCreateVpc(
	opt *handlers.HandlerOption, vendor enumor.Vendor, application *dataproto.ApplicationResp,
) (handlers.ApplicationHandler, error) {
	switch vendor {
	case enumor.TCloud:
		req, err := parseReqFromApplicationContent[csvpc.TCloudVpcCreateReq](application.Content)
		if err != nil {
			return nil, err
		}
		return tcloudvpchandler.NewApplicationOfCreateTCloudVpc(opt, req), nil
	case enumor.Aws:
		req, err := parseReqFromApplicationContent[csvpc.AwsVpcCreateReq](application.Content)
		if err != nil {
			return nil, err
		}
		return awsvpchandler.NewApplicationOfCreateAwsVpc(opt, req), nil
	case enumor.HuaWei:
		req, err := parseReqFromApplicationContent[csvpc.HuaWeiVpcCreateReq](application.Content)
		if err != nil {
			return nil, err
		}
		return huaweivpchandler.NewApplicationOfCreateHuaWeiVpc(opt, req), nil
	case enumor.Gcp:
		req, err := parseReqFromApplicationContent[csvpc.GcpVpcCreateReq](application.Content)
		if err != nil {
			return nil, err
		}
		return gcpvpchandler.NewApplicationOfCreateGcpVpc(opt, req), nil
	case enumor.Azure:
		req, err := parseReqFromApplicationContent[csvpc.AzureVpcCreateReq](application.Content)
		if err != nil {
			return nil, err
		}
		return azurevpchandler.NewApplicationOfCreateAzureVpc(opt, req), nil
	}

	return nil, fmt.Errorf("not support handler of create %s vpc", vendor)
}

func (a *applicationSvc) getHandlerOfCreateDisk(
	opt *handlers.HandlerOption, vendor enumor.Vendor, application *dataproto.ApplicationResp,
) (handlers.ApplicationHandler, error) {
	switch vendor {
	case enumor.TCloud:
		req, err := parseReqFromApplicationContent[csdisk.TCloudDiskCreateReq](application.Content)
		if err != nil {
			return nil, err
		}
		return tclouddiskhandler.NewApplicationOfCreateTCloudDisk(opt, req), nil
	case enumor.Gcp:
		req, err := parseReqFromApplicationContent[csdisk.GcpDiskCreateReq](application.Content)
		if err != nil {
			return nil, err
		}
		return gcpdiskhandler.NewApplicationOfCreateGcpDisk(opt, req), nil
	case enumor.Aws:
		req, err := parseReqFromApplicationContent[csdisk.AwsDiskCreateReq](application.Content)
		if err != nil {
			return nil, err
		}
		return awsdiskhandler.NewApplicationOfCreateAwsDisk(opt, req), nil
	case enumor.HuaWei:
		req, err := parseReqFromApplicationContent[csdisk.HuaWeiDiskCreateReq](application.Content)
		if err != nil {
			return nil, err
		}
		return huaweidiskhandler.NewApplicationOfCreateHuaWeiDisk(opt, req), nil
	case enumor.Azure:
		req, err := parseReqFromApplicationContent[csdisk.AzureDiskCreateReq](application.Content)
		if err != nil {
			return nil, err
		}
		return azurediskhandler.NewApplicationOfCreateAzureDisk(opt, req), nil
	default:
		return nil, fmt.Errorf("not support handler of create %s disk", vendor)
	}
}

func (a *applicationSvc) getHandlerOfCreateLoadBalancer(opt *handlers.HandlerOption, vendor enumor.Vendor,
	application *dataproto.ApplicationResp) (handlers.ApplicationHandler, error) {

	switch vendor {
	case enumor.TCloud:
		req, err := parseReqFromApplicationContent[hclb.TCloudLoadBalancerCreateReq](application.Content)
		if err != nil {
			return nil, err
		}
		return tcloud.NewApplicationOfCreateTCloudLB(opt, req), nil
	default:
		return nil, fmt.Errorf("not support handler of create %s load balancer", vendor)
	}
}

func (a *applicationSvc) getHandlerByApplication(cts *rest.Contexts, application *dataproto.ApplicationResp) (
	handlers.ApplicationHandler, error) {

	opt := a.getHandlerOption(cts)

	// 只解析申请单的vendor
	onlyVendor, err := parseReqFromApplicationContent[struct {
		Vendor enumor.Vendor `json:"vendor"`
	}](application.Content)
	if err != nil {
		return nil, err
	}
	vendor := onlyVendor.Vendor

	// 根据不同申请单类型，生成对应的Handler
	switch application.Type {
	case enumor.AddAccount:
		req, err := parseReqFromApplicationContent[proto.AccountAddReq](application.Content)
		if err != nil {
			return nil, err
		}
		return accounthandler.NewApplicationOfAddAccount(opt, a.authorizer, req), nil
	case enumor.CreateCvm:
		return a.getHandlerOfCreateCvm(opt, vendor, application)
	case enumor.CreateVpc:
		return a.getHandlerOfCreateVpc(opt, vendor, application)
	case enumor.CreateDisk:
		return a.getHandlerOfCreateDisk(opt, vendor, application)
	case enumor.CreateLoadBalancer:
		return a.getHandlerOfCreateLoadBalancer(opt, vendor, application)
	}
	return nil, errors.New("not handler to support")
}

func (a *applicationSvc) deliver(cts *rest.Contexts, application *dataproto.ApplicationResp) {
	// 将执行人设置为申请人
	cts.Kit.User = application.Applicant

	// 除非交付成功，否则都属于交付失败状态
	deliverStatus := enumor.DeliverError
	deliveryDetailStr := `{"error": "unknown deliver error"}`
	defer func() {
		err := a.updateStatusWithDetail(cts, application.ID, deliverStatus, deliveryDetailStr)
		if err != nil {
			logs.Errorf("%s execute application[id=%s] delivery of %s failed, updateStatusWithDetail err: %s, rid: %s",
				constant.ApplicationDeliverFailed, application.ID, application.Type, err, cts.Kit.Rid)
			return
		}
	}()

	// 根据不同申请单类型，获取对应的Handler
	handler, err := a.getHandlerByApplication(cts, application)
	if err != nil {
		logs.Errorf("execute application[id=%s] delivery of %s failed, NewHandler err: %s, rid: %s",
			application.ID, application.Type, err, cts.Kit.Rid)
		deliveryDetailStr = fmt.Sprintf(`{"error": "get handler by application failed, err: %v"}`, err)
		return
	}

	// 预处理申请内容数据，来自DB的数据
	err = handler.PrepareReqFromContent()
	if err != nil {
		logs.Errorf("execute application[id=%s] delivery of %s failed, PrepareReqFromContent err: %s, rid: %s",
			application.ID, application.Type, err, cts.Kit.Rid)
		deliveryDetailStr = fmt.Sprintf(`{"error": "prepare request from content failed, err: %v"}`, err)
		return
	}

	// 再次校验数据正确性（特别是唯一性校验，申请时可能通过，但是审批后可能已经有其他存在了）
	err = handler.CheckReq()
	if err != nil {
		logs.Errorf("execute application[id=%s] delivery of %s failed, CheckReq err: %s, rid: %s",
			application.ID, application.Type, err, cts.Kit.Rid)
		deliveryDetailStr = fmt.Sprintf(`{"error": "check request failed, err: %v"}`, err)
		return
	}

	// 执行交付
	deliverStatus, deliveryDetail, err := handler.Deliver()
	// Note: 排查需要，这里无论失败还是成功，都记录日志，因为没有异步框架可以记录这些信息
	logs.Infof("execute application[id=%s] delivery of %s, deliver status: %s, detail: %+v, rid: %s",
		application.ID, application.Type, deliverStatus, deliveryDetail, cts.Kit.Rid)

	if err != nil {
		logs.Errorf("execute application[id=%s] delivery of %s failed, err: %s, rid: %s", application.ID,
			application.Type, err, cts.Kit.Rid)

		deliverStatus = enumor.DeliverError
	}

	// 更新DB里单据的交付状态和详情
	deliveryDetailStr, err = json.MarshalToString(deliveryDetail)
	if err != nil {
		logs.Errorf("marshal deliver detail failed, err: %v, detail: %+v, rid: %s", err, deliveryDetail, cts.Kit.Rid)

		deliverStatus = enumor.DeliverError
		deliveryDetailStr = `{"error": "marshal deliver detail failed"}`
	}
}

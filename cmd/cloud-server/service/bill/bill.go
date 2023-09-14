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

// Package bill ...
package bill

import (
	"fmt"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	cloudserver "hcm/pkg/api/cloud-server"
	csbill "hcm/pkg/api/cloud-server/bill"
	"hcm/pkg/api/core"
	hcbill "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
)

// InitBillService initialize the bill service.
func InitBillService(c *capability.Capability) {
	svc := &billSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	h.Add("ListBills", "POST", "/vendors/{vendor}/bills/list", svc.ListBills)
	h.Add("ListBillsConfig", "POST", "/bills/config/list", svc.ListBillsConfig)

	h.Load(c.WebService)
}

type billSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

// ListBills list bills.
func (b *billSvc) ListBills(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	// 校验用户是否有拉取云账单的权限
	if err := b.checkPermission(cts, meta.CostManage, meta.Find); err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.Aws:
		billReq, err := b.getAwsListReq(cts)
		if err != nil {
			return nil, err
		}
		return b.client.HCService().Aws.Bill.List(cts.Kit.Ctx, cts.Kit.Header(), billReq)
	case enumor.TCloud:
		billReq, err := b.getTCloudListReq(cts)
		if err != nil {
			return nil, err
		}
		return b.client.HCService().TCloud.Bill.List(cts.Kit.Ctx, cts.Kit.Header(), billReq)
	case enumor.HuaWei:
		billReq, err := b.getHuaWeiListReq(cts)
		if err != nil {
			return nil, err
		}
		return b.client.HCService().HuaWei.Bill.List(cts.Kit.Ctx, cts.Kit.Header(), billReq)
	case enumor.Azure:
		billReq, err := b.getAzureListReq(cts)
		if err != nil {
			return nil, err
		}
		return b.client.HCService().Azure.Bill.List(cts.Kit.Ctx, cts.Kit.Header(), billReq)
	case enumor.Gcp:
		billReq, err := b.getGcpListReq(cts)
		if err != nil {
			return nil, err
		}
		return b.client.HCService().Gcp.Bill.List(cts.Kit.Ctx, cts.Kit.Header(), billReq)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

func (b *billSvc) getAwsListReq(cts *rest.Contexts) (*hcbill.AwsBillListReq, error) {
	req := new(csbill.AwsBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	billReq := &hcbill.AwsBillListReq{
		AccountID: req.AccountID,
		BeginDate: req.BeginDate,
		EndDate:   req.EndDate,
		Page:      (*hcbill.AwsBillListPage)(req.Page),
	}
	return billReq, nil
}

func (b *billSvc) getTCloudListReq(cts *rest.Contexts) (*hcbill.TCloudBillListReq, error) {
	req := new(csbill.TCloudBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	billReq := &hcbill.TCloudBillListReq{
		AccountID: req.AccountID,
		Month:     req.Month,
		BeginDate: req.BeginDate,
		EndDate:   req.EndDate,
		Page:      req.Page,
		Context:   req.Context,
	}
	return billReq, nil
}

func (b *billSvc) getHuaWeiListReq(cts *rest.Contexts) (*hcbill.HuaWeiBillListReq, error) {
	req := new(csbill.HuaWeiBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	billReq := &hcbill.HuaWeiBillListReq{
		AccountID: req.AccountID,
		Month:     req.Month,
		Page:      req.Page,
	}
	return billReq, nil
}

func (b *billSvc) getAzureListReq(cts *rest.Contexts) (*hcbill.AzureBillListReq, error) {
	req := new(csbill.AzureBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	billReq := &hcbill.AzureBillListReq{
		AccountID: req.AccountID,
		BeginDate: req.BeginDate,
		EndDate:   req.EndDate,
		Page:      req.Page,
	}
	return billReq, nil
}

func (b *billSvc) getGcpListReq(cts *rest.Contexts) (*hcbill.GcpBillListReq, error) {
	req := new(csbill.GcpBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	billReq := &hcbill.GcpBillListReq{
		BillAccountID: req.BillAccountID,
		AccountID:     req.AccountID,
		Month:         req.Month,
		BeginDate:     req.BeginDate,
		EndDate:       req.EndDate,
		Page:          req.Page,
	}
	return billReq, nil
}

func (b *billSvc) checkPermission(cts *rest.Contexts, resType meta.ResourceType, action meta.Action) error {
	return b.checkPermissions(cts, resType, action)
}

// checkPermissions check permissions
func (b *billSvc) checkPermissions(cts *rest.Contexts, resType meta.ResourceType, action meta.Action) error {
	resources := make([]meta.ResourceAttribute, 0)
	resources = append(resources, meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   resType,
			Action: action,
		},
	})

	_, authorized, err := b.authorizer.Authorize(cts.Kit, resources...)
	if err != nil {
		return errf.NewFromErr(
			errf.PermissionDenied,
			fmt.Errorf("check %s account permissions failed, err: %v", action, err),
		)
	}

	if !authorized {
		return errf.NewFromErr(errf.PermissionDenied, fmt.Errorf("you have not permission of %s", action))
	}

	return nil
}

// ListBillsConfig list bills config.
func (b *billSvc) ListBillsConfig(cts *rest.Contexts) (interface{}, error) {
	// 校验用户是否有拉取权限
	if err := b.checkPermission(cts, meta.CostManage, meta.Find); err != nil {
		return nil, err
	}

	req := new(cloudserver.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Filter: req.Filter,
		Page:   req.Page,
	}
	return b.client.DataService().Global.Bill.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
}

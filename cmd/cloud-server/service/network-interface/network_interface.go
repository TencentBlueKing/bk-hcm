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

package networkinterface

import (
	"fmt"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/hooks/handler"
)

// InitNetworkInterfaceService initialize the network interface service.
func InitNetworkInterfaceService(c *capability.Capability) {
	svc := &netSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	h.Add("ListNetworkInterface", "POST", "/network_interfaces/list", svc.ListNetworkInterface)
	h.Add("GetNetworkInterface", "GET", "/network_interfaces/{id}", svc.GetNetworkInterface)
	h.Add("ListNetworkInterfaceExtByCvmID", "GET", "/vendors/{vendor}/network_interfaces/cvms/{cvm_id}",
		svc.ListNetworkInterfaceExtByCvmID)
	h.Add("AssignNetworkInterfaceToBiz", "POST", "/network_interfaces/assign/bizs",
		svc.AssignNetworkInterfaceToBiz)

	// network interface biz apis
	h.Add("ListBizNetworkInterface", "POST", "/bizs/{bk_biz_id}/network_interfaces/list", svc.ListBizNetworkInterface)
	h.Add("GetBizNetworkInterface", "GET", "/bizs/{bk_biz_id}/network_interfaces/{id}", svc.GetBizNetworkInterface)
	h.Add("ListBizNICExtByCvmID", "GET", "/bizs/{bk_biz_id}/vendors/{vendor}/network_interfaces/cvms/{cvm_id}",
		svc.ListBizNICExtByCvmID)

	h.Load(c.WebService)
}

type netSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

// ListNetworkInterface list network interface.
func (svc *netSvc) ListNetworkInterface(cts *rest.Contexts) (interface{}, error) {
	return svc.listNetworkInterface(cts, handler.ListResourceAuthRes)
}

// ListBizNetworkInterface list biz network interface.
func (svc *netSvc) ListBizNetworkInterface(cts *rest.Contexts) (interface{}, error) {
	return svc.listNetworkInterface(cts, handler.ListBizAuthRes)
}

func (svc *netSvc) listNetworkInterface(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{},
	error) {

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.NetworkInterface, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &cloudserver.NetworkInterfaceListResult{Details: make([]coreni.BaseNetworkInterface, 0)}, nil
	}
	req.Filter = expr

	// list network interface
	res, err := svc.client.DataService().Global.NetworkInterface.List(cts.Kit.Ctx, cts.Kit.Header(), req)
	if err != nil {
		return nil, err
	}

	return &cloudserver.NetworkInterfaceListResult{Count: res.Count, Details: res.Details}, nil
}

// GetNetworkInterface get network interface details.
func (svc *netSvc) GetNetworkInterface(cts *rest.Contexts) (interface{}, error) {
	return svc.getNetworkInterface(cts, handler.ResValidWithAuth)
}

// GetBizNetworkInterface get biz network interface details.
func (svc *netSvc) GetBizNetworkInterface(cts *rest.Contexts) (interface{}, error) {
	return svc.getNetworkInterface(cts, handler.BizValidWithAuth)
}

func (svc *netSvc) getNetworkInterface(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{},
	error) {

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		enumor.NetworkInterfaceCloudResType, id)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.NetworkInterface,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	// get detail info
	switch basicInfo.Vendor {
	case enumor.Azure:
		return svc.client.DataService().Azure.NetworkInterface.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	case enumor.Gcp:
		return svc.client.DataService().Gcp.NetworkInterface.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	case enumor.HuaWei:
		return svc.client.DataService().HuaWei.NetworkInterface.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", basicInfo.Vendor)
	}
}

// CheckNIInBiz check if network interface are in the specified biz.
func CheckNIInBiz(kt *kit.Kit, client *client.ClientSet, rule filter.RuleFactory, bizID int64) error {
	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "bk_biz_id", Op: filter.NotEqual.Factory(), Value: bizID}, rule,
			},
		},
		Page: &core.BasePage{
			Count: true,
		},
	}
	result, err := client.DataService().Global.NetworkInterface.List(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("count network interface that are not in biz failed, err: %v, req: %+v, rid: %s",
			err, req, kt.Rid)
		return err
	}

	if result.Count != 0 {
		return fmt.Errorf("%d network interface are already assigned", result.Count)
	}

	return nil
}

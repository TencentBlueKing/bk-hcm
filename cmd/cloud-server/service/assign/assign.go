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

// Package assign ...
package assign

import (
	"net/http"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	proto "hcm/pkg/api/cloud-server/assign"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

// InitService initialize the assign service.
func InitService(c *capability.Capability) {
	s := &svc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	h.Add("AssignResourceToBiz", http.MethodPost, "/resources/assign/bizs", s.AssignResourceToBiz)

	h.Load(c.WebService)
}

type svc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

// AssignResourceToBiz assign an account's cloud resource to biz, **only for ui**.
func (svc *svc) AssignResourceToBiz(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AssignResourceToBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.CloudResource, Action: meta.Assign,
		ResourceID: req.AccountID}, BizID: req.BkBizID}
	err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	// check if account is related to assigned biz
	accountBizReq := &core.ListReq{
		Filter: tools.EqualExpression("account_id", req.AccountID),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"bk_biz_id"},
	}
	accountBizRes, err := svc.client.DataService().Global.Account.ListAccountBizRel(cts.Kit.Ctx, cts.Kit.Header(),
		accountBizReq)
	if err != nil {
		logs.Errorf("get account biz relation failed, err: %v, req: %+v, rid: %s", err, accountBizReq, cts.Kit.Rid)
		return nil, err
	}

	if len(accountBizRes.Details) == 0 || accountBizRes.Details[0].BkBizID != req.BkBizID {
		return nil, errf.Newf(errf.InvalidParameter, "account(%s) and biz(%d) not matches", req.AccountID, req.BkBizID)
	}

	// compatible for assign all resource scenario
	if req.IsAllResType {
		req.ResTypes = []enumor.CloudResourceType{enumor.CvmCloudResType, enumor.DiskCloudResType,
			enumor.EipCloudResType, enumor.NetworkInterfaceCloudResType, enumor.SecurityGroupCloudResType,
			enumor.GcpFirewallRuleCloudResType, enumor.VpcCloudResType, enumor.SubnetCloudResType,
			enumor.RouteTableCloudResType}
	}

	// check if all vpc has cloud area id
	if slice.IsItemInSlice(req.ResTypes, enumor.VpcCloudResType) {
		vpcReq := &core.ListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{tools.EqualExpression("account_id", req.AccountID), &filter.AtomRule{
					Field: "bk_cloud_id", Op: filter.LessThanEqual.Factory(), Value: 0}},
			},
			Page: &core.BasePage{Count: true},
		}

		vpcRes, err := svc.client.DataService().Global.Vpc.List(cts.Kit.Ctx, cts.Kit.Header(), vpcReq)
		if err != nil {
			logs.Errorf("count not bind cloud area vpc failed, err: %v, req: %+v, rid: %s", err, vpcReq, cts.Kit.Rid)
			return nil, err
		}

		if vpcRes.Count > 0 {
			return nil, errf.Newf(errf.InvalidParameter, "%d vpcs are not bind with cloud area", vpcRes.Count)
		}
	}

	assignReq := &cloud.AssignResourceToBizReq{
		AccountID: req.AccountID,
		BkBizID:   req.BkBizID,
		ResTypes:  req.ResTypes,
	}
	err = svc.client.DataService().Global.Cloud.AssignResourceToBiz(cts.Kit.Ctx, cts.Kit.Header(), assignReq)
	if err != nil {
		logs.Errorf("assign resource to biz failed, err: %v, req: %+v, rid: %s", err, assignReq, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

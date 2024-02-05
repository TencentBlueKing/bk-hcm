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

// Package argstpl ...
package argstpl

import (
	proto "hcm/pkg/api/cloud-server"
	csargstpl "hcm/pkg/api/cloud-server/argument-template"
	csprotoargstpl "hcm/pkg/api/cloud-server/argument-template"
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// ListArgsTpl list resource argument template.
func (svc *argsTplSvc) ListArgsTpl(cts *rest.Contexts) (interface{}, error) {
	return svc.listArgsTpl(cts, handler.ListResourceAuthRes)
}

// ListBizArgsTpl list biz argument template.
func (svc *argsTplSvc) ListBizArgsTpl(cts *rest.Contexts) (interface{}, error) {
	return svc.listArgsTpl(cts, handler.ListBizAuthRes)
}

func (svc *argsTplSvc) listArgsTpl(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	req := new(proto.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.Biz, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	listReq := &core.ListReq{
		Filter: expr,
		Page:   req.Page,
	}
	return svc.client.DataService().Global.ArgsTpl.ListArgsTpl(cts.Kit, listReq)
}

// ListArgsTplBindInstanceRule list resource argument template bind instance rule.
func (svc *argsTplSvc) ListArgsTplBindInstanceRule(cts *rest.Contexts) (interface{}, error) {
	return svc.listArgsTplBindInstanceRule(cts, handler.ListResourceAuthRes)
}

// ListBizArgsTplBindInstanceRule list biz argument template bind instance rule.
func (svc *argsTplSvc) ListBizArgsTplBindInstanceRule(cts *rest.Contexts) (interface{}, error) {
	return svc.listArgsTplBindInstanceRule(cts, handler.ListBizAuthRes)
}

func (svc *argsTplSvc) listArgsTplBindInstanceRule(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (
	any, error) {

	req := new(csprotoargstpl.ArgsTplBatchIDsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	flt := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "id",
				Op:    filter.In.Factory(),
				Value: req.IDs,
			},
		},
	}
	_, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.Biz, Action: meta.Find, Filter: flt})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	listReq := &protocloud.TCloudSGRuleListReq{
		Filter: &filter.Expression{
			Op: filter.Or,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "service_id", Op: filter.In.Factory(), Value: req.IDs},
				&filter.AtomRule{Field: "service_group_id", Op: filter.In.Factory(), Value: req.IDs},
				&filter.AtomRule{Field: "address_id", Op: filter.In.Factory(), Value: req.IDs},
				&filter.AtomRule{Field: "address_group_id", Op: filter.In.Factory(), Value: req.IDs},
			},
		},
		Page: core.NewDefaultBasePage(),
	}

	sgRuleList, err := svc.client.DataService().TCloud.SecurityGroup.ListSecurityGroupRuleExt(
		cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		return nil, err
	}

	sgMap := make(map[string]int64)
	for _, item := range sgRuleList.SecurityGroup {
		if _, ok := sgMap[item.ID]; !ok {
			sgMap[item.ID] += 1
		}
	}

	list := svc.buildArgsTplBinding(sgRuleList, sgMap, req.IDs)
	return list, nil
}

// buildArgsTplBinding build argument template binding.
// description: sgRuleList.SecurityGroupRule里面会同时包含AddressID、AddressGroupID、ServiceID、ServiceGroupID
func (svc *argsTplSvc) buildArgsTplBinding(sgRuleList *protocloud.TCloudSGRuleListExtResult,
	sgMap map[string]int64, ids []string) []*csargstpl.BindArgsTplInstanceRuleResp {

	argsTplMap := make(map[string]*csargstpl.BindArgsTplInstanceRuleResp)
	for _, item := range sgRuleList.SecurityGroupRule {
		if item.AddressID != nil {
			tmpID := converter.PtrToVal(item.AddressID)
			if _, ok := argsTplMap[tmpID]; !ok {
				argsTplMap[tmpID] = &csargstpl.BindArgsTplInstanceRuleResp{
					ID: tmpID,
				}
			}
			argsTplMap[tmpID].InstanceNum = sgMap[item.SecurityGroupID]
			argsTplMap[tmpID].RuleNum++
		}

		if item.AddressGroupID != nil {
			tmpID := converter.PtrToVal(item.AddressGroupID)
			if _, ok := argsTplMap[tmpID]; !ok {
				argsTplMap[tmpID] = &csargstpl.BindArgsTplInstanceRuleResp{
					ID: tmpID,
				}
			}
			argsTplMap[tmpID].InstanceNum = sgMap[item.SecurityGroupID]
			argsTplMap[tmpID].RuleNum++
		}

		if item.ServiceID != nil {
			tmpID := converter.PtrToVal(item.ServiceID)
			if _, ok := argsTplMap[tmpID]; !ok {
				argsTplMap[tmpID] = &csargstpl.BindArgsTplInstanceRuleResp{
					ID: tmpID,
				}
			}
			argsTplMap[tmpID].InstanceNum = sgMap[item.SecurityGroupID]
			argsTplMap[tmpID].RuleNum++
		}

		if item.ServiceGroupID != nil {
			tmpID := converter.PtrToVal(item.ServiceGroupID)
			if _, ok := argsTplMap[tmpID]; !ok {
				argsTplMap[tmpID] = &csargstpl.BindArgsTplInstanceRuleResp{
					ID: tmpID,
				}
			}
			argsTplMap[tmpID].InstanceNum = sgMap[item.SecurityGroupID]
			argsTplMap[tmpID].RuleNum++
		}
	}

	list := make([]*csargstpl.BindArgsTplInstanceRuleResp, 0)
	for _, tmpID := range ids {
		if val, ok := argsTplMap[tmpID]; ok {
			list = append(list, val)
		}
	}

	return list
}

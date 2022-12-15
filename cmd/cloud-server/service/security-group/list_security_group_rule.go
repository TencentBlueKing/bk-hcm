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

package securitygroup

import (
	proto "hcm/pkg/api/cloud-server"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/rest"
)

// ListSecurityGroupRule list security group rule.
func (svc securityGroupSvc) ListSecurityGroupRule(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(proto.SecurityGroupRuleListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		listReq := &dataproto.TCloudSGRuleListReq{
			Filter: tools.AllExpression(),
			Page: &types.BasePage{
				Count: req.Page.Count,
				Start: req.Page.Start,
				Limit: req.Page.Limit,
			},
		}
		return svc.client.DataService().TCloud.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx,
			cts.Kit.Header(), listReq, sgID)

	case enumor.Aws:
		listReq := &dataproto.AwsSGRuleListReq{
			Filter: tools.AllExpression(),
			Page: &types.BasePage{
				Count: req.Page.Count,
				Start: req.Page.Start,
				Limit: req.Page.Limit,
			},
		}
		return svc.client.DataService().Aws.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(),
			listReq, sgID)

	case enumor.HuaWei:
		listReq := &dataproto.HuaWeiSGRuleListReq{
			Filter: tools.AllExpression(),
			Page: &types.BasePage{
				Count: req.Page.Count,
				Start: req.Page.Start,
				Limit: req.Page.Limit,
			},
		}
		return svc.client.DataService().HuaWei.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(),
			listReq, sgID)

	case enumor.Azure:
		listReq := &dataproto.AzureSGRuleListReq{
			Filter: tools.AllExpression(),
			Page: &types.BasePage{
				Count: req.Page.Count,
				Start: req.Page.Start,
				Limit: req.Page.Limit,
			},
		}
		return svc.client.DataService().Azure.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(),
			listReq, sgID)

	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}
}

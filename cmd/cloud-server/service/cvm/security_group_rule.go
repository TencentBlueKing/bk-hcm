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

package cvm

import (
	"fmt"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListCvmSecurityGroupRules ...
func (svc *cvmSvc) ListCvmSecurityGroupRules(cts *rest.Contexts) (interface{}, error) {

	cvmID := cts.PathParameter("cvm_id").String()
	if len(cvmID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "cvm id is required")
	}
	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.CvmCloudResType, cvmID)
	if err != nil {
		return nil, err
	}
	// validate biz and authorize
	err = handler.BizOperateAuth(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Cvm,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	return svc.listCvmSecurityGroupRules(cts.Kit, basicInfo.Vendor, cvmID, sgID, req)
}

func (svc *cvmSvc) listCvmSecurityGroupRules(kt *kit.Kit, vendor enumor.Vendor, cvmID, sgID string, req *core.ListReq) (
	interface{}, error) {

	err := svc.checkCvmAndSecurityGroupRel(kt, cvmID, sgID)
	if err != nil {
		logs.Errorf("check cvm and security group relation failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		listReq := &dataproto.TCloudSGRuleListReq{
			Filter: req.Filter,
			Page:   req.Page,
		}
		return svc.client.DataService().TCloud.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq, sgID)
	case enumor.Aws:
		listReq := &dataproto.AwsSGRuleListReq{
			Filter: req.Filter,
			Page:   req.Page,
		}
		return svc.client.DataService().Aws.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq, sgID)
	case enumor.HuaWei:
		listReq := &dataproto.HuaWeiSGRuleListReq{
			Filter: req.Filter,
			Page:   req.Page,
		}
		return svc.client.DataService().HuaWei.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq, sgID)
	case enumor.Azure:
		listReq := &dataproto.AzureSGRuleListReq{
			Filter: req.Filter,
			Page:   req.Page,
		}
		return svc.client.DataService().Azure.SecurityGroup.ListSecurityGroupRule(kt.Ctx, kt.Header(), listReq, sgID)
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support for get cvm security group rules", vendor)
	}
}

func (svc *cvmSvc) checkCvmAndSecurityGroupRel(kt *kit.Kit, cvmID, sgID string) error {
	checkReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("cvm_id", cvmID),
			tools.RuleEqual("security_group_id", sgID),
		),
		Page: core.NewCountPage(),
	}
	rels, err := svc.client.DataService().Global.SGCvmRel.ListSgCvmRels(kt.Ctx, kt.Header(), checkReq)
	if err != nil {
		logs.Errorf("check cvm and security group relation failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	if rels.Count == 0 {
		logs.Errorf("cvm and security group relation not found, cvm_id: %s, security_group_id: %s, rid: %s",
			cvmID, sgID, kt.Rid)
		return fmt.Errorf("cvm and security group relation not found, cvm_id: %s, security_group_id: %s", cvmID, sgID)
	}
	return nil
}

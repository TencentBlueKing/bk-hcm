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
	"fmt"

	"hcm/cmd/cloud-server/service/common"
	cloudserver "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	dataproto "hcm/pkg/api/data-service/cloud"
	proto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// CloneBizSecurityGroup clone biz security group
func (svc *securityGroupSvc) CloneBizSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}

	sgID := cts.PathParameter("id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(cloudserver.SecurityGroupCloneReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg, err := svc.getSecurityGroupFromDB(cts.Kit, sgID)
	if err != nil {
		logs.Errorf("get security group from db failed, err: %v, sgID: %s, rid: %s", err, sgID, cts.Kit.Rid)
		return nil, err
	}

	err = svc.authorizeCloneSecurityGroup(cts, sg.AccountID, bizID)
	if err != nil {
		logs.Errorf("authorize clone security group failed, err: %v, sgID: %s, rid: %s", err, sgID, cts.Kit.Rid)
		return nil, err
	}

	switch sg.Vendor {
	case enumor.TCloud:
		result, err := svc.tcloudCloneSecurityGroup(cts.Kit, bizID, sg, req)
		if err != nil {
			logs.Errorf("clone security group failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
			return nil, err
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported vendor: %s for clone security group", sg.Vendor)
	}
}

func (svc *securityGroupSvc) authorizeCloneSecurityGroup(cts *rest.Contexts, accountID string, bizID int64) error {
	// validate  create permission
	err := handler.BizOperateAuth(cts,
		&handler.ValidWithAuthOption{
			Authorizer: svc.authorizer, ResType: meta.SecurityGroup, Action: meta.Create,
			BasicInfo: common.GetCloudResourceBasicInfo(accountID, bizID)},
	)
	if err != nil {
		return err
	}

	// validate find permission
	_, noPerm, err := handler.ListBizAuthRes(cts,
		&handler.ListAuthResOption{
			Authorizer: svc.authorizer, ResType: meta.SecurityGroup, Action: meta.Find,
		},
	)
	if err != nil {
		return err
	}
	if noPerm {
		return errf.New(errf.PermissionDenied, "permission denied for access security group")
	}
	return nil
}

func (svc *securityGroupSvc) tcloudCloneSecurityGroup(kt *kit.Kit, bizID int64, sg *cloud.BaseSecurityGroup,
	req *cloudserver.SecurityGroupCloneReq) (*core.CreateResult, error) {

	cloneReq := &proto.TCloudSecurityGroupCloneReq{
		SecurityGroupID: sg.ID,
		Manager:         req.Manager,
		BakManager:      req.BakManager,
		ManagementBizID: bizID,
		TargetRegion:    req.TargetRegion,
	}
	if req.Name == nil {
		cloneReq.GroupName = fmt.Sprintf("%s-copy", sg.Name)
	} else {
		cloneReq.GroupName = converter.PtrToVal(req.Name)
	}
	result, err := svc.client.HCService().TCloud.SecurityGroup.CloneSecurityGroup(kt, cloneReq)
	if err != nil {
		logs.Errorf("clone security group failed, err: %v, req: %+v, rid: %s", err, cloneReq, kt.Rid)
		return nil, err
	}
	return result, nil
}

func (svc *securityGroupSvc) getSecurityGroupFromDB(kt *kit.Kit, sgID string) (*cloud.BaseSecurityGroup, error) {
	listReq := &dataproto.SecurityGroupListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("id", sgID),
		),
		Page: core.NewDefaultBasePage(),
	}
	resp, err := svc.client.DataService().Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list security group failed, id: %s, err: %v, rid: %s", sgID, err, kt.Rid)
		return nil, err
	}
	if len(resp.Details) == 0 {
		return nil, fmt.Errorf("security group(%s) not found", sgID)
	}
	return &resp.Details[0], nil
}

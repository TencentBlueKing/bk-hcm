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
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// InitSecurityGroupService initial the security group service
func InitSecurityGroupService(c *capability.Capability) {
	svc := &securityGroupSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()
	h.Add("GetSecurityGroup", http.MethodGet, "/security_groups/{id}", svc.GetSecurityGroup)
	h.Add("BatchUpdateSecurityGroup", http.MethodPatch, "/security_groups/{id}", svc.UpdateSecurityGroup)
	h.Add("BatchDeleteSecurityGroup", http.MethodDelete, "/security_groups/batch", svc.BatchDeleteSecurityGroup)
	h.Add("ListSecurityGroup", http.MethodPost, "/security_groups/list", svc.ListSecurityGroup)
	h.Add("AssignSecurityGroupToBiz", http.MethodPost, "/security_groups/assign/bizs", svc.AssignSecurityGroupToBiz)

	h.Add("CreateSecurityGroupRule", http.MethodPost,
		"/vendors/{vendor}/security_groups/{security_group_id}/rules/create", svc.CreateSecurityGroupRule)
	h.Add("ListSecurityGroupRule", http.MethodPost,
		"/vendors/{vendor}/security_groups/{security_group_id}/rules/list", svc.ListSecurityGroupRule)
	h.Add("UpdateSecurityGroupRule", http.MethodPut,
		"/vendors/{vendor}/security_groups/{security_group_id}/rules/{id}", svc.UpdateSecurityGroupRule)
	h.Add("DeleteSecurityGroupRule", http.MethodDelete,
		"/vendors/{vendor}/security_groups/{security_group_id}/rules/{id}", svc.DeleteSecurityGroupRule)

	h.Load(c.WebService)
}

type securityGroupSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// GetSecurityGroup get security group.
func (svc securityGroupSvc) GetSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	baseInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		enumor.SecurityGroupCloudResType, id)
	if err != nil {
		logs.Errorf("get resource vendor failed, id: %s, err: %s, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	switch baseInfo.Vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)

	case enumor.Aws:
		return svc.client.DataService().Aws.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)

	case enumor.HuaWei:
		return svc.client.DataService().HuaWei.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)

	case enumor.Azure:
		return svc.client.DataService().Azure.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)

	default:
		return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support", id, baseInfo.Vendor)
	}
}

// UpdateSecurityGroup update security group.
func (svc securityGroupSvc) UpdateSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(proto.SecurityGroupUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	baseInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		enumor.SecurityGroupCloudResType, id)
	if err != nil {
		logs.Errorf("get resource vendor failed, id: %s, err: %s, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	updateReq := &hcproto.SecurityGroupUpdateReq{
		Name: req.Name,
		Memo: req.Memo,
	}
	switch baseInfo.Vendor {
	case enumor.TCloud:
		err = svc.client.HCService().TCloud.SecurityGroup.UpdateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
			id, updateReq)

	case enumor.HuaWei:
		err = svc.client.HCService().HuaWei.SecurityGroup.UpdateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
			id, updateReq)

	case enumor.Azure:
		err = svc.client.HCService().Azure.SecurityGroup.UpdateSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
			id, updateReq)

	default:
		return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support", id, baseInfo.Vendor)
	}
	if err != nil {
		logs.Errorf("update security group failed, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchDeleteSecurityGroup batch delete security group.
func (svc securityGroupSvc) BatchDeleteSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.SecurityGroupBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	successIDs := make([]string, 0)
	for _, id := range req.IDs {
		baseInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
			enumor.SecurityGroupCloudResType, id)
		if err != nil {
			return core.BatchDeleteResp{
				Succeeded: successIDs,
				Failed: &core.FailedInfo{
					ID:    id,
					Error: err.Error(),
				},
			}, errf.NewFromErr(errf.PartialFailed, err)
		}

		switch baseInfo.Vendor {
		case enumor.TCloud:
			err = svc.client.HCService().TCloud.SecurityGroup.DeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
		case enumor.Aws:
			err = svc.client.HCService().Aws.SecurityGroup.DeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
		case enumor.HuaWei:
			err = svc.client.HCService().HuaWei.SecurityGroup.DeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
		case enumor.Azure:
			err = svc.client.HCService().Azure.SecurityGroup.DeleteSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)
		default:
			return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support", id, baseInfo.Vendor)
		}
		if err != nil {
			return core.BatchDeleteResp{
				Succeeded: successIDs,
				Failed: &core.FailedInfo{
					ID:    id,
					Error: err.Error(),
				},
			}, errf.NewFromErr(errf.PartialFailed, err)
		}

		successIDs = append(successIDs, id)
	}

	return nil, nil
}

// ListSecurityGroup list security group.
func (svc securityGroupSvc) ListSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.SecurityGroupListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	dataReq := &dataproto.SecurityGroupListReq{
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.client.DataService().Global.SecurityGroup.ListSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(),
		dataReq)
	if err != nil {
		logs.Errorf("list security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return &proto.SecurityGroupListResult{
		Count:   result.Count,
		Details: result.Details,
	}, nil
}

// AssignSecurityGroupToBiz assign security group to biz.
func (svc securityGroupSvc) AssignSecurityGroupToBiz(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AssignSecurityGroupToBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &dataproto.SecurityGroupListReq{
		Field: []string{"id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "id",
					Op:    filter.In.Factory(),
					Value: req.SecurityGroupIDs,
				},
				&filter.AtomRule{
					Field: "bk_biz_id",
					Op:    filter.NotEqual.Factory(),
					Value: constant.UnassignedBiz,
				},
			},
		},
		Page: core.DefaultBasePage,
	}
	result, err := svc.client.DataService().Global.SecurityGroup.ListSecurityGroup(cts.Kit.Ctx,
		cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("ListSecurityGroup failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(result.Details) != 0 {
		ids := make([]string, len(result.Details))
		for index, one := range result.Details {
			ids[index] = one.ID
		}
		return nil, fmt.Errorf("security group%v already assigned", ids)
	}

	update := &dataproto.SecurityGroupCommonInfoBatchUpdateReq{
		IDs:     req.SecurityGroupIDs,
		BkBizID: req.BkBizID,
	}
	if err := svc.client.DataService().Global.SecurityGroup.BatchUpdateSecurityGroupCommonInfo(cts.Kit.Ctx,
		cts.Kit.Header(), update); err != nil {

		logs.Errorf("BatchUpdateSecurityGroupCommonInfo failed, err: %v, req: %v, rid: %s", err, update,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

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

package security_group

import (
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
	proto "hcm/pkg/api/cloud-server"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitSecurityGroupService initial the security group service
func InitSecurityGroupService(c *capability.Capability) {
	svc := &securityGroupSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()
	h.Add("GetSecurityGroup", http.MethodGet, "/security_groups/{id}", svc.GetSecurityGroup)
	h.Add("UpdateSecurityGroup", http.MethodPatch, "/security_groups/{id}", svc.UpdateSecurityGroup)
	h.Add("BatchDeleteSecurityGroup", http.MethodDelete, "/security_groups/batch", svc.BatchDeleteSecurityGroup)
	h.Add("ListSecurityGroup", http.MethodPost, "/security_groups/list", svc.ListSecurityGroup)
	h.Add("AssignSecurityGroupToBiz", http.MethodPost, "/security_groups/assign/bizs", svc.AssignSecurityGroupToBiz)

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

	vendor, err := svc.client.DataService().Global.Cloud.GetResourceVendor(cts.Kit.Ctx, cts.Kit.Header(),
		enumor.SecurityGroupCloudResType, id)
	if err != nil {
		logs.Errorf("get resource vendor failed, id: %s, err: %s, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)

	case enumor.Aws:
		return svc.client.DataService().Aws.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)

	case enumor.HuaWei:
		return svc.client.DataService().HuaWei.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)

	case enumor.Azure:
		return svc.client.DataService().Azure.SecurityGroup.GetSecurityGroup(cts.Kit.Ctx, cts.Kit.Header(), id)

	default:
		return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support get security group", id, vendor)
	}
}

// UpdateSecurityGroup update security group.
func (svc securityGroupSvc) UpdateSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	vendor, err := svc.client.DataService().Global.Cloud.GetResourceVendor(cts.Kit.Ctx, cts.Kit.Header(),
		enumor.SecurityGroupCloudResType, id)
	if err != nil {
		logs.Errorf("get resource vendor failed, id: %s, err: %s, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

}

// BatchDeleteSecurityGroup batch delete security group.
func (svc securityGroupSvc) BatchDeleteSecurityGroup(cts *rest.Contexts) (interface{}, error) {

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

	dataReq := &protocloud.SecurityGroupListReq{
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.client.DataService().Global.SecurityGroup.List(cts.Kit.Ctx, cts.Kit.Header(), dataReq)
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

}

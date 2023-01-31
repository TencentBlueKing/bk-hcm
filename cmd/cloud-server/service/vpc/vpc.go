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

package vpc

import (
	"fmt"

	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// InitVpcService initialize the vpc service.
func InitVpcService(c *capability.Capability) {
	svc := &vpcSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()

	h.Add("GetVpc", "GET", "/vpcs/{id}", svc.GetVpc)
	h.Add("ListVpc", "POST", "/vpcs/list", svc.ListVpc)
	h.Add("UpdateVpc", "PATCH", "/vpcs/{id}", svc.UpdateVpc)
	h.Add("BatchDeleteVpc", "DELETE", "/vpcs/batch", svc.BatchDeleteVpc)
	h.Add("AssignVpcToBiz", "POST", "/vpcs/assign/bizs", svc.AssignVpcToBiz)
	h.Add("BindVpcWithCloudArea", "POST", "/vpcs/bind/cloud_areas", svc.BindVpcWithCloudArea)

	h.Load(c.WebService)
}

type vpcSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// UpdateVpc update vpc.
func (svc *vpcSvc) UpdateVpc(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudserver.VpcUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		enumor.VpcCloudResType, id)
	if err != nil {
		return nil, err
	}

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Resource, Action: meta.Update,
		ResourceID: basicInfo.AccountID}}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	// update vpc
	switch basicInfo.Vendor {
	case enumor.TCloud:
		err = svc.client.HCService().TCloud.Vpc.Update(cts.Kit.Ctx, cts.Kit.Header(), id, nil)
	case enumor.Aws:
		err = svc.client.HCService().Aws.Vpc.Update(cts.Kit.Ctx, cts.Kit.Header(), id, nil)
	case enumor.Gcp:
		updateReq := &hcservice.VpcUpdateReq{
			Memo: req.Memo,
		}
		err = svc.client.HCService().Gcp.Vpc.Update(cts.Kit.Ctx, cts.Kit.Header(), id, updateReq)
	case enumor.Azure:
		err = svc.client.HCService().Azure.Vpc.Update(cts.Kit.Ctx, cts.Kit.Header(), id, nil)
	case enumor.HuaWei:
		updateReq := &hcservice.VpcUpdateReq{
			Memo: req.Memo,
		}
		err = svc.client.HCService().HuaWei.Vpc.Update(cts.Kit.Ctx, cts.Kit.Header(), id, updateReq)
	}

	if err != nil {
		return nil, err
	}

	return nil, nil
}

// GetVpc get vpc details.
func (svc *vpcSvc) GetVpc(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		enumor.VpcCloudResType, id)
	if err != nil {
		return nil, err
	}

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Resource, Action: meta.Find,
		ResourceID: basicInfo.AccountID}}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	// get vpc detail info
	switch basicInfo.Vendor {
	case enumor.TCloud:
		vpc, err := svc.client.DataService().TCloud.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}
		return vpc, err
	case enumor.Aws:
		vpc, err := svc.client.DataService().Aws.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}
		return vpc, err
	case enumor.Gcp:
		vpc, err := svc.client.DataService().Gcp.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}
		return vpc, err
	case enumor.HuaWei:
		vpc, err := svc.client.DataService().HuaWei.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}
		return vpc, err
	case enumor.Azure:
		vpc, err := svc.client.DataService().Azure.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}
		return vpc, err
	}

	return nil, nil
}

// ListVpc list vpcs.
func (svc *vpcSvc) ListVpc(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	authOpt := &meta.ListAuthResInput{Type: meta.Resource, Action: meta.Find}
	authInst, err := svc.authorizer.ListAuthorizedInstances(cts.Kit, authOpt)
	if err != nil {
		return nil, err
	}

	if !authInst.IsAny {
		if len(authInst.IDs) == 0 {
			return &cloudserver.VpcListResult{Count: 0, Details: make([]corecloud.BaseVpc, 0)}, nil
		}
		// TODO add account id filter
		//req.Filter.
	}

	// list vpcs
	res, err := svc.client.DataService().Global.Vpc.List(cts.Kit.Ctx, cts.Kit.Header(), req)
	if err != nil {
		return nil, err
	}

	return &cloudserver.VpcListResult{Count: res.Count, Details: res.Details}, nil
}

// BatchDeleteVpc batch delete vpcs.
func (svc *vpcSvc) BatchDeleteVpc(cts *rest.Contexts) (interface{}, error) {
	req := new(core.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.VpcCloudResType,
		IDs:          req.IDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	// authorize
	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Resource, Action: meta.Delete,
			ResourceID: info.AccountID}})
	}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
	if err != nil {
		return nil, err
	}

	// delete vpcs
	succeeded := make([]string, 0)
	for _, id := range req.IDs {
		basicInfo, exists := basicInfoMap[id]
		if !exists {
			return nil, errf.New(errf.InvalidParameter, fmt.Sprintf("id %s has no corresponding vendor", id))
		}

		switch basicInfo.Vendor {
		case enumor.TCloud:
			err = svc.client.HCService().TCloud.Vpc.Delete(cts.Kit.Ctx, cts.Kit.Header(), id)
		case enumor.Aws:
			err = svc.client.HCService().Aws.Vpc.Delete(cts.Kit.Ctx, cts.Kit.Header(), id)
		case enumor.Gcp:
			err = svc.client.HCService().Gcp.Vpc.Delete(cts.Kit.Ctx, cts.Kit.Header(), id)
		case enumor.Azure:
			err = svc.client.HCService().Azure.Vpc.Delete(cts.Kit.Ctx, cts.Kit.Header(), id)
		case enumor.HuaWei:
			err = svc.client.HCService().HuaWei.Vpc.Delete(cts.Kit.Ctx, cts.Kit.Header(), id)
		}

		if err != nil {
			return core.BatchDeleteResp{
				Succeeded: succeeded,
				Failed: &core.FailedInfo{
					ID:    id,
					Error: err.Error(),
				},
			}, errf.NewFromErr(errf.PartialFailed, err)
		}

		succeeded = append(succeeded, id)
	}

	return core.BatchDeleteResp{Succeeded: succeeded}, nil
}

// AssignVpcToBiz assign vpcs to biz.
func (svc *vpcSvc) AssignVpcToBiz(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudserver.AssignVpcToBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize
	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.VpcCloudResType,
		IDs:          req.VpcIDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Resource, Action: meta.Assign,
			ResourceID: info.AccountID}})
	}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
	if err != nil {
		return nil, err
	}

	// check if all vpcs are not assigned
	assignedReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.VpcIDs},
				&filter.AtomRule{Field: "bk_biz_id", Op: filter.NotEqual.Factory(), Value: constant.UnassignedBiz},
			},
		},
		Page: &core.BasePage{
			Count: true,
		},
	}
	result, err := svc.client.DataService().Global.Vpc.List(cts.Kit.Ctx, cts.Kit.Header(), assignedReq)
	if err != nil {
		logs.Errorf("count assigned vpc failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if result.Count != 0 {
		return nil, fmt.Errorf("%d vpcs are already assigned", result.Count)
	}

	// update vpc biz relations
	createReq := &cloud.VpcBaseInfoBatchUpdateReq{
		Vpcs: []cloud.VpcBaseInfoUpdateReq{{
			IDs: req.VpcIDs,
			Data: &cloud.VpcUpdateBaseInfo{
				BkBizID: req.BkBizID,
			},
		}},
	}

	err = svc.client.DataService().Global.Vpc.BatchUpdateBaseInfo(cts.Kit.Ctx, cts.Kit.Header(), createReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// BindVpcWithCloudArea bind vpcs with cloud areas.
func (svc *vpcSvc) BindVpcWithCloudArea(cts *rest.Contexts) (interface{}, error) {
	req := make(cloudserver.BindVpcWithCloudAreaReq, 0)
	if err := cts.DecodeInto(&req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0, len(req))
	for _, rel := range req {
		ids = append(ids, rel.VpcID)
	}

	// authorize
	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.VpcCloudResType,
		IDs:          ids,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	vpcIDs := make([]string, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Resource, Action: meta.Assign,
			ResourceID: info.AccountID}})
		vpcIDs = append(vpcIDs, info.ID)
	}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
	if err != nil {
		return nil, err
	}

	// check if all vpcs are not assigned
	assignedReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: vpcIDs},
				&filter.AtomRule{Field: "bk_cloud_id", Op: filter.NotEqual.Factory(), Value: constant.UnbindBkCloudID},
			},
		},
		Page: &core.BasePage{
			Count: true,
		},
	}
	result, err := svc.client.DataService().Global.Vpc.List(cts.Kit.Ctx, cts.Kit.Header(), assignedReq)
	if err != nil {
		logs.Errorf("count assigned vpc failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if result.Count != 0 {
		return nil, fmt.Errorf("%d vpcs are already assigned", result.Count)
	}

	// update vpc cloud area relations
	updateReqs := make([]cloud.VpcBaseInfoUpdateReq, 0, len(req))
	for _, rel := range req {
		updateReqs = append(updateReqs, cloud.VpcBaseInfoUpdateReq{
			IDs:  []string{rel.VpcID},
			Data: &cloud.VpcUpdateBaseInfo{BkCloudID: rel.BkCloudID},
		})
	}

	batchUpdateReq := &cloud.VpcBaseInfoBatchUpdateReq{
		Vpcs: updateReqs,
	}

	err = svc.client.DataService().Global.Vpc.BatchUpdateBaseInfo(cts.Kit.Ctx, cts.Kit.Header(), batchUpdateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

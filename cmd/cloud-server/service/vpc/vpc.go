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

	"hcm/cmd/cloud-server/logics/audit"
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
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// InitVpcService initialize the vpc service.
func InitVpcService(c *capability.Capability) {
	svc := &vpcSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	h.Add("GetVpc", "GET", "/vpcs/{id}", svc.GetVpc)
	h.Add("ListVpc", "POST", "/vpcs/list", svc.ListVpc)
	h.Add("UpdateVpc", "PATCH", "/vpcs/{id}", svc.UpdateVpc)
	h.Add("BatchDeleteVpc", "DELETE", "/vpcs/batch", svc.BatchDeleteVpc)
	h.Add("AssignVpcToBiz", "POST", "/vpcs/assign/bizs", svc.AssignVpcToBiz)
	h.Add("BindVpcWithCloudArea", "POST", "/vpcs/bind/cloud_areas", svc.BindVpcWithCloudArea)

	// vpc apis in biz
	h.Add("GetBizVpc", "GET", "/bizs/{bk_biz_id}/vpcs/{id}", svc.GetBizVpc)
	h.Add("ListBizVpc", "POST", "/bizs/{bk_biz_id}/vpcs/list", svc.ListBizVpc)
	h.Add("UpdateBizVpc", "PATCH", "/bizs/{bk_biz_id}/vpcs/{id}", svc.UpdateBizVpc)
	h.Add("BatchDeleteBizVpc", "DELETE", "/bizs/{bk_biz_id}/vpcs/batch", svc.BatchDeleteBizVpc)

	h.Load(c.WebService)
}

type vpcSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

// UpdateVpc update vpc.
func (svc *vpcSvc) UpdateVpc(cts *rest.Contexts) (interface{}, error) {
	return svc.updateVpc(cts, handler.ResValidWithAuth)
}

// UpdateBizVpc update biz vpc.
func (svc *vpcSvc) UpdateBizVpc(cts *rest.Contexts) (interface{}, error) {
	return svc.updateVpc(cts, handler.BizValidWithAuth)
}

func (svc *vpcSvc) updateVpc(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
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

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Vpc,
		Action: meta.Update, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	// create update audit.
	updateFields, err := converter.StructToMap(req)
	if err != nil {
		logs.Errorf("convert request to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err = svc.audit.ResUpdateAudit(cts.Kit, enumor.VpcCloudAuditResType, id, updateFields); err != nil {
		logs.Errorf("create update audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// update vpc
	updateReq := &hcservice.VpcUpdateReq{
		Memo: req.Memo,
	}
	switch basicInfo.Vendor {
	case enumor.TCloud:
		err = svc.client.HCService().TCloud.Vpc.Update(cts.Kit.Ctx, cts.Kit.Header(), id, updateReq)
	case enumor.Aws:
		err = svc.client.HCService().Aws.Vpc.Update(cts.Kit.Ctx, cts.Kit.Header(), id, updateReq)
	case enumor.Gcp:
		err = svc.client.HCService().Gcp.Vpc.Update(cts.Kit.Ctx, cts.Kit.Header(), id, updateReq)
	case enumor.Azure:
		err = svc.client.HCService().Azure.Vpc.Update(cts.Kit.Ctx, cts.Kit.Header(), id, updateReq)
	case enumor.HuaWei:
		err = svc.client.HCService().HuaWei.Vpc.Update(cts.Kit.Ctx, cts.Kit.Header(), id, updateReq)
	}

	if err != nil {
		return nil, err
	}

	return nil, nil
}

// GetVpc get vpc details.
func (svc *vpcSvc) GetVpc(cts *rest.Contexts) (interface{}, error) {
	return svc.getVpc(cts, handler.ResValidWithAuth)
}

// GetBizVpc get biz vpc details.
func (svc *vpcSvc) GetBizVpc(cts *rest.Contexts) (interface{}, error) {
	return svc.getVpc(cts, handler.BizValidWithAuth)
}

func (svc *vpcSvc) getVpc(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		enumor.VpcCloudResType, id)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Vpc,
		Action: meta.Find, BasicInfo: basicInfo})
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

// ListVpc list vpc.
func (svc *vpcSvc) ListVpc(cts *rest.Contexts) (interface{}, error) {
	return svc.listVpc(cts, handler.ListResourceAuthRes)
}

// ListBizVpc list biz vpc.
func (svc *vpcSvc) ListBizVpc(cts *rest.Contexts) (interface{}, error) {
	return svc.listVpc(cts, handler.ListBizAuthRes)
}

func (svc *vpcSvc) listVpc(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer, ResType: meta.Vpc,
		Action: meta.Find, Filter: req.Filter})
	if noPermFlag {
		return &cloudserver.VpcListResult{Count: 0, Details: make([]corecloud.BaseVpc, 0)}, nil
	}
	req.Filter = expr

	// list vpcs
	res, err := svc.client.DataService().Global.Vpc.List(cts.Kit.Ctx, cts.Kit.Header(), req)
	if err != nil {
		return nil, err
	}

	return &cloudserver.VpcListResult{Count: res.Count, Details: res.Details}, nil
}

// BatchDeleteVpc batch delete vpc.
func (svc *vpcSvc) BatchDeleteVpc(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteVpc(cts, handler.ResValidWithAuth)
}

// BatchDeleteBizVpc batch delete biz vpc.
func (svc *vpcSvc) BatchDeleteBizVpc(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteVpc(cts, handler.BizValidWithAuth)
}

func (svc *vpcSvc) batchDeleteVpc(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
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

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Vpc,
		Action: meta.Find, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	// create delete audit.
	if err := svc.audit.ResDeleteAudit(cts.Kit, enumor.VpcCloudAuditResType, req.IDs); err != nil {
		logs.Errorf("create delete audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
	err := svc.authorizeVpcBatchOp(cts.Kit, meta.Assign, req.VpcIDs, req.BkBizID)
	if err != nil {
		return nil, err
	}

	// check if all vpcs are not assigned to biz, right now assigning resource twice is not allowed
	vpcFilter := &filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.VpcIDs}
	err = svc.checkVpcsInBiz(cts.Kit, vpcFilter, constant.UnassignedBiz)
	if err != nil {
		return nil, err
	}

	// create assign audit.
	err = svc.audit.ResBizAssignAudit(cts.Kit, enumor.VpcCloudAuditResType, req.VpcIDs, req.BkBizID)
	if err != nil {
		logs.Errorf("create assign audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
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

	// authorize
	ids := make([]string, 0, len(req))
	for _, rel := range req {
		ids = append(ids, rel.VpcID)
	}

	err := svc.authorizeVpcBatchOp(cts.Kit, meta.Update, ids, 0)
	if err != nil {
		return nil, err
	}

	// check if all vpcs are not assigned to biz, cannot operate biz resource in account api
	vpcFilter := &filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: ids}
	err = svc.checkVpcsInBiz(cts.Kit, vpcFilter, constant.UnassignedBiz)
	if err != nil {
		return nil, err
	}

	// check if all vpcs are not assigned
	assignedReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: ids},
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

	// create assign audit.
	auditOpt := make([]audit.ResCloudAreaAssignOption, 0, len(req))
	for _, info := range req {
		auditOpt = append(auditOpt, audit.ResCloudAreaAssignOption{
			ResID:   info.VpcID,
			CloudID: info.BkCloudID,
		})
	}
	err = svc.audit.ResCloudAreaAssignAudit(cts.Kit, enumor.VpcCloudAuditResType, auditOpt)
	if err != nil {
		logs.Errorf("create assign audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
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

func (svc *vpcSvc) authorizeVpcBatchOp(kt *kit.Kit, action meta.Action, ids []string, bizID int64) error {
	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.VpcCloudResType,
		IDs:          ids,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(kt.Ctx, kt.Header(), basicInfoReq)
	if err != nil {
		return err
	}

	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Vpc, Action: action,
			ResourceID: info.AccountID}, BizID: bizID})
	}
	err = svc.authorizer.AuthorizeWithPerm(kt, authRes...)
	if err != nil {
		return err
	}

	return nil
}

// checkVpcsInBiz check if vpcs are in the specified biz.
func (svc *vpcSvc) checkVpcsInBiz(kt *kit.Kit, rule filter.RuleFactory, bizID int64) error {
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
	result, err := svc.client.DataService().Global.Vpc.List(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("count vpcs that are not in biz failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return err
	}

	if result.Count != 0 {
		return fmt.Errorf("%d vpcs are already assigned", result.Count)
	}

	return nil
}

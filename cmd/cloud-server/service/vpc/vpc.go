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

// Package vpc ...
package vpc

import (
	"encoding/json"
	"fmt"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/cmd/cloud-server/service/common"
	cloudserver "hcm/pkg/api/cloud-server"
	csvpc "hcm/pkg/api/cloud-server/vpc"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service/vpc"
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
	h.Add("CreateVpc", "POST", "/vpcs/create", svc.CreateVpc)
	h.Add("UpdateVpc", "PATCH", "/vpcs/{id}", svc.UpdateVpc)
	h.Add("DeleteVpc", "DELETE", "/vpcs/{id}", svc.DeleteVpc)
	h.Add("AssignVpcToBiz", "POST", "/vpcs/assign/bizs", svc.AssignVpcToBiz)
	h.Add("ListResVpcExt", "POST", "/vendors/{vendor}/vpcs/list", svc.ListResVpcExt)

	// vpc apis in biz
	h.Add("GetBizVpc", "GET", "/bizs/{bk_biz_id}/vpcs/{id}", svc.GetBizVpc)
	h.Add("ListBizVpc", "POST", "/bizs/{bk_biz_id}/vpcs/list", svc.ListBizVpc)
	h.Add("ListBizVpcExt", "POST", "/bizs/{bk_biz_id}/vendors/{vendor}/vpcs/list", svc.ListBizVpcExt)
	h.Add("UpdateBizVpc", "PATCH", "/bizs/{bk_biz_id}/vpcs/{id}", svc.UpdateBizVpc)
	h.Add("DeleteBizVpc", "DELETE", "/bizs/{bk_biz_id}/vpcs/{id}", svc.DeleteBizVpc)

	h.Load(c.WebService)
}

type vpcSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

// CreateVpc create vpc.
func (svc *vpcSvc) CreateVpc(cts *rest.Contexts) (interface{}, error) {

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create vpc request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Vpc, Action: meta.Create,
		ResourceID: req.AccountID}}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("create vpc auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 获取资源公共参数，如厂商
	info, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 根据厂商信息转到下方具体的实现
	switch info.Vendor {
	case enumor.TCloud:
		return svc.createTCloudVpc(cts.Kit, req.Data)
	case enumor.Aws:
		return svc.createAwsVpc(cts.Kit, req.Data)
	case enumor.HuaWei:
		return svc.createHuaWeiVpc(cts.Kit, req.Data)
	case enumor.Gcp:
		return svc.createGcpVpc(cts.Kit, req.Data)
	case enumor.Azure:
		return svc.createAzureVpc(cts.Kit, req.Data)
	default:
		return nil, fmt.Errorf("vendor: %s not support", info.Vendor)
	}

}

func (svc *vpcSvc) createTCloudVpc(kt *kit.Kit, body json.RawMessage) (interface{}, error) {

	req := new(csvpc.TCloudVpcCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	// 校验参数，不要求业务id
	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	// 转换参数并调用HCService进行创建流程
	result, err := svc.client.HCService().TCloud.Vpc.Create(kt.Ctx, kt.Header(), common.ConvTCloudVpcCreateReq(req))
	if err != nil {
		logs.Errorf("batch create tcloud vpc failed, err: %v, result: %v, rid: %s", err, result, kt.Rid)
		return result, err
	}

	return result, nil
}

func (svc *vpcSvc) createAzureVpc(kt *kit.Kit, body json.RawMessage) (interface{}, error) {

	req := new(csvpc.AzureVpcCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().Azure.Vpc.Create(kt.Ctx, kt.Header(), common.ConvAzureVpcCreateReq(req))
	if err != nil {
		logs.Errorf("batch create azure vpc failed, err: %v, result: %v, rid: %s", err, result, kt.Rid)
		return result, err
	}

	return result, nil
}

func (svc *vpcSvc) createHuaWeiVpc(kt *kit.Kit, body json.RawMessage) (interface{}, error) {

	req := new(csvpc.HuaWeiVpcCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().HuaWei.Vpc.Create(kt.Ctx, kt.Header(),
		common.ConvHuaWeiVpcCreateReq(req))
	if err != nil {
		logs.Errorf("batch create huawei vpc failed, err: %v, result: %v, rid: %s", err, result, kt.Rid)
		return result, err
	}

	return result, nil
}

func (svc *vpcSvc) createGcpVpc(kt *kit.Kit, body json.RawMessage) (interface{}, error) {

	req := new(csvpc.GcpVpcCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().Gcp.Vpc.Create(kt.Ctx, kt.Header(), common.ConvGcpVpcCreateReq(req))
	if err != nil {
		logs.Errorf("batch create gcp vpc failed, err: %v, result: %v, rid: %s", err, result, kt.Rid)
		return result, err
	}

	return result, nil
}

func (svc *vpcSvc) createAwsVpc(kt *kit.Kit, body json.RawMessage) (interface{}, error) {

	req := new(csvpc.AwsVpcCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().Aws.Vpc.Create(kt.Ctx, kt.Header(), common.ConvAwsVpcCreateReq(req))
	if err != nil {
		logs.Errorf("batch create aws vpc failed, err: %v, result: %v, rid: %s", err, result, kt.Rid)
		return result, err
	}

	return result, nil
}

// UpdateVpc update vpc.
func (svc *vpcSvc) UpdateVpc(cts *rest.Contexts) (interface{}, error) {
	return svc.updateVpc(cts, handler.ResOperateAuth)
}

// UpdateBizVpc update biz vpc.
func (svc *vpcSvc) UpdateBizVpc(cts *rest.Contexts) (interface{}, error) {
	return svc.updateVpc(cts, handler.BizOperateAuth)
}

func (svc *vpcSvc) updateVpc(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	req := new(csvpc.VpcUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
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
	return svc.getVpc(cts, handler.ResOperateAuth)
}

// GetBizVpc get biz vpc details.
func (svc *vpcSvc) GetBizVpc(cts *rest.Contexts) (interface{}, error) {
	return svc.getVpc(cts, handler.BizOperateAuth)
}

func (svc *vpcSvc) getVpc(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.VpcCloudResType, id)
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
		vpc, err := svc.client.DataService().Azure.Vpc.Get(cts.Kit, id)
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
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &csvpc.VpcListResult{Count: 0, Details: make([]corecloud.BaseVpc, 0)}, nil
	}

	req.Filter = expr

	// list vpcs
	res, err := svc.client.DataService().Global.Vpc.List(cts.Kit.Ctx, cts.Kit.Header(), req)
	if err != nil {
		return nil, err
	}

	return &csvpc.VpcListResult{Count: res.Count, Details: res.Details}, nil
}

// ListResVpcExt list resource vpc with extension.
func (svc *vpcSvc) ListResVpcExt(cts *rest.Contexts) (interface{}, error) {
	return svc.listVpcExt(cts, handler.ListResourceAuthRes)
}

// ListBizVpcExt list biz vpc with extension.
func (svc *vpcSvc) ListBizVpcExt(cts *rest.Contexts) (interface{}, error) {
	return svc.listVpcExt(cts, handler.ListBizAuthRes)
}

func (svc *vpcSvc) listVpcExt(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())

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
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &csvpc.VpcListResult{Count: 0, Details: make([]corecloud.BaseVpc, 0)}, nil
	}
	req.Filter = expr

	switch vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.Vpc.ListVpcExt(cts.Kit.Ctx, cts.Kit.Header(), req)
	case enumor.Aws:
		return svc.client.DataService().Aws.Vpc.ListVpcExt(cts.Kit.Ctx, cts.Kit.Header(), req)
	case enumor.Gcp:
		return svc.client.DataService().Gcp.Vpc.ListVpcExt(cts.Kit, req)
	case enumor.Azure:
		return svc.client.DataService().Azure.Vpc.ListVpcExt(cts.Kit.Ctx, cts.Kit.Header(), req)
	case enumor.HuaWei:
		return svc.client.DataService().HuaWei.Vpc.ListVpcExt(cts.Kit.Ctx, cts.Kit.Header(), req)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "vendor: %s not support", vendor)
	}
}

// DeleteVpc delete vpc.
func (svc *vpcSvc) DeleteVpc(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteVpc(cts, handler.ResOperateAuth)
}

// DeleteBizVpc delete biz vpc.
func (svc *vpcSvc) DeleteBizVpc(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteVpc(cts, handler.BizOperateAuth)
}

func (svc *vpcSvc) deleteVpc(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.VpcCloudResType, id)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Vpc,
		Action: meta.Delete, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	// create delete audit.
	if err := svc.audit.ResDeleteAudit(cts.Kit, enumor.VpcCloudAuditResType, []string{id}); err != nil {
		logs.Errorf("create delete audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// delete vpcs
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
		logs.Errorf("delete vpc %s failed, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// AssignVpcToBiz assign vpcs to biz.
func (svc *vpcSvc) AssignVpcToBiz(cts *rest.Contexts) (interface{}, error) {
	req := new(csvpc.AssignVpcToBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := common.ValidateTargetBizID(cts.Kit, svc.client.DataService(), enumor.VpcCloudResType, req.VpcIDs, req.BkBizID)
	if err != nil {
		return nil, err
	}

	// authorize
	err = svc.authorizeVpcBatchOp(cts.Kit, meta.Assign, req.VpcIDs, req.BkBizID)
	if err != nil {
		return nil, err
	}

	// check if all vpcs are not assigned to biz, right now assigning resource twice is not allowed
	listReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.VpcIDs},
				&filter.AtomRule{Field: "bk_biz_id", Op: filter.Equal.Factory(), Value: constant.UnassignedBiz},
			},
		},
		Page: core.NewCountPage(),
	}
	result, err := svc.client.DataService().Global.Vpc.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("count vpcs that are not in biz failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	if int(result.Count) != len(req.VpcIDs) {
		return nil, fmt.Errorf("%d vpcs are already assigned biz or unBind cloud area",
			len(req.VpcIDs)-int(result.Count))
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

func (svc *vpcSvc) authorizeVpcBatchOp(kt *kit.Kit, action meta.Action, ids []string, bizID int64) error {
	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.VpcCloudResType,
		IDs:          ids,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(kt, basicInfoReq)
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

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

// Package subnet ...
package subnet

import (
	"encoding/json"
	"fmt"

	"hcm/cmd/cloud-server/logics/async"
	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/cmd/cloud-server/service/common"
	actionsubnet "hcm/cmd/task-server/logics/action/subnet"
	cloudserver "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service/subnet"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
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
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/uuid"
)

// InitSubnetService initialize the subnet service.
func InitSubnetService(c *capability.Capability) {
	svc := &subnetSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	h.Add("CreateSubnet", "POST", "/subnets/create", svc.CreateSubnet)
	h.Add("GetSubnet", "GET", "/subnets/{id}", svc.GetSubnet)
	h.Add("ListSubnet", "POST", "/subnets/list", svc.ListSubnet)
	h.Add("UpdateSubnet", "PATCH", "/subnets/{id}", svc.UpdateSubnet)
	h.Add("BatchDeleteSubnet", "DELETE", "/subnets/batch", svc.BatchDeleteSubnet)
	h.Add("AssignSubnetToBiz", "POST", "/subnets/assign/bizs", svc.AssignSubnetToBiz)
	h.Add("CountSubnetAvailableIPs", "POST", "/subnets/{id}/ips/count", svc.CountSubnetAvailableIPs)
	h.Add("ListCountResSubnetAvailIPs", "POST", "/subnets/ips/count/list",
		svc.ListCountResSubnetAvailIPs)

	// subnet apis in biz
	h.Add("CreateBizSubnet", "POST", "/bizs/{bk_biz_id}/subnets/create", svc.CreateBizSubnet)
	h.Add("GetBizSubnet", "GET", "/bizs/{bk_biz_id}/subnets/{id}", svc.GetBizSubnet)
	h.Add("ListBizSubnet", "POST", "/bizs/{bk_biz_id}/subnets/list", svc.ListBizSubnet)
	h.Add("UpdateBizSubnet", "PATCH", "/bizs/{bk_biz_id}/subnets/{id}", svc.UpdateBizSubnet)
	h.Add("BatchDeleteBizSubnet", "DELETE", "/bizs/{bk_biz_id}/subnets/batch", svc.BatchDeleteBizSubnet)
	h.Add("CountBizSubnetAvailIPs", "POST", "/bizs/{bk_biz_id}/subnets/{id}/ips/count", svc.CountBizSubnetAvailIPs)
	h.Add("ListCountBizSubnetAvailIPs", "POST", "/bizs/{bk_biz_id}/subnets/ips/count/list",
		svc.ListCountBizSubnetAvailIPs)

	h.Load(c.WebService)
}

type subnetSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

// CreateSubnet create subnet.
func (svc *subnetSvc) CreateSubnet(cts *rest.Contexts) (interface{}, error) {
	bizID := int64(constant.UnassignedBiz)
	return svc.createSubnet(cts, bizID, handler.ResOperateAuth)
}

// CreateBizSubnet create biz subnet.
func (svc *subnetSvc) CreateBizSubnet(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	return svc.createSubnet(cts, bizID, handler.BizOperateAuth)
}

func (svc *subnetSvc) createSubnet(cts *rest.Contexts, bizID int64,
	validHandler handler.ValidWithAuthHandler) (interface{}, error) {

	accountID, err := common.ExtractAccountID(cts)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// validate authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Subnet,
		Action: meta.Create, BasicInfo: common.GetCloudResourceBasicInfo(accountID, bizID)})
	if err != nil {
		return nil, err
	}

	req := new(cloudserver.RawCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	switch req.Vendor {
	case enumor.TCloud:
		return svc.createTCloudSubnet(cts.Kit, bizID, req.Data)
	case enumor.Aws:
		return svc.createAwsSubnet(cts.Kit, bizID, req.Data)
	case enumor.Gcp:
		return svc.createGcpSubnet(cts.Kit, bizID, req.Data)
	case enumor.Azure:
		return svc.createAzureSubnet(cts.Kit, bizID, req.Data)
	case enumor.HuaWei:
		return svc.createHuaWeiSubnet(cts.Kit, bizID, req.Data)
	}
	return nil, nil
}

// createTCloudSubnet create tcloud subnet.
func (svc *subnetSvc) createTCloudSubnet(kt *kit.Kit, bizID int64, data json.RawMessage) (
	interface{}, error) {

	req := new(cloudserver.TCloudSubnetCreateReq)
	if err := json.Unmarshal(data, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &hcservice.TCloudSubnetBatchCreateReq{
		BkBizID:    bizID,
		AccountID:  req.AccountID,
		Region:     req.Region,
		CloudVpcID: req.CloudVpcID,

		Subnets: []hcservice.TCloudOneSubnetCreateReq{{
			IPv4Cidr:          req.IPv4Cidr,
			Name:              req.Name,
			Zone:              req.Zone,
			CloudRouteTableID: req.CloudRouteTableID,
			Memo:              req.Memo,
		}},
	}
	createRes, err := svc.client.HCService().TCloud.Subnet.BatchCreate(kt.Ctx, kt.Header(), opt)
	if err != nil {
		return nil, err
	}

	return core.CreateResult{ID: createRes.IDs[0]}, nil
}

// createAwsSubnet create aws subnet.
func (svc *subnetSvc) createAwsSubnet(kt *kit.Kit, bizID int64, data json.RawMessage) (
	interface{}, error) {

	req := new(cloudserver.AwsSubnetCreateReq)
	if err := json.Unmarshal(data, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &hcservice.SubnetCreateReq[hcservice.AwsSubnetCreateExt]{
		BaseSubnetCreateReq: convertBaseSubnetCreateReq(bizID, req.BaseSubnetCreateReq),
		Extension: &hcservice.AwsSubnetCreateExt{
			Region:   req.Region,
			Zone:     req.Zone,
			IPv4Cidr: req.IPv4Cidr,
			IPv6Cidr: req.IPv6Cidr,
		},
	}
	createRes, err := svc.client.HCService().Aws.Subnet.Create(kt.Ctx, kt.Header(), opt)
	if err != nil {
		return nil, err
	}

	return createRes, nil
}

// createGcpSubnet create gcp subnet.
func (svc *subnetSvc) createGcpSubnet(kt *kit.Kit, bizID int64, data json.RawMessage) (
	interface{}, error) {

	req := new(cloudserver.GcpSubnetCreateReq)
	if err := json.Unmarshal(data, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &hcservice.SubnetCreateReq[hcservice.GcpSubnetCreateExt]{
		BaseSubnetCreateReq: convertBaseSubnetCreateReq(bizID, req.BaseSubnetCreateReq),
		Extension: &hcservice.GcpSubnetCreateExt{
			Region:                req.Region,
			IPv4Cidr:              req.IPv4Cidr,
			PrivateIpGoogleAccess: req.PrivateIpGoogleAccess,
			EnableFlowLogs:        req.EnableFlowLogs,
		},
	}
	createRes, err := svc.client.HCService().Gcp.Subnet.Create(kt.Ctx, kt.Header(), opt)
	if err != nil {
		return nil, err
	}

	return createRes, nil
}

// createAzureSubnet create azure subnet.
func (svc *subnetSvc) createAzureSubnet(kt *kit.Kit, bizID int64, data json.RawMessage) (
	interface{}, error) {

	req := new(cloudserver.AzureSubnetCreateReq)
	if err := json.Unmarshal(data, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// check azure subnet params
	if err := svc.checkAzureSubnetParams(req); err != nil {
		return nil, err
	}

	opt := &hcservice.SubnetCreateReq[hcservice.AzureSubnetCreateExt]{
		BaseSubnetCreateReq: convertBaseSubnetCreateReq(bizID, req.BaseSubnetCreateReq),
		Extension: &hcservice.AzureSubnetCreateExt{
			ResourceGroup:        req.ResourceGroup,
			IPv4Cidr:             req.IPv4Cidr,
			IPv6Cidr:             req.IPv6Cidr,
			CloudRouteTableID:    req.CloudRouteTableID,
			NatGateway:           req.NatGateway,
			NetworkSecurityGroup: req.NetworkSecurityGroup,
		},
	}
	createRes, err := svc.client.HCService().Azure.Subnet.Create(kt.Ctx, kt.Header(), opt)
	if err != nil {
		return nil, err
	}

	return createRes, nil
}

// createHuaWeiSubnet create huawei subnet.
func (svc *subnetSvc) createHuaWeiSubnet(kt *kit.Kit, bizID int64, data json.RawMessage) (
	interface{}, error) {

	req := new(cloudserver.HuaWeiSubnetCreateReq)
	if err := json.Unmarshal(data, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &hcservice.SubnetCreateReq[hcservice.HuaWeiSubnetCreateExt]{
		BaseSubnetCreateReq: convertBaseSubnetCreateReq(bizID, req.BaseSubnetCreateReq),
		Extension: &hcservice.HuaWeiSubnetCreateExt{
			Region:     req.Region,
			Zone:       req.Zone,
			IPv4Cidr:   req.IPv4Cidr,
			Ipv6Enable: req.Ipv6Enable,
			GatewayIp:  req.GatewayIp,
		},
	}
	createRes, err := svc.client.HCService().HuaWei.Subnet.Create(kt.Ctx, kt.Header(), opt)
	if err != nil {
		return nil, err
	}

	return createRes, nil
}

// convertBaseSubnetCreateReq convert base subnet create request.
func convertBaseSubnetCreateReq(bizID int64, req *cloudserver.BaseSubnetCreateReq) *hcservice.BaseSubnetCreateReq {
	return &hcservice.BaseSubnetCreateReq{
		AccountID:  req.AccountID,
		Name:       req.Name,
		Memo:       req.Memo,
		CloudVpcID: req.CloudVpcID,
		BkBizID:    bizID,
	}
}

// UpdateSubnet update subnet.
func (svc *subnetSvc) UpdateSubnet(cts *rest.Contexts) (interface{}, error) {
	return svc.updateSubnet(cts, handler.ResOperateAuth)
}

// UpdateBizSubnet update biz subnet.
func (svc *subnetSvc) UpdateBizSubnet(cts *rest.Contexts) (interface{}, error) {
	return svc.updateSubnet(cts, handler.BizOperateAuth)
}

func (svc *subnetSvc) updateSubnet(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	req := new(cloudserver.SubnetUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.SubnetCloudResType, id)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Subnet,
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
	if err = svc.audit.ResUpdateAudit(cts.Kit, enumor.SubnetAuditResType, id, updateFields); err != nil {
		logs.Errorf("create update audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// update subnet
	updateReq := &hcservice.SubnetUpdateReq{
		Memo: req.Memo,
	}
	switch basicInfo.Vendor {
	case enumor.TCloud:
		err = svc.client.HCService().TCloud.Subnet.Update(cts.Kit.Ctx, cts.Kit.Header(), id, updateReq)
	case enumor.Aws:
		err = svc.client.HCService().Aws.Subnet.Update(cts.Kit.Ctx, cts.Kit.Header(), id, updateReq)
	case enumor.Gcp:
		err = svc.client.HCService().Gcp.Subnet.Update(cts.Kit.Ctx, cts.Kit.Header(), id, updateReq)
	case enumor.Azure:
		err = svc.client.HCService().Azure.Subnet.Update(cts.Kit.Ctx, cts.Kit.Header(), id, updateReq)
	case enumor.HuaWei:
		err = svc.client.HCService().HuaWei.Subnet.Update(cts.Kit.Ctx, cts.Kit.Header(), id, updateReq)
	}

	if err != nil {
		return nil, err
	}

	return nil, nil
}

// GetSubnet get subnet details.
func (svc *subnetSvc) GetSubnet(cts *rest.Contexts) (interface{}, error) {
	return svc.getSubnet(cts, handler.ResOperateAuth)
}

// GetBizSubnet get biz subnet details.
func (svc *subnetSvc) GetBizSubnet(cts *rest.Contexts) (interface{}, error) {
	return svc.getSubnet(cts, handler.BizOperateAuth)
}

func (svc *subnetSvc) getSubnet(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.SubnetCloudResType, id)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Subnet,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	// get subnet detail info
	switch basicInfo.Vendor {
	case enumor.TCloud:
		subnet, err := svc.client.DataService().TCloud.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}
		return subnet, err
	case enumor.Aws:
		subnet, err := svc.client.DataService().Aws.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}
		return subnet, err
	case enumor.Gcp:
		subnet, err := svc.client.DataService().Gcp.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}
		return subnet, err
	case enumor.HuaWei:
		subnet, err := svc.client.DataService().HuaWei.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}
		return subnet, err
	case enumor.Azure:
		subnet, err := svc.client.DataService().Azure.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}
		return subnet, err
	}

	return nil, nil
}

// ListSubnet list subnet.
func (svc *subnetSvc) ListSubnet(cts *rest.Contexts) (interface{}, error) {
	return svc.listSubnet(cts, handler.ListResourceAuthRes)
}

// ListBizSubnet list biz subnet.
func (svc *subnetSvc) ListBizSubnet(cts *rest.Contexts) (interface{}, error) {
	return svc.listSubnet(cts, handler.ListBizAuthRes)
}

func (svc *subnetSvc) listSubnet(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.Subnet, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &cloudserver.SubnetListResult{Count: 0, Details: make([]corecloud.BaseSubnet, 0)}, nil
	}
	req.Filter = expr

	// list subnets
	res, err := svc.client.DataService().Global.Subnet.List(cts.Kit.Ctx, cts.Kit.Header(), req)
	if err != nil {
		return nil, err
	}

	return &cloudserver.SubnetListResult{Count: res.Count, Details: res.Details}, nil
}

// BatchDeleteSubnet batch delete subnet.
func (svc *subnetSvc) BatchDeleteSubnet(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteSubnet(cts, handler.ResOperateAuth)
}

// BatchDeleteBizSubnet batch delete biz subnet.
func (svc *subnetSvc) BatchDeleteBizSubnet(cts *rest.Contexts) (interface{}, error) {
	return svc.batchDeleteSubnet(cts, handler.BizOperateAuth)
}

func (svc *subnetSvc) batchDeleteSubnet(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{},
	error) {

	req := new(core.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.SubnetCloudResType,
		IDs:          req.IDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Subnet,
		Action: meta.Delete, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	// create delete audit.
	if err := svc.audit.ResDeleteAudit(cts.Kit, enumor.SubnetAuditResType, req.IDs); err != nil {
		logs.Errorf("create delete audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// delete subnets
	tasks := make([]ts.CustomFlowTask, 0, len(req.IDs))
	for _, id := range req.IDs {
		basicInfo, exists := basicInfoMap[id]
		if !exists {
			return nil, errf.New(errf.InvalidParameter, fmt.Sprintf("id %s has no corresponding vendor", id))
		}

		tasks = append(tasks, ts.CustomFlowTask{
			ActionID:   action.ActIDType(uuid.UUID()),
			ActionName: enumor.ActionDeleteSubnet,
			Params: &actionsubnet.DeleteSubnetOption{
				Vendor: basicInfo.Vendor,
				ID:     id,
			},
		})
	}

	addReq := &ts.AddCustomFlowReq{
		Name:  enumor.FlowDeleteSubnet,
		Tasks: tasks,
	}
	result, err := svc.client.TaskServer().CreateCustomFlow(cts.Kit, addReq)
	if err != nil {
		logs.Errorf("call taskserver to create custom flow failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, async.WaitTaskToEnd(cts.Kit, svc.client.TaskServer(), result.ID)
}

// AssignSubnetToBiz assign subnets to biz.
func (svc *subnetSvc) AssignSubnetToBiz(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudserver.AssignSubnetToBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := common.ValidateTargetBizID(cts.Kit, svc.client.DataService(), enumor.SubnetCloudResType, req.SubnetIDs,
		req.BkBizID)
	if err != nil {
		return nil, err
	}

	// authorize
	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.SubnetCloudResType,
		IDs:          req.SubnetIDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		return nil, err
	}

	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Subnet, Action: meta.Assign,
			ResourceID: info.AccountID}, BizID: req.BkBizID})
	}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
	if err != nil {
		return nil, err
	}

	// check if all subnets are not assigned, right now assigning resource twice is not allowed
	subnetFilter := &filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.SubnetIDs}
	err = CheckSubnetsInBiz(cts.Kit, svc.client, subnetFilter, constant.UnassignedBiz)
	if err != nil {
		return nil, err
	}

	// create assign audit.
	err = svc.audit.ResBizAssignAudit(cts.Kit, enumor.SubnetAuditResType, req.SubnetIDs, req.BkBizID)
	if err != nil {
		logs.Errorf("create assign audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// update subnet biz relations
	updateReq := &cloud.SubnetBaseInfoBatchUpdateReq{
		Subnets: []cloud.SubnetBaseInfoUpdateReq{{
			IDs: req.SubnetIDs,
			Data: &cloud.SubnetUpdateBaseInfo{
				BkBizID: req.BkBizID,
			},
		}},
	}

	err = svc.client.DataService().Global.Subnet.BatchUpdateBaseInfo(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// CheckSubnetsInBiz check if subnets are in the specified biz.
func CheckSubnetsInBiz(kt *kit.Kit, client *client.ClientSet, rule filter.RuleFactory, bizID int64) error {
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
	result, err := client.DataService().Global.Subnet.List(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("count subnets that are not in biz failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return err
	}

	if result.Count != 0 {
		return fmt.Errorf("%d subnets are already assigned", result.Count)
	}

	return nil
}

// checkAzureSubnetParams check azure subnet params
func (svc *subnetSvc) checkAzureSubnetParams(req *cloudserver.AzureSubnetCreateReq) error {
	if !assert.IsSameCaseString(req.Name) {
		return errf.New(errf.InvalidParameter, "name can only be lowercase")
	}

	if !assert.IsSameCaseString(req.CloudVpcID) {
		return errf.New(errf.InvalidParameter, "cloud_vpc_id can only be lowercase")
	}

	if !assert.IsSameCaseString(req.ResourceGroup) {
		return errf.New(errf.InvalidParameter, "resource_group can only be lowercase")
	}

	if !assert.IsSameCaseString(req.CloudRouteTableID) {
		return errf.New(errf.InvalidParameter, "cloud_route_table_id can only be lowercase")
	}

	if !assert.IsSameCaseString(req.NetworkSecurityGroup) {
		return errf.New(errf.InvalidParameter, "network_security_group can only be lowercase")
	}

	return nil
}

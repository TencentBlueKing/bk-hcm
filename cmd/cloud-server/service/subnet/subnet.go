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

package subnet

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
)

// InitSubnetService initialize the subnet service.
func InitSubnetService(c *capability.Capability) {
	svc := &subnetSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	h.Add("GetSubnet", "GET", "/subnets/{id}", svc.GetSubnet)
	h.Add("ListSubnet", "POST", "/subnets/list", svc.ListSubnet)
	h.Add("UpdateSubnet", "PATCH", "/subnets/{id}", svc.UpdateSubnet)
	h.Add("BatchDeleteSubnet", "DELETE", "/subnets/batch", svc.BatchDeleteSubnet)
	h.Add("AssignSubnetToBiz", "POST", "/subnets/assign/bizs", svc.AssignSubnetToBiz)

	h.Load(c.WebService)
}

type subnetSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

// UpdateSubnet update subnet.
func (svc *subnetSvc) UpdateSubnet(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudserver.SubnetUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		enumor.SubnetCloudResType, id)
	if err != nil {
		return nil, err
	}

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Subnet, Action: meta.Update,
		ResourceID: basicInfo.AccountID}}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	// check if all subnets are not assigned, cannot operate biz resource in account api
	subnetFilter := &filter.AtomRule{Field: "id", Op: filter.Equal.Factory(), Value: id}
	err = svc.checkSubnetsInBiz(cts.Kit, subnetFilter, constant.UnassignedBiz)
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
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		enumor.SubnetCloudResType, id)
	if err != nil {
		return nil, err
	}

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Subnet, Action: meta.Find,
		ResourceID: basicInfo.AccountID}}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	// check if all subnets are not assigned, cannot operate biz resource in account api
	subnetFilter := &filter.AtomRule{Field: "id", Op: filter.Equal.Factory(), Value: id}
	err = svc.checkSubnetsInBiz(cts.Kit, subnetFilter, constant.UnassignedBiz)
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

// ListSubnet list subnets.
func (svc *subnetSvc) ListSubnet(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// check if all subnets are not assigned, cannot operate biz resource in account api
	err := svc.checkSubnetsInBiz(cts.Kit, req.Filter, constant.UnassignedBiz)
	if err != nil {
		return nil, err
	}

	// list authorized instances
	authOpt := &meta.ListAuthResInput{Type: meta.Subnet, Action: meta.Find}
	expr, noPermFlag, err := svc.authorizer.ListAuthInstWithFilter(cts.Kit, authOpt, req.Filter, "account_id")
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

// BatchDeleteSubnet batch delete subnets.
func (svc *subnetSvc) BatchDeleteSubnet(cts *rest.Contexts) (interface{}, error) {
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
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	// authorize
	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Subnet, Action: meta.Delete,
			ResourceID: info.AccountID}})
	}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
	if err != nil {
		return nil, err
	}

	// check if all subnets are not assigned, cannot operate biz resource in account api
	subnetFilter := &filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.IDs}
	err = svc.checkSubnetsInBiz(cts.Kit, subnetFilter, constant.UnassignedBiz)
	if err != nil {
		return nil, err
	}

	// create delete audit.
	if err := svc.audit.ResDeleteAudit(cts.Kit, enumor.SubnetAuditResType, req.IDs); err != nil {
		logs.Errorf("create delete audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// delete subnets
	succeeded := make([]string, 0)
	for _, id := range req.IDs {
		basicInfo, exists := basicInfoMap[id]
		if !exists {
			return nil, errf.New(errf.InvalidParameter, fmt.Sprintf("id %s has no corresponding vendor", id))
		}

		switch basicInfo.Vendor {
		case enumor.TCloud:
			err = svc.client.HCService().TCloud.Subnet.Delete(cts.Kit.Ctx, cts.Kit.Header(), id)
		case enumor.Aws:
			err = svc.client.HCService().Aws.Subnet.Delete(cts.Kit.Ctx, cts.Kit.Header(), id)
		case enumor.Gcp:
			err = svc.client.HCService().Gcp.Subnet.Delete(cts.Kit.Ctx, cts.Kit.Header(), id)
		case enumor.Azure:
			err = svc.client.HCService().Azure.Subnet.Delete(cts.Kit.Ctx, cts.Kit.Header(), id)
		case enumor.HuaWei:
			err = svc.client.HCService().HuaWei.Subnet.Delete(cts.Kit.Ctx, cts.Kit.Header(), id)
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

// AssignSubnetToBiz assign subnets to biz.
func (svc *subnetSvc) AssignSubnetToBiz(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudserver.AssignSubnetToBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// authorize
	basicInfoReq := cloud.ListResourceBasicInfoReq{
		ResourceType: enumor.SubnetCloudResType,
		IDs:          req.SubnetIDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		basicInfoReq)
	if err != nil {
		return nil, err
	}

	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Subnet, Action: meta.Assign,
			ResourceID: info.AccountID}})
	}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
	if err != nil {
		return nil, err
	}

	// check if all subnets are not assigned, right now assigning resource twice is not allowed
	subnetFilter := &filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: req.SubnetIDs}
	err = svc.checkSubnetsInBiz(cts.Kit, subnetFilter, constant.UnassignedBiz)
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

// checkSubnetsInBiz check if subnets are in the specified biz.
func (svc *subnetSvc) checkSubnetsInBiz(kt *kit.Kit, rule filter.RuleFactory, bizID int64) error {
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
	result, err := svc.client.DataService().Global.Subnet.List(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("count subnets that are not in biz failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return err
	}

	if result.Count != 0 {
		return fmt.Errorf("%d subnets are already assigned", result.Count)
	}

	return nil
}

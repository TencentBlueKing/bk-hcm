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
	"strings"

	cloudserver "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/classifier"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// ListResourceIdBySecurityGroup list resource id by security group
func (svc *securityGroupSvc) ListResourceIdBySecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.listResourceIdBySecurityGroup(cts, handler.ResOperateAuth)
}

// ListBizResourceIDBySecurityGroup list biz resource id by security group
func (svc *securityGroupSvc) ListBizResourceIDBySecurityGroup(cts *rest.Contexts) (interface{}, error) {
	return svc.listResourceIdBySecurityGroup(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) listResourceIdBySecurityGroup(cts *rest.Contexts,
	validHandler handler.ValidWithAuthHandler) (interface{}, error) {

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	baseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, id)
	if err != nil {
		logs.Errorf("get security group basic info failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Find, BasicInfo: baseInfo})
	if err != nil {
		return nil, err
	}

	return svc.client.DataService().Global.SGCommonRel.ListSgCommonRels(cts.Kit, req)
}

// QueryBizRelatedResourceCount query biz related resource count
func (svc *securityGroupSvc) QueryBizRelatedResourceCount(cts *rest.Contexts) (interface{}, error) {
	return svc.queryRelatedResourceCount(cts, handler.ListBizAuthRes)
}

// QueryRelatedResourceCount query related resource count
func (svc *securityGroupSvc) QueryRelatedResourceCount(cts *rest.Contexts) (interface{}, error) {
	return svc.queryRelatedResourceCount(cts, handler.ListResourceAuthRes)
}

func (svc *securityGroupSvc) queryRelatedResourceCount(cts *rest.Contexts, validHandler handler.ListAuthResHandler) (
	interface{}, error) {

	req := new(cloudserver.SecurityGroupQueryRelatedResourceCountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authFilter, noPerm, err := validHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.SecurityGroup, Action: meta.Find})
	if err != nil {
		return nil, err
	}
	if noPerm {
		return nil, errf.New(errf.PermissionDenied, "permission denied for query related resource count")
	}

	securityGroups, err := svc.listSecurityGroupByIDsAndFilter(cts.Kit, req.IDs, authFilter)
	if err != nil {
		logs.Errorf("list security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return svc.queryRelatedResourceCountFromCloud(cts.Kit, securityGroups)

}

func (svc *securityGroupSvc) listSecurityGroupByIDsAndFilter(kt *kit.Kit, ids []string,
	authFilter *filter.Expression) ([]cloud.BaseSecurityGroup, error) {

	resultMap := make(map[string]cloud.BaseSecurityGroup, len(ids))
	for _, sgIDs := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		var finalFilter *filter.Expression
		var err error
		if authFilter != nil {
			finalFilter, err = tools.And(authFilter, tools.ContainersExpression("id", sgIDs))
			if err != nil {
				logs.Errorf("build filter failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		} else {
			finalFilter = tools.ContainersExpression("id", sgIDs)
		}

		listReq := &dataproto.SecurityGroupListReq{
			Filter: finalFilter,
			Page:   core.NewDefaultBasePage(),
		}
		resp, err := svc.client.DataService().Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("ListSecurityGroup failed, err: %v, ids: %v, rid: %s", err, sgIDs, kt.Rid)
			return nil, err
		}
		for _, detail := range resp.Details {
			resultMap[detail.ID] = detail
		}
	}
	result := make([]cloud.BaseSecurityGroup, 0, len(ids))
	for _, id := range ids {
		item, ok := resultMap[id]
		if !ok {
			logs.Errorf("security group %s not found, rid: %s", id, kt.Rid)
			return nil, fmt.Errorf("security group %s not found", id)
		}
		result = append(result, item)
	}
	return result, nil
}

func (svc *securityGroupSvc) queryRelatedResourceCountFromCloud(kt *kit.Kit,
	securityGroups []cloud.BaseSecurityGroup) (*cloudserver.ListSecurityGroupStatisticResp, error) {

	sgByVendor := classifier.ClassifySlice(securityGroups, func(sg cloud.BaseSecurityGroup) string {
		// key: vendor+accountID+region
		return fmt.Sprintf("%s,%s,%s", sg.Vendor, sg.AccountID, sg.Region)
	})
	resultMap := make(map[string]*cloudserver.SecurityGroupStatisticItem)
	for key, groups := range sgByVendor {
		arr := strings.Split(key, ",")
		vendor := enumor.Vendor(arr[0])

		accountID, region := arr[1], arr[2]
		listFunc, err := svc.chooseListSGStatisticFunc(vendor)
		if err != nil {
			logs.Errorf("choose list security group statistic func failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		ids := make([]string, 0, len(groups))
		for _, info := range groups {
			ids = append(ids, info.ID)
		}
		req := &hcservice.ListSecurityGroupStatisticReq{
			SecurityGroupIDs: ids,
			Region:           region,
			AccountID:        accountID,
		}
		resp, err := listFunc(kt, req)
		if err != nil {
			logs.Errorf("list security group statistic failed, err: %v, req: %v, rid: %s",
				err, req, kt.Rid)
			for _, id := range ids {
				resultMap[id] = &cloudserver.SecurityGroupStatisticItem{
					ID:    id,
					Error: cvt.ValToPtr(err.Error()),
				}
			}
			continue
		}
		for _, detail := range resp.Details {
			resultMap[detail.ID] = &cloudserver.SecurityGroupStatisticItem{
				ID:        detail.ID,
				Resources: detail.Resources,
			}
		}
	}
	result := &cloudserver.ListSecurityGroupStatisticResp{}
	for _, group := range securityGroups {
		statistic, ok := resultMap[group.ID]
		if !ok {
			logs.Errorf("security group %s statistic not found, rid: %s", group.ID, kt.Rid)
			return nil, fmt.Errorf("security group %s statistic not found", group.ID)
		}
		result.Details = append(result.Details, statistic)
	}

	return result, nil
}

type callListSecurityGroupStatisticFunc func(kt *kit.Kit, req *hcservice.ListSecurityGroupStatisticReq) (
	*hcservice.ListSecurityGroupStatisticResp, error)

func (svc *securityGroupSvc) chooseListSGStatisticFunc(vendor enumor.Vendor) (
	callListSecurityGroupStatisticFunc, error) {

	switch vendor {
	case enumor.TCloud:
		return svc.client.HCService().TCloud.SecurityGroup.ListSecurityGroupStatistic, nil
	case enumor.Aws:
		return svc.client.HCService().Aws.SecurityGroup.ListSecurityGroupStatistic, nil
	case enumor.HuaWei:
		return svc.client.HCService().HuaWei.SecurityGroup.ListSecurityGroupStatistic, nil
	case enumor.Azure:
		return svc.client.HCService().Azure.SecurityGroup.ListSecurityGroupStatistic, nil
	default:
		return nil, fmt.Errorf("vendor: %s not support for ListSecurityGroupStatistic", vendor)
	}
}

// ListSecurityGroupRelBusiness list security group rel business
func (svc *securityGroupSvc) ListSecurityGroupRelBusiness(cts *rest.Contexts) (interface{}, error) {
	return svc.listSecurityGroupRelBusiness(cts, constant.UnassignedBiz, handler.ResOperateAuth)
}

// ListBizSecurityGroupRelBusiness list biz security group rel business
func (svc *securityGroupSvc) ListBizSecurityGroupRelBusiness(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.New(errf.InvalidParameter, "bk_biz_id must be int64")
	}

	return svc.listSecurityGroupRelBusiness(cts, bizID, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) listSecurityGroupRelBusiness(cts *rest.Contexts, bizID int64,
	validHandler handler.ValidWithAuthHandler) (interface{}, error) {

	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security_group_id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, sgID)
	if err != nil {
		logs.Errorf("get security group basic info failed, id: %s, err: %v, rid: %s", sgID, err, cts.Kit.Rid)
		return nil, err
	}

	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	return svc.sgLogic.ListSGRelBusiness(cts.Kit, bizID, sgID)
}

// ListSGRelCVMByBizID list security group rel cvm by biz id
func (svc *securityGroupSvc) ListSGRelCVMByBizID(cts *rest.Contexts) (interface{}, error) {
	return svc.listSGRelCVMByBizID(cts, handler.ResOperateAuth)
}

// ListBizSGRelCVMByBizID list biz security group rel cvm by biz id
func (svc *securityGroupSvc) ListBizSGRelCVMByBizID(cts *rest.Contexts) (interface{}, error) {
	return svc.listSGRelCVMByBizID(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) listSGRelCVMByBizID(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	sgID := cts.PathParameter("sg_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "sg_id is required")
	}

	resBizID, err := cts.PathParameter("res_biz_id").Int64()
	if err != nil {
		return nil, errf.New(errf.InvalidParameter, "res_biz_id need be int")
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, sgID)
	if err != nil {
		logs.Errorf("get security group basic info failed, id: %s, err: %v, rid: %s", sgID, err, cts.Kit.Rid)
		return nil, err
	}

	// 非管理业务，不允许查看其他业务的绑定资源详情
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err == nil {
		if basicInfo.BkBizID != bizID && resBizID != bizID {
			return nil, errf.New(errf.InvalidParameter,
				"non-management business can only list its own resources")
		}
	}

	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	return svc.sgLogic.ListSGRelCVM(cts.Kit, sgID, resBizID, req)
}

// ListSGRelLBByBizID list security group rel load balancer by biz id
func (svc *securityGroupSvc) ListSGRelLBByBizID(cts *rest.Contexts) (interface{}, error) {
	return svc.listSGRelLBByBizID(cts, handler.ResOperateAuth)
}

// ListBizSGRelLBByBizID list biz security group rel load balancer by biz id
func (svc *securityGroupSvc) ListBizSGRelLBByBizID(cts *rest.Contexts) (interface{}, error) {
	return svc.listSGRelLBByBizID(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) listSGRelLBByBizID(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	sgID := cts.PathParameter("sg_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "sg_id is required")
	}

	resBizID, err := cts.PathParameter("res_biz_id").Int64()
	if err != nil {
		return nil, errf.New(errf.InvalidParameter, "res_biz_id need be int")
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, sgID)
	if err != nil {
		logs.Errorf("get security group basic info failed, id: %s, err: %v, rid: %s", sgID, err, cts.Kit.Rid)
		return nil, err
	}

	// 非管理业务，不允许查看其他业务的绑定资源详情
	bizIDStr := cts.PathParameter("bk_biz_id")
	if bizIDStr != "" {
		bizID, err := bizIDStr.Int64()
		if err != nil {
			return nil, errf.New(errf.InvalidParameter, "bk_biz_id need be int")
		}

		if basicInfo.BkBizID != bizID && resBizID != bizID {
			return nil, errf.New(errf.InvalidParameter,
				"non-management business can only list its own resources")
		}
	}

	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	return svc.sgLogic.ListSGRelLoadBalancer(cts.Kit, sgID, resBizID, req)
}

// ListSGRelCVM ...
func (svc *securityGroupSvc) ListSGRelCVM(cts *rest.Contexts) (interface{}, error) {
	return svc.listSGRelCVM(cts, handler.ResOperateAuth)
}

func (svc *securityGroupSvc) listSGRelCVM(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	sgID := cts.PathParameter("sg_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "sg_id is required")
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, sgID)
	if err != nil {
		logs.Errorf("get security group basic info failed, id: %s, err: %v, rid: %s", sgID, err, cts.Kit.Rid)
		return nil, err
	}

	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	listReq := &dataproto.SGCommonRelListReq{
		SGIDs:   []string{sgID},
		ListReq: *req,
	}
	return svc.client.DataService().Global.SGCommonRel.ListWithCVMSummary(cts.Kit, listReq)
}

// ListSGRelLB ...
func (svc *securityGroupSvc) ListSGRelLB(cts *rest.Contexts) (interface{}, error) {
	return svc.listSGRelLB(cts, handler.ResOperateAuth)
}

func (svc *securityGroupSvc) listSGRelLB(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	sgID := cts.PathParameter("sg_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "sg_id is required")
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, sgID)
	if err != nil {
		logs.Errorf("get security group basic info failed, id: %s, err: %v, rid: %s", sgID, err, cts.Kit.Rid)
		return nil, err
	}

	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	listReq := &dataproto.SGCommonRelListReq{
		SGIDs:   []string{sgID},
		ListReq: *req,
	}
	return svc.client.DataService().Global.SGCommonRel.ListWithLBSummary(cts.Kit, listReq)
}

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

	cloudserver "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// BatchListResSecurityGroups ...
func (svc *securityGroupSvc) BatchListResSecurityGroups(cts *rest.Contexts) (interface{}, error) {
	return svc.listResSecurityGroups(cts, handler.ResOperateAuth)
}

// BizBatchListResSecurityGroups ...
func (svc *securityGroupSvc) BizBatchListResSecurityGroups(cts *rest.Contexts) (interface{}, error) {
	return svc.listResSecurityGroups(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) listResSecurityGroups(cts *rest.Contexts, authHandler handler.ValidWithAuthHandler) (
	[]cloudserver.ResSGRel, error) {

	resType := enumor.CloudResourceType(cts.PathParameter("res_type").String())
	if len(resType) == 0 {
		return nil, errf.New(errf.InvalidParameter, "res_type is required")
	} else if resType != enumor.CvmCloudResType && resType != enumor.LoadBalancerCloudResType {
		return nil, errf.New(errf.InvalidParameter, fmt.Sprintf("invalid res_type %s", resType))
	}

	req := new(cloudserver.BatchGetResRelatedSecurityGroupsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: resType,
		IDs:          req.ResIDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		return nil, err
	}
	err = authHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer,
		ResType: meta.ResourceType(resType), Action: meta.Find, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	rels, err := svc.listSGCommonRels(cts.Kit, resType, req.ResIDs)
	if err != nil {
		logs.Errorf("list cvm security group rels failed, err: %v, cvm_ids: %v, rid: %s",
			err, req.ResIDs, cts.Kit.Rid)
		return nil, err
	}

	sgIDsMap := make(map[string]struct{})
	for _, rel := range rels {
		sgIDsMap[rel.SecurityGroupID] = struct{}{}
	}
	securityGroupsMap, err := svc.getSecurityGroupsMap(cts.Kit, converter.MapKeyToSlice(sgIDsMap))
	if err != nil {
		return nil, err
	}

	itemMap := convSGInfo(securityGroupsMap)
	cvmToSgMap := make(map[string][]cloudserver.SGInfo, len(req.ResIDs))
	for _, rel := range rels {
		cvmToSgMap[rel.ResID] = append(cvmToSgMap[rel.ResID], itemMap[rel.SecurityGroupID])
	}

	return buildBatchListResSecurityGroupsResp(req.ResIDs, cvmToSgMap), nil
}

func buildBatchListResSecurityGroupsResp(resIDs []string,
	cvmToSgMap map[string][]cloudserver.SGInfo) []cloudserver.ResSGRel {

	result := make([]cloudserver.ResSGRel, len(resIDs))
	for i, resID := range resIDs {
		sgList, ok := cvmToSgMap[resID]
		if !ok {
			sgList = make([]cloudserver.SGInfo, 0)
		}
		result[i] = cloudserver.ResSGRel{
			ResID:          resID,
			SecurityGroups: sgList,
		}
	}
	return result
}

func convSGInfo(
	m map[string]cloud.BaseSecurityGroup) map[string]cloudserver.SGInfo {

	result := make(map[string]cloudserver.SGInfo, len(m))
	for id, sg := range m {
		result[id] = cloudserver.SGInfo{
			ID:      sg.ID,
			Name:    sg.Name,
			CloudId: sg.CloudID,
		}
	}
	return result
}

func (svc *securityGroupSvc) getSecurityGroupsMap(kt *kit.Kit, sgIDs []string) (
	map[string]cloud.BaseSecurityGroup, error) {

	result := make(map[string]cloud.BaseSecurityGroup, len(sgIDs))
	for _, ids := range slice.Split(sgIDs, int(core.DefaultMaxPageLimit)) {
		req := &dataproto.SecurityGroupListReq{
			Filter: tools.ContainersExpression("id", ids),
			Page:   core.NewDefaultBasePage(),
		}
		securityGroups, err := svc.client.DataService().Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("list security groups failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
			return nil, err
		}

		for _, detail := range securityGroups.Details {
			result[detail.ID] = detail
		}
	}
	if len(result) != len(sgIDs) {
		for _, sgID := range sgIDs {
			if _, ok := result[sgID]; !ok {
				logs.Errorf("security group not found, id: %s, rid: %s", sgID, kt.Rid)
				return nil, fmt.Errorf("security group not found, id: %s", sgID)
			}
		}
	}
	return result, nil
}

func (svc *securityGroupSvc) listSGCommonRels(kt *kit.Kit, resType enumor.CloudResourceType, resIDs []string) (
	[]cloud.SecurityGroupCommonRel, error) {

	result := make([]cloud.SecurityGroupCommonRel, 0)

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleIn("res_id", resIDs),
			tools.RuleEqual("res_type", resType),
		),
		Page: core.NewDefaultBasePage(),
	}

	for {
		sgCvmRels, err := svc.client.DataService().Global.SGCommonRel.ListSgCommonRels(kt, req)
		if err != nil {
			return nil, err
		}
		if len(sgCvmRels.Details) == 0 {
			break
		}
		result = append(result, sgCvmRels.Details...)
		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return result, nil
}

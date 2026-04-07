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

package permissionpolicylibrary

import (
	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

// ListPermissionPolicyLibrary list permission policy library.
func (svc *svc) ListPermissionPolicyLibrary(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(proto.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("decode list permission policy library request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.PermissionPolicyLibrary, Action: meta.Find}}
	_, authorized, err := svc.authorizer.Authorize(cts.Kit, authRes)
	if err != nil {
		logs.Errorf("list permission policy library authorize failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "no permission to list permission policy library")
	}

	listFilter := req.Filter
	listFilter, err = tools.And(listFilter, tools.EqualExpression("vendor", vendor))
	if err != nil {
		logs.Errorf("merge vendor filter failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	dsReq := &protocloud.PermissionPolicyLibraryListReq{
		Filter: listFilter,
		Page:   req.Page,
	}
	dsResult, err := svc.client.DataService().Global.PermissionPolicyLibrary.ListPermissionPolicyLibrary(cts.Kit, dsReq)
	if err != nil {
		logs.Errorf("list permission policy library from data service failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &proto.PermissionPolicyLibraryListResult{Count: dsResult.Count}, nil
	}

	libraryIDs := make([]string, 0, len(dsResult.Details))
	for _, item := range dsResult.Details {
		libraryIDs = append(libraryIDs, item.ID)
	}

	countMap, err := svc.buildLibraryAccountCountMap(cts.Kit, libraryIDs)
	if err != nil {
		logs.Errorf("build library account count map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	details := make([]proto.PermissionPolicyLibraryResult, 0, len(dsResult.Details))
	for _, item := range dsResult.Details {
		details = append(details, proto.PermissionPolicyLibraryResult{
			BasePermissionPolicyLibrary: item,
			AssociatedAccountCount:      countMap[item.ID],
		})
	}

	return &proto.PermissionPolicyLibraryListResult{Count: 0, Details: details}, nil
}

// buildLibraryAccountCountMap queries permission_template by policy_library_id and returns
// a map of libraryID to unique associated account count.
func (svc *svc) buildLibraryAccountCountMap(kt *kit.Kit, libraryIDs []string) (map[string]int, error) {
	countMap := make(map[string]int, len(libraryIDs))
	if len(libraryIDs) == 0 {
		return countMap, nil
	}

	accountSets := make(map[string]map[string]struct{}, len(libraryIDs))
	req := &protocloud.PermissionTemplateListReq{
		Filter: tools.ContainersExpression("policy_library_id", libraryIDs),
		Page:   core.NewDefaultBasePage(),
	}
	for {
		result, err := svc.client.DataService().Global.PermissionTemplate.ListPermissionTemplate(kt, req)
		if err != nil {
			logs.Errorf("list permission template failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, tmpl := range result.Details {
			if tmpl.PolicyLibraryID == nil {
				continue
			}
			libID := cvt.PtrToVal(tmpl.PolicyLibraryID)
			if _, ok := accountSets[libID]; !ok {
				accountSets[libID] = make(map[string]struct{})
			}
			accountSets[libID][tmpl.AccountID] = struct{}{}
		}

		if uint(len(result.Details)) < core.DefaultMaxPageLimit {
			break
		}
		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	for libID, accountSet := range accountSets {
		countMap[libID] = len(accountSet)
	}
	return countMap, nil
}

// ListPermissionPolicyLibraryUnappliedAccountIDs returns account IDs that have not applied the given policy library.
func (svc *svc) ListPermissionPolicyLibraryUnappliedAccountIDs(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:       meta.PermissionPolicyLibrary,
			Action:     meta.Find,
			ResourceID: id,
		},
	}
	_, authorized, err := svc.authorizer.Authorize(cts.Kit, authRes)
	if err != nil {
		logs.Errorf("list unapplied account ids auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "no permission to query unapplied account ids")
	}

	applier := NewPolicyLibraryApplier(svc.client, svc.audit)
	accountIDs, err := applier.ListUnappliedAccountIDs(cts.Kit, vendor, id)
	if err != nil {
		return nil, err
	}

	if accountIDs == nil {
		accountIDs = make([]string, 0)
	}
	return &proto.PermissionPolicyLibraryAccountIDsResult{AccountIDs: accountIDs}, nil
}

// ListPermissionPolicyLibraryPermissionTemplates returns all permission templates applied from the given policy library
// that are still within the library's current biz scope.
func (svc *svc) ListPermissionPolicyLibraryPermissionTemplates(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:       meta.PermissionPolicyLibrary,
			Action:     meta.Find,
			ResourceID: id,
		},
	}
	_, authorized, err := svc.authorizer.Authorize(cts.Kit, authRes)
	if err != nil {
		logs.Errorf("list permission templates auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "no permission to list permission templates")
	}

	applier := NewPolicyLibraryApplier(svc.client, svc.audit)
	details, err := applier.ListTemplatesInScope(cts.Kit, vendor, id)
	if err != nil {
		logs.Errorf("list permission templates in scope failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	return &proto.PermissionPolicyLibraryPermTmplResult{Details: details}, nil
}

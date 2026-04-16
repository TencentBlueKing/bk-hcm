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
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
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

	libAccountIDsMap, err := svc.buildLibraryAccountIDsMap(cts.Kit, libraryIDs)
	if err != nil {
		logs.Errorf("build library account ids map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	details := make([]proto.PermissionPolicyLibraryResult, 0, len(dsResult.Details))
	for _, item := range dsResult.Details {
		details = append(details, proto.PermissionPolicyLibraryResult{
			BasePermissionPolicyLibrary: item,
			AssociatedAccountCount:      len(libAccountIDsMap[item.ID]),
		})
	}

	return &proto.PermissionPolicyLibraryListResult{Count: 0, Details: details}, nil
}

// buildLibraryAccountIDsMap queries permission_template by policy_library_id and returns
// a map of libraryID to unique associated account ids.
func (svc *svc) buildLibraryAccountIDsMap(kt *kit.Kit, libraryIDs []string) (map[string][]string, error) {
	libAccountIDsMap := make(map[string][]string, len(libraryIDs))
	if len(libraryIDs) == 0 {
		return libAccountIDsMap, nil
	}

	for _, batch := range slice.Split(libraryIDs, int(core.DefaultMaxPageLimit)) {
		accountSets := make(map[string]map[string]struct{}, len(batch))
		req := &protocloud.PermissionTemplateListReq{
			Filter: tools.ContainersExpression("policy_library_id", batch),
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
			libAccountIDsMap[libID] = maps.Keys(accountSet)
		}
	}

	return libAccountIDsMap, nil
}

// ListPermissionPolicyLibraryUnAppliedAccountIDs returns account IDs that have not applied the given policy library.
func (svc *svc) ListPermissionPolicyLibraryUnAppliedAccountIDs(cts *rest.Contexts) (interface{}, error) {
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
	accountIDs, err := applier.ListUnAppliedAccountIDs(cts.Kit, vendor, id)
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

// ListPermissionPolicyLibraryAccountIDs returns all account IDs associated with the given policy library.
func (svc *svc) ListPermissionPolicyLibraryAccountIDs(cts *rest.Contexts) (interface{}, error) {
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
		logs.Errorf("list permission policy library account ids auth failed, err: %v, id: %s, vendor: %s, rid: %s", err,
			id, vendor, cts.Kit.Rid)
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "no permission to list permission policy library apply account ids")
	}

	applier := NewPolicyLibraryApplier(svc.client, svc.audit)
	accountIDs, err := applier.ListAllAppliedAccountIDs(cts.Kit, id)
	if err != nil {
		logs.Errorf("list all applied account ids failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	if accountIDs == nil {
		accountIDs = make([]string, 0)
	}
	return &proto.PermissionPolicyLibraryAccountIDsResult{AccountIDs: accountIDs}, nil
}

// ListBizPermissionPolicyLibraryAccountIDs returns account IDs associated with the given policy library,
// filtered to only those accounts whose management biz equals bk_biz_id.
func (svc *svc) ListBizPermissionPolicyLibraryAccountIDs(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err = vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bizID,
	}
	_, authorized, err := svc.authorizer.Authorize(cts.Kit, authRes)
	if err != nil {
		logs.Errorf("list biz permission policy library account ids auth failed, err: %v, id: %s, vendor: %s, rid: %s",
			err, id, vendor, cts.Kit.Rid)
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "no permission to query account ids under this biz")
	}

	applier := NewPolicyLibraryApplier(svc.client, svc.audit)
	library, err := applier.GetPolicyLibraryDetail(cts.Kit, id)
	if err != nil {
		logs.Errorf("get policy library detail failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	inScope := false
	for _, biz := range library.BkBizIDs {
		if biz == bizID {
			inScope = true
			break
		}
	}
	if !inScope {
		return nil, errf.Newf(errf.InvalidParameter, "bk_biz_id %d is not in policy library scope", bizID)
	}

	allAccountIDs, err := applier.ListAllAppliedAccountIDs(cts.Kit, id)
	if err != nil {
		logs.Errorf("list all applied account ids failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	accountIDs, err := svc.filterAccountIDsByBizID(cts.Kit, allAccountIDs, bizID)
	if err != nil {
		logs.Errorf("filter account ids by biz failed, bizID: %d, err: %v, rid: %s", bizID, err, cts.Kit.Rid)
		return nil, err
	}

	return &proto.PermissionPolicyLibraryAccountIDsResult{AccountIDs: accountIDs}, nil
}

// ListBizPermissionPolicyLibraryUnAppliedAccountIDs returns account IDs under the given biz
// that have not applied the given policy library.
func (svc *svc) ListBizPermissionPolicyLibraryUnAppliedAccountIDs(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err = vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bizID,
	}
	_, authorized, err := svc.authorizer.Authorize(cts.Kit, authRes)
	if err != nil {
		logs.Errorf("list biz unapplied account ids auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "no permission to query unapplied account ids under this biz")
	}

	applier := NewPolicyLibraryApplier(svc.client, svc.audit)
	accountIDs, err := applier.ListBizUnAppliedAccountIDs(cts.Kit, vendor, id, bizID)
	if err != nil {
		logs.Errorf("list biz unapplied account ids failed, id: %s, bizID: %d, err: %v, rid: %s",
			id, bizID, err, cts.Kit.Rid)
		return nil, err
	}

	if accountIDs == nil {
		accountIDs = make([]string, 0)
	}
	return &proto.PermissionPolicyLibraryAccountIDsResult{AccountIDs: accountIDs}, nil
}

// ListBizPermissionPolicyLibrary list permission policy library under a biz.
func (svc *svc) ListBizPermissionPolicyLibrary(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err = vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(proto.ListReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("decode list biz permission policy library request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err = req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access}, BizID: bizID}
	_, authorized, err := svc.authorizer.Authorize(cts.Kit, authRes)
	if err != nil {
		logs.Errorf("list biz permission policy library authorize failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "no permission to list permission policy library")
	}

	listFilter := req.Filter
	listFilter, err = tools.And(listFilter, tools.EqualExpression("vendor", vendor),
		tools.RuleJSONContains("bk_biz_ids", bizID))
	if err != nil {
		logs.Errorf("merge vendor filter failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	listReq := &protocloud.PermissionPolicyLibraryListReq{
		Filter: listFilter,
		Page:   req.Page,
	}
	result, err := svc.client.DataService().Global.PermissionPolicyLibrary.ListPermissionPolicyLibrary(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("list biz permission policy library from data service failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &proto.PermissionPolicyLibraryListResult{Count: result.Count}, nil
	}

	libraryIDs := make([]string, 0, len(result.Details))
	for _, item := range result.Details {
		libraryIDs = append(libraryIDs, item.ID)
	}

	countMap, err := svc.buildBizLibraryAccountCountMap(cts.Kit, libraryIDs, bizID)
	if err != nil {
		logs.Errorf("build biz library account count map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	details := make([]proto.PermissionPolicyLibraryResult, 0, len(result.Details))
	for _, item := range result.Details {
		details = append(details, proto.PermissionPolicyLibraryResult{
			BasePermissionPolicyLibrary: item,
			AssociatedAccountCount:      countMap[item.ID],
		})
	}

	return &proto.PermissionPolicyLibraryListResult{Count: 0, Details: details}, nil
}

func (svc *svc) buildBizLibraryAccountCountMap(kt *kit.Kit, libraryIDs []string, bizID int64) (map[string]int, error) {
	libAccountIDsMap, err := svc.buildLibraryAccountIDsMap(kt, libraryIDs)
	if err != nil {
		logs.Errorf("build library account ids map failed, err: %v, libraryIDs: %v, bizID: %d, rid: %s", err,
			libraryIDs, bizID, kt.Rid)
		return nil, err
	}

	countMap := make(map[string]int)
	for libraryID, accountIDs := range libAccountIDsMap {
		accountIDs, err = svc.filterAccountIDsByBizID(kt, accountIDs, bizID)
		if err != nil {
			logs.Errorf("filter account ids by biz id failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		countMap[libraryID] = len(accountIDs)
	}
	return countMap, nil
}

// filterAccountIDsByBizID queries the given account IDs and returns only those whose bk_biz_id equals bizID.
func (svc *svc) filterAccountIDsByBizID(kt *kit.Kit, accountIDs []string, bizID int64) ([]string, error) {
	if len(accountIDs) == 0 {
		return make([]string, 0), nil
	}

	result := make([]string, 0)
	for _, batch := range slice.Split(accountIDs, int(core.DefaultMaxPageLimit)) {
		req := &protocloud.AccountListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", batch), tools.RuleEqual("bk_biz_id", bizID)),
			Fields: []string{"id"},
			Page:   core.NewDefaultBasePage(),
		}
		listResult, err := svc.client.DataService().Global.Account.List(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("list accounts for biz filter failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		for _, account := range listResult.Details {
			result = append(result, account.ID)
		}
	}
	return result, nil
}

// ListBizPermissionPolicyLibraryPermissionTemplates returns all permission templates applied from the given library,
// filtered to those belonging to accounts whose management biz equals bk_biz_id.
func (svc *svc) ListBizPermissionPolicyLibraryPermissionTemplates(cts *rest.Contexts) (interface{}, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err = vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	authRes := meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bizID,
	}
	_, authorized, err := svc.authorizer.Authorize(cts.Kit, authRes)
	if err != nil {
		logs.Errorf("list biz permission templates auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if !authorized {
		return nil, errf.New(errf.PermissionDenied, "no permission to list permission templates under this biz")
	}

	applier := NewPolicyLibraryApplier(svc.client, svc.audit)
	details, err := applier.ListBizTemplatesInScope(cts.Kit, vendor, id, bizID)
	if err != nil {
		logs.Errorf("list biz permission templates failed, id: %s, bizID: %d, err: %v, rid: %s",
			id, bizID, err, cts.Kit.Rid)
		return nil, err
	}

	return &proto.PermissionPolicyLibraryPermTmplResult{Details: details}, nil
}

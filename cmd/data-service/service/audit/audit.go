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

// Package audit ...
package audit

import (
	"fmt"
	"net/http"

	"hcm/cmd/data-service/service/audit/cloud"
	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	coreaudit "hcm/pkg/api/core/audit"
	proto "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitAuditService initial the Audit service
func InitAuditService(cap *capability.Capability) {
	svc := &svc{
		cloudAudit: cloud.NewCloudAudit(cap.Dao),
		dao:        cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("CloudResourceUpdateAudit", http.MethodPost, "/cloud/resources/update_audits/create",
		svc.cloudAudit.CloudResourceUpdateAudit)
	h.Add("CloudResourceDeleteAudit", http.MethodPost, "/cloud/resources/delete_audits/create",
		svc.cloudAudit.CloudResourceDeleteAudit)
	h.Add("CloudResourceAssignAudit", http.MethodPost, "/cloud/resources/assign_audits/create",
		svc.cloudAudit.CloudResourceAssignAudit)
	h.Add("CloudResourceOperationAudit", http.MethodPost, "/cloud/resources/operation_audits/create",
		svc.cloudAudit.CloudResourceOperationAudit)
	h.Add("CloudResourceRecycleAudit", http.MethodPost, "/cloud/resources/recycle_audits/create",
		svc.cloudAudit.CloudResourceRecycleAudit)
	h.Add("ListAudit", http.MethodPost, "/audits/list", svc.ListAudit)
	h.Add("GetAudit", http.MethodGet, "/audits/{id}", svc.GetAudit)

	h.Load(cap.WebService)
}

// Audit define audit service.
type svc struct {
	cloudAudit *cloud.Audit
	dao        dao.Set
}

// ListAudit list audits.
func (svc *svc) ListAudit(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 考虑到性能问题，如果用户不指定查询审计详情，列表返回信息没有审计详情信息
	if req.Fields == nil {
		req.Fields = []string{"id", "res_id", "cloud_res_id", "res_name", "res_type", "action", "bk_biz_id", "vendor",
			"account_id", "operator", "source", "rid", "app_code", "created_at"}
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}
	result, err := svc.dao.Audit().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list audit failed, err: %v", err)
	}
	if req.Page.Count {
		return &proto.ListResult{Count: result.Count}, nil
	}

	details := make([]coreaudit.Audit, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, coreaudit.Audit{
			ID:         one.ID,
			ResID:      one.ResID,
			CloudResID: one.CloudResID,
			ResName:    one.ResName,
			ResType:    one.ResType,
			Action:     one.Action,
			BkBizID:    one.BkBizID,
			Vendor:     one.Vendor,
			AccountID:  one.AccountID,
			Operator:   one.Operator,
			Source:     one.Source,
			Rid:        one.Rid,
			AppCode:    one.AppCode,
			CreatedAt:  one.CreatedAt.String(),
		})
	}

	return &proto.ListResult{Details: details}, nil
}

// GetAudit get audits.
func (svc *svc) GetAudit(cts *rest.Contexts) (interface{}, error) {
	id, err := cts.PathParameter("id").Uint64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.dao.Audit().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list audit failed, err: %v", err)
	}

	if len(result.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "audit: %d not found", id)
	}

	audit := &coreaudit.Audit{
		ID:         result.Details[0].ID,
		ResID:      result.Details[0].ResID,
		CloudResID: result.Details[0].CloudResID,
		ResName:    result.Details[0].ResName,
		ResType:    result.Details[0].ResType,
		Action:     result.Details[0].Action,
		BkBizID:    result.Details[0].BkBizID,
		Vendor:     result.Details[0].Vendor,
		AccountID:  result.Details[0].AccountID,
		Operator:   result.Details[0].Operator,
		Source:     result.Details[0].Source,
		Rid:        result.Details[0].Rid,
		AppCode:    result.Details[0].AppCode,
		Detail:     result.Details[0].Detail,
		CreatedAt:  result.Details[0].CreatedAt.String(),
	}

	return audit, nil
}

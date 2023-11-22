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

package csselection

import (
	"errors"
	"fmt"

	csselection "hcm/pkg/api/cloud-server/cloud-selection"
	"hcm/pkg/api/core"
	dsselection "hcm/pkg/api/data-service/cloud-selection"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// BatchDeleteScheme ..
func (svc *service) BatchDeleteScheme(cts *rest.Contexts) (interface{}, error) {
	req := new(core.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	reses := make([]meta.ResourceAttribute, 0, len(req.IDs))
	for _, one := range req.IDs {
		reses = append(reses, meta.ResourceAttribute{
			Basic: &meta.Basic{
				Type:       meta.CloudSelectionIdc,
				Action:     meta.Delete,
				ResourceID: one,
			},
		})
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, reses...); err != nil {
		logs.Errorf("batch delete scheme auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := svc.client.DataService().Global.CloudSelection.BatchDeleteScheme(cts.Kit, req); err != nil {
		logs.Errorf("call dataservice to batch delete scheme failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// CreateScheme ...
func (svc *service) CreateScheme(cts *rest.Contexts) (interface{}, error) {
	req := new(csselection.SchemeCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	res := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   meta.CloudSelectionIdc,
			Action: meta.Create,
		},
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, res); err != nil {
		logs.Errorf("create scheme auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	vendors, err := svc.getIdcVendorByIDs(cts.Kit, req.ResultIdcIDs)
	if err != nil {
		logs.Errorf("get idc vendor by ids failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	createReq := &dsselection.SchemeCreateReq{
		BkBizID:                req.BkBizID,
		Name:                   req.Name,
		BizType:                req.BizType,
		Vendors:                vendors,
		DeploymentArchitecture: req.DeploymentArchitecture,
		CoverPing:              req.CoverPing,
		CompositeScore:         req.CompositeScore,
		NetScore:               req.NetScore,
		CostScore:              req.CostScore,
		CoverRate:              req.CoverRate,
		UserDistribution:       req.UserDistribution,
		ResultIdcIDs:           req.ResultIdcIDs,
	}
	result, err := svc.client.DataService().Global.CloudSelection.CreateScheme(cts.Kit, createReq)
	if err != nil {
		logs.Errorf("call dataservice to create scheme failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// UpdateScheme ...
func (svc *service) UpdateScheme(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errors.New("id is required")
	}

	req := new(csselection.SchemeUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	res := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:       meta.CloudSelectionIdc,
			Action:     meta.Update,
			ResourceID: id,
		},
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, res); err != nil {
		logs.Errorf("update scheme auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	updateReq := &dsselection.SchemeUpdateReq{
		Name:    req.Name,
		BkBizID: req.BkBizID,
	}
	if err := svc.client.DataService().Global.CloudSelection.UpdateScheme(cts.Kit, id, updateReq); err != nil {
		logs.Errorf("call dataservice to update scheme failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// GetScheme ...
func (svc *service) GetScheme(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errors.New("id is required")
	}

	res := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:       meta.CloudSelectionIdc,
			Action:     meta.Find,
			ResourceID: id,
		},
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, res); err != nil {
		logs.Errorf("get scheme auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	req := &core.ListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.client.DataService().Global.CloudSelection.ListScheme(cts.Kit, req)
	if err != nil {
		logs.Errorf("call dataservice to list scheme failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(result.Details) == 0 {
		return nil, fmt.Errorf("scheme: %s not found", id)
	}

	return result.Details[0], nil
}

// ListScheme ...
func (svc *service) ListScheme(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	res := &meta.ListAuthResInput{
		Type:   meta.CloudSelectionScheme,
		Action: meta.Find,
	}
	expr, noPermFlag, err := svc.authorizer.ListAuthInstWithFilter(cts.Kit, res, req.Filter, "id")
	if err != nil {
		logs.Errorf("list scheme failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}
	req.Filter = expr

	result, err := svc.client.DataService().Global.CloudSelection.ListScheme(cts.Kit, req)
	if err != nil {
		logs.Errorf("call dataservice to list scheme failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

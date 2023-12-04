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

package cloudselection

import (
	"fmt"

	"hcm/pkg/api/core"
	coreselection "hcm/pkg/api/core/cloud-selection"
	dsselection "hcm/pkg/api/data-service/cloud-selection"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableselection "hcm/pkg/dal/table/cloud-selection"
	types2 "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"
)

// ListScheme ...
func (svc *service) ListScheme(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.CloudSelectionScheme().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list cloud selection scheme failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	details := make([]coreselection.Scheme, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, coreselection.Scheme{
			ID:                     one.ID,
			BkBizID:                one.BkBizID,
			Name:                   one.Name,
			BizType:                one.BizType,
			Vendors:                one.Vendors,
			DeploymentArchitecture: one.DeploymentArchitecture,
			CoverPing:              one.CoverPing,
			CompositeScore:         one.CompositeScore,
			NetScore:               one.NetScore,
			CostScore:              one.CostScore,
			CoverRate:              one.CoverRate,
			UserDistribution:       one.UserDistribution,
			ResultIdcIDs:           one.ResultIdcIDs,
			Revision: core.Revision{
				Creator:   one.Creator,
				Reviser:   one.Reviser,
				CreatedAt: one.CreatedAt.String(),
				UpdatedAt: one.UpdatedAt.String(),
			},
		})
	}

	return &core.ListResultT[coreselection.Scheme]{Count: result.Count, Details: details}, nil
}

// BatchDeleteScheme ...
func (svc *service) BatchDeleteScheme(cts *rest.Contexts) (interface{}, error) {
	req := new(core.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	flt := tools.ContainersExpression("id", req.IDs)
	if err := svc.dao.CloudSelectionScheme().Delete(cts.Kit, flt); err != nil {
		logs.Errorf("delete cloud selection scheme failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// CreateScheme ...
func (svc *service) CreateScheme(cts *rest.Contexts) (interface{}, error) {
	req := new(dsselection.SchemeCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	model := &tableselection.SchemeTable{
		ID:      "",
		BkBizID: req.BkBizID,
		Name:    req.Name,
		BizType: req.BizType,
		Vendors: req.Vendors,
		DeploymentArchitecture: types2.StringArray(slice.Map(req.DeploymentArchitecture,
			func(a enumor.SchemeDeployArch) string { return string(a) })),
		CoverPing:        req.CoverPing,
		CompositeScore:   req.CompositeScore,
		NetScore:         req.NetScore,
		CostScore:        req.CostScore,
		CoverRate:        req.CoverRate,
		UserDistribution: req.UserDistribution,
		ResultIdcIDs:     req.ResultIdcIDs,
		Creator:          cts.Kit.User,
		Reviser:          cts.Kit.User,
	}
	id, err := svc.dao.CloudSelectionScheme().Create(cts.Kit, model)
	if err != nil {
		logs.Errorf("create cloud selection scheme failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return &core.CreateResult{ID: id}, nil
}

// UpdateScheme ...
func (svc *service) UpdateScheme(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, fmt.Errorf("id is required")
	}

	req := new(dsselection.SchemeUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	model := &tableselection.SchemeTable{
		BkBizID: req.BkBizID,
		Name:    req.Name,
		Reviser: cts.Kit.User,
	}
	if err := svc.dao.CloudSelectionScheme().UpdateByID(cts.Kit, id, model); err != nil {
		logs.Errorf("update cloud selection scheme failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

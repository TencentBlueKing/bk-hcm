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

package eipcvmrel

import (
	"fmt"

	eipcvmrel "hcm/pkg/api/core/cloud/eip-cvm-rel"
	datarelproto "hcm/pkg/api/data-service/cloud"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table/cloud/eip"
	"hcm/pkg/rest"
)

// ListEipCvmRels ...
func (svc *relSvc) ListEipCvmRels(cts *rest.Contexts) (interface{}, error) {
	req := new(datarelproto.EipCvmRelListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	data, err := svc.dao.EipCvmRel().List(cts.Kit, opt)
	if err != nil {
		return nil, fmt.Errorf("list eip cvm rels failed, err: %v", err)
	}

	if req.Page.Count {
		return &datarelproto.EipCvmRelListResult{Count: data.Count}, nil
	}

	details := make([]*datarelproto.EipCvmRelResult, len(data.Details))
	for idx, r := range data.Details {
		details[idx] = &datarelproto.EipCvmRelResult{
			ID:        r.ID,
			EipID:     r.EipID,
			CvmID:     r.CvmID,
			Creator:   r.Creator,
			CreatedAt: r.CreatedAt.String(),
		}
	}

	return &datarelproto.EipCvmRelListResult{Details: details}, nil
}

// ListWithEip ...
func (svc *relSvc) ListWithEip(cts *rest.Contexts) (interface{}, error) {
	req := new(datarelproto.EipCvmRelWithEipListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	data, err := svc.dao.EipCvmRel().ListJoinEip(cts.Kit, req.CvmIDs)
	if err != nil {
		return nil, err
	}

	eips := make([]*datarelproto.EipWithCvmID, len(data.Details))
	for idx, d := range data.Details {
		eips[idx] = toProtoEipWithCvmID(d)
	}

	return eips, nil
}

// ListWithEipExt ...
func (svc *relSvc) ListWithEipExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(datarelproto.EipCvmRelWithEipExtListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	data, err := svc.dao.EipCvmRel().ListJoinEip(cts.Kit, req.CvmIDs)
	if err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return toProtoEipExtWithCvmIDs[dataproto.TCloudEipExtensionResult](data)
	case enumor.Aws:
		return toProtoEipExtWithCvmIDs[dataproto.AwsEipExtensionResult](data)
	case enumor.Gcp:
		return toProtoEipExtWithCvmIDs[dataproto.GcpEipExtensionResult](data)
	case enumor.Azure:
		return toProtoEipExtWithCvmIDs[dataproto.AzureEipExtensionResult](data)
	case enumor.HuaWei:
		return toProtoEipExtWithCvmIDs[dataproto.HuaWeiEipExtensionResult](data)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// ListEipWithoutCvm ...
func (svc *relSvc) ListEipWithoutCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(datarelproto.ListEipWithoutCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.EipCvmRel().ListEipLeftJoinRel(cts.Kit, opt)
	if err != nil {
		return nil, fmt.Errorf("list eip left join eip_cvm_rel failed, err: %v", err)
	}

	if req.Page.Count {
		return &datarelproto.ListEipWithoutCvmResult{Count: result.Count}, nil
	}

	details := make([]eipcvmrel.RelWithEip, len(result.Details))
	for index, one := range result.Details {
		details[index] = eipcvmrel.RelWithEip{
			EipModel: eip.EipModel{
				ID:        one.ID,
				Vendor:    one.Vendor,
				AccountID: one.AccountID,
				CloudID:   one.CloudID,
				BkBizID:   one.BkBizID,
				Name:      one.Name,
				Region:    one.Region,
				Status:    one.Status,
				PublicIp:  one.PublicIp,
				PrivateIp: one.PrivateIp,
				Creator:   one.Creator,
				Reviser:   one.Reviser,
				Extension: one.Extension,
				CreatedAt: one.CreatedAt,
				UpdatedAt: one.UpdatedAt,
			},
			RelCreator: one.RelCreator,
		}
	}

	return &datarelproto.ListEipWithoutCvmResult{Details: details}, nil
}

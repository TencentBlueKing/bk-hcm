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

// Package cfs 如下:
// cfs handler
package cfs

import (
	"fmt"

	"hcm/pkg/api/core"
	corecfs "hcm/pkg/api/core/cloud/cfs"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecfs "hcm/pkg/dal/table/cloud/cfs"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/pkg/errors"
)

// ListCfsExt cfs.
func (svc *cfsSvc) ListCfsExt(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.CfsListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("ListCfsExt decode cfs list req failed, err: %s, rid: %s", err.Error(), cts.Kit.Rid)
		return nil, errors.Wrapf(err, "ListCfsExt decode cfs list req failed, rid: %s", cts.Kit.Rid)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Field,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.Cfs().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list cfs failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list cfs failed, err: %v", err)
	}
	if req.Page.Count {

		return &protocloud.CfsExtListResp[corecfs.TCloudCfsExtension]{
			Data: &protocloud.CfsExtListResult[corecfs.TCloudCfsExtension]{
				Count: result.Count,
			},
		}, nil

		//return &protocloud.CfsExtListResult[corecfs.TCloudCfsExtension]{Count: result.Count}, nil
	}

	resp, err := convCfsListResult[corecfs.TCloudCfsExtension](result.Details)
	if err != nil {
		logs.Errorf("list cfs conv failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errors.Wrapf(err, "list cfs conv failed, err: %v", err)
	}

	return resp, nil
}

// GetCfs cfs.
func (svc *cfsSvc) GetCfs(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "cfs id is required")
	}

	cfsTable, err := svc.getCfsByID(cts.Kit, id)
	if err != nil {
		return nil, err
	}

	base := convTableToBaseCfs(cfsTable)

	switch cfsTable.Vendor {
	case enumor.TCloud:
		return convCfsGetResult[corecfs.TCloudCfsExtension](base, cfsTable.Extension)

	//case enumor.Aws:
	//	return convCfsGetResult[corecfs.AwsCfsExtension](base, cfsTable.Extension)
	//case enumor.HuaWei:
	//	return convCfsGetResult[corecfs.HuaWeiCfsExtension](base, cfsTable.Extension)
	//case enumor.Azure:
	//	return convCfsGetResult[corecfs.AzureCfsExtension](base, cfsTable.Extension)
	//case enumor.Gcp:
	//	return convCfsGetResult[corecfs.GcpCfsExtension](base, cfsTable.Extension)
	//case enumor.Other:
	//	return convCfsGetResult[corecfs.OtherCfsExtension](base, cfsTable.Extension)

	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

// convCfsGetResult get
func convCfsGetResult[T corecfs.Extension](base corecfs.BaseCfs, extJson tabletype.JsonField) (
	*corecfs.Cfs[T], error) {

	extension := new(T)
	if len(extJson) != 0 {
		if err := json.UnmarshalFromString(string(extJson), &extension); err != nil {
			return nil, fmt.Errorf("UnmarshalFromString cfs json extension failed, err: %v", err)
		}
	}

	return &corecfs.Cfs[T]{
		BaseCfs:   base,
		Extension: extension,
	}, nil
}

// getCfsByID get
func (svc *cfsSvc) getCfsByID(kt *kit.Kit, id string) (*tablecfs.Table, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.dao.Cfs().List(kt, opt)
	if err != nil {
		logs.Errorf("get cfs failed, err: %v, rid: %s", kt.Rid)
		return nil, fmt.Errorf("get cfs failed, err: %v", err)
	}

	if len(result.Details) != 1 {
		return nil, errf.New(errf.RecordNotFound, "cfs not found")
	}

	return &result.Details[0], nil
}

//// ListCfs cfs.
//func (svc *cfsSvc) ListCfs(cts *rest.Contexts) (interface{}, error) {
//	req := new(protocloud.CfsListReq)
//	if err := cts.DecodeInto(req); err != nil {
//		return nil, err
//	}
//	if err := req.Validate(); err != nil {
//		return nil, errf.NewFromErr(errf.InvalidParameter, err)
//	}
//
//	opt := &types.ListOption{
//		Fields: req.Field,
//		Filter: req.Filter,
//		Page:   req.Page,
//	}
//	result, err := svc.dao.Cfs().List(cts.Kit, opt)
//	if err != nil {
//		logs.Errorf("list cfs failed, err: %v, rid: %s", err, cts.Kit.Rid)
//		return nil, fmt.Errorf("list cfs failed, err: %v", err)
//	}
//	if req.Page.Count {
//		return &protocloud.CfsListResult{Count: result.Count}, nil
//	}
//
//	logs.Infof("list cfs result: %v, rid: %s", result, cts.Kit.Rid) // note: debug log
//
//	details := make([]*corecfs.BaseCfs, 0, len(result.Details))
//	for _, one := range result.Details {
//		details = append(details, convTableToBaseCfs(one))
//	}
//
//	return &protocloud.CfsListResult{Details: details, Count: uint64(len(result.Details))}, nil
//}

// convCfsListResult list
func convCfsListResult[T corecfs.Extension](tables []tablecfs.Table) (*protocloud.CfsExtListResult[T], error) {
	details := make([]corecfs.Cfs[T], 0, len(tables))

	for _, one := range tables {
		extension := new(T)
		if len(one.Extension) != 0 {
			if err := json.UnmarshalFromString(string(one.Extension), &extension); err != nil {
				return nil, fmt.Errorf("UnmarshalFromString cfs json extension failed, err: %v", err)
			}
		}

		details = append(details, corecfs.Cfs[T]{
			Extension: extension,
			BaseCfs:   convTableToBaseCfs(&one),
		})
	}

	return &protocloud.CfsExtListResult[T]{
		Details: details,
		Count:   uint64(len(details)),
	}, nil
}

// convTableToBaseCfs table to base
func convTableToBaseCfs(one *tablecfs.Table) corecfs.BaseCfs {
	base := corecfs.BaseCfs{
		ID:        one.ID,
		BkBizID:   one.BkBizID,
		AccountID: one.AccountID,
		Vendor:    one.Vendor,
		//
		CloudID:        one.CloudID,
		Name:           one.Name,
		Region:         one.Region,
		Zone:           one.Zone,
		SizeLimit:      one.SizeLimit,
		SizeByte:       one.SizeByte,
		AvailCapacity:  one.AvailCapacity,
		BandwidthLimit: one.BandwidthLimit,
		Protocol:       one.Protocol,
		StorageType:    one.StorageType,
		Encrypted:      one.Encrypted,
		CryptKeyId:     one.CryptKeyId,
		CloudVpcIDs:    one.CloudVpcIDs,
		VpcIDs:         one.VpcIDs,
		CloudSubnetIDs: one.CloudSubnetIDs,
		SubnetIDs:      one.SubnetIDs,
		//
		Status:           one.Status,
		Memo:             one.Memo,
		CloudCreatedTime: one.CloudCreatedTime,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}

	return base
}

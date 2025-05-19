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

package cvm

import (
	"fmt"

	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecvm "hcm/pkg/dal/table/cloud/cvm"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

// ListCvm cvm.
func (svc *cvmSvc) ListCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.CvmListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Field,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.Cvm().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list cvm failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.CvmListResult{Count: result.Count}, nil
	}

	details := make([]corecvm.BaseCvm, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, *convTableToBaseCvm(&one))
	}

	return &protocloud.CvmListResult{Details: details}, nil
}

// GetCvm cvm.
func (svc *cvmSvc) GetCvm(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "cvm id is required")
	}

	cvmTable, err := svc.getCvmByID(cts.Kit, id)
	if err != nil {
		return nil, err
	}

	// TODO: 添加查询关联信息逻辑

	base := convTableToBaseCvm(cvmTable)

	switch cvmTable.Vendor {
	case enumor.TCloud:
		return convCvmGetResult[corecvm.TCloudCvmExtension](base, cvmTable.Extension)
	case enumor.Aws:
		return convCvmGetResult[corecvm.AwsCvmExtension](base, cvmTable.Extension)
	case enumor.HuaWei:
		return convCvmGetResult[corecvm.HuaWeiCvmExtension](base, cvmTable.Extension)
	case enumor.Azure:
		return convCvmGetResult[corecvm.AzureCvmExtension](base, cvmTable.Extension)
	case enumor.Gcp:
		return convCvmGetResult[corecvm.GcpCvmExtension](base, cvmTable.Extension)
	case enumor.Other:
		return convCvmGetResult[corecvm.OtherCvmExtension](base, cvmTable.Extension)

	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

func convCvmGetResult[T corecvm.Extension](base *corecvm.BaseCvm, extJson tabletype.JsonField) (
	*corecvm.Cvm[T], error) {

	extension := new(T)
	if len(extJson) != 0 {
		if err := json.UnmarshalFromString(string(extJson), &extension); err != nil {
			return nil, fmt.Errorf("UnmarshalFromString cvm json extension failed, err: %v", err)
		}
	}

	return &corecvm.Cvm[T]{
		BaseCvm:   *base,
		Extension: extension,
	}, nil
}

func convTableToBaseCvm(one *tablecvm.Table) *corecvm.BaseCvm {
	base := &corecvm.BaseCvm{
		ID:                   one.ID,
		CloudID:              one.CloudID,
		Name:                 one.Name,
		Vendor:               one.Vendor,
		BkBizID:              one.BkBizID,
		BkHostID:             one.BkHostID,
		BkCloudID:            converter.PtrToVal(one.BkCloudID),
		AccountID:            one.AccountID,
		Region:               one.Region,
		Zone:                 one.Zone,
		CloudVpcIDs:          one.CloudVpcIDs,
		VpcIDs:               one.VpcIDs,
		CloudSubnetIDs:       one.CloudSubnetIDs,
		SubnetIDs:            one.SubnetIDs,
		CloudImageID:         one.CloudImageID,
		ImageID:              one.ImageID,
		OsName:               one.OsName,
		Memo:                 one.Memo,
		Status:               one.Status,
		RecycleStatus:        one.RecycleStatus,
		PrivateIPv4Addresses: one.PrivateIPv4Addresses,
		PrivateIPv6Addresses: one.PrivateIPv6Addresses,
		PublicIPv4Addresses:  one.PublicIPv4Addresses,
		PublicIPv6Addresses:  one.PublicIPv6Addresses,
		MachineType:          one.MachineType,
		CloudCreatedTime:     one.CloudCreatedTime,
		CloudLaunchedTime:    one.CloudLaunchedTime,
		CloudExpiredTime:     one.CloudExpiredTime,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}

	return base
}

func (svc *cvmSvc) getCvmByID(kt *kit.Kit, id string) (*tablecvm.Table, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.dao.Cvm().List(kt, opt)
	if err != nil {
		logs.Errorf("list cvm failed, err: %v, rid: %s", kt.Rid)
		return nil, fmt.Errorf("list cvm failed, err: %v", err)
	}

	if len(result.Details) != 1 {
		return nil, errf.New(errf.RecordNotFound, "cvm not found")
	}

	return &result.Details[0], nil
}

// ListCvmExt cvm with extension.
func (svc *cvmSvc) ListCvmExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(protocloud.CvmExtListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Field,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.Cvm().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list cvm failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.CvmExtListResult[corecvm.TCloudCvmExtension]{Count: result.Count}, nil
	}

	switch vendor {
	case enumor.TCloud:
		return convCvmListResult[corecvm.TCloudCvmExtension](result.Details)
	case enumor.Aws:
		return convCvmListResult[corecvm.AwsCvmExtension](result.Details)
	case enumor.HuaWei:
		return convCvmListResult[corecvm.HuaWeiCvmExtension](result.Details)
	case enumor.Azure:
		return convCvmListResult[corecvm.AzureCvmExtension](result.Details)
	case enumor.Gcp:
		return convCvmListResult[corecvm.GcpCvmExtension](result.Details)

	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

func convCvmListResult[T corecvm.Extension](tables []tablecvm.Table) (*protocloud.CvmExtListResult[T], error) {

	details := make([]corecvm.Cvm[T], 0, len(tables))
	for _, one := range tables {
		extension := new(T)
		if len(one.Extension) != 0 {
			if err := json.UnmarshalFromString(string(one.Extension), &extension); err != nil {
				return nil, fmt.Errorf("UnmarshalFromString cvm json extension failed, err: %v", err)
			}
		}

		details = append(details, corecvm.Cvm[T]{
			BaseCvm:   *convTableToBaseCvm(&one),
			Extension: extension,
		})
	}

	return &protocloud.CvmExtListResult[T]{
		Details: details,
	}, nil
}

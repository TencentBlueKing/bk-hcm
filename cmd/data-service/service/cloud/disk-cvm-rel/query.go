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

package diskcvmrel

import (
	"fmt"

	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	coredisk "hcm/pkg/api/core/cloud/disk"
	diskcvmrel "hcm/pkg/api/core/cloud/disk-cvm-rel"
	datarelproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table/cloud/disk"
	"hcm/pkg/rest"
)

// ListDiskCvmRel ...
func (svc *relSvc) ListDiskCvmRel(cts *rest.Contexts) (interface{}, error) {
	req := new(datarelproto.DiskCvmRelListReq)
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
	data, err := svc.dao.DiskCvmRel().List(cts.Kit, opt)
	if err != nil {
		return nil, fmt.Errorf("list disk cvm rels failed, err: %v", err)
	}

	if req.Page.Count {
		return &datarelproto.DiskCvmRelListResult{Count: data.Count}, nil
	}

	details := make([]*datarelproto.DiskCvmRelResult, len(data.Details))
	for idx, r := range data.Details {
		details[idx] = &datarelproto.DiskCvmRelResult{
			ID:        r.ID,
			DiskID:    r.DiskID,
			CvmID:     r.CvmID,
			Creator:   r.Creator,
			CreatedAt: r.CreatedAt.String(),
		}
	}

	return &datarelproto.DiskCvmRelListResult{Details: details}, nil
}

// ListWithCvm ...
func (svc *relSvc) ListWithCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(datarelproto.ListWithCvmReq)
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
	result, err := svc.dao.DiskCvmRel().ListCvmIDLeftJoinRel(cts.Kit, opt, req.NotEqualDiskID)
	if err != nil {
		return nil, fmt.Errorf("list cvm left join disk_cvm_rel failed, err: %v", err)
	}

	if req.Page.Count {
		return &datarelproto.ListCvmResult{Count: result.Count}, nil
	}

	details := make([]corecvm.BaseCvm, len(result.Details))
	for index, one := range result.Details {
		details[index] = corecvm.BaseCvm{
			ID:                   one.ID,
			CloudID:              one.CloudID,
			Name:                 one.Name,
			Vendor:               one.Vendor,
			BkBizID:              one.BkBizID,
			BkHostID:             one.BkHostID,
			BkCloudID:            *one.BkCloudID,
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
	}

	return &datarelproto.ListCvmResult{Details: details}, nil
}

// ListDiskWithoutCvm ...
func (svc *relSvc) ListDiskWithoutCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(datarelproto.ListDiskWithoutCvmReq)
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
	result, err := svc.dao.DiskCvmRel().ListDiskLeftJoinRel(cts.Kit, opt)
	if err != nil {
		return nil, fmt.Errorf("list cvm left join disk_cvm_rel failed, err: %v", err)
	}

	if req.Page.Count {
		return &datarelproto.ListDiskWithoutCvmResult{Count: result.Count}, nil
	}

	details := make([]diskcvmrel.RelWithDisk, len(result.Details))
	for index, one := range result.Details {
		details[index] = diskcvmrel.RelWithDisk{
			DiskModel: disk.DiskModel{
				ID:            one.ID,
				Vendor:        one.Vendor,
				AccountID:     one.AccountID,
				CloudID:       one.CloudID,
				BkBizID:       one.BkBizID,
				Name:          one.Name,
				Region:        one.Region,
				Zone:          one.Zone,
				DiskSize:      one.DiskSize,
				DiskType:      one.DiskType,
				Status:        one.Status,
				RecycleStatus: one.RecycleStatus,
				IsSystemDisk:  one.IsSystemDisk,
				Memo:          one.Memo,
				Extension:     one.Extension,
				Creator:       one.Creator,
				Reviser:       one.Reviser,
				CreatedAt:     one.CreatedAt,
				UpdatedAt:     one.UpdatedAt,
			},
			RelCreator: one.RelCreator,
		}
	}

	return &datarelproto.ListDiskWithoutCvmResult{Details: details}, nil
}

// ListWithDisk ...
func (svc *relSvc) ListWithDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(datarelproto.DiskCvmRelWithDiskListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	data, err := svc.dao.DiskCvmRel().ListJoinDisk(cts.Kit, req.CvmIDs)
	if err != nil {
		return nil, err
	}

	disks := make([]*datarelproto.DiskWithCvmID, len(data.Details))
	for idx, d := range data.Details {
		disks[idx] = toProtoDiskWithCvmID(d)
	}
	return disks, nil
}

// ListWithDiskExt ...
func (svc *relSvc) ListWithDiskExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(datarelproto.DiskCvmRelWithDiskExtListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	data, err := svc.dao.DiskCvmRel().ListJoinDisk(cts.Kit, req.CvmIDs)
	if err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return toProtoDiskExtWithCvmIDs[coredisk.TCloudExtension](data)
	case enumor.Aws:
		return toProtoDiskExtWithCvmIDs[coredisk.AwsExtension](data)
	case enumor.Gcp:
		return toProtoDiskExtWithCvmIDs[coredisk.GcpExtension](data)
	case enumor.Azure:
		return toProtoDiskExtWithCvmIDs[coredisk.AzureExtension](data)
	case enumor.HuaWei:
		return toProtoDiskExtWithCvmIDs[coredisk.HuaWeiExtension](data)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

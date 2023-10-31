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
	coredisk "hcm/pkg/api/core/cloud/disk"
	"hcm/pkg/api/data-service/cloud"
	reltypes "hcm/pkg/dal/dao/types/cloud"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

func toProtoDiskExtWithCvmIDs[T coredisk.Extension](
	data *reltypes.DiskCvmRelJoinDiskListResult,
) ([]*cloud.DiskExtWithCvmID[T], error) {
	details := make([]*cloud.DiskExtWithCvmID[T], len(data.Details))
	for idx, d := range data.Details {
		extResult, err := toProtoDiskExtWithCvmID[T](d)
		if err != nil {
			return nil, err
		}
		details[idx] = extResult
	}
	return details, nil
}

func toProtoDiskExtWithCvmID[T coredisk.Extension](
	d *reltypes.DiskWithCvmID,
) (*cloud.DiskExtWithCvmID[T], error) {
	extension := new(T)
	err := json.UnmarshalFromString(string(d.Extension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString db extension failed, err: %v", err)
	}
	return &cloud.DiskExtWithCvmID[T]{
		Disk: coredisk.Disk[T]{
			BaseDisk: coredisk.BaseDisk{
				ID:           d.ID,
				Vendor:       d.Vendor,
				AccountID:    d.AccountID,
				Name:         d.Name,
				BkBizID:      d.BkBizID,
				CloudID:      d.CloudID,
				Region:       d.Region,
				Zone:         d.Zone,
				DiskSize:     d.DiskSize,
				DiskType:     d.DiskType,
				Status:       d.Status,
				IsSystemDisk: converter.PtrToVal(d.IsSystemDisk),
				Memo:         d.Memo,
				Revision: core.Revision{
					Creator:   d.Creator,
					Reviser:   d.Reviser,
					CreatedAt: d.CreatedAt.String(),
					UpdatedAt: d.UpdatedAt.String(),
				},
			},
			Extension: extension,
		},
		CvmID:        d.CvmID,
		RelCreator:   d.RelCreator,
		RelCreatedAt: d.RelCreatedAt.String(),
	}, nil
}

func toProtoDiskWithCvmID(d *reltypes.DiskWithCvmID) *cloud.DiskWithCvmID {
	return &cloud.DiskWithCvmID{
		BaseDisk: coredisk.BaseDisk{
			ID:           d.ID,
			Vendor:       d.Vendor,
			AccountID:    d.AccountID,
			Name:         d.Name,
			BkBizID:      d.BkBizID,
			CloudID:      d.CloudID,
			Region:       d.Region,
			Zone:         d.Zone,
			DiskSize:     d.DiskSize,
			DiskType:     d.DiskType,
			Status:       d.Status,
			IsSystemDisk: converter.PtrToVal(d.IsSystemDisk),
			Memo:         d.Memo,
			Revision: core.Revision{
				Creator:   d.Creator,
				Reviser:   d.Reviser,
				CreatedAt: d.CreatedAt.String(),
				UpdatedAt: d.UpdatedAt.String(),
			},
		},
		CvmID:        d.CvmID,
		RelCreator:   d.RelCreator,
		RelCreatedAt: d.RelCreatedAt.String(),
	}
}

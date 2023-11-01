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

package disk

import (
	"fmt"

	"hcm/pkg/api/core"
	coredisk "hcm/pkg/api/core/cloud/disk"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	"hcm/pkg/dal/dao/types/cloud"
	tablecloud "hcm/pkg/dal/table/cloud/disk"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

func toProtoDiskExtListResult[T coredisk.Extension](
	data *cloud.DiskListResult,
) (*dataproto.ListExtResult[T], error) {
	details := make([]*coredisk.Disk[T], len(data.Details))
	for indx, d := range data.Details {
		extResult, err := toProtoDiskExtResult[T](d)
		if err != nil {
			return nil, err
		}
		details[indx] = extResult
	}

	return &dataproto.ListExtResult[T]{Count: data.Count, Details: details}, nil
}

func toProtoDiskExtResult[T coredisk.Extension](
	m *tablecloud.DiskModel,
) (*coredisk.Disk[T], error) {
	extension := new(T)

	err := json.UnmarshalFromString(string(m.Extension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString db extension failed, err: %v", err)
	}
	return &coredisk.Disk[T]{
		BaseDisk: coredisk.BaseDisk{
			ID:           m.ID,
			Vendor:       m.Vendor,
			AccountID:    m.AccountID,
			Name:         m.Name,
			BkBizID:      m.BkBizID,
			CloudID:      m.CloudID,
			Region:       m.Region,
			Zone:         m.Zone,
			DiskSize:     m.DiskSize,
			DiskType:     m.DiskType,
			Status:       m.Status,
			IsSystemDisk: converter.PtrToVal(m.IsSystemDisk),
			Memo:         m.Memo,
			Revision: core.Revision{
				Creator:   m.Creator,
				Reviser:   m.Reviser,
				CreatedAt: m.CreatedAt.String(),
				UpdatedAt: m.UpdatedAt.String(),
			},
		},
		Extension: extension,
	}, nil
}

func toProtoDiskResult(m *tablecloud.DiskModel) *coredisk.BaseDisk {
	return &coredisk.BaseDisk{
		ID:            m.ID,
		Vendor:        m.Vendor,
		AccountID:     m.AccountID,
		Name:          m.Name,
		BkBizID:       m.BkBizID,
		CloudID:       m.CloudID,
		Region:        m.Region,
		Zone:          m.Zone,
		DiskSize:      m.DiskSize,
		DiskType:      m.DiskType,
		Status:        m.Status,
		RecycleStatus: m.RecycleStatus,
		IsSystemDisk:  converter.PtrToVal(m.IsSystemDisk),
		Memo:          m.Memo,
		Revision: core.Revision{
			Creator:   m.Creator,
			Reviser:   m.Reviser,
			CreatedAt: m.CreatedAt.String(),
			UpdatedAt: m.UpdatedAt.String(),
		},
	}
}

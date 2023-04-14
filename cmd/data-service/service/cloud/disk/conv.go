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

	dataproto "hcm/pkg/api/data-service/cloud/disk"
	"hcm/pkg/dal/dao/types/cloud"
	tablecloud "hcm/pkg/dal/table/cloud/disk"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

func toProtoDiskExtListResult[T dataproto.DiskExtensionResult](
	data *cloud.DiskListResult,
) (*dataproto.DiskExtListResult[T], error) {
	details := make([]*dataproto.DiskExtResult[T], len(data.Details))
	for indx, d := range data.Details {
		extResult, err := toProtoDiskExtResult[T](d)
		if err != nil {
			return nil, err
		}
		details[indx] = extResult
	}

	return &dataproto.DiskExtListResult[T]{Count: data.Count, Details: details}, nil
}

func toProtoDiskExtResult[T dataproto.DiskExtensionResult](
	m *tablecloud.DiskModel,
) (*dataproto.DiskExtResult[T], error) {
	extension := new(T)
	err := json.UnmarshalFromString(string(m.Extension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString db extension failed, err: %v", err)
	}
	return &dataproto.DiskExtResult[T]{
		ID:           m.ID,
		Vendor:       m.Vendor,
		AccountID:    m.AccountID,
		BkBizID:      m.BkBizID,
		CloudID:      m.CloudID,
		Name:         m.Name,
		Region:       m.Region,
		Zone:         m.Zone,
		DiskSize:     m.DiskSize,
		DiskType:     m.DiskType,
		Status:       m.Status,
		IsSystemDisk: converter.PtrToVal(m.IsSystemDisk),
		Memo:         m.Memo,
		Creator:      m.Creator,
		Reviser:      m.Reviser,
		CreatedAt:    m.CreatedAt.String(),
		UpdatedAt:    m.UpdatedAt.String(),
		Extension:    extension,
	}, nil
}

func toProtoDiskResult(m *tablecloud.DiskModel) *dataproto.DiskResult {
	return &dataproto.DiskResult{
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
		Creator:      m.Creator,
		Reviser:      m.Reviser,
		CreatedAt:    m.CreatedAt.String(),
		UpdatedAt:    m.UpdatedAt.String(),
	}
}

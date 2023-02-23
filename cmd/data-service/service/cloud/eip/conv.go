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

package eip

import (
	"fmt"

	dataproto "hcm/pkg/api/data-service/cloud/eip"
	"hcm/pkg/dal/dao/types/cloud"
	tablecloud "hcm/pkg/dal/table/cloud/eip"
	"hcm/pkg/tools/json"
)

func toProtoEipExtListResult[T dataproto.EipExtensionResult](
	data *cloud.EipListResult,
) (*dataproto.EipExtListResult[T], error) {
	details := make([]*dataproto.EipExtResult[T], len(data.Details))
	for indx, d := range data.Details {
		extResult, err := toProtoEipExtResult[T](d)
		if err != nil {
			return nil, err
		}
		details[indx] = extResult
	}

	return &dataproto.EipExtListResult[T]{Count: data.Count, Details: details}, nil
}

func toProtoEipExtResult[T dataproto.EipExtensionResult](m *tablecloud.EipModel) (*dataproto.EipExtResult[T], error) {
	extension := new(T)
	err := json.UnmarshalFromString(string(m.Extension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString db extension failed, err: %v", err)
	}
	return &dataproto.EipExtResult[T]{
		ID:           m.ID,
		AccountID:    m.AccountID,
		Vendor:       m.Vendor,
		Name:         m.Name,
		CloudID:      m.CloudID,
		BkBizID:      m.BkBizID,
		Region:       m.Region,
		InstanceId:   m.InstanceId,
		InstanceType: m.InstanceType,
		Status:       m.Status,
		PublicIp:     m.PublicIp,
		PrivateIp:    m.PrivateIp,
		Creator:      m.Creator,
		Reviser:      m.Reviser,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		Extension:    extension,
	}, nil
}

func toProtoEipResult(m *tablecloud.EipModel) *dataproto.EipResult {
	return &dataproto.EipResult{
		ID:           m.ID,
		AccountID:    m.AccountID,
		Vendor:       m.Vendor,
		Name:         m.Name,
		CloudID:      m.CloudID,
		BkBizID:      m.BkBizID,
		Region:       m.Region,
		InstanceId:   m.InstanceId,
		InstanceType: m.InstanceType,
		Status:       m.Status,
		PublicIp:     m.PublicIp,
		PrivateIp:    m.PrivateIp,
		Creator:      m.Creator,
		Reviser:      m.Reviser,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

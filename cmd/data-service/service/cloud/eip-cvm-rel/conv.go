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

	"hcm/pkg/api/data-service/cloud"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	reltypes "hcm/pkg/dal/dao/types/cloud"
	"hcm/pkg/tools/json"
)

func toProtoEipExtWithCvmIDs[T dataproto.EipExtensionResult](
	data *reltypes.EipCvmRelJoinEipListResult,
) ([]*cloud.EipExtWithCvmID[T], error) {
	details := make([]*cloud.EipExtWithCvmID[T], len(data.Details))
	for idx, d := range data.Details {
		extResult, err := toProtoEipExtWithCvmID[T](d)
		if err != nil {
			return nil, err
		}
		details[idx] = extResult
	}
	return details, nil
}

func toProtoEipExtWithCvmID[T dataproto.EipExtensionResult](
	d *reltypes.EipWithCvmID,
) (*cloud.EipExtWithCvmID[T], error) {
	extension := new(T)
	err := json.UnmarshalFromString(string(d.Extension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString db extension failed, err: %v", err)
	}
	return &cloud.EipExtWithCvmID[T]{
		EipExtResult: dataproto.EipExtResult[T]{
			ID:        d.ID,
			Vendor:    d.Vendor,
			CloudID:   d.CloudID,
			AccountID: d.AccountID,
			Name:      d.Name,
			BkBizID:   d.BkBizID,
			Region:    d.Region,
			Status:    d.Status,
			PublicIp:  d.PublicIp,
			PrivateIp: d.PrivateIp,
			Creator:   d.Creator,
			Reviser:   d.Reviser,
			CreatedAt: d.CreatedAt.String(),
			UpdatedAt: d.UpdatedAt.String(),
			Extension: extension,
		},
		CvmID:        d.CvmID,
		RelCreator:   d.RelCreator,
		RelCreatedAt: d.RelCreatedAt.String(),
	}, nil
}

func toProtoEipWithCvmID(d *reltypes.EipWithCvmID) *cloud.EipWithCvmID {
	return &cloud.EipWithCvmID{
		EipResult: dataproto.EipResult{
			ID:        d.ID,
			Vendor:    d.Vendor,
			CloudID:   d.CloudID,
			AccountID: d.AccountID,
			Name:      d.Name,
			BkBizID:   d.BkBizID,
			Region:    d.Region,
			Status:    d.Status,
			PublicIp:  d.PublicIp,
			PrivateIp: d.PrivateIp,
			Creator:   d.Creator,
			Reviser:   d.Reviser,
			CreatedAt: d.CreatedAt.String(),
			UpdatedAt: d.UpdatedAt.String(),
		},
		CvmID:        d.CvmID,
		RelCreator:   d.RelCreator,
		RelCreatedAt: d.RelCreatedAt.String(),
	}
}

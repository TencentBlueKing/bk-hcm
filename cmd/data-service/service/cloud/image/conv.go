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

package image

import (
	"fmt"

	"hcm/pkg/api/core"
	coreimage "hcm/pkg/api/core/cloud/image"
	dataproto "hcm/pkg/api/data-service/cloud/image"
	"hcm/pkg/dal/dao/types/cloud"
	tablecloud "hcm/pkg/dal/table/cloud/image"
	"hcm/pkg/tools/json"
)

func toProtoImageExtListResult[T coreimage.Extension](data *cloud.ImageListResult) (
	*dataproto.ListExtResult[T], error) {

	details := make([]*coreimage.Image[T], len(data.Details))
	for index, d := range data.Details {
		extResult, err := toProtoImageExtResult[T](d)
		if err != nil {
			return nil, err
		}
		details[index] = extResult
	}
	return &dataproto.ListExtResult[T]{Count: data.Count, Details: details}, nil
}

func toProtoImageExtResult[T coreimage.Extension](m *tablecloud.ImageModel) (*coreimage.Image[T], error) {

	extension := new(T)
	err := json.UnmarshalFromString(string(m.Extension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString db extension failed, err: %v", err)
	}

	return &coreimage.Image[T]{
		BaseImage: coreimage.BaseImage{
			ID:           m.ID,
			Vendor:       m.Vendor,
			CloudID:      m.CloudID,
			Name:         m.Name,
			Architecture: m.Architecture,
			Platform:     m.Platform,
			State:        m.State,
			Type:         m.Type,
			OsType:       m.OsType,
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

func toProtoImageResult(m *tablecloud.ImageModel) *coreimage.BaseImage {
	return &coreimage.BaseImage{
		ID:           m.ID,
		Vendor:       m.Vendor,
		CloudID:      m.CloudID,
		Name:         m.Name,
		Architecture: m.Architecture,
		Platform:     m.Platform,
		State:        m.State,
		Type:         m.Type,
		OsType:       m.OsType,
		Revision: core.Revision{
			Creator:   m.Creator,
			Reviser:   m.Reviser,
			CreatedAt: m.CreatedAt.String(),
			UpdatedAt: m.UpdatedAt.String(),
		},
	}
}

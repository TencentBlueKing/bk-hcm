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

package cslb

import (
	"testing"

	"hcm/pkg/criteria/enumor"

	"github.com/stretchr/testify/require"
)

func TestListExclusiveClusterTagsReq_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		req     ListExclusiveClusterTagsReq
		wantErr bool
	}{
		{
			name: "valid required fields only",
			req: ListExclusiveClusterTagsReq{
				AccountID: "acc-1", Region: "ap-guangzhou", Isp: enumor.BGPClusterIsp,
			},
		},
		{
			name: "valid single zone without back_zones",
			req: ListExclusiveClusterTagsReq{
				AccountID: "acc-1", Region: "ap-guangzhou", Isp: enumor.BGPClusterIsp,
				Zones: []string{"ap-guangzhou-1"},
			},
		},
		{
			name: "valid master and slave zones",
			req: ListExclusiveClusterTagsReq{
				AccountID: "acc-1", Region: "ap-guangzhou", Isp: enumor.BGPClusterIsp,
				Zones: []string{"ap-guangzhou-1"}, BackZones: []string{"ap-guangzhou-2"},
			},
		},
		{
			name: "empty zone element is invalid",
			req: ListExclusiveClusterTagsReq{
				AccountID: "acc-1", Region: "ap-guangzhou", Isp: enumor.BGPClusterIsp,
				Zones: []string{""},
			},
			wantErr: true,
		},
		{
			name: "empty back_zone element is invalid",
			req: ListExclusiveClusterTagsReq{
				AccountID: "acc-1", Region: "ap-guangzhou", Isp: enumor.BGPClusterIsp,
				BackZones: []string{""},
			},
			wantErr: true,
		},
		{
			name: "missing isp is invalid",
			req: ListExclusiveClusterTagsReq{
				AccountID: "acc-1", Region: "ap-guangzhou",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

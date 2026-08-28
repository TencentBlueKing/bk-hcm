/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package account_test

import (
	"testing"

	"hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core"

	"github.com/stretchr/testify/assert"
)

func TestResCondSyncReq_Validate(t *testing.T) {
	testCases := []struct {
		name       string
		req        account.ResCondSyncReq
		needRegion bool
		wantErr    bool
	}{
		{
			name: "regions required for subnet",
			req: account.ResCondSyncReq{
				Regions: []string{},
			},
			needRegion: true,
			wantErr:    true,
		},
		{
			name: "regions max 5",
			req: account.ResCondSyncReq{
				Regions: []string{"r1", "r2", "r3", "r4", "r5", "r6"},
			},
			needRegion: true,
			wantErr:    true,
		},
		{
			name: "cloud_ids and tag_filters mutually exclusive",
			req: account.ResCondSyncReq{
				Regions:    []string{"ap-guangzhou"},
				CloudIDs:   []string{"subnet-1"},
				TagFilters: core.MultiValueTagMap{"biz": {"1"}},
			},
			needRegion: true,
			wantErr:    true,
		},
		{
			name: "cloud_ids requires single region",
			req: account.ResCondSyncReq{
				Regions:  []string{"ap-guangzhou", "ap-shanghai"},
				CloudIDs: []string{"subnet-1"},
			},
			needRegion: true,
			wantErr:    true,
		},
		{
			name: "valid regions only",
			req: account.ResCondSyncReq{
				Regions: []string{"ap-guangzhou"},
			},
			needRegion: true,
			wantErr:    false,
		},
		{
			name: "valid cloud_ids with single region",
			req: account.ResCondSyncReq{
				Regions:  []string{"ap-guangzhou"},
				CloudIDs: []string{"subnet-1"},
			},
			needRegion: true,
			wantErr:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate(tc.needRegion)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

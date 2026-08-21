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

package region

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func boolPtr(v bool) *bool {
	return &v
}

func TestRegionBatchUpdateSyncEnableReq_Validate_EmptyIDs(t *testing.T) {
	req := &RegionBatchUpdateSyncEnableReq{
		IDs:        []string{},
		SyncEnable: boolPtr(false),
	}

	err := req.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "ids is required")
}

func TestRegionBatchUpdateSyncEnableReq_Validate_NilIDs(t *testing.T) {
	req := &RegionBatchUpdateSyncEnableReq{
		IDs:        nil,
		SyncEnable: boolPtr(true),
	}

	err := req.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "ids is required")
}

func TestRegionBatchUpdateSyncEnableReq_Validate_TooManyIDs(t *testing.T) {
	ids := make([]string, 101)
	for i := 0; i < 101; i++ {
		ids[i] = "id-" + string(rune(i))
	}

	req := &RegionBatchUpdateSyncEnableReq{
		IDs:        ids,
		SyncEnable: boolPtr(false),
	}

	err := req.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "ids count should <= 100")
}

func TestRegionBatchUpdateSyncEnableReq_Validate_NilSyncEnable(t *testing.T) {
	req := &RegionBatchUpdateSyncEnableReq{
		IDs:        []string{"id-1"},
		SyncEnable: nil,
	}

	err := req.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "SyncEnable")
}

func TestRegionBatchUpdateSyncEnableReq_Validate_ValidRequest(t *testing.T) {
	req := &RegionBatchUpdateSyncEnableReq{
		IDs:        []string{"id-1", "id-2"},
		SyncEnable: boolPtr(true),
	}

	err := req.Validate()
	require.NoError(t, err)
}

func TestRegionBatchUpdateSyncEnableReq_Validate_MaxIDs(t *testing.T) {
	ids := make([]string, 100)
	for i := 0; i < 100; i++ {
		ids[i] = "id-" + string(rune(i))
	}

	req := &RegionBatchUpdateSyncEnableReq{
		IDs:        ids,
		SyncEnable: boolPtr(false),
	}

	err := req.Validate()
	require.NoError(t, err)
}

func TestRegionBatchUpdateSyncEnableReq_Validate_SingleID(t *testing.T) {
	req := &RegionBatchUpdateSyncEnableReq{
		IDs:        []string{"region-id-1"},
		SyncEnable: boolPtr(false),
	}

	err := req.Validate()
	require.NoError(t, err)
}

/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package account

import (
	"testing"

	"hcm/pkg/api/core/cloud/region"

	"github.com/stretchr/testify/require"
)

func TestBuildAwsCondSyncRegionRules(t *testing.T) {
	t.Run("specified regions", func(t *testing.T) {
		rules := buildAwsCondSyncRegionRules("acc-1", []string{"us-east-1", "us-west-2"})
		require.Len(t, rules, 2)
		require.Equal(t, "account_id", rules[0].Field)
		require.Equal(t, "acc-1", rules[0].Value)
		require.Equal(t, "region_id", rules[1].Field)
		require.Equal(t, []string{"us-east-1", "us-west-2"}, rules[1].Value)
	})

	t.Run("empty regions", func(t *testing.T) {
		rules := buildAwsCondSyncRegionRules("acc-1", nil)
		require.Len(t, rules, 2)
		require.Equal(t, "account_id", rules[0].Field)
		require.Equal(t, "sync_enable", rules[1].Field)
		require.Equal(t, true, rules[1].Value)
	})
}

func TestCheckAwsRequestRegions(t *testing.T) {
	regionList := []region.AwsRegion{
		{RegionID: "us-east-1", SyncEnable: true},
		{RegionID: "me-south-1", SyncEnable: false},
	}

	t.Run("enabled", func(t *testing.T) {
		err := checkAwsRequestRegions([]string{"us-east-1"}, regionList)
		require.NoError(t, err)
	})

	t.Run("not exist", func(t *testing.T) {
		err := checkAwsRequestRegions([]string{"us-west-2"}, regionList)
		require.Error(t, err)
		require.Contains(t, err.Error(), "don't exist")
		require.Contains(t, err.Error(), "us-west-2")
		require.NotContains(t, err.Error(), "sync is disabled")
	})

	t.Run("sync disabled", func(t *testing.T) {
		err := checkAwsRequestRegions([]string{"me-south-1"}, regionList)
		require.Error(t, err)
		require.Contains(t, err.Error(), "sync is disabled")
		require.Contains(t, err.Error(), "me-south-1")
		require.NotContains(t, err.Error(), "don't exist")
	})

	t.Run("both", func(t *testing.T) {
		err := checkAwsRequestRegions([]string{"us-west-2", "me-south-1"}, regionList)
		require.Error(t, err)
		require.Contains(t, err.Error(), "don't exist")
		require.Contains(t, err.Error(), "sync is disabled")
	})
}

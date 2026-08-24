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

package aws

import (
	"testing"

	protocore "hcm/pkg/api/core/cloud/region"
	"hcm/pkg/runtime/filter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildAccountRegionListFilter(t *testing.T) {
	accountID := "test-account-123"
	filterExpr := buildAccountRegionListFilter(accountID)

	require.NotNil(t, filterExpr)
	require.Len(t, filterExpr.Rules, 1)

	rule, ok := filterExpr.Rules[0].(filter.AtomRule)
	require.True(t, ok)
	assert.Equal(t, "account_id", rule.Field)
	assert.Equal(t, accountID, rule.Value)
}

func TestListRegion_OnlyEnabledRegions(t *testing.T) {
	details := []protocore.AwsRegion{
		{RegionID: "us-east-1", SyncEnable: true},
		{RegionID: "me-south-1", SyncEnable: false},
		{RegionID: "us-west-2", SyncEnable: true},
	}

	regions, disabledRegions, err := parseSyncEnabledRegions(details)
	require.NoError(t, err)
	assert.Equal(t, []string{"us-east-1", "us-west-2"}, regions)
	assert.Equal(t, []string{"me-south-1"}, disabledRegions)
}

func TestListRegion_AllDisabled_ReturnsError(t *testing.T) {
	details := []protocore.AwsRegion{
		{RegionID: "me-south-1", SyncEnable: false},
	}

	regions, disabledRegions, err := parseSyncEnabledRegions(details)
	require.Error(t, err)
	assert.Equal(t, "aws region is empty", err.Error())
	assert.Nil(t, regions)
	assert.Equal(t, []string{"me-south-1"}, disabledRegions)
}

func TestListRegion_AllEnabled(t *testing.T) {
	details := []protocore.AwsRegion{
		{RegionID: "us-east-1", SyncEnable: true},
		{RegionID: "us-west-2", SyncEnable: true},
	}

	regions, disabledRegions, err := parseSyncEnabledRegions(details)
	require.NoError(t, err)
	assert.Equal(t, []string{"us-east-1", "us-west-2"}, regions)
	assert.Empty(t, disabledRegions)
}

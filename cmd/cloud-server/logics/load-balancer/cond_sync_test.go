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
 * to the current version of the project delivered to anyone in the future.
 */

package lblogic

import (
	"testing"

	corelb "hcm/pkg/api/core/cloud/load-balancer"

	"github.com/stretchr/testify/require"
)

func TestDiffLoadBalancerBrief(t *testing.T) {
	region := "ap-guangzhou"

	tests := []struct {
		name       string
		cloudLBs   []corelb.LoadBalancerBrief
		dbLBs      []corelb.LoadBalancerBrief
		wantCreate []string
		wantUpdate []string
		wantDelete []string
	}{
		{
			name:       "empty input",
			cloudLBs:   nil,
			dbLBs:      nil,
			wantCreate: nil,
			wantUpdate: nil,
			wantDelete: nil,
		},
		{
			name: "cloud only, all create",
			cloudLBs: []corelb.LoadBalancerBrief{
				{CloudID: "lb-1"}, {CloudID: "lb-2"},
			},
			dbLBs:      nil,
			wantCreate: []string{"lb-1", "lb-2"},
			wantUpdate: nil,
			wantDelete: nil,
		},
		{
			name:     "db only, all delete",
			cloudLBs: nil,
			dbLBs: []corelb.LoadBalancerBrief{
				{CloudID: "lb-1"}, {CloudID: "lb-2"},
			},
			wantCreate: nil,
			wantUpdate: nil,
			wantDelete: []string{"lb-1", "lb-2"},
		},
		{
			name: "both exist, all update",
			cloudLBs: []corelb.LoadBalancerBrief{
				{CloudID: "lb-1"}, {CloudID: "lb-2"},
			},
			dbLBs: []corelb.LoadBalancerBrief{
				{CloudID: "lb-1"}, {CloudID: "lb-2"},
			},
			wantCreate: nil,
			wantUpdate: []string{"lb-1", "lb-2"},
			wantDelete: nil,
		},
		{
			name: "mixed create update delete",
			cloudLBs: []corelb.LoadBalancerBrief{
				{CloudID: "lb-create"}, {CloudID: "lb-update"},
			},
			dbLBs: []corelb.LoadBalancerBrief{
				{CloudID: "lb-update"}, {CloudID: "lb-delete"},
			},
			wantCreate: []string{"lb-create"},
			wantUpdate: []string{"lb-update"},
			wantDelete: []string{"lb-delete"},
		},
		{
			name: "cloud duplicate cloud_id deduplicated",
			cloudLBs: []corelb.LoadBalancerBrief{
				{CloudID: "lb-1"}, {CloudID: "lb-1"}, {CloudID: "lb-2"},
			},
			dbLBs:      nil,
			wantCreate: []string{"lb-1", "lb-2"},
			wantUpdate: nil,
			wantDelete: nil,
		},
		{
			name: "cloud duplicate with db match still update once",
			cloudLBs: []corelb.LoadBalancerBrief{
				{CloudID: "lb-1"}, {CloudID: "lb-1"},
			},
			dbLBs: []corelb.LoadBalancerBrief{
				{CloudID: "lb-1"},
			},
			wantCreate: nil,
			wantUpdate: []string{"lb-1"},
			wantDelete: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := DiffLoadBalancerBrief(region, tt.cloudLBs, tt.dbLBs)

			require.Equal(t, region, diff.Region)
			require.Equal(t, tt.wantCreate, extractCloudIDs(diff.Create))
			require.Equal(t, tt.wantUpdate, extractCloudIDs(diff.Update))
			require.Equal(t, tt.wantDelete, extractCloudIDs(diff.Delete))
		})
	}
}

func extractCloudIDs(briefs []corelb.LoadBalancerBrief) []string {
	if len(briefs) == 0 {
		return nil
	}
	ids := make([]string, 0, len(briefs))
	for _, b := range briefs {
		ids = append(ids, b.CloudID)
	}
	return ids
}

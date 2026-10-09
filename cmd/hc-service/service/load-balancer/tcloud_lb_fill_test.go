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

package loadbalancer

import (
	"testing"

	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/criteria/constant"

	"github.com/stretchr/testify/assert"
)

func TestBuildTCloudLBBizUpdates(t *testing.T) {
	const reqBiz int64 = 100
	cases := []struct {
		name         string
		lbs          []corelb.BaseLoadBalancer
		wantUpdates  []*tcloudLBUpdate
		wantOtherBiz map[string]int64
	}{
		{
			name: "all unassigned",
			lbs: []corelb.BaseLoadBalancer{
				{ID: "id-1", CloudID: "lb-1", BkBizID: constant.UnassignedBiz},
				{ID: "id-2", CloudID: "lb-2", BkBizID: constant.UnassignedBiz},
			},
			wantUpdates:  []*tcloudLBUpdate{{ID: "id-1", BkBizID: reqBiz}, {ID: "id-2", BkBizID: reqBiz}},
			wantOtherBiz: map[string]int64{},
		},
		{
			name: "part assigned to other biz",
			lbs: []corelb.BaseLoadBalancer{
				{ID: "id-1", CloudID: "lb-1", BkBizID: constant.UnassignedBiz},
				{ID: "id-2", CloudID: "lb-2", BkBizID: 200},
			},
			wantUpdates:  []*tcloudLBUpdate{{ID: "id-1", BkBizID: reqBiz}},
			wantOtherBiz: map[string]int64{"lb-2": 200},
		},
		{
			name: "all already request biz",
			lbs: []corelb.BaseLoadBalancer{
				{ID: "id-1", CloudID: "lb-1", BkBizID: reqBiz},
				{ID: "id-2", CloudID: "lb-2", BkBizID: reqBiz},
			},
			wantUpdates:  []*tcloudLBUpdate{},
			wantOtherBiz: map[string]int64{},
		},
		{
			name:         "no records",
			wantUpdates:  []*tcloudLBUpdate{},
			wantOtherBiz: map[string]int64{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			updates, otherBiz := buildTCloudLBBizUpdates(c.lbs, reqBiz)
			assert.Equal(t, c.wantUpdates, updates)
			assert.Equal(t, c.wantOtherBiz, otherBiz)
		})
	}
}

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

package tcloud

import (
	"testing"

	typeslb "hcm/pkg/adaptor/types/load-balancer"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	cvt "hcm/pkg/tools/converter"

	"github.com/stretchr/testify/assert"
	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

func newTestTCloudClb(backupZones ...*tclb.ZoneInfo) typeslb.TCloudClb {
	return typeslb.TCloudClb{LoadBalancer: &tclb.LoadBalancer{
		LoadBalancerId:   cvt.ValToPtr("lb-1"),
		LoadBalancerType: cvt.ValToPtr(string(typeslb.OpenLoadBalancerType)),
		MasterZone:       &tclb.ZoneInfo{Zone: cvt.ValToPtr("ap-guangzhou-3")},
		BackupZoneSet:    backupZones,
		TargetRegionInfo: &tclb.TargetRegionInfo{},
	}}
}

func TestGetTCloudBackupZones(t *testing.T) {
	cases := []struct {
		name  string
		zones []*tclb.ZoneInfo
		want  []string
	}{
		{
			name: "cloud returns null",
			want: nil,
		},
		{
			name:  "skip nil and empty zone",
			zones: []*tclb.ZoneInfo{nil, {Zone: nil}, {Zone: cvt.ValToPtr("")}},
			want:  nil,
		},
		{
			name:  "keep cloud order",
			zones: []*tclb.ZoneInfo{{Zone: cvt.ValToPtr("ap-guangzhou-4")}, {Zone: cvt.ValToPtr("ap-guangzhou-6")}},
			want:  []string{"ap-guangzhou-4", "ap-guangzhou-6"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, getTCloudBackupZones(newTestTCloudClb(c.zones...)))
		})
	}
}

func TestConvCloudToDBBackupZones(t *testing.T) {
	cloud := newTestTCloudClb(&tclb.ZoneInfo{Zone: cvt.ValToPtr("ap-guangzhou-4")})
	want := []string{"ap-guangzhou-4"}

	created := convCloudToDBCreate(cloud, "account", "ap-guangzhou", nil, nil)
	assert.Equal(t, want, created.BackupZones)
	assert.Equal(t, []string{"ap-guangzhou-3"}, created.Zones)

	updated := convCloudToDBUpdate("id-1", cloud, nil, nil, "ap-guangzhou")
	assert.Equal(t, want, updated.BackupZones)

	// 云上未返回备可用区时更新项为空，data-service 不覆盖本地已有值
	updated = convCloudToDBUpdate("id-1", newTestTCloudClb(), nil, nil, "ap-guangzhou")
	assert.Nil(t, updated.BackupZones)
}

func TestIsLBChangeBackupZones(t *testing.T) {
	cloud := newTestTCloudClb(&tclb.ZoneInfo{Zone: cvt.ValToPtr("ap-guangzhou-4")})
	db := corelb.TCloudLoadBalancer{BaseLoadBalancer: corelb.BaseLoadBalancer{
		IPVersion: cloud.GetIPVersion(),
		Status:    "0",
	}}

	assert.True(t, isLBChange(cloud, db), "missing local backup zones should trigger update")

	db.BackupZones = []string{"ap-guangzhou-6"}
	assert.True(t, isLBChange(cloud, db), "different backup zones should trigger update")
}

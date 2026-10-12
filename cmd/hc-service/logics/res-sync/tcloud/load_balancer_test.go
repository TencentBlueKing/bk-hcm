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

// newTestLBPair returns a cloud load balancer and a db load balancer which have no difference.
// A null cluster tag from cloud is stored as an empty string, so the db side holds an empty string pointer.
func newTestLBPair() (typeslb.TCloudClb, corelb.TCloudLoadBalancer) {
	cloud := typeslb.TCloudClb{LoadBalancer: &tclb.LoadBalancer{
		TargetRegionInfo: &tclb.TargetRegionInfo{},
	}}
	db := corelb.TCloudLoadBalancer{}
	db.Extension = &corelb.TCloudClbExtension{ClusterTag: cvt.ValToPtr("")}

	return cloud, db
}

func TestIsLBExclusiveChange(t *testing.T) {
	tests := []struct {
		name   string
		modify func(cloud *typeslb.TCloudClb, db *corelb.TCloudLoadBalancer)
		want   bool
	}{
		{name: "no difference", modify: func(_ *typeslb.TCloudClb, _ *corelb.TCloudLoadBalancer) {}, want: false},
		{
			name: "exclusive equal",
			modify: func(cloud *typeslb.TCloudClb, db *corelb.TCloudLoadBalancer) {
				cloud.Exclusive = cvt.ValToPtr(uint64(1))
				db.Extension.Exclusive = cvt.ValToPtr(uint64(1))
			},
			want: false,
		},
		{
			name: "exclusive changed",
			modify: func(cloud *typeslb.TCloudClb, db *corelb.TCloudLoadBalancer) {
				cloud.Exclusive = cvt.ValToPtr(uint64(1))
				db.Extension.Exclusive = cvt.ValToPtr(uint64(0))
			},
			want: true,
		},
		{
			name: "exclusive changed from null to value",
			modify: func(cloud *typeslb.TCloudClb, _ *corelb.TCloudLoadBalancer) {
				cloud.Exclusive = cvt.ValToPtr(uint64(0))
			},
			want: true,
		},
		{
			name: "cluster ids changed",
			modify: func(cloud *typeslb.TCloudClb, db *corelb.TCloudLoadBalancer) {
				cloud.ClusterIds = []*string{cvt.ValToPtr("tgw-a")}
				db.Extension.ClusterIds = []*string{cvt.ValToPtr("tgw-b")}
			},
			want: true,
		},
		{
			name: "cluster tag changed",
			modify: func(cloud *typeslb.TCloudClb, db *corelb.TCloudLoadBalancer) {
				cloud.ClusterTag = cvt.ValToPtr("tag-a")
				db.Extension.ClusterTag = cvt.ValToPtr("tag-b")
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloud, db := newTestLBPair()
			tt.modify(&cloud, &db)
			assert.Equal(t, tt.want, isLBExclusiveChange(cloud, db))
		})
	}
}

func TestIsLBNetworkAttrChange(t *testing.T) {
	newNetworkPair := func() (typeslb.TCloudClb, corelb.TCloudLoadBalancer) {
		cloud, db := newTestLBPair()
		cloud.NetworkAttributes = &tclb.InternetAccessible{
			InternetMaxBandwidthOut: cvt.ValToPtr(int64(10)),
			InternetChargeType:      cvt.ValToPtr("BANDWIDTH_POSTPAID_BY_HOUR"),
			BandwidthpkgSubType:     cvt.ValToPtr("BGP"),
		}
		db.BandWidth = 10
		db.Extension.InternetMaxBandwidthOut = cvt.ValToPtr(int64(10))
		db.Extension.InternetChargeType = cvt.ValToPtr("BANDWIDTH_POSTPAID_BY_HOUR")
		db.Extension.BandwidthpkgSubType = cvt.ValToPtr("BGP")

		return cloud, db
	}

	tests := []struct {
		name   string
		modify func(cloud *typeslb.TCloudClb, db *corelb.TCloudLoadBalancer)
		want   bool
	}{
		{name: "equal", modify: func(_ *typeslb.TCloudClb, _ *corelb.TCloudLoadBalancer) {}, want: false},
		{
			name: "cloud network attributes null",
			modify: func(cloud *typeslb.TCloudClb, db *corelb.TCloudLoadBalancer) {
				cloud.NetworkAttributes = nil
				db.BandWidth = 99
			},
			want: false,
		},
		{
			name: "extension bandwidth changed",
			modify: func(cloud *typeslb.TCloudClb, _ *corelb.TCloudLoadBalancer) {
				cloud.NetworkAttributes.InternetMaxBandwidthOut = cvt.ValToPtr(int64(20))
			},
			want: true,
		},
		{
			name: "db bandwidth changed",
			modify: func(_ *typeslb.TCloudClb, db *corelb.TCloudLoadBalancer) {
				db.BandWidth = 20
			},
			want: true,
		},
		{
			name: "charge type changed",
			modify: func(cloud *typeslb.TCloudClb, _ *corelb.TCloudLoadBalancer) {
				cloud.NetworkAttributes.InternetChargeType = cvt.ValToPtr("TRAFFIC_POSTPAID_BY_HOUR")
			},
			want: true,
		},
		{
			name: "bandwidth package sub type changed",
			modify: func(cloud *typeslb.TCloudClb, _ *corelb.TCloudLoadBalancer) {
				cloud.NetworkAttributes.BandwidthpkgSubType = cvt.ValToPtr("SINGLEISP")
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloud, db := newNetworkPair()
			tt.modify(&cloud, &db)
			assert.Equal(t, tt.want, isLBNetworkAttrChange(cloud, db))
		})
	}
}

// TestIsLBExtensionChange checks the extracted helpers are still wired into isLBExtensionChange.
func TestIsLBExtensionChange(t *testing.T) {
	t.Run("extension null", func(t *testing.T) {
		cloud, db := newTestLBPair()
		db.Extension = nil
		assert.True(t, isLBExtensionChange(cloud, db))
	})

	t.Run("no difference", func(t *testing.T) {
		cloud, db := newTestLBPair()
		assert.False(t, isLBExtensionChange(cloud, db))
	})

	t.Run("exclusive changed", func(t *testing.T) {
		cloud, db := newTestLBPair()
		cloud.Exclusive = cvt.ValToPtr(uint64(1))
		assert.True(t, isLBExtensionChange(cloud, db))
	})

	t.Run("network attributes changed", func(t *testing.T) {
		cloud, db := newTestLBPair()
		cloud.NetworkAttributes = &tclb.InternetAccessible{InternetMaxBandwidthOut: cvt.ValToPtr(int64(5))}
		assert.True(t, isLBExtensionChange(cloud, db))
	})
}

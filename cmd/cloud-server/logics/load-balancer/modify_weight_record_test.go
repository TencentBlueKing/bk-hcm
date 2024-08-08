/*
 *
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

package lblogic

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModifyWeightRecord_validateVIPsAndVPorts(t *testing.T) {
	cases := []struct {
		input     *ModifyWeightRecord
		expectErr bool
	}{
		{
			input: &ModifyWeightRecord{
				VIP:         "192.168.1.3",
				VPorts:      []int{80},
				HaveEndPort: false,
			},
			expectErr: false,
		},
		{
			input: &ModifyWeightRecord{
				VIP:         "192.168.1.3",
				HaveEndPort: false,
			},
			expectErr: true,
		},
		{
			input: &ModifyWeightRecord{
				VPorts:      []int{80},
				HaveEndPort: false,
			},
			expectErr: true,
		},
		{
			input: &ModifyWeightRecord{
				VIP:         "192.168.1.3",
				VPorts:      []int{80, 100},
				HaveEndPort: false,
			},
			expectErr: true,
		},
		{
			input: &ModifyWeightRecord{
				VIP:         "192.168.1.3",
				VPorts:      []int{80},
				HaveEndPort: true,
			},
			expectErr: true,
		},
		{
			input: &ModifyWeightRecord{
				VIP:         "192.168.1.3",
				VPorts:      []int{80, 1000},
				HaveEndPort: true,
			},
			expectErr: false,
		},
		{
			input: &ModifyWeightRecord{
				VIP:         "192.168.1.3",
				VPorts:      []int{80, 1000000},
				HaveEndPort: true,
			},
			expectErr: true,
		},
	}

	for i, c := range cases {
		err := c.input.validateVIPsAndVPorts()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}

}

func TestModifyWeightRecord_validateProtocol(t *testing.T) {
	cases := []struct {
		input     *ModifyWeightRecord
		expectErr bool
	}{
		{
			input: &ModifyWeightRecord{
				Protocol: "TCP",
			},
			expectErr: false,
		},
		{
			input: &ModifyWeightRecord{
				Protocol: "UDP",
			},
			expectErr: false,
		},
		{
			input: &ModifyWeightRecord{
				Protocol: "HTTPS",
			},
			expectErr: false,
		},
		{
			input: &ModifyWeightRecord{
				Protocol: "HTTPS",
			},
			expectErr: false,
		},
		{
			input: &ModifyWeightRecord{
				Protocol: "grpc",
			},
			expectErr: true,
		},
		{
			input: &ModifyWeightRecord{
				Protocol: "HTTP",
			},
			expectErr: false,
		},
	}

	for i, c := range cases {
		err := c.input.validateProtocol()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}

}

func TestModifyWeightRecord_validateWeight(t *testing.T) {
	cases := []struct {
		input     *ModifyWeightRecord
		expectErr bool
	}{
		{
			input: &ModifyWeightRecord{
				OldWeight: []int{1},
				Weight:    []int{1},
			},
			expectErr: false,
		},
		{
			input: &ModifyWeightRecord{
				OldWeight: []int{1},
				Weight:    []int{101},
			},
			expectErr: true,
		},
		{
			input: &ModifyWeightRecord{
				OldWeight: []int{1},
				Weight:    []int{-1},
			},
			expectErr: true,
		},
		{
			input: &ModifyWeightRecord{
				OldWeight: []int{1, 2},
				Weight:    []int{2},
			},
			expectErr: true,
		},
	}

	for i, c := range cases {
		err := c.input.validateWeight()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}
}

func TestModifyWeightRecord_validateListenerName(t *testing.T) {
	cases := []struct {
		input     *ModifyWeightRecord
		expectErr bool
	}{
		{
			input: &ModifyWeightRecord{
				ListenerName: "test-rule",
			},
			expectErr: false,
		},
		{
			input: &ModifyWeightRecord{
				ListenerName: strings.Repeat("a", 256),
			},
			expectErr: true,
		},
		{
			input:     &ModifyWeightRecord{},
			expectErr: true,
		},
	}

	for i, c := range cases {
		err := c.input.validateListenerName()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}
}

func TestModifyWeightRecord_validateIpDomainType(t *testing.T) {
	cases := []struct {
		input     *ModifyWeightRecord
		expectErr bool
	}{
		{
			input: &ModifyWeightRecord{
				IPDomainType: ipDomainTypeIPv4,
				VIP:          "192.168.1.1",
			},
			expectErr: false,
		},
		{
			input: &ModifyWeightRecord{
				IPDomainType: ipDomainTypeIPv6,
				VIP:          "2408:8722:840:f8::83",
			},
			expectErr: false,
		},
		{
			input: &ModifyWeightRecord{
				IPDomainType: ipDomainTypeDomain,
				VIP:          "example.com",
			},
			expectErr: false,
		},
		// bad cases
		{
			input: &ModifyWeightRecord{
				IPDomainType: ipDomainTypeIPv4,
				VIP:          "2408:8722:840:f8::83",
			},
			expectErr: true,
		},
		{
			input: &ModifyWeightRecord{
				IPDomainType: ipDomainTypeIPv6,
				VIP:          "192.168.1.1",
			},
			expectErr: true,
		},
	}

	for i, c := range cases {
		err := c.input.validateIpDomainType()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}
}

func TestModifyWeightRecord_validateRS(t *testing.T) {
	cases := []struct {
		input     *ModifyWeightRecord
		expectErr bool
	}{
		{
			input: &ModifyWeightRecord{
				RSIPs:     []string{"192.168.1.3"},
				RSPorts:   []int{80},
				Weight:    []int{1},
				OldWeight: []int{1},
			},
			expectErr: false,
		},
		{
			input: &ModifyWeightRecord{
				RSIPs:     []string{"192.168.1.3"},
				RSPorts:   []int{80, 8081},
				Weight:    []int{1, 2},
				OldWeight: []int{1, 2},
			},
			expectErr: true,
		},
		{
			input: &ModifyWeightRecord{
				RSIPs:     []string{"192.168.1.3"},
				RSPorts:   []int{80, 8081},
				Weight:    []int{1},
				OldWeight: []int{1, 2},
			},
			expectErr: true,
		},
		{
			input: &ModifyWeightRecord{
				RSIPs:     []string{"192.168.1.3", "192.168.1.4"},
				RSPorts:   []int{80},
				Weight:    []int{1},
				OldWeight: []int{1},
			},
			expectErr: false,
		},
		{
			input: &ModifyWeightRecord{
				RSIPs:     []string{"192.168.1.3", "192.168.1.4"},
				RSPorts:   []int{80, 8081},
				Weight:    []int{1, 2},
				OldWeight: []int{1, 2},
			},
			expectErr: false,
		},
		{
			input: &ModifyWeightRecord{
				RSIPs:     []string{"192.168.1.3", "192.168.1.4"},
				RSPorts:   []int{80},
				Weight:    []int{1, 2},
				OldWeight: []int{1, 2},
			},
			expectErr: false,
		},
		{
			input: &ModifyWeightRecord{
				RSIPs:     []string{"192.168.1.3", "192.168.1.4"},
				RSPorts:   []int{80},
				Weight:    []int{1, 2, 3},
				OldWeight: []int{1, 2, 3},
			},
			expectErr: true,
		},
		{
			input:     &ModifyWeightRecord{},
			expectErr: true,
		},
		{
			input: &ModifyWeightRecord{
				RSIPs: []string{"192.168.1.3", "192.168.1.4"},
			},
			expectErr: true,
		},
		{
			input: &ModifyWeightRecord{
				RSPorts: []int{80},
			},
			expectErr: true,
		},
		{
			input: &ModifyWeightRecord{
				Weight:    []int{1},
				OldWeight: []int{1},
			},
			expectErr: true,
		},
		{
			input: &ModifyWeightRecord{
				RSIPs:     []string{"192.168.1.3", "192.168.1.4"},
				RSPorts:   []int{65535, 65536},
				Weight:    []int{1, 2},
				OldWeight: []int{1, 2},
			},
			expectErr: true,
		},
		//EndPort cases
		{
			input: &ModifyWeightRecord{
				HaveEndPort: true,
				RSIPs:       []string{"192.168.1.3"},
				RSPorts:     []int{80},
				Weight:      []int{1},
				OldWeight:   []int{1},
			},
			expectErr: true,
		},
		{ // 端口段 rsIP weight 不唯一
			input: &ModifyWeightRecord{
				HaveEndPort: true,
				RSIPs:       []string{"192.168.1.3", "192.168.1.4"},
				RSPorts:     []int{80, 8081},
				VPorts:      []int{80, 8081},
				Weight:      []int{1, 2},
				OldWeight:   []int{1, 2},
			},
			expectErr: true,
		},
		{
			input: &ModifyWeightRecord{
				HaveEndPort: true,
				RSIPs:       []string{"192.168.1.3"},
				RSPorts:     []int{80, 8081},
				VPorts:      []int{80, 8080},
				Weight:      []int{1},
				OldWeight:   []int{1},
			},
			expectErr: true,
		},
		{ // 溢出
			input: &ModifyWeightRecord{
				HaveEndPort: true,
				RSIPs:       []string{"192.168.1.3"},
				RSPorts:     []int{65535, 65536},
				VPorts:      []int{80, 81},
				Weight:      []int{1},
			},
			expectErr: true,
		},
		{
			input: &ModifyWeightRecord{
				HaveEndPort: true,
				RSIPs:       []string{"192.168.1.3"},
				RSPorts:     []int{80, 8081},
				VPorts:      []int{80, 8081},
				Weight:      []int{1},
				OldWeight:   []int{1},
			},
			expectErr: false,
		},
	}

	for i, c := range cases {
		err := c.input.validateRS()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}
}

func TestModifyWeightRecord_validateRSInfos(t *testing.T) {
	cases := []struct {
		input     *ModifyWeightRecord
		expectErr bool
	}{
		{
			input: &ModifyWeightRecord{
				RSInfos: []*RSUpdateInfo{
					{
						IP:   "192.168.1.3",
						Port: 8080,
					},
				},
			},
			expectErr: false,
		},
		{
			input: &ModifyWeightRecord{
				RSInfos: []*RSUpdateInfo{
					{
						IP:   "192.168.1.3",
						Port: 8080,
					},
					{
						IP:   "192.168.1.3",
						Port: 8080,
					},
				},
			},
			expectErr: true,
		},
	}

	for i, c := range cases {
		err := c.input.validateRSInfoDuplicate()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}
}

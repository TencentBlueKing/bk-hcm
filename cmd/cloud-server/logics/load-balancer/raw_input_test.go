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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBindRSRawInput_SplitRecord(t *testing.T) {
	cases := []struct {
		input     *BindRSRawInput
		expect    int
		expectErr bool
	}{
		{
			input: &BindRSRawInput{
				VIPs:        []string{"192.168.1.3", "192.168.1.4"},
				VPorts:      []int{80, 81},
				haveEndPort: true,
			},
			expect:    2,
			expectErr: false,
		},
		{ // bad case
			input: &BindRSRawInput{
				VIPs:        []string{"192.168.1.3", "192.168.1.4"},
				VPorts:      []int{80, 81, 82},
				haveEndPort: false,
			},
			expect:    6,
			expectErr: false,
		},
		{
			input: &BindRSRawInput{
				VIPs:        []string{"192.168.1.3", "192.168.1.5", "192.168.1.4"},
				VPorts:      []int{80, 81, 82},
				haveEndPort: false,
			},
			expect:    9,
			expectErr: false,
		},
		{
			input: &BindRSRawInput{
				VIPs:   []string{"192.168.1.3"},
				VPorts: []int{80, 81, 82},
			},
			expect:    3,
			expectErr: false,
		},
		{
			input: &BindRSRawInput{
				VIPs:   []string{"192.168.1.3", "192.168.1.4"},
				VPorts: []int{80},
			},
			expect:    2,
			expectErr: false,
		},
		{
			input: &BindRSRawInput{
				VIPs:   []string{"192.168.1.3", "192.168.1.4"},
				VPorts: []int{},
			},
			expect:    0,
			expectErr: true,
		},
	}

	for i, c := range cases {
		records, err := c.input.SplitRecord()
		assert.Equalf(t, c.expect, len(records), "case %d failed", i)
		if c.expectErr {
			assert.Errorf(t, err, "case %d failed", i)
		}
	}
}

func TestModifyWeightRawInput_SplitRecord(t *testing.T) {
	cases := []struct {
		input     *ModifyWeightRawInput
		expect    int
		expectErr bool
	}{
		{
			input: &ModifyWeightRawInput{
				VIPs:        []string{"192.168.1.3", "192.168.1.4"},
				VPorts:      []int{80, 81},
				haveEndPort: true,
			},
			expect:    2,
			expectErr: false,
		},
		{ // bad case
			input: &ModifyWeightRawInput{
				VIPs:        []string{"192.168.1.3", "192.168.1.4"},
				VPorts:      []int{80, 81, 82},
				haveEndPort: false,
			},
			expect:    6,
			expectErr: false,
		},
		{
			input: &ModifyWeightRawInput{
				VIPs:        []string{"192.168.1.3", "192.168.1.5", "192.168.1.4"},
				VPorts:      []int{80, 81, 82},
				haveEndPort: false,
			},
			expect:    9,
			expectErr: false,
		},
		{
			input: &ModifyWeightRawInput{
				VIPs:   []string{"192.168.1.3"},
				VPorts: []int{80, 81, 82},
			},
			expect:    3,
			expectErr: false,
		},
		{
			input: &ModifyWeightRawInput{
				VIPs:   []string{"192.168.1.3", "192.168.1.4"},
				VPorts: []int{80},
			},
			expect:    2,
			expectErr: false,
		},
		{
			input: &ModifyWeightRawInput{
				VIPs:   []string{"192.168.1.3", "192.168.1.4"},
				VPorts: []int{},
			},
			expect:    0,
			expectErr: true,
		},
	}

	for i, c := range cases {
		records, err := c.input.SplitRecord()
		assert.Equalf(t, c.expect, len(records), "case %d failed", i)
		if c.expectErr {
			assert.Errorf(t, err, "case %d failed", i)
		}
	}
}

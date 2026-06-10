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

package cvm

import (
	"testing"

	"hcm/pkg/criteria/enumor"
)

func TestIsGPUMachineAzureAndEmpty(t *testing.T) {
	got := isGPUMachine(enumor.Azure, "Standard_NC6", nil)
	if got {
		t.Errorf("azure machine is_gpu should be false, got: %v", got)
	}

	got = isGPUMachine(enumor.TCloud, "", map[string]struct{}{"GN10X": {}})
	if got {
		t.Errorf("empty machine type is_gpu should be false, got: %v", got)
	}
}

func TestMatchGPUMachineType(t *testing.T) {
	tcloudPrefixes := map[string]struct{}{
		"GN10X": {},
	}
	huaweiPrefixes := map[string]struct{}{
		"pi2": {},
	}
	awsTypes := map[string]struct{}{
		"g4dn.xlarge": {},
	}
	gcpTypes := map[string]struct{}{
		"n1-standard-8": {},
	}

	tests := []struct {
		name         string
		vendor       enumor.Vendor
		machineType  string
		machineTypes map[string]struct{}
		want         bool
	}{
		{
			name:         "tcloud prefix hit",
			vendor:       enumor.TCloud,
			machineType:  "GN10X.2XLARGE40",
			machineTypes: tcloudPrefixes,
			want:         true,
		},
		{
			name:         "tcloud prefix miss",
			vendor:       enumor.TCloud,
			machineType:  "S5.LARGE8",
			machineTypes: tcloudPrefixes,
			want:         false,
		},
		{
			name:         "huawei prefix hit",
			vendor:       enumor.HuaWei,
			machineType:  "pi2.2xlarge.4",
			machineTypes: huaweiPrefixes,
			want:         true,
		},
		{
			name:         "aws exact hit",
			vendor:       enumor.Aws,
			machineType:  "g4dn.xlarge",
			machineTypes: awsTypes,
			want:         true,
		},
		{
			name:         "aws exact miss by prefix only",
			vendor:       enumor.Aws,
			machineType:  "g4dn.2xlarge",
			machineTypes: awsTypes,
			want:         false,
		},
		{
			name:         "gcp exact hit",
			vendor:       enumor.Gcp,
			machineType:  "n1-standard-8",
			machineTypes: gcpTypes,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchGPUMachineType(tt.vendor, tt.machineType, tt.machineTypes)
			if got != tt.want {
				t.Errorf("matchGPUMachineType() = %v, want %v", got, tt.want)
			}
		})
	}
}

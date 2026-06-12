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

func buildMachineTypeSet(machineTypes []string) map[string]struct{} {
	set := make(map[string]struct{}, len(machineTypes))
	for _, machineType := range machineTypes {
		set[machineType] = struct{}{}
	}
	return set
}

func TestIsGPUMachine(t *testing.T) {
	// Keep this in sync with current global_config snapshot.
	awsMachineTypes := []string{"g4dn.xlarge", "g4dn.2xlarge", "g4dn.4xlarge", "g4dn.8xlarge", "g4dn.16xlarge",
		"g4ad.xlarge", "g7e.48xlarge", "g5.xlarge", "g5.2xlarge", "g5.4xlarge", "g5.8xlarge", "g5.12xlarge",
		"g5.16xlarge", "g5.24xlarge", "g5.48xlarge", "g5g.xlarge", "g6.xlarge", "g6.2xlarge", "g6.4xlarge",
		"g6.8xlarge", "g6.12xlarge", "g6.16xlarge", "g6.24xlarge", "g6.48xlarge", "g6e.xlarge", "g6e.2xlarge",
		"g6e.4xlarge", "g6e.8xlarge", "g6e.12xlarge", "g6e.16xlarge", "g6e.24xlarge", "g6e.48xlarge", "p3.2xlarge",
		"p3.8xlarge", "p3.16xlarge", "p3dn.24xlarge", "p4d.24xlarge", "p4de.24xlarge", "p5.48xlarge", "p5.2xlarge",
		"p5.4xlarge", "p5en.48xlarge", "p5e.48xlarge", "p6-b200.48xlarge", "p6-b300.48xlarge", "ml.g4dn.xlarge",
		"ml.g4dn.2xlarge", "ml.g4dn.4xlarge", "ml.g4dn.8xlarge", "ml.g4dn.12xlarge", "ml.g4dn.16xlarge", "ml.g5.xlarge",
		"ml.g5.2xlarge", "ml.g5.4xlarge", "ml.g5.8xlarge", "ml.g5.12xlarge", "ml.g5.16xlarge", "ml.g5.24xlarge",
		"ml.g5.48xlarge", "ml.g6.xlarge", "ml.g6.2xlarge", "ml.g6.4xlarge", "ml.g6e.xlarge", "ml.g6e.2xlarge",
		"ml.g6e.4xlarge", "ml.g6e.8xlarge", "ml.g6e.48xlarge", "ml.p3.2xlarge", "ml.p3.8xlarge", "ml.p3.16xlarge",
		"ml.p4d.24xlarge", "ml.p4de.24xlarge", "ml.p5.48xlarge", "ml.p5.4xlarge", "ml.p5e.48xlarge", "ml.p5en.48xlarge",
		"p6-b300",
	}
	awsTypes := buildMachineTypeSet(awsMachineTypes)
	tcloudPrefixes := buildMachineTypeSet([]string{"g4dn", "GN7", "GN8", "GN10X", "GN10Xp", "GN7vw", "PNV5b"})
	azurePrefixes := buildMachineTypeSet([]string{"Standard_NC", "Standard_ND", "Standard_NV", "Standard_NG"})
	gcpPrefixes := buildMachineTypeSet([]string{"a2-", "a3-", "a4-", "a4x-", "g2-", "g4-"})
	huaweiPrefixes := buildMachineTypeSet([]string{
		"p1", "p2v", "p2vs", "p2s", "pi1", "pi2", "g1", "g3", "g5", "g6", "g6v"})

	tests := []struct {
		name         string
		vendor       enumor.Vendor
		machineType  string
		machineTypes map[string]struct{}
		extension    string
		want         bool
	}{
		{name: "tcloud prefix hit", vendor: enumor.TCloud, machineType: "GN10X.2XLARGE40",
			machineTypes: tcloudPrefixes, want: true},
		{name: "tcloud configured g4dn prefix hit", vendor: enumor.TCloud, machineType: "g4dn.8xlarge",
			machineTypes: tcloudPrefixes, want: true},
		{name: "tcloud prefix miss", vendor: enumor.TCloud, machineType: "S5.LARGE8", machineTypes: tcloudPrefixes,
			want: false},
		{name: "huawei prefix hit", vendor: enumor.HuaWei, machineType: "pi2.2xlarge.4", machineTypes: huaweiPrefixes,
			want: true},
		{name: "huawei g-series prefix hit", vendor: enumor.HuaWei, machineType: "g6v.4xlarge.8",
			machineTypes: huaweiPrefixes, want: true,
		},
		{name: "huawei prefix miss", vendor: enumor.HuaWei, machineType: "s6.large.2", machineTypes: huaweiPrefixes,
			want: false,
		},
		{name: "aws exact hit", vendor: enumor.Aws, machineType: "g4dn.xlarge", machineTypes: awsTypes,
			want: true,
		},
		{name: "aws exact miss by prefix only", vendor: enumor.Aws, machineType: "g4dn", machineTypes: awsTypes,
			want: false,
		},
		{name: "aws configured ml type exact hit", vendor: enumor.Aws, machineType: "ml.g6e.48xlarge",
			machineTypes: awsTypes,
			want:         true,
		},
		{
			name: "gcp prefix hit", vendor: enumor.Gcp, machineType: "a2-highgpu-1g", machineTypes: gcpPrefixes,
			want: true,
		},
		{name: "gcp g2 prefix hit", vendor: enumor.Gcp, machineType: "g2-standard-16", machineTypes: gcpPrefixes,
			want: true,
		},
		{name: "gcp prefix miss", vendor: enumor.Gcp, machineType: "n2-standard-8", machineTypes: gcpPrefixes,
			want: false,
		},
		{name: "gcp n1 with attached gpu by card count", vendor: enumor.Gcp, machineType: "n1-standard-8",
			machineTypes: gcpPrefixes, extension: `{"guest_accelerators":[{"accelerator_count":1}]}`, want: true,
		},
		{name: "gcp prefix hit ignores empty extension", vendor: enumor.Gcp, machineType: "g2-standard-16",
			machineTypes: gcpPrefixes, extension: "", want: true,
		},
		{name: "gcp non gpu with no accelerators", vendor: enumor.Gcp, machineType: "n1-standard-8",
			machineTypes: gcpPrefixes, extension: `{"guest_accelerators":[]}`, want: false,
		},
		{name: "non gcp extension ignored", vendor: enumor.TCloud, machineType: "S5.LARGE8",
			machineTypes: tcloudPrefixes, extension: `{"guest_accelerators":[{"accelerator_count":8}]}`, want: false,
		},
		{name: "azure prefix hit", vendor: enumor.Azure, machineType: "Standard_NC6s_v3", machineTypes: azurePrefixes,
			want: true,
		},
		{name: "azure nv prefix hit", vendor: enumor.Azure, machineType: "Standard_NV36ads_A10_v5",
			machineTypes: azurePrefixes, want: true,
		},
		{name: "azure prefix miss", vendor: enumor.Azure, machineType: "Standard_D2s_v5", machineTypes: azurePrefixes,
			want: false,
		},
		{name: "azure e-series prefix miss", vendor: enumor.Azure, machineType: "Standard_E4s_v5",
			machineTypes: azurePrefixes, want: false,
		},
		{name: "azure d prefix miss", vendor: enumor.Azure, machineType: "Standard_D4s_v5",
			machineTypes: azurePrefixes, want: false,
		},
		{name: "empty machine type", vendor: enumor.TCloud, machineType: "", machineTypes: tcloudPrefixes,
			want: false,
		},
		{name: "unsupported vendor", vendor: enumor.Vendor("unknown"), machineType: "GN10X.2XLARGE40",
			machineTypes: tcloudPrefixes, want: false,
		},
		{name: "unsupported vendor", vendor: enumor.Vendor("unknown"), machineType: "GN10X.2XLARGE40",
			machineTypes: tcloudPrefixes, want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isGPUMachine(tt.vendor, tt.machineType, tt.machineTypes, tt.extension)
			if got != tt.want {
				t.Errorf("isGPUMachine() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("aws full config all hit", func(t *testing.T) {
		for _, machineType := range awsMachineTypes {
			if !isGPUMachine(enumor.Aws, machineType, awsTypes, "") {
				t.Errorf("aws machine type should match: %s", machineType)
			}
		}
	})
}

func TestMatchGPUMachineTypeByPrefix(t *testing.T) {
	prefixes := map[string]struct{}{
		"GN10X": {},
		"pi2":   {},
	}

	tests := []struct {
		name        string
		machineType string
		prefixes    map[string]struct{}
		want        bool
	}{
		{
			name:        "prefix hit",
			machineType: "GN10X.2XLARGE40",
			prefixes:    prefixes,
			want:        true,
		},
		{
			name:        "prefix miss",
			machineType: "S5.LARGE8",
			prefixes:    prefixes,
			want:        false,
		},
		{
			name:        "case insensitive machine type",
			machineType: "gn10x.2xlarge40",
			prefixes:    prefixes,
			want:        true,
		},
		{
			name:        "case insensitive prefix",
			machineType: "PI2.2xlarge.4",
			prefixes:    map[string]struct{}{"pi2": {}},
			want:        true,
		},
		{
			name:        "empty machine type",
			machineType: "",
			prefixes:    prefixes,
			want:        false,
		},
		{
			name:        "empty prefix map",
			machineType: "GN10X.2XLARGE40",
			prefixes:    map[string]struct{}{},
			want:        false,
		},
		{
			name:        "exact prefix match",
			machineType: "GN10X",
			prefixes:    map[string]struct{}{"GN10X": {}},
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchGPUMachineTypeByPrefix(tt.machineType, tt.prefixes)
			if got != tt.want {
				t.Errorf("matchGPUMachineTypeByPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

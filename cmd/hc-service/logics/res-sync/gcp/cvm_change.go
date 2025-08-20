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

package gcp

import (
	"fmt"
	"time"

	"hcm/pkg/adaptor/gcp"
	typescvm "hcm/pkg/adaptor/types/cvm"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/times"
)

// isCvmChange checks if the GCP CVM has changed by comparing its cloud data with the database data.
func isCvmChange(cloud typescvm.GcpCvm, db corecvm.Cvm[corecvm.GcpCvmExtension]) bool {

	if isCvmBaseInfoChange(cloud, db) ||
		isCvmNetWorkInterfaceChange(cloud, db) ||
		isCvmIPChange(cloud, db) ||
		isCvmTimeChange(cloud, db) ||
		isCvmDiskChange(cloud, db) ||
		isCvmReservationAffinityChange(cloud, db) ||
		isCvmAdvanceMacheFeaturesChange(cloud, db) ||
		isCvmExtensionInfoChange(cloud, db) {
		return true
	}

	return false
}

// isCvmExtensionInfoChange checks if the extension information of the GCP CVM has changed.
func isCvmExtensionInfoChange(cloud typescvm.GcpCvm, db corecvm.Cvm[corecvm.GcpCvmExtension]) bool {
	if db.Extension.DeletionProtection != cloud.DeletionProtection {
		return true
	}

	if db.Extension.CpuPlatform != cloud.CpuPlatform {
		return true
	}

	if db.Extension.CanIpForward != cloud.CanIpForward {
		return true
	}

	if db.Extension.SelfLink != cloud.SelfLink {
		return true
	}

	if db.Extension.MinCpuPlatform != cloud.MinCpuPlatform {
		return true
	}

	if db.Extension.StartRestricted != cloud.StartRestricted {
		return true
	}

	if !assert.IsStringSliceEqual(db.Extension.ResourcePolicies, cloud.ResourcePolicies) {
		return true
	}

	if db.Extension.Fingerprint != cloud.Fingerprint {
		return true
	}
	return false
}

// isCvmReservationAffinityChange checks if the reservation affinity of the GCP CVM has changed.
func isCvmReservationAffinityChange(cloud typescvm.GcpCvm, db corecvm.Cvm[corecvm.GcpCvmExtension]) bool {
	if (db.Extension.ReservationAffinity == nil && cloud.ReservationAffinity != nil) ||
		(db.Extension.ReservationAffinity != nil && cloud.ReservationAffinity == nil) {
		return true
	}

	if db.Extension.ReservationAffinity != nil && cloud.ReservationAffinity != nil {
		if db.Extension.ReservationAffinity.ConsumeReservationType != cloud.ReservationAffinity.ConsumeReservationType {
			return true
		}

		if db.Extension.ReservationAffinity.Key != cloud.ReservationAffinity.Key {
			return true
		}

		if !assert.IsStringSliceEqual(db.Extension.ReservationAffinity.Values, cloud.ReservationAffinity.Values) {
			return true
		}
	}

	return false
}

// isCvmIPChange checks if the IP addresses of the GCP CVM have changed.
func isCvmIPChange(cloud typescvm.GcpCvm, db corecvm.Cvm[corecvm.GcpCvmExtension]) bool {
	priIPv4, pubIPv4, priIPv6, pubIPv6 := gcp.GetGcpIPAddresses(cloud.NetworkInterfaces)

	if !assert.IsStringSliceEqual(db.PrivateIPv4Addresses, priIPv4) {
		return true
	}

	if !assert.IsStringSliceEqual(db.PublicIPv4Addresses, pubIPv4) {
		return true
	}

	if !assert.IsStringSliceEqual(db.PrivateIPv6Addresses, priIPv6) {
		return true
	}

	if !assert.IsStringSliceEqual(db.PublicIPv6Addresses, pubIPv6) {
		return true
	}

	return false
}

// isCvmTimeChange checks if the creation and last start timestamps of the GCP CVM have changed.
func isCvmTimeChange(cloud typescvm.GcpCvm, db corecvm.Cvm[corecvm.GcpCvmExtension]) bool {
	createTime, err := times.ParseToStdTime(time.RFC3339Nano, cloud.CreationTimestamp)
	if err != nil {
		logs.Errorf("[%s] conv CreationTimestamp to std time failed, err: %v", enumor.Gcp, err)
		return true
	}

	if db.CloudCreatedTime != createTime {
		return true
	}

	startTime, err := times.ParseToStdTime(time.RFC3339Nano, cloud.LastStartTimestamp)
	if err != nil {
		logs.Errorf("[%s] conv LastStartTimestamp to std time failed, err: %v", enumor.Gcp, err)
		return true
	}
	if db.CloudLaunchedTime != startTime {
		return true
	}

	return false
}

// isCvmNetWorkInterfaceChange checks if the network interfaces of the GCP CVM have changed.
func isCvmNetWorkInterfaceChange(cloud typescvm.GcpCvm, db corecvm.Cvm[corecvm.GcpCvmExtension]) bool {

	vpcSelfLinks := make([]string, 0)
	subnetSelfLinks := make([]string, 0)
	cloudNetWorkInterfaceIDs := make([]string, 0)
	if len(cloud.NetworkInterfaces) > 0 {
		for _, networkInterface := range cloud.NetworkInterfaces {
			if networkInterface != nil {
				cloudNetInterfaceID := fmt.Sprintf("%d", cloud.Id) + "_" + networkInterface.Name
				cloudNetWorkInterfaceIDs = append(cloudNetWorkInterfaceIDs, cloudNetInterfaceID)
				vpcSelfLinks = append(vpcSelfLinks, networkInterface.Network)
				subnetSelfLinks = append(subnetSelfLinks, networkInterface.Subnetwork)
			}
		}
	}

	if len(db.Extension.VpcSelfLinks) == 0 || len(vpcSelfLinks) == 0 ||
		(db.Extension.VpcSelfLinks[0] != vpcSelfLinks[0]) {
		return true
	}

	if len(db.Extension.SubnetSelfLinks) == 0 || len(subnetSelfLinks) == 0 ||
		!assert.IsStringSliceEqual(db.Extension.SubnetSelfLinks, subnetSelfLinks) {
		return true
	}

	if !assert.IsStringSliceEqual(db.Extension.CloudNetworkInterfaceIDs, cloudNetWorkInterfaceIDs) {
		return true
	}

	return false
}

// isCvmBaseInfoChange checks if the base information of the GCP CVM has changed.
func isCvmBaseInfoChange(cloud typescvm.GcpCvm, db corecvm.Cvm[corecvm.GcpCvmExtension]) bool {
	if db.CloudID != fmt.Sprintf("%d", cloud.Id) {
		return true
	}

	if db.Name != cloud.Name {
		return true
	}

	if db.CloudImageID != cloud.SourceMachineImage {
		return true
	}

	if db.Status != cloud.Status {
		return true
	}
	if db.MachineType != gcp.GetMachineType(cloud.MachineType) {
		return true
	}

	if !assert.IsStringMapEqual(db.Extension.Labels, cloud.Labels) {
		return true
	}

	return false
}

// isCvmDiskChange checks if the disks of the GCP CVM have changed.
func isCvmDiskChange(cloud typescvm.GcpCvm, db corecvm.Cvm[corecvm.GcpCvmExtension]) bool {
	for _, dbValue := range db.Extension.Disks {
		isEqual := false
		for _, cloudValue := range cloud.Disks {
			if dbValue.Boot == cloudValue.Boot && dbValue.Index == cloudValue.Index &&
				dbValue.SelfLink == cloudValue.Source && dbValue.DeviceName == cloudValue.DeviceName {
				isEqual = true
				break
			}
		}
		if !isEqual {
			return true
		}
	}

	return false
}

// isCvmAdvanceMacheFeaturesChange checks if the advanced machine features of the GCP CVM have changed.
func isCvmAdvanceMacheFeaturesChange(cloud typescvm.GcpCvm, db corecvm.Cvm[corecvm.GcpCvmExtension]) bool {

	if (db.Extension.AdvancedMachineFeatures != nil && cloud.AdvancedMachineFeatures == nil) ||
		(db.Extension.AdvancedMachineFeatures == nil && cloud.AdvancedMachineFeatures != nil) {
		return true
	}

	if db.Extension.AdvancedMachineFeatures != nil && cloud.AdvancedMachineFeatures != nil {
		if db.Extension.AdvancedMachineFeatures.EnableNestedVirtualization !=
			cloud.AdvancedMachineFeatures.EnableNestedVirtualization {
			return true
		}

		if db.Extension.AdvancedMachineFeatures.EnableUefiNetworking != cloud.AdvancedMachineFeatures.EnableUefiNetworking {
			return true
		}

		if db.Extension.AdvancedMachineFeatures.ThreadsPerCore != cloud.AdvancedMachineFeatures.ThreadsPerCore {
			return true
		}

		if db.Extension.AdvancedMachineFeatures.VisibleCoreCount != cloud.AdvancedMachineFeatures.ThreadsPerCore {
			return true
		}
	}
	return false
}

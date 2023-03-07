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
	"fmt"

	"hcm/cmd/hc-service/logics/sync/disk"
	synceip "hcm/cmd/hc-service/logics/sync/eip"
	syncnetworkinterface "hcm/cmd/hc-service/logics/sync/network-interface"
	"hcm/cmd/hc-service/logics/sync/subnet"
	"hcm/cmd/hc-service/logics/sync/vpc"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/adaptor/gcp"
	typecore "hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	protocvm "hcm/pkg/api/hc-service/cvm"
	protodisk "hcm/pkg/api/hc-service/disk"
	protoeip "hcm/pkg/api/hc-service/eip"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
)

// SyncGcpCvmOption ...
type SyncGcpCvmOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	Zone      string   `json:"zone" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"required"`
}

// SyncGcpCvm sync cvm self
func SyncGcpCvm(kt *kit.Kit, ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client,
	req *SyncGcpCvmOption) (interface{}, error) {

	client, err := ad.Gcp(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	cloudAllIDs := make(map[string]bool)
	for {
		opt := &typecvm.GcpListOption{
			Region: req.Region,
			Zone:   req.Zone,
			Page: &typecore.GcpPage{
				PageToken: nextToken,
				PageSize:  int64(filter.DefaultMaxInLimit),
			},
		}

		if nextToken != "" {
			opt.Page.PageToken = nextToken
		}

		if len(req.CloudIDs) > 0 {
			opt.Page = nil
			opt.CloudIDs = req.CloudIDs
		}

		datas, token, err := client.ListCvm(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor to list gcp cvm failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if len(datas) <= 0 {
			break
		}

		cloudMap := make(map[string]*GcpCvmSync)
		cloudIDs := make([]string, 0, len(datas))
		for _, data := range datas {
			cvmSync := new(GcpCvmSync)
			cvmSync.IsUpdate = false
			cvmSync.Cvm = data
			cloudMap[fmt.Sprintf("%d", data.Id)] = cvmSync
			cloudIDs = append(cloudIDs, fmt.Sprintf("%d", data.Id))
			cloudAllIDs[fmt.Sprintf("%d", data.Id)] = true
		}

		updateIDs, dsMap, err := getGcpCvmDSSync(kt, cloudIDs, req, dataCli)
		if err != nil {
			logs.Errorf("request getGcpCvmDSSync failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if len(updateIDs) > 0 {
			err := syncGcpCvmUpdate(kt, updateIDs, cloudMap, dsMap, dataCli)
			if err != nil {
				logs.Errorf("request syncGcpCvmUpdate failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}

		addIDs := make([]string, 0)
		for _, id := range updateIDs {
			if _, ok := cloudMap[id]; ok {
				cloudMap[id].IsUpdate = true
			}
		}

		for k, v := range cloudMap {
			if !v.IsUpdate {
				addIDs = append(addIDs, k)
			}
		}

		if len(addIDs) > 0 {
			err := syncGcpCvmAdd(kt, addIDs, req, cloudMap, dataCli)
			if err != nil {
				logs.Errorf("request syncGcpCvmAdd failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}

		if len(token) == 0 {
			break
		}
		nextToken = token
	}

	dsIDs, err := getGcpCvmAllDSByVendor(kt, req, enumor.Gcp, dataCli)
	if err != nil {
		logs.Errorf("request getGcpCvmAllDSByVendor failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	deleteIDs := make([]string, 0)
	for _, id := range dsIDs {
		if _, ok := cloudAllIDs[id]; !ok {
			deleteIDs = append(deleteIDs, id)
		}
	}

	if len(deleteIDs) > 0 {
		realDeleteIDs := make([]string, 0)
		nextToken := ""
		for {
			opt := &typecvm.GcpListOption{
				Region: req.Region,
				Zone:   req.Zone,
				Page: &typecore.GcpPage{
					PageToken: nextToken,
					PageSize:  int64(filter.DefaultMaxInLimit),
				},
			}

			if nextToken != "" {
				opt.Page.PageToken = nextToken
			}

			if len(req.CloudIDs) > 0 {
				opt.Page = nil
				opt.CloudIDs = req.CloudIDs
			}

			datas, token, err := client.ListCvm(kt, opt)
			if err != nil {
				logs.Errorf("request adaptor to list gcp cvm failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}

			for _, id := range deleteIDs {
				realDeleteFlag := true
				for _, data := range datas {
					if fmt.Sprintf("%d", data.Id) == id {
						realDeleteFlag = false
						break
					}
				}

				if realDeleteFlag {
					realDeleteIDs = append(realDeleteIDs, id)
				}
			}

			if len(token) == 0 {
				break
			}
			nextToken = token
		}

		if len(realDeleteIDs) > 0 {
			err := syncCvmDelete(kt, realDeleteIDs, dataCli)
			if err != nil {
				logs.Errorf("request syncCvmDelete failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}
	}

	return nil, nil
}

func isChangeGcp(cloud *GcpCvmSync, db *GcpDSCvmSync) bool {

	if db.Cvm.CloudID != fmt.Sprintf("%d", cloud.Cvm.Id) {
		return true
	}

	if db.Cvm.Name != cloud.Cvm.Name {
		return true
	}

	cloudVpcIDs := make([]string, 0)
	cloudSubnetIDs := make([]string, 0)
	cloudNetWorkInterfaceIDs := make([]string, 0)
	if len(cloud.Cvm.NetworkInterfaces) > 0 {
		for _, networkInterface := range cloud.Cvm.NetworkInterfaces {
			if networkInterface != nil {
				cloudNetInterfaceID := fmt.Sprintf("%d", cloud.Cvm.Id) + "_" + networkInterface.Name
				cloudNetWorkInterfaceIDs = append(cloudNetWorkInterfaceIDs, cloudNetInterfaceID)
				cloudVpcIDs = append(cloudVpcIDs, networkInterface.Network)
				cloudSubnetIDs = append(cloudSubnetIDs, networkInterface.Subnetwork)
			}
		}
	}

	if len(db.Cvm.CloudVpcIDs) == 0 || len(cloudVpcIDs) == 0 || (db.Cvm.CloudVpcIDs[0] != cloudVpcIDs[0]) {
		return true
	}

	if len(db.Cvm.CloudSubnetIDs) == 0 || len(cloudSubnetIDs) == 0 ||
		!assert.IsStringSliceEqual(db.Cvm.CloudSubnetIDs, cloudSubnetIDs) {
		return true
	}

	if db.Cvm.CloudImageID != cloud.Cvm.SourceMachineImage {
		return true
	}

	if db.Cvm.Status != cloud.Cvm.Status {
		return true
	}

	priIPv4, pubIPv4, priIPv6, pubIPv6 := gcp.GetGcpIPAddresses(cloud.Cvm.NetworkInterfaces)

	if !assert.IsStringSliceEqual(db.Cvm.PrivateIPv4Addresses, priIPv4) {
		return true
	}

	if !assert.IsStringSliceEqual(db.Cvm.PublicIPv4Addresses, pubIPv4) {
		return true
	}

	if !assert.IsStringSliceEqual(db.Cvm.PrivateIPv6Addresses, priIPv6) {
		return true
	}

	if !assert.IsStringSliceEqual(db.Cvm.PublicIPv6Addresses, pubIPv6) {
		return true
	}

	if db.Cvm.MachineType != cloud.Cvm.MachineType {
		return true
	}

	if db.Cvm.CloudCreatedTime != cloud.Cvm.CreationTimestamp {
		return true
	}

	if db.Cvm.CloudExpiredTime != cloud.Cvm.LastStopTimestamp {
		return true
	}

	if db.Cvm.Extension.DeletionProtection != cloud.Cvm.DeletionProtection {
		return true
	}

	if db.Cvm.Extension.CpuPlatform != cloud.Cvm.CpuPlatform {
		return true
	}

	if db.Cvm.Extension.CanIpForward != cloud.Cvm.CanIpForward {
		return true
	}

	if !assert.IsStringSliceEqual(db.Cvm.Extension.CloudNetworkInterfaceIDs, cloudNetWorkInterfaceIDs) {
		return true
	}

	if db.Cvm.Extension.SelfLink != cloud.Cvm.SelfLink {
		return true
	}

	if db.Cvm.Extension.MinCpuPlatform != cloud.Cvm.MinCpuPlatform {
		return true
	}

	if db.Cvm.Extension.StartRestricted != cloud.Cvm.StartRestricted {
		return true
	}

	if !assert.IsStringSliceEqual(db.Cvm.Extension.ResourcePolicies, cloud.Cvm.ResourcePolicies) {
		return true
	}

	if db.Cvm.Extension.Fingerprint != cloud.Cvm.Fingerprint {
		return true
	}

	if db.Cvm.Extension.ReservationAffinity.ConsumeReservationType != cloud.Cvm.ReservationAffinity.ConsumeReservationType {
		return true
	}

	if db.Cvm.Extension.ReservationAffinity.Key != cloud.Cvm.ReservationAffinity.Key {
		return true
	}

	if !assert.IsStringSliceEqual(db.Cvm.Extension.ReservationAffinity.Values, cloud.Cvm.ReservationAffinity.Values) {
		return true
	}

	if db.Cvm.Extension.AdvancedMachineFeatures.EnableNestedVirtualization != cloud.Cvm.AdvancedMachineFeatures.EnableNestedVirtualization {
		return true
	}

	if db.Cvm.Extension.AdvancedMachineFeatures.EnableUefiNetworking != cloud.Cvm.AdvancedMachineFeatures.EnableUefiNetworking {
		return true
	}

	if db.Cvm.Extension.AdvancedMachineFeatures.ThreadsPerCore != cloud.Cvm.AdvancedMachineFeatures.ThreadsPerCore {
		return true
	}

	if db.Cvm.Extension.AdvancedMachineFeatures.VisibleCoreCount != cloud.Cvm.AdvancedMachineFeatures.ThreadsPerCore {
		return true
	}

	if !assert.IsStringMapEqual(db.Cvm.Extension.Labels, cloud.Cvm.Labels) {
		return true
	}

	disks := make([]corecvm.GcpAttachedDisk, 0)
	if len(cloud.Cvm.Disks) > 0 {
		for _, v := range cloud.Cvm.Disks {
			tmp := corecvm.GcpAttachedDisk{
				Boot:    v.Boot,
				Index:   v.Index,
				CloudID: v.Source,
			}
			disks = append(disks, tmp)
		}
	}

	for _, dbValue := range db.Cvm.Extension.Disks {
		isEqual := false
		for _, cloudValue := range cloud.Cvm.Disks {
			if dbValue.Boot == cloudValue.Boot && dbValue.Index == cloudValue.Index && dbValue.CloudID == cloudValue.Source {
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

func syncGcpCvmUpdate(kt *kit.Kit, updateIDs []string, cloudMap map[string]*GcpCvmSync,
	dsMap map[string]*GcpDSCvmSync, dataCli *dataservice.Client) error {

	lists := make([]dataproto.CvmBatchUpdate[corecvm.GcpCvmExtension], 0)

	for _, id := range updateIDs {
		if !isChangeGcp(cloudMap[id], dsMap[id]) {
			continue
		}

		cloudVpcIDs := make([]string, 0)
		cloudSubnetIDs := make([]string, 0)
		cloudNetWorkInterfaceIDs := make([]string, 0)
		if len(cloudMap[id].Cvm.NetworkInterfaces) > 0 {
			for _, networkInterface := range cloudMap[id].Cvm.NetworkInterfaces {
				if networkInterface != nil {
					cloudNetInterfaceID := fmt.Sprintf("%d", cloudMap[id].Cvm.Id) + "_" + networkInterface.Name
					cloudNetWorkInterfaceIDs = append(cloudNetWorkInterfaceIDs, cloudNetInterfaceID)
					cloudVpcIDs = append(cloudVpcIDs, networkInterface.Network)
					cloudSubnetIDs = append(cloudSubnetIDs, networkInterface.Subnetwork)
				}
			}
		}

		if len(cloudVpcIDs) <= 0 {
			return fmt.Errorf("gcp cvm: %s no vpc", fmt.Sprintf("%d", cloudMap[id].Cvm.Id))
		}

		if len(cloudVpcIDs) > 1 {
			logs.Errorf("gcp cvm: %s more than one vpc", fmt.Sprintf("%d", cloudMap[id].Cvm.Id))
		}

		vpcID, bkCloudID, err := queryVpcID(kt, dataCli, cloudVpcIDs[0])
		if err != nil {
			return err
		}

		subnetIDs, err := querySubnetIDs(kt, dataCli, cloudSubnetIDs)
		if err != nil {
			return err
		}

		disks := make([]corecvm.GcpAttachedDisk, 0)
		if len(cloudMap[id].Cvm.Disks) > 0 {
			for _, v := range cloudMap[id].Cvm.Disks {
				tmp := corecvm.GcpAttachedDisk{
					Boot:    v.Boot,
					Index:   v.Index,
					CloudID: v.Source,
				}
				disks = append(disks, tmp)
			}
		}

		priIPv4, pubIPv4, priIPv6, pubIPv6 := gcp.GetGcpIPAddresses(cloudMap[id].Cvm.NetworkInterfaces)
		cvm := dataproto.CvmBatchUpdate[corecvm.GcpCvmExtension]{
			ID:                   dsMap[id].Cvm.ID,
			Name:                 cloudMap[id].Cvm.Name,
			BkBizID:              constant.UnassignedBiz,
			BkCloudID:            bkCloudID,
			CloudVpcIDs:          cloudVpcIDs,
			VpcIDs:               []string{vpcID},
			CloudSubnetIDs:       cloudSubnetIDs,
			SubnetIDs:            subnetIDs,
			Memo:                 &cloudMap[id].Cvm.Description,
			Status:               cloudMap[id].Cvm.Status,
			PrivateIPv4Addresses: priIPv4,
			PrivateIPv6Addresses: pubIPv4,
			PublicIPv4Addresses:  priIPv6,
			PublicIPv6Addresses:  pubIPv6,
			CloudLaunchedTime:    cloudMap[id].Cvm.LastStartTimestamp,
			CloudExpiredTime:     cloudMap[id].Cvm.LastStopTimestamp,
			Extension: &corecvm.GcpCvmExtension{
				DeletionProtection:       cloudMap[id].Cvm.DeletionProtection,
				CpuPlatform:              cloudMap[id].Cvm.CpuPlatform,
				CanIpForward:             cloudMap[id].Cvm.CanIpForward,
				CloudNetworkInterfaceIDs: cloudNetWorkInterfaceIDs,
				Disks:                    disks,
				SelfLink:                 cloudMap[id].Cvm.SelfLink,
				Labels:                   cloudMap[id].Cvm.Labels,
				MinCpuPlatform:           cloudMap[id].Cvm.MinCpuPlatform,
				StartRestricted:          cloudMap[id].Cvm.StartRestricted,
				ResourcePolicies:         cloudMap[id].Cvm.ResourcePolicies,
				ReservationAffinity: &corecvm.GcpReservationAffinity{
					ConsumeReservationType: cloudMap[id].Cvm.ReservationAffinity.ConsumeReservationType,
					Key:                    cloudMap[id].Cvm.ReservationAffinity.Key,
					Values:                 cloudMap[id].Cvm.ReservationAffinity.Values,
				},
				Fingerprint: cloudMap[id].Cvm.Fingerprint,
				AdvancedMachineFeatures: &corecvm.GcpAdvancedMachineFeatures{
					EnableNestedVirtualization: cloudMap[id].Cvm.AdvancedMachineFeatures.EnableNestedVirtualization,
					EnableUefiNetworking:       cloudMap[id].Cvm.AdvancedMachineFeatures.EnableUefiNetworking,
					ThreadsPerCore:             cloudMap[id].Cvm.AdvancedMachineFeatures.ThreadsPerCore,
					VisibleCoreCount:           cloudMap[id].Cvm.AdvancedMachineFeatures.VisibleCoreCount,
				},
			},
		}

		lists = append(lists, cvm)
	}

	updateReq := dataproto.CvmBatchUpdateReq[corecvm.GcpCvmExtension]{
		Cvms: lists,
	}

	if len(updateReq.Cvms) > 0 {
		if err := dataCli.Gcp.Cvm.BatchUpdateCvm(kt.Ctx, kt.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateCvm failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func syncGcpCvmAdd(kt *kit.Kit, addIDs []string, req *SyncGcpCvmOption,
	cloudMap map[string]*GcpCvmSync, dataCli *dataservice.Client) error {

	lists := make([]dataproto.CvmBatchCreate[corecvm.GcpCvmExtension], 0)

	for _, id := range addIDs {
		cloudVpcIDs := make([]string, 0)
		cloudSubnetIDs := make([]string, 0)
		cloudNetWorkInterfaceIDs := make([]string, 0)
		if len(cloudMap[id].Cvm.NetworkInterfaces) > 0 {
			for _, networkInterface := range cloudMap[id].Cvm.NetworkInterfaces {
				if networkInterface != nil {
					cloudNetInterfaceID := fmt.Sprintf("%d", cloudMap[id].Cvm.Id) + "_" + networkInterface.Name
					cloudNetWorkInterfaceIDs = append(cloudNetWorkInterfaceIDs, cloudNetInterfaceID)
					cloudVpcIDs = append(cloudVpcIDs, networkInterface.Network)
					cloudSubnetIDs = append(cloudSubnetIDs, networkInterface.Subnetwork)
				}
			}
		}

		if len(cloudVpcIDs) <= 0 {
			return fmt.Errorf("gcp cvm: %s no vpc", fmt.Sprintf("%d", cloudMap[id].Cvm.Id))
		}

		if len(cloudVpcIDs) > 1 {
			logs.Errorf("gcp cvm: %s more than one vpc", fmt.Sprintf("%d", cloudMap[id].Cvm.Id))
		}

		vpcID, bkCloudID, err := queryVpcID(kt, dataCli, cloudVpcIDs[0])
		if err != nil {
			return err
		}

		subnetIDs, err := querySubnetIDs(kt, dataCli, cloudSubnetIDs)
		if err != nil {
			return err
		}

		disks := make([]corecvm.GcpAttachedDisk, 0)
		if len(cloudMap[id].Cvm.Disks) > 0 {
			for _, v := range cloudMap[id].Cvm.Disks {
				tmp := corecvm.GcpAttachedDisk{
					Boot:    v.Boot,
					Index:   v.Index,
					CloudID: v.Source,
				}
				disks = append(disks, tmp)
			}
		}

		priIPv4, pubIPv4, priIPv6, pubIPv6 := gcp.GetGcpIPAddresses(cloudMap[id].Cvm.NetworkInterfaces)
		cvm := dataproto.CvmBatchCreate[corecvm.GcpCvmExtension]{
			CloudID:        fmt.Sprintf("%d", cloudMap[id].Cvm.Id),
			Name:           cloudMap[id].Cvm.Name,
			BkBizID:        constant.UnassignedBiz,
			BkCloudID:      bkCloudID,
			AccountID:      req.AccountID,
			Region:         req.Region,
			Zone:           req.Zone,
			CloudVpcIDs:    cloudVpcIDs,
			VpcIDs:         []string{vpcID},
			CloudSubnetIDs: cloudSubnetIDs,
			SubnetIDs:      subnetIDs,
			CloudImageID:   cloudMap[id].Cvm.SourceMachineImage,
			// gcp镜像是与硬盘绑定的
			OsName:               "",
			Memo:                 &cloudMap[id].Cvm.Description,
			Status:               cloudMap[id].Cvm.Status,
			PrivateIPv4Addresses: priIPv4,
			PrivateIPv6Addresses: pubIPv4,
			PublicIPv4Addresses:  priIPv6,
			PublicIPv6Addresses:  pubIPv6,
			MachineType:          cloudMap[id].Cvm.MachineType,
			CloudCreatedTime:     cloudMap[id].Cvm.CreationTimestamp,
			CloudLaunchedTime:    cloudMap[id].Cvm.LastStartTimestamp,
			CloudExpiredTime:     cloudMap[id].Cvm.LastStopTimestamp,
			Extension: &corecvm.GcpCvmExtension{
				DeletionProtection:       cloudMap[id].Cvm.DeletionProtection,
				CpuPlatform:              cloudMap[id].Cvm.CpuPlatform,
				CanIpForward:             cloudMap[id].Cvm.CanIpForward,
				CloudNetworkInterfaceIDs: cloudNetWorkInterfaceIDs,
				Disks:                    disks,
				SelfLink:                 cloudMap[id].Cvm.SelfLink,
				Labels:                   cloudMap[id].Cvm.Labels,
				MinCpuPlatform:           cloudMap[id].Cvm.MinCpuPlatform,
				StartRestricted:          cloudMap[id].Cvm.StartRestricted,
				ResourcePolicies:         cloudMap[id].Cvm.ResourcePolicies,
				ReservationAffinity: &corecvm.GcpReservationAffinity{
					ConsumeReservationType: cloudMap[id].Cvm.ReservationAffinity.ConsumeReservationType,
					Key:                    cloudMap[id].Cvm.ReservationAffinity.Key,
					Values:                 cloudMap[id].Cvm.ReservationAffinity.Values,
				},
				Fingerprint: cloudMap[id].Cvm.Fingerprint,
				AdvancedMachineFeatures: &corecvm.GcpAdvancedMachineFeatures{
					EnableNestedVirtualization: cloudMap[id].Cvm.AdvancedMachineFeatures.EnableNestedVirtualization,
					EnableUefiNetworking:       cloudMap[id].Cvm.AdvancedMachineFeatures.EnableUefiNetworking,
					ThreadsPerCore:             cloudMap[id].Cvm.AdvancedMachineFeatures.ThreadsPerCore,
					VisibleCoreCount:           cloudMap[id].Cvm.AdvancedMachineFeatures.VisibleCoreCount,
				},
			},
		}

		lists = append(lists, cvm)
	}

	createReq := dataproto.CvmBatchCreateReq[corecvm.GcpCvmExtension]{
		Cvms: lists,
	}

	if len(createReq.Cvms) > 0 {
		_, err := dataCli.Gcp.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(), &createReq)
		if err != nil {
			logs.Errorf("request dataservice to create gcp cvm failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func getGcpCvmDSSync(kt *kit.Kit, cloudIDs []string, req *SyncGcpCvmOption,
	dataCli *dataservice.Client) ([]string, map[string]*GcpDSCvmSync, error) {

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*GcpDSCvmSync)

	start := 0
	for {
		dataReq := &dataproto.CvmListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "vendor",
						Op:    filter.Equal.Factory(),
						Value: enumor.Gcp,
					},
					&filter.AtomRule{
						Field: "account_id",
						Op:    filter.Equal.Factory(),
						Value: req.AccountID,
					},
					&filter.AtomRule{
						Field: "region",
						Op:    filter.Equal.Factory(),
						Value: req.Region,
					},
					&filter.AtomRule{
						Field: "zone",
						Op:    filter.Equal.Factory(),
						Value: req.Zone,
					},
					&filter.AtomRule{
						Field: "cloud_id",
						Op:    filter.In.Factory(),
						Value: cloudIDs,
					},
				},
			},
			Page: &core.BasePage{
				Start: uint32(start),
				Limit: core.DefaultMaxPageLimit,
			},
		}

		results, err := dataCli.Gcp.Cvm.ListCvmExt(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list cvm failed, err: %v, rid: %s", err, kt.Rid)
			return updateIDs, dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				updateIDs = append(updateIDs, detail.CloudID)
				dsImageSync := new(GcpDSCvmSync)
				dsImageSync.Cvm = detail
				dsMap[detail.CloudID] = dsImageSync
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	return updateIDs, dsMap, nil
}

func getGcpCvmAllDSByVendor(kt *kit.Kit, req *SyncGcpCvmOption,
	vendor enumor.Vendor, dataCli *dataservice.Client) ([]string, error) {

	dsIDs := make([]string, 0)

	start := 0
	for {
		dataReq := &dataproto.CvmListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "vendor",
						Op:    filter.Equal.Factory(),
						Value: vendor,
					},
					&filter.AtomRule{
						Field: "account_id",
						Op:    filter.Equal.Factory(),
						Value: req.AccountID,
					},
					&filter.AtomRule{
						Field: "region",
						Op:    filter.Equal.Factory(),
						Value: req.Region,
					},
					&filter.AtomRule{
						Field: "zone",
						Op:    filter.Equal.Factory(),
						Value: req.Zone,
					},
				},
			},
			Page: &core.BasePage{
				Start: uint32(start),
				Limit: core.DefaultMaxPageLimit,
			},
		}

		if len(req.CloudIDs) > 0 {
			filter := filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: req.CloudIDs}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		results, err := dataCli.Gcp.Cvm.ListCvmExt(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list cvm failed, err: %v, rid: %s", err, kt.Rid)
			return dsIDs, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				dsIDs = append(dsIDs, detail.CloudID)
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	return dsIDs, nil
}

// SyncGcpCvmWithRelResource sync cvm all rel resource
func SyncGcpCvmWithRelResource(kt *kit.Kit, ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client,
	req *SyncGcpCvmOption) (interface{}, error) {

	client, err := ad.Gcp(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	cloudNetInterMap, cloudVpcMap, cloudSubnetMap, cloudEipMap, cloudDiskMap, err := getGcpCVMRelResourcesIDs(kt,
		req, client)
	if err != nil {
		logs.Errorf("request get gcp cvm rel resource ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(cloudVpcMap) > 0 {
		vpcCloudIDs := make([]string, 0)
		for _, id := range cloudVpcMap {
			vpcCloudIDs = append(vpcCloudIDs, id.RelID)
		}
		req := &hcservice.GcpResourceSyncReq{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  vpcCloudIDs,
		}
		_, err := vpc.GcpVpcSync(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("request to sync gcp vpc logic failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudSubnetMap) > 0 {
		subnetCloudIDs := make([]string, 0)
		for _, id := range cloudSubnetMap {
			subnetCloudIDs = append(subnetCloudIDs, id.RelID)
		}
		req := &hcservice.GcpResourceSyncReq{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  subnetCloudIDs,
		}
		_, err := subnet.GcpSubnetSync(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("request to sync gcp subnet logic failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudEipMap) > 0 {
		eipCloudIDs := make([]string, 0)
		for _, id := range cloudEipMap {
			eipCloudIDs = append(eipCloudIDs, id.RelID)
		}
		req := &protoeip.EipSyncReq{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  eipCloudIDs,
		}
		_, err := synceip.SyncGcpEip(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("sync gcp cvm rel eip failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudNetInterMap) > 0 {
		netInterCloudIDs := make([]string, 0)
		for _, id := range cloudNetInterMap {
			netInterCloudIDs = append(netInterCloudIDs, id.RelID)
		}
		req := &hcservice.GcpNetworkInterfaceSyncReq{
			AccountID:   req.AccountID,
			Zone:        req.Zone,
			CloudCvmIDs: netInterCloudIDs,
		}
		_, err := syncnetworkinterface.GcpNetworkInterfaceSync(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("request to sync gcp networkinterface logic failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudDiskMap) > 0 {
		diskCloudIDs := make([]string, 0)
		for _, id := range cloudDiskMap {
			diskCloudIDs = append(diskCloudIDs, id.RelID)
		}
		req := &protodisk.DiskSyncReq{
			AccountID: req.AccountID,
			Zone:      req.Zone,
			Region:    req.Region,
			CloudIDs:  diskCloudIDs,
		}
		_, err := disk.SyncGcpDisk(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("sync gcp cvm rel disk failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	cvmReq := &SyncGcpCvmOption{
		AccountID: req.AccountID,
		Region:    req.Region,
		Zone:      req.Zone,
		CloudIDs:  req.CloudIDs,
	}
	_, err = SyncGcpCvm(kt, ad, dataCli, cvmReq)
	if err != nil {
		logs.Errorf("sync gcp cvm self failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	hcReq := &protocvm.OperateSyncReq{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  req.CloudIDs,
	}

	err = getDiskHcIDs(kt, hcReq, dataCli, cloudDiskMap)
	if err != nil {
		logs.Errorf("request get cvm disk rel resource hc ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	err = getEipHcIDs(kt, hcReq, dataCli, cloudEipMap)
	if err != nil {
		logs.Errorf("request get cvm eip rel resource hc ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	err = getNetworkInterfaceHcIDs(kt, hcReq, dataCli, cloudNetInterMap)
	if err != nil {
		logs.Errorf("request get cvm networkinterface rel resource hc ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(cloudNetInterMap) > 0 {
		err := SyncCvmNetworkInterfaceRel(kt, cloudNetInterMap, dataCli)
		if err != nil {
			logs.Errorf("sync gcp cvm networkinterface rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudEipMap) > 0 {
		err := SyncCvmEipRel(kt, cloudEipMap, dataCli)
		if err != nil {
			logs.Errorf("sync gcp cvm eip rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudDiskMap) > 0 {
		err := SyncCvmDiskRel(kt, cloudDiskMap, dataCli)
		if err != nil {
			logs.Errorf("sync gcp cvm disk rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func getGcpCVMRelResourcesIDs(kt *kit.Kit, req *SyncGcpCvmOption,
	client *gcp.Gcp) (map[string]*CVMOperateSync, map[string]*CVMOperateSync,
	map[string]*CVMOperateSync, map[string]*CVMOperateSync, map[string]*CVMOperateSync, error) {

	netInterMap := make(map[string]*CVMOperateSync)
	vpcMap := make(map[string]*CVMOperateSync)
	subnetMap := make(map[string]*CVMOperateSync)
	eipMap := make(map[string]*CVMOperateSync)
	diskMap := make(map[string]*CVMOperateSync)

	opt := &typecvm.GcpListOption{
		Region:   req.Region,
		Zone:     req.Zone,
		CloudIDs: req.CloudIDs,
	}

	datas, _, err := client.ListCvm(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to list gcp cvm failed, err: %v, rid: %s", err, kt.Rid)
		return netInterMap, vpcMap, subnetMap, eipMap, diskMap, err
	}

	for _, data := range datas {
		if len(data.NetworkInterfaces) > 0 {
			for _, networkInterface := range data.NetworkInterfaces {
				if networkInterface != nil {
					netInterId := fmt.Sprintf("%d", data.Id) + "_" + networkInterface.Name
					netInterMapId := getCVMRelID(netInterId, fmt.Sprintf("%d", data.Id))
					netInterMap[netInterMapId] = &CVMOperateSync{RelID: netInterId, InstanceID: fmt.Sprintf("%d", data.Id)}

					netWorkId := getCVMRelID(networkInterface.Network, fmt.Sprintf("%d", data.Id))
					vpcMap[netWorkId] = &CVMOperateSync{RelID: networkInterface.Network, InstanceID: fmt.Sprintf("%d", data.Id)}

					subNetId := getCVMRelID(networkInterface.Subnetwork, fmt.Sprintf("%d", data.Id))
					subnetMap[subNetId] = &CVMOperateSync{RelID: networkInterface.Subnetwork, InstanceID: fmt.Sprintf("%d", data.Id)}

					if len(networkInterface.AccessConfigs) > 0 {
						ipAddresses := make([]string, 0, len(networkInterface.AccessConfigs))
						for _, config := range networkInterface.AccessConfigs {
							ipAddresses = append(ipAddresses, config.NatIP)
						}
						opt := &eip.GcpEipAggregatedListOption{
							IPAddresses: ipAddresses,
						}
						// 外部临时IP该接口无法查询出来，但这部分IP会飘，随着主机重启，不属于弹性IP，属于正常情况。
						eips, err := client.ListAggregatedEip(kt, opt)
						if err != nil {
							logs.Errorf("request adaptor to aggregate list gcp eip by ip failed, err: %v, rid: %s",
								err, kt.Rid)
							return nil, nil, nil, nil, nil, err
						}

						for _, one := range eips {
							id := getCVMRelID(fmt.Sprintf("%d", one.Id), fmt.Sprintf("%d", data.Id))
							eipMap[id] = &CVMOperateSync{RelID: fmt.Sprintf("%d", one.Id),
								InstanceID: fmt.Sprintf("%d", data.Id)}
						}
					}
				}
			}
		}

		if len(data.Disks) > 0 {
			for _, disk := range data.Disks {
				if disk != nil {
					id := getCVMRelID(disk.Source, fmt.Sprintf("%d", data.Id))
					diskMap[id] = &CVMOperateSync{RelID: disk.Source, InstanceID: fmt.Sprintf("%d", data.Id)}
				}
			}
		}
	}

	return netInterMap, vpcMap, subnetMap, eipMap, diskMap, nil
}

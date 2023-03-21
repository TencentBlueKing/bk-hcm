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
	diskproto "hcm/pkg/api/data-service/cloud/disk"
	hcservice "hcm/pkg/api/hc-service"
	protocvm "hcm/pkg/api/hc-service/cvm"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/slice"
)

// SyncGcpCvmOption ...
type SyncGcpCvmOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	Zone      string   `json:"zone" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncGcpCvmOption
func (opt SyncGcpCvmOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.RelResourceOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.RelResourceOperationMaxLimit)
	}

	return nil
}

// SyncGcpCvm sync cvm self
func SyncGcpCvm(kt *kit.Kit, ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client,
	req *SyncGcpCvmOption) (interface{}, error) {

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

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

	vpcSelfLinks := make([]string, 0)
	subnetSelfLinks := make([]string, 0)
	cloudNetWorkInterfaceIDs := make([]string, 0)
	if len(cloud.Cvm.NetworkInterfaces) > 0 {
		for _, networkInterface := range cloud.Cvm.NetworkInterfaces {
			if networkInterface != nil {
				cloudNetInterfaceID := fmt.Sprintf("%d", cloud.Cvm.Id) + "_" + networkInterface.Name
				cloudNetWorkInterfaceIDs = append(cloudNetWorkInterfaceIDs, cloudNetInterfaceID)
				vpcSelfLinks = append(vpcSelfLinks, networkInterface.Network)
				subnetSelfLinks = append(subnetSelfLinks, networkInterface.Subnetwork)
			}
		}
	}

	if len(db.Cvm.Extension.VpcSelfLinks) == 0 || len(vpcSelfLinks) == 0 ||
		(db.Cvm.Extension.VpcSelfLinks[0] != vpcSelfLinks[0]) {
		return true
	}

	if len(db.Cvm.Extension.SubnetSelfLinks) == 0 || len(subnetSelfLinks) == 0 ||
		!assert.IsStringSliceEqual(db.Cvm.Extension.SubnetSelfLinks, subnetSelfLinks) {
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

	if db.Cvm.MachineType != gcp.GetMachineType(cloud.Cvm.MachineType) {
		return true
	}

	if db.Cvm.CloudCreatedTime != cloud.Cvm.CreationTimestamp {
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

	if (db.Cvm.Extension.AdvancedMachineFeatures != nil && cloud.Cvm.AdvancedMachineFeatures == nil) ||
		(db.Cvm.Extension.AdvancedMachineFeatures == nil && cloud.Cvm.AdvancedMachineFeatures != nil) {
		return true
	}

	if db.Cvm.Extension.AdvancedMachineFeatures != nil && cloud.Cvm.AdvancedMachineFeatures != nil {
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
	}

	if !assert.IsStringMapEqual(db.Cvm.Extension.Labels, cloud.Cvm.Labels) {
		return true
	}

	for _, dbValue := range db.Cvm.Extension.Disks {
		isEqual := false
		for _, cloudValue := range cloud.Cvm.Disks {
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

func syncGcpCvmUpdate(kt *kit.Kit, updateIDs []string, cloudMap map[string]*GcpCvmSync,
	dsMap map[string]*GcpDSCvmSync, dataCli *dataservice.Client) error {

	lists := make([]dataproto.CvmBatchUpdate[corecvm.GcpCvmExtension], 0)

	for _, id := range updateIDs {
		if !isChangeGcp(cloudMap[id], dsMap[id]) {
			continue
		}

		vpcSelfLinks := make([]string, 0)
		subnetSelfLinks := make([]string, 0)
		cloudNetWorkInterfaceIDs := make([]string, 0)
		if len(cloudMap[id].Cvm.NetworkInterfaces) > 0 {
			for _, networkInterface := range cloudMap[id].Cvm.NetworkInterfaces {
				if networkInterface != nil {
					cloudNetInterfaceID := fmt.Sprintf("%d", cloudMap[id].Cvm.Id) + "_" + networkInterface.Name
					cloudNetWorkInterfaceIDs = append(cloudNetWorkInterfaceIDs, cloudNetInterfaceID)
					vpcSelfLinks = append(vpcSelfLinks, networkInterface.Network)
					subnetSelfLinks = append(subnetSelfLinks, networkInterface.Subnetwork)
				}
			}
		}

		if len(vpcSelfLinks) == 0 {
			return fmt.Errorf("gcp cvm: %s no vpc", fmt.Sprintf("%d", cloudMap[id].Cvm.Id))
		}

		if len(vpcSelfLinks) > 1 {
			logs.Errorf("gcp cvm: %s more than one vpc", fmt.Sprintf("%d", cloudMap[id].Cvm.Id))
		}

		vpcID, cloudVpcID, bkCloudID, err := queryVpcIDBySelfLink(kt, dataCli, vpcSelfLinks[0])
		if err != nil {
			return err
		}

		subnetIDs, cloudSubIDs, err := querySubnetIDsBySelfLink(kt, dataCli, subnetSelfLinks)
		if err != nil {
			return err
		}

		diskSelfLinks := make([]string, 0, len(cloudMap[id].Cvm.Disks))
		for _, one := range cloudMap[id].Cvm.Disks {
			diskSelfLinks = append(diskSelfLinks, one.Source)
		}

		cloudIDMap, err := queryCloudDiskIDMapBySelfLink(kt, dataCli, diskSelfLinks)
		if err != nil {
			return err
		}

		disks := make([]corecvm.GcpAttachedDisk, 0)
		for _, v := range cloudMap[id].Cvm.Disks {
			cloudID, exist := cloudIDMap[v.Source]
			if !exist {
				return fmt.Errorf("cvm: %d not found disk: %s in db", cloudMap[id].Cvm.Id, v.Source)
			}

			tmp := corecvm.GcpAttachedDisk{
				SelfLink:   v.Source,
				Boot:       v.Boot,
				Index:      v.Index,
				CloudID:    cloudID,
				DeviceName: v.DeviceName,
			}
			disks = append(disks, tmp)
		}

		priIPv4, pubIPv4, priIPv6, pubIPv6 := gcp.GetGcpIPAddresses(cloudMap[id].Cvm.NetworkInterfaces)
		cvm := dataproto.CvmBatchUpdate[corecvm.GcpCvmExtension]{
			ID:                   dsMap[id].Cvm.ID,
			Name:                 cloudMap[id].Cvm.Name,
			BkCloudID:            bkCloudID,
			CloudVpcIDs:          []string{cloudVpcID},
			VpcIDs:               []string{vpcID},
			CloudSubnetIDs:       cloudSubIDs,
			SubnetIDs:            subnetIDs,
			Memo:                 &cloudMap[id].Cvm.Description,
			Status:               cloudMap[id].Cvm.Status,
			PrivateIPv4Addresses: priIPv4,
			PrivateIPv6Addresses: priIPv6,
			PublicIPv4Addresses:  pubIPv4,
			PublicIPv6Addresses:  pubIPv6,
			CloudLaunchedTime:    cloudMap[id].Cvm.LastStartTimestamp,
			CloudExpiredTime:     cloudMap[id].Cvm.LastStopTimestamp,
			Extension: &corecvm.GcpCvmExtension{
				VpcSelfLinks:             vpcSelfLinks,
				SubnetSelfLinks:          subnetSelfLinks,
				DeletionProtection:       cloudMap[id].Cvm.DeletionProtection,
				CanIpForward:             cloudMap[id].Cvm.CanIpForward,
				CloudNetworkInterfaceIDs: cloudNetWorkInterfaceIDs,
				Disks:                    disks,
				SelfLink:                 cloudMap[id].Cvm.SelfLink,
				CpuPlatform:              cloudMap[id].Cvm.CpuPlatform,
				Labels:                   cloudMap[id].Cvm.Labels,
				MinCpuPlatform:           cloudMap[id].Cvm.MinCpuPlatform,
				StartRestricted:          cloudMap[id].Cvm.StartRestricted,
				ResourcePolicies:         cloudMap[id].Cvm.ResourcePolicies,
				ReservationAffinity:      nil,
				Fingerprint:              cloudMap[id].Cvm.Fingerprint,
				AdvancedMachineFeatures:  nil,
			},
		}

		if cloudMap[id].Cvm.ReservationAffinity != nil {
			cvm.Extension.ReservationAffinity = &corecvm.GcpReservationAffinity{
				ConsumeReservationType: cloudMap[id].Cvm.ReservationAffinity.ConsumeReservationType,
				Key:                    cloudMap[id].Cvm.ReservationAffinity.Key,
				Values:                 cloudMap[id].Cvm.ReservationAffinity.Values,
			}
		}

		if cloudMap[id].Cvm.AdvancedMachineFeatures != nil {
			cvm.Extension.AdvancedMachineFeatures = &corecvm.GcpAdvancedMachineFeatures{
				EnableNestedVirtualization: cloudMap[id].Cvm.AdvancedMachineFeatures.EnableNestedVirtualization,
				EnableUefiNetworking:       cloudMap[id].Cvm.AdvancedMachineFeatures.EnableUefiNetworking,
				ThreadsPerCore:             cloudMap[id].Cvm.AdvancedMachineFeatures.ThreadsPerCore,
				VisibleCoreCount:           cloudMap[id].Cvm.AdvancedMachineFeatures.VisibleCoreCount,
			}
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

func queryCloudDiskIDMapBySelfLink(kt *kit.Kit, dataCli *dataservice.Client, selfLinks []string) (
	map[string]string, error) {

	unique := slice.Unique(selfLinks)
	req := &diskproto.DiskListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "extension.self_link",
					Op:    filter.JSONIn.Factory(),
					Value: unique,
				},
			},
		},
		Page:   core.DefaultBasePage,
		Fields: []string{"cloud_id", "extension"},
	}
	result, err := dataCli.Gcp.ListDisk(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("list disk failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	if len(result.Details) != len(unique) {
		logs.Errorf("list disk but some disk not found, selfLinks: %v, count: %d, rid: %s", unique,
			len(result.Details), kt.Rid)
		return nil, errf.Newf(errf.RecordNotFound, "some disk not found")
	}

	cloudIDMap := make(map[string]string)
	for _, v := range result.Details {
		cloudIDMap[v.Extension.SelfLink] = v.CloudID
	}
	return cloudIDMap, nil
}

func syncGcpCvmAdd(kt *kit.Kit, addIDs []string, req *SyncGcpCvmOption,
	cloudMap map[string]*GcpCvmSync, dataCli *dataservice.Client) error {

	lists := make([]dataproto.CvmBatchCreate[corecvm.GcpCvmExtension], 0)

	for _, id := range addIDs {
		vpcSelfLinks := make([]string, 0)
		subnetSelfLinks := make([]string, 0)
		cloudNetWorkInterfaceIDs := make([]string, 0)
		if len(cloudMap[id].Cvm.NetworkInterfaces) > 0 {
			for _, networkInterface := range cloudMap[id].Cvm.NetworkInterfaces {
				if networkInterface != nil {
					cloudNetInterfaceID := fmt.Sprintf("%d", cloudMap[id].Cvm.Id) + "_" + networkInterface.Name
					cloudNetWorkInterfaceIDs = append(cloudNetWorkInterfaceIDs, cloudNetInterfaceID)
					vpcSelfLinks = append(vpcSelfLinks, networkInterface.Network)
					subnetSelfLinks = append(subnetSelfLinks, networkInterface.Subnetwork)
				}
			}
		}

		if len(vpcSelfLinks) <= 0 {
			return fmt.Errorf("gcp cvm: %s no vpc", fmt.Sprintf("%d", cloudMap[id].Cvm.Id))
		}

		if len(vpcSelfLinks) > 1 {
			logs.Errorf("gcp cvm: %s more than one vpc", fmt.Sprintf("%d", cloudMap[id].Cvm.Id))
		}

		vpcID, cloudVpcID, bkCloudID, err := queryVpcIDBySelfLink(kt, dataCli, vpcSelfLinks[0])
		if err != nil {
			return err
		}

		subnetIDs, cloudSubIDs, err := querySubnetIDsBySelfLink(kt, dataCli, subnetSelfLinks)
		if err != nil {
			return err
		}

		diskSelfLinks := make([]string, 0, len(cloudMap[id].Cvm.Disks))
		for _, one := range cloudMap[id].Cvm.Disks {
			diskSelfLinks = append(diskSelfLinks, one.Source)
		}

		cloudIDMap, err := queryCloudDiskIDMapBySelfLink(kt, dataCli, diskSelfLinks)
		if err != nil {
			return err
		}

		disks := make([]corecvm.GcpAttachedDisk, 0)
		for _, v := range cloudMap[id].Cvm.Disks {
			cloudID, exist := cloudIDMap[v.Source]
			if !exist {
				return fmt.Errorf("cvm: %d not found disk: %s in db", cloudMap[id].Cvm.Id, v.Source)
			}

			tmp := corecvm.GcpAttachedDisk{
				SelfLink:   v.Source,
				Boot:       v.Boot,
				Index:      v.Index,
				CloudID:    cloudID,
				DeviceName: v.DeviceName,
			}
			disks = append(disks, tmp)
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
			CloudVpcIDs:    []string{cloudVpcID},
			VpcIDs:         []string{vpcID},
			CloudSubnetIDs: cloudSubIDs,
			SubnetIDs:      subnetIDs,
			CloudImageID:   cloudMap[id].Cvm.SourceMachineImage,
			// gcp镜像是与硬盘绑定的
			OsName:               "",
			Memo:                 &cloudMap[id].Cvm.Description,
			Status:               cloudMap[id].Cvm.Status,
			PrivateIPv4Addresses: priIPv4,
			PrivateIPv6Addresses: priIPv6,
			PublicIPv4Addresses:  pubIPv4,
			PublicIPv6Addresses:  pubIPv6,
			MachineType:          gcp.GetMachineType(cloudMap[id].Cvm.MachineType),
			CloudCreatedTime:     cloudMap[id].Cvm.CreationTimestamp,
			CloudLaunchedTime:    cloudMap[id].Cvm.LastStartTimestamp,
			CloudExpiredTime:     cloudMap[id].Cvm.LastStopTimestamp,
			Extension: &corecvm.GcpCvmExtension{
				VpcSelfLinks:             vpcSelfLinks,
				SubnetSelfLinks:          subnetSelfLinks,
				DeletionProtection:       cloudMap[id].Cvm.DeletionProtection,
				CanIpForward:             cloudMap[id].Cvm.CanIpForward,
				CloudNetworkInterfaceIDs: cloudNetWorkInterfaceIDs,
				Disks:                    disks,
				SelfLink:                 cloudMap[id].Cvm.SelfLink,
				CpuPlatform:              cloudMap[id].Cvm.CpuPlatform,
				Labels:                   cloudMap[id].Cvm.Labels,
				MinCpuPlatform:           cloudMap[id].Cvm.MinCpuPlatform,
				StartRestricted:          cloudMap[id].Cvm.StartRestricted,
				ResourcePolicies:         cloudMap[id].Cvm.ResourcePolicies,
				ReservationAffinity:      nil,
				Fingerprint:              cloudMap[id].Cvm.Fingerprint,
				AdvancedMachineFeatures:  nil,
			},
		}

		if cloudMap[id].Cvm.ReservationAffinity != nil {
			cvm.Extension.ReservationAffinity = &corecvm.GcpReservationAffinity{
				ConsumeReservationType: cloudMap[id].Cvm.ReservationAffinity.ConsumeReservationType,
				Key:                    cloudMap[id].Cvm.ReservationAffinity.Key,
				Values:                 cloudMap[id].Cvm.ReservationAffinity.Values,
			}
		}

		if cloudMap[id].Cvm.AdvancedMachineFeatures != nil {
			cvm.Extension.AdvancedMachineFeatures = &corecvm.GcpAdvancedMachineFeatures{
				EnableNestedVirtualization: cloudMap[id].Cvm.AdvancedMachineFeatures.EnableNestedVirtualization,
				EnableUefiNetworking:       cloudMap[id].Cvm.AdvancedMachineFeatures.EnableUefiNetworking,
				ThreadsPerCore:             cloudMap[id].Cvm.AdvancedMachineFeatures.ThreadsPerCore,
				VisibleCoreCount:           cloudMap[id].Cvm.AdvancedMachineFeatures.VisibleCoreCount,
			}
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

	if len(cloudIDs) <= 0 {
		return updateIDs, dsMap, nil
	}

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

	niCloudIDMap, vpcSLMap, subnetSLMap, eipCloudIDMap, diskSLMap, err := getGcpCVMRelResourcesIDs(kt, req, client)
	if err != nil {
		logs.Errorf("request get gcp cvm rel resource ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(vpcSLMap) > 0 {
		vpcSelfLinks := make([]string, 0)
		for _, id := range vpcSLMap {
			vpcSelfLinks = append(vpcSelfLinks, id.RelID)
		}
		req := &vpc.SyncGcpOption{
			AccountID: req.AccountID,
			SelfLinks: vpcSelfLinks,
		}
		_, err := vpc.GcpVpcSync(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("request to sync gcp vpc logic failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(subnetSLMap) > 0 {
		subnetSelfLinks := make([]string, 0)
		for _, id := range subnetSLMap {
			subnetSelfLinks = append(subnetSelfLinks, id.RelID)
		}
		req := &subnet.SyncGcpOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			SelfLinks: subnetSelfLinks,
		}
		_, err := subnet.GcpSubnetSync(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("request to sync gcp subnet logic failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(eipCloudIDMap) > 0 {
		cloudIDs := make([]string, 0)
		for _, id := range eipCloudIDMap {
			cloudIDs = append(cloudIDs, id.RelID)
		}
		req := &synceip.SyncGcpEipOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  cloudIDs,
		}
		_, err := synceip.SyncGcpEip(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("sync gcp cvm rel eip failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(niCloudIDMap) > 0 {
		netInterCloudIDs := make([]string, 0)
		for _, id := range niCloudIDMap {
			netInterCloudIDs = append(netInterCloudIDs, id.RelID)
		}
		req := &hcservice.GcpNetworkInterfaceSyncReq{
			AccountID:   req.AccountID,
			Zone:        req.Zone,
			CloudCvmIDs: req.CloudIDs,
		}
		_, err := syncnetworkinterface.GcpNetworkInterfaceSync(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("request to sync gcp networkinterface logic failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(diskSLMap) > 0 {
		diskSelfLinks := make([]string, 0)
		for _, id := range diskSLMap {
			diskSelfLinks = append(diskSelfLinks, id.RelID)
		}
		req := &disk.SyncGcpDiskOption{
			AccountID: req.AccountID,
			Zone:      req.Zone,
			SelfLinks: diskSelfLinks,
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
	err = getDiskHcIDsForGcp(kt, hcReq, dataCli, diskSLMap)
	if err != nil {
		logs.Errorf("request get cvm disk rel resource hc ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	err = getEipHcIDs(kt, hcReq, dataCli, eipCloudIDMap)
	if err != nil {
		logs.Errorf("request get cvm eip rel resource hc ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	err = getNetworkInterfaceHcIDs(kt, hcReq, dataCli, niCloudIDMap)
	if err != nil {
		logs.Errorf("request get cvm networkinterface rel resource hc ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(niCloudIDMap) > 0 {
		err := SyncCvmNetworkInterfaceRel(kt, niCloudIDMap, dataCli)
		if err != nil {
			logs.Errorf("sync gcp cvm networkinterface rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(eipCloudIDMap) > 0 {
		err := SyncCvmEipRel(kt, eipCloudIDMap, dataCli)
		if err != nil {
			logs.Errorf("sync gcp cvm eip rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(diskSLMap) > 0 {
		err := SyncCvmDiskRel(kt, diskSLMap, dataCli)
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
		return nil, nil, nil, nil, nil, err
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

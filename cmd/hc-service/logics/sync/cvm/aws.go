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
	securitygroup "hcm/cmd/hc-service/logics/sync/security-group"
	"hcm/cmd/hc-service/logics/sync/subnet"
	"hcm/cmd/hc-service/logics/sync/vpc"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/adaptor/aws"
	typecore "hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	dataproto "hcm/pkg/api/data-service/cloud"
	protocvm "hcm/pkg/api/hc-service/cvm"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
)

// SyncAwsCvmOption define sync aws cvm option.
type SyncAwsCvmOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncAwsCvmOption
func (opt SyncAwsCvmOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// SyncAwsCvm sync cvm self
func SyncAwsCvm(kt *kit.Kit, ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client,
	req *SyncAwsCvmOption) (interface{}, error) {

	client, err := ad.Aws(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	nextToken := ""
	cloudAllIDs := make(map[string]bool)
	for {
		opt := &typecvm.AwsListOption{
			Region: req.Region,
			Page: &typecore.AwsPage{
				MaxResults: converter.ValToPtr(int64(filter.DefaultMaxInLimit)),
			},
		}

		if nextToken != "" {
			opt.Page.NextToken = converter.ValToPtr(nextToken)
		}

		if len(req.CloudIDs) > 0 {
			opt.Page = nil
			opt.CloudIDs = req.CloudIDs
		}

		datas, err := client.ListCvm(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor to list aws cvm failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		cloudMap := make(map[string]*AwsCvmSync)
		cloudIDs := make([]string, 0)
		for _, reservation := range datas.Reservations {
			for _, data := range reservation.Instances {
				cvmSync := new(AwsCvmSync)
				cvmSync.IsUpdate = false
				cvmSync.Cvm = data
				cloudMap[*data.InstanceId] = cvmSync
				cloudIDs = append(cloudIDs, *data.InstanceId)
				cloudAllIDs[*data.InstanceId] = true
			}
		}

		if len(cloudIDs) <= 0 {
			break
		}

		updateIDs, dsMap, err := getAwsCvmDSSync(kt, cloudIDs, req, dataCli)
		if err != nil {
			logs.Errorf("request getAwsCvmDSSync failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if len(updateIDs) > 0 {
			err := syncAwsCvmUpdate(kt, updateIDs, cloudMap, dsMap, dataCli)
			if err != nil {
				logs.Errorf("request syncAwsCvmUpdate failed, err: %v, rid: %s", err, kt.Rid)
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
			err := syncAwsCvmAdd(kt, addIDs, req, cloudMap, dataCli)
			if err != nil {
				logs.Errorf("request syncAwsCvmAdd failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}

		if datas.NextToken == nil {
			break
		}
		nextToken = *datas.NextToken
	}

	dsIDs, err := getAwsCvmAllDSByVendor(kt, req, enumor.Aws, dataCli)
	if err != nil {
		logs.Errorf("request getAwsCvmAllDSByVendor failed, err: %v, rid: %s", err, kt.Rid)
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
			opt := &typecvm.AwsListOption{
				Region: req.Region,
				Page: &typecore.AwsPage{
					MaxResults: converter.ValToPtr(int64(filter.DefaultMaxInLimit)),
				},
			}

			if nextToken != "" {
				opt.Page.NextToken = converter.ValToPtr(nextToken)
			}

			if len(req.CloudIDs) > 0 {
				opt.Page = nil
				opt.CloudIDs = req.CloudIDs
			}

			datas, err := client.ListCvm(kt, opt)
			if err != nil {
				logs.Errorf("request adaptor to list aws cvm failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}

			for _, id := range deleteIDs {
				realDeleteFlag := true
				for _, reservation := range datas.Reservations {
					for _, data := range reservation.Instances {
						if *data.InstanceId == id {
							realDeleteFlag = false
							break
						}
					}
				}

				if realDeleteFlag {
					realDeleteIDs = append(realDeleteIDs, id)
				}
			}

			if datas.NextToken == nil {
				break
			}
			nextToken = *datas.NextToken
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

func isChangeAws(cloud *AwsCvmSync, db *AwsDSCvmSync) bool {

	if db.Cvm.CloudID != *cloud.Cvm.InstanceId {
		return true
	}

	if db.Cvm.Name != converter.PtrToVal(aws.GetCvmNameFromTags(cloud.Cvm.Tags)) {
		return true
	}

	if db.Cvm.CloudImageID != *cloud.Cvm.ImageId {
		return true
	}

	if db.Cvm.OsName != *cloud.Cvm.PlatformDetails {
		return true
	}

	if db.Cvm.Status != *cloud.Cvm.State.Name {
		return true
	}

	if len(db.Cvm.CloudVpcIDs) == 0 || (db.Cvm.CloudVpcIDs[0] != *cloud.Cvm.VpcId) {
		return true
	}

	if len(db.Cvm.CloudSubnetIDs) == 0 || (db.Cvm.CloudSubnetIDs[0] != *cloud.Cvm.SubnetId) {
		return true
	}

	privateIPv4Addresses := make([]string, 0)
	if cloud.Cvm.PrivateIpAddress != nil {
		privateIPv4Addresses = append(privateIPv4Addresses, *cloud.Cvm.PrivateIpAddress)
	}
	publicIPv4Addresses := make([]string, 0)
	if cloud.Cvm.PublicIpAddress != nil {
		publicIPv4Addresses = append(publicIPv4Addresses, *cloud.Cvm.PublicIpAddress)
	}
	publicIPv6Addresses := make([]string, 0)
	if cloud.Cvm.Ipv6Address != nil {
		publicIPv6Addresses = append(publicIPv6Addresses, *cloud.Cvm.Ipv6Address)
	}

	if !assert.IsStringSliceEqual(privateIPv4Addresses, db.Cvm.PrivateIPv4Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(publicIPv4Addresses, db.Cvm.PublicIPv4Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(publicIPv6Addresses, db.Cvm.PublicIPv6Addresses) {
		return true
	}

	if db.Cvm.MachineType != *cloud.Cvm.InstanceType {
		return true
	}

	if db.Cvm.CloudExpiredTime != cloud.Cvm.LaunchTime.String() {
		return true
	}

	if !assert.IsPtrStringEqual(db.Cvm.Extension.Platform, cloud.Cvm.Platform) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Cvm.Extension.DnsName, cloud.Cvm.PublicDnsName) {
		return true
	}

	if !assert.IsPtrBoolEqual(db.Cvm.Extension.EbsOptimized, cloud.Cvm.EbsOptimized) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Cvm.Extension.PrivateDnsName, cloud.Cvm.PrivateDnsName) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Cvm.Extension.CloudRamDiskID, cloud.Cvm.RamdiskId) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Cvm.Extension.RootDeviceName, cloud.Cvm.RootDeviceName) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Cvm.Extension.PrivateDnsName, cloud.Cvm.PrivateDnsName) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Cvm.Extension.RootDeviceType, cloud.Cvm.RootDeviceType) {
		return true
	}

	if !assert.IsPtrBoolEqual(db.Cvm.Extension.SourceDestCheck, cloud.Cvm.SourceDestCheck) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Cvm.Extension.SriovNetSupport, cloud.Cvm.SriovNetSupport) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Cvm.Extension.VirtualizationType, cloud.Cvm.VirtualizationType) {
		return true
	}

	if !assert.IsPtrInt64Equal(db.Cvm.Extension.CpuOptions.CoreCount, cloud.Cvm.CpuOptions.CoreCount) {
		return true
	}

	if !assert.IsPtrInt64Equal(db.Cvm.Extension.CpuOptions.ThreadsPerCore, cloud.Cvm.CpuOptions.ThreadsPerCore) {
		return true
	}

	if !assert.IsPtrBoolEqual(db.Cvm.Extension.PrivateDnsNameOptions.EnableResourceNameDnsAAAARecord,
		cloud.Cvm.PrivateDnsNameOptions.EnableResourceNameDnsAAAARecord) {
		return true
	}

	if !assert.IsPtrBoolEqual(db.Cvm.Extension.PrivateDnsNameOptions.EnableResourceNameDnsARecord, cloud.Cvm.PrivateDnsNameOptions.EnableResourceNameDnsARecord) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Cvm.Extension.PrivateDnsNameOptions.HostnameType, cloud.Cvm.PrivateDnsNameOptions.HostnameType) {
		return true
	}

	sgIDs := make([]string, 0)
	if len(cloud.Cvm.SecurityGroups) > 0 {
		for _, sg := range cloud.Cvm.SecurityGroups {
			if sg.GroupId != nil {
				sgIDs = append(sgIDs, *sg.GroupId)
			}
		}
	}
	if !assert.IsStringSliceEqual(db.Cvm.Extension.CloudSecurityGroupIDs, sgIDs) {
		return true
	}

	for _, dbValue := range db.Cvm.Extension.BlockDeviceMapping {
		isEqual := false
		for _, cloudValue := range cloud.Cvm.BlockDeviceMappings {
			if dbValue.CloudVolumeID == cloudValue.Ebs.VolumeId && dbValue.Status == cloudValue.Ebs.Status {
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

func syncAwsCvmUpdate(kt *kit.Kit, updateIDs []string, cloudMap map[string]*AwsCvmSync,
	dsMap map[string]*AwsDSCvmSync, dataCli *dataservice.Client) error {

	lists := make([]dataproto.CvmBatchUpdate[corecvm.AwsCvmExtension], 0)

	for _, id := range updateIDs {
		if !isChangeAws(cloudMap[id], dsMap[id]) {
			continue
		}

		cloudVpcIDs := make([]string, 0)
		if cloudMap[id].Cvm.VpcId != nil {
			cloudVpcIDs = append(cloudVpcIDs, *cloudMap[id].Cvm.VpcId)
		}

		cloudSubnetIDs := make([]string, 0)
		if cloudMap[id].Cvm.SubnetId != nil {
			cloudSubnetIDs = append(cloudSubnetIDs, *cloudMap[id].Cvm.SubnetId)
		}

		if len(cloudVpcIDs) <= 0 {
			return fmt.Errorf("aws cvm: %s no vpc", *cloudMap[id].Cvm.InstanceId)
		}

		if len(cloudVpcIDs) > 1 {
			logs.Errorf("aws cvm: %s more than one vpc", cloudMap[id].Cvm.InstanceId)
		}

		vpcID, bkCloudID, err := queryVpcIDByCloudID(kt, dataCli, cloudVpcIDs[0])
		if err != nil {
			return err
		}

		subnetIDs, err := querySubnetIDsByCloudID(kt, dataCli, cloudSubnetIDs)
		if err != nil {
			return err
		}

		privateIPv4Addresses := make([]string, 0)
		if cloudMap[id].Cvm.PrivateIpAddress != nil {
			privateIPv4Addresses = append(privateIPv4Addresses, *cloudMap[id].Cvm.PrivateIpAddress)
		}
		publicIPv4Addresses := make([]string, 0)
		if cloudMap[id].Cvm.PublicIpAddress != nil {
			publicIPv4Addresses = append(publicIPv4Addresses, *cloudMap[id].Cvm.PublicIpAddress)
		}
		publicIPv6Addresses := make([]string, 0)
		if cloudMap[id].Cvm.Ipv6Address != nil {
			publicIPv6Addresses = append(publicIPv6Addresses, *cloudMap[id].Cvm.Ipv6Address)
		}

		sgIDs := make([]string, 0)
		if len(cloudMap[id].Cvm.SecurityGroups) > 0 {
			for _, sg := range cloudMap[id].Cvm.SecurityGroups {
				if sg.GroupId != nil {
					sgIDs = append(sgIDs, *sg.GroupId)
				}
			}
		}

		awsBlockDeviceMapping := make([]corecvm.AwsBlockDeviceMapping, 0)
		if len(cloudMap[id].Cvm.BlockDeviceMappings) > 0 {
			for _, v := range cloudMap[id].Cvm.BlockDeviceMappings {
				if v != nil {
					tmp := corecvm.AwsBlockDeviceMapping{
						Status:        v.Ebs.Status,
						CloudVolumeID: v.Ebs.VolumeId,
					}
					awsBlockDeviceMapping = append(awsBlockDeviceMapping, tmp)
				}
			}
		}

		cvm := dataproto.CvmBatchUpdate[corecvm.AwsCvmExtension]{
			ID:             dsMap[id].Cvm.ID,
			Name:           converter.PtrToVal(aws.GetCvmNameFromTags(cloudMap[id].Cvm.Tags)),
			BkCloudID:      bkCloudID,
			CloudVpcIDs:    cloudVpcIDs,
			VpcIDs:         []string{vpcID},
			CloudSubnetIDs: cloudSubnetIDs,
			SubnetIDs:      subnetIDs,
			// 云上不支持该字段
			Memo:                 nil,
			Status:               converter.PtrToVal(cloudMap[id].Cvm.State.Name),
			PrivateIPv4Addresses: privateIPv4Addresses,
			// 云上不支持该字段
			PrivateIPv6Addresses: nil,
			PublicIPv4Addresses:  publicIPv4Addresses,
			PublicIPv6Addresses:  publicIPv6Addresses,
			CloudLaunchedTime:    cloudMap[id].Cvm.LaunchTime.String(),
			// 云上不支持该字段
			CloudExpiredTime: "",
			Extension: &corecvm.AwsCvmExtension{
				CpuOptions: &corecvm.AwsCpuOptions{
					CoreCount:      cloudMap[id].Cvm.CpuOptions.CoreCount,
					ThreadsPerCore: cloudMap[id].Cvm.CpuOptions.ThreadsPerCore,
				},
				Platform:              cloudMap[id].Cvm.Platform,
				DnsName:               cloudMap[id].Cvm.PublicDnsName,
				EbsOptimized:          cloudMap[id].Cvm.EbsOptimized,
				CloudSecurityGroupIDs: sgIDs,
				PrivateDnsName:        cloudMap[id].Cvm.PrivateDnsName,
				PrivateDnsNameOptions: &corecvm.AwsPrivateDnsNameOptions{
					EnableResourceNameDnsAAAARecord: cloudMap[id].Cvm.PrivateDnsNameOptions.EnableResourceNameDnsAAAARecord,
					EnableResourceNameDnsARecord:    cloudMap[id].Cvm.PrivateDnsNameOptions.EnableResourceNameDnsARecord,
					HostnameType:                    cloudMap[id].Cvm.PrivateDnsNameOptions.HostnameType,
				},
				CloudRamDiskID:     cloudMap[id].Cvm.RamdiskId,
				RootDeviceName:     cloudMap[id].Cvm.RootDeviceName,
				RootDeviceType:     cloudMap[id].Cvm.RootDeviceType,
				SourceDestCheck:    cloudMap[id].Cvm.SourceDestCheck,
				SriovNetSupport:    cloudMap[id].Cvm.SriovNetSupport,
				VirtualizationType: cloudMap[id].Cvm.VirtualizationType,
				BlockDeviceMapping: awsBlockDeviceMapping,
			},
		}

		lists = append(lists, cvm)
	}

	updateReq := dataproto.CvmBatchUpdateReq[corecvm.AwsCvmExtension]{
		Cvms: lists,
	}

	if len(updateReq.Cvms) > 0 {
		if err := dataCli.Aws.Cvm.BatchUpdateCvm(kt.Ctx, kt.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateCvm failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func syncAwsCvmAdd(kt *kit.Kit, addIDs []string, req *SyncAwsCvmOption,
	cloudMap map[string]*AwsCvmSync, dataCli *dataservice.Client) error {

	lists := make([]dataproto.CvmBatchCreate[corecvm.AwsCvmExtension], 0)

	for _, id := range addIDs {
		cloudVpcIDs := make([]string, 0)
		if cloudMap[id].Cvm.VpcId != nil {
			cloudVpcIDs = append(cloudVpcIDs, *cloudMap[id].Cvm.VpcId)
		}

		cloudSubnetIDs := make([]string, 0)
		if cloudMap[id].Cvm.SubnetId != nil {
			cloudSubnetIDs = append(cloudSubnetIDs, *cloudMap[id].Cvm.SubnetId)
		}

		if len(cloudVpcIDs) <= 0 {
			return fmt.Errorf("aws cvm: %s no vpc", *cloudMap[id].Cvm.InstanceId)
		}

		if len(cloudVpcIDs) > 1 {
			logs.Errorf("aws cvm: %s more than one vpc", cloudMap[id].Cvm.InstanceId)
		}

		vpcID, bkCloudID, err := queryVpcIDByCloudID(kt, dataCli, cloudVpcIDs[0])
		if err != nil {
			return err
		}

		subnetIDs, err := querySubnetIDsByCloudID(kt, dataCli, cloudSubnetIDs)
		if err != nil {
			return err
		}

		privateIPv4Addresses := make([]string, 0)
		if cloudMap[id].Cvm.PrivateIpAddress != nil {
			privateIPv4Addresses = append(privateIPv4Addresses, *cloudMap[id].Cvm.PrivateIpAddress)
		}
		publicIPv4Addresses := make([]string, 0)
		if cloudMap[id].Cvm.PublicIpAddress != nil {
			publicIPv4Addresses = append(publicIPv4Addresses, *cloudMap[id].Cvm.PublicIpAddress)
		}
		publicIPv6Addresses := make([]string, 0)
		if cloudMap[id].Cvm.Ipv6Address != nil {
			publicIPv6Addresses = append(publicIPv6Addresses, *cloudMap[id].Cvm.Ipv6Address)
		}

		sgIDs := make([]string, 0)
		if len(cloudMap[id].Cvm.SecurityGroups) > 0 {
			for _, sg := range cloudMap[id].Cvm.SecurityGroups {
				if sg.GroupId != nil {
					sgIDs = append(sgIDs, *sg.GroupId)
				}
			}
		}

		awsBlockDeviceMapping := make([]corecvm.AwsBlockDeviceMapping, 0)
		if len(cloudMap[id].Cvm.BlockDeviceMappings) > 0 {
			for _, v := range cloudMap[id].Cvm.BlockDeviceMappings {
				if v != nil {
					tmp := corecvm.AwsBlockDeviceMapping{
						Status:        v.Ebs.Status,
						CloudVolumeID: v.Ebs.VolumeId,
					}
					awsBlockDeviceMapping = append(awsBlockDeviceMapping, tmp)
				}
			}
		}

		cvm := dataproto.CvmBatchCreate[corecvm.AwsCvmExtension]{
			CloudID:        converter.PtrToVal(cloudMap[id].Cvm.InstanceId),
			Name:           converter.PtrToVal(aws.GetCvmNameFromTags(cloudMap[id].Cvm.Tags)),
			BkBizID:        constant.UnassignedBiz,
			BkCloudID:      bkCloudID,
			AccountID:      req.AccountID,
			Region:         req.Region,
			Zone:           converter.PtrToVal(cloudMap[id].Cvm.Placement.AvailabilityZone),
			CloudVpcIDs:    cloudVpcIDs,
			VpcIDs:         []string{vpcID},
			CloudSubnetIDs: cloudSubnetIDs,
			SubnetIDs:      subnetIDs,
			CloudImageID:   converter.PtrToVal(cloudMap[id].Cvm.ImageId),
			OsName:         *cloudMap[id].Cvm.PlatformDetails,
			// 云上不支持该字段
			Memo:                 nil,
			Status:               converter.PtrToVal(cloudMap[id].Cvm.State.Name),
			PrivateIPv4Addresses: privateIPv4Addresses,
			// 云上不支持该字段
			PrivateIPv6Addresses: nil,
			PublicIPv4Addresses:  publicIPv4Addresses,
			PublicIPv6Addresses:  publicIPv6Addresses,
			MachineType:          converter.PtrToVal(cloudMap[id].Cvm.InstanceType),
			// 云上不支持该字段
			CloudCreatedTime:  "",
			CloudLaunchedTime: cloudMap[id].Cvm.LaunchTime.String(),
			// 云上不支持该字段
			CloudExpiredTime: "",
			Extension: &corecvm.AwsCvmExtension{
				CpuOptions: &corecvm.AwsCpuOptions{
					CoreCount:      cloudMap[id].Cvm.CpuOptions.CoreCount,
					ThreadsPerCore: cloudMap[id].Cvm.CpuOptions.ThreadsPerCore,
				},
				Platform:              cloudMap[id].Cvm.Platform,
				DnsName:               cloudMap[id].Cvm.PublicDnsName,
				EbsOptimized:          cloudMap[id].Cvm.EbsOptimized,
				CloudSecurityGroupIDs: sgIDs,
				PrivateDnsName:        cloudMap[id].Cvm.PrivateDnsName,
				PrivateDnsNameOptions: &corecvm.AwsPrivateDnsNameOptions{
					EnableResourceNameDnsAAAARecord: cloudMap[id].Cvm.PrivateDnsNameOptions.EnableResourceNameDnsAAAARecord,
					EnableResourceNameDnsARecord:    cloudMap[id].Cvm.PrivateDnsNameOptions.EnableResourceNameDnsARecord,
					HostnameType:                    cloudMap[id].Cvm.PrivateDnsNameOptions.HostnameType,
				},
				CloudRamDiskID:     cloudMap[id].Cvm.RamdiskId,
				RootDeviceName:     cloudMap[id].Cvm.RootDeviceName,
				RootDeviceType:     cloudMap[id].Cvm.RootDeviceType,
				SourceDestCheck:    cloudMap[id].Cvm.SourceDestCheck,
				SriovNetSupport:    cloudMap[id].Cvm.SriovNetSupport,
				VirtualizationType: cloudMap[id].Cvm.VirtualizationType,
				BlockDeviceMapping: awsBlockDeviceMapping,
			},
		}

		lists = append(lists, cvm)
	}

	createReq := dataproto.CvmBatchCreateReq[corecvm.AwsCvmExtension]{
		Cvms: lists,
	}

	if len(createReq.Cvms) > 0 {
		_, err := dataCli.Aws.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(), &createReq)
		if err != nil {
			logs.Errorf("request dataservice to create aws cvm failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func getAwsCvmDSSync(kt *kit.Kit, cloudIDs []string, req *SyncAwsCvmOption,
	dataCli *dataservice.Client) ([]string, map[string]*AwsDSCvmSync, error) {

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*AwsDSCvmSync)

	start := 0
	for {
		dataReq := &dataproto.CvmListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "vendor",
						Op:    filter.Equal.Factory(),
						Value: enumor.Aws,
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

		results, err := dataCli.Aws.Cvm.ListCvmExt(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list cvm failed, err: %v, rid: %s", err, kt.Rid)
			return updateIDs, dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				updateIDs = append(updateIDs, detail.CloudID)
				dsImageSync := new(AwsDSCvmSync)
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

func getAwsCvmAllDSByVendor(kt *kit.Kit, req *SyncAwsCvmOption,
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

		results, err := dataCli.Aws.Cvm.ListCvmExt(kt.Ctx, kt.Header(), dataReq)
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

// SyncAwsCvmWithRelResource sync all cvm rel resource
func SyncAwsCvmWithRelResource(kt *kit.Kit, req *SyncAwsCvmOption,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	client, err := ad.Aws(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	cloudVpcMap, cloudSubnetMap, cloudEipMap, cloudSGMap, cloudDiskMap, err := getAwsCVMRelResourcesIDs(kt, req, client)
	if err != nil {
		logs.Errorf("request get aws cvm rel resource ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(cloudVpcMap) > 0 {
		vpcCloudIDs := make([]string, 0)
		for _, id := range cloudVpcMap {
			vpcCloudIDs = append(vpcCloudIDs, id.RelID)
		}
		vpcOpt := &vpc.SyncAwsOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  vpcCloudIDs,
		}
		_, err := vpc.AwsVpcSync(kt, ad, dataCli, vpcOpt)
		if err != nil {
			logs.Errorf("request to sync aws vpc logic failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudSubnetMap) > 0 {
		subnetCloudIDs := make([]string, 0)
		for _, id := range cloudSubnetMap {
			subnetCloudIDs = append(subnetCloudIDs, id.RelID)
		}
		req := &subnet.SyncAwsOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  subnetCloudIDs,
		}
		_, err := subnet.AwsSubnetSync(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("request to sync aws subnet logic failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudEipMap) > 0 {
		eipCloudIDs := make([]string, 0)
		for _, id := range cloudEipMap {
			eipCloudIDs = append(eipCloudIDs, id.RelID)
		}
		req := &synceip.SyncAwsEipOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  eipCloudIDs,
		}
		_, err := synceip.SyncAwsEip(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("sync aws cvm rel eip failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudSGMap) > 0 {
		sGCloudIDs := make([]string, 0)
		for _, id := range cloudSGMap {
			sGCloudIDs = append(sGCloudIDs, id.RelID)
		}
		req := &securitygroup.SyncAwsSecurityGroupOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  sGCloudIDs,
		}
		_, err := securitygroup.SyncAwsSecurityGroup(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("sync aws cvm rel security group failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudDiskMap) > 0 {
		diskCloudIDs := make([]string, 0)
		for _, id := range cloudDiskMap {
			diskCloudIDs = append(diskCloudIDs, id.RelID)
		}
		req := &disk.SyncAwsDiskOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  diskCloudIDs,
		}
		_, err := disk.SyncAwsDisk(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("sync aws cvm rel disk failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	cvmReq := &SyncAwsCvmOption{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  req.CloudIDs,
	}
	_, err = SyncAwsCvm(kt, ad, dataCli, cvmReq)
	if err != nil {
		logs.Errorf("sync aws cvm self failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	hcReq := &protocvm.OperateSyncReq{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  req.CloudIDs,
	}

	err = getSGHcIDs(kt, hcReq, dataCli, cloudSGMap)
	if err != nil {
		logs.Errorf("request get cvm sg rel resource hc ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
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

	if len(cloudEipMap) > 0 {
		err := SyncCvmEipRel(kt, cloudEipMap, dataCli)
		if err != nil {
			logs.Errorf("sync aws cvm eip rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudSGMap) > 0 {
		err := SyncCvmSGRel(kt, cloudSGMap, dataCli)
		if err != nil {
			logs.Errorf("sync aws cvm sg rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudDiskMap) > 0 {
		err := SyncCvmDiskRel(kt, cloudDiskMap, dataCli)
		if err != nil {
			logs.Errorf("sync aws cvm disk rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func getAwsCVMRelResourcesIDs(kt *kit.Kit, req *SyncAwsCvmOption,
	client *aws.Aws) (map[string]*CVMOperateSync, map[string]*CVMOperateSync,
	map[string]*CVMOperateSync, map[string]*CVMOperateSync, map[string]*CVMOperateSync, error) {

	vpcMap := make(map[string]*CVMOperateSync)
	subnetMap := make(map[string]*CVMOperateSync)
	eipMap := make(map[string]*CVMOperateSync)
	sGMap := make(map[string]*CVMOperateSync)
	diskMap := make(map[string]*CVMOperateSync)
	eipIps := make([]string, 0)

	opt := &typecvm.AwsListOption{
		Region:   req.Region,
		CloudIDs: req.CloudIDs,
	}

	datas, err := client.ListCvm(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to list aws cvm failed, err: %v, rid: %s", err, kt.Rid)
		return vpcMap, subnetMap, eipMap, sGMap, diskMap, err
	}

	for _, reservation := range datas.Reservations {
		for _, instance := range reservation.Instances {
			if len(instance.SecurityGroups) > 0 {
				for _, sg := range instance.SecurityGroups {
					if sg != nil {
						id := getCVMRelID(*sg.GroupId, *instance.InstanceId)
						sGMap[id] = &CVMOperateSync{RelID: *sg.GroupId, InstanceID: *instance.InstanceId}
					}
				}
			}

			if instance.VpcId != nil {
				id := getCVMRelID(*instance.VpcId, *instance.InstanceId)
				vpcMap[id] = &CVMOperateSync{RelID: *instance.VpcId, InstanceID: *instance.InstanceId}
			}

			if instance.SubnetId != nil {
				id := getCVMRelID(*instance.SubnetId, *instance.InstanceId)
				subnetMap[id] = &CVMOperateSync{RelID: *instance.SubnetId, InstanceID: *instance.InstanceId}
			}

			if len(instance.BlockDeviceMappings) > 0 {
				for _, disk := range instance.BlockDeviceMappings {
					if disk.Ebs != nil {
						id := getCVMRelID(*disk.Ebs.VolumeId, *instance.InstanceId)
						diskMap[id] = &CVMOperateSync{RelID: *disk.Ebs.VolumeId, InstanceID: *instance.InstanceId}
					}
				}
			}

			if instance.PublicIpAddress != nil {
				eipIps = append(eipIps, *instance.PublicIpAddress)
			}
		}
	}

	if len(eipIps) > 0 {
		opt := &eip.AwsEipListOption{
			Region: req.Region,
			Ips:    eipIps,
		}

		eips, err := client.ListEip(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor to list aws eip failed, err: %v, rid: %s", err, kt.Rid)
		}

		for _, eip := range eips.Details {
			id := getCVMRelID(eip.CloudID, *eip.InstanceId)
			eipMap[id] = &CVMOperateSync{RelID: eip.CloudID, InstanceID: *eip.InstanceId}
		}
	}

	return vpcMap, subnetMap, eipMap, sGMap, diskMap, err
}

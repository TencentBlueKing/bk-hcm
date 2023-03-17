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
	"errors"
	"fmt"

	"hcm/cmd/hc-service/logics/sync/disk"
	synceip "hcm/cmd/hc-service/logics/sync/eip"
	securitygroup "hcm/cmd/hc-service/logics/sync/security-group"
	"hcm/cmd/hc-service/logics/sync/subnet"
	"hcm/cmd/hc-service/logics/sync/vpc"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/adaptor/tcloud"
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
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"

	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

// SyncTCloudCvmOption define sync tcloud cvm option.
type SyncTCloudCvmOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"required"`
}

// Validate SyncTCloudCvmOption
func (opt SyncTCloudCvmOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) == 0 {
		return errors.New("cloudIDs is required")
	}

	if len(opt.CloudIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// SyncTCloudCvm ...
func SyncTCloudCvm(kt *kit.Kit, ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client,
	opt *SyncTCloudCvmOption) error {

	client, err := ad.TCloud(kt, opt.AccountID)
	if err != nil {
		return err
	}

	// 查询Cvm数据从云上
	listOpt := &typecvm.TCloudListOption{
		Region:   opt.Region,
		CloudIDs: opt.CloudIDs,
	}
	cvmsFromCloud, err := client.ListCvm(kt, listOpt)
	if err != nil {
		logs.Errorf("request adaptor to list tcloud cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if len(cvmsFromCloud) == 0 {
		return nil
	}

	// 查询Cvm数据从db
	cloudIDCvmMapFromDB, err := listTCloudCvmFromDB(kt, dataCli, opt)
	if err != nil {
		logs.Errorf("request listTCloudCvmFromDB failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cloudIDCvmMapFromCloudTmp := make(map[string]*cvm.Instance, len(cvmsFromCloud))
	cloudIDCvmMapFromCloud := make(map[string]*cvm.Instance, len(cvmsFromCloud))
	for _, data := range cvmsFromCloud {
		cloudIDCvmMapFromCloud[*data.InstanceId] = data
		cloudIDCvmMapFromCloudTmp[*data.InstanceId] = data
	}

	// 对更新、新增、删除主机进行分类
	delCloudIDs := make([]string, 0)
	updateCloudIDs := make([]string, 0)
	for cloudID, cvmFromDB := range cloudIDCvmMapFromDB {
		cvmFromCloud, exist := cloudIDCvmMapFromCloudTmp[cloudID]
		if !exist {
			delCloudIDs = append(delCloudIDs, cloudID)
		}

		if isTCloudCvmChange(cvmFromCloud, cvmFromDB) {
			updateCloudIDs = append(updateCloudIDs, cloudID)
			delete(cloudIDCvmMapFromCloudTmp, cloudID)
		}
	}

	addCloudIDs := make([]string, 0, len(cloudIDCvmMapFromCloudTmp))
	for cloudID := range cloudIDCvmMapFromCloudTmp {
		addCloudIDs = append(addCloudIDs, cloudID)
	}

	if len(updateCloudIDs) > 0 {
		if err := updateTCloudCvm(kt, dataCli, client, updateCloudIDs, cloudIDCvmMapFromCloud,
			cloudIDCvmMapFromDB); err != nil {

			logs.Errorf("request updateTCloudCvm failed, err: %v, cloudIDs: %v, rid: %s", err, updateCloudIDs, kt.Rid)
			return err
		}
	}

	if len(addCloudIDs) > 0 {
		if err := addTCloudCvm(kt, dataCli, client, addCloudIDs, cloudIDCvmMapFromCloud, opt); err != nil {
			logs.Errorf("request addTCloudCvm failed, err: %v, cloudIDs: %v, rid: %s", err, addCloudIDs, kt.Rid)
			return err
		}
	}

	if len(delCloudIDs) > 0 {
		if err := deleteTCloudCvm(kt, dataCli, client, delCloudIDs, opt); err != nil {
			logs.Errorf("request deleteTCloudCvm failed, err: %v, cloudIDs: %v, rid: %s", err, delCloudIDs, kt.Rid)
			return err
		}
	}

	return nil
}

// isTCloudCvmChange 判断cvm是否有字段不一致
func isTCloudCvmChange(cloud *cvm.Instance, db corecvm.Cvm[corecvm.TCloudCvmExtension]) bool {

	if db.CloudID != *cloud.InstanceId {
		return true
	}

	if db.Name != *cloud.InstanceName {
		return true
	}

	if len(db.CloudVpcIDs) == 0 || (db.CloudVpcIDs[0] != *cloud.VirtualPrivateCloud.VpcId) {
		return true
	}

	if len(db.CloudSubnetIDs) == 0 || (db.CloudSubnetIDs[0] != *cloud.VirtualPrivateCloud.SubnetId) {
		return true
	}

	if db.CloudImageID != *cloud.ImageId {
		return true
	}

	if db.OsName != *cloud.OsName {
		return true
	}

	if db.Status != *cloud.InstanceState {
		return true
	}

	if len(cloud.IPv6Addresses) == 0 && (len(db.PublicIPv6Addresses) != 0 || len(db.PrivateIPv6Addresses) != 0) {
		return true
	}

	if len(cloud.IPv6Addresses) != 0 {
		if len(db.PublicIPv6Addresses) == 0 && len(db.PrivateIPv6Addresses) == 0 {
			return true
		}

		tmpMap := converter.StringSliceToMap(converter.PtrToSlice(cloud.IPv6Addresses))
		for _, address := range db.PublicIPv6Addresses {
			delete(tmpMap, address)
		}

		for _, address := range db.PrivateIPv6Addresses {
			delete(tmpMap, address)
		}

		if len(tmpMap) != 0 {
			return true
		}
	}

	if !assert.IsStringSliceEqual(converter.PtrToSlice(cloud.PrivateIpAddresses), db.PrivateIPv4Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(converter.PtrToSlice(cloud.PublicIpAddresses), db.PublicIPv4Addresses) {
		return true
	}

	if db.MachineType != *cloud.InstanceType {
		return true
	}

	if db.CloudCreatedTime != *cloud.CreatedTime {
		return true
	}

	if db.CloudExpiredTime != converter.PtrToVal(cloud.ExpiredTime) {
		return true
	}

	if db.Extension.Placement.CloudProjectID != cloud.Placement.ProjectId {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.InstanceChargeType, cloud.InstanceChargeType) {
		return true
	}

	if !assert.IsPtrInt64Equal(db.Extension.Cpu, cloud.CPU) {
		return true
	}

	if !assert.IsPtrInt64Equal(db.Extension.Memory, cloud.Memory) {
		return true
	}

	if !assert.IsPtrStringEqual(cloud.SystemDisk.DiskId, db.Extension.CloudSystemDiskID) {
		return true
	}

	dataDiskIds := make([]string, 0, len(cloud.DataDisks))
	for _, one := range cloud.DataDisks {
		dataDiskIds = append(dataDiskIds, *one.DiskId)
	}
	if !assert.IsStringSliceEqual(dataDiskIds, db.Extension.CloudDataDiskIDs) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.InternetAccessible.InternetChargeType,
		cloud.InternetAccessible.InternetChargeType) {
		return true
	}

	if !assert.IsPtrInt64Equal(db.Extension.InternetAccessible.InternetMaxBandwidthOut,
		cloud.InternetAccessible.InternetMaxBandwidthOut) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.InternetAccessible.CloudBandwidthPackageID,
		cloud.InternetAccessible.BandwidthPackageId) {
		return true
	}

	if !assert.IsPtrBoolEqual(db.Extension.InternetAccessible.PublicIPAssigned,
		cloud.InternetAccessible.PublicIpAssigned) {
		return true
	}

	if !assert.IsPtrBoolEqual(db.Extension.VirtualPrivateCloud.AsVpcGateway, cloud.VirtualPrivateCloud.AsVpcGateway) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.RenewFlag, cloud.RenewFlag) {
		return true
	}

	if !assert.IsStringSliceEqual(db.Extension.CloudSecurityGroupIDs, converter.PtrToSlice(cloud.SecurityGroupIds)) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.StopChargingMode, cloud.StopChargingMode) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.UUID, cloud.Uuid) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Extension.IsolatedSource, cloud.IsolatedSource) {
		return true
	}

	if !assert.IsPtrBoolEqual(db.Extension.DisableApiTermination, cloud.DisableApiTermination) {
		return true
	}

	return true
}

func deleteTCloudCvm(kt *kit.Kit, dataCli *dataservice.Client, tcloudCli *tcloud.TCloud,
	delCloudID []string, opt *SyncTCloudCvmOption) error {

	listOpt := &typecvm.TCloudListOption{
		Region:   opt.Region,
		CloudIDs: delCloudID,
	}
	cvmsFromCloud, err := tcloudCli.ListCvm(kt, listOpt)
	if err != nil {
		return err
	}

	if len(cvmsFromCloud) > 0 {
		existCloudIDs := make([]string, len(cvmsFromCloud))
		for _, cvm := range cvmsFromCloud {
			existCloudIDs = append(existCloudIDs, *cvm.InstanceId)
		}

		return fmt.Errorf("cvm(cloudIDs=%v) exist in tcloud, can not delete", existCloudIDs)
	}

	batchDeleteReq := &dataproto.CvmBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudID),
	}

	if err := dataCli.Global.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), batchDeleteReq); err != nil {
		logs.Errorf("request dataservice delete tcloud cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func updateTCloudCvm(kt *kit.Kit, dataCli *dataservice.Client, tcloud *tcloud.TCloud, updateCloudID []string,
	cloudIDCvmMapFromCloud map[string]*cvm.Instance,
	cloudIDCvmMapFromDB map[string]corecvm.Cvm[corecvm.TCloudCvmExtension]) error {

	lists := make([]dataproto.CvmBatchUpdate[corecvm.TCloudCvmExtension], 0, len(updateCloudID))
	for _, cloudID := range updateCloudID {
		cvmFromCloud, exist := cloudIDCvmMapFromCloud[cloudID]
		if !exist {
			return fmt.Errorf("cvm: %s not found from cloud", cloudID)
		}

		cvmDB, exist := cloudIDCvmMapFromDB[cloudID]
		if !exist {
			return fmt.Errorf("cvm: %s not found from db", cloudID)
		}

		vpcID, bkCloudID, err := queryVpcIDByCloudID(kt, dataCli, *cvmFromCloud.VirtualPrivateCloud.VpcId)
		if err != nil {
			return err
		}

		cloudSubnetIDs := make([]string, 0)
		cloudSubnetIDs = append(cloudSubnetIDs, *cvmFromCloud.VirtualPrivateCloud.SubnetId)
		subnetIDs, err := querySubnetIDsByCloudID(kt, dataCli, cloudSubnetIDs)
		if err != nil {
			return err
		}

		dataDiskIDs := make([]string, 0)
		for _, disk := range cvmFromCloud.DataDisks {
			dataDiskIDs = append(dataDiskIDs, *disk.DiskId)
		}

		one := dataproto.CvmBatchUpdate[corecvm.TCloudCvmExtension]{
			ID:             cvmDB.ID,
			Name:           converter.PtrToVal(cvmFromCloud.InstanceName),
			BkCloudID:      bkCloudID,
			CloudVpcIDs:    []string{converter.PtrToVal(cvmFromCloud.VirtualPrivateCloud.VpcId)},
			VpcIDs:         []string{vpcID},
			CloudSubnetIDs: []string{converter.PtrToVal(cvmFromCloud.VirtualPrivateCloud.SubnetId)},
			SubnetIDs:      subnetIDs,
			// 备注字段云上没有，仅限hcm内部使用
			Memo:                 nil,
			Status:               converter.PtrToVal(cvmFromCloud.InstanceState),
			PrivateIPv4Addresses: converter.PtrToSlice(cvmFromCloud.PrivateIpAddresses),
			PublicIPv4Addresses:  converter.PtrToSlice(cvmFromCloud.PublicIpAddresses),
			// 云上该字段没有
			CloudLaunchedTime: "",
			CloudExpiredTime:  converter.PtrToVal(cvmFromCloud.ExpiredTime),
			Extension: &corecvm.TCloudCvmExtension{
				Placement: &corecvm.TCloudPlacement{
					CloudProjectID: cvmFromCloud.Placement.ProjectId,
				},
				InstanceChargeType: cvmFromCloud.InstanceChargeType,
				Cpu:                cvmFromCloud.CPU,
				Memory:             cvmFromCloud.Memory,
				CloudSystemDiskID:  cvmFromCloud.SystemDisk.DiskId,
				CloudDataDiskIDs:   dataDiskIDs,
				InternetAccessible: &corecvm.TCloudInternetAccessible{
					InternetChargeType:      cvmFromCloud.InternetAccessible.InternetChargeType,
					InternetMaxBandwidthOut: cvmFromCloud.InternetAccessible.InternetMaxBandwidthOut,
					PublicIPAssigned:        cvmFromCloud.InternetAccessible.PublicIpAssigned,
					CloudBandwidthPackageID: cvmFromCloud.InternetAccessible.BandwidthPackageId,
				},
				VirtualPrivateCloud: &corecvm.TCloudVirtualPrivateCloud{
					AsVpcGateway: cvmFromCloud.VirtualPrivateCloud.AsVpcGateway,
				},
				RenewFlag:             cvmFromCloud.RenewFlag,
				CloudSecurityGroupIDs: converter.PtrToSlice(cvmFromCloud.SecurityGroupIds),
				StopChargingMode:      cvmFromCloud.StopChargingMode,
				UUID:                  cvmFromCloud.Uuid,
				IsolatedSource:        cvmFromCloud.IsolatedSource,
				DisableApiTermination: cvmFromCloud.DisableApiTermination,
			},
		}

		if len(cvmFromCloud.IPv6Addresses) != 0 {
			cvmFromCloud.PublicIpAddresses, cvmFromCloud.PrivateIpAddresses, err = tcloud.DetermineIPv6Type(kt,
				cvmDB.Region, cvmFromCloud.IPv6Addresses)
			if err != nil {
				return err
			}
		}

		lists = append(lists, one)
	}

	updateReq := dataproto.CvmBatchUpdateReq[corecvm.TCloudCvmExtension]{
		Cvms: lists,
	}
	if len(updateReq.Cvms) > 0 {
		if err := dataCli.TCloud.Cvm.BatchUpdateCvm(kt.Ctx, kt.Header(), &updateReq); err != nil {
			logs.Errorf("request tcloud dataservice BatchUpdateCvm failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func addTCloudCvm(kt *kit.Kit, dataCli *dataservice.Client, tcloud *tcloud.TCloud, addCloudIDs []string,
	cloudIDCvmMapFromCloud map[string]*cvm.Instance, opt *SyncTCloudCvmOption) error {

	lists := make([]dataproto.CvmBatchCreate[corecvm.TCloudCvmExtension], 0)

	for _, cloudID := range addCloudIDs {
		cvmFromCloud, exist := cloudIDCvmMapFromCloud[cloudID]
		if !exist {
			return fmt.Errorf("cvm: %s not found", cloudID)
		}

		vpcID, bkCloudID, err := queryVpcIDByCloudID(kt, dataCli, *cvmFromCloud.VirtualPrivateCloud.VpcId)
		if err != nil {
			return err
		}

		cloudSubnetIDs := make([]string, 0)
		cloudSubnetIDs = append(cloudSubnetIDs, *cvmFromCloud.VirtualPrivateCloud.SubnetId)
		subnetIDs, err := querySubnetIDsByCloudID(kt, dataCli, cloudSubnetIDs)
		if err != nil {
			return err
		}

		dataDiskIDs := make([]string, 0)
		for _, disk := range cvmFromCloud.DataDisks {
			dataDiskIDs = append(dataDiskIDs, *disk.DiskId)
		}

		one := dataproto.CvmBatchCreate[corecvm.TCloudCvmExtension]{
			CloudID:        converter.PtrToVal(cvmFromCloud.InstanceId),
			Name:           converter.PtrToVal(cvmFromCloud.InstanceName),
			BkBizID:        constant.UnassignedBiz,
			BkCloudID:      bkCloudID,
			AccountID:      opt.AccountID,
			Region:         opt.Region,
			Zone:           converter.PtrToVal(cvmFromCloud.Placement.Zone),
			CloudVpcIDs:    []string{converter.PtrToVal(cvmFromCloud.VirtualPrivateCloud.VpcId)},
			VpcIDs:         []string{vpcID},
			CloudSubnetIDs: []string{converter.PtrToVal(cvmFromCloud.VirtualPrivateCloud.SubnetId)},
			SubnetIDs:      subnetIDs,
			CloudImageID:   converter.PtrToVal(cvmFromCloud.ImageId),
			OsName:         converter.PtrToVal(cvmFromCloud.OsName),
			// 备注字段云上没有，仅限hcm内部使用
			Memo:                 nil,
			Status:               converter.PtrToVal(cvmFromCloud.InstanceState),
			PrivateIPv4Addresses: converter.PtrToSlice(cvmFromCloud.PrivateIpAddresses),
			PublicIPv4Addresses:  converter.PtrToSlice(cvmFromCloud.PublicIpAddresses),
			MachineType:          *cvmFromCloud.InstanceType,
			CloudCreatedTime:     *cvmFromCloud.CreatedTime,
			// 该字段云上没有
			CloudLaunchedTime: "",
			CloudExpiredTime:  converter.PtrToVal(cvmFromCloud.ExpiredTime),
			Extension: &corecvm.TCloudCvmExtension{
				Placement: &corecvm.TCloudPlacement{
					CloudProjectID: cvmFromCloud.Placement.ProjectId,
				},
				InstanceChargeType: cvmFromCloud.InstanceChargeType,
				Cpu:                cvmFromCloud.CPU,
				Memory:             cvmFromCloud.Memory,
				CloudSystemDiskID:  cvmFromCloud.SystemDisk.DiskId,
				CloudDataDiskIDs:   dataDiskIDs,
				InternetAccessible: &corecvm.TCloudInternetAccessible{
					InternetChargeType:      cvmFromCloud.InternetAccessible.InternetChargeType,
					InternetMaxBandwidthOut: cvmFromCloud.InternetAccessible.InternetMaxBandwidthOut,
					PublicIPAssigned:        cvmFromCloud.InternetAccessible.PublicIpAssigned,
					CloudBandwidthPackageID: cvmFromCloud.InternetAccessible.BandwidthPackageId,
				},
				VirtualPrivateCloud: &corecvm.TCloudVirtualPrivateCloud{
					AsVpcGateway: cvmFromCloud.VirtualPrivateCloud.AsVpcGateway,
				},
				RenewFlag:             cvmFromCloud.RenewFlag,
				CloudSecurityGroupIDs: converter.PtrToSlice(cvmFromCloud.SecurityGroupIds),
				StopChargingMode:      cvmFromCloud.StopChargingMode,
				UUID:                  cvmFromCloud.Uuid,
				IsolatedSource:        cvmFromCloud.IsolatedSource,
				DisableApiTermination: cvmFromCloud.DisableApiTermination,
			},
		}

		if len(cvmFromCloud.IPv6Addresses) != 0 {
			cvmFromCloud.PublicIpAddresses, cvmFromCloud.PrivateIpAddresses, err = tcloud.DetermineIPv6Type(kt,
				opt.Region, cvmFromCloud.IPv6Addresses)
			if err != nil {
				return err
			}
		}

		lists = append(lists, one)
	}

	createReq := dataproto.CvmBatchCreateReq[corecvm.TCloudCvmExtension]{
		Cvms: lists,
	}

	_, err := dataCli.TCloud.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(), &createReq)
	if err != nil {
		logs.Errorf("request dataservice to create tcloud cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func listTCloudCvmFromDB(kt *kit.Kit, dataCli *dataservice.Client, opt *SyncTCloudCvmOption) (
	map[string]corecvm.Cvm[corecvm.TCloudCvmExtension], error) {

	dataReq := &dataproto.CvmListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: enumor.TCloud,
				},
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: opt.AccountID,
				},
				&filter.AtomRule{
					Field: "region",
					Op:    filter.Equal.Factory(),
					Value: opt.Region,
				},
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: opt.CloudIDs,
				},
			},
		},
		Page: core.DefaultBasePage,
	}

	results, err := dataCli.TCloud.Cvm.ListCvmExt(kt.Ctx, kt.Header(), dataReq)
	if err != nil {
		logs.Errorf("from data-service list tcloud cvm failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	cloudIDCvmMap := make(map[string]corecvm.Cvm[corecvm.TCloudCvmExtension])
	for _, detail := range results.Details {
		cloudIDCvmMap[detail.CloudID] = detail
	}

	return cloudIDCvmMap, nil
}

// SyncTCloudCvmWithRelResource sync cvm all rel resource
func SyncTCloudCvmWithRelResource(kt *kit.Kit, ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client,
	option *SyncTCloudCvmOption) (interface{}, error) {

	client, err := ad.TCloud(kt, option.AccountID)
	if err != nil {
		return nil, err
	}

	cloudSGMap, cloudVpcMap, cloudSubnetMap, cloudDiskMap, cloudEipMap, err := getTCloudCVMRelResourcesIDs(kt,
		option, client)
	if err != nil {
		logs.Errorf("request get tcloud cvm rel resource ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(cloudVpcMap) > 0 {
		vpcCloudIDs := make([]string, 0)
		for _, id := range cloudVpcMap {
			vpcCloudIDs = append(vpcCloudIDs, id.RelID)
		}
		req := &vpc.SyncTCloudOption{
			AccountID: option.AccountID,
			Region:    option.Region,
			CloudIDs:  vpcCloudIDs,
		}
		_, err := vpc.TCloudVpcSync(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("request to sync tcloud vpc logic failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudSubnetMap) > 0 {
		subnetCloudIDs := make([]string, 0)
		for _, id := range cloudSubnetMap {
			subnetCloudIDs = append(subnetCloudIDs, id.RelID)
		}
		req := &subnet.SyncTCloudOption{
			AccountID: option.AccountID,
			Region:    option.Region,
			CloudIDs:  subnetCloudIDs,
		}
		_, err := subnet.TCloudSubnetSync(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("request to sync tcloud subnet logic failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudSGMap) > 0 {
		sGCloudIDs := make([]string, 0)
		for _, id := range cloudSGMap {
			sGCloudIDs = append(sGCloudIDs, id.RelID)
		}
		req := &securitygroup.SyncTCloudSecurityGroupOption{
			AccountID: option.AccountID,
			Region:    option.Region,
			CloudIDs:  sGCloudIDs,
		}
		_, err := securitygroup.SyncTCloudSecurityGroup(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("sync tcloud cvm rel security group failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudDiskMap) > 0 {
		diskCloudIDs := make([]string, 0)
		for _, id := range cloudDiskMap {
			diskCloudIDs = append(diskCloudIDs, id.RelID)
		}
		req := &disk.SyncTCloudDiskOption{
			AccountID: option.AccountID,
			Region:    option.Region,
			CloudIDs:  diskCloudIDs,
		}
		_, err := disk.SyncTCloudDisk(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("sync tcloud cvm rel disk failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudEipMap) > 0 {
		eipCloudIDs := make([]string, 0)
		for _, id := range cloudEipMap {
			eipCloudIDs = append(eipCloudIDs, id.RelID)
		}
		req := &synceip.SyncTCloudEipOption{
			AccountID: option.AccountID,
			Region:    option.Region,
			CloudIDs:  eipCloudIDs,
		}
		_, err := synceip.SyncTCloudEip(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("sync tcloud cvm rel eip failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	cvmOpt := &SyncTCloudCvmOption{
		AccountID: option.AccountID,
		Region:    option.Region,
		CloudIDs:  option.CloudIDs,
	}
	if err = SyncTCloudCvm(kt, ad, dataCli, cvmOpt); err != nil {
		logs.Errorf("sync tcloud cvm self failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	hcReq := &protocvm.OperateSyncReq{
		AccountID: option.AccountID,
		Region:    option.Region,
		CloudIDs:  option.CloudIDs,
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

	if len(cloudSGMap) > 0 {
		err := SyncCvmSGRel(kt, cloudSGMap, dataCli)
		if err != nil {
			logs.Errorf("sync tcloud cvm sg rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudDiskMap) > 0 {
		err := SyncCvmDiskRel(kt, cloudDiskMap, dataCli)
		if err != nil {
			logs.Errorf("sync tcloud cvm disk rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudEipMap) > 0 {
		err := SyncCvmEipRel(kt, cloudEipMap, dataCli)
		if err != nil {
			logs.Errorf("sync tcloud cvm eip rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func getTCloudCVMRelResourcesIDs(kt *kit.Kit, option *SyncTCloudCvmOption,
	client *tcloud.TCloud) (map[string]*CVMOperateSync, map[string]*CVMOperateSync,
	map[string]*CVMOperateSync, map[string]*CVMOperateSync, map[string]*CVMOperateSync, error) {

	sGMap := make(map[string]*CVMOperateSync)
	vpcMap := make(map[string]*CVMOperateSync)
	subnetMap := make(map[string]*CVMOperateSync)
	diskMap := make(map[string]*CVMOperateSync)
	eipMap := make(map[string]*CVMOperateSync)

	opt := &typecvm.TCloudListOption{
		Region:   option.Region,
		CloudIDs: option.CloudIDs,
	}

	datas, err := client.ListCvm(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to list tcloud cvm failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, nil, nil, err
	}

	for _, data := range datas {
		if len(data.SecurityGroupIds) > 0 {
			for _, SecurityGroupId := range data.SecurityGroupIds {
				if SecurityGroupId == nil {
					continue
				}
				id := getCVMRelID(*SecurityGroupId, *data.InstanceId)
				sGMap[id] = &CVMOperateSync{RelID: *SecurityGroupId, InstanceID: *data.InstanceId}
			}
		}

		if data.VirtualPrivateCloud != nil {
			if data.VirtualPrivateCloud.VpcId != nil {
				id := getCVMRelID(*data.VirtualPrivateCloud.VpcId, *data.InstanceId)
				vpcMap[id] = &CVMOperateSync{RelID: *data.VirtualPrivateCloud.VpcId, InstanceID: *data.InstanceId}
			}

			if data.VirtualPrivateCloud.SubnetId != nil {
				id := getCVMRelID(*data.VirtualPrivateCloud.SubnetId, *data.InstanceId)
				subnetMap[id] = &CVMOperateSync{RelID: *data.VirtualPrivateCloud.SubnetId, InstanceID: *data.InstanceId}
			}
		}

		if data.SystemDisk != nil {
			if data.SystemDisk.DiskId != nil {
				id := getCVMRelID(*data.SystemDisk.DiskId, *data.InstanceId)
				diskMap[id] = &CVMOperateSync{RelID: *data.SystemDisk.DiskId, InstanceID: *data.InstanceId}
			}
		}

		if len(data.DataDisks) > 0 {
			for _, disk := range data.DataDisks {
				if disk.DiskId == nil {
					continue
				}
				id := getCVMRelID(*disk.DiskId, *data.InstanceId)
				diskMap[id] = &CVMOperateSync{RelID: *disk.DiskId, InstanceID: *data.InstanceId}
			}
		}

		if len(data.PublicIpAddresses) > 0 {
			ips := make([]string, 0)
			for _, ip := range data.PublicIpAddresses {
				if ip != nil {
					ips = append(ips, *ip)
				}
			}

			opt := &eip.TCloudEipListOption{
				Region: option.Region,
				Ips:    ips,
			}
			eips, err := client.ListEip(kt, opt)
			if err != nil {
				logs.Errorf("request adaptor to list tcloud eip failed, err: %v, rid: %s", err, kt.Rid)
				return nil, nil, nil, nil, nil, err
			}
			for _, eip := range eips.Details {
				id := getCVMRelID(eip.CloudID, *data.InstanceId)
				eipMap[id] = &CVMOperateSync{RelID: eip.CloudID, InstanceID: *data.InstanceId}
			}
		}
	}

	return sGMap, vpcMap, subnetMap, diskMap, eipMap, nil
}

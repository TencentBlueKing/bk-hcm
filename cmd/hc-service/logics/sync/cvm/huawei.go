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
	"strconv"

	"hcm/cmd/hc-service/logics/sync/disk"
	synceip "hcm/cmd/hc-service/logics/sync/eip"
	syncnetworkinterface "hcm/cmd/hc-service/logics/sync/network-interface"
	securitygroup "hcm/cmd/hc-service/logics/sync/security-group"
	"hcm/cmd/hc-service/logics/sync/subnet"
	"hcm/cmd/hc-service/logics/sync/vpc"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/adaptor/huawei"
	typecore "hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/adaptor/types/eip"
	networkinterface "hcm/pkg/adaptor/types/network-interface"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	dataproto "hcm/pkg/api/data-service/cloud"
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
	"hcm/pkg/tools/converter"
)

// SyncHuaWeiCvmOption define sync huawei cvm option.
type SyncHuaWeiCvmOption struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncHuaWeiCvmOption
func (opt SyncHuaWeiCvmOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.RelResourceOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.RelResourceOperationMaxLimit)
	}

	return nil
}

// SyncHuaWeiCvm sync cvm self
func SyncHuaWeiCvm(kt *kit.Kit, req *SyncHuaWeiCvmOption,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := ad.HuaWei(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	cloudAllIDs := make(map[string]bool)

	listOpt := &typecvm.HuaWeiListOption{
		Region: req.Region,
		Page: &typecore.HuaWeiCvmOffsetPage{
			Offset: int32(0),
			Limit:  int32(constant.BatchOperationMaxLimit),
		},
	}
	for {
		if len(req.CloudIDs) > 0 {
			listOpt.Page = nil
			listOpt.CloudIDs = req.CloudIDs
		}

		datas, err := client.ListCvm(kt, listOpt)
		if err != nil {
			logs.Errorf("request adaptor to list huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		cloudMap := make(map[string]*HuaWeiCvmSync)
		cloudIDs := make([]string, 0)
		if datas != nil {
			for _, data := range *datas {
				cvmSync := new(HuaWeiCvmSync)
				cvmSync.IsUpdate = false
				cvmSync.Cvm = data
				cloudMap[data.Id] = cvmSync
				cloudIDs = append(cloudIDs, data.Id)
				cloudAllIDs[data.Id] = true
			}
		}

		updateIDs, dsMap, err := getHuaWeiCvmDSSync(kt, cloudIDs, req, dataCli)
		if err != nil {
			logs.Errorf("request getHuaWeiCvmDSSync failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		if len(updateIDs) > 0 {
			err := syncHuaWeiCvmUpdate(kt, req, updateIDs, cloudMap, dsMap, client, dataCli)
			if err != nil {
				logs.Errorf("request syncHuaWeiCvmUpdate failed, err: %v, rid: %s", err, kt.Rid)
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
			err := syncHuaWeiCvmAdd(kt, addIDs, req, cloudMap, client, dataCli)
			if err != nil {
				logs.Errorf("request syncHuaWeiCvmAdd failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}

		if datas == nil || len(*datas) < typecore.TCloudQueryLimit {
			break
		}
		listOpt.Page.Offset += typecore.TCloudQueryLimit
	}

	dsIDs, err := getHuaWeiCvmAllDSByVendor(kt, req, enumor.HuaWei, dataCli)
	if err != nil {
		logs.Errorf("request getHuaWeiCvmAllDSByVendor failed, err: %v, rid: %s", err, kt.Rid)
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
		for {
			if len(req.CloudIDs) > 0 {
				listOpt.Page = nil
				listOpt.CloudIDs = req.CloudIDs
			}

			datas, err := client.ListCvm(kt, listOpt)
			if err != nil {
				logs.Errorf("request adaptor to list huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}

			for _, id := range deleteIDs {
				realDeleteFlag := true
				if datas != nil {
					for _, data := range *datas {
						if data.Id == id {
							realDeleteFlag = false
							break
						}
					}
				}

				if realDeleteFlag {
					realDeleteIDs = append(realDeleteIDs, id)
				}
			}

			if datas == nil || len(*datas) < typecore.TCloudQueryLimit {
				break
			}
			listOpt.Page.Offset += typecore.TCloudQueryLimit
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

func isChangeHuaWei(cloud *HuaWeiCvmSync, db *HuaWeiDSCvmSync, kt *kit.Kit,
	req *SyncHuaWeiCvmOption, client *huawei.HuaWei) bool {
	if db.Cvm.CloudID != cloud.Cvm.Id {
		return true
	}

	if db.Cvm.Name != cloud.Cvm.Name {
		return true
	}

	if db.Cvm.CloudImageID != cloud.Cvm.Image.Id {
		return true
	}

	if db.Cvm.OsName != cloud.Cvm.OSEXTSRVATTRhost {
		return true
	}

	if db.Cvm.Status != cloud.Cvm.Status {
		return true
	}

	opt := &networkinterface.HuaWeiNIListOption{
		Region:   req.Region,
		ServerID: cloud.Cvm.Id,
	}
	netInterDatas, err := client.ListNetworkInterface(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to list huawei network interface failed, err: %v, rid: %s", err, kt.Rid)
		return false
	}

	cloudVpcIDs := make([]string, 0)
	cloudSubnetIDs := make([]string, 0)
	for _, v := range netInterDatas.Details {
		if v.CloudVpcID != nil {
			cloudVpcIDs = append(cloudVpcIDs, *v.CloudVpcID)
		}
		if v.CloudSubnetID != nil {
			cloudSubnetIDs = append(cloudSubnetIDs, *v.CloudSubnetID)
		}
	}
	cloudVpcIDs = append(cloudVpcIDs, cloud.Cvm.Metadata["vpc_id"])

	if len(db.Cvm.CloudVpcIDs) == 0 || len(cloudVpcIDs) == 0 || db.Cvm.CloudVpcIDs[0] != cloudVpcIDs[0] {
		return true
	}

	if len(db.Cvm.CloudSubnetIDs) == 0 || len(cloudSubnetIDs) == 0 ||
		!assert.IsStringSliceEqual(db.Cvm.CloudSubnetIDs, cloudSubnetIDs) {
		return true
	}

	privateIPv4Addresses := make([]string, 0)
	publicIPv4Addresses := make([]string, 0)
	privateIPv6Addresses := make([]string, 0)
	publicIPv6Addresses := make([]string, 0)
	for _, addresses := range cloud.Cvm.Addresses {
		for _, addresse := range addresses {
			if addresse.Version == "4" && addresse.OSEXTIPStype.Value() == "fixed" {
				privateIPv4Addresses = append(privateIPv4Addresses, addresse.Addr)
			}
			if addresse.Version == "4" && addresse.OSEXTIPStype.Value() == "floating" {
				publicIPv4Addresses = append(publicIPv4Addresses, addresse.Addr)
			}
			if addresse.Version == "6" && addresse.OSEXTIPStype.Value() == "fixed" {
				privateIPv6Addresses = append(privateIPv4Addresses, addresse.Addr)
			}
			if addresse.Version == "6" && addresse.OSEXTIPStype.Value() == "floating" {
				publicIPv6Addresses = append(publicIPv4Addresses, addresse.Addr)
			}
		}
	}

	if !assert.IsStringSliceEqual(privateIPv4Addresses, db.Cvm.PrivateIPv4Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(publicIPv4Addresses, db.Cvm.PublicIPv4Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(privateIPv6Addresses, db.Cvm.PrivateIPv6Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(publicIPv6Addresses, db.Cvm.PublicIPv6Addresses) {
		return true
	}

	if db.Cvm.MachineType != cloud.Cvm.OSEXTSTSvmState {
		return true
	}

	if db.Cvm.CloudCreatedTime != cloud.Cvm.Created {
		return true
	}

	if db.Cvm.CloudExpiredTime != cloud.Cvm.AutoTerminateTime {
		return true
	}

	if db.Cvm.Extension.AliasName != cloud.Cvm.OSEXTSRVATTRinstanceName {
		return true
	}

	if db.Cvm.Extension.HypervisorHostname != cloud.Cvm.OSEXTSRVATTRhypervisorHostname {
		return true
	}

	sgIDs := make([]string, 0)
	for _, v := range cloud.Cvm.SecurityGroups {
		sgIDs = append(sgIDs, v.Id)
	}
	if !assert.IsStringSliceEqual(db.Cvm.Extension.CloudSecurityGroupIDs, sgIDs) {
		return true
	}

	if db.Cvm.Extension.CloudTenantID != cloud.Cvm.TenantId {
		return true
	}

	if !assert.IsPtrStringEqual(db.Cvm.Extension.DiskConfig, cloud.Cvm.OSDCFdiskConfig) {
		return true
	}

	if db.Cvm.Extension.PowerState != cloud.Cvm.OSEXTSTSpowerState {
		return true
	}

	if db.Cvm.Extension.ConfigDrive != cloud.Cvm.ConfigDrive {
		return true
	}

	osDiskId := ""
	dataDiskIds := make([]string, 0)
	for _, v := range cloud.Cvm.OsExtendedVolumesvolumesAttached {
		if *v.BootIndex == "0" {
			osDiskId = v.Id
		} else {
			dataDiskIds = append(dataDiskIds, v.Id)
		}
	}
	if db.Cvm.Extension.CloudOSVolumeID != osDiskId {
		return true
	}
	if !assert.IsStringSliceEqual(db.Cvm.Extension.CloudDataVolumeIDs, dataDiskIds) {
		return true
	}

	if db.Cvm.Extension.RootDeviceName != cloud.Cvm.OSEXTSRVATTRrootDeviceName {
		return true
	}

	if !assert.IsPtrStringEqual(db.Cvm.Extension.CloudEnterpriseProjectID, cloud.Cvm.EnterpriseProjectId) {
		return true
	}

	if !assert.IsPtrInt32Equal(db.Cvm.Extension.CpuOptions.CpuThreads, cloud.Cvm.CpuOptions.HwcpuThreads) {
		return true
	}

	if db.Cvm.Extension.Flavor.CloudID != cloud.Cvm.Flavor.Id {
		return true
	}

	if db.Cvm.Extension.Flavor.Name != cloud.Cvm.Flavor.Name {
		return true
	}

	if db.Cvm.Extension.Flavor.Disk != cloud.Cvm.Flavor.Disk {
		return true
	}

	if db.Cvm.Extension.Flavor.VCpus != cloud.Cvm.Flavor.Vcpus {
		return true
	}

	if db.Cvm.Extension.Flavor.Ram != cloud.Cvm.Flavor.Ram {
		return true
	}

	if db.Cvm.Extension.Metadata.ChargingMode != cloud.Cvm.Metadata["charging_mode"] {
		return true
	}

	if db.Cvm.Extension.Metadata.CloudOrderID != cloud.Cvm.Metadata["metering.order_id"] {
		return true
	}

	if db.Cvm.Extension.Metadata.CloudProductID != cloud.Cvm.Metadata["metering.product_id"] {
		return true
	}

	if db.Cvm.Extension.Metadata.EcmResStatus != cloud.Cvm.Metadata["EcmResStatus"] {
		return true
	}

	if db.Cvm.Extension.Metadata.ImageType != cloud.Cvm.Metadata["metering.imagetype"] {
		return true
	}

	if db.Cvm.Extension.Metadata.ResourceSpecCode != cloud.Cvm.Metadata["metering.resourcespeccode"] {
		return true
	}

	if db.Cvm.Extension.Metadata.ResourceType != cloud.Cvm.Metadata["metering.resourcetype"] {
		return true
	}

	if db.Cvm.Extension.Metadata.InstanceExtraInfo != cloud.Cvm.Metadata["cascaded.instance_extrainfo"] {
		return true
	}

	if db.Cvm.Extension.Metadata.ImageName != cloud.Cvm.Metadata["image_name"] {
		return true
	}

	if db.Cvm.Extension.Metadata.AgencyName != cloud.Cvm.Metadata["agency_name"] {
		return true
	}

	if db.Cvm.Extension.Metadata.OSBit != cloud.Cvm.Metadata["os_bit"] {
		return true
	}

	if db.Cvm.Extension.Metadata.SupportAgentList != cloud.Cvm.Metadata["__support_agent_list"] {
		return true
	}

	return false
}

func syncHuaWeiCvmUpdate(kt *kit.Kit, req *SyncHuaWeiCvmOption, updateIDs []string, cloudMap map[string]*HuaWeiCvmSync,
	dsMap map[string]*HuaWeiDSCvmSync, client *huawei.HuaWei, dataCli *dataservice.Client) error {

	lists := make([]dataproto.CvmBatchUpdate[corecvm.HuaWeiCvmExtension], 0)

	for _, id := range updateIDs {
		if !isChangeHuaWei(cloudMap[id], dsMap[id], kt, req, client) {
			continue
		}

		opt := &networkinterface.HuaWeiNIListOption{
			Region:   req.Region,
			ServerID: cloudMap[id].Cvm.Id,
		}
		netInterDatas, err := client.ListNetworkInterface(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor to list huawei network interface failed, err: %v, rid: %s", err, kt.Rid)
			continue
		}

		cloudSubnetIDs := make([]string, 0)
		for _, v := range netInterDatas.Details {
			if v.CloudSubnetID != nil {
				cloudSubnetIDs = append(cloudSubnetIDs, *v.CloudSubnetID)
			}
		}
		cloudVpcID := cloudMap[id].Cvm.Metadata["vpc_id"]
		vpcID, bkCloudID, err := queryVpcIDByCloudID(kt, dataCli, cloudVpcID)
		if err != nil {
			return err
		}

		subnetIDs, err := querySubnetIDsByCloudID(kt, dataCli, cloudSubnetIDs)
		if err != nil {
			return err
		}

		privateIPv4Addresses := make([]string, 0)
		publicIPv4Addresses := make([]string, 0)
		privateIPv6Addresses := make([]string, 0)
		publicIPv6Addresses := make([]string, 0)
		for _, addresses := range cloudMap[id].Cvm.Addresses {
			for _, addresse := range addresses {
				if addresse.Version == "4" && addresse.OSEXTIPStype.Value() == "fixed" {
					privateIPv4Addresses = append(privateIPv4Addresses, addresse.Addr)
				}
				if addresse.Version == "4" && addresse.OSEXTIPStype.Value() == "floating" {
					publicIPv4Addresses = append(publicIPv4Addresses, addresse.Addr)
				}
				if addresse.Version == "6" && addresse.OSEXTIPStype.Value() == "fixed" {
					privateIPv6Addresses = append(privateIPv4Addresses, addresse.Addr)
				}
				if addresse.Version == "6" && addresse.OSEXTIPStype.Value() == "floating" {
					publicIPv6Addresses = append(publicIPv4Addresses, addresse.Addr)
				}
			}
		}

		sgIDs := make([]string, 0)
		for _, v := range cloudMap[id].Cvm.SecurityGroups {
			sgIDs = append(sgIDs, v.Id)
		}

		osDiskId := ""
		dataDiskIds := make([]string, 0)
		for _, v := range cloudMap[id].Cvm.OsExtendedVolumesvolumesAttached {
			if converter.PtrToVal(v.BootIndex) == "0" {
				osDiskId = v.Id
			} else {
				dataDiskIds = append(dataDiskIds, v.Id)
			}
		}

		cvm := dataproto.CvmBatchUpdate[corecvm.HuaWeiCvmExtension]{
			ID:                   dsMap[id].Cvm.ID,
			Name:                 cloudMap[id].Cvm.Name,
			BkCloudID:            bkCloudID,
			CloudVpcIDs:          []string{cloudVpcID},
			VpcIDs:               []string{vpcID},
			CloudSubnetIDs:       cloudSubnetIDs,
			SubnetIDs:            subnetIDs,
			Memo:                 cloudMap[id].Cvm.Description,
			Status:               cloudMap[id].Cvm.Status,
			PrivateIPv4Addresses: privateIPv4Addresses,
			PrivateIPv6Addresses: privateIPv6Addresses,
			PublicIPv4Addresses:  publicIPv4Addresses,
			PublicIPv6Addresses:  publicIPv6Addresses,
			CloudLaunchedTime:    cloudMap[id].Cvm.OSSRVUSGlaunchedAt,
			CloudExpiredTime:     cloudMap[id].Cvm.AutoTerminateTime,
			Extension: &corecvm.HuaWeiCvmExtension{
				AliasName:             cloudMap[id].Cvm.OSEXTSRVATTRinstanceName,
				HypervisorHostname:    cloudMap[id].Cvm.OSEXTSRVATTRhypervisorHostname,
				Flavor:                nil,
				CloudSecurityGroupIDs: sgIDs,
				CloudTenantID:         cloudMap[id].Cvm.TenantId,
				DiskConfig:            cloudMap[id].Cvm.OSDCFdiskConfig,
				PowerState:            cloudMap[id].Cvm.OSEXTSTSpowerState,
				ConfigDrive:           cloudMap[id].Cvm.ConfigDrive,
				Metadata: &corecvm.HuaWeiMetadata{
					ChargingMode:      cloudMap[id].Cvm.Metadata["charging_mode"],
					CloudOrderID:      cloudMap[id].Cvm.Metadata["metering.order_id"],
					CloudProductID:    cloudMap[id].Cvm.Metadata["metering.product_id"],
					EcmResStatus:      cloudMap[id].Cvm.Metadata["EcmResStatus"],
					ImageType:         cloudMap[id].Cvm.Metadata["metering.imagetype"],
					ResourceSpecCode:  cloudMap[id].Cvm.Metadata["metering.resourcespeccode"],
					ResourceType:      cloudMap[id].Cvm.Metadata["metering.resourcetype"],
					InstanceExtraInfo: cloudMap[id].Cvm.Metadata["cascaded.instance_extrainfo"],
					ImageName:         cloudMap[id].Cvm.Metadata["image_name"],
					AgencyName:        cloudMap[id].Cvm.Metadata["agency_name"],
					OSBit:             cloudMap[id].Cvm.Metadata["os_bit"],
					SupportAgentList:  cloudMap[id].Cvm.Metadata["__support_agent_list"],
				},
				CloudOSVolumeID:          osDiskId,
				CloudDataVolumeIDs:       dataDiskIds,
				RootDeviceName:           cloudMap[id].Cvm.OSEXTSRVATTRrootDeviceName,
				CloudEnterpriseProjectID: cloudMap[id].Cvm.EnterpriseProjectId,
				CpuOptions:               nil,
			},
		}

		if cloudMap[id].Cvm.Flavor != nil {
			ramInt, err := strconv.Atoi(cloudMap[id].Cvm.Flavor.Ram)
			if err != nil {
				logs.Errorf("request huawei cvm ram strconv atoi, err: %v, rid: %s", err, kt.Rid)
				return err
			}
			ram := strconv.Itoa(ramInt / 1024)
			cvm.Extension.Flavor = &corecvm.HuaWeiFlavor{
				CloudID: cloudMap[id].Cvm.Flavor.Id,
				Name:    cloudMap[id].Cvm.Flavor.Name,
				Disk:    cloudMap[id].Cvm.Flavor.Disk,
				VCpus:   cloudMap[id].Cvm.Flavor.Vcpus,
				Ram:     ram,
			}
		}

		if cloudMap[id].Cvm.CpuOptions != nil {
			cvm.Extension.CpuOptions = &corecvm.HuaWeiCpuOptions{
				CpuThreads: cloudMap[id].Cvm.CpuOptions.HwcpuThreads,
			}
		}

		lists = append(lists, cvm)
	}

	updateReq := dataproto.CvmBatchUpdateReq[corecvm.HuaWeiCvmExtension]{
		Cvms: lists,
	}

	if len(updateReq.Cvms) > 0 {
		if err := dataCli.HuaWei.Cvm.BatchUpdateCvm(kt.Ctx, kt.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateCvm failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func syncHuaWeiCvmAdd(kt *kit.Kit, addIDs []string, req *SyncHuaWeiCvmOption,
	cloudMap map[string]*HuaWeiCvmSync, client *huawei.HuaWei, dataCli *dataservice.Client) error {

	lists := make([]dataproto.CvmBatchCreate[corecvm.HuaWeiCvmExtension], 0)

	for _, id := range addIDs {

		opt := &networkinterface.HuaWeiNIListOption{
			Region:   req.Region,
			ServerID: cloudMap[id].Cvm.Id,
		}
		netInterDatas, err := client.ListNetworkInterface(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor to list huawei network interface failed, err: %v, rid: %s", err, kt.Rid)
			continue
		}

		vpcIDs := make([]string, 0)
		subnetIDs := make([]string, 0)
		for _, v := range netInterDatas.Details {
			if v.CloudVpcID != nil {
				vpcIDs = append(vpcIDs, *v.CloudVpcID)
			}
			if v.CloudSubnetID != nil {
				subnetIDs = append(subnetIDs, *v.CloudSubnetID)
			}
		}
		vpcIDs = append(vpcIDs, cloudMap[id].Cvm.Metadata["vpc_id"])

		if len(vpcIDs) <= 0 {
			return fmt.Errorf("huawei cvm: %s no vpc", cloudMap[id].Cvm.Id)
		}

		if len(vpcIDs) > 1 {
			logs.Errorf("huawei cvm: %s more than one vpc", cloudMap[id].Cvm.Id)
		}

		vpcID, bkCloudID, err := queryVpcIDByCloudID(kt, dataCli, vpcIDs[0])
		if err != nil {
			return err
		}

		subIDs, err := querySubnetIDsByCloudID(kt, dataCli, subnetIDs)
		if err != nil {
			return err
		}

		privateIPv4Addresses := make([]string, 0)
		publicIPv4Addresses := make([]string, 0)
		privateIPv6Addresses := make([]string, 0)
		publicIPv6Addresses := make([]string, 0)
		for _, addresses := range cloudMap[id].Cvm.Addresses {
			for _, addresse := range addresses {
				if addresse.Version == "4" && addresse.OSEXTIPStype.Value() == "fixed" {
					privateIPv4Addresses = append(privateIPv4Addresses, addresse.Addr)
				}
				if addresse.Version == "4" && addresse.OSEXTIPStype.Value() == "floating" {
					publicIPv4Addresses = append(publicIPv4Addresses, addresse.Addr)
				}
				if addresse.Version == "6" && addresse.OSEXTIPStype.Value() == "fixed" {
					privateIPv6Addresses = append(privateIPv4Addresses, addresse.Addr)
				}
				if addresse.Version == "6" && addresse.OSEXTIPStype.Value() == "floating" {
					publicIPv6Addresses = append(publicIPv4Addresses, addresse.Addr)
				}
			}
		}

		sgIDs := make([]string, 0)
		for _, v := range cloudMap[id].Cvm.SecurityGroups {
			sgIDs = append(sgIDs, v.Id)
		}

		osDiskId := ""
		dataDiskIds := make([]string, 0)
		for _, v := range cloudMap[id].Cvm.OsExtendedVolumesvolumesAttached {
			if converter.PtrToVal(v.BootIndex) == "0" {
				osDiskId = v.Id
			} else {
				dataDiskIds = append(dataDiskIds, v.Id)
			}
		}

		cvm := dataproto.CvmBatchCreate[corecvm.HuaWeiCvmExtension]{
			CloudID:              cloudMap[id].Cvm.Id,
			Name:                 cloudMap[id].Cvm.Name,
			BkBizID:              constant.UnassignedBiz,
			BkCloudID:            bkCloudID,
			AccountID:            req.AccountID,
			Region:               req.Region,
			Zone:                 cloudMap[id].Cvm.OSEXTAZavailabilityZone,
			CloudVpcIDs:          vpcIDs,
			VpcIDs:               []string{vpcID},
			CloudSubnetIDs:       subIDs,
			SubnetIDs:            subnetIDs,
			CloudImageID:         cloudMap[id].Cvm.Image.Id,
			OsName:               cloudMap[id].Cvm.Metadata["os_type"],
			Memo:                 cloudMap[id].Cvm.Description,
			Status:               cloudMap[id].Cvm.Status,
			PrivateIPv4Addresses: privateIPv4Addresses,
			PrivateIPv6Addresses: privateIPv6Addresses,
			PublicIPv4Addresses:  publicIPv4Addresses,
			PublicIPv6Addresses:  publicIPv6Addresses,
			MachineType:          cloudMap[id].Cvm.Flavor.Id,
			CloudCreatedTime:     cloudMap[id].Cvm.Created,
			CloudLaunchedTime:    cloudMap[id].Cvm.OSSRVUSGlaunchedAt,
			CloudExpiredTime:     cloudMap[id].Cvm.AutoTerminateTime,
			Extension: &corecvm.HuaWeiCvmExtension{
				AliasName:             cloudMap[id].Cvm.OSEXTSRVATTRinstanceName,
				HypervisorHostname:    cloudMap[id].Cvm.OSEXTSRVATTRhypervisorHostname,
				Flavor:                nil,
				CloudSecurityGroupIDs: sgIDs,
				CloudTenantID:         cloudMap[id].Cvm.TenantId,
				DiskConfig:            cloudMap[id].Cvm.OSDCFdiskConfig,
				PowerState:            cloudMap[id].Cvm.OSEXTSTSpowerState,
				ConfigDrive:           cloudMap[id].Cvm.ConfigDrive,
				Metadata: &corecvm.HuaWeiMetadata{
					ChargingMode:      cloudMap[id].Cvm.Metadata["charging_mode"],
					CloudOrderID:      cloudMap[id].Cvm.Metadata["metering.order_id"],
					CloudProductID:    cloudMap[id].Cvm.Metadata["metering.product_id"],
					EcmResStatus:      cloudMap[id].Cvm.Metadata["EcmResStatus"],
					ImageType:         cloudMap[id].Cvm.Metadata["metering.imagetype"],
					ResourceSpecCode:  cloudMap[id].Cvm.Metadata["metering.resourcespeccode"],
					ResourceType:      cloudMap[id].Cvm.Metadata["metering.resourcetype"],
					InstanceExtraInfo: cloudMap[id].Cvm.Metadata["cascaded.instance_extrainfo"],
					ImageName:         cloudMap[id].Cvm.Metadata["image_name"],
					AgencyName:        cloudMap[id].Cvm.Metadata["agency_name"],
					OSBit:             cloudMap[id].Cvm.Metadata["os_bit"],
					SupportAgentList:  cloudMap[id].Cvm.Metadata["__support_agent_list"],
				},
				CloudOSVolumeID:          osDiskId,
				CloudDataVolumeIDs:       dataDiskIds,
				RootDeviceName:           cloudMap[id].Cvm.OSEXTSRVATTRrootDeviceName,
				CloudEnterpriseProjectID: cloudMap[id].Cvm.EnterpriseProjectId,
				CpuOptions:               nil,
			},
		}

		if cloudMap[id].Cvm.Flavor != nil {
			ramInt, err := strconv.Atoi(cloudMap[id].Cvm.Flavor.Ram)
			if err != nil {
				logs.Errorf("request huawei cvm ram strconv atoi, err: %v, rid: %s", err, kt.Rid)
				return err
			}
			ram := strconv.Itoa(ramInt / 1024)
			cvm.Extension.Flavor = &corecvm.HuaWeiFlavor{
				CloudID: cloudMap[id].Cvm.Flavor.Id,
				Name:    cloudMap[id].Cvm.Flavor.Name,
				Disk:    cloudMap[id].Cvm.Flavor.Disk,
				VCpus:   cloudMap[id].Cvm.Flavor.Vcpus,
				Ram:     ram,
			}
		}

		if cloudMap[id].Cvm.CpuOptions != nil {
			cvm.Extension.CpuOptions = &corecvm.HuaWeiCpuOptions{
				CpuThreads: cloudMap[id].Cvm.CpuOptions.HwcpuThreads,
			}
		}

		lists = append(lists, cvm)
	}

	createReq := dataproto.CvmBatchCreateReq[corecvm.HuaWeiCvmExtension]{
		Cvms: lists,
	}

	if len(createReq.Cvms) > 0 {
		_, err := dataCli.HuaWei.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(), &createReq)
		if err != nil {
			logs.Errorf("request dataservice to create huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func getHuaWeiCvmDSSync(kt *kit.Kit, cloudIDs []string, req *SyncHuaWeiCvmOption,
	dataCli *dataservice.Client) ([]string, map[string]*HuaWeiDSCvmSync, error) {

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*HuaWeiDSCvmSync)

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
						Value: enumor.HuaWei,
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

		results, err := dataCli.HuaWei.Cvm.ListCvmExt(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list cvm failed, err: %v, rid: %s", err, kt.Rid)
			return updateIDs, dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				updateIDs = append(updateIDs, detail.CloudID)
				dsImageSync := new(HuaWeiDSCvmSync)
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

func getHuaWeiCvmAllDSByVendor(kt *kit.Kit, req *SyncHuaWeiCvmOption,
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

		results, err := dataCli.HuaWei.Cvm.ListCvmExt(kt.Ctx, kt.Header(), dataReq)
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

// SyncHuaWeiCvmWithRelResource sync all cvm rel resource
func SyncHuaWeiCvmWithRelResource(kt *kit.Kit, req *SyncHuaWeiCvmOption,
	ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) (interface{}, error) {

	client, err := ad.HuaWei(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	cloudSGMap, cloudVpcMap, cloudNetInterMap, cloudDiskMap, cloudEipMap, cloudSubnetMap, bootMap, err :=
		getHuaWeiCVMRelResourcesCloudIDs(kt, req, client)
	if err != nil {
		logs.Errorf("request get huawei cvm rel resource ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(cloudVpcMap) > 0 {
		opt := &vpc.SyncHuaWeiOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  converter.MapKeyToStringSlice(cloudVpcMap),
		}
		_, err := vpc.HuaWeiVpcSync(kt, opt, ad, dataCli)
		if err != nil {
			logs.Errorf("request to sync huawei vpc logic failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudSubnetMap) > 0 {
		for cloudVpcID, cloudSubnetIDMap := range cloudSubnetMap {
			req := &subnet.SyncHuaWeiOption{
				AccountID:  req.AccountID,
				Region:     req.Region,
				CloudVpcID: cloudVpcID,
				CloudIDs:   converter.MapKeyToStringSlice(cloudSubnetIDMap),
			}
			_, err := subnet.SyncHuaWeiSubnet(kt, req, ad, dataCli)
			if err != nil {
				logs.Errorf("request to sync huawei subnet logic failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}
	}

	if len(cloudEipMap) > 0 {
		eipCloudIDs := make([]string, 0)
		for _, id := range cloudEipMap {
			eipCloudIDs = append(eipCloudIDs, id.RelID)
		}

		req := &synceip.SyncHuaWeiEipOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  eipCloudIDs,
		}
		_, err := synceip.SyncHuaWeiEip(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("sync huawei cvm rel eip failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudSGMap) > 0 {
		sGCloudIDs := make([]string, 0)
		for _, id := range cloudSGMap {
			sGCloudIDs = append(sGCloudIDs, id.RelID)
		}
		req := &securitygroup.SyncHuaWeiSecurityGroupOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  sGCloudIDs,
		}
		_, err := securitygroup.SyncHuaWeiSecurityGroup(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("sync huawei cvm rel security group failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudDiskMap) > 0 {
		diskCloudIDs := make([]string, 0)
		for _, id := range cloudDiskMap {
			diskCloudIDs = append(diskCloudIDs, id.RelID)
		}
		req := &disk.SyncHuaWeiDiskOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  diskCloudIDs,
		}
		_, err := disk.SyncHuaWeiDiskWithBoot(kt, req, bootMap, ad, dataCli)
		if err != nil {
			logs.Errorf("sync huawei cvm rel disk failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	cvmReq := &SyncHuaWeiCvmOption{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  req.CloudIDs,
	}
	_, err = SyncHuaWeiCvm(kt, cvmReq, ad, dataCli)
	if err != nil {
		logs.Errorf("sync huawei cvm self failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(cloudNetInterMap) > 0 {
		niSyncReq := &hcservice.HuaWeiNetworkInterfaceSyncReq{
			AccountID:   req.AccountID,
			Region:      req.Region,
			CloudCvmIDs: req.CloudIDs,
		}
		_, err = syncnetworkinterface.HuaWeiNetworkInterfaceSync(kt, niSyncReq, ad, dataCli)
		if err != nil {
			logs.Errorf("request to sync huawei networkinterface logic failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
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

	err = getNetworkInterfaceHcIDs(kt, hcReq, dataCli, cloudNetInterMap)
	if err != nil {
		logs.Errorf("request get cvm networkinterface rel resource hc ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(cloudNetInterMap) > 0 {
		err := SyncCvmNetworkInterfaceRel(kt, cloudNetInterMap, dataCli, req.AccountID, req.CloudIDs)
		if err != nil {
			logs.Errorf("sync huawei cvm networkinterface rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudSGMap) > 0 {
		err := SyncCvmSGRel(kt, cloudSGMap, dataCli, req.AccountID, req.CloudIDs)
		if err != nil {
			logs.Errorf("sync huawei cvm sg rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudDiskMap) > 0 {
		err := SyncCvmDiskRel(kt, cloudDiskMap, dataCli, req.AccountID, req.CloudIDs)
		if err != nil {
			logs.Errorf("sync huawei cvm disk rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudEipMap) > 0 {
		err := SyncCvmEipRel(kt, cloudEipMap, dataCli, req.AccountID, req.CloudIDs)
		if err != nil {
			logs.Errorf("sync huawei cvm eip rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func getHuaWeiCVMRelResourcesCloudIDs(kt *kit.Kit, req *SyncHuaWeiCvmOption, client *huawei.HuaWei) (
	map[string]*CVMOperateSync, map[string]struct{}, map[string]*CVMOperateSync, map[string]*CVMOperateSync,
	map[string]*CVMOperateSync, map[string]map[string]struct{}, map[string]struct{}, error) {

	sGMap := make(map[string]*CVMOperateSync)
	netInterMap := make(map[string]*CVMOperateSync)
	diskMap := make(map[string]*CVMOperateSync)
	eipMap := make(map[string]*CVMOperateSync)
	subnetMap := make(map[string]map[string]struct{}, 0)
	eipIpMap := make(map[string]string)
	vpcIDMap := make(map[string]struct{}, 0)
	bootMap := make(map[string]struct{})

	opt := &typecvm.HuaWeiListOption{
		Region:   req.Region,
		CloudIDs: req.CloudIDs,
	}

	datas, err := client.ListCvm(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to list huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	for _, data := range *datas {
		if len(data.SecurityGroups) > 0 {
			for _, sg := range data.SecurityGroups {
				id := getCVMRelID(sg.Id, data.Id)
				sGMap[id] = &CVMOperateSync{RelID: sg.Id, InstanceID: data.Id}
			}
		}

		if len(data.OsExtendedVolumesvolumesAttached) > 0 {
			for _, v := range data.OsExtendedVolumesvolumesAttached {
				id := getCVMRelID(v.Id, data.Id)
				diskMap[id] = &CVMOperateSync{RelID: v.Id, InstanceID: data.Id}
				if v.BootIndex != nil && *v.BootIndex == "0" {
					bootMap[v.Id] = struct{}{}
				}
			}
		}

		if len(data.Addresses) > 0 {
			for _, address := range data.Addresses {
				if len(address) > 0 {
					for _, v := range address {
						if v.Version == "4" && v.OSEXTIPStype.Value() == "floating" {
							eipIpMap[v.Addr] = data.Id
						}
					}
				}
			}
		}

		opt := &networkinterface.HuaWeiNIListOption{
			Region:   req.Region,
			ServerID: data.Id,
		}
		netInterDatas, err := client.ListNetworkInterface(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor to list huawei network interface failed, err: %v, rid: %s", err, kt.Rid)
			continue
		}

		vpcID := data.Metadata["vpc_id"]
		vpcIDMap[vpcID] = struct{}{}
		for _, v := range netInterDatas.Details {
			if v.CloudID != nil {
				id := getCVMRelID(*v.CloudID, data.Id)
				netInterMap[id] = &CVMOperateSync{RelID: *v.CloudID, InstanceID: data.Id}
			}
			if v.CloudSubnetID != nil {
				if _, exist := subnetMap[vpcID]; !exist {
					subnetMap[vpcID] = make(map[string]struct{}, 0)
				}

				subnetMap[vpcID][*v.CloudSubnetID] = struct{}{}
			}
		}

	}

	if len(eipIpMap) > 0 {
		eipIps := make([]string, 0)
		for ip := range eipIpMap {
			eipIps = append(eipIps, ip)
		}
		opt := &eip.HuaWeiEipListOption{
			Region: req.Region,
			Ips:    eipIps,
		}

		eips, err := client.ListEip(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor to list huawei eip failed, err: %v, rid: %s", err, kt.Rid)
		}

		for _, eip := range eips.Details {
			id := getCVMRelID(eip.CloudID, eipIpMap[*eip.PublicIp])
			eipMap[id] = &CVMOperateSync{RelID: eip.CloudID, InstanceID: eipIpMap[*eip.PublicIp]}
		}
	}

	return sGMap, vpcIDMap, netInterMap, diskMap, eipMap, subnetMap, bootMap, err
}

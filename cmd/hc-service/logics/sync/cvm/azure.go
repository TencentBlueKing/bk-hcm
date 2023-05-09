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
	"hcm/pkg/adaptor/azure"
	typescore "hcm/pkg/adaptor/types/core"
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
	"hcm/pkg/tools/times"
)

// SyncAzureCvmOption ...
type SyncAzureCvmOption struct {
	AccountID         string   `json:"account_id" validate:"required"`
	ResourceGroupName string   `json:"resource_group_name" validate:"required"`
	CloudIDs          []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate SyncAzureCvmOption
func (opt SyncAzureCvmOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cloudIDs should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// SyncAzureCvm sync cvm self
func SyncAzureCvm(kt *kit.Kit, ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client,
	req *SyncAzureCvmOption) (interface{}, error) {

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := ad.Azure(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	cloudAllIDs := make(map[string]bool)

	opt := &typescore.AzureListByIDOption{
		ResourceGroupName: req.ResourceGroupName,
	}
	if len(req.CloudIDs) > 0 {
		opt.CloudIDs = req.CloudIDs
	}

	datas, err := client.ListCvmByID(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to list azure cvm failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	cloudMap := make(map[string]*AzureCvmSync)
	cloudIDs := make([]string, 0, len(datas))
	for _, data := range datas {
		cvmSync := new(AzureCvmSync)
		cvmSync.IsUpdate = false
		cvmSync.Cvm = data
		cloudMap[*data.ID] = cvmSync
		cloudIDs = append(cloudIDs, *data.ID)
		cloudAllIDs[*data.ID] = true
	}

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*AzureDSCvmSync)

	start := 0
	step := int(filter.DefaultMaxInLimit)
	for {
		var tmpCloudIDs []string
		if start+step > len(cloudIDs) {
			tmpCloudIDs = make([]string, len(cloudIDs)-start)
			copy(tmpCloudIDs, cloudIDs[start:])
		} else {
			tmpCloudIDs = make([]string, step)
			copy(tmpCloudIDs, cloudIDs[start:start+step])
		}

		if len(tmpCloudIDs) > 0 {
			tmpIDs, tmpMap, err := getAzureCvmDSSync(kt, tmpCloudIDs, req, dataCli)
			if err != nil {
				logs.Errorf("request getAzureEipDSSync failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}

			updateIDs = append(updateIDs, tmpIDs...)
			for k, v := range tmpMap {
				dsMap[k] = v
			}
		}

		start = start + step
		if start > len(cloudIDs) {
			break
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

	dsIDs, err := geAzureCvmAllDSByVendor(kt, req, enumor.Azure, dataCli)
	if err != nil {
		logs.Errorf("request geAzureCvmAllDSByVendor failed, err: %v, rid: %s", err, kt.Rid)
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

		datas, err := client.ListCvmByID(kt, opt)
		if err != nil {
			logs.Errorf("request adaptor to list azure cvm failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, id := range deleteIDs {
			realDeleteFlag := true
			for _, data := range datas {
				if *data.ID == id {
					realDeleteFlag = false
					break
				}
			}

			if realDeleteFlag {
				realDeleteIDs = append(realDeleteIDs, id)
			}
		}

		if len(realDeleteIDs) > 0 {
			if err = syncCvmDelete(kt, realDeleteIDs, dataCli); err != nil {
				logs.Errorf("request syncCvmDelete failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
			logs.Infof("[%s] account[%s] sync cvm to delete cvm success, count: %d, ids: %v, rid: %s", enumor.Azure,
				req.AccountID, len(realDeleteIDs), realDeleteIDs, kt.Rid)
		}
	}

	if len(updateIDs) > 0 {
		err := syncAzureCvmUpdate(kt, updateIDs, cloudMap, dsMap, dataCli, req, client)
		if err != nil {
			logs.Errorf("request syncAzureCvmUpdate failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		logs.Infof("[%s] account[%s] sync cvm to update cvm success, count: %d, ids: %v, rid: %s", enumor.Azure,
			req.AccountID, len(updateIDs), updateIDs, kt.Rid)
	}

	if len(addIDs) > 0 {
		err := syncAzureCvmAdd(kt, addIDs, req, cloudMap, dataCli, client)
		if err != nil {
			logs.Errorf("request syncAzureCvmAdd failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		logs.Infof("[%s] account[%s] sync cvm to add cvm success, count: %d, ids: %v, rid: %s", enumor.Azure,
			req.AccountID, len(addIDs), addIDs, kt.Rid)
	}

	return nil, nil
}

func isChangeAzure(kt *kit.Kit, cloud *AzureCvmSync, db *AzureDSCvmSync,
	req *SyncAzureCvmOption, client *azure.Azure) bool {

	if db.Cvm.CloudID != *cloud.Cvm.ID {
		return true
	}

	if db.Cvm.Name != *cloud.Cvm.Name {
		return true
	}

	netInterIDs := cloud.Cvm.NetworkInterfaceIDs

	netInterOpt := &typescore.AzureListByIDOption{
		ResourceGroupName: req.ResourceGroupName,
		CloudIDs:          netInterIDs,
	}

	netInterDatas, err := client.ListNetworkInterfaceByID(kt, netInterOpt)
	if err != nil {
		logs.Errorf("request adaptor to list azure net interface failed, err: %v, rid: %s", err, kt.Rid)
		return false
	}

	cloudVpcIDs := make([]string, 0)
	cloudSubnetIDs := make([]string, 0)
	privateIPv4Addresses := make([]string, 0)
	privateIPv6Addresses := make([]string, 0)
	publicIPv4Addresses := make([]string, 0)
	publicIPv6Addresses := make([]string, 0)
	for _, netInter := range netInterDatas.Details {
		if netInter.CloudVpcID != nil {
			cloudVpcIDs = append(cloudVpcIDs, *netInter.CloudVpcID)
		}
		if netInter.CloudSubnetID != nil {
			cloudSubnetIDs = append(cloudSubnetIDs, *netInter.CloudSubnetID)
		}

		privateIPv4Addresses = append(privateIPv4Addresses, netInter.PrivateIPv4...)
		privateIPv6Addresses = append(privateIPv6Addresses, netInter.PrivateIPv6...)
		publicIPv4Addresses = append(publicIPv4Addresses, netInter.PublicIPv4...)
		publicIPv6Addresses = append(publicIPv6Addresses, netInter.PublicIPv6...)
	}

	if len(db.Cvm.CloudVpcIDs) == 0 || len(cloudVpcIDs) == 0 || db.Cvm.CloudVpcIDs[0] != cloudVpcIDs[0] {
		return true
	}

	if len(db.Cvm.CloudSubnetIDs) == 0 || len(cloudSubnetIDs) == 0 ||
		!assert.IsStringSliceEqual(db.Cvm.CloudSubnetIDs, cloudSubnetIDs) {
		return true
	}

	if db.Cvm.CloudImageID != converter.PtrToVal(cloud.Cvm.CloudImageID) {
		return true
	}

	if db.Cvm.OsName != *cloud.Cvm.ComputerName {
		return true
	}

	if db.Cvm.Status != *cloud.Cvm.Status {
		return true
	}

	if !assert.IsStringSliceEqual(privateIPv4Addresses, db.Cvm.PrivateIPv4Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(privateIPv6Addresses, db.Cvm.PrivateIPv6Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(publicIPv4Addresses, db.Cvm.PublicIPv4Addresses) {
		return true
	}

	if !assert.IsStringSliceEqual(publicIPv6Addresses, db.Cvm.PublicIPv6Addresses) {
		return true
	}

	if db.Cvm.MachineType != string(*cloud.Cvm.VMSize) {
		return true
	}

	if db.Cvm.Extension.AdditionalCapabilities != nil {
		if !assert.IsPtrBoolEqual(db.Cvm.Extension.AdditionalCapabilities.HibernationEnabled,
			cloud.Cvm.HibernationEnabled) {
			return true
		}

		if !assert.IsPtrBoolEqual(db.Cvm.Extension.AdditionalCapabilities.UltraSSDEnabled,
			cloud.Cvm.UltraSSDEnabled) {
			return true
		}
	}

	if db.Cvm.Extension.BillingProfile != nil {
		if !assert.IsPtrFloat64Equal(db.Cvm.Extension.BillingProfile.MaxPrice,
			cloud.Cvm.MaxPrice) {
			return true
		}
	}

	if !assert.IsPtrStringEqual(db.Cvm.Extension.EvictionPolicy,
		(*string)(cloud.Cvm.EvictionPolicy)) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Cvm.Extension.LicenseType, cloud.Cvm.LicenseType) {
		return true
	}

	if !assert.IsStringSliceEqual(db.Cvm.Extension.CloudNetworkInterfaceIDs, netInterIDs) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Cvm.Extension.Priority,
		(*string)(cloud.Cvm.Priority)) {
		return true
	}

	if !assert.IsStringSliceEqual(db.Cvm.Extension.Zones, converter.PtrToSlice(cloud.Cvm.Zones)) {
		return true
	}

	if db.Cvm.Extension.StorageProfile.CloudOsDiskID != cloud.Cvm.CloudOsDiskID {
		return true
	}

	if !assert.IsStringSliceEqual(db.Cvm.Extension.StorageProfile.CloudDataDiskIDs, cloud.Cvm.CloudDataDiskIDs) {
		return true
	}

	if !assert.IsPtrStringEqual(db.Cvm.Extension.HardwareProfile.VmSize,
		(*string)(cloud.Cvm.VMSize)) {
		return true
	}

	if db.Cvm.Extension.HardwareProfile.VmSizeProperties != nil {
		if !assert.IsPtrInt32Equal(db.Cvm.Extension.HardwareProfile.VmSizeProperties.VCPUsAvailable,
			cloud.Cvm.VCPUsAvailable) {
			return true
		}

		if !assert.IsPtrInt32Equal(db.Cvm.Extension.HardwareProfile.VmSizeProperties.VCPUsPerCore,
			cloud.Cvm.VCPUsPerCore) {
			return true
		}
	}

	return false
}

func syncAzureCvmUpdate(kt *kit.Kit, updateIDs []string, cloudMap map[string]*AzureCvmSync,
	dsMap map[string]*AzureDSCvmSync, dataCli *dataservice.Client, req *SyncAzureCvmOption, client *azure.Azure) error {

	lists := make([]dataproto.CvmBatchUpdate[corecvm.AzureCvmExtension], 0)

	for _, id := range updateIDs {
		if !isChangeAzure(kt, cloudMap[id], dsMap[id], req, client) {
			continue
		}

		netInterOpt := &typescore.AzureListByIDOption{
			ResourceGroupName: req.ResourceGroupName,
			CloudIDs:          cloudMap[id].Cvm.NetworkInterfaceIDs,
		}

		netInterDatas, err := client.ListNetworkInterfaceByID(kt, netInterOpt)
		if err != nil {
			logs.Errorf("request adaptor to list azure net interface failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		cloudVpcIDs := make([]string, 0)
		cloudSubnetIDs := make([]string, 0)
		privateIPv4Addresses := make([]string, 0)
		privateIPv6Addresses := make([]string, 0)
		publicIPv4Addresses := make([]string, 0)
		publicIPv6Addresses := make([]string, 0)
		for _, netInter := range netInterDatas.Details {
			if netInter.CloudVpcID != nil {
				cloudVpcIDs = append(cloudVpcIDs, *netInter.CloudVpcID)
			}
			if netInter.CloudSubnetID != nil {
				cloudSubnetIDs = append(cloudSubnetIDs, *netInter.CloudSubnetID)
			}

			privateIPv4Addresses = append(privateIPv4Addresses, netInter.PrivateIPv4...)
			privateIPv6Addresses = append(privateIPv6Addresses, netInter.PrivateIPv6...)
			publicIPv4Addresses = append(publicIPv4Addresses, netInter.PublicIPv4...)
			publicIPv6Addresses = append(publicIPv6Addresses, netInter.PublicIPv6...)
		}

		if len(cloudVpcIDs) <= 0 {
			return fmt.Errorf("azure cvm: %s no vpc", converter.PtrToVal(cloudMap[id].Cvm.ID))
		}

		if len(cloudVpcIDs) > 1 {
			logs.Errorf("azure cvm: %s more than one vpc", converter.PtrToVal(cloudMap[id].Cvm.ID))
		}

		vpcID, bkCloudID, err := queryVpcIDByCloudID(kt, dataCli, cloudVpcIDs[0])
		if err != nil {
			return err
		}

		subnetIDs, err := querySubnetIDsByCloudID(kt, dataCli, cloudSubnetIDs)
		if err != nil {
			return err
		}

		cvm := dataproto.CvmBatchUpdate[corecvm.AzureCvmExtension]{
			ID:             dsMap[id].Cvm.ID,
			Name:           converter.PtrToVal(cloudMap[id].Cvm.Name),
			BkCloudID:      bkCloudID,
			CloudVpcIDs:    cloudVpcIDs,
			VpcIDs:         []string{vpcID},
			CloudSubnetIDs: cloudSubnetIDs,
			SubnetIDs:      subnetIDs,
			// 云上不支持该字段
			Memo:                 nil,
			Status:               converter.PtrToVal(cloudMap[id].Cvm.Status),
			PrivateIPv4Addresses: privateIPv4Addresses,
			PrivateIPv6Addresses: privateIPv6Addresses,
			PublicIPv4Addresses:  publicIPv4Addresses,
			PublicIPv6Addresses:  publicIPv6Addresses,
			// 云上不支持该字段
			CloudLaunchedTime: "",
			// 云上不支持该字段
			CloudExpiredTime: "",
			Extension: &corecvm.AzureCvmExtension{
				ResourceGroupName: req.ResourceGroupName,
				AdditionalCapabilities: &corecvm.AzureAdditionalCapabilities{
					HibernationEnabled: cloudMap[id].Cvm.HibernationEnabled,
					UltraSSDEnabled:    cloudMap[id].Cvm.UltraSSDEnabled,
				},
				BillingProfile: &corecvm.AzureBillingProfile{
					MaxPrice: cloudMap[id].Cvm.MaxPrice,
				},
				EvictionPolicy: (*string)(cloudMap[id].Cvm.EvictionPolicy),
				HardwareProfile: &corecvm.AzureHardwareProfile{
					VmSize: (*string)(cloudMap[id].Cvm.VMSize),
					VmSizeProperties: &corecvm.AzureVmSizeProperties{
						VCPUsAvailable: cloudMap[id].Cvm.VCPUsAvailable,
						VCPUsPerCore:   cloudMap[id].Cvm.VCPUsPerCore,
					},
				},
				LicenseType:              cloudMap[id].Cvm.LicenseType,
				CloudNetworkInterfaceIDs: cloudMap[id].Cvm.NetworkInterfaceIDs,
				Priority:                 (*string)(cloudMap[id].Cvm.Priority),
				StorageProfile: &corecvm.AzureStorageProfile{
					CloudDataDiskIDs: cloudMap[id].Cvm.CloudDataDiskIDs,
					CloudOsDiskID:    cloudMap[id].Cvm.CloudOsDiskID,
				},
				Zones: converter.PtrToSlice(cloudMap[id].Cvm.Zones),
			},
		}

		lists = append(lists, cvm)
	}

	updateReq := dataproto.CvmBatchUpdateReq[corecvm.AzureCvmExtension]{
		Cvms: lists,
	}

	if len(updateReq.Cvms) > 0 {
		if err := dataCli.Azure.Cvm.BatchUpdateCvm(kt.Ctx, kt.Header(), &updateReq); err != nil {
			logs.Errorf("request dataservice BatchUpdateCvm failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func syncAzureCvmAdd(kt *kit.Kit, addIDs []string, req *SyncAzureCvmOption,
	cloudMap map[string]*AzureCvmSync, dataCli *dataservice.Client, client *azure.Azure) error {

	lists := make([]dataproto.CvmBatchCreate[corecvm.AzureCvmExtension], 0)

	for _, id := range addIDs {

		netInterOpt := &typescore.AzureListByIDOption{
			ResourceGroupName: req.ResourceGroupName,
			CloudIDs:          cloudMap[id].Cvm.NetworkInterfaceIDs,
		}

		netInterDatas, err := client.ListNetworkInterfaceByID(kt, netInterOpt)
		if err != nil {
			logs.Errorf("request adaptor to list azure net interface failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		cloudVpcIDs := make([]string, 0)
		cloudSubnetIDs := make([]string, 0)
		privateIPv4Addresses := make([]string, 0)
		privateIPv6Addresses := make([]string, 0)
		publicIPv4Addresses := make([]string, 0)
		publicIPv6Addresses := make([]string, 0)
		for _, netInter := range netInterDatas.Details {
			if netInter.CloudVpcID != nil {
				cloudVpcIDs = append(cloudVpcIDs, *netInter.CloudVpcID)
			}
			if netInter.CloudSubnetID != nil {
				cloudSubnetIDs = append(cloudSubnetIDs, *netInter.CloudSubnetID)
			}

			privateIPv4Addresses = append(privateIPv4Addresses, netInter.PrivateIPv4...)
			privateIPv6Addresses = append(privateIPv6Addresses, netInter.PrivateIPv6...)
			publicIPv4Addresses = append(publicIPv4Addresses, netInter.PublicIPv4...)
			publicIPv6Addresses = append(publicIPv6Addresses, netInter.PublicIPv6...)
		}

		if len(cloudVpcIDs) <= 0 {
			return fmt.Errorf("azure cvm: %s no vpc", converter.PtrToVal(cloudMap[id].Cvm.ID))
		}

		if len(cloudVpcIDs) > 1 {
			logs.Errorf("azure cvm: %s more than one vpc", converter.PtrToVal(cloudMap[id].Cvm.ID))
		}

		vpcID, bkCloudID, err := queryVpcIDByCloudID(kt, dataCli, cloudVpcIDs[0])
		if err != nil {
			return err
		}

		subnetIDs, err := querySubnetIDsByCloudID(kt, dataCli, cloudSubnetIDs)
		if err != nil {
			return err
		}

		cvm := dataproto.CvmBatchCreate[corecvm.AzureCvmExtension]{
			CloudID:   converter.PtrToVal(cloudMap[id].Cvm.ID),
			Name:      converter.PtrToVal(cloudMap[id].Cvm.Name),
			BkBizID:   constant.UnassignedBiz,
			BkCloudID: bkCloudID,
			AccountID: req.AccountID,
			Region:    converter.PtrToVal(cloudMap[id].Cvm.Location),
			// 云上不支持该字段，azure可用区非地域概念
			Zone:           "",
			CloudVpcIDs:    cloudVpcIDs,
			VpcIDs:         []string{vpcID},
			CloudSubnetIDs: cloudSubnetIDs,
			SubnetIDs:      subnetIDs,
			CloudImageID:   converter.PtrToVal(cloudMap[id].Cvm.CloudImageID),
			OsName:         converter.PtrToVal(cloudMap[id].Cvm.ComputerName),
			// 云上不支持该字段
			Memo:                 nil,
			Status:               converter.PtrToVal(cloudMap[id].Cvm.Status),
			PrivateIPv4Addresses: privateIPv4Addresses,
			PrivateIPv6Addresses: privateIPv6Addresses,
			PublicIPv4Addresses:  publicIPv4Addresses,
			PublicIPv6Addresses:  publicIPv6Addresses,
			MachineType:          string(converter.PtrToVal(cloudMap[id].Cvm.VMSize)),
			CloudCreatedTime:     times.ConvStdTimeFormat(*cloudMap[id].Cvm.TimeCreated),
			// 云上不支持该字段
			CloudLaunchedTime: "",
			// 云上不支持该字段
			CloudExpiredTime: "",
			Extension: &corecvm.AzureCvmExtension{
				ResourceGroupName: req.ResourceGroupName,
				AdditionalCapabilities: &corecvm.AzureAdditionalCapabilities{
					HibernationEnabled: cloudMap[id].Cvm.HibernationEnabled,
					UltraSSDEnabled:    cloudMap[id].Cvm.UltraSSDEnabled,
				},
				BillingProfile: &corecvm.AzureBillingProfile{
					MaxPrice: cloudMap[id].Cvm.MaxPrice,
				},
				EvictionPolicy: (*string)(cloudMap[id].Cvm.EvictionPolicy),
				HardwareProfile: &corecvm.AzureHardwareProfile{
					VmSize: (*string)(cloudMap[id].Cvm.VMSize),
					VmSizeProperties: &corecvm.AzureVmSizeProperties{
						VCPUsAvailable: cloudMap[id].Cvm.VCPUsAvailable,
						VCPUsPerCore:   cloudMap[id].Cvm.VCPUsPerCore,
					},
				},
				LicenseType:              cloudMap[id].Cvm.LicenseType,
				CloudNetworkInterfaceIDs: cloudMap[id].Cvm.NetworkInterfaceIDs,
				Priority:                 (*string)(cloudMap[id].Cvm.Priority),
				StorageProfile: &corecvm.AzureStorageProfile{
					CloudDataDiskIDs: cloudMap[id].Cvm.CloudDataDiskIDs,
					CloudOsDiskID:    cloudMap[id].Cvm.CloudOsDiskID,
				},
				Zones: converter.PtrToSlice(cloudMap[id].Cvm.Zones),
			},
		}

		lists = append(lists, cvm)
	}

	createReq := dataproto.CvmBatchCreateReq[corecvm.AzureCvmExtension]{
		Cvms: lists,
	}

	if len(createReq.Cvms) > 0 {
		_, err := dataCli.Azure.Cvm.BatchCreateCvm(kt.Ctx, kt.Header(), &createReq)
		if err != nil {
			logs.Errorf("request dataservice to create azure cvm failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

func getAzureCvmDSSync(kt *kit.Kit, cloudIDs []string, req *SyncAzureCvmOption,
	dataCli *dataservice.Client) ([]string, map[string]*AzureDSCvmSync, error) {

	updateIDs := make([]string, 0)
	dsMap := make(map[string]*AzureDSCvmSync)

	start := 0
	for {
		dataReq := &dataproto.CvmListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "vendor",
						Op:    filter.Equal.Factory(),
						Value: enumor.Azure,
					},
					&filter.AtomRule{
						Field: "account_id",
						Op:    filter.Equal.Factory(),
						Value: req.AccountID,
					},
					&filter.AtomRule{
						Field: "extension.resource_group_name",
						Op:    filter.JSONEqual.Factory(),
						Value: req.ResourceGroupName,
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

		results, err := dataCli.Azure.Cvm.ListCvmExt(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list cvm failed, err: %v, rid: %s", err, kt.Rid)
			return updateIDs, dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				updateIDs = append(updateIDs, detail.CloudID)
				dsImageSync := new(AzureDSCvmSync)
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

func geAzureCvmAllDSByVendor(kt *kit.Kit, req *SyncAzureCvmOption,
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
						Field: "extension.resource_group_name",
						Op:    filter.JSONEqual.Factory(),
						Value: req.ResourceGroupName,
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

		results, err := dataCli.Azure.Cvm.ListCvmExt(kt.Ctx, kt.Header(), dataReq)
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

// SyncAzureCvmWithRelResource sync cvm all rel resource
func SyncAzureCvmWithRelResource(kt *kit.Kit, ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client,
	req *SyncAzureCvmOption) (interface{}, error) {

	client, err := ad.Azure(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	cloudNetInterMap, cloudVpcMap, cloudSubnetMap, cloudEipMap, cloudDiskMap, osDiskMap, err :=
		getAzureCVMRelResourcesIDs(kt, req, client)
	if err != nil {
		logs.Errorf("request get azure cvm rel resource ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(cloudVpcMap) > 0 {
		req := &hcservice.AzureResourceSyncReq{
			AccountID:         req.AccountID,
			ResourceGroupName: req.ResourceGroupName,
			CloudIDs:          converter.MapKeyToStringSlice(cloudVpcMap),
		}
		_, err := vpc.AzureVpcSync(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("request to sync azure vpc logic failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudSubnetMap) > 0 {
		for cloudVpcID, cloudSubnetIDMap := range cloudSubnetMap {
			req := &hcservice.AzureResourceSyncReq{
				AccountID:         req.AccountID,
				ResourceGroupName: req.ResourceGroupName,
				CloudVpcID:        cloudVpcID,
				CloudIDs:          converter.MapKeyToStringSlice(cloudSubnetIDMap),
			}
			_, err := subnet.AzureSubnetSync(kt, req, ad, dataCli)
			if err != nil {
				logs.Errorf("request to sync azure subnet logic failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
		}
	}

	if len(cloudEipMap) > 0 {
		eipCloudIDs := make([]string, 0)
		for _, id := range cloudEipMap {
			eipCloudIDs = append(eipCloudIDs, id.RelID)
		}
		req := &synceip.SyncAzureEipOption{
			AccountID:         req.AccountID,
			ResourceGroupName: req.ResourceGroupName,
			CloudIDs:          eipCloudIDs,
		}
		_, err := synceip.SyncAzureEip(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("sync azure cvm rel eip failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudDiskMap) > 0 {
		diskCloudIDs := make([]string, 0)
		for _, id := range cloudDiskMap {
			diskCloudIDs = append(diskCloudIDs, id.RelID)
		}
		req := &disk.SyncAzureDiskOption{
			AccountID:         req.AccountID,
			ResourceGroupName: req.ResourceGroupName,
			CloudIDs:          diskCloudIDs,
		}
		_, err := disk.SyncAzureDiskWithOs(kt, req, osDiskMap, ad, dataCli)
		if err != nil {
			logs.Errorf("sync azure cvm rel disk failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	cvmReq := &SyncAzureCvmOption{
		AccountID:         req.AccountID,
		ResourceGroupName: req.ResourceGroupName,
		CloudIDs:          req.CloudIDs,
	}
	_, err = SyncAzureCvm(kt, ad, dataCli, cvmReq)
	if err != nil {
		logs.Errorf("sync azure cvm self failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(cloudNetInterMap) > 0 {
		netInterCloudIDs := make([]string, 0)
		for _, id := range cloudNetInterMap {
			netInterCloudIDs = append(netInterCloudIDs, id.RelID)
		}
		req := &hcservice.AzureNetworkInterfaceSyncReq{
			AccountID:         req.AccountID,
			ResourceGroupName: req.ResourceGroupName,
			CloudIDs:          netInterCloudIDs,
		}
		_, err := syncnetworkinterface.AzureNetworkInterfaceSync(kt, req, ad, dataCli)
		if err != nil {
			logs.Errorf("request to sync azure networkinterface logic failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	hcReq := &protocvm.OperateSyncReq{
		AccountID: req.AccountID,
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
		err := SyncCvmNetworkInterfaceRel(kt, cloudNetInterMap, dataCli, req.AccountID, req.CloudIDs)
		if err != nil {
			logs.Errorf("sync azure cvm networkinterface rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudEipMap) > 0 {
		err := SyncCvmEipRel(kt, cloudEipMap, dataCli, req.AccountID, req.CloudIDs)
		if err != nil {
			logs.Errorf("sync azure cvm eip rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if len(cloudDiskMap) > 0 {
		err := SyncCvmDiskRel(kt, cloudDiskMap, dataCli, req.AccountID, req.CloudIDs)
		if err != nil {
			logs.Errorf("sync azure cvm disk rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	return nil, nil
}

func getAzureCVMRelResourcesIDs(kt *kit.Kit, req *SyncAzureCvmOption,
	client *azure.Azure) (map[string]*CVMOperateSync, map[string]struct{}, map[string]map[string]struct{},
	map[string]*CVMOperateSync, map[string]*CVMOperateSync, map[string]struct{}, error) {

	netInterMap := make(map[string]*CVMOperateSync)
	vpcMap := make(map[string]struct{}, 0)
	subnetMap := make(map[string]map[string]struct{})
	eipMap := make(map[string]*CVMOperateSync)
	diskMap := make(map[string]*CVMOperateSync)
	osDiskMap := make(map[string]struct{}, 0)
	netInterIDs := make([]string, 0)

	opt := &typescore.AzureListByIDOption{
		ResourceGroupName: req.ResourceGroupName,
		CloudIDs:          req.CloudIDs,
	}
	datas, err := client.ListCvmByID(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to list azure cvm failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, nil, nil, nil, err
	}

	for _, data := range datas {
		if data == nil {
			continue
		}

		for _, one := range data.NetworkInterfaceIDs {
			id := getCVMRelID(one, *data.ID)
			netInterMap[id] = &CVMOperateSync{RelID: one, InstanceID: *data.ID}
			netInterIDs = append(netInterIDs, one)
		}

		osDiskId := getCVMRelID(data.CloudOsDiskID, *data.ID)
		diskMap[osDiskId] = &CVMOperateSync{RelID: data.CloudOsDiskID, InstanceID: *data.ID}
		osDiskMap[data.CloudOsDiskID] = struct{}{}

		for _, one := range data.CloudDataDiskIDs {
			id := getCVMRelID(one, *data.ID)
			diskMap[id] = &CVMOperateSync{RelID: one, InstanceID: *data.ID}
		}
	}

	netInterOpt := &typescore.AzureListByIDOption{
		ResourceGroupName: req.ResourceGroupName,
		CloudIDs:          netInterIDs,
	}

	netInterDatas, err := client.ListNetworkInterfaceByID(kt, netInterOpt)
	if err != nil {
		logs.Errorf("request adaptor to list azure net interface failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, nil, nil, nil, err
	}

	for _, netInter := range netInterDatas.Details {
		if netInter.CloudVpcID != nil && netInter.CloudSubnetID != nil {
			vpcMap[*netInter.CloudVpcID] = struct{}{}

			if _, exist := subnetMap[*netInter.CloudVpcID]; !exist {
				subnetMap[*netInter.CloudVpcID] = make(map[string]struct{}, 0)
			}
			subnetMap[*netInter.CloudVpcID][*netInter.CloudSubnetID] = struct{}{}
		}

		if len(netInter.Extension.IPConfigurations) > 0 {
			for _, ip := range netInter.Extension.IPConfigurations {
				if ip.Properties.PublicIPAddress != nil {
					id := getCVMRelID(*ip.Properties.PublicIPAddress.CloudID, *netInter.InstanceID)
					eipMap[id] = &CVMOperateSync{RelID: *ip.Properties.PublicIPAddress.CloudID, InstanceID: *netInter.InstanceID}
				}
			}
		}
	}

	return netInterMap, vpcMap, subnetMap, eipMap, diskMap, osDiskMap, nil
}

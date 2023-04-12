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

// Package networkinterface defines network interface service.
package networkinterface

import (
	"fmt"

	"hcm/cmd/hc-service/logics/sync/logics"
	subnetlogics "hcm/cmd/hc-service/logics/sync/logics/subnet"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	typesniproto "hcm/pkg/adaptor/types/network-interface"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	datacloudproto "hcm/pkg/api/data-service/cloud"
	dataproto "hcm/pkg/api/data-service/cloud/network-interface"
	hcservice "hcm/pkg/api/hc-service"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/uuid"
)

// HuaWeiNetworkInterfaceSync sync huawei cloud network interface.
func HuaWeiNetworkInterfaceSync(kt *kit.Kit, req *hcservice.HuaWeiNetworkInterfaceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	if len(req.CloudCvmIDs) == 0 {
		return nil, nil
	}

	// syncs network interface list from cloudapi.
	allCloudIDMap, err := SyncHuaWeiNetworkInterfaceList(kt, req, adaptor, dataCli)
	if err != nil {
		logs.Errorf("%s-networkinterface request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	// compare and delete network interface idmap from db.
	err = compareDeleteHuaWeiNetworkInterfaceList(kt, req, allCloudIDMap, dataCli)
	if err != nil {
		logs.Errorf("%s-networkinterface compare delete and dblist failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// SyncHuaWeiNetworkInterfaceList sync network interface list from cloudapi.
func SyncHuaWeiNetworkInterfaceList(kt *kit.Kit, req *hcservice.HuaWeiNetworkInterfaceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (map[string]bool, error) {

	cloudIDs, allCloudIDMap, cloudList, err := GetHuaWeiNetworkInterfaceList(kt, req, adaptor, dataCli)
	if err != nil {
		logs.Errorf("%s-networkinterface batch get cloudapi failed. accountID: %s, cvmIDs: %s, region: %s, "+
			"err: %v", enumor.HuaWei, req.AccountID, req.CloudCvmIDs, req.Region, err)
		return nil, err
	}

	// get network interface info from db.
	resourceDBMap, err := BatchHuaWeiGetNetworkInterfaceMapFromDB(kt, enumor.HuaWei, cloudIDs, dataCli)
	if err != nil {
		logs.Errorf("%s-networkinterface get routetabledblist failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return allCloudIDMap, err
	}

	// compare and update network interface list.
	err = compareUpdateHuaWeiNetworkInterfaceList(kt, req, cloudList, resourceDBMap, dataCli)
	if err != nil {
		logs.Errorf("%s-networkinterface compare and update routetabledblist failed. accountID: %s, "+
			"region: %s, err: %v", enumor.HuaWei, req.AccountID, req.Region, err)
		return allCloudIDMap, err
	}

	return allCloudIDMap, nil
}

func GetHuaWeiNetworkInterfaceList(kt *kit.Kit, req *hcservice.HuaWeiNetworkInterfaceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (
	[]string, map[string]bool, *typesniproto.HuaWeiInterfaceListResult, error) {

	if len(req.CloudCvmIDs) == 0 {
		return nil, nil, nil, errf.New(errf.InvalidParameter, "cloud_cvm_ids is empty")
	}
	if len(req.CloudCvmIDs) > int(core.DefaultMaxPageLimit) {
		return nil, nil, nil, errf.New(errf.TooManyRequest, fmt.Sprintf("cloud_cvm_ids length should <= %d",
			core.DefaultMaxPageLimit))
	}

	// get cvm map by cloud_cvm_id
	cvmMapInfo, err := GetCvmMapByCloudIDs(kt, dataCli, enumor.HuaWei, req.CloudCvmIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	cli, err := adaptor.HuaWei(kt, req.AccountID)
	if err != nil {
		return nil, nil, nil, err
	}

	var (
		list          = &typesniproto.HuaWeiInterfaceListResult{}
		result        = make(map[string][]typesniproto.HuaWeiNI, len(req.CloudCvmIDs))
		cloudIDs      = make([]string, 0)
		allCloudIDMap = make(map[string]bool, 0)
	)
	for _, cvmID := range req.CloudCvmIDs {
		opt := &typesniproto.HuaWeiNIListOption{
			ServerID: cvmID,
			Region:   req.Region,
		}
		if cvmItem, ok := cvmMapInfo[cvmID]; ok {
			opt.Zone = cvmItem.Zone
		}

		tmpList, err := cli.ListNetworkInterface(kt, opt)
		if err != nil {
			logs.Errorf("%s-networkinterface batch get cloudapi failed. accountID: %s, cvmIDs: %s, region: %s, "+
				"err: %v", enumor.HuaWei, req.AccountID, req.CloudCvmIDs, req.Region, err)
			return nil, nil, nil, err
		}

		result[cvmID] = append(result[cvmID], tmpList.Details...)
		niDetails := make([]typesniproto.HuaWeiNI, 0, len(tmpList.Details))
		for _, item := range tmpList.Details {
			tmpID := converter.PtrToVal(item.CloudID)
			cloudIDs = append(cloudIDs, tmpID)
			allCloudIDMap[tmpID] = true

			// get subnet info by cloud_subnet_id
			subnetDetail, err := GetHuaWeiCloudSubnetInfoByID(kt, adaptor, dataCli, req.AccountID,
				converter.PtrToVal(item.CloudSubnetID), converter.PtrToVal(item.CloudVpcID), req.Region)
			if err != nil {
				return nil, nil, nil, err
			}
			item.SubnetID = converter.ValToPtr(subnetDetail.ID)
			item.VpcID = converter.ValToPtr(subnetDetail.VpcID)
			item.CloudVpcID = converter.ValToPtr(subnetDetail.CloudVpcID)

			niDetails = append(niDetails, item)
		}
		list.Details = append(list.Details, niDetails...)
	}

	return cloudIDs, allCloudIDMap, list, nil
}

// BatchHuaWeiGetNetworkInterfaceMapFromDB batch get network interface info from db.
func BatchHuaWeiGetNetworkInterfaceMapFromDB(kt *kit.Kit, vendor enumor.Vendor, cloudIDs []string,
	dataCli *dataclient.Client) (map[string]coreni.NetworkInterface[coreni.HuaWeiNIExtension], error) {

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "vendor",
				Op:    filter.Equal.Factory(),
				Value: vendor,
			},
			&filter.AtomRule{
				Field: "cloud_id",
				Op:    filter.In.Factory(),
				Value: cloudIDs,
			},
		},
	}
	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   &core.BasePage{Count: false, Start: 0, Limit: core.DefaultMaxPageLimit},
	}
	dbList, err := dataCli.HuaWei.NetworkInterface.ListNetworkInterfaceExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("%s-networkinterface batch get networkinterfacelist db error. limit: %d, err: %v",
			vendor, core.DefaultMaxPageLimit, err)
		return nil, err
	}

	resourceMap := make(map[string]coreni.NetworkInterface[coreni.HuaWeiNIExtension], 0)
	if len(dbList.Details) == 0 {
		return resourceMap, nil
	}

	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// compareUpdateHuaWeiNetworkInterfaceList compare and update network interface list.
func compareUpdateHuaWeiNetworkInterfaceList(kt *kit.Kit, req *hcservice.HuaWeiNetworkInterfaceSyncReq,
	list *typesniproto.HuaWeiInterfaceListResult,
	resourceDBMap map[string]coreni.NetworkInterface[coreni.HuaWeiNIExtension], dataCli *dataclient.Client) error {

	createResources, updateResources, err := filterHuaWeiNetworkInterfaceList(kt, req, list, resourceDBMap, dataCli)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &dataproto.NetworkInterfaceBatchUpdateReq[dataproto.HuaWeiNICreateExt]{
			NetworkInterfaces: updateResources,
		}
		if err = dataCli.HuaWei.NetworkInterface.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-networkinterface batch compare db update failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &dataproto.NetworkInterfaceBatchCreateReq[dataproto.HuaWeiNICreateExt]{
			NetworkInterfaces: createResources,
		}
		if _, err = dataCli.HuaWei.NetworkInterface.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("%s-networkinterface batch compare db create failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, err)
			return err
		}
	}

	return nil
}

// filterHuaWeiNetworkInterfaceList filter huawei network interface list
func filterHuaWeiNetworkInterfaceList(kt *kit.Kit, req *hcservice.HuaWeiNetworkInterfaceSyncReq,
	list *typesniproto.HuaWeiInterfaceListResult,
	resourceDBMap map[string]coreni.NetworkInterface[coreni.HuaWeiNIExtension],
	dataCli *dataclient.Client) (
	createResources []dataproto.NetworkInterfaceReq[dataproto.HuaWeiNICreateExt],
	updateResources []dataproto.NetworkInterfaceUpdateReq[dataproto.HuaWeiNICreateExt], err error) {

	if list == nil || len(list.Details) == 0 {
		return createResources, updateResources,
			fmt.Errorf("cloudapi networkinterfacelist is empty, accountID: %s, region: %s",
				req.AccountID, req.Region)
	}

	for _, item := range list.Details {
		// when sync add, if cvm is set bk_biz_id ,ni set the same bk_biz_id
		bkBizID, err := getCvmBkBizIDFromDB(kt, req.AccountID, converter.PtrToVal(item.InstanceID), dataCli)
		if err != nil {
			logs.Errorf("%s-networkinterface get cvm data from db error, err: %v", enumor.HuaWei, err)
			return nil, nil, err
		}
		// need compare and update resource data
		tmpCloudID := converter.PtrToVal(item.CloudID)
		if resourceInfo, ok := resourceDBMap[tmpCloudID]; ok {
			if !isHuaWeiChange(item, resourceInfo) {
				continue
			}

			tmpRes := dataproto.NetworkInterfaceUpdateReq[dataproto.HuaWeiNICreateExt]{
				ID:            resourceInfo.ID,
				AccountID:     req.AccountID,
				Name:          converter.PtrToVal(item.Name),
				Region:        converter.PtrToVal(item.Region),
				Zone:          converter.PtrToVal(item.Zone),
				CloudID:       converter.PtrToVal(item.CloudID),
				VpcID:         converter.PtrToVal(item.VpcID),
				CloudVpcID:    converter.PtrToVal(item.CloudVpcID),
				SubnetID:      converter.PtrToVal(item.SubnetID),
				CloudSubnetID: converter.PtrToVal(item.CloudSubnetID),
				PrivateIPv4:   item.PrivateIPv4,
				PrivateIPv6:   item.PrivateIPv6,
				PublicIPv4:    item.PublicIPv4,
				PublicIPv6:    item.PublicIPv6,
				InstanceID:    converter.PtrToVal(item.InstanceID),
			}
			if item.Extension != nil {
				tmpRes.Extension = &dataproto.HuaWeiNICreateExt{
					// MacAddr 网卡Mac地址信息。
					MacAddr: item.Extension.MacAddr,
					// NetId 网卡端口所属网络ID。
					NetId: item.Extension.NetId,
					// PortState 网卡端口状态。
					PortState: item.Extension.PortState,
					// DeleteOnTermination 卸载网卡时，是否删除网卡。
					DeleteOnTermination: item.Extension.DeleteOnTermination,
					// DriverMode 从guest os中，网卡的驱动类型。可选值为virtio和hinic，默认为virtio
					DriverMode: item.Extension.DriverMode,
					// MinRate 网卡带宽下限。
					MinRate: item.Extension.MinRate,
					// MultiqueueNum 网卡多队列个数。
					MultiqueueNum: item.Extension.MultiqueueNum,
					// PciAddress 弹性网卡在Linux GuestOS里的BDF号
					PciAddress:            item.Extension.PciAddress,
					IpV6:                  item.Extension.IpV6,
					Addresses:             (*dataproto.EipNetwork)(item.Extension.Addresses),
					CloudSecurityGroupIDs: slice.Unique(item.Extension.CloudSecurityGroupIDs),
				}
				// 网卡私网IP信息列表
				var tmpFixIps []dataproto.ServerInterfaceFixedIp
				for _, fixIpItem := range item.Extension.FixedIps {
					tmpFixIps = append(tmpFixIps, dataproto.ServerInterfaceFixedIp{
						IpAddress: fixIpItem.IpAddress,
						SubnetId:  fixIpItem.SubnetId,
					})
				}
				tmpRes.Extension.FixedIps = tmpFixIps

				var tmpVirtualIps []dataproto.NetVirtualIP
				for _, virtualIpItem := range item.Extension.VirtualIPList {
					tmpVirtualIps = append(tmpVirtualIps, dataproto.NetVirtualIP{
						IP:           virtualIpItem.IP,
						ElasticityIP: virtualIpItem.ElasticityIP,
					})
				}
				tmpRes.Extension.VirtualIPList = tmpVirtualIps
			}

			updateResources = append(updateResources, tmpRes)
		} else {
			// need add resource data
			tmpRes := dataproto.NetworkInterfaceReq[dataproto.HuaWeiNICreateExt]{
				AccountID:     req.AccountID,
				Vendor:        string(enumor.HuaWei),
				Name:          converter.PtrToVal(item.Name),
				Region:        converter.PtrToVal(item.Region),
				Zone:          converter.PtrToVal(item.Zone),
				CloudID:       converter.PtrToVal(item.CloudID),
				VpcID:         converter.PtrToVal(item.VpcID),
				CloudVpcID:    converter.PtrToVal(item.CloudVpcID),
				SubnetID:      converter.PtrToVal(item.SubnetID),
				CloudSubnetID: converter.PtrToVal(item.CloudSubnetID),
				PrivateIPv4:   item.PrivateIPv4,
				PrivateIPv6:   item.PrivateIPv6,
				PublicIPv4:    item.PublicIPv4,
				PublicIPv6:    item.PublicIPv6,
				InstanceID:    converter.PtrToVal(item.InstanceID),
				BkBizID:       bkBizID,
			}
			if item.Extension != nil {
				tmpRes.Extension = &dataproto.HuaWeiNICreateExt{
					// MacAddr 网卡Mac地址信息。
					MacAddr: item.Extension.MacAddr,
					// NetId 网卡端口所属网络ID。
					NetId: item.Extension.NetId,
					// PortState 网卡端口状态。
					PortState: item.Extension.PortState,
					// DeleteOnTermination 卸载网卡时，是否删除网卡。
					DeleteOnTermination: item.Extension.DeleteOnTermination,
					// DriverMode 从guest os中，网卡的驱动类型。可选值为virtio和hinic，默认为virtio
					DriverMode: item.Extension.DriverMode,
					// MinRate 网卡带宽下限。
					MinRate: item.Extension.MinRate,
					// MultiqueueNum 网卡多队列个数。
					MultiqueueNum: item.Extension.MultiqueueNum,
					// PciAddress 弹性网卡在Linux GuestOS里的BDF号
					PciAddress:            item.Extension.PciAddress,
					IpV6:                  item.Extension.IpV6,
					Addresses:             (*dataproto.EipNetwork)(item.Extension.Addresses),
					CloudSecurityGroupIDs: slice.Unique(item.Extension.CloudSecurityGroupIDs),
				}
				// 网卡私网IP信息列表
				var tmpFixIps []dataproto.ServerInterfaceFixedIp
				for _, fixIpItem := range item.Extension.FixedIps {
					tmpFixIps = append(tmpFixIps, dataproto.ServerInterfaceFixedIp{
						IpAddress: fixIpItem.IpAddress,
						SubnetId:  fixIpItem.SubnetId,
					})
				}
				tmpRes.Extension.FixedIps = tmpFixIps

				var tmpVirtualIps []dataproto.NetVirtualIP
				for _, virtualIpItem := range item.Extension.VirtualIPList {
					tmpVirtualIps = append(tmpVirtualIps, dataproto.NetVirtualIP{
						IP:           virtualIpItem.IP,
						ElasticityIP: virtualIpItem.ElasticityIP,
					})
				}
				tmpRes.Extension.VirtualIPList = tmpVirtualIps
			}

			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, nil
}

func isHuaWeiChange(item typesniproto.HuaWeiNI, dbInfo coreni.NetworkInterface[coreni.HuaWeiNIExtension]) bool {
	if dbInfo.Name != converter.PtrToVal(item.Name) || dbInfo.Region != converter.PtrToVal(item.Region) ||
		dbInfo.Zone != converter.PtrToVal(item.Zone) || dbInfo.CloudID != converter.PtrToVal(item.CloudID) ||
		dbInfo.CloudVpcID != converter.PtrToVal(item.CloudVpcID) ||
		dbInfo.CloudSubnetID != converter.PtrToVal(item.CloudSubnetID) ||
		dbInfo.InstanceID != converter.PtrToVal(item.InstanceID) {
		return true
	}
	if !assert.IsStringSliceEqual(item.PrivateIPv4, dbInfo.PrivateIPv4) {
		return true
	}
	if !assert.IsStringSliceEqual(item.PrivateIPv6, dbInfo.PrivateIPv6) {
		return true
	}
	if !assert.IsStringSliceEqual(item.PublicIPv4, dbInfo.PublicIPv4) {
		return true
	}
	if !assert.IsStringSliceEqual(item.PublicIPv6, dbInfo.PublicIPv6) {
		return true
	}
	if dbInfo.Extension == nil {
		return false
	}
	extRet := checkHuaWeiExt(item, dbInfo)
	if extRet {
		return true
	}
	return false
}

func checkHuaWeiExt(item typesniproto.HuaWeiNI, dbInfo coreni.NetworkInterface[coreni.HuaWeiNIExtension]) bool {
	if !assert.IsPtrStringEqual(item.Extension.MacAddr, dbInfo.Extension.MacAddr) {
		return true
	}
	if item.Extension.FixedIps != nil {
		for index, remote := range item.Extension.FixedIps {
			if len(dbInfo.Extension.FixedIps) > index {
				dbFixIpInfo := dbInfo.Extension.FixedIps[index]
				if !assert.IsPtrStringEqual(remote.SubnetId, dbFixIpInfo.SubnetId) {
					return true
				}
				if !assert.IsPtrStringEqual(remote.IpAddress, dbFixIpInfo.IpAddress) {
					return true
				}
			}
		}
	}
	if !assert.IsPtrStringEqual(item.Extension.NetId, dbInfo.Extension.NetId) {
		return true
	}
	if !assert.IsPtrStringEqual(item.Extension.PortState, dbInfo.Extension.PortState) {
		return true
	}
	if !assert.IsPtrBoolEqual(item.Extension.DeleteOnTermination, dbInfo.Extension.DeleteOnTermination) {
		return true
	}
	if !assert.IsPtrStringEqual(item.Extension.DriverMode, dbInfo.Extension.DriverMode) {
		return true
	}
	if !assert.IsPtrInt32Equal(item.Extension.MinRate, dbInfo.Extension.MinRate) {
		return true
	}
	if !assert.IsPtrInt32Equal(item.Extension.MultiqueueNum, dbInfo.Extension.MultiqueueNum) {
		return true
	}
	if !assert.IsPtrStringEqual(item.Extension.PciAddress, dbInfo.Extension.PciAddress) {
		return true
	}
	if !assert.IsPtrStringEqual(item.Extension.IpV6, dbInfo.Extension.IpV6) {
		return true
	}
	if !assert.IsStringSliceEqual(item.Extension.CloudSecurityGroupIDs, dbInfo.Extension.CloudSecurityGroupIDs) {
		return true
	}
	if item.Extension.Addresses != nil {
		if item.Extension.Addresses.BandwidthID != dbInfo.Extension.Addresses.BandwidthID {
			return true
		}
		if item.Extension.Addresses.BandwidthSize != dbInfo.Extension.Addresses.BandwidthSize {
			return true
		}
		if item.Extension.Addresses.BandwidthType != dbInfo.Extension.Addresses.BandwidthType {
			return true
		}
	}
	return false
}

// compareDeleteHuaWeiNetworkInterfaceList compare and delete network interface list from db.
func compareDeleteHuaWeiNetworkInterfaceList(kt *kit.Kit, req *hcservice.HuaWeiNetworkInterfaceSyncReq,
	allCloudIDMap map[string]bool, dataCli *dataclient.Client) error {

	page := uint32(0)
	for {
		count := core.DefaultMaxPageLimit
		offset := page * uint32(count)
		expr := &filter.Expression{
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
					Field: "instance_id",
					Op:    filter.In.Factory(),
					Value: req.CloudCvmIDs,
				},
			},
		}
		dbQueryReq := &core.ListReq{
			Filter: expr,
			Page:   &core.BasePage{Count: false, Start: offset, Limit: count},
		}
		dbList, err := dataCli.Global.NetworkInterface.List(kt.Ctx, kt.Header(), dbQueryReq)
		if err != nil {
			logs.Errorf("%s-networkinterface batch get networkinterfacelist db error. offset: %d, limit: %d, "+
				"err: %v", enumor.HuaWei, offset, count, err)
			return err
		}

		if len(dbList.Details) == 0 {
			return nil
		}

		deleteCloudIDMap := make(map[string]string, 0)
		for _, item := range dbList.Details {
			if _, ok := allCloudIDMap[item.CloudID]; !ok {
				deleteCloudIDMap[item.CloudID] = item.ID
			}
		}

		// batch query need delete network interface list
		deleteIDs := GetNeedDeleteHuaWeiNetworkInterfaceList(kt, req, deleteCloudIDMap)
		if len(deleteIDs) > 0 {
			err = BatchDeleteNetworkInterfaceByIDs(kt, deleteIDs, dataCli)
			if err != nil {
				logs.Errorf("%s-networkinterface batch compare db delete failed. deleteIDs: %v, err: %v",
					enumor.HuaWei, deleteIDs, err)
				return err
			}
		}
		deleteIDs = nil

		if len(dbList.Details) < int(count) {
			break
		}
		page++
	}
	allCloudIDMap = nil

	return nil
}

// GetNeedDeleteHuaWeiNetworkInterfaceList get need delete huawei network interface list
func GetNeedDeleteHuaWeiNetworkInterfaceList(_ *kit.Kit, _ *hcservice.HuaWeiNetworkInterfaceSyncReq,
	deleteCloudIDMap map[string]string) []string {

	deleteIDs := make([]string, 0, len(deleteCloudIDMap))
	if len(deleteCloudIDMap) == 0 {
		return deleteIDs
	}

	for _, tmpID := range deleteCloudIDMap {
		deleteIDs = append(deleteIDs, tmpID)
	}

	return deleteIDs
}

// GetHuaWeiCloudSubnetInfoByID get subnet info by cloud_subnet_id
func GetHuaWeiCloudSubnetInfoByID(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client,
	accountID, cloudSubnetID, cloudVpcID, region string) (*cloud.BaseSubnet, error) {

	// query subnet info by cloud_subnet_id
	opt := &subnetlogics.QuerySubnetIDsAndSyncOption{
		Vendor:         enumor.HuaWei,
		AccountID:      accountID,
		CloudSubnetIDs: []string{cloudSubnetID},
		Region:         region,
		CloudVpcID:     cloudVpcID,
	}
	subnetMap, err := subnetlogics.QuerySubnetIDsAndSync(kt, adaptor, dataCli, opt)
	if err != nil {
		logs.Errorf("get network interface subnet list failed, vendor: %s, accountID: %s, cloudSubnetID: %s, "+
			"region: %s, err: %v, rid: %s", enumor.HuaWei, accountID, cloudSubnetID, region, err, kt.Rid)
		return nil, err
	}
	subnetDetail, ok := subnetMap[cloudSubnetID]
	if !ok {
		return &cloud.BaseSubnet{}, nil
	}

	if len(subnetDetail.VpcID) == 0 && len(subnetDetail.CloudVpcID) != 0 {
		vpcOpt := &logics.QueryVpcIDsAndSyncOption{
			Vendor:      enumor.HuaWei,
			AccountID:   accountID,
			CloudVpcIDs: []string{subnetDetail.CloudVpcID},
			Region:      region,
		}
		vpcMap, err := logics.QueryVpcIDsAndSync(kt, adaptor, dataCli, vpcOpt)
		if err != nil {
			logs.Errorf("query vpcIDs and sync failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		if tmpVpcID, ok := vpcMap[subnetDetail.CloudVpcID]; ok {
			subnetDetail.VpcID = tmpVpcID
		}
	}

	return &subnetDetail, nil
}

// GetCvmMapByCloudIDs get cvm map by cloud_cvm_id
func GetCvmMapByCloudIDs(kt *kit.Kit, dataCli *dataclient.Client, vendor enumor.Vendor, cvmCloudIDs []string) (
	map[string]corecvm.BaseCvm, error) {

	listReq := &datacloudproto.CvmListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor},
				&filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: cvmCloudIDs},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}
	result, err := dataCli.Global.Cvm.ListCvm(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("get network interface list cvm from db failed, vendor: %s, cvmIDs: %v, err: %v, rid: %s",
			vendor, cvmCloudIDs, err, kt.Rid)
		return nil, err
	}

	var cloudCvmMap = make(map[string]corecvm.BaseCvm, len(result.Details))
	for _, item := range result.Details {
		cloudCvmMap[item.CloudID] = item
	}

	return cloudCvmMap, nil
}

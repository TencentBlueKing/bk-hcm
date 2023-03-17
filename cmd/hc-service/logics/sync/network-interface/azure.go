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
	securitygrouplogics "hcm/cmd/hc-service/logics/sync/logics/security-group"
	subnetlogics "hcm/cmd/hc-service/logics/sync/logics/subnet"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	adaptorazure "hcm/pkg/adaptor/azure"
	adcore "hcm/pkg/adaptor/types/core"
	typesniproto "hcm/pkg/adaptor/types/network-interface"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	dataservice "hcm/pkg/api/data-service"
	dataproto "hcm/pkg/api/data-service/cloud/network-interface"
	hcservice "hcm/pkg/api/hc-service"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// AzureNetworkInterfaceSync sync azure cloud network interface.
func AzureNetworkInterfaceSync(kt *kit.Kit, req *hcservice.AzureNetworkInterfaceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	// sync network interface list from cloudapi.
	var allCloudIDMap map[string]bool
	var err error
	if len(req.CloudIDs) == 0 {
		allCloudIDMap, err = SyncAzureNetworkInterfaceAll(kt, req, adaptor, dataCli)
	} else {
		allCloudIDMap, err = SyncAzureNetworkInterfaceByID(kt, req, adaptor, dataCli)
	}
	if err != nil {
		logs.Errorf("%s-networkinterface request cloudapi response failed. accountID: %s, resGroupName: %s, "+
			"err: %v", enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	// compare and delete network interface idmap from db.
	err = compareDeleteAzureNetworkInterfaceList(kt, req, allCloudIDMap, adaptor, dataCli)
	if err != nil {
		logs.Errorf("%s-networkinterface compare delete and dblist failed. accountID: %s, resGroupName: %s, "+
			"err: %v", enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// SyncAzureNetworkInterfaceAll sync network interface all from cloudapi.
func SyncAzureNetworkInterfaceAll(kt *kit.Kit, req *hcservice.AzureNetworkInterfaceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (map[string]bool, error) {

	if len(req.CloudIDs) > 0 {
		return nil, errf.New(errf.InvalidParameter, "cloud_ids is not empty")
	}

	cli, err := adaptor.Azure(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	pager, err := cli.ListNetworkInterfacePage()
	if err != nil {
		logs.Errorf("%s-routetable batch get cloudapi failed. accountID: %s, resGroupName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	allCloudIDMap := make(map[string]bool, 0)
	for pager.More() {
		page, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("list azure route table but get next page failed, err: %v", err)
		}

		tmpList := &typesniproto.AzureInterfaceListResult{}
		details := make([]typesniproto.AzureNI, 0, len(page.Value))
		for _, niItem := range page.Value {
			details = append(details, converter.PtrToVal(cli.ConvertCloudNetworkInterface(niItem)))
		}
		tmpList.Details = details

		allCloudIDMap, err = processCompareAzureNetworkInterface(kt, req, adaptor, dataCli, tmpList, allCloudIDMap)
		if err != nil {
			return nil, err
		}
	}
	return allCloudIDMap, nil
}

// SyncAzureNetworkInterfaceByID sync network interface by id from cloudapi.
func SyncAzureNetworkInterfaceByID(kt *kit.Kit, req *hcservice.AzureNetworkInterfaceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (map[string]bool, error) {

	if len(req.CloudIDs) == 0 {
		return nil, errf.New(errf.InvalidParameter, "cloud_ids is empty")
	}
	if len(req.CloudIDs) > int(core.DefaultMaxPageLimit) {
		return nil, errf.New(errf.TooManyRequest, fmt.Sprintf("cloud_ids length should <= %d",
			core.DefaultMaxPageLimit))
	}

	cli, err := adaptor.Azure(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adcore.AzureListByIDOption{
		ResourceGroupName: req.ResourceGroupName,
		CloudIDs:          req.CloudIDs,
	}
	pager, err := cli.ListNetworkInterfaceByIDPage(opt)
	if err != nil {
		logs.Errorf("%s-routetable batch get cloudapi failed. accountID: %s, resGroupName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	idMap := converter.StringSliceToMap(req.CloudIDs)
	allCloudIDMap := make(map[string]bool, 0)
	for pager.More() {
		page, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("list azure route table but get next page failed, err: %v", err)
		}

		tmpList := &typesniproto.AzureInterfaceListResult{}
		details := make([]typesniproto.AzureNI, 0, len(idMap))
		for _, niItem := range page.Value {
			id := adaptorazure.SPtrToLowerStr(niItem.ID)
			if _, exist := idMap[id]; !exist {
				continue
			}

			details = append(details, converter.PtrToVal(cli.ConvertCloudNetworkInterface(niItem)))
			delete(idMap, id)
			if len(idMap) == 0 {
				tmpList.Details = details
				break
			}
		}

		allCloudIDMap, err = processCompareAzureNetworkInterface(kt, req, adaptor, dataCli, tmpList, allCloudIDMap)
		if err != nil {
			return nil, err
		}
	}
	return allCloudIDMap, nil
}

func processCompareAzureNetworkInterface(kt *kit.Kit, req *hcservice.AzureNetworkInterfaceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client, tmpList *typesniproto.AzureInterfaceListResult,
	allCloudIDMap map[string]bool) (map[string]bool, error) {

	cloudIDs := make([]string, 0)
	niDetails := make([]typesniproto.AzureNI, 0, len(tmpList.Details))
	for _, item := range tmpList.Details {
		tmpID := converter.PtrToVal(item.CloudID)
		cloudIDs = append(cloudIDs, tmpID)
		allCloudIDMap[tmpID] = true

		// get subnet info by cloud_subnet_id
		subnetDetail, err := GetAzureCloudSubnetInfoByID(kt, adaptor, dataCli, req.AccountID,
			converter.PtrToVal(item.CloudSubnetID), converter.PtrToVal(item.CloudVpcID), req.ResourceGroupName)
		if err != nil {
			return nil, err
		}
		item.SubnetID = converter.ValToPtr(subnetDetail.ID)
		item.VpcID = converter.ValToPtr(subnetDetail.VpcID)

		// get security group ids by cloud_security_group_ids
		opt := &securitygrouplogics.QuerySecurityGroupIDsAndSyncOption{
			Vendor:                enumor.Azure,
			AccountID:             req.AccountID,
			CloudSecurityGroupIDs: make([]string, 0),
			ResourceGroupName:     req.ResourceGroupName,
			Region:                converter.PtrToVal(item.Region),
		}
		tmpSGCloudID := adaptorazure.SPtrToLowerStr(item.Extension.CloudSecurityGroupID)
		if len(tmpSGCloudID) != 0 {
			opt.CloudSecurityGroupIDs = append(opt.CloudSecurityGroupIDs, tmpSGCloudID)
		}
		securityGroupMap, err := securitygrouplogics.QuerySecurityGroupIDsAndSync(kt, adaptor, dataCli, opt)
		if err != nil {
			logs.Errorf("%s-networkinterface query security_group_ids and sync failed. accountID: %s, "+
				"resGroupName: %s, opt: %+v, err: %v, rid: %s",
				enumor.Azure, req.AccountID, req.ResourceGroupName, opt, err, kt.Rid)
			return nil, err
		}
		if tmpSGID, ok := securityGroupMap[tmpSGCloudID]; ok {
			item.Extension.SecurityGroupID = converter.ValToPtr(tmpSGID)
		}

		niDetails = append(niDetails, item)
	}
	tmpList.Details = niDetails

	// get network interface info from db.
	resourceDBMap, err := BatchAzureGetNetworkInterfaceMapFromDB(kt, enumor.Azure, cloudIDs, dataCli)
	if err != nil {
		logs.Errorf("%s-networkinterface get routetabledblist failed. accountID: %s, resGroupName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	// compare and update network interface list.
	err = compareUpdateAzureNetworkInterfaceList(kt, req, tmpList, resourceDBMap, dataCli)
	if err != nil {
		logs.Errorf("%s-networkinterface compare and update routetabledblist failed. accountID: %s, "+
			"resGroupName: %s, err: %v", enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}
	return allCloudIDMap, nil
}

// BatchAzureGetNetworkInterfaceMapFromDB batch get network interface info from db.
func BatchAzureGetNetworkInterfaceMapFromDB(kt *kit.Kit, vendor enumor.Vendor, cloudIDs []string,
	dataCli *dataclient.Client) (map[string]coreni.NetworkInterface[coreni.AzureNIExtension], error) {

	if len(cloudIDs) <= 0 {
		return make(map[string]coreni.NetworkInterface[coreni.AzureNIExtension]), nil
	}

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
	dbList, err := dataCli.Azure.NetworkInterface.ListNetworkInterfaceExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("%s-networkinterface batch get networkinterfacelist db error. limit: %d, err: %v",
			vendor, core.DefaultMaxPageLimit, err)
		return nil, err
	}

	resourceMap := make(map[string]coreni.NetworkInterface[coreni.AzureNIExtension], 0)
	if len(dbList.Details) == 0 {
		return resourceMap, nil
	}

	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// GetAzureNetworkInterfaceInfoFromDB get network interface info from db.
func GetAzureNetworkInterfaceInfoFromDB(kt *kit.Kit, id string, dataCli *dataclient.Client) (
	*coreni.NetworkInterface[coreni.AzureNIExtension], error) {

	info, err := dataCli.Azure.NetworkInterface.Get(kt.Ctx, kt.Header(), id)
	if err != nil {
		logs.Errorf("%s-networkinterface get networkinterfaceinfo db error. id: %s, err: %v",
			enumor.Azure, id, err)
		return nil, err
	}

	return info, nil
}

// compareUpdateAzureNetworkInterfaceList compare and update network interface list.
func compareUpdateAzureNetworkInterfaceList(kt *kit.Kit, req *hcservice.AzureNetworkInterfaceSyncReq,
	list *typesniproto.AzureInterfaceListResult,
	resourceDBMap map[string]coreni.NetworkInterface[coreni.AzureNIExtension], dataCli *dataclient.Client) error {

	createResources, updateResources, err := filterAzureNetworkInterfaceList(kt, req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &dataproto.NetworkInterfaceBatchUpdateReq[dataproto.AzureNICreateExt]{
			NetworkInterfaces: updateResources,
		}
		if err = dataCli.Azure.NetworkInterface.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-networkinterface batch compare db update failed. accountID: %s, resGroupName: %s, "+
				"err: %v", enumor.Azure, req.AccountID, req.ResourceGroupName, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &dataproto.NetworkInterfaceBatchCreateReq[dataproto.AzureNICreateExt]{
			NetworkInterfaces: createResources,
		}
		if _, err = dataCli.Azure.NetworkInterface.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("%s-networkinterface batch compare db create failed. accountID: %s, resGroupName: %s, "+
				"err: %v", enumor.Azure, req.AccountID, req.ResourceGroupName, err)
			return err
		}
	}

	return nil
}

// filterAzureNetworkInterfaceList filter azure network interface list
func filterAzureNetworkInterfaceList(_ *kit.Kit, req *hcservice.AzureNetworkInterfaceSyncReq,
	list *typesniproto.AzureInterfaceListResult,
	resourceDBMap map[string]coreni.NetworkInterface[coreni.AzureNIExtension]) (
	createResources []dataproto.NetworkInterfaceReq[dataproto.AzureNICreateExt],
	updateResources []dataproto.NetworkInterfaceUpdateReq[dataproto.AzureNICreateExt], err error) {

	if list == nil || len(list.Details) == 0 {
		return createResources, updateResources,
			fmt.Errorf("cloudapi networkinterfacelist is empty, accountID: %s, resGroupName: %s",
				req.AccountID, req.ResourceGroupName)
	}

	for _, item := range list.Details {
		// need compare and update resource data
		tmpCloudID := converter.PtrToVal(item.CloudID)
		if resourceInfo, ok := resourceDBMap[tmpCloudID]; ok {
			if !isAzureChange(item, resourceInfo) {
				continue
			}

			tmpRes := dataproto.NetworkInterfaceUpdateReq[dataproto.AzureNICreateExt]{
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
				tmpRes.Extension = &dataproto.AzureNICreateExt{
					ResourceGroupName: item.Extension.ResourceGroupName,
					MacAddress:        converter.PtrToVal(item.Extension.MacAddress),
					// EnableAcceleratedNetworking 是否加速网络
					EnableAcceleratedNetworking: item.Extension.EnableAcceleratedNetworking,
					// EnableIPForwarding 是否允许IP转发
					EnableIPForwarding: item.Extension.EnableIPForwarding,
					// DNSSettings DNS设置
					DNSSettings: item.Extension.DNSSettings,
					// GatewayLoadBalancerID 网关负载均衡器ID
					CloudGatewayLoadBalancerID: item.Extension.CloudGatewayLoadBalancerID,
					// CloudSecurityGroupID 网络安全组ID
					CloudSecurityGroupID: item.Extension.CloudSecurityGroupID,
					SecurityGroupID:      item.Extension.SecurityGroupID,
				}
				// IPConfigurations IP配置列表
				var tmpIPConfigs []*coreni.InterfaceIPConfiguration
				for _, cidrItem := range item.Extension.IPConfigurations {
					tmpIPConfigs = append(tmpIPConfigs, cidrItem)
				}
				tmpRes.Extension.IPConfigurations = tmpIPConfigs
			}

			updateResources = append(updateResources, tmpRes)
		} else {
			// need add resource data
			tmpRes := dataproto.NetworkInterfaceReq[dataproto.AzureNICreateExt]{
				AccountID:     req.AccountID,
				Vendor:        string(enumor.Azure),
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
				tmpRes.Extension = &dataproto.AzureNICreateExt{
					ResourceGroupName: item.Extension.ResourceGroupName,
					MacAddress:        converter.PtrToVal(item.Extension.MacAddress),
					// EnableAcceleratedNetworking 是否加速网络
					EnableAcceleratedNetworking: item.Extension.EnableAcceleratedNetworking,
					// EnableIPForwarding 是否允许IP转发
					EnableIPForwarding: item.Extension.EnableIPForwarding,
					// DNSSettings DNS设置
					DNSSettings: item.Extension.DNSSettings,
					// GatewayLoadBalancerID 网关负载均衡器ID
					CloudGatewayLoadBalancerID: item.Extension.CloudGatewayLoadBalancerID,
					// CloudSecurityGroupID 网络安全组ID
					CloudSecurityGroupID: item.Extension.CloudSecurityGroupID,
					SecurityGroupID:      item.Extension.SecurityGroupID,
				}
				// IPConfigurations IP配置列表
				var tmpIPConfigs []*coreni.InterfaceIPConfiguration
				for _, cidrItem := range item.Extension.IPConfigurations {
					tmpIPConfigs = append(tmpIPConfigs, cidrItem)
				}
				tmpRes.Extension.IPConfigurations = tmpIPConfigs
			}

			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, nil
}

func isAzureChange(item typesniproto.AzureNI, dbInfo coreni.NetworkInterface[coreni.AzureNIExtension]) bool {
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

	extRet := checkAzureExt(item, dbInfo)
	if extRet {
		return true
	}

	ipExtRet := checkAzureIPConfigIsUpdate(item, dbInfo)
	if ipExtRet {
		return true
	}
	return true
}

func checkAzureExt(item typesniproto.AzureNI, dbInfo coreni.NetworkInterface[coreni.AzureNIExtension]) bool {
	if item.Extension.ResourceGroupName != dbInfo.Extension.ResourceGroupName {
		return true
	}
	if !assert.IsPtrStringEqual(item.Extension.MacAddress, dbInfo.Extension.MacAddress) {
		return true
	}
	if !assert.IsPtrBoolEqual(item.Extension.EnableAcceleratedNetworking,
		dbInfo.Extension.EnableAcceleratedNetworking) {
		return true
	}
	if !assert.IsPtrBoolEqual(item.Extension.EnableIPForwarding, dbInfo.Extension.EnableIPForwarding) {
		return true
	}
	if item.Extension.DNSSettings != nil {
		if !assert.IsPtrStringSliceEqual(item.Extension.DNSSettings.DNSServers,
			dbInfo.Extension.DNSSettings.DNSServers) {
			return true
		}
	}
	if !assert.IsPtrStringEqual(item.Extension.CloudGatewayLoadBalancerID,
		dbInfo.Extension.CloudGatewayLoadBalancerID) {
		return true
	}
	if !assert.IsPtrStringEqual(item.Extension.CloudSecurityGroupID, dbInfo.Extension.CloudSecurityGroupID) {
		return true
	}
	return false
}

func checkAzureIPConfigIsUpdate(item typesniproto.AzureNI,
	dbInfo coreni.NetworkInterface[coreni.AzureNIExtension]) bool {

	for index, remote := range item.Extension.IPConfigurations {
		if len(dbInfo.Extension.IPConfigurations) > index {
			dbIpInfo := dbInfo.Extension.IPConfigurations[index]
			if !assert.IsPtrStringEqual(remote.CloudID, dbIpInfo.CloudID) {
				return true
			}
			if !assert.IsPtrStringEqual(remote.Name, dbIpInfo.Name) {
				return true
			}
			if !assert.IsPtrStringEqual(remote.Type, dbIpInfo.Type) {
				return true
			}
			if dbIpInfo.Properties != nil {
				if !assert.IsPtrBoolEqual(remote.Properties.Primary, dbIpInfo.Properties.Primary) {
					return true
				}
				if !assert.IsPtrStringEqual(remote.Properties.PrivateIPAddress, dbIpInfo.Properties.PrivateIPAddress) {
					return true
				}
				if !assert.IsPtrStringEqual((*string)(remote.Properties.PrivateIPAddressVersion),
					(*string)(dbIpInfo.Properties.PrivateIPAddressVersion)) {
					return true
				}
				if !assert.IsPtrStringEqual((*string)(remote.Properties.PrivateIPAllocationMethod),
					(*string)(dbIpInfo.Properties.PrivateIPAllocationMethod)) {
					return true
				}
				if !assert.IsPtrStringEqual(remote.Properties.CloudSubnetID, dbIpInfo.Properties.CloudSubnetID) {
					return true
				}
				if dbIpInfo.Properties.PublicIPAddress != nil && dbIpInfo.Properties.PublicIPAddress.Properties != nil {
					if !assert.IsPtrStringEqual(remote.Properties.PublicIPAddress.Properties.IPAddress,
						dbIpInfo.Properties.PublicIPAddress.Properties.IPAddress) {
						return true
					}
					if !assert.IsPtrStringEqual(
						(*string)(remote.Properties.PublicIPAddress.Properties.PublicIPAddressVersion),
						(*string)(dbIpInfo.Properties.PublicIPAddress.Properties.PublicIPAddressVersion)) {
						return true
					}
					if !assert.IsPtrStringEqual(
						(*string)(remote.Properties.PublicIPAddress.Properties.PublicIPAllocationMethod),
						(*string)(dbIpInfo.Properties.PublicIPAddress.Properties.PublicIPAllocationMethod)) {
						return true
					}
				}
			}
		}
	}
	return false
}

// compareDeleteAzureNetworkInterfaceList compare and delete network interface list from db.
func compareDeleteAzureNetworkInterfaceList(kt *kit.Kit, req *hcservice.AzureNetworkInterfaceSyncReq,
	allCloudIDMap map[string]bool, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) error {

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
					Value: enumor.Azure,
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
				"err: %v", enumor.Azure, offset, count, err)
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
		deleteIDs := GetNeedDeleteAzureNetworkInterfaceList(kt, req, deleteCloudIDMap, adaptor, dataCli)
		if len(deleteIDs) > 0 {
			err = BatchDeleteNetworkInterfaceByIDs(kt, deleteIDs, dataCli)
			if err != nil {
				logs.Errorf("%s-networkinterface batch compare db delete failed. deleteIDs: %v, err: %v",
					enumor.Azure, deleteIDs, err)
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

// BatchDeleteNetworkInterfaceByIDs batch delete network interface ids
func BatchDeleteNetworkInterfaceByIDs(kt *kit.Kit, deleteIDs []string, dataCli *dataclient.Client) error {
	querySize := int(filter.DefaultMaxInLimit)
	times := len(deleteIDs) / querySize
	if len(deleteIDs)%querySize != 0 {
		times++
	}

	for i := 0; i < times; i++ {
		var newDeleteIDs []string
		if i == times-1 {
			newDeleteIDs = append(newDeleteIDs, deleteIDs[i*querySize:]...)
		} else {
			newDeleteIDs = append(newDeleteIDs, deleteIDs[i*querySize:(i+1)*querySize]...)
		}

		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("id", newDeleteIDs),
		}
		if err := dataCli.Global.NetworkInterface.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
			return err
		}
	}

	return nil
}

// GetNeedDeleteAzureNetworkInterfaceList get need delete azure network interface list
func GetNeedDeleteAzureNetworkInterfaceList(kt *kit.Kit, req *hcservice.AzureNetworkInterfaceSyncReq,
	deleteCloudIDMap map[string]string, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) []string {

	deleteIDs := make([]string, 0, len(deleteCloudIDMap))
	if len(deleteCloudIDMap) == 0 {
		return deleteIDs
	}

	cli, err := adaptor.Azure(kt, req.AccountID)
	if err != nil {
		logs.Errorf("%s-networkinterface get account failed. accountID: %s, rgName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return deleteIDs
	}

	for deleteCloudID, tmpID := range deleteCloudIDMap {
		// get network interface info from db
		niInfo, err := GetAzureNetworkInterfaceInfoFromDB(kt, tmpID, dataCli)
		if err != nil {
			deleteIDs = append(deleteIDs, tmpID)
			continue
		}
		if niInfo.Extension == nil || len(niInfo.Extension.ResourceGroupName) == 0 {
			deleteIDs = append(deleteIDs, tmpID)
			continue
		}

		opt := &adcore.AzureListOption{
			ResourceGroupName:    niInfo.Extension.ResourceGroupName,
			NetworkInterfaceName: niInfo.Name,
		}

		niDetail, tmpErr := cli.GetNetworkInterface(kt, opt)
		if tmpErr != nil || converter.PtrToVal(niDetail.CloudID) != deleteCloudID {
			logs.Errorf("%s-networkinterface batch get cloudapi failed. accountID: %s, rgName: %s, opt: %+v, "+
				"err: %v", enumor.Azure, req.AccountID, req.ResourceGroupName, opt, tmpErr)
			deleteIDs = append(deleteIDs, tmpID)
			continue
		}
	}

	return deleteIDs
}

// GetAzureCloudSubnetInfoByID get subnet info by cloud_subnet_id
func GetAzureCloudSubnetInfoByID(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client,
	accountID, cloudSubnetID, cloudVpcID, resourceGroupName string) (*cloud.BaseSubnet, error) {

	// query subnet info by cloud_subnet_id
	opt := &subnetlogics.QuerySubnetIDsAndSyncOption{
		Vendor:            enumor.Azure,
		AccountID:         accountID,
		CloudSubnetIDs:    []string{cloudSubnetID},
		ResourceGroupName: resourceGroupName,
		CloudVpcID:        cloudVpcID,
	}
	subnetMap, err := subnetlogics.QuerySubnetIDsAndSync(kt, adaptor, dataCli, opt)
	if err != nil {
		logs.Errorf("get network interface subnet list failed, vendor: %s, accountID: %s, cloudSubnetID: %s, "+
			"rgName: %s, err: %v, rid: %s", enumor.Azure, accountID, cloudSubnetID, resourceGroupName, err, kt.Rid)
		return nil, err
	}
	subnetDetail, ok := subnetMap[cloudSubnetID]
	if !ok {
		return &cloud.BaseSubnet{}, nil
	}

	if len(subnetDetail.VpcID) == 0 && len(subnetDetail.CloudVpcID) != 0 {
		vpcOpt := &logics.QueryVpcIDsAndSyncOption{
			Vendor:            enumor.Azure,
			AccountID:         accountID,
			CloudVpcIDs:       []string{subnetDetail.CloudVpcID},
			ResourceGroupName: resourceGroupName,
		}
		vpcMap, err := logics.QueryVpcIDsAndSync(kt, adaptor, dataCli, vpcOpt)
		if err != nil {
			logs.Errorf("get network interface query vpc_ids and sync failed, vendor: %s, accountID: %s, "+
				"vpcOpt: %+v, err: %v, rid: %s", enumor.Azure, accountID, vpcOpt, err, kt.Rid)
			return nil, err
		}
		if tmpVpcID, ok := vpcMap[subnetDetail.CloudVpcID]; ok {
			subnetDetail.VpcID = tmpVpcID
		}
	}

	return &subnetDetail, nil
}

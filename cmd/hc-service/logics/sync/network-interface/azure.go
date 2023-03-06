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

	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	adcore "hcm/pkg/adaptor/types/core"
	typesniproto "hcm/pkg/adaptor/types/network-interface"
	"hcm/pkg/api/core"
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
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// AzureNetworkInterfaceSync sync azure cloud network interface.
func AzureNetworkInterfaceSync(kt *kit.Kit, req *hcservice.AzureNetworkInterfaceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	if len(req.CloudIDs) > 0 && len(req.CloudIDs) > 100 {
		return nil, errf.New(errf.TooManyRequest, "cloud_ids length should <= 100")
	}

	// syncs network interface list from cloudapi.
	allCloudIDMap, err := SyncAzureNetworkInterfaceList(kt, req, adaptor, dataCli)
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

// SyncAzureNetworkInterfaceList sync network interface from cloudapi.
func SyncAzureNetworkInterfaceList(kt *kit.Kit, req *hcservice.AzureNetworkInterfaceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (map[string]bool, error) {

	if len(req.CloudIDs) > 0 && len(req.CloudIDs) > 100 {
		return nil, errf.New(errf.TooManyRequest, "cloud_ids length should <= 100")
	}

	cli, err := adaptor.Azure(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	// 查询指定CloudIDs
	var tmpList *typesniproto.AzureInterfaceListResult
	if len(req.CloudIDs) > 0 {
		opt := &adcore.AzureListByIDOption{
			ResourceGroupName: req.ResourceGroupName,
			CloudIDs:          req.CloudIDs,
		}
		tmpList, err = cli.ListNetworkInterfaceByID(kt, opt)
	} else {
		tmpList, err = cli.ListNetworkInterface(kt)
	}

	if err != nil {
		logs.Errorf("%s-networkinterface batch get cloudapi failed. accountID: %s, resGroupName: %s, err: %v",
			enumor.Azure, req.AccountID, req.ResourceGroupName, err)
		return nil, err
	}

	cloudIDs := make([]string, 0)
	allCloudIDMap := make(map[string]bool, 0)
	for _, item := range tmpList.Details {
		tmpID := converter.PtrToVal(item.CloudID)
		cloudIDs = append(cloudIDs, tmpID)
		allCloudIDMap[tmpID] = true
	}

	// get network interface info from db.
	resourceDBMap, err := BatchGetNetworkInterfaceMapFromDB(kt, enumor.Azure, cloudIDs, dataCli)
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

// BatchGetNetworkInterfaceMapFromDB batch get network interface info from db.
func BatchGetNetworkInterfaceMapFromDB(kt *kit.Kit, vendor enumor.Vendor, cloudIDs []string,
	dataCli *dataclient.Client) (map[string]coreni.BaseNetworkInterface, error) {

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
	dbList, err := dataCli.Global.NetworkInterface.List(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("%s-networkinterface batch get networkinterfacelist db error. limit: %d, err: %v",
			vendor, core.DefaultMaxPageLimit, err)
		return nil, err
	}

	resourceMap := make(map[string]coreni.BaseNetworkInterface, 0)
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
	list *typesniproto.AzureInterfaceListResult, resourceDBMap map[string]coreni.BaseNetworkInterface,
	dataCli *dataclient.Client) error {

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
	list *typesniproto.AzureInterfaceListResult, resourceDBMap map[string]coreni.BaseNetworkInterface) (
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
			if resourceInfo.Name == converter.PtrToVal(item.Name) &&
				resourceInfo.Region == converter.PtrToVal(item.Region) &&
				resourceInfo.CloudVpcID == converter.PtrToVal(item.CloudVpcID) &&
				resourceInfo.CloudSubnetID == converter.PtrToVal(item.CloudSubnetID) &&
				resourceInfo.PrivateIP == converter.PtrToVal(item.PrivateIP) &&
				resourceInfo.PublicIP == converter.PtrToVal(item.PublicIP) &&
				resourceInfo.InstanceID == converter.PtrToVal(item.InstanceID) {
				continue
			}

			tmpRes := dataproto.NetworkInterfaceUpdateReq[dataproto.AzureNICreateExt]{
				ID:            resourceInfo.ID,
				AccountID:     req.AccountID,
				Name:          converter.PtrToVal(item.Name),
				Region:        converter.PtrToVal(item.Region),
				Zone:          converter.PtrToVal(item.Zone),
				CloudID:       converter.PtrToVal(item.CloudID),
				CloudVpcID:    converter.PtrToVal(item.CloudVpcID),
				CloudSubnetID: converter.PtrToVal(item.CloudSubnetID),
				PrivateIP:     converter.PtrToVal(item.PrivateIP),
				PublicIP:      converter.PtrToVal(item.PublicIP),
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
				CloudVpcID:    converter.PtrToVal(item.CloudVpcID),
				CloudSubnetID: converter.PtrToVal(item.CloudSubnetID),
				PrivateIP:     converter.PtrToVal(item.PrivateIP),
				PublicIP:      converter.PtrToVal(item.PublicIP),
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

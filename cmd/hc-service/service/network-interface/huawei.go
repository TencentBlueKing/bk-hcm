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

	typesniproto "hcm/pkg/adaptor/types/network-interface"
	"hcm/pkg/api/core"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	dataproto "hcm/pkg/api/data-service/cloud/network-interface"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// HuaWeiNetworkInterfaceSync sync huawei cloud network interface.
func (n networkInterfaceAdaptor) HuaWeiNetworkInterfaceSync(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.HuaWeiNetworkInterfaceSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	err := req.Validate()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// syncs network interface list from cloudapi.
	allCloudIDMap, err := n.SyncHuaWeiNetworkInterfaceList(cts, req)
	if err != nil {
		logs.Errorf("%s-networkinterface request cloudapi response failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return nil, err
	}

	// compare and delete network interface idmap from db.
	err = n.compareDeleteHuaWeiNetworkInterfaceList(cts, req, allCloudIDMap)
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
func (n networkInterfaceAdaptor) SyncHuaWeiNetworkInterfaceList(cts *rest.Contexts,
	req *hcservice.HuaWeiNetworkInterfaceSyncReq) (map[string]bool, error) {

	allCloudIDMap := make(map[string]bool, 0)
	cli, err := n.adaptor.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return allCloudIDMap, err
	}

	opt := &typesniproto.HuaWeiNIListOption{
		AccountID: req.AccountID,
		ServerID:  req.ServerID,
		Region:    req.Region,
	}
	tmpList, tmpErr := cli.ListNetworkInterface(cts.Kit, opt)
	if tmpErr != nil {
		logs.Errorf("%s-networkinterface batch get cloudapi failed. accountID: %s, serverID: %s, region: %s, "+
			"err: %v", enumor.HuaWei, req.AccountID, req.ServerID, req.Region, tmpErr)
		return allCloudIDMap, tmpErr
	}

	cloudIDs := make([]string, 0)
	for _, item := range tmpList.Details {
		tmpID := converter.PtrToVal(item.CloudID)
		cloudIDs = append(cloudIDs, tmpID)
		allCloudIDMap[tmpID] = true
	}

	// get network interface info from db.
	resourceDBMap, err := n.BatchGetNetworkInterfaceMapFromDB(cts, enumor.HuaWei, cloudIDs)
	if err != nil {
		logs.Errorf("%s-networkinterface get routetabledblist failed. accountID: %s, region: %s, err: %v",
			enumor.HuaWei, req.AccountID, req.Region, err)
		return allCloudIDMap, err
	}

	// compare and update network interface list.
	err = n.compareUpdateHuaWeiNetworkInterfaceList(cts, req, tmpList, resourceDBMap)
	if err != nil {
		logs.Errorf("%s-networkinterface compare and update routetabledblist failed. accountID: %s, "+
			"region: %s, err: %v", enumor.HuaWei, req.AccountID, req.Region, err)
		return allCloudIDMap, err
	}

	return allCloudIDMap, nil
}

// compareUpdateHuaWeiNetworkInterfaceList compare and update network interface list.
func (n networkInterfaceAdaptor) compareUpdateHuaWeiNetworkInterfaceList(cts *rest.Contexts,
	req *hcservice.HuaWeiNetworkInterfaceSyncReq, list *typesniproto.HuaWeiInterfaceListResult,
	resourceDBMap map[string]coreni.BaseNetworkInterface) error {
	createResources, updateResources, err := n.filterHuaWeiNetworkInterfaceList(cts, req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &dataproto.NetworkInterfaceBatchUpdateReq[dataproto.HuaWeiNICreateExt]{
			NetworkInterfaces: updateResources,
		}
		if err = n.dataCli.HuaWei.NetworkInterface.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
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
		if _, err = n.dataCli.HuaWei.NetworkInterface.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq); err != nil {
			logs.Errorf("%s-networkinterface batch compare db create failed. accountID: %s, region: %s, err: %v",
				enumor.HuaWei, req.AccountID, req.Region, err)
			return err
		}
	}

	return nil
}

// filterHuaWeiNetworkInterfaceList filter huawei network interface list
func (n networkInterfaceAdaptor) filterHuaWeiNetworkInterfaceList(_ *rest.Contexts,
	req *hcservice.HuaWeiNetworkInterfaceSyncReq, list *typesniproto.HuaWeiInterfaceListResult,
	resourceDBMap map[string]coreni.BaseNetworkInterface) (
	createResources []dataproto.NetworkInterfaceReq[dataproto.HuaWeiNICreateExt],
	updateResources []dataproto.NetworkInterfaceUpdateReq[dataproto.HuaWeiNICreateExt], err error) {

	if list == nil || len(list.Details) == 0 {
		return createResources, updateResources,
			fmt.Errorf("cloudapi networkinterfacelist is empty, accountID: %s, region: %s",
				req.AccountID, req.Region)
	}

	for _, item := range list.Details {
		// need compare and update resource data
		tmpCloudID := converter.PtrToVal(item.CloudID)
		if resourceInfo, ok := resourceDBMap[tmpCloudID]; ok {
			if resourceInfo.Name == converter.PtrToVal(item.Name) &&
				resourceInfo.PrivateIP == converter.PtrToVal(item.PrivateIP) {
				continue
			}

			tmpRes := dataproto.NetworkInterfaceUpdateReq[dataproto.HuaWeiNICreateExt]{
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
				tmpRes.Extension = &dataproto.HuaWeiNICreateExt{
					// MacAddr 网卡Mac地址信息。
					MacAddr: item.Extension.MacAddr,
					// NetId 网卡端口所属网络ID。
					NetId: item.Extension.NetId,
					// PortId 网卡端口ID。
					PortId: item.Extension.PortId,
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
					PciAddress: item.Extension.PciAddress,
					IpV6:       item.Extension.IpV6,
					Addresses:  (*dataproto.EipNetwork)(item.Extension.Addresses),
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

				// 安全组
				if len(item.Extension.SecurityGroups) > 0 {
					var sgArr []dataproto.NovaServerSecurityGroup
					for _, sgItem := range item.Extension.SecurityGroups {
						sgArr = append(sgArr, dataproto.NovaServerSecurityGroup{
							ID:   sgItem.CloudID,
							Name: sgItem.Name,
						})
					}
					tmpRes.Extension.SecurityGroups = sgArr
				}
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
				CloudVpcID:    converter.PtrToVal(item.CloudVpcID),
				CloudSubnetID: converter.PtrToVal(item.CloudSubnetID),
				PrivateIP:     converter.PtrToVal(item.PrivateIP),
				PublicIP:      converter.PtrToVal(item.PublicIP),
				InstanceID:    converter.PtrToVal(item.InstanceID),
			}
			if item.Extension != nil {
				tmpRes.Extension = &dataproto.HuaWeiNICreateExt{
					// MacAddr 网卡Mac地址信息。
					MacAddr: item.Extension.MacAddr,
					// NetId 网卡端口所属网络ID。
					NetId: item.Extension.NetId,
					// PortId 网卡端口ID。
					PortId: item.Extension.PortId,
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
					PciAddress: item.Extension.PciAddress,
					IpV6:       item.Extension.IpV6,
					Addresses:  (*dataproto.EipNetwork)(item.Extension.Addresses),
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

				// 安全组
				if len(item.Extension.SecurityGroups) > 0 {
					var sgArr []dataproto.NovaServerSecurityGroup
					for _, sgItem := range item.Extension.SecurityGroups {
						sgArr = append(sgArr, dataproto.NovaServerSecurityGroup{
							ID:   sgItem.CloudID,
							Name: sgItem.Name,
						})
					}
					tmpRes.Extension.SecurityGroups = sgArr
				}
			}

			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, nil
}

// compareDeleteHuaWeiNetworkInterfaceList compare and delete network interface list from db.
func (n networkInterfaceAdaptor) compareDeleteHuaWeiNetworkInterfaceList(cts *rest.Contexts,
	req *hcservice.HuaWeiNetworkInterfaceSyncReq, allCloudIDMap map[string]bool) error {

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
			},
		}
		dbQueryReq := &core.ListReq{
			Filter: expr,
			Page:   &core.BasePage{Count: false, Start: offset, Limit: count},
		}
		dbList, err := n.dataCli.Global.NetworkInterface.List(cts.Kit.Ctx, cts.Kit.Header(), dbQueryReq)
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
		deleteIDs := n.GetNeedDeleteHuaWeiNetworkInterfaceList(cts, req, deleteCloudIDMap)
		if len(deleteIDs) > 0 {
			err = n.BatchDeleteNetworkInterfaceByIDs(cts, deleteIDs)
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
func (n networkInterfaceAdaptor) GetNeedDeleteHuaWeiNetworkInterfaceList(_ *rest.Contexts,
	_ *hcservice.HuaWeiNetworkInterfaceSyncReq, deleteCloudIDMap map[string]string) []string {

	deleteIDs := make([]string, 0, len(deleteCloudIDMap))
	if len(deleteCloudIDMap) == 0 {
		return deleteIDs
	}

	for _, tmpID := range deleteCloudIDMap {
		deleteIDs = append(deleteIDs, tmpID)
	}

	return deleteIDs
}

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
	"hcm/pkg/api/core/cloud"
	coreni "hcm/pkg/api/core/cloud/network-interface"
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
	"hcm/pkg/tools/uuid"

	"google.golang.org/api/compute/v1"
)

// GcpNetworkInterfaceSync sync gcp cloud network interface.
func GcpNetworkInterfaceSync(kt *kit.Kit, req *hcservice.GcpNetworkInterfaceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (interface{}, error) {

	// sync network interface list from cloudapi.
	var (
		err           error
		allCloudIDMap = make(map[string]bool, 0)
	)
	if len(req.CloudCvmIDs) == 0 {
		allCloudIDMap, err = SyncGcpNetworkInterfaceListAll(kt, req, adaptor, dataCli)
	} else {
		allCloudIDMap, err = SyncGcpNetworkInterfaceListByCloudIDs(kt, req, adaptor, dataCli)
	}
	if err != nil {
		logs.Errorf("%s-networkinterface request cloudapi response failed. accountID: %s, zone: %s, err: %v",
			enumor.Gcp, req.AccountID, req.Zone, err)
		return nil, err
	}

	// compare and delete network interface idmap from db.
	err = compareDeleteGcpNetworkInterfaceList(kt, req, allCloudIDMap, dataCli)
	if err != nil {
		logs.Errorf("%s-networkinterface compare delete and dblist failed. accountID: %s, zone: %s, err: %v",
			enumor.Gcp, req.AccountID, req.Zone, err)
		return nil, err
	}

	return &hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

func SyncGcpNetworkInterfaceListAll(kt *kit.Kit, req *hcservice.GcpNetworkInterfaceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (map[string]bool, error) {

	cli, err := adaptor.Gcp(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adcore.GcpListOption{
		Zone: req.Zone,
	}
	listCall, err := cli.ListNetworkInterfacePage(kt, opt)
	if err != nil {
		logs.Errorf("%s-networkinterface batch get cloudapi failed. accountID: %s, zone: %s, err: %v",
			enumor.Gcp, req.AccountID, req.Zone, err)
		return nil, err
	}

	allCloudIDMap := make(map[string]bool, 0)
	if err := listCall.Pages(kt.Ctx, func(page *compute.InstanceList) error {
		details := make([]typesniproto.GcpNI, 0, len(page.Items))
		for _, item := range page.Items {
			for _, niItem := range item.NetworkInterfaces {
				details = append(details, converter.PtrToVal(cli.ConvertNetworkInterface(item, niItem)))
			}
		}

		list := &typesniproto.GcpInterfaceListResult{}
		niDetails := make([]typesniproto.GcpNI, 0, len(details))
		cloudIDs := make([]string, 0)
		for _, item := range details {
			tmpID := converter.PtrToVal(item.CloudID)
			cloudIDs = append(cloudIDs, tmpID)
			allCloudIDMap[tmpID] = true

			// get gcp vpc info by vpc selflink
			vpcDetail, err := GetCloudVpcInfoBySelfLink(kt, req, enumor.Gcp, item.CloudVpcID, dataCli)
			if err != nil {
				return err
			}
			item.CloudVpcID = converter.ValToPtr(vpcDetail.CloudID)
			item.VpcID = converter.ValToPtr(vpcDetail.ID)

			// get gcp subnet info by subnet selflink
			subnetDetail, err := GetCloudSubnetInfoBySelfLink(kt, req, enumor.Gcp, item.CloudSubnetID, dataCli)
			if err != nil {
				return err
			}
			item.CloudSubnetID = converter.ValToPtr(subnetDetail.CloudID)
			item.SubnetID = converter.ValToPtr(subnetDetail.ID)
			niDetails = append(niDetails, item)
		}
		list.Details = niDetails

		// get network interface info from db.
		resourceDBMap, err := BatchGcpGetNetworkInterfaceMapFromDB(kt, enumor.Gcp, cloudIDs, dataCli)
		if err != nil {
			logs.Errorf("%s-networkinterface get routetabledblist failed. accountID: %s, zone: %s, err: %v",
				enumor.Gcp, req.AccountID, req.Zone, err)
			return err
		}

		// compare and update network interface list.
		err = compareUpdateGcpNetworkInterfaceList(kt, req, list, resourceDBMap, dataCli)
		if err != nil {
			logs.Errorf("%s-networkinterface compare and update routetabledblist failed. accountID: %s, "+
				"zone: %s, err: %v", enumor.Gcp, req.AccountID, req.Zone, err)
			return err
		}

		return nil
	}); err != nil {
		logs.Errorf("cloudapi failed to list gcp network interface, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return allCloudIDMap, nil
}

func SyncGcpNetworkInterfaceListByCloudIDs(kt *kit.Kit, req *hcservice.GcpNetworkInterfaceSyncReq,
	adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client) (map[string]bool, error) {

	if len(req.CloudCvmIDs) == 0 {
		return nil, errf.New(errf.InvalidParameter, "cloud_cvm_ids is empty")
	}
	if len(req.CloudCvmIDs) > int(core.DefaultMaxPageLimit) {
		return nil, errf.New(errf.TooManyRequest, fmt.Sprintf("cloud_cvm_ids length should <= %d",
			core.DefaultMaxPageLimit))
	}

	cli, err := adaptor.Gcp(kt, req.AccountID)
	if err != nil {
		return nil, err
	}

	var cvmMap map[string] /*CloudCvmID*/ []typesniproto.GcpNI
	opt := &typesniproto.GcpListByCvmIDOption{
		Zone:        req.Zone,
		CloudCvmIDs: req.CloudCvmIDs,
	}
	cvmMap, err = cli.ListNetworkInterfaceByCvmID(kt, opt)
	if err != nil {
		logs.Errorf("%s-networkinterface batch get cloudapi failed. accountID: %s, zone: %s, err: %v",
			enumor.Gcp, req.AccountID, req.Zone, err)
		return nil, err
	}

	var allCloudIDMap = make(map[string]bool, 0)
	for _, niList := range cvmMap {
		var cloudIDs = make([]string, 0)
		var list = &typesniproto.GcpInterfaceListResult{}
		for _, niItem := range niList {
			tmpID := converter.PtrToVal(niItem.CloudID)
			cloudIDs = append(cloudIDs, tmpID)
			allCloudIDMap[tmpID] = true

			// get gcp vpc info by vpc-selflink
			vpcDetail, err := GetCloudVpcInfoBySelfLink(kt, req, enumor.Gcp, niItem.CloudVpcID, dataCli)
			if err != nil {
				return nil, err
			}
			niItem.CloudVpcID = converter.ValToPtr(vpcDetail.CloudID)
			niItem.VpcID = converter.ValToPtr(vpcDetail.ID)

			// get gcp subnet info by subnet-selflink
			subnetDetail, err := GetCloudSubnetInfoBySelfLink(kt, req, enumor.Gcp, niItem.CloudSubnetID, dataCli)
			if err != nil {
				return nil, err
			}
			niItem.CloudSubnetID = converter.ValToPtr(subnetDetail.CloudID)
			niItem.SubnetID = converter.ValToPtr(subnetDetail.ID)
			list.Details = append(list.Details, niItem)
		}

		// get network interface info from db.
		resourceDBMap, err := BatchGcpGetNetworkInterfaceMapFromDB(kt, enumor.Gcp, cloudIDs, dataCli)
		if err != nil {
			logs.Errorf("%s-networkinterface get routetabledblist failed. accountID: %s, zone: %s, err: %v",
				enumor.Gcp, req.AccountID, req.Zone, err)
			return nil, err
		}

		// compare and update network interface list.
		err = compareUpdateGcpNetworkInterfaceList(kt, req, list, resourceDBMap, dataCli)
		if err != nil {
			logs.Errorf("%s-networkinterface compare and update routetabledblist failed. accountID: %s, "+
				"zone: %s, err: %v", enumor.Gcp, req.AccountID, req.Zone, err)
			return nil, err
		}
	}

	return allCloudIDMap, nil
}

// BatchGcpGetNetworkInterfaceMapFromDB batch get network interface info from db.
func BatchGcpGetNetworkInterfaceMapFromDB(kt *kit.Kit, vendor enumor.Vendor, cloudIDs []string,
	dataCli *dataclient.Client) (map[string]coreni.NetworkInterface[coreni.GcpNIExtension], error) {

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
	dbList, err := dataCli.Gcp.NetworkInterface.ListNetworkInterfaceExt(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("%s-networkinterface batch get networkinterfacelist db error. limit: %d, err: %v",
			vendor, core.DefaultMaxPageLimit, err)
		return nil, err
	}

	resourceMap := make(map[string]coreni.NetworkInterface[coreni.GcpNIExtension], 0)
	if len(dbList.Details) == 0 {
		return resourceMap, nil
	}

	for _, item := range dbList.Details {
		resourceMap[item.CloudID] = item
	}

	return resourceMap, nil
}

// compareUpdateGcpNetworkInterfaceList compare and update network interface list.
func compareUpdateGcpNetworkInterfaceList(kt *kit.Kit, req *hcservice.GcpNetworkInterfaceSyncReq,
	list *typesniproto.GcpInterfaceListResult, resourceDBMap map[string]coreni.NetworkInterface[coreni.GcpNIExtension],
	dataCli *dataclient.Client) error {

	createResources, updateResources, err := filterGcpNetworkInterfaceList(kt, req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &dataproto.NetworkInterfaceBatchUpdateReq[dataproto.GcpNICreateExt]{
			NetworkInterfaces: updateResources,
		}
		if err = dataCli.Gcp.NetworkInterface.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
			logs.Errorf("%s-networkinterface batch compare db update failed. accountID: %s, zone: %s, err: %v",
				enumor.Gcp, req.AccountID, req.Zone, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &dataproto.NetworkInterfaceBatchCreateReq[dataproto.GcpNICreateExt]{
			NetworkInterfaces: createResources,
		}
		if _, err = dataCli.Gcp.NetworkInterface.BatchCreate(kt.Ctx, kt.Header(), createReq); err != nil {
			logs.Errorf("%s-networkinterface batch compare db create failed. accountID: %s, zone: %s, err: %v",
				enumor.Gcp, req.AccountID, req.Zone, err)
			return err
		}
	}

	return nil
}

// filterGcpNetworkInterfaceList filter gcp network interface list
func filterGcpNetworkInterfaceList(_ *kit.Kit, req *hcservice.GcpNetworkInterfaceSyncReq,
	list *typesniproto.GcpInterfaceListResult,
	resourceDBMap map[string]coreni.NetworkInterface[coreni.GcpNIExtension]) (
	createResources []dataproto.NetworkInterfaceReq[dataproto.GcpNICreateExt],
	updateResources []dataproto.NetworkInterfaceUpdateReq[dataproto.GcpNICreateExt], err error) {

	if list == nil || len(list.Details) == 0 {
		return createResources, updateResources,
			fmt.Errorf("cloudapi networkinterfacelist is empty, accountID: %s, zone: %s",
				req.AccountID, req.Zone)
	}

	for _, item := range list.Details {
		// need compare and update resource data
		tmpCloudID := converter.PtrToVal(item.CloudID)
		if resourceInfo, ok := resourceDBMap[tmpCloudID]; ok {
			if !isGcpChange(item, resourceInfo) {
				continue
			}

			tmpRes := dataproto.NetworkInterfaceUpdateReq[dataproto.GcpNICreateExt]{
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
				tmpRes.Extension = &dataproto.GcpNICreateExt{
					CanIpForward: item.Extension.CanIpForward,
					Status:       item.Extension.Status,
					StackType:    item.Extension.StackType,
				}
				// 网卡私网IP信息列表
				var tmpAccConfigs []*dataproto.AccessConfig
				for _, accConfigItem := range item.Extension.AccessConfigs {
					tmpAccConfigs = append(tmpAccConfigs, &dataproto.AccessConfig{
						Name:        accConfigItem.Name,
						NatIP:       accConfigItem.NatIP,
						NetworkTier: accConfigItem.NetworkTier,
						Type:        accConfigItem.Type,
					})
				}
				tmpRes.Extension.AccessConfigs = tmpAccConfigs
			}

			updateResources = append(updateResources, tmpRes)
		} else {
			// need add resource data
			tmpRes := dataproto.NetworkInterfaceReq[dataproto.GcpNICreateExt]{
				AccountID:     req.AccountID,
				Vendor:        string(enumor.Gcp),
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
				if item.Extension != nil {
					tmpRes.Extension = &dataproto.GcpNICreateExt{
						CanIpForward: item.Extension.CanIpForward,
						Status:       item.Extension.Status,
						StackType:    item.Extension.StackType,
					}
					// 网卡私网IP信息列表
					var tmpAccConfigs []*dataproto.AccessConfig
					for _, accConfigItem := range item.Extension.AccessConfigs {
						tmpAccConfigs = append(tmpAccConfigs, &dataproto.AccessConfig{
							Name:        accConfigItem.Name,
							NatIP:       accConfigItem.NatIP,
							NetworkTier: accConfigItem.NetworkTier,
							Type:        accConfigItem.Type,
						})
					}
					tmpRes.Extension.AccessConfigs = tmpAccConfigs
				}
			}

			createResources = append(createResources, tmpRes)
		}
	}

	return createResources, updateResources, nil
}

func isGcpChange(item typesniproto.GcpNI, dbInfo coreni.NetworkInterface[coreni.GcpNIExtension]) bool {
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
	extRet := checkGcpExt(item, dbInfo)
	if extRet {
		return true
	}
	return false
}

func checkGcpExt(item typesniproto.GcpNI, dbInfo coreni.NetworkInterface[coreni.GcpNIExtension]) bool {
	if item.Extension.CanIpForward != dbInfo.Extension.CanIpForward {
		return true
	}
	if item.Extension.Status != dbInfo.Extension.Status {
		return true
	}
	if item.Extension.StackType != dbInfo.Extension.StackType {
		return true
	}
	if !assert.IsStringSliceEqual(item.PublicIPv4, dbInfo.PublicIPv4) {
		return true
	}
	if !assert.IsStringSliceEqual(item.PublicIPv6, dbInfo.PublicIPv6) {
		return true
	}

	for _, remote := range item.Extension.AccessConfigs {
		for _, db := range dbInfo.Extension.AccessConfigs {
			if remote.Name != db.Name || remote.NatIP != db.NatIP || remote.NetworkTier != db.NetworkTier ||
				remote.Type != db.Type {
				return true
			}
		}
	}
	return false
}

// compareDeleteGcpNetworkInterfaceList compare and delete network interface list from db.
func compareDeleteGcpNetworkInterfaceList(kt *kit.Kit, req *hcservice.GcpNetworkInterfaceSyncReq,
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
					Value: enumor.Gcp,
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
				"err: %v", enumor.Gcp, offset, count, err)
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
		deleteIDs := GetNeedDeleteGcpNetworkInterfaceList(kt, req, deleteCloudIDMap)
		if len(deleteIDs) > 0 {
			err = BatchDeleteNetworkInterfaceByIDs(kt, deleteIDs, dataCli)
			if err != nil {
				logs.Errorf("%s-networkinterface batch compare db delete failed. deleteIDs: %v, err: %v",
					enumor.Gcp, deleteIDs, err)
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

// GetNeedDeleteGcpNetworkInterfaceList get need delete gcp network interface list
func GetNeedDeleteGcpNetworkInterfaceList(_ *kit.Kit, _ *hcservice.GcpNetworkInterfaceSyncReq,
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

// GetCloudVpcInfoBySelfLink get vpc info by selflink
func GetCloudVpcInfoBySelfLink(kt *kit.Kit, req *hcservice.GcpNetworkInterfaceSyncReq, vendor enumor.Vendor,
	selfLink *string, dataCli *dataclient.Client) (*cloud.BaseVpc, error) {

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "account_id",
				Op:    filter.Equal.Factory(),
				Value: req.AccountID,
			},
			&filter.AtomRule{
				Field: "vendor",
				Op:    filter.Equal.Factory(),
				Value: vendor,
			},
			&filter.AtomRule{
				Field: "extension.self_link",
				Op:    filter.JSONIn.Factory(),
				Value: []*string{selfLink},
			},
		},
	}

	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}
	dbList, err := dataCli.Global.Vpc.List(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("get gcp vpc list from db failed, accountID: %s, vendor: %s, err: %v",
			req.AccountID, vendor, err)
		return nil, err
	}
	if len(dbList.Details) == 0 {
		logs.Errorf("get gcp vpc info is not found, accountID: %s, vendor: %s, selfLink: %s",
			req.AccountID, vendor, converter.PtrToVal(selfLink))
		return nil, errf.New(errf.RecordNotFound, "get gcp vpc info is not found.")
	}

	return &dbList.Details[0], nil
}

// GetCloudSubnetInfoBySelfLink get subnet info by selflink
func GetCloudSubnetInfoBySelfLink(kt *kit.Kit, req *hcservice.GcpNetworkInterfaceSyncReq, vendor enumor.Vendor,
	selfLink *string, dataCli *dataclient.Client) (*cloud.BaseSubnet, error) {

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "account_id",
				Op:    filter.Equal.Factory(),
				Value: req.AccountID,
			},
			&filter.AtomRule{
				Field: "vendor",
				Op:    filter.Equal.Factory(),
				Value: vendor,
			},
			&filter.AtomRule{
				Field: "extension.self_link",
				Op:    filter.JSONIn.Factory(),
				Value: []*string{selfLink},
			},
		},
	}

	dbQueryReq := &core.ListReq{
		Filter: expr,
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}
	dbList, err := dataCli.Global.Subnet.List(kt.Ctx, kt.Header(), dbQueryReq)
	if err != nil {
		logs.Errorf("get gcp subnet list from db failed, accountID: %s, vendor: %s, err: %v",
			req.AccountID, vendor, err)
		return nil, err
	}
	if len(dbList.Details) == 0 {
		logs.Errorf("get gcp subnet info is not found, accountID: %s, vendor: %s, selfLink: %s",
			req.AccountID, vendor, converter.PtrToVal(selfLink))
		return nil, errf.New(errf.RecordNotFound, "get gcp subnet info is not found.")
	}

	return &dbList.Details[0], nil
}

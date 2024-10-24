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

package huawei

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	typesni "hcm/pkg/adaptor/types/network-interface"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	dataproto "hcm/pkg/api/data-service"
	datacloud "hcm/pkg/api/data-service/cloud"
	datani "hcm/pkg/api/data-service/cloud/network-interface"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/concurrence"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SyncNIOption ...
type SyncNIOption struct {
}

// Validate ...
func (opt SyncNIOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// NetworkInterface 网络接口的同步依赖主机同步，db的网络接口是通过主机关联关系查询出来的。
func (cli *client) NetworkInterface(kt *kit.Kit, params *SyncBaseParams, opt *SyncNIOption) (*SyncResult, error) {

	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := concurrence.BaseExec(constant.SyncConcurrencyDefaultMaxLimit, params.CloudIDs, func(param string) error {
		syncOpt := &syncNIOption{
			AccountID:  params.AccountID,
			Region:     params.Region,
			CloudCvmID: param,
		}
		if _, err := cli.syncNetworkInterface(kt, syncOpt); err != nil {
			logs.ErrorDepthf(1, "[%s] account: %s cvm: %s sync network interface failed, err: %v, rid: %s",
				enumor.HuaWei, params.AccountID, param, err, kt.Rid)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return new(SyncResult), nil
}

type syncNIOption struct {
	AccountID  string `json:"account_id" validate:"required"`
	Region     string `json:"region" validate:"required"`
	CloudCvmID string `json:"cloud_cvm_id" validate:"required"`
}

// Validate ...
func (opt syncNIOption) Validate() error {
	return validator.Validate.Struct(opt)
}

func (cli *client) syncNetworkInterface(kt *kit.Kit, opt *syncNIOption) (*SyncResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	networkInterfaceFromCloud, err := cli.listNetworkInterfaceFromCloud(kt, opt.Region, opt.CloudCvmID)
	if err != nil {
		return nil, err
	}

	networkInterfaceFromDB, cvm, err := cli.listNetworkInterfaceFromDB(kt, opt)
	if err != nil {
		return nil, err
	}

	if len(networkInterfaceFromCloud) == 0 && len(networkInterfaceFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typesni.HuaWeiNI, coreni.
		NetworkInterface[coreni.HuaWeiNIExtension]](networkInterfaceFromCloud, networkInterfaceFromDB, isNIChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteNetworkInterface(kt, delCloudIDs, opt); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createNetworkInterface(kt, opt.AccountID, cvm, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateNetworkInterface(kt, opt.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (cli *client) deleteNetworkInterface(kt *kit.Kit, delCloudIDs []string, opt *syncNIOption) error {

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("delete network interfaces, network interfaces id is required")
	}

	niLists, err := cli.listNetworkInterfaceFromCloud(kt, opt.Region, opt.CloudCvmID)
	if err != nil {
		return err
	}

	delMap := converter.StringSliceToMap(delCloudIDs)
	for _, one := range niLists {
		if _, exist := delMap[*one.CloudID]; exist {
			logs.Errorf("[%s] validate networkInterface not exist failed, before delete, opt: %v, failed_count: %d, "+
				"rid: %s", enumor.HuaWei, opt, len(delCloudIDs), kt.Rid)
			return fmt.Errorf("validate networkInterface not exist failed, before delete")
		}
	}

	deleteReq := &dataproto.BatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.NetworkInterface.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete networkInterface failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return err
	}

	logs.V(3).Infof("[%s] sync network interface to delete ni success, accountID: %s, count: %d, rid: %s",
		enumor.HuaWei, opt.AccountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) updateNetworkInterface(kt *kit.Kit, accountID string, updateMap map[string]typesni.HuaWeiNI) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("update network interfaces, network interfaces is required")
	}

	if err := cli.completeNetworkInterfaceUpdateInfo(kt, updateMap); err != nil {
		return fmt.Errorf("complete network interface update info failed, err: %v", err)
	}

	lists := make([]datani.NetworkInterfaceUpdateReq[datani.HuaWeiNICreateExt], 0)
	for id, item := range updateMap {
		tmpRes := datani.NetworkInterfaceUpdateReq[datani.HuaWeiNICreateExt]{
			ID:            id,
			AccountID:     accountID,
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
			tmpRes.Extension = &datani.HuaWeiNICreateExt{
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
				Addresses:             (*datani.EipNetwork)(item.Extension.Addresses),
				CloudSecurityGroupIDs: slice.Unique(item.Extension.CloudSecurityGroupIDs),
			}
			// 网卡私网IP信息列表
			var tmpFixIps []datani.ServerInterfaceFixedIp
			for _, fixIpItem := range item.Extension.FixedIps {
				tmpFixIps = append(tmpFixIps, datani.ServerInterfaceFixedIp{
					IpAddress: fixIpItem.IpAddress,
					SubnetId:  fixIpItem.SubnetId,
				})
			}
			tmpRes.Extension.FixedIps = tmpFixIps

			var tmpVirtualIps []datani.NetVirtualIP
			for _, virtualIpItem := range item.Extension.VirtualIPList {
				tmpVirtualIps = append(tmpVirtualIps, datani.NetVirtualIP{
					IP:           virtualIpItem.IP,
					ElasticityIP: virtualIpItem.ElasticityIP,
				})
			}
			tmpRes.Extension.VirtualIPList = tmpVirtualIps
		}

		lists = append(lists, tmpRes)
	}

	updateReq := &datani.NetworkInterfaceBatchUpdateReq[datani.HuaWeiNICreateExt]{
		NetworkInterfaces: lists,
	}
	if err := cli.dbCli.HuaWei.NetworkInterface.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
		logs.Errorf("[%s] request dataservice huawei BatchUpdateNetworkInterface failed, err: %v, rid: %s",
			enumor.HuaWei, err, kt.Rid)
		return err
	}

	logs.V(3).Infof("[%s] sync network interface to update ni success, accountID: %s, count: %d, rid: %s",
		enumor.HuaWei, accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) completeNetworkInterfaceUpdateInfo(kt *kit.Kit, niMap map[string]typesni.HuaWeiNI) error {
	subnetCloudIDMap := make(map[string]struct{}, 0)
	for _, ni := range niMap {
		subnetCloudIDMap[*ni.CloudSubnetID] = struct{}{}
	}

	subnetMap, err := cli.getSubnetMapByCloudID(kt, converter.MapKeyToStringSlice(subnetCloudIDMap))
	if err != nil {
		return err
	}

	for _, ni := range niMap {
		cloudID := *ni.CloudSubnetID
		subnet, exist := subnetMap[cloudID]
		if !exist {
			return fmt.Errorf("subnet: %s not found", cloudID)
		}

		ni.CloudVpcID = converter.ValToPtr(subnet.CloudVpcID)
		ni.VpcID = converter.ValToPtr(subnet.VpcID)
		ni.SubnetID = converter.ValToPtr(subnet.ID)
	}

	return nil
}

func (cli *client) completeNetworkInterfaceCreateInfo(kt *kit.Kit, nis []typesni.HuaWeiNI) error {
	subnetCloudIDMap := make(map[string]struct{}, 0)
	for _, ni := range nis {
		subnetCloudIDMap[*ni.CloudSubnetID] = struct{}{}
	}

	subnetMap, err := cli.getSubnetMapByCloudID(kt, converter.MapKeyToStringSlice(subnetCloudIDMap))
	if err != nil {
		return err
	}

	for index := range nis {
		cloudID := *nis[index].CloudSubnetID
		subnet, exist := subnetMap[cloudID]
		if !exist {
			return fmt.Errorf("subnet: %s not found", cloudID)
		}

		nis[index].CloudVpcID = converter.ValToPtr(subnet.CloudVpcID)
		nis[index].VpcID = converter.ValToPtr(subnet.VpcID)
		nis[index].SubnetID = converter.ValToPtr(subnet.ID)
	}

	return nil
}

func (cli *client) getSubnetMapByCloudID(kt *kit.Kit, cloudIDs []string) (map[string]cloud.BaseSubnet, error) {
	subnetMap := make(map[string]cloud.BaseSubnet, len(cloudIDs))
	split := slice.Split(cloudIDs, int(core.DefaultMaxPageLimit))
	for _, part := range split {
		req := &core.ListReq{
			Filter: tools.ContainersExpression("cloud_id", part),
			Page:   core.NewDefaultBasePage(),
		}
		result, err := cli.dbCli.Global.Subnet.List(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] list subnet failed, err: %v, rid: %s", enumor.HuaWei, err, kt.Rid)
			return nil, err
		}

		for _, one := range result.Details {
			subnetMap[one.CloudID] = one
		}
	}
	return subnetMap, nil
}

func (cli *client) createNetworkInterface(kt *kit.Kit, accountID string, cvm *corecvm.BaseCvm,
	addSlice []typesni.HuaWeiNI) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("create network interfaces, network interfaces is required")
	}

	if err := cli.completeNetworkInterfaceCreateInfo(kt, addSlice); err != nil {
		return fmt.Errorf("complete network interface create info failed, err: %v", err)
	}

	lists := make([]datani.NetworkInterfaceReq[datani.HuaWeiNICreateExt], 0)
	for _, item := range addSlice {
		tmpRes := datani.NetworkInterfaceReq[datani.HuaWeiNICreateExt]{
			AccountID:     accountID,
			Vendor:        string(enumor.HuaWei),
			Name:          converter.PtrToVal(item.Name),
			Region:        converter.PtrToVal(item.Region),
			Zone:          cvm.Zone,
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
			BkBizID:       cvm.BkBizID,
		}
		if item.Extension != nil {
			tmpRes.Extension = &datani.HuaWeiNICreateExt{
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
				Addresses:             (*datani.EipNetwork)(item.Extension.Addresses),
				CloudSecurityGroupIDs: slice.Unique(item.Extension.CloudSecurityGroupIDs),
			}
			// 网卡私网IP信息列表
			var tmpFixIps []datani.ServerInterfaceFixedIp
			for _, fixIpItem := range item.Extension.FixedIps {
				tmpFixIps = append(tmpFixIps, datani.ServerInterfaceFixedIp{
					IpAddress: fixIpItem.IpAddress,
					SubnetId:  fixIpItem.SubnetId,
				})
			}
			tmpRes.Extension.FixedIps = tmpFixIps

			var tmpVirtualIps []datani.NetVirtualIP
			for _, virtualIpItem := range item.Extension.VirtualIPList {
				tmpVirtualIps = append(tmpVirtualIps, datani.NetVirtualIP{
					IP:           virtualIpItem.IP,
					ElasticityIP: virtualIpItem.ElasticityIP,
				})
			}
			tmpRes.Extension.VirtualIPList = tmpVirtualIps
		}

		lists = append(lists, tmpRes)
	}

	createReq := &datani.NetworkInterfaceBatchCreateReq[datani.HuaWeiNICreateExt]{
		NetworkInterfaces: lists,
	}
	createResult, err := cli.dbCli.HuaWei.NetworkInterface.BatchCreate(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create network interface failed, err: %v, rid: %s",
			enumor.HuaWei, err, kt.Rid)
		return err
	}

	relLists := make([]datacloud.NetworkInterfaceCvmRelCreateReq, 0)
	for _, id := range createResult.IDs {
		relLists = append(relLists, datacloud.NetworkInterfaceCvmRelCreateReq{
			NetworkInterfaceID: id,
			CvmID:              cvm.ID,
		})
	}

	createRelReq := &datacloud.NetworkInterfaceCvmRelBatchCreateReq{
		Rels: relLists,
	}
	if err = cli.dbCli.Global.NetworkInterfaceCvmRel.BatchCreateNetworkCvmRels(kt.Ctx, kt.Header(), createRelReq); err != nil {
		logs.Errorf("[%s] request dataservice to create cvm_ni_rel failed, err: %v, rid: %s",
			enumor.HuaWei, err, kt.Rid)
		return err
	}

	logs.V(3).Infof("[%s] sync network interface to create ni success, accountID: %s, count: %d, rid: %s",
		enumor.HuaWei, accountID, len(addSlice), kt.Rid)

	return nil
}

func (cli *client) listNetworkInterfaceFromCloud(kt *kit.Kit, region, cloudCvmID string) ([]typesni.HuaWeiNI, error) {

	listOpt := &typesni.HuaWeiNIListOption{
		ServerID: cloudCvmID,
		Region:   region,
	}
	result, err := cli.cloudCli.ListNetworkInterface(kt, listOpt)
	if err != nil {
		logs.Errorf("[%s] list networkInterface from cloud failed, err: %v, rid: %s", enumor.HuaWei, err, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func (cli *client) listNetworkInterfaceFromDB(kt *kit.Kit, opt *syncNIOption) (
	[]coreni.NetworkInterface[coreni.HuaWeiNIExtension], *corecvm.BaseCvm, error) {

	if err := opt.Validate(); err != nil {
		return nil, nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listCvmReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "cloud_id", Op: filter.Equal.Factory(), Value: opt.CloudCvmID},
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: opt.AccountID},
				&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: opt.Region},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	cvmResult, err := cli.dbCli.Global.Cvm.ListCvm(kt, listCvmReq)
	if err != nil {
		logs.Errorf("[%s] list cvm failed, err: %v, rid: %s", enumor.HuaWei, err, kt.Rid)
		return nil, nil, err
	}

	if len(cvmResult.Details) == 0 {
		return nil, nil, fmt.Errorf("cvm: %s not found", opt.CloudCvmID)
	}

	listRelReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "cvm_id", Op: filter.Equal.Factory(), Value: cvmResult.Details[0].ID},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	relResult, err := cli.dbCli.Global.NetworkInterfaceCvmRel.ListNetworkCvmRels(kt, listRelReq)
	if err != nil {
		logs.Errorf("[%s] list ni_cvm_rel failed, err: %v, rid: %s", enumor.HuaWei, err, kt.Rid)
		return nil, nil, err
	}

	if len(relResult.Details) == 0 {
		return make([]coreni.NetworkInterface[coreni.HuaWeiNIExtension], 0), &cvmResult.Details[0], nil
	}

	niIDs := make([]string, 0, len(relResult.Details))
	for _, one := range relResult.Details {
		niIDs = append(niIDs, one.NetworkInterfaceID)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: opt.AccountID},
				&filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: niIDs},
				&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: opt.Region},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.HuaWei.NetworkInterface.ListNetworkInterfaceExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list networkInterface from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.HuaWei,
			err, opt.AccountID, req, kt.Rid)
		return nil, nil, err
	}

	return result.Details, &cvmResult.Details[0], nil
}

func isNIChange(item typesni.HuaWeiNI, dbInfo coreni.NetworkInterface[coreni.HuaWeiNIExtension]) bool {

	if dbInfo.Name != converter.PtrToVal(item.Name) {
		return true
	}

	if dbInfo.Region != converter.PtrToVal(item.Region) {
		return true
	}

	if dbInfo.CloudID != converter.PtrToVal(item.CloudID) {
		return true
	}

	if dbInfo.CloudSubnetID != converter.PtrToVal(item.CloudSubnetID) {
		return true
	}

	if dbInfo.InstanceID != converter.PtrToVal(item.InstanceID) {
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

	if isNIExtChange(item, dbInfo) {
		return true
	}

	return false
}

func isNIExtChange(item typesni.HuaWeiNI, dbInfo coreni.NetworkInterface[coreni.HuaWeiNIExtension]) bool {
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

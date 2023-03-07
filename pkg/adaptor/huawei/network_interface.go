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
	"strings"

	typesniproto "hcm/pkg/adaptor/types/network-interface"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	ecsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	eipmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v3/model"
	vpcmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
)

// ListNetworkInterface 查看网卡列表
// reference: https://support.huaweicloud.com/intl/zh-cn/api-ecs/ecs_02_0505.html
func (h *HuaWei) ListNetworkInterface(kt *kit.Kit, opt *typesniproto.HuaWeiNIListOption) (
	*typesniproto.HuaWeiInterfaceListResult, error) {

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return nil, err
	}

	req := new(ecsmodel.ListServerInterfacesRequest)
	req.ServerId = opt.ServerID
	resp, err := client.ListServerInterfaces(req)
	if err != nil {
		logs.Errorf("list huawei network interface failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	details := make([]typesniproto.HuaWeiNI, 0, len(*resp.InterfaceAttachments))
	vnicPortIDs := make([]string, 0, len(*resp.InterfaceAttachments))
	for _, item := range *resp.InterfaceAttachments {
		tmpPortID := converter.PtrToVal(item.PortId)
		vnicPortIDs = append(vnicPortIDs, tmpPortID)
		details = append(details, converter.PtrToVal(h.convertCloudNetworkInterface(kt, opt, &item)))
	}

	details = h.replenishEipInfo(kt, opt, vnicPortIDs, details)

	return &typesniproto.HuaWeiInterfaceListResult{
		Details: details,
	}, nil
}

// replenishEipInfo replenish eip info
func (h *HuaWei) replenishEipInfo(kt *kit.Kit, opt *typesniproto.HuaWeiNIListOption, vnicPortIDs []string,
	details []typesniproto.HuaWeiNI) []typesniproto.HuaWeiNI {

	// 获取弹性IP信息
	eipMap, err := h.GetPublicIpsMapByPortIDs(kt, &typesniproto.HuaWeiEipListOption{
		Region:      opt.Region,
		VnicPortIDs: vnicPortIDs,
	})
	if err != nil {
		logs.Errorf("list huawei public eips map failed, region: %s, vnicPortIDs: %v, err: %v, rid: %s",
			opt.Region, vnicPortIDs, err, kt.Rid)
	}

	for _, item := range details {
		if item.Extension == nil {
			continue
		}

		tmpPortID := converter.PtrToVal(item.Extension.PortId)
		if eipInfo, ok := eipMap[tmpPortID]; ok {
			item.Extension.Addresses = eipInfo
		}
	}

	return details
}

func (h *HuaWei) convertCloudNetworkInterface(kt *kit.Kit, opt *typesniproto.HuaWeiNIListOption,
	data *ecsmodel.InterfaceAttachment) *typesniproto.HuaWeiNI {

	if data == nil {
		return nil
	}

	v := &typesniproto.HuaWeiNI{
		CloudID:    data.PortId,                      // 网卡端口ID
		InstanceID: converter.ValToPtr(opt.ServerID), // 关联的实例ID
		Region:     converter.ValToPtr(opt.Region),   // 区域
		Extension: &coreni.HuaWeiNIExtension{
			// FixedIps 网卡私网IP信息列表。
			FixedIps: []coreni.ServerInterfaceFixedIp{},
			// MacAddr 网卡Mac地址信息。
			MacAddr: data.MacAddr,
			// PortState 网卡端口状态。
			PortState: data.PortState,
			// DeleteOnTermination 卸载网卡时，是否删除网卡。
			DeleteOnTermination: data.DeleteOnTermination,
			// DriverMode 从guest os中，网卡的驱动类型。可选值为virtio和hinic，默认为virtio
			DriverMode: data.DriverMode,
			// MinRate 网卡带宽下限。
			MinRate: data.MinRate,
			// MultiqueueNum 网卡多队列个数。
			MultiqueueNum: data.MultiqueueNum,
			// PciAddress 弹性网卡在Linux GuestOS里的BDF号
			PciAddress: data.PciAddress,
		},
	}

	// 网卡私网IP信息列表
	ipv4Map := make(map[string]bool, 0)
	ipv6Map := make(map[string]bool, 0)
	if data.FixedIps != nil {
		for _, tmpFi := range *data.FixedIps {
			v.Extension.FixedIps = append(v.Extension.FixedIps, coreni.ServerInterfaceFixedIp{
				SubnetId:  tmpFi.SubnetId,
				IpAddress: tmpFi.IpAddress,
			})
			if checkIsIPv4(*tmpFi.IpAddress) {
				v.PrivateIPv4 = append(v.PrivateIPv4, *tmpFi.IpAddress)
				ipv4Map[*tmpFi.IpAddress] = true
			} else {
				v.PrivateIPv6 = append(v.PrivateIPv6, *tmpFi.IpAddress)
				ipv6Map[*tmpFi.IpAddress] = true
			}
		}
	}
	v.CloudSubnetID = data.NetId
	v.Name = converter.ValToPtr(fmt.Sprintf("name:%s", converter.PtrToVal(data.PortId)))

	// get security groups by port id
	tmpNetID := converter.PtrToVal(data.NetId)
	securityGroupMap, virtualIPs, err := h.GetSecurityGroupsByNetID(kt, &typesniproto.HuaWeiPortInfoOption{
		Region:         opt.Region,
		NetID:          tmpNetID,
		IPv4AddressMap: ipv4Map,
		IPv6AddressMap: ipv6Map,
	})
	if err != nil {
		logs.Errorf("list huawei security group map failed, region: %s, tmpNetID: %s, err: %v, rid: %s",
			opt.Region, tmpNetID, err, kt.Rid)
	}
	if sgList, ok := securityGroupMap[tmpNetID]; ok {
		v.Extension.CloudSecurityGroupIDs = sgList
	}
	v.Extension.VirtualIPList = virtualIPs

	return v
}

// GetPublicIpsMapByPortIDs get public ips map
func (h *HuaWei) GetPublicIpsMapByPortIDs(kt *kit.Kit, opt *typesniproto.HuaWeiEipListOption) (
	map[string]*coreni.EipNetwork, error) {

	client, err := h.clientSet.eipV3Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req := new(eipmodel.ListPublicipsRequest)
	req.VnicPortId = converter.ValToPtr(opt.VnicPortIDs)
	resp, err := client.ListPublicips(req)
	if err != nil {
		logs.Errorf("list huawei eip failed, region: %s, err: %v, rid: %s", opt.Region, err, kt.Rid)
		return nil, err
	}

	vnicPortMap := make(map[string]*coreni.EipNetwork, len(*resp.Publicips))
	for _, item := range *resp.Publicips {
		if item.Vnic == nil || item.Vnic.PortId == nil {
			continue
		}

		tmpVnicPortID := converter.PtrToVal(item.Vnic.PortId)
		vnicPortMap[tmpVnicPortID] = &coreni.EipNetwork{
			IPVersion: item.IpVersion.Value(),
			// PublicIPAddress IP地址。
			PublicIPAddress: converter.PtrToVal(item.PublicIpAddress),
			// PublicIPV6Address IPV6地址。
			PublicIPV6Address: converter.PtrToVal(item.PublicIpv6Address),
			// BandwidthType 带宽类型，示例:5_bgp(全动态BGP)
			BandwidthType: converter.PtrToVal(item.PublicipPoolName),
		}
		if item.Bandwidth != nil {
			// 带宽ID
			vnicPortMap[tmpVnicPortID].BandwidthID = converter.PtrToVal(item.Bandwidth.Id)
			// 带宽大小
			vnicPortMap[tmpVnicPortID].BandwidthSize = converter.PtrToVal(item.Bandwidth.Size)
		}
	}

	return vnicPortMap, nil
}

// GetSecurityGroupsByPortID get security groups by port id
func (h *HuaWei) GetSecurityGroupsByPortID(kt *kit.Kit, opt *typesniproto.HuaWeiPortInfoOption) (
	map[string][]string, error) {

	client, err := h.clientSet.vpcClientV2(opt.Region)
	if err != nil {
		return nil, err
	}

	req := new(vpcmodel.ShowPortRequest)
	req.PortId = opt.PortID
	resp, err := client.ShowPort(req)
	if err != nil {
		logs.Errorf("list huawei port info failed, region: %s, portID: %s, err: %v, rid: %s",
			opt.Region, opt.PortID, err, kt.Rid)
		return nil, err
	}

	var portMap = make(map[string][]string, 0)
	if resp.Port == nil || len(resp.Port.SecurityGroups) == 0 {
		return portMap, nil
	}

	var tmpSgList = make([]string, 0)
	for _, sgID := range resp.Port.SecurityGroups {
		tmpSgList = append(tmpSgList, sgID)
	}
	portMap[resp.Port.Id] = tmpSgList

	return portMap, nil
}

// GetSecurityGroupsByNetID get security groups by net id
func (h *HuaWei) GetSecurityGroupsByNetID(kt *kit.Kit, opt *typesniproto.HuaWeiPortInfoOption) (
	map[string][]string, []coreni.NetVirtualIP, error) {

	client, err := h.clientSet.vpcClientV2(opt.Region)
	if err != nil {
		return nil, []coreni.NetVirtualIP{}, err
	}

	req := new(vpcmodel.ListPortsRequest)
	req.NetworkId = converter.ValToPtr(opt.NetID)
	resp, err := client.ListPorts(req)
	if err != nil {
		logs.Errorf("list huawei security group failed, region: %s, netID: %s, err: %v, rid: %s",
			opt.Region, opt.NetID, err, kt.Rid)
		return nil, []coreni.NetVirtualIP{}, err
	}

	if resp.Ports == nil {
		return nil, []coreni.NetVirtualIP{}, nil
	}

	var virtualIPs = make([]coreni.NetVirtualIP, 0, len(*resp.Ports))
	for _, portItem := range *resp.Ports {
		if len(portItem.AllowedAddressPairs) == 0 || len(portItem.FixedIps) == 0 {
			continue
		}
		for _, item := range portItem.AllowedAddressPairs {
			if _, ok := opt.IPv4AddressMap[item.IpAddress]; ok {
				for _, fixIpItem := range portItem.FixedIps {
					virtualIPs = append(virtualIPs, coreni.NetVirtualIP{
						IP: converter.PtrToVal(fixIpItem.IpAddress),
					})
				}
			}
			if _, ok := opt.IPv6AddressMap[item.IpAddress]; ok {
				for _, fixIpItem := range portItem.FixedIps {
					virtualIPs = append(virtualIPs, coreni.NetVirtualIP{
						IP: converter.PtrToVal(fixIpItem.IpAddress),
					})
				}
			}
		}
	}

	var (
		securityGroupMap = make(map[string][]string, 0)
		tmpSgList        = make([]string, 0)
	)
	for _, portItem := range *resp.Ports {
		if len(portItem.SecurityGroups) == 0 {
			continue
		}
		for _, sgID := range portItem.SecurityGroups {
			tmpSgList = append(tmpSgList, sgID)
		}
		securityGroupMap[portItem.NetworkId] = tmpSgList
	}

	return securityGroupMap, virtualIPs, nil
}

func checkIsIPv4(ip string) bool {
	if len(ip) == 0 {
		return false
	}

	ipArr := strings.Split(ip, ".")
	return len(ipArr) == 4
}

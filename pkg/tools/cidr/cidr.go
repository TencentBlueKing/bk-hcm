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

// Package cidr ...
package cidr

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"regexp"
	"sort"

	"hcm/pkg/criteria/enumor"
)

// IsSubnetContained 判断父网络是否包含子网络
func IsSubnetContained(parent, child string) error {
	_, parentNet, err := net.ParseCIDR(parent)
	if err != nil {
		return fmt.Errorf("failed to parse parent subnet: %w", err)
	}

	_, childNet, err := net.ParseCIDR(child)
	if err != nil {
		return fmt.Errorf("failed to parse child subnet: %w", err)
	}

	if parentNet.Contains(childNet.IP) {
		maskSizeParent, _ := parentNet.Mask.Size()
		maskSizeChild, _ := childNet.Mask.Size()
		if maskSizeChild >= maskSizeParent {
			return nil
		}
	}

	return fmt.Errorf("cidr[%s] not belong cidr[%s]", child, parent)
}

// CidrIPAddressType get cidr ip address type.
func CidrIPAddressType(cidr string) (enumor.IPAddressType, error) {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}

	if ip.To4().Equal(ip) {
		return enumor.Ipv4, nil
	}

	if ip.To16().Equal(ip) {
		return enumor.Ipv6, nil
	}

	return "", fmt.Errorf("%s ip address type is invalid", ip)
}

// CidrIPCounts get ip counts by cidr
func CidrIPCounts(cidr string) (int, error) {
	_, net, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0, err
	}

	ones, bits := net.Mask.Size()
	hostBits := bits - ones
	totalIPs := 1 << uint(hostBits)

	return totalIPs - 2, nil
}

// IpNumToMasklen calculate the netmask len by number of ip, if number of ip less then 4 treat as 4.
func IpNumToMasklen(ipnum int) int {
	// calculat ceil(log2(x)), if x < 4 , treat x as 4
	if ipnum <= 4 {
		return 30
	}
	return 32 - int(math.Ceil(math.Log2(float64(ipnum))))
}

// NextAvailableNetByIpNum find next available net
// Params:
// 1. outer: 待分配的网段
// 2. used: outer中已经分配出去的子网
// 3. IPNum: 待分配子网所需的ip数量（包括网络号和主机号）
func NextAvailableNetByIpNum(outer net.IPNet, used []net.IPNet, ipNum int) (net.IPNet, error) {
	return NextAvailableNet(outer, used, IpNumToMasklen(ipNum))
}

// NextAvailableNet find next available net
// Params:
// 1. outer: 待分配的网段
// 2. used: outer中已经分配出去的子网
// 3. masklen: 待分配的网段掩码长度
func NextAvailableNet(outer net.IPNet, used []net.IPNet, masklen int) (net.IPNet, error) {
	// 给定一个网段、该网段中已分配的子网段列表，以及需要的新网段的掩码长度，返回下一个可用的网段。
	// '下一个可用网段'为已用网段中最后一个网段的下一个可用网段。
	var nextAvailable net.IPNet
	outerMasklen, _ := outer.Mask.Size()
	if masklen < outerMasklen {
		return nextAvailable, errors.New("new net mask length is shorter than outer net")
	}
	nextAvailable.Mask = net.CIDRMask(masklen, 32)
	// cidr排序，降序，要求给定的CIDR之间不相交
	sort.Slice(used, func(i, j int) bool { return bytes.Compare(used[i].IP, used[j].IP) > 0 })
	var lastBlock *net.IPNet
	for _, u := range used {
		// 剔除范围外的地址块
		if !outer.Contains(u.IP) {
			continue
		}
		lastBlock = &u
		break
	}

	if lastBlock == nil {
		// 该地址空间未分配有效网段
		// 直接分配主ip地址段即可
		nextAvailable.IP = outer.IP
		return nextAvailable, nil

	}
	// lastBlock 的下一个就是结果
	// 计算下一个网段，如果当前网络掩码长度比预期的长，则剪短掩码（网段范围变大），然后计算下一个。如果比较短则直接按当前掩码计算下一个
	lastMasklen, _ := lastBlock.Mask.Size()
	if lastMasklen > masklen {
		// 剪短掩码
		lastBlock.Mask = nextAvailable.Mask
		lastMasklen = masklen
	}

	// 先mask自己，剪掉多余的网络号
	lastBlock.IP = lastBlock.IP.Mask(lastBlock.Mask)

	// 网络号+1，计算下一个网络
	// 转换为整数形式方便计算
	netid := binary.BigEndian.Uint32(lastBlock.IP)
	netid >>= (32 - lastMasklen)
	netid += 1
	netid <<= (32 - lastMasklen)
	nextAvailable.IP = make(net.IP, 4)
	binary.BigEndian.PutUint32(nextAvailable.IP, netid)

	// 检查是否超出范围
	if !outer.Contains(nextAvailable.IP) {
		return net.IPNet{}, errors.New("out of range")
	}
	return nextAvailable, nil

}

// IsIPv4 检查字符串是否包含 IPv4 地址
func IsIPv4(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() != nil
}

// IsIPv6 检查字符串是否包含 IPv6 地址
func IsIPv6(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() == nil && ip.To16() != nil
}

// IsDomainName 检查字符串是否包含域名
func IsDomainName(s string) bool {
	// 使用正则表达式检查域名
	domainRegex := `^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(domainRegex, s)
	return matched
}

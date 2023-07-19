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

package cidr

import (
	"fmt"
	"net"

	"hcm/pkg/criteria/enumor"
)

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

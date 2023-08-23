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
	"testing"
)

type NextAvailableNetResult struct {
	NetStr string
	Err    error
}

func TestNextAvailableNet(t *testing.T) {

	_, testNet, _ := net.ParseCIDR("*.*.*.0/24")
	usedNetStr := []string{
		// "*.*.*.0/24",
		"*.*.*.0/28",
		"*.*.*.144/28",
		"*.*.*.96/27",
		"*.*.*.64/29",
		"*.*.*.32/27",
		"*.*.*.80/28",
		"*.*.*.128/29",
	}
	usedNetList := make([]net.IPNet, len(usedNetStr))

	result := []NextAvailableNetResult{
		// 23
		{"", fmt.Errorf("new net mask length is shorter than outer net")},
		// 24
		{"", fmt.Errorf("out of range")},
		// 25
		{"", fmt.Errorf("out of range")},
		{"*.*.*.192/26", nil},
		{"*.*.*.160/27", nil},
		{"*.*.*.144/28", nil},
		{"*.*.*.136/29", nil},
		{"*.*.*.136/30", nil},
		{"*.*.*.136/31", nil},
	}
	for idx, netStr := range usedNetStr {
		_, _net, _ := net.ParseCIDR(netStr)
		usedNetList[idx] = *_net
	}
	for i := range result {
		t.Run(fmt.Sprint("used-", i+23), func(t *testing.T) {

			next, err := NextAvailableNet(*testNet, usedNetList, 23+i)

			if err != nil && result[i].Err != nil && err.Error() == result[i].Err.Error() {
				return
			}
			if err != nil || next.String() != result[i].NetStr {
				t.Errorf("got next=%v,err=%v, except=%v, except err=%v", next.String(), err, result[i].NetStr, result[i].Err)
			}

		})

	}

}

func TestNextAvailableNet2(t *testing.T) {

	_, testNet, _ := net.ParseCIDR("*.*.*.0/24")
	result := []NextAvailableNetResult{
		// 23
		{"", fmt.Errorf("new net mask length is shorter than outer net")},
		{"*.*.*.0/24", nil},
		{"*.*.*.0/25", nil},
		{"*.*.*.0/26", nil},
		{"*.*.*.0/27", nil},
		{"*.*.*.0/28", nil},
		{"*.*.*.0/29", nil},
		{"*.*.*.0/30", nil},
		{"*.*.*.0/31", nil},
	}
	for i := range result {
		t.Run(fmt.Sprint("unused", i+23), func(t *testing.T) {

			next, err := NextAvailableNet(*testNet, []net.IPNet{}, 23+i)

			if err != nil && result[i].Err != nil && err.Error() == result[i].Err.Error() {
				return
			}
			if err != nil || next.String() != result[i].NetStr {
				t.Errorf("got next=%v,err=%v, except=%v, except err=%v", next.String(), err, result[i].NetStr, result[i].Err)
			}

		})

	}
}

func TestIpNumToMasklen(t *testing.T) {
	resutl := []int{
		// 0-4
		30, 30, 30, 30, 30,
		// 5-8
		29, 29, 29, 29,
		// 9-10
		28, 28,
	}
	for i, ans := range resutl {
		t.Run(fmt.Sprint("ipnum= ", i), func(t *testing.T) {
			if got := IpNumToMasklen(i); ans != got {
				t.Errorf("except %v got %v", ans, got)
			}
		})

	}
}

func TestNextAvailableNetByIPNum(t *testing.T) {
	_, testNet, _ := net.ParseCIDR("*.*.*.0/24")
	usedNetStr := []string{
		// "*.*.*.0/24",
		"*.*.*.0/28",
		"*.*.*.144/28",
		"*.*.*.96/27",
		"*.*.*.64/29",
		"*.*.*.32/27",
		"*.*.*.80/28",
		"*.*.*.128/29",
	}
	usedNetList := make([]net.IPNet, len(usedNetStr))

	result := []NextAvailableNetResult{
		// 23
		{"", fmt.Errorf("new net mask length is shorter than outer net")},
		// 24
		{"", fmt.Errorf("out of range")},
		// 25
		{"", fmt.Errorf("out of range")},
		{"*.*.*.192/26", nil},
		{"*.*.*.160/27", nil},
		{"*.*.*.144/28", nil},
		{"*.*.*.136/29", nil},
		{"*.*.*.136/30", nil},
		{"*.*.*.136/30", nil},
		{"*.*.*.136/30", nil},
		{"*.*.*.136/30", nil},
	}
	for idx, netStr := range usedNetStr {
		_, _net, _ := net.ParseCIDR(netStr)
		usedNetList[idx] = *_net
	}
	for i, num := range []int{512, 255, 126, 61, 28, 11, 7, 4, 3, 2, 1} {
		t.Run(fmt.Sprint(num), func(t *testing.T) {

			next, err := NextAvailableNetByIpNum(*testNet, usedNetList, num)

			if err != nil && result[i].Err != nil && err.Error() == result[i].Err.Error() {
				return
			}
			if err != nil || next.String() != result[i].NetStr {
				t.Errorf("got next=%v,err=%v, except=%v, except err=%v", next.String(), err, result[i].NetStr, result[i].Err)
			}

		})

	}
}

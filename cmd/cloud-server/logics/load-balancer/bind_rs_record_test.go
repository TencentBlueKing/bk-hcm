/*
 *
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package lblogic

import (
	"strings"
	"testing"

	"hcm/pkg/criteria/enumor"

	"github.com/stretchr/testify/assert"
)

func TestBindRSRecord_validateRS(t *testing.T) {
	cases := []struct {
		input     *BindRSRecord
		expectErr bool
	}{
		{
			input: &BindRSRecord{
				RSIPs:   []string{"192.168.1.3"},
				RSPorts: []int{80},
				Weight:  []int{1},
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				RSIPs:   []string{"192.168.1.3"},
				RSPorts: []int{80, 8081},
				Weight:  []int{1, 2},
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				RSIPs:   []string{"192.168.1.3"},
				RSPorts: []int{80, 8081},
				Weight:  []int{1},
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				RSIPs:   []string{"192.168.1.3", "192.168.1.4"},
				RSPorts: []int{80},
				Weight:  []int{1},
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				RSIPs:   []string{"192.168.1.3", "192.168.1.4"},
				RSPorts: []int{80, 8081},
				Weight:  []int{1, 2},
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				RSIPs:   []string{"192.168.1.3", "192.168.1.4"},
				RSPorts: []int{80},
				Weight:  []int{1, 2},
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				RSIPs:   []string{"192.168.1.3", "192.168.1.4"},
				RSPorts: []int{80},
				Weight:  []int{1, 2, 3},
			},
			expectErr: true,
		},
		{
			input:     &BindRSRecord{},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				RSIPs: []string{"192.168.1.3", "192.168.1.4"},
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				RSPorts: []int{80},
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Weight: []int{1},
			},
			expectErr: true,
		},
		//EndPort cases
		{
			input: &BindRSRecord{
				HaveEndPort: true,
				RSIPs:       []string{"192.168.1.3"},
				RSPorts:     []int{80},
				Weight:      []int{1},
			},
			expectErr: true,
		},
		{ // 端口段 rsIP weight 不唯一
			input: &BindRSRecord{
				HaveEndPort: true,
				RSIPs:       []string{"192.168.1.3", "192.168.1.4"},
				RSPorts:     []int{80, 8081},
				VPorts:      []int{80, 8081},
				Weight:      []int{1, 2},
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				HaveEndPort: true,
				RSIPs:       []string{"192.168.1.3"},
				RSPorts:     []int{80, 8081},
				VPorts:      []int{80, 8080},
				Weight:      []int{1},
			},
			expectErr: true,
		},
		{ // 溢出
			input: &BindRSRecord{
				HaveEndPort: true,
				RSIPs:       []string{"192.168.1.3"},
				RSPorts:     []int{65535, 65536},
				VPorts:      []int{80, 81},
				Weight:      []int{1},
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				HaveEndPort: true,
				RSIPs:       []string{"192.168.1.3"},
				RSPorts:     []int{80, 8081},
				VPorts:      []int{80, 8081},
				Weight:      []int{1},
			},
			expectErr: false,
		},
	}

	for i, c := range cases {
		err := c.input.validateRS()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}
}

func TestBindRSRecord_validateProtocol(t *testing.T) {
	cases := []struct {
		input     *BindRSRecord
		expectErr bool
	}{
		{
			input: &BindRSRecord{
				Protocol: "TCP",
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				Protocol: "UDP",
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				Protocol: "HTTPS",
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				Protocol: "HTTPS",
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				Protocol: "grpc",
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol: "HTTP",
			},
			expectErr: false,
		},
	}

	for i, c := range cases {
		err := c.input.validateProtocol()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}

}

func TestBindRSRecord_validateWeight(t *testing.T) {
	cases := []struct {
		input     *BindRSRecord
		expectErr bool
	}{
		{
			input: &BindRSRecord{
				Weight: []int{1},
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				Weight: []int{101},
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Weight: []int{-1},
			},
			expectErr: true,
		},
	}

	for i, c := range cases {
		err := c.input.validateWeight()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}
}

func TestBindRSRecord_validateListenerName(t *testing.T) {
	cases := []struct {
		input     *BindRSRecord
		expectErr bool
	}{
		{
			input: &BindRSRecord{
				ListenerName: "test-rule",
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				ListenerName: strings.Repeat("a", 256),
			},
			expectErr: true,
		},
		{
			input:     &BindRSRecord{},
			expectErr: true,
		},
	}

	for i, c := range cases {
		err := c.input.validateListenerName()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}
}

func TestBindRSRecord_validateVIPsAndVPorts(t *testing.T) {
	cases := []struct {
		input     *BindRSRecord
		expectErr bool
	}{
		{
			input: &BindRSRecord{
				VIP:         "192.168.1.3",
				VPorts:      []int{80},
				HaveEndPort: false,
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				VIP:         "192.168.1.3",
				HaveEndPort: false,
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				VPorts:      []int{80},
				HaveEndPort: false,
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				VIP:         "192.168.1.3",
				VPorts:      []int{80, 100},
				HaveEndPort: false,
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				VIP:         "192.168.1.3",
				VPorts:      []int{80},
				HaveEndPort: true,
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				VIP:         "192.168.1.3",
				VPorts:      []int{80, 1000},
				HaveEndPort: true,
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				VIP:         "192.168.1.3",
				VPorts:      []int{80, 1000000},
				HaveEndPort: true,
			},
			expectErr: true,
		},
	}

	for i, c := range cases {
		err := c.input.validateVIPsAndVPorts()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}

}

func TestBindRSRecord_validateCertAndURL(t *testing.T) {
	cases := []struct {
		input     *BindRSRecord
		expectErr bool
	}{
		// HTTPS case
		{
			input: &BindRSRecord{
				Protocol: "HTTPS",
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol: "HTTPS",
				Domain:   "example.com",
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol: "HTTPS",
				URLPath:  "/api",
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol:   "HTTPS",
				Domain:     "example.com",
				ServerCert: []string{"server.crt"},
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol:   "HTTPS",
				Domain:     "example.com",
				URLPath:    "/api",
				ClientCert: "client.crt",
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol:   "HTTPS",
				Domain:     "example.com",
				URLPath:    "/api",
				ServerCert: []string{""},
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol:   "HTTPS",
				Domain:     "example.com",
				URLPath:    "/api",
				ServerCert: []string{"server.crt", "server.crt", "server.crt"},
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol:   "HTTPS",
				Domain:     "example.com",
				URLPath:    "/api",
				ServerCert: []string{"server.crt"},
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				Protocol:   "HTTPS",
				Domain:     "example.com",
				URLPath:    "/api",
				ServerCert: []string{"server.crt"},
				ClientCert: "client.crt",
			},
			expectErr: false,
		},
		// HTTP case
		{
			input: &BindRSRecord{
				Protocol: "HTTP",
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol: "HTTP",
				Domain:   "example.com",
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol: "HTTP",
				URLPath:  "/api",
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol:   "HTTP",
				Domain:     "example.com",
				URLPath:    "/api",
				ServerCert: []string{"server.crt"},
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol:   "HTTP",
				Domain:     "example.com",
				URLPath:    "/api",
				ClientCert: "client.crt",
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol:   "HTTP",
				Domain:     "example.com",
				URLPath:    "/api",
				ServerCert: []string{"server.crt"},
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol:   "HTTP",
				Domain:     "example.com",
				URLPath:    "/api",
				ServerCert: []string{"server.crt"},
				ClientCert: "client.crt",
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol: "HTTP",
				Domain:   "example.com",
				URLPath:  "/api",
			},
			expectErr: false,
		},
		// TCP cases
		{
			input: &BindRSRecord{
				Protocol: "TCP",
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				Protocol: "TCP",
				Domain:   "example.com",
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol: "TCP",
				URLPath:  "/api",
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol:   "TCP",
				ServerCert: []string{"server.crt"},
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol:   "TCP",
				ClientCert: "client.crt",
			},
			expectErr: true,
		},
		// UDP cases
		{
			input: &BindRSRecord{
				Protocol: "UDP",
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				Protocol: "UDP",
				Domain:   "example.com",
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol: "UDP",
				URLPath:  "/api",
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol:   "UDP",
				ServerCert: []string{"server.crt"},
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				Protocol:   "UDP",
				ClientCert: "client.crt",
			},
			expectErr: true,
		},
	}

	for i, c := range cases {
		err := c.input.validateCertAndURL()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}
}

func TestBindRSRecord_validateInstType(t *testing.T) {
	cases := []struct {
		input     *BindRSRecord
		expectErr bool
	}{
		{
			input: &BindRSRecord{
				InstType: "ENI",
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				InstType: "CVM",
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				InstType: "Unknown Type",
			},
			expectErr: true,
		},
		{
			input:     &BindRSRecord{},
			expectErr: true,
		},
	}

	for i, c := range cases {
		err := c.input.validateInstType()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}
}

func TestBindRSRecord_validateScheduler(t *testing.T) {
	cases := []struct {
		input     *BindRSRecord
		expectErr bool
	}{
		{
			input: &BindRSRecord{
				Scheduler: "WRR",
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				Scheduler: "LEAST_CONN",
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				Scheduler: "Unknown Type",
			},
			expectErr: true,
		},
		{
			input:     &BindRSRecord{},
			expectErr: true,
		},
	}

	for i, c := range cases {
		err := c.input.validateScheduler()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}
}

func TestBindRSRecord_validateSessionExpired(t *testing.T) {
	cases := []struct {
		input     *BindRSRecord
		expectErr bool
	}{
		{
			input: &BindRSRecord{
				SessionExpired: -1,
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				SessionExpired: 10,
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				SessionExpired: 100000000,
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				SessionExpired: 30,
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				SessionExpired: 50,
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				SessionExpired: 0,
			},
			expectErr: false,
		},
	}

	for i, c := range cases {
		err := c.input.validateSessionExpired()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}
}

func TestBindRSRecord_validateRSInfos(t *testing.T) {
	cases := []struct {
		input     *BindRSRecord
		expectErr bool
	}{
		{
			input: &BindRSRecord{
				RSInfos: []*RSInfo{
					{
						IP:   "192.168.1.3",
						Port: 8080,
					},
				},
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				RSInfos: []*RSInfo{
					{
						IP:   "192.168.1.3",
						Port: 8080,
					},
					{
						IP:   "192.168.1.3",
						Port: 8080,
					},
				},
			},
			expectErr: true,
		},
	}

	for i, c := range cases {
		err := c.input.validateRSInfoDuplicate()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}
}

func TestGetIPVersion(t *testing.T) {
	cases := []struct {
		input       string
		expectValue enumor.IPAddressType
	}{
		{
			input:       "192.168.1.1",
			expectValue: enumor.Ipv4,
		},
		{
			input:       "2001:db8::1",
			expectValue: enumor.Ipv6,
		},
		{
			input:       "192.168.1.1:80",
			expectValue: "",
		},
	}

	for i, c := range cases {
		value, _ := getIPVersion(c.input)
		assert.Equalf(t, c.expectValue, value, "case %d failed", i)
	}
}

func TestBindRSRecord_validateIpDomainType(t *testing.T) {
	cases := []struct {
		input     *BindRSRecord
		expectErr bool
	}{
		{
			input: &BindRSRecord{
				IPDomainType: ipDomainTypeIPv4,
				VIP:          "192.168.1.1",
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				IPDomainType: ipDomainTypeIPv6,
				VIP:          "2408:8722:840:f8::83",
			},
			expectErr: false,
		},
		{
			input: &BindRSRecord{
				IPDomainType: ipDomainTypeDomain,
				VIP:          "example.com",
			},
			expectErr: false,
		},
		// bad cases
		{
			input: &BindRSRecord{
				IPDomainType: ipDomainTypeIPv4,
				VIP:          "2408:8722:840:f8::83",
			},
			expectErr: true,
		},
		{
			input: &BindRSRecord{
				IPDomainType: ipDomainTypeIPv6,
				VIP:          "192.168.1.1",
			},
			expectErr: true,
		},
	}

	for i, c := range cases {

		err := c.input.validateIpDomainType()
		if c.expectErr {
			assert.Error(t, err, "case %d failed", i)
		} else {
			assert.NoError(t, err, "case %d failed", i)
		}
	}
}

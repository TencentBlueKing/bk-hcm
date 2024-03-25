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

package cloud

import (
	"hcm/cmd/data-service/service/audit/cloud/cvm"
	"hcm/cmd/data-service/service/audit/cloud/firewall"
	loadbalancer "hcm/cmd/data-service/service/audit/cloud/load-balancer"
	networkinterface "hcm/cmd/data-service/service/audit/cloud/network-interface"
	routetable "hcm/cmd/data-service/service/audit/cloud/route-table"
	securitygroup "hcm/cmd/data-service/service/audit/cloud/security-group"
	"hcm/cmd/data-service/service/audit/cloud/subnet"
	"hcm/pkg/dal/dao"
)

// NewCloudAudit new audit svc.
func NewCloudAudit(dao dao.Set) *Audit {
	return &Audit{
		dao:              dao,
		securityGroup:    securitygroup.NewSecurityGroup(dao),
		firewall:         firewall.NewFirewall(dao),
		cvm:              cvm.NewCvm(dao),
		subnet:           subnet.NewSubnet(dao),
		networkInterface: networkinterface.NewNetworkInterface(dao),
		routeTable:       routetable.NewRouteTable(dao),
		loadBalancer:     loadbalancer.NewLoadBalancer(dao),
	}
}

// Audit define cloud audit.
type Audit struct {
	dao              dao.Set
	securityGroup    *securitygroup.SecurityGroup
	firewall         *firewall.Firewall
	cvm              *cvm.Cvm
	subnet           *subnet.Subnet
	networkInterface *networkinterface.NetworkInterface
	routeTable       *routetable.RouteTable
	loadBalancer     *loadbalancer.LoadBalancer
}

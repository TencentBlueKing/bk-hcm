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

package ziyan

import (
	ziyan "hcm/pkg/adaptor/tcloud-ziyan"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/kit"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
)

// Interface support resource sync.
type Interface interface {
	CloudCli() ziyan.TCloudZiyan

	// SecurityGroup 同步安全组
	SecurityGroup(kt *kit.Kit, params *SyncBaseParams, opt *SyncSGOption) (*SyncResult, error)
	RemoveSecurityGroupDeleteFromCloud(kt *kit.Kit, accountID string, region string) error
	RemoveSecurityGroupDeleteFromCloudV2(kt *kit.Kit, params *SyncRemovedParams,
		allCloudIDMap map[string]struct{}) error
	SecurityGroupRule(kt *kit.Kit, params *SyncBaseParams, opt *SyncSGRuleOption) (*SyncResult, error)
	SecurityGroupUsageBiz(kt *kit.Kit, params *SyncSGUsageBizParams) error

	// Subnet 同步子网
	Subnet(kt *kit.Kit, params *SyncBaseParams, opt *SyncSubnetOption) (*SyncResult, error)
	RemoveSubnetDeleteFromCloud(kt *kit.Kit, accountID string, region string) error

	// Vpc 同步VPC
	Vpc(kt *kit.Kit, params *SyncBaseParams, opt *SyncVpcOption) (*SyncResult, error)
	RemoveVpcDeleteFromCloud(kt *kit.Kit, accountID string, region string) error

	Zone(kt *kit.Kit, opt *SyncZoneOption) (*SyncResult, error)
	Region(kt *kit.Kit, opt *SyncRegionOption) (*SyncResult, error)

	ArgsTplAddress(kt *kit.Kit, params *SyncBaseParams, opt *SyncArgsTplOption) (*SyncResult, error)
	RemoveArgsTplAddressDeleteFromCloud(kt *kit.Kit, accountID string, region string) error
	ArgsTplAddressGroup(kt *kit.Kit, params *SyncBaseParams, opt *SyncArgsTplOption) (*SyncResult, error)
	RemoveArgsTplAddressGroupDeleteFromCloud(kt *kit.Kit, accountID string, region string) error
	ArgsTplService(kt *kit.Kit, params *SyncBaseParams, opt *SyncArgsTplOption) (*SyncResult, error)
	RemoveArgsTplServiceDeleteFromCloud(kt *kit.Kit, accountID string, region string) error
	ArgsTplServiceGroup(kt *kit.Kit, params *SyncBaseParams, opt *SyncArgsTplOption) (*SyncResult, error)
	RemoveArgsTplServiceGroupDeleteFromCloud(kt *kit.Kit, accountID string, region string) error

	Cert(kt *kit.Kit, params *SyncBaseParams, opt *SyncCertOption) (*SyncResult, error)
	RemoveCertDeleteFromCloud(kt *kit.Kit, accountID string, region string) error

	LoadBalancer(kt *kit.Kit, params *SyncBaseParams, opt *SyncLBOption) (*SyncResult, error)
	RemoveLoadBalancerDeleteFromCloud(kt *kit.Kit, params *SyncRemovedParams) error
	RemoveLoadBalancerDeleteFromCloudV2(kt *kit.Kit, param *SyncRemovedParams, allCloudIDMap map[string]struct{}) error

	// LoadBalancerWithListener 同步负载均衡及监听器
	LoadBalancerWithListener(kt *kit.Kit, params *SyncBaseParams, opt *SyncLBOption) (*SyncResult, error)

	// Listener 同步指定负载均衡下的指定云id 负载均衡
	Listener(kt *kit.Kit, params *SyncBaseParams, opt *SyncListenerOption) (*SyncResult, error)
	RemoveListenerDeleteFromCloud(kt *kit.Kit, params *ListenerSyncRemovedParams) error

	RemoveHostFromCC(kt *kit.Kit, params *DelHostParams) error
	HostWithRelRes(kt *kit.Kit, params *SyncHostParams) (*SyncResult, error)

	// Image 同步镜像
	Image(kt *kit.Kit, params *SyncBaseParams, opt *SyncImageOption) (*SyncResult, error)
	RemoveImageDeleteFromCloud(kt *kit.Kit, accountID string, region string) error

	// DeviceType 同步机型
	DeviceType(kt *kit.Kit, params *SyncDeviceTypeParams) (*SyncResult, error)
	// RemoveDeviceTypeDeleteFromCloud 删除云上已不存在的机型
	RemoveDeviceTypeDeleteFromCloud(kt *kit.Kit, params *SyncRemovedParams, allCloudIDMap map[string]struct{}) error
}

var _ Interface = new(client)

type client struct {
	accountID string
	cloudCli  ziyan.TCloudZiyan
	dbCli     *dataservice.Client
	// 引入esb 是为了同步时是去cc查询业务信息，后期考虑将clb 业务同步改到CloudServer中异步执行
	cmdbCli cmdb.Client
}

// CloudCli return tcloud client.
func (cli *client) CloudCli() ziyan.TCloudZiyan {
	return cli.cloudCli
}

// NewClient new sync client.
func NewClient(dbCli *dataservice.Client, cloudCli ziyan.TCloudZiyan, cmdbCli cmdb.Client) Interface {
	// 获取 cmdb
	return &client{
		cloudCli: cloudCli,
		dbCli:    dbCli,
		cmdbCli:  cmdbCli,
	}
}

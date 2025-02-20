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

package aws

import (
	"strings"

	"hcm/pkg/adaptor/aws"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/kit"
)

// Interface support resource sync.
type Interface interface {
	CloudCli() *aws.Aws

	Cvm(kt *kit.Kit, params *SyncBaseParams, opt *SyncCvmOption) (*SyncResult, error)
	CvmWithRelRes(kt *kit.Kit, params *SyncBaseParams, opt *SyncCvmWithRelResOption) (*SyncResult, error)
	RemoveCvmDeleteFromCloud(kt *kit.Kit, accountID string, region string) error

	Disk(kt *kit.Kit, params *SyncBaseParams, opt *SyncDiskOption) (*SyncResult, error)
	RemoveDiskDeleteFromCloud(kt *kit.Kit, accountID string, region string) error

	Eip(kt *kit.Kit, params *SyncBaseParams, opt *SyncEipOption) (*SyncResult, error)
	RemoveEipDeleteFromCloud(kt *kit.Kit, accountID string, region string) error

	RouteTable(kt *kit.Kit, params *SyncBaseParams, opt *SyncRouteTableOption) (*SyncResult, error)
	RemoveRouteTableDeleteFromCloud(kt *kit.Kit, accountID string, region string) error

	SecurityGroup(kt *kit.Kit, params *SyncBaseParams, opt *SyncSGOption) (*SyncResult, error)
	RemoveSecurityGroupDeleteFromCloud(kt *kit.Kit, accountID string, region string) error
	SecurityGroupUsageBiz(kt *kit.Kit, params *SyncSGUsageBizParams) error

	Subnet(kt *kit.Kit, params *SyncBaseParams, opt *SyncSubnetOption) (*SyncResult, error)
	RemoveSubnetDeleteFromCloud(kt *kit.Kit, accountID string, region string) error

	Image(kt *kit.Kit, params *SyncBaseParams, opt *SyncImageOption) (*SyncResult, error)
	RemoveImageDeleteFromCloud(kt *kit.Kit, accountID string, region string) error

	Vpc(kt *kit.Kit, params *SyncBaseParams, opt *SyncVpcOption) (*SyncResult, error)
	RemoveVpcDeleteFromCloud(kt *kit.Kit, accountID string, region string) error

	SecurityGroupRule(kt *kit.Kit, params *SyncBaseParams, opt *SyncSGRuleOption) (*SyncResult, error)

	Route(kt *kit.Kit, params *SyncBaseParams, opt *SyncRouteOption) (*SyncResult, error)

	Zone(kt *kit.Kit, opt *SyncZoneOption) (*SyncResult, error)

	Region(kt *kit.Kit, opt *SyncRegionOption) (*SyncResult, error)

	SubAccount(kt *kit.Kit, opt *SyncSubAccountOption) (*SyncResult, error)
}

var _ Interface = new(client)

// NewClient new client.
func NewClient(dbCli *dataservice.Client, cloudCli *aws.Aws) Interface {
	return &client{
		dbCli:    dbCli,
		cloudCli: cloudCli,
	}
}

type client struct {
	accountID string
	cloudCli  *aws.Aws
	dbCli     *dataservice.Client
}

// CloudCli ...
func (cli *client) CloudCli() *aws.Aws {
	return cli.cloudCli
}

func removeNotFoundCloudID(cloudIDs []string, err error) ([]string, string) {
	errStr := err.Error()
	delCloudID := errStr[strings.Index(errStr, "'")+1 : strings.LastIndex(errStr, "'")]

	result := make([]string, 0, len(cloudIDs)-1)
	for _, one := range cloudIDs {
		if one != delCloudID {
			result = append(result, one)
		}
	}

	return result, delCloudID
}

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

package hcziyancli

import (
	"hcm/pkg/rest"
)

// Client is a tcloud api client
type Client struct {
	Account       *AccountClient
	SecurityGroup *SecurityGroupClient
	Zone          *ZoneClient
	Region        *RegionClient
	ArgsTpl       *ArgsTplClient
	Cvm           *CvmClient
	Image         *ImageClient
	Application   *ApplicationClient
	Cert          *CertClient
	Clb           *ClbClient
	Subnet        *SubnetClient
	Vpc           *VpcClient
	BandPkg       *BandwidthPackageClient
	Tag           *TagClient
	Cos           *CosClient
	DeviceType    *DeviceTypeClient
}

// NewClient create a new tcloud api client.
func NewClient(client rest.ClientInterface) *Client {
	return &Client{
		Account:       NewAccountClient(client),
		SecurityGroup: NewCloudSecurityGroupClient(client),
		Zone:          NewZoneClient(client),
		Region:        NewRegionClient(client),
		ArgsTpl:       NewArgsTplClient(client),
		Cvm:           NewCvmClient(client),
		Image:         NewCloudPublicClient(client),
		Application:   NewApplicationClient(client),
		Cert:          NewCertClient(client),
		Clb:           NewClbClient(client),
		Vpc:           NewVpcClient(client),
		Subnet:        NewSubnetClient(client),
		BandPkg:       NewBandPkgClient(client),
		Tag:           NewTagClient(client),
		Cos:           NewCosClient(client),
		DeviceType:    NewDeviceTypeClient(client),
	}
}

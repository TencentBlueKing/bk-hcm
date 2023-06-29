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

package tcloudtypes

import (
	"fmt"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	"hcm/pkg/ad/provider"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
	"hcm/pkg/tools/cidr"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

// Vpc define
type Vpc struct {
	CloudID string `json:"cloud_id"`
	Name    string `json:"name"`
	Region  string `json:"region"`

	// IPv4Cidr 创建使用字段。
	IPv4Cidr string `json:"ipv4_cidr"`

	Extension *corecloud.TCloudVpcExtension `json:"extension"`
}

// ConvProviderVpc conv provider vpc.
func (v *Vpc) ConvProviderVpc() (*provider.Vpc, error) {
	ext, err := json.NewExtMessage(v.Extension)
	if err != nil {
		return nil, err
	}

	return &provider.Vpc{
		CloudID:    v.CloudID,
		Name:       v.Name,
		Region:     v.Region,
		ExtMessage: ext,
	}, nil
}

// CreateValidate vpc.
func (v Vpc) CreateValidate() error {
	// TODO: 添加校验
	return nil
}

// UpdateValidate vpc.
func (v Vpc) UpdateValidate() error {
	// TODO: 添加校验
	return nil
}

// ParseProviderVpc parse provider vpc.
func ParseProviderVpc(source *provider.Vpc) (*Vpc, error) {
	vpc := &Vpc{
		Name:     source.Name,
		Region:   source.Region,
		IPv4Cidr: "",
	}

	err := source.ExtMessage.UnmarshalObject(vpc)
	if err != nil {
		return nil, fmt.Errorf("unmarshal extMessage to vpc failed, err: %v", err)
	}

	return vpc, nil
}

// ParseCloudVpc parse cloud vpc.
func ParseCloudVpc(one *vpc.Vpc, region string) Vpc {
	tmp := Vpc{
		CloudID: converter.PtrToVal(one.VpcId),
		Name:    converter.PtrToVal(one.VpcName),
		Region:  region,
		Extension: &corecloud.TCloudVpcExtension{
			Cidr:            nil,
			IsDefault:       converter.PtrToVal(one.IsDefault),
			EnableMulticast: converter.PtrToVal(one.EnableMulticast),
			DnsServerSet:    converter.PtrToSlice(one.DnsServerSet),
			DomainName:      converter.PtrToVal(one.DomainName),
		},
	}

	if one.CidrBlock != nil && len(*one.CidrBlock) != 0 {
		tmp.Extension.Cidr = append(tmp.Extension.Cidr, corecloud.TCloudCidr{
			Type:     enumor.Ipv4,
			Cidr:     *one.CidrBlock,
			Category: enumor.MasterTCloudCidr,
		})
	}

	if one.Ipv6CidrBlock != nil && len(*one.Ipv6CidrBlock) != 0 {
		tmp.Extension.Cidr = append(tmp.Extension.Cidr, corecloud.TCloudCidr{
			Type:     enumor.Ipv6,
			Cidr:     *one.Ipv6CidrBlock,
			Category: enumor.MasterTCloudCidr,
		})
	}

	for _, asstCidr := range one.AssistantCidrSet {
		if asstCidr == nil {
			continue
		}

		cidrBlock := converter.PtrToVal(asstCidr.CidrBlock)
		addressType, err := cidr.CidrIPAddressType(cidrBlock)
		if err != nil {
			logs.Errorf("get cidr ip address type failed, cidr: %v, err: %v", cidrBlock, err)
		}

		tcloudCidr := corecloud.TCloudCidr{
			Type: addressType,
			Cidr: cidrBlock,
		}

		switch converter.PtrToVal(asstCidr.AssistantType) {
		case 0:
			tcloudCidr.Category = enumor.AssistantTCloudCidr
		case 1:
			tcloudCidr.Category = enumor.ContainerTCloudCidr
		}

		tmp.Extension.Cidr = append(tmp.Extension.Cidr, tcloudCidr)
	}

	return tmp
}

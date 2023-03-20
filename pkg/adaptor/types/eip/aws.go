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

package eip

import (
	"hcm/pkg/criteria/validator"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// AwsEipListOption ...
type AwsEipListOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"omitempty"`
	Ips      []string `json:"ips" validate:"omitempty"`
}

// Validate ...
func (o *AwsEipListOption) Validate() error {
	return validator.Validate.Struct(o)
}

// AwsEipListResult ...
type AwsEipListResult struct {
	Details []*AwsEip
}

// AwsEip ...
type AwsEip struct {
	CloudID        string
	Name           *string
	Region         string
	InstanceId     *string
	Status         *string
	PublicIp       *string
	PrivateIp      *string
	PublicIpv4Pool *string
	Domain         *string
	AssociationId  *string
}

// AwsEipDeleteOption ...
type AwsEipDeleteOption struct {
	Region  string `json:"region" validate:"required"`
	CloudID string `json:"cloud_id" validate:"required"`
}

// Validate ...
func (opt *AwsEipDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToReleaseAddressInput ...
func (opt *AwsEipDeleteOption) ToReleaseAddressInput() (*ec2.ReleaseAddressInput, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}
	return &ec2.ReleaseAddressInput{AllocationId: aws.String(opt.CloudID)}, nil
}

// AwsEipAssociateOption ...
type AwsEipAssociateOption struct {
	Region     string `json:"region" validate:"required"`
	CloudCvmID string `json:"cloud_cvm_id" validate:"required"`
	PublicIp   string `json:"public_ip" validate:"required"`
}

// Validate ...
func (opt *AwsEipAssociateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToAssociateAddressInput ...
func (opt *AwsEipAssociateOption) ToAssociateAddressInput() (*ec2.AssociateAddressInput, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	return &ec2.AssociateAddressInput{InstanceId: aws.String(opt.CloudCvmID), PublicIp: aws.String(opt.PublicIp)}, nil
}

// AwsEipDisassociateOption ...
type AwsEipDisassociateOption struct {
	Region   string `json:"region" validate:"required"`
	PublicIp string `json:"public_ip" validate:"required"`
}

// Validate ...
func (opt *AwsEipDisassociateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToDisassociateAddressInput ...
func (opt *AwsEipDisassociateOption) ToDisassociateAddressInput() (*ec2.DisassociateAddressInput, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}
	return &ec2.DisassociateAddressInput{}, nil
}

// AwsEipCreateOption ...
type AwsEipCreateOption struct {
	Region             string `json:"region" validate:"required"`
	PublicIpv4Pool     string `json:"public_ipv4_pool" validate:"required"`
	NetworkBorderGroup string `json:"network_border_group" validate:"required"`
}

// Validate ...
func (opt *AwsEipCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToAllocateAddressInput ...
func (opt *AwsEipCreateOption) ToAllocateAddressInput() (*ec2.AllocateAddressInput, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	return &ec2.AllocateAddressInput{
		PublicIpv4Pool:     aws.String(opt.PublicIpv4Pool),
		NetworkBorderGroup: aws.String(opt.NetworkBorderGroup),
	}, nil
}

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
	"hcm/pkg/tools/converter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v2/model"
)

// HuaWeiBandwidthShareTypeEnum ...
var HuaWeiBandwidthShareTypeEnum = map[string]model.CreatePublicipBandwidthOptionShareType{
	"WHOLE": model.GetCreatePublicipBandwidthOptionShareTypeEnum().WHOLE,
	"PER":   model.GetCreatePublicipBandwidthOptionShareTypeEnum().PER,
}

// HuaWeiBandwidthChargeModeTypeEnum ...
var HuaWeiBandwidthChargeModeTypeEnum = map[string]model.CreatePublicipBandwidthOptionChargeMode{
	"bandwidth": model.GetCreatePublicipBandwidthOptionChargeModeEnum().BANDWIDTH,
	"traffic":   model.GetCreatePublicipBandwidthOptionChargeModeEnum().TRAFFIC,
}

// HuaWeiBandwidthChargeModeTypeEnum ...
var HuaWeiChargeModeTypeEnum = map[string]model.CreatePrePaidPublicipExtendParamOptionChargeMode{
	"prePaid":  model.GetCreatePrePaidPublicipExtendParamOptionChargeModeEnum().PRE_PAID,
	"postPaid": model.GetCreatePrePaidPublicipExtendParamOptionChargeModeEnum().POST_PAID,
}

// HuaWeiPeriodTypeEnum ...
var HuaWeiPeriodTypeEnum = map[string]model.CreatePrePaidPublicipExtendParamOptionPeriodType{
	"month": model.GetCreatePrePaidPublicipExtendParamOptionPeriodTypeEnum().MONTH,
	"year":  model.GetCreatePrePaidPublicipExtendParamOptionPeriodTypeEnum().YEAR,
}

// HuaWeiEipListOption ...
type HuaWeiEipListOption struct {
	Region   string   `json:"region" validate:"required"`
	Limit    *int32   `json:"limit" validate:"omitempty"`
	Marker   *string  `json:"marker" validate:"omitempty"`
	CloudIDs []string `json:"cloud_ids" validate:"omitempty"`
	Ips      []string `json:"ips" validate:"omitempty"`
}

// Validate ...
func (o *HuaWeiEipListOption) Validate() error {
	return validator.Validate.Struct(o)
}

// HuaWeiEipListResult ...
type HuaWeiEipListResult struct {
	Details []*HuaWeiEip
}

// HuaWeiEip ...
type HuaWeiEip struct {
	CloudID             string
	Name                *string
	Region              string
	InstanceId          *string
	Status              *string
	PublicIp            *string
	PrivateIp           *string
	PortID              *string
	BandwidthId         *string
	BandwidthName       *string
	BandwidthSize       *int32
	EnterpriseProjectId *string
	Type                *string
	BandwidthShareType  string
	ChargeMode          string
}

// GetCloudID ...
func (eip *HuaWeiEip) GetCloudID() string {
	return eip.CloudID
}

// HuaWeiEipDeleteOption ...
type HuaWeiEipDeleteOption struct {
	CloudID string `json:"cloud_id" validate:"required"`
	Region  string `json:"region" validate:"required"`
}

// Validate ...
func (opt *HuaWeiEipDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToDeletePublicipRequest ...
func (opt *HuaWeiEipDeleteOption) ToDeletePublicipRequest() (*model.DeletePublicipRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}
	return &model.DeletePublicipRequest{PublicipId: opt.CloudID}, nil
}

// HuaWeiEipAssociateOption ...
type HuaWeiEipAssociateOption struct {
	CloudEipID              string `json:"cloud_eip_id" validate:"required"`
	CloudNetworkInterfaceID string `json:"cloud_network_interface_id" validate:"required"`
	Region                  string `json:"region" validate:"required"`
}

// Validate ...
func (opt *HuaWeiEipAssociateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToUpdatePublicipRequest ...
func (opt *HuaWeiEipAssociateOption) ToUpdatePublicipRequest() (*model.UpdatePublicipRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	req := &model.UpdatePublicipRequest{PublicipId: opt.CloudEipID}
	req.Body = &model.UpdatePublicipsRequestBody{
		Publicip: &model.UpdatePublicipOption{PortId: &opt.CloudNetworkInterfaceID},
	}
	return req, nil
}

// HuaWeiEipDisassociateOption ...
type HuaWeiEipDisassociateOption struct {
	CloudEipID string `json:"cloud_eip_id" validate:"required"`
	Region     string `json:"region" validate:"required"`
}

// Validate ...
func (opt *HuaWeiEipDisassociateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToUpdatePublicipRequest ...
func (opt *HuaWeiEipDisassociateOption) ToUpdatePublicipRequest() (*model.UpdatePublicipRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	req := &model.UpdatePublicipRequest{PublicipId: opt.CloudEipID, Body: &model.UpdatePublicipsRequestBody{
		Publicip: &model.UpdatePublicipOption{PortId: nil},
	}}
	return req, nil
}

// HuaWeiEipCreateOption ...
type HuaWeiEipCreateOption struct {
	Region                string                      `json:"region" validate:"required"`
	EipName               *string                     `json:"eip_name"`
	EipType               string                      `json:"eip_type"  validate:"required,eq=5_bgp|eq=5_sbgp"`
	EipCount              int64                       `json:"eip_count"  validate:"required"`
	InternetChargeType    string                      `json:"internet_charge_type" validate:"required,eq=prePaid|eq=postPaid"`
	InternetChargePrepaid *HuaWeiAddressChargePrepaid `json:"internet_charge_prepaid"`
	BandwidthOption       *HuaWeiBandwidthOption      `json:"bandwidth_option" validate:"required"`
}

// Validate ...
func (opt *HuaWeiEipCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToCreatePublicipRequest 按需付费参数
func (opt *HuaWeiEipCreateOption) ToCreatePublicipRequest() (*model.CreatePublicipRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	req := &model.CreatePublicipRequest{}
	req.Body = &model.CreatePublicipRequestBody{}

	bandwidthOpt := opt.BandwidthOption
	chargeMode := HuaWeiBandwidthChargeModeTypeEnum[bandwidthOpt.ChargeMode]
	req.Body.Bandwidth = &model.CreatePublicipBandwidthOption{
		ShareType:  HuaWeiBandwidthShareTypeEnum[bandwidthOpt.ShareType],
		ChargeMode: &chargeMode,
		Name:       bandwidthOpt.Name,
		Id:         bandwidthOpt.Id,
		Size:       bandwidthOpt.Size,
	}

	req.Body.Publicip = &model.CreatePublicipOption{Alias: opt.EipName, Type: opt.EipType}

	return req, nil
}

// ToCreatePrePaidPublicipRequest 包年/包月付费参数
func (opt *HuaWeiEipCreateOption) ToCreatePrePaidPublicipRequest() (*model.CreatePrePaidPublicipRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	req := &model.CreatePrePaidPublicipRequest{}
	req.Body = &model.CreatePrePaidPublicipRequestBody{}

	bandwidthOpt := opt.BandwidthOption
	chargeMode := HuaWeiBandwidthChargeModeTypeEnum[bandwidthOpt.ChargeMode]
	req.Body.Bandwidth = &model.CreatePublicipBandwidthOption{
		ShareType:  HuaWeiBandwidthShareTypeEnum[bandwidthOpt.ShareType],
		ChargeMode: &chargeMode,
		Name:       bandwidthOpt.Name,
		Id:         bandwidthOpt.Id,
		Size:       bandwidthOpt.Size,
	}

	req.Body.Publicip = &model.CreatePrePaidPublicipOption{Alias: opt.EipName, Type: opt.EipType}

	periodType := HuaWeiPeriodTypeEnum[opt.InternetChargePrepaid.PeriodType]
	prePaidChargeMode := HuaWeiChargeModeTypeEnum["prePaid"]
	req.Body.ExtendParam = &model.CreatePrePaidPublicipExtendParamOption{
		IsAutoRenew: &opt.InternetChargePrepaid.IsAutoRenew,
		PeriodNum:   &opt.InternetChargePrepaid.PeriodNum,
		PeriodType:  &periodType,
		ChargeMode:  &prePaidChargeMode,
		// **注意：现在默认设置为自动提交订单，后续形态需要支持订单的话需要调整此处**
		IsAutoPay: converter.ValToPtr(true),
	}

	return req, nil
}

// HuaWeiBandwidthOption ...
type HuaWeiBandwidthOption struct {
	ShareType  string  `json:"share_type" validate:"required,eq=PER|eq=WHOLE"`
	ChargeMode string  `json:"charge_mode" validate:"required,eq=bandwidth|eq=traffic"`
	Name       *string `json:"name"`
	Id         *string `json:"id"`
	Size       *int32  `json:"size"`
}

// HuaWeiAddressChargePrepaid ...
type HuaWeiAddressChargePrepaid struct {
	PeriodNum   int32  `json:"period_num"`
	PeriodType  string `json:"period_type"`
	IsAutoRenew bool   `json:"is_auto_renew"`
}

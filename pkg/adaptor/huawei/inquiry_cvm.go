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

package huawei

import (
	"fmt"

	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"

	bssmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bssintl/v2/model"
)

// InquiryPriceCvm 创建云主机询价
// reference: https://console-intl.huaweicloud.com/apiexplorer/#/openapi/BSSINTL/debug?api=ListRateOnPeriodDetail
// reference: https://console-intl.huaweicloud.com/apiexplorer/#/openapi/BSSINTL/debug?api=ListOnDemandResourceRatings
func (h *HuaWei) InquiryPriceCvm(kt *kit.Kit, opt *typecvm.HuaWeiCreateOption) (
	*typecvm.InquiryPriceResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "option is required")
	}

	projectID, err := h.GetProjectID(kt, opt.Region)
	if err != nil {
		logs.Errorf("get project id failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	switch opt.InstanceCharge.ChargingMode {
	case typecvm.PrePaid:
		return h.inquiryPricePrepaidCvm(kt, opt, projectID)
	case typecvm.PostPaid:
		return h.inquiryPricePostPaidCvm(kt, opt, projectID)
	default:
		return nil, fmt.Errorf("invalid charge type %s", opt.InstanceCharge.ChargingMode)
	}
}

func (h *HuaWei) inquiryPricePrepaidCvm(kt *kit.Kit, opt *typecvm.HuaWeiCreateOption, projectID string) (
	*typecvm.InquiryPriceResult, error) {

	client, err := h.clientSet.bssintlGlobalClient()
	if err != nil {
		return nil, err
	}

	periodType := int32(0)
	switch converter.PtrToVal(opt.InstanceCharge.PeriodType) {
	case typecvm.Year:
		periodType = 3
	case typecvm.Month:
		periodType = 2
	default:
		return nil, fmt.Errorf("invalid period type %s", converter.PtrToVal(opt.InstanceCharge.PeriodType))
	}

	infos := make([]bssmodel.PeriodProductInfo, 0)
	infos = append(infos, bssmodel.PeriodProductInfo{
		Id:               uuid.UUID(),
		CloudServiceType: "hws.service.type.ec2",
		ResourceType:     "hws.resource.type.vm",
		ResourceSpec:     opt.InstanceType,
		Region:           opt.Region,
		AvailableZone:    converter.ValToPtr(opt.Zone),
		PeriodType:       periodType,
		PeriodNum:        converter.PtrToVal(opt.InstanceCharge.PeriodNum),
		SubscriptionNum:  1,
	})

	infos = append(infos, bssmodel.PeriodProductInfo{
		Id:               uuid.UUID(),
		CloudServiceType: "hws.service.type.ebs",
		ResourceType:     "hws.resource.type.volume",
		ResourceSpec:     string(opt.RootVolume.VolumeType),
		Region:           opt.Region,
		AvailableZone:    converter.ValToPtr(opt.Zone),
		ResourceSize:     converter.ValToPtr(opt.RootVolume.SizeGB),
		SizeMeasureId:    converter.ValToPtr(int32(17)),
		PeriodType:       periodType,
		PeriodNum:        converter.PtrToVal(opt.InstanceCharge.PeriodNum),
		SubscriptionNum:  1,
	})

	for _, one := range opt.DataVolume {
		infos = append(infos, bssmodel.PeriodProductInfo{
			Id:               uuid.UUID(),
			CloudServiceType: "hws.service.type.ebs",
			ResourceType:     "hws.resource.type.volume",
			ResourceSpec:     string(one.VolumeType),
			Region:           opt.Region,
			AvailableZone:    converter.ValToPtr(opt.Zone),
			ResourceSize:     converter.ValToPtr(one.SizeGB),
			SizeMeasureId:    converter.ValToPtr(int32(17)),
			PeriodType:       periodType,
			PeriodNum:        converter.PtrToVal(opt.InstanceCharge.PeriodNum),
			SubscriptionNum:  1,
		})
	}

	req := &bssmodel.ListRateOnPeriodDetailRequest{
		Body: &bssmodel.RateOnPeriodReq{
			ProjectId:    projectID,
			ProductInfos: infos,
		},
	}
	resp, err := client.ListRateOnPeriodDetail(req)
	if err != nil {
		logs.Errorf("list rate on period detail failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := &typecvm.InquiryPriceResult{
		DiscountPrice: converter.PtrToVal(resp.OfficialWebsiteRatingResult.OfficialWebsiteAmount),
		OriginalPrice: 0,
	}

	return result, nil
}

func (h *HuaWei) inquiryPricePostPaidCvm(kt *kit.Kit, opt *typecvm.HuaWeiCreateOption, projectID string) (
	*typecvm.InquiryPriceResult, error) {

	client, err := h.clientSet.bssintlGlobalClient()
	if err != nil {
		return nil, err
	}

	infos := make([]bssmodel.DemandProductInfo, 0)
	infos = append(infos, bssmodel.DemandProductInfo{
		Id:               uuid.UUID(),
		CloudServiceType: "hws.service.type.ec2",
		ResourceType:     "hws.resource.type.vm",
		ResourceSpec:     opt.InstanceType,
		Region:           opt.Region,
		AvailableZone:    converter.ValToPtr(opt.Zone),
		UsageFactor:      "Duration",
		UsageValue:       1,
		UsageMeasureId:   4,
		SubscriptionNum:  1,
	})

	infos = append(infos, bssmodel.DemandProductInfo{
		Id:               uuid.UUID(),
		CloudServiceType: "hws.service.type.ebs",
		ResourceType:     "hws.resource.type.volume",
		ResourceSpec:     string(opt.RootVolume.VolumeType),
		Region:           opt.Region,
		AvailableZone:    converter.ValToPtr(opt.Zone),
		ResourceSize:     converter.ValToPtr(opt.RootVolume.SizeGB),
		SizeMeasureId:    converter.ValToPtr(int32(17)),
		UsageFactor:      "Duration",
		UsageValue:       1,
		UsageMeasureId:   4,
		SubscriptionNum:  1,
	})

	for _, one := range opt.DataVolume {
		infos = append(infos, bssmodel.DemandProductInfo{
			Id:               uuid.UUID(),
			CloudServiceType: "hws.service.type.ebs",
			ResourceType:     "hws.resource.type.volume",
			ResourceSpec:     string(one.VolumeType),
			Region:           opt.Region,
			AvailableZone:    converter.ValToPtr(opt.Zone),
			ResourceSize:     converter.ValToPtr(one.SizeGB),
			SizeMeasureId:    converter.ValToPtr(int32(17)),
			UsageFactor:      "Duration",
			UsageValue:       1,
			UsageMeasureId:   4,
			SubscriptionNum:  1,
		})
	}

	req := &bssmodel.ListOnDemandResourceRatingsRequest{
		Body: &bssmodel.RateOnDemandReq{
			ProjectId:    projectID,
			ProductInfos: infos,
		},
	}
	resp, err := client.ListOnDemandResourceRatings(req)
	if err != nil {
		logs.Errorf("list rate on period detail failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := &typecvm.InquiryPriceResult{
		DiscountPrice: converter.PtrToVal(resp.OfficialWebsiteAmount),
		OriginalPrice: 0,
	}

	return result, nil
}

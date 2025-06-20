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
	"errors"
	"fmt"
	"strconv"
	"strings"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/times"
	"hcm/pkg/tools/uuid"

	bssmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bssintl/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
)

const (
	// IPTypeFixed fixed：代表私有IP地址
	IPTypeFixed = "fixed"
	// IPTypeFloating floating：代表浮动IP地址。
	IPTypeFloating = "floating"

	// IPVersion4 ipv4
	IPVersion4 = "4"
	// IPVersion6 ipv6
	IPVersion6 = "6"
)

// ListCvm list cvm.
// reference: https://support.huaweicloud.com/api-ecs/zh-cn_topic_0094148850.html
func (h *HuaWei) ListCvm(kt *kit.Kit, opt *typecvm.HuaWeiListOption) ([]typecvm.HuaWeiCvm, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new ecs client failed, err: %v", err)
	}

	req := new(model.ListServersDetailsRequest)

	if len(opt.CloudIDs) != 0 {
		req.ServerId = converter.ValToPtr(strings.Join(opt.CloudIDs, ","))
	}

	if opt.Page != nil {
		req.Limit = converter.ValToPtr(opt.Page.Limit)
		req.Offset = converter.ValToPtr(opt.Page.Offset)
	}
	// 暂不支持裸金属服务器，因此此处屏蔽裸金属服务器
	// 注意，若指定了ServerID，该筛选条件会自动失效
	req.NotTags = converter.ValToPtr("__type_baremetal")

	resp, err := client.ListServersDetails(req)
	if err != nil {
		return nil, err
	}

	cvms := make([]typecvm.HuaWeiCvm, 0, len(converter.PtrToVal(resp.Servers)))
	for _, one := range converter.PtrToVal(resp.Servers) {

		inst := typecvm.HuaWeiCvm{ServerDetail: one}
		privateIPv4, publicIPv4, privateIPv6, publicIPv6 := getIps(one.Addresses)
		inst.PrivateIPv4Addresses = privateIPv4
		inst.PublicIPv4Addresses = publicIPv4
		inst.PrivateIPv6Addresses = privateIPv6
		inst.PublicIPv6Addresses = publicIPv6

		dataDiskIds := make([]string, 0)
		for _, v := range one.OsExtendedVolumesvolumesAttached {
			if converter.PtrToVal(v.BootIndex) == "0" {
				inst.CloudOSDiskID = v.Id
			} else {
				dataDiskIds = append(dataDiskIds, v.Id)
			}
		}
		inst.CLoudDataDiskIDs = dataDiskIds

		startTime, err := times.ParseToStdTime("2006-01-02T15:04:05.000000", one.OSSRVUSGlaunchedAt)
		if err != nil {
			logs.Errorf("[%s] conv LastStartTimestamp to std time failed, err: %v", enumor.HuaWei, err)
			return nil, err
		}
		inst.CloudLaunchedTime = startTime

		if one.Flavor != nil {
			flavor, err := h.convFlavor(kt, one.Flavor)
			if err != nil {
				logs.Errorf("conv huawei flavor failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
			inst.Flavor = flavor
		}
		cvms = append(cvms, inst)
	}

	return cvms, err
}

func (h *HuaWei) convFlavor(kt *kit.Kit, cloudFlavor *model.ServerFlavor) (*corecvm.HuaWeiFlavor, error) {
	ramInt, err := strconv.Atoi(cloudFlavor.Ram)
	if err != nil {
		logs.Errorf("convert huawei cvm ram to int failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	ram := strconv.Itoa(ramInt / 1024)
	flavor := &corecvm.HuaWeiFlavor{
		CloudID: cloudFlavor.Id,
		Name:    cloudFlavor.Name,
		Disk:    cloudFlavor.Disk,
		VCpus:   cloudFlavor.Vcpus,
		Ram:     ram,
	}
	return flavor, nil
}

func getIps(serverAddress map[string][]model.ServerAddress) (privateIPv4, publicIPv4, privateIPv6, publicIPv6 []string) {

	for _, addresses := range serverAddress {
		for _, addr := range addresses {

			if addr.Version == IPVersion4 && addr.OSEXTIPStype.Value() == IPTypeFixed {
				privateIPv4 = append(privateIPv4, addr.Addr)
			}
			if addr.Version == IPVersion4 && addr.OSEXTIPStype.Value() == IPTypeFloating {
				publicIPv4 = append(publicIPv4, addr.Addr)
			}
			if addr.Version == IPVersion6 && addr.OSEXTIPStype.Value() == IPTypeFixed {
				privateIPv6 = append(privateIPv6, addr.Addr)
			}
			if addr.Version == IPVersion6 && addr.OSEXTIPStype.Value() == IPTypeFloating {
				publicIPv6 = append(publicIPv6, addr.Addr)
			}
		}
	}

	return privateIPv4, publicIPv4, privateIPv6, publicIPv6
}

// DeleteCvm reference: https://support.huaweicloud.com/api-ecs/ecs_02_0103.html
func (h *HuaWei) DeleteCvm(kt *kit.Kit, opt *typecvm.HuaWeiDeleteOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new ecs client failed, err: %v", err)
	}

	svrIDs := make([]model.ServerId, 0, len(opt.CloudIDs))
	for _, one := range opt.CloudIDs {
		svrIDs = append(svrIDs, model.ServerId{
			Id: one,
		})
	}

	req := &model.DeleteServersRequest{
		Body: &model.DeleteServersRequestBody{
			DeletePublicip: converter.ValToPtr(opt.DeletePublicIP),
			DeleteVolume:   converter.ValToPtr(opt.DeleteVolume),
			Servers:        svrIDs,
		},
	}

	_, err = client.DeleteServers(req)
	if err != nil {
		logs.Errorf("delete huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return err
}

// StartCvm reference: https://support.huaweicloud.com/api-ecs/ecs_02_0301.html
func (h *HuaWei) StartCvm(kt *kit.Kit, opt *typecvm.HuaWeiStartOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "start option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new ecs client failed, err: %v", err)
	}

	svrIDs := make([]model.ServerId, 0, len(opt.CloudIDs))
	for _, one := range opt.CloudIDs {
		svrIDs = append(svrIDs, model.ServerId{
			Id: one,
		})
	}

	req := &model.BatchStartServersRequest{
		Body: &model.BatchStartServersRequestBody{
			OsStart: &model.BatchStartServersOption{
				Servers: svrIDs,
			},
		},
	}

	resp, err := client.BatchStartServers(req)
	if err != nil {
		logs.Errorf("batch start huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 判断批量操作任务是否失败
	handler := &jobPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*HuaWei, []model.SubJob, poller.BaseDoneResult]{Handler: handler}
	_, err = respPoller.PollUntilDone(h, kt, []*string{resp.JobId}, types.NewBatchOperateCvmPollerOpt())
	if err != nil {
		return err
	}

	// 等待主机状态改变
	startHandler := &startCvmPollingHandler{
		opt.Region,
	}
	startPoller := poller.Poller[*HuaWei, []model.ServerDetail, poller.BaseDoneResult]{Handler: startHandler}
	_, err = startPoller.PollUntilDone(h, kt, converter.SliceToPtr(opt.CloudIDs),
		types.NewBatchOperateCvmPollerOpt())
	if err != nil {
		return err
	}

	return err
}

// StopCvm reference: https://support.huaweicloud.com/api-ecs/ecs_02_0303.html
func (h *HuaWei) StopCvm(kt *kit.Kit, opt *typecvm.HuaWeiStopOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "stop option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new ecs client failed, err: %v", err)
	}

	svrIDs := make([]model.ServerId, 0, len(opt.CloudIDs))
	for _, one := range opt.CloudIDs {
		svrIDs = append(svrIDs, model.ServerId{
			Id: one,
		})
	}

	var stopType model.BatchStopServersOptionType
	if opt.Force {
		stopType = model.GetBatchStopServersOptionTypeEnum().SOFT
	} else {
		stopType = model.GetBatchStopServersOptionTypeEnum().HARD
	}

	req := &model.BatchStopServersRequest{
		Body: &model.BatchStopServersRequestBody{
			OsStop: &model.BatchStopServersOption{
				Type:    &stopType,
				Servers: svrIDs,
			},
		},
	}

	resp, err := client.BatchStopServers(req)
	if err != nil {
		logs.Errorf("batch stop huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 判断批量操作任务是否失败
	handler := &jobPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*HuaWei, []model.SubJob, poller.BaseDoneResult]{Handler: handler}
	_, err = respPoller.PollUntilDone(h, kt, []*string{resp.JobId}, types.NewBatchOperateCvmPollerOpt())
	if err != nil {
		return err
	}

	// 等待主机状态改变
	stopHandler := &stopCvmPollingHandler{
		opt.Region,
	}
	stopPoller := poller.Poller[*HuaWei, []model.ServerDetail, poller.BaseDoneResult]{Handler: stopHandler}
	_, err = stopPoller.PollUntilDone(h, kt, converter.SliceToPtr(opt.CloudIDs), types.NewBatchOperateCvmPollerOpt())
	if err != nil {
		return err
	}

	return err
}

// RebootCvm reference: https://support.huaweicloud.com/api-ecs/ecs_02_0302.html
func (h *HuaWei) RebootCvm(kt *kit.Kit, opt *typecvm.HuaWeiRebootOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "reboot option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new ecs client failed, err: %v", err)
	}

	svrIDs := make([]model.ServerId, 0, len(opt.CloudIDs))
	for _, one := range opt.CloudIDs {
		svrIDs = append(svrIDs, model.ServerId{
			Id: one,
		})
	}

	var rebootType model.BatchRebootSeversOptionType
	if opt.Force {
		rebootType = model.GetBatchRebootSeversOptionTypeEnum().SOFT
	} else {
		rebootType = model.GetBatchRebootSeversOptionTypeEnum().HARD
	}

	req := &model.BatchRebootServersRequest{
		Body: &model.BatchRebootServersRequestBody{
			Reboot: &model.BatchRebootSeversOption{
				Type:    rebootType,
				Servers: svrIDs,
			},
		},
	}

	resp, err := client.BatchRebootServers(req)
	if err != nil {
		logs.Errorf("batch reboot huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 判断批量操作任务是否失败
	handler := &jobPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*HuaWei, []model.SubJob, poller.BaseDoneResult]{Handler: handler}
	_, err = respPoller.PollUntilDone(h, kt, []*string{resp.JobId}, types.NewBatchOperateCvmPollerOpt())
	if err != nil {
		return err
	}

	// 等待主机状态改变
	rebootHandler := &rebootCvmPollingHandler{
		opt.Region,
	}
	rebootPoller := poller.Poller[*HuaWei, []model.ServerDetail, poller.BaseDoneResult]{Handler: rebootHandler}
	_, err = rebootPoller.PollUntilDone(h, kt, converter.SliceToPtr(opt.CloudIDs), types.NewBatchOperateCvmPollerOpt())
	if err != nil {
		return err
	}

	return err
}

// ResetCvmPwd reference: https://support.huaweicloud.com/api-ecs/ecs_02_0306.html
func (h *HuaWei) ResetCvmPwd(kt *kit.Kit, opt *typecvm.HuaWeiResetPwdOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "reset pwd option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new ecs client failed, err: %v", err)
	}

	svrIDs := make([]model.ServerId, 0, len(opt.CloudIDs))
	for _, one := range opt.CloudIDs {
		svrIDs = append(svrIDs, model.ServerId{
			Id: one,
		})
	}

	req := &model.BatchResetServersPasswordRequest{
		Body: &model.BatchResetServersPasswordRequestBody{
			NewPassword: opt.Password,
			Servers:     svrIDs,
		},
	}

	_, err = client.BatchResetServersPassword(req)
	if err != nil {
		logs.Errorf("batch reset pwd huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	handler := &resetpwdCvmPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*HuaWei, []model.ServerDetail, poller.BaseDoneResult]{Handler: handler}
	_, err = respPoller.PollUntilDone(h, kt, converter.SliceToPtr(opt.CloudIDs),
		types.NewBatchOperateCvmPollerOpt())
	if err != nil {
		return err
	}

	return err
}

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

// CreateCvm reference: https://support.huaweicloud.com/api-ecs/ecs_02_0101.html
func (h *HuaWei) CreateCvm(kt *kit.Kit, opt *typecvm.HuaWeiCreateOption) (*poller.BaseDoneResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "reset pwd option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new ecs client failed, err: %v", err)
	}

	volumeType, err := opt.RootVolume.VolumeType.RootVolumeType()
	if err != nil {
		return nil, err
	}

	chargingMode, err := opt.InstanceCharge.ChargingMode.ChargingMode()
	if err != nil {
		return nil, err
	}
	req := &model.CreateServersRequest{
		XClientToken: opt.ClientToken,
		Body: &model.CreateServersRequestBody{
			DryRun: converter.ValToPtr(opt.DryRun),
			Server: &model.PrePaidServer{
				ImageRef:  opt.CloudImageID,
				FlavorRef: opt.InstanceType,
				Name:      opt.Name,
				AdminPass: converter.ValToPtr(opt.Password),
				Vpcid:     opt.CloudVpcID,
				Nics: []model.PrePaidServerNic{
					{
						SubnetId: opt.CloudSubnetID,
					},
				},
				RootVolume: &model.PrePaidServerRootVolume{
					Volumetype: volumeType,
					Size:       converter.ValToPtr(opt.RootVolume.SizeGB),
				},
				Count:            converter.ValToPtr(opt.RequiredCount),
				AvailabilityZone: converter.ValToPtr(opt.Zone),
				Description:      opt.Description,
				Extendparam: &model.PrePaidServerExtendParam{
					ChargingMode: converter.ValToPtr(chargingMode),
					IsAutoRenew:  converter.ValToPtr(model.GetPrePaidServerExtendParamIsAutoRenewEnum().TRUE),
					IsAutoPay:    converter.ValToPtr(model.GetPrePaidServerExtendParamIsAutoPayEnum().TRUE),
				},
			},
		},
	}

	if opt.PublicIPAssigned {
		mode, err := opt.Eip.ChargingMode.EipChargingMode()
		if err != nil {
			return nil, err
		}

		req.Body.Server.Publicip = &model.PrePaidServerPublicip{
			Eip: &model.PrePaidServerEip{
				Iptype: string(opt.Eip.Type),
				Bandwidth: &model.PrePaidServerEipBandwidth{
					Size:      converter.ValToPtr(opt.Eip.Size),
					Sharetype: model.GetPrePaidServerEipBandwidthSharetypeEnum().PER,
				},
				Extendparam: &model.PrePaidServerEipExtendParam{
					ChargingMode: converter.ValToPtr(mode),
				},
			},
		}
	}

	if opt.InstanceCharge.PeriodType != nil {
		periodType, err := opt.InstanceCharge.PeriodType.PeriodType()
		if err != nil {
			return nil, err
		}
		req.Body.Server.Extendparam.PeriodType = converter.ValToPtr(periodType)
		req.Body.Server.Extendparam.PeriodNum = opt.InstanceCharge.PeriodNum
	}

	if opt.InstanceCharge.IsAutoRenew != nil {
		if *opt.InstanceCharge.IsAutoRenew {
			req.Body.Server.Extendparam.IsAutoRenew = converter.ValToPtr(
				model.GetPrePaidServerExtendParamIsAutoRenewEnum().TRUE)
		} else {
			req.Body.Server.Extendparam.IsAutoRenew = converter.ValToPtr(
				model.GetPrePaidServerExtendParamIsAutoRenewEnum().FALSE)
		}
	}

	if len(opt.CloudSecurityGroupIDs) != 0 {
		req.Body.Server.SecurityGroups = new([]model.PrePaidServerSecurityGroup)
		for _, sgID := range opt.CloudSecurityGroupIDs {
			*req.Body.Server.SecurityGroups = append(*req.Body.Server.SecurityGroups, model.PrePaidServerSecurityGroup{
				Id: converter.ValToPtr(sgID),
			})
		}
	}

	if len(opt.DataVolume) != 0 {
		req.Body.Server.DataVolumes = new([]model.PrePaidServerDataVolume)
		for _, one := range opt.DataVolume {
			volType, err := one.VolumeType.DataVolumeType()
			if err != nil {
				return nil, err
			}
			*req.Body.Server.DataVolumes = append(*req.Body.Server.DataVolumes, model.PrePaidServerDataVolume{
				Volumetype: volType,
				Size:       one.SizeGB,
			})
		}
	}

	resp, err := client.CreateServers(req)
	if err != nil {
		logs.Errorf("create huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if opt.DryRun {
		return new(poller.BaseDoneResult), nil
	}

	handler := &createCvmPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*HuaWei, []model.ServerDetail, poller.BaseDoneResult]{Handler: handler}
	result, err := respPoller.PollUntilDone(h, kt, converter.SliceToPtr(converter.PtrToVal(resp.ServerIds)),
		types.NewBatchCreateCvmPollerOption())
	if err != nil {
		return nil, err
	}

	return result, err
}

type jobPollingHandler struct {
	region string
}

// Done ...
func (h *jobPollingHandler) Done(jobs []model.SubJob) (bool, *poller.BaseDoneResult) {

	result := &poller.BaseDoneResult{
		SuccessCloudIDs: make([]string, 0),
		FailedCloudIDs:  make([]string, 0),
		UnknownCloudIDs: make([]string, 0),
		FailedMessage:   "",
	}
	for _, job := range jobs {
		if converter.PtrToVal(job.Status) == model.GetSubJobStatusEnum().RUNNING {
			return false, result
		}

		if converter.PtrToVal(job.Status) == model.GetSubJobStatusEnum().FAIL {
			result.FailedCloudIDs = append(result.FailedCloudIDs, converter.PtrToVal(job.Entities.ServerId))
			result.FailedMessage = converter.PtrToVal(job.FailReason)
		}

		if converter.PtrToVal(job.Status) == model.GetSubJobStatusEnum().SUCCESS {
			result.SuccessCloudIDs = append(result.SuccessCloudIDs, converter.PtrToVal(job.Entities.ServerId))
		}
	}

	return true, result
}

// Poll ...
func (h *jobPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]model.SubJob, error) {
	if len(cloudIDs) == 0 {
		return nil, errors.New("job id is required")
	}

	ecsCli, err := client.clientSet.ecsClient(h.region)
	if err != nil {
		logs.Errorf("new ecs client failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("new ecs client failed, err: %v", err)
	}

	req := &model.ShowJobRequest{
		JobId: *cloudIDs[0],
	}
	resp, err := ecsCli.ShowJob(req)
	if err != nil {
		logs.Errorf("show job failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return converter.PtrToVal(resp.Entities.SubJobs), nil
}

type startCvmPollingHandler struct {
	region string
}

// Done ...
func (h *startCvmPollingHandler) Done(cvms []model.ServerDetail) (bool, *poller.BaseDoneResult) {
	return done(cvms, "ACTIVE")
}

// Poll ...
func (h *startCvmPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]model.ServerDetail, error) {
	return poll(client, kt, h.region, cloudIDs)
}

type stopCvmPollingHandler struct {
	region string
}

// Done ...
func (h *stopCvmPollingHandler) Done(cvms []model.ServerDetail) (bool, *poller.BaseDoneResult) {
	return done(cvms, "SHUTOFF")
}

// Poll ...
func (h *stopCvmPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]model.ServerDetail, error) {
	return poll(client, kt, h.region, cloudIDs)
}

type resetpwdCvmPollingHandler struct {
	region string
}

// Done ...
func (h *resetpwdCvmPollingHandler) Done(cvms []model.ServerDetail) (bool, *poller.BaseDoneResult) {
	return done(cvms, "ACTIVE")
}

// Poll ...
func (h *resetpwdCvmPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) (
	[]model.ServerDetail, error) {

	return poll(client, kt, h.region, cloudIDs)
}

type rebootCvmPollingHandler struct {
	region string
}

// Done ...
func (h *rebootCvmPollingHandler) Done(cvms []model.ServerDetail) (bool, *poller.BaseDoneResult) {
	return done(cvms, "ACTIVE")
}

// Poll ...
func (h *rebootCvmPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]model.ServerDetail, error) {
	return poll(client, kt, h.region, cloudIDs)
}

func done(cvms []model.ServerDetail, succeed string) (bool, *poller.BaseDoneResult) {
	result := new(poller.BaseDoneResult)

	flag := true
	for _, instance := range cvms {
		// not done
		if instance.Status != succeed {
			flag = false
			continue
		}

		result.SuccessCloudIDs = append(result.SuccessCloudIDs, instance.Id)
	}

	return flag, result
}

func poll(client *HuaWei, kt *kit.Kit, region string, cloudIDs []*string) ([]model.ServerDetail, error) {
	cloudIDSplit := slice.Split(cloudIDs, core.HuaWeiQueryLimit)

	cvms := make([]model.ServerDetail, 0, len(cloudIDs))
	for _, partIDs := range cloudIDSplit {
		req := new(model.ListServersDetailsRequest)
		req.ServerId = converter.ValToPtr(strings.Join(converter.PtrToSlice(partIDs), ","))
		req.Limit = converter.ValToPtr(int32(core.HuaWeiQueryLimit))

		cvmCli, err := client.clientSet.ecsClient(region)
		if err != nil {
			return nil, err
		}

		resp, err := cvmCli.ListServersDetails(req)
		if err != nil {
			logs.Errorf("list servers detail failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		cvms = append(cvms, *resp.Servers...)
	}

	return cvms, nil
}

type createCvmPollingHandler struct {
	region string
}

// Done ...
func (h *createCvmPollingHandler) Done(cvms []model.ServerDetail) (bool, *poller.BaseDoneResult) {

	result := &poller.BaseDoneResult{
		SuccessCloudIDs: make([]string, 0),
		FailedCloudIDs:  make([]string, 0),
	}
	flag := true
	for _, instance := range cvms {
		// 创建中
		if instance.Status == "BUILD" {
			flag = false
			result.UnknownCloudIDs = append(result.UnknownCloudIDs, instance.Id)
			continue
		}

		// 生产失败
		if instance.Status == "ERROR" || instance.Status == "UNKNOWN" {
			result.FailedCloudIDs = append(result.FailedCloudIDs, instance.Id)
			result.FailedMessage = instance.Fault.String()
			continue
		}

		result.SuccessCloudIDs = append(result.SuccessCloudIDs, instance.Id)
	}

	return flag, result
}

// Poll ...
func (h *createCvmPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]model.ServerDetail, error) {

	cloudIDSplit := slice.Split(cloudIDs, core.HuaWeiQueryLimit)

	cvms := make([]model.ServerDetail, 0, len(cloudIDs))
	for _, partIDs := range cloudIDSplit {
		req := new(model.ListServersDetailsRequest)
		req.ServerId = converter.ValToPtr(strings.Join(converter.PtrToSlice(partIDs), ","))
		req.Limit = converter.ValToPtr(int32(core.HuaWeiQueryLimit))

		cvmCli, err := client.clientSet.ecsClient(h.region)
		if err != nil {
			return nil, err
		}

		resp, err := cvmCli.ListServersDetails(req)
		if err != nil {
			return nil, err
		}

		cvms = append(cvms, *resp.Servers...)
	}

	if len(cvms) != len(cloudIDs) {
		return nil, fmt.Errorf("query cvm count: %d not equal return count: %d", len(cloudIDs), len(cvms))
	}

	return cvms, nil
}

var _ poller.PollingHandler[*HuaWei, []model.ServerDetail, poller.BaseDoneResult] = new(createCvmPollingHandler)

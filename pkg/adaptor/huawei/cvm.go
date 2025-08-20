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

	req, err := buildCreateCvmReq(opt, chargingMode, volumeType)
	if err != nil {
		logs.Errorf("build create cvm request failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("build create cvm request failed, err: %v", err)
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

func buildCreateCvmReq(opt *typecvm.HuaWeiCreateOption, chargingMode model.PrePaidServerExtendParamChargingMode,
	volumeType model.PrePaidServerRootVolumeVolumetype) (*model.CreateServersRequest, error) {

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

	createCvmReqSetSecurityGroup(req, opt.CloudSecurityGroupIDs)
	if err := createCvmReqSetDataVolume(req, opt.DataVolume); err != nil {
		return nil, err
	}
	return req, nil
}

func createCvmReqSetSecurityGroup(req *model.CreateServersRequest, securityGroupIDs []string) {
	if len(securityGroupIDs) == 0 {
		return
	}
	req.Body.Server.SecurityGroups = new([]model.PrePaidServerSecurityGroup)
	for _, sgID := range securityGroupIDs {
		*req.Body.Server.SecurityGroups = append(*req.Body.Server.SecurityGroups, model.PrePaidServerSecurityGroup{
			Id: converter.ValToPtr(sgID),
		})
	}

}

func createCvmReqSetDataVolume(req *model.CreateServersRequest, dataVolumes []typecvm.HuaWeiVolume) error {
	if len(dataVolumes) == 0 {
		return nil
	}

	req.Body.Server.DataVolumes = new([]model.PrePaidServerDataVolume)
	for _, one := range dataVolumes {
		volType, err := one.VolumeType.DataVolumeType()
		if err != nil {
			return err
		}
		*req.Body.Server.DataVolumes = append(*req.Body.Server.DataVolumes, model.PrePaidServerDataVolume{
			Volumetype: volType,
			Size:       one.SizeGB,
		})
	}
	return nil
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

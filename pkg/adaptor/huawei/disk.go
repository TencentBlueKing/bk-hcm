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
	"strings"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/uuid"

	bssmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bssintl/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2/model"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

// CreateDisk 创建云硬盘
// reference: https://support.huaweicloud.com/api-evs/evs_04_2003.html
func (h *HuaWei) CreateDisk(kt *kit.Kit, opt *disk.HuaWeiDiskCreateOption) (*poller.BaseDoneResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "huawei disk create option is required")
	}

	resp, err := h.createDisk(opt)
	if err != nil {
		return nil, err
	}

	if resp.VolumeIds == nil || len(*resp.VolumeIds) == 0 {
		return nil, fmt.Errorf("create disk return volume_ids is empty, orderID: %v", converter.ValToPtr(resp.OrderId))
	}

	respPoller := poller.Poller[*HuaWei, []disk.HuaWeiDisk, poller.BaseDoneResult]{
		Handler: &createDiskPollingHandler{region: opt.Region},
	}
	return respPoller.PollUntilDone(h, kt, common.StringPtrs(*resp.VolumeIds), nil)
}

// InquiryPriceDisk 创建云硬盘询价
// reference: https://console-intl.huaweicloud.com/apiexplorer/#/openapi/BSSINTL/debug?api=ListRateOnPeriodDetail
// reference: https://console-intl.huaweicloud.com/apiexplorer/#/openapi/BSSINTL/debug?api=ListOnDemandResourceRatings
func (h *HuaWei) InquiryPriceDisk(kt *kit.Kit, opt *disk.HuaWeiDiskCreateOption) (
	*disk.InquiryPriceResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "option is required")
	}

	projectID, err := h.GetProjectID(kt, opt.Region)
	if err != nil {
		logs.Errorf("get project id failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	switch converter.PtrToVal(opt.DiskChargeType) {
	case disk.HuaWeiDiskChargeTypeEnum.PRE_PAID:
		return h.inquiryPricePrepaidDisk(kt, opt, projectID)
	case disk.HuaWeiDiskChargeTypeEnum.POST_PAID:
		return h.inquiryPricePostPaidDisk(kt, opt, projectID)
	default:
		return nil, fmt.Errorf("invalid charge type %s", converter.PtrToVal(opt.DiskChargeType))
	}
}

func (h *HuaWei) inquiryPricePostPaidDisk(kt *kit.Kit, opt *disk.HuaWeiDiskCreateOption, projectID string) (
	*disk.InquiryPriceResult, error) {

	client, err := h.clientSet.bssintlGlobalClient()
	if err != nil {
		return nil, err
	}

	req := &bssmodel.ListOnDemandResourceRatingsRequest{
		Body: &bssmodel.RateOnDemandReq{
			ProjectId: projectID,
			ProductInfos: []bssmodel.DemandProductInfo{
				{
					Id:               uuid.UUID(),
					CloudServiceType: "hws.service.type.ebs",
					ResourceType:     "hws.resource.type.volume",
					ResourceSpec:     opt.DiskType,
					Region:           opt.Region,
					AvailableZone:    converter.ValToPtr(opt.Zone),
					ResourceSize:     converter.ValToPtr(opt.DiskSize),
					SizeMeasureId:    converter.ValToPtr(int32(17)),
					UsageFactor:      "Duration",
					UsageValue:       1,
					UsageMeasureId:   4,
					SubscriptionNum:  1,
				},
			},
		},
	}
	resp, err := client.ListOnDemandResourceRatings(req)
	if err != nil {
		logs.Errorf("list rate on period detail failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := &disk.InquiryPriceResult{
		DiscountPrice: converter.PtrToVal(resp.OfficialWebsiteAmount),
		OriginalPrice: 0,
	}

	return result, nil
}

func (h *HuaWei) inquiryPricePrepaidDisk(kt *kit.Kit, opt *disk.HuaWeiDiskCreateOption, projectID string) (
	*disk.InquiryPriceResult, error) {

	client, err := h.clientSet.bssintlGlobalClient()
	if err != nil {
		return nil, err
	}

	periodType := int32(0)
	switch converter.PtrToVal(opt.DiskChargePrepaid.PeriodType) {
	case disk.HuaWeiDiskPeriodTypeEnum.YEAR:
		periodType = 3
	case disk.HuaWeiDiskPeriodTypeEnum.MONTH:
		periodType = 2
	default:
		return nil, fmt.Errorf("invalid period type %s", converter.PtrToVal(opt.DiskChargePrepaid.PeriodType))
	}

	req := &bssmodel.ListRateOnPeriodDetailRequest{
		Body: &bssmodel.RateOnPeriodReq{
			ProjectId: projectID,
			ProductInfos: []bssmodel.PeriodProductInfo{
				{
					Id:               uuid.UUID(),
					CloudServiceType: "hws.service.type.ebs",
					ResourceType:     "hws.resource.type.volume",
					ResourceSpec:     opt.DiskType,
					Region:           opt.Region,
					AvailableZone:    converter.ValToPtr(opt.Zone),
					ResourceSize:     converter.ValToPtr(opt.DiskSize),
					SizeMeasureId:    converter.ValToPtr(int32(17)),
					PeriodType:       periodType,
					PeriodNum:        converter.PtrToVal(opt.DiskChargePrepaid.PeriodNum),
					SubscriptionNum:  1,
				},
			},
		},
	}
	resp, err := client.ListRateOnPeriodDetail(req)
	if err != nil {
		logs.Errorf("list rate on period detail failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := &disk.InquiryPriceResult{
		DiscountPrice: converter.PtrToVal(resp.OfficialWebsiteRatingResult.OfficialWebsiteAmount),
		OriginalPrice: 0,
	}

	return result, nil
}

func (h *HuaWei) createDisk(opt *disk.HuaWeiDiskCreateOption) (*model.CreateVolumeResponse, error) {
	client, err := h.clientSet.evsClient(opt.Region)
	if err != nil {
		return nil, err
	}

	req, err := opt.ToCreateVolumeRequest()
	if err != nil {
		return nil, err
	}

	return client.CreateVolume(req)
}

// ListDisk 查看云硬盘
// reference: https://support.huaweicloud.com/api-evs/evs_04_2006.html
func (h *HuaWei) ListDisk(kt *kit.Kit, opt *disk.HuaWeiDiskListOption) ([]disk.HuaWeiDisk, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "huawei disk list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.evsClient(opt.Region)
	if err != nil {
		return nil, err
	}

	req := new(model.ListVolumesRequest)

	if opt.Page != nil {
		req.Marker = opt.Page.Marker
		req.Limit = opt.Page.Limit
	}

	if len(opt.CloudIDs) > 0 {
		req.Ids = converter.StringSliceToSliceStringPtr(opt.CloudIDs)
	}

	resp, err := client.ListVolumes(req)
	if err != nil {
		if strings.Contains(err.Error(), ErrDataNotFound) {
			return make([]disk.HuaWeiDisk, 0), nil
		}
		logs.Errorf("list huawei disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	disks, err := h.buildDisk(kt, *resp.Volumes)
	if err != nil {
		return nil, err
	}

	return disks, nil
}

func (h *HuaWei) buildDisk(kt *kit.Kit, details []model.VolumeDetail) ([]disk.HuaWeiDisk, error) {

	client, err := h.clientSet.bssintlGlobalClient()
	if err != nil {
		return nil, err
	}

	// 获取包年包月硬盘的计费信息
	ids := slice.FilterMap(details, func(one model.VolumeDetail) bool {
		_, exist := one.Metadata["orderID"]
		return exist
	}, func(one model.VolumeDetail) string { return one.Id })

	split := slice.Split(ids, 50)
	idChargeMap := make(map[string]bssmodel.OrderInstanceV2, len(ids))
	for _, partIDs := range split {
		req := &bssmodel.ListPayPerUseCustomerResourcesRequest{
			Body: &bssmodel.QueryResourcesReq{
				ResourceIds: converter.ValToPtr(partIDs),
			},
		}
		resp, err := client.ListPayPerUseCustomerResources(req)
		if err != nil {
			logs.Errorf("list pay per use customer resource failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, one := range converter.PtrToVal(resp.Data) {
			idChargeMap[converter.PtrToVal(one.ResourceId)] = one
		}
	}

	disks := make([]disk.HuaWeiDisk, 0, len(details))
	for _, one := range details {
		tmp := disk.HuaWeiDisk{
			VolumeDetail: one,
		}
		_, exist := one.Metadata["orderID"]
		if !exist {
			// 如果没有 orderID，则表明是按需付费
			tmp.DiskChargeType = disk.HuaWeiDiskChargeTypeEnum.POST_PAID
		} else {
			// 如果有 orderID，则表明是包年包月计费模式
			charge, exist := idChargeMap[one.Id]
			if exist {
				tmp.DiskChargeType = disk.HuaWeiDiskChargeTypeEnum.PRE_PAID
				tmp.ExpireTime = converter.PtrToVal(charge.ExpireTime)
			}
		}

		disks = append(disks, tmp)
	}

	return disks, nil
}

// DeleteDisk 删除云盘
// reference: https://support.huaweicloud.com/api-evs/evs_04_2008.html
func (h *HuaWei) DeleteDisk(kt *kit.Kit, chargeType string, opt *disk.HuaWeiDiskDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "huawei disk delete option is required")
	}

	switch chargeType {
	case disk.HuaWeiDiskChargeTypeEnum.POST_PAID:
		client, err := h.clientSet.evsClient(opt.Region)
		if err != nil {
			return err
		}

		req, err := opt.ToDeleteVolumeRequest()
		if err != nil {
			return err
		}

		_, err = client.DeleteVolume(req)
		if err != nil {
			logs.Errorf("huawei delete disk failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

	case disk.HuaWeiDiskChargeTypeEnum.PRE_PAID:
		if err := h.DeletePrePaidResource(kt, []string{opt.CloudID}); err != nil {
			logs.Errorf("delete pre paid resource failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

// DeletePrePaidResource ...
func (h *HuaWei) DeletePrePaidResource(kt *kit.Kit, cloudIDs []string) error {

	client, err := h.clientSet.bssintlGlobalClient()
	if err != nil {
		return err
	}

	request := &bssmodel.CancelResourcesSubscriptionRequest{
		Body: &bssmodel.UnsubscribeResourcesReq{
			UnsubscribeType: int32(1),
			ResourceIds:     cloudIDs,
		}}
	_, err = client.CancelResourcesSubscription(request)
	if err != nil {
		logs.Errorf("huawei cancel resource subscription failed, err: %v, req: %+v, rid: %s", err, request, kt.Rid)
		return err
	}

	return nil
}

// AttachDisk 挂载云盘
// reference: https://support.huaweicloud.com/api-ecs/ecs_02_0605.html
func (h *HuaWei) AttachDisk(kt *kit.Kit, opt *disk.HuaWeiDiskAttachOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "huawei disk attach option is required")
	}

	req, err := opt.ToAttachServerVolumeRequest()
	if err != nil {
		return err
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return err
	}

	_, err = client.AttachServerVolume(req)
	if err != nil {
		logs.Errorf("huawei attach disk failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	respPoller := poller.Poller[*HuaWei, []disk.HuaWeiDisk, poller.BaseDoneResult]{
		Handler: &attachDiskPollingHandler{region: opt.Region},
	}
	_, err = respPoller.PollUntilDone(h, kt, []*string{&opt.CloudDiskID}, nil)
	return err
}

// DetachDisk 卸载云盘
// reference: https://support.huaweicloud.com/api-ecs/ecs_02_0606.html
func (h *HuaWei) DetachDisk(kt *kit.Kit, opt *disk.HuaWeiDiskDetachOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "huawei disk detach option is required")
	}

	req, err := opt.ToDetachServerVolumeRequest()
	if err != nil {
		return err
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return err
	}

	_, err = client.DetachServerVolume(req)
	if err != nil {
		logs.Errorf("huawei detach disk failed, err: %v, rid: %s, job id: %s", err, kt.Rid)
		return err
	}

	respPoller := poller.Poller[*HuaWei, []disk.HuaWeiDisk, poller.BaseDoneResult]{
		Handler: &detachDiskPollingHandler{region: opt.Region},
	}
	_, err = respPoller.PollUntilDone(h, kt, []*string{&opt.CloudDiskID}, nil)
	return err
}

type createDiskPollingHandler struct {
	region string
}

func (h *createDiskPollingHandler) Done(pollResult []disk.HuaWeiDisk) (bool, *poller.BaseDoneResult) {
	successCloudIDs := make([]string, 0)
	unknownCloudIDs := make([]string, 0)

	for _, r := range pollResult {
		if r.Status == "creating" {
			unknownCloudIDs = append(unknownCloudIDs, r.Id)
		} else {
			successCloudIDs = append(successCloudIDs, r.Id)
		}
	}

	isDone := false
	if len(successCloudIDs) != 0 && len(successCloudIDs) == len(pollResult) {
		isDone = true
	}

	return isDone, &poller.BaseDoneResult{
		SuccessCloudIDs: successCloudIDs,
		UnknownCloudIDs: unknownCloudIDs,
	}
}

func (h *createDiskPollingHandler) Poll(
	client *HuaWei,
	kt *kit.Kit,
	cloudIDs []*string,
) ([]disk.HuaWeiDisk, error) {
	cIDs := converter.PtrToSlice(cloudIDs)
	result, err := client.ListDisk(kt, &disk.HuaWeiDiskListOption{Region: h.region, CloudIDs: cIDs})
	if err != nil {
		return nil, err
	}

	if len(result) != len(cloudIDs) {
		return nil, fmt.Errorf("query cloudID count: %d, but return %d", len(cloudIDs), len(result))
	}

	return result, nil
}

var _ poller.PollingHandler[*HuaWei, []disk.HuaWeiDisk, poller.BaseDoneResult] = new(createDiskPollingHandler)

type attachDiskPollingHandler struct {
	region string
}

func (h *attachDiskPollingHandler) Done(pollResult []disk.HuaWeiDisk) (bool, *poller.BaseDoneResult) {
	r := pollResult[0]
	if r.Status != "in-use" {
		return false, nil
	}
	return true, nil
}

func (h *attachDiskPollingHandler) Poll(
	client *HuaWei,
	kt *kit.Kit,
	cloudIDs []*string,
) ([]disk.HuaWeiDisk, error) {
	if len(cloudIDs) != 1 {
		return nil, fmt.Errorf("poll only support one id param, but get %v. rid: %s", cloudIDs, kt.Rid)
	}

	cIDs := converter.PtrToSlice(cloudIDs)
	result, err := client.ListDisk(kt, &disk.HuaWeiDiskListOption{Region: h.region, CloudIDs: cIDs})
	if err != nil {
		return nil, err
	}

	return result, nil
}

type detachDiskPollingHandler struct {
	region string
}

func (h *detachDiskPollingHandler) Done(pollResult []disk.HuaWeiDisk) (bool, *poller.BaseDoneResult) {
	r := pollResult[0]
	if r.Status != "available" {
		return false, nil
	}
	return true, nil
}

func (h *detachDiskPollingHandler) Poll(
	client *HuaWei,
	kt *kit.Kit,
	cloudIDs []*string,
) ([]disk.HuaWeiDisk, error) {
	if len(cloudIDs) != 1 {
		return nil, fmt.Errorf("poll only support one id param, but get %v. rid: %s", cloudIDs, kt.Rid)
	}

	cIDs := converter.PtrToSlice(cloudIDs)
	result, err := client.ListDisk(kt, &disk.HuaWeiDiskListOption{Region: h.region, CloudIDs: cIDs})
	if err != nil {
		return nil, err
	}

	return result, nil
}

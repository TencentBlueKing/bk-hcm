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

package tcloud

import (
	"fmt"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/adaptor/types/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

// CreateDisk 创建云硬盘
// reference: https://cloud.tencent.com/document/api/362/16312
func (t *TCloudImpl) CreateDisk(kt *kit.Kit, opt *disk.TCloudDiskCreateOption) (*poller.BaseDoneResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "tcloud disk create option is required")
	}

	resp, err := t.createDisk(kt, opt)
	if err != nil {
		return nil, err
	}

	respPoller := poller.Poller[*TCloudImpl, []disk.TCloudDisk, poller.BaseDoneResult]{
		Handler: &createDiskPollingHandler{region: opt.Region},
	}
	return respPoller.PollUntilDone(t, kt, resp.Response.DiskIdSet, nil)
}

// InquiryPriceDisk 创建云硬盘询价
// reference: https://cloud.tencent.com/document/api/362/16314
func (t *TCloudImpl) InquiryPriceDisk(kt *kit.Kit, opt *disk.TCloudDiskCreateOption) (
	*typecvm.InquiryPriceResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "option is required")
	}

	client, err := t.clientSet.CbsClient(opt.Region)
	if err != nil {
		return nil, err
	}

	req := cbs.NewInquiryPriceCreateDisksRequest()
	req.DiskType = common.StringPtr(opt.DiskType)
	req.DiskCount = opt.DiskCount
	req.DiskSize = opt.DiskSize
	req.DiskChargeType = common.StringPtr(opt.DiskChargeType)
	// 预付费模式需要设定 ChargePrepaid
	if *req.DiskChargeType == disk.TCloudDiskChargeTypeEnum.PREPAID {
		req.DiskChargePrepaid = &cbs.DiskChargePrepaid{
			Period:              opt.DiskChargePrepaid.Period,
			RenewFlag:           opt.DiskChargePrepaid.RenewFlag,
			CurInstanceDeadline: opt.DiskChargePrepaid.CurInstanceDeadline,
		}
	}

	resp, err := client.InquiryPriceCreateDisksWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("inquiry price create disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := new(typecvm.InquiryPriceResult)
	switch opt.DiskChargeType {
	case disk.TCloudDiskChargeTypeEnum.PREPAID:
		result.OriginalPrice = converter.PtrToVal(resp.Response.DiskPrice.OriginalPrice)
		result.DiscountPrice = converter.PtrToVal(resp.Response.DiskPrice.DiscountPrice)
	case disk.TCloudDiskChargeTypeEnum.POSTPAID_BY_HOUR:
		result.OriginalPrice = converter.PtrToVal(resp.Response.DiskPrice.UnitPrice)
		result.DiscountPrice = converter.PtrToVal(resp.Response.DiskPrice.UnitPriceDiscount)

	default:
		return nil, fmt.Errorf("charge type: %s not support", opt.DiskChargeType)
	}

	return result, nil
}

func (t *TCloudImpl) createDisk(kt *kit.Kit, opt *disk.TCloudDiskCreateOption) (*cbs.CreateDisksResponse, error) {
	client, err := t.clientSet.CbsClient(opt.Region)
	if err != nil {
		return nil, err
	}

	req, err := opt.ToCreateDisksRequest()
	if err != nil {
		return nil, err
	}

	return client.CreateDisksWithContext(kt.Ctx, req)
}

// ListDisk 查询云硬盘列表
// reference: https://cloud.tencent.com/document/api/362/16315
func (t *TCloudImpl) ListDisk(kt *kit.Kit, opt *core.TCloudListOption) ([]disk.TCloudDisk, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "tcloud disk list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.CbsClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud cbs client failed, err: %v", err)
	}

	req := cbs.NewDescribeDisksRequest()
	if len(opt.CloudIDs) != 0 {
		req.DiskIds = common.StringPtrs(opt.CloudIDs)
		req.Limit = common.Uint64Ptr(uint64(core.TCloudQueryLimit))
	}
	if opt.Page != nil {
		req.Offset = &opt.Page.Offset
		req.Limit = &opt.Page.Limit
	}

	resp, err := client.DescribeDisksWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	disks := make([]disk.TCloudDisk, 0, len(resp.Response.DiskSet))
	for _, one := range resp.Response.DiskSet {
		var systemDisk bool
		if one.DiskUsage != nil && converter.PtrToVal(one.DiskUsage) == "SYSTEM_DISK" {
			systemDisk = true
		}
		disks = append(disks, disk.TCloudDisk{Boot: systemDisk, Disk: one})
	}
	return disks, nil
}

// CountDisk 基于 DescribeDisksWithContext
// reference: https://cloud.tencent.com/document/api/362/16315
func (t *TCloudImpl) CountDisk(kt *kit.Kit, region string) (int32, error) {

	client, err := t.clientSet.CbsClient(region)
	if err != nil {
		return 0, fmt.Errorf("new tcloud cbs client failed, err: %v", err)
	}

	req := cbs.NewDescribeDisksRequest()
	req.Limit = converter.ValToPtr(uint64(1))
	resp, err := client.DescribeDisksWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("count tcloud disk failed, err: %v, region: %s, rid: %s", err, region, kt.Rid)
		return 0, err
	}
	return int32(*resp.Response.TotalCount), nil
}

// DeleteDisk 删除云盘
// reference: https://cloud.tencent.com/document/product/362/16321
func (t *TCloudImpl) DeleteDisk(kt *kit.Kit, opt *disk.TCloudDiskDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "tcloud disk delete option is required")
	}

	req, err := opt.ToTerminateDisksRequest()
	if err != nil {
		return err
	}

	client, err := t.clientSet.CbsClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new tcloud cbs client failed, err: %v", err)
	}

	_, err = client.TerminateDisksWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("tcloud delete disk failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// AttachDisk 挂载云盘
// reference: https://cloud.tencent.com/document/product/362/16313
func (t *TCloudImpl) AttachDisk(kt *kit.Kit, opt *disk.TCloudDiskAttachOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "tcloud disk attach option is required")
	}

	req, err := opt.ToAttachDisksRequest()
	if err != nil {
		return err
	}

	client, err := t.clientSet.CbsClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new tcloud cbs client failed, err: %v", err)
	}

	_, err = client.AttachDisksWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("tcloud attach disk failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	respPoller := poller.Poller[*TCloudImpl, []disk.TCloudDisk, poller.BaseDoneResult]{
		Handler: &attachDiskPollingHandler{region: opt.Region},
	}
	_, err = respPoller.PollUntilDone(t, kt, converter.SliceToPtr(opt.CloudDiskIDs), nil)
	return err
}

// DetachDisk 卸载云盘
// reference: https://cloud.tencent.com/document/product/362/16316
func (t *TCloudImpl) DetachDisk(kt *kit.Kit, opt *disk.TCloudDiskDetachOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "tcloud disk detach option is required")
	}

	req, err := opt.ToDetachDisksRequest()
	if err != nil {
		return err
	}

	client, err := t.clientSet.CbsClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new tcloud cbs client failed, err: %v", err)
	}

	_, err = client.DetachDisksWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("tcloud detach disk failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	respPoller := poller.Poller[*TCloudImpl, []disk.TCloudDisk, poller.BaseDoneResult]{
		Handler: &detachDiskPollingHandler{region: opt.Region},
	}
	_, err = respPoller.PollUntilDone(t, kt, converter.SliceToPtr(opt.CloudDiskIDs), nil)
	return err
}

type createDiskPollingHandler struct {
	region string
}

// Done 判断是否满足结束条件
func (h *createDiskPollingHandler) Done(pollResult []disk.TCloudDisk) (bool, *poller.BaseDoneResult) {
	successCloudIDs := make([]string, 0)
	unknownCloudIDs := make([]string, 0)

	for _, r := range pollResult {
		if r.DiskState == nil {
			unknownCloudIDs = append(unknownCloudIDs, *r.DiskId)
		} else {
			successCloudIDs = append(successCloudIDs, *r.DiskId)
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

// Poll 轮询结果
func (h *createDiskPollingHandler) Poll(client *TCloudImpl, kt *kit.Kit, cloudIDs []*string) ([]disk.TCloudDisk,
	error) {
	cIDs := converter.PtrToSlice(cloudIDs)

	req := &core.TCloudListOption{
		Region:   h.region,
		CloudIDs: cIDs,
		Page: &core.TCloudPage{
			Offset: 0,
			Limit:  core.TCloudQueryLimit,
		},
	}
	disks, err := client.ListDisk(kt, req)
	if err != nil {
		return nil, err
	}

	return disks, nil
}

var _ poller.PollingHandler[*TCloudImpl, []disk.TCloudDisk, poller.BaseDoneResult] = new(createDiskPollingHandler)

type attachDiskPollingHandler struct {
	region string
}

// Done 判断结束
func (h *attachDiskPollingHandler) Done(pollResult []disk.TCloudDisk) (bool, *poller.BaseDoneResult) {
	r := pollResult[0]
	if converter.PtrToVal(r.DiskState) != "ATTACHED" {
		return false, nil
	}
	return true, nil
}

// Poll 轮询
func (h *attachDiskPollingHandler) Poll(client *TCloudImpl, kt *kit.Kit, cloudIDs []*string) ([]disk.TCloudDisk,
	error) {
	if len(cloudIDs) != 1 {
		return nil, fmt.Errorf("poll only support one id param, but get %v. rid: %s", cloudIDs, kt.Rid)
	}

	cIDs := converter.PtrToSlice(cloudIDs)
	return client.ListDisk(kt,
		&core.TCloudListOption{Region: h.region, CloudIDs: cIDs, Page: &core.TCloudPage{Limit: core.TCloudQueryLimit}})
}

type detachDiskPollingHandler struct {
	region string
}

// Done ...
func (h *detachDiskPollingHandler) Done(pollResult []disk.TCloudDisk) (bool, *poller.BaseDoneResult) {
	r := pollResult[0]
	if converter.PtrToVal(r.DiskState) != "UNATTACHED" {
		return false, nil
	}
	return true, nil
}

// Poll ...
func (h *detachDiskPollingHandler) Poll(client *TCloudImpl, kt *kit.Kit, cloudIDs []*string) ([]disk.TCloudDisk,
	error) {
	if len(cloudIDs) != 1 {
		return nil, fmt.Errorf("poll only support one id param, but get %v. rid: %s", cloudIDs, kt.Rid)
	}

	cIDs := converter.PtrToSlice(cloudIDs)
	return client.ListDisk(kt,
		&core.TCloudListOption{Region: h.region, CloudIDs: cIDs, Page: &core.TCloudPage{Limit: core.TCloudQueryLimit}})
}

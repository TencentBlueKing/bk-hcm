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
	"strings"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2/model"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

// CreateDisk 创建云硬盘
// reference: https://support.huaweicloud.com/api-evs/evs_04_2003.html
func (h *HuaWei) CreateDisk(kt *kit.Kit, opt *disk.HuaWeiDiskCreateOption) ([]string, error) {
	resp, err := h.createDisk(opt)
	if err != nil {
		return nil, err
	}

	respPoller := poller.Poller[*HuaWei, []model.VolumeDetail, poller.BaseDoneResult]{Handler: new(createDiskPollingHandler)}
	_, err = respPoller.PollUntilDone(h, kt, common.StringPtrs(*resp.VolumeIds), nil)
	if err != nil {
		return nil, err
	}

	return *resp.VolumeIds, nil
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
func (h *HuaWei) ListDisk(kt *kit.Kit, opt *disk.HuaWeiDiskListOption) ([]model.VolumeDetail, error) {
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
			return make([]model.VolumeDetail, 0), nil
		}
		logs.Errorf("list huawei disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return *resp.Volumes, nil
}

// DeleteDisk 删除云盘
// reference: https://support.huaweicloud.com/api-evs/evs_04_2008.html
func (h *HuaWei) DeleteDisk(kt *kit.Kit, opt *disk.HuaWeiDiskDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "huawei disk delete option is required")
	}

	client, err := h.clientSet.evsClient(opt.Region)
	if err != nil {
		return err
	}

	req, err := opt.ToDeleteVolumeRequest()
	if err != nil {
		return err
	}

	resp, err := client.DeleteVolume(req)
	if err != nil {
		logs.Errorf(
			"huawei delete disk failed, err: %v, rid: %s, job id: %s",
			err,
			kt.Rid,
			resp.JobId,
		)
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

	resp, err := client.AttachServerVolume(req)
	if err != nil {
		logs.Errorf(
			"huawei attach disk failed, err: %v, rid: %s, job id: %s",
			err,
			kt.Rid,
			resp.JobId,
		)
		return err
	}

	return nil
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

	resp, err := client.DetachServerVolume(req)
	if err != nil {
		logs.Errorf(
			"huawei detach disk failed, err: %v, rid: %s, job id: %s",
			err,
			kt.Rid,
			resp.JobId,
		)
		return err
	}

	return nil
}

type createDiskPollingHandler struct{}

func (h *createDiskPollingHandler) Done(pollResult []model.VolumeDetail) (bool, *poller.BaseDoneResult) {
	return true, nil
}

func (h *createDiskPollingHandler) Poll(
	client *HuaWei,
	kt *kit.Kit,
	cloudIDs []*string,
) ([]model.VolumeDetail, error) {
	return nil, nil
}

var _ poller.PollingHandler[*HuaWei, []model.VolumeDetail, poller.BaseDoneResult] = new(createDiskPollingHandler)

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

package gcp

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"google.golang.org/api/compute/v1"
)

// CreateDisk 创建云硬盘
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/disks/insert
func (g *Gcp) CreateDisk(kt *kit.Kit, opt *disk.GcpDiskCreateOption) (*poller.BaseDoneResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "gcp disk create option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	diskCloudIDs := make([]string, 0)

	if *opt.DiskCount == 1 {
		_, err := g.createDisk(kt, opt)
		if err != nil {
			return nil, err
		}
		cloudID, err := g.getDiskCloudID(kt, opt.Zone, opt.DiskName)
		if err != nil {
			return nil, err
		}

		diskCloudIDs = append(diskCloudIDs, *cloudID)
	} else {
		diskName := opt.DiskName
		for i := uint64(1); i <= *opt.DiskCount; i++ {
			opt.DiskName = fmt.Sprintf("%s-%d", diskName, i)
			_, err := g.createDisk(kt, opt)
			if err != nil {
				return nil, err
			}

			cloudID, err := g.getDiskCloudID(kt, opt.Zone, opt.DiskName)
			if err != nil {
				return nil, err
			}

			diskCloudIDs = append(diskCloudIDs, *cloudID)
		}
	}

	respPoller := poller.Poller[*Gcp, []*compute.Disk, poller.BaseDoneResult]{
		Handler: &createDiskPollingHandler{Zone: opt.Zone},
	}
	return respPoller.PollUntilDone(g, kt, converter.SliceToPtr(diskCloudIDs), nil)
}

func (g *Gcp) createDisk(kt *kit.Kit, opt *disk.GcpDiskCreateOption) (*compute.Operation, error) {
	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	cloudProjectID := g.CloudProjectID()
	req, err := opt.ToCreateDiskRequest(cloudProjectID)
	if err != nil {
		return nil, err
	}

	var call *compute.DisksInsertCall
	call = client.Disks.Insert(cloudProjectID, opt.Zone, req).Context(kt.Ctx)
	return call.Do()
}

// ListDisk 查看云硬盘
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/disks/list
func (g *Gcp) ListDisk(kt *kit.Kit, opt *disk.GcpDiskListOption) ([]*compute.Disk, string, error) {
	if opt == nil {
		return nil, "", errf.New(errf.InvalidParameter, "gcp disk list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, "", errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, "", err
	}

	request := client.Disks.List(g.clientSet.credential.CloudProjectID, opt.Zone).Context(kt.Ctx)

	if len(opt.CloudIDs) > 0 {
		request.Filter(generateResourceIDsFilter(opt.CloudIDs))
	}

	if len(opt.SelfLinks) > 0 {
		request.Filter(generateResourceFilter("selfLink", opt.SelfLinks))
	}

	if opt.Page != nil {
		request.MaxResults(opt.Page.PageSize).PageToken(opt.Page.PageToken)
	}

	resp, err := request.Do()
	if err != nil {
		logs.Errorf("list disks failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, "", err
	}

	for index := range resp.Items {
		resp.Items[index].Region = resp.Items[index].
			Zone[strings.LastIndex(resp.Items[index].Zone, "/")+1 : strings.LastIndex(resp.Items[index].Zone, "-")]
	}

	return resp.Items, resp.NextPageToken, nil
}

// DeleteDisk 删除云盘
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/disks/delete
func (g *Gcp) DeleteDisk(kt *kit.Kit, opt *disk.GcpDiskDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "gcp disk delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	_, err = client.Disks.Delete(g.CloudProjectID(), opt.Zone, opt.DiskName).Context(kt.Ctx).Do()
	return err
}

// AttachDisk 挂载云盘
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/attachDisk
func (g *Gcp) AttachDisk(kt *kit.Kit, opt *disk.GcpDiskAttachOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "gcp disk attach option is required")
	}

	req, err := opt.ToAttachDiskRequest()
	if err != nil {
		return err
	}

	req.Source = fmt.Sprintf("projects/%s/zones/%s/disks/%s", g.CloudProjectID(), opt.Zone, opt.DiskName)

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	_, err = client.Instances.AttachDisk(g.CloudProjectID(), opt.Zone, opt.CvmName, req).
		Context(kt.Ctx).
		Do()

	if err != nil {
		logs.Errorf("attach disk failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
	}

	return err
}

// DetachDisk 卸载云盘
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/detachDisk
func (g *Gcp) DetachDisk(kt *kit.Kit, opt *disk.GcpDiskDetachOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "gcp disk detach option is required")
	}

	if err := opt.Validate(); err != nil {
		return err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	_, err = client.Instances.DetachDisk(g.CloudProjectID(), opt.Zone, opt.CvmName, opt.DeviceName).
		Context(kt.Ctx).
		Do()

	if err != nil {
		logs.Errorf("detach disk failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
	}

	return err
}

func (g *Gcp) getDiskCloudID(kt *kit.Kit, zone string, diskName string) (*string, error) {
	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	endTime := time.Now().Add(time.Duration(1) * time.Minute)
	for {
		if time.Now().After(endTime) {
			return nil, fmt.Errorf("disk not found, zone: %s, diskName: %s", zone, diskName)
		}

		resp, err := client.Disks.Get(g.CloudProjectID(), zone, diskName).Context(kt.Ctx).Do()
		if err != nil {
			return nil, err
		}

		if resp != nil && resp.Name == diskName {
			cloudID := strconv.FormatUint(resp.Id, 10)
			return &cloudID, nil
		}
	}
}

type createDiskPollingHandler struct {
	Zone string
}

func (h *createDiskPollingHandler) Done(pollResult []*compute.Disk) (bool, *poller.BaseDoneResult) {
	successCloudIDs := make([]string, 0)
	unknownCloudIDs := make([]string, 0)

	for _, r := range pollResult {
		if r.Status == "CREATING" {
			unknownCloudIDs = append(unknownCloudIDs, strconv.FormatUint(r.Id, 10))
		} else {
			successCloudIDs = append(successCloudIDs, strconv.FormatUint(r.Id, 10))
		}
	}

	isDone := false
	if len(unknownCloudIDs) == 0 {
		isDone = true
	}

	return isDone, &poller.BaseDoneResult{
		SuccessCloudIDs: successCloudIDs,
		UnknownCloudIDs: unknownCloudIDs,
	}
}

func (h *createDiskPollingHandler) Poll(client *Gcp, kt *kit.Kit, cloudIDs []*string) ([]*compute.Disk, error) {
	cIDs := converter.PtrToSlice(cloudIDs)
	result, _, err := client.ListDisk(
		kt,
		&disk.GcpDiskListOption{Zone: h.Zone, CloudIDs: cIDs},
	)
	return result, err
}

var _ poller.PollingHandler[*Gcp, []*compute.Disk, poller.BaseDoneResult] = new(createDiskPollingHandler)

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

// Package tcloud 如下:
// 使用 SDK-Client 调用腾讯云 API
package tcloud

import (
	"fmt"
	"strings"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	typescfs "hcm/pkg/adaptor/types/cfs"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	"github.com/pkg/errors"
	cfs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cfs/v20190719"
)

// 轮询初始化化
var (
	// cfs poll
	_ poller.PollingHandler[*TCloudImpl, []*cfs.FileSystemInfo, poller.BaseDoneResult] = new(createCfsPollingHandler)
)

// createCfsPollingHandler poll handler for create status
type createCfsPollingHandler struct {
	region string

	kt *kit.Kit
}

// CreateStorage 创建文件存储
// reference: https://cloud.tencent.com/document/api/582/38174
func (t *TCloudImpl) CreateStorage(kt *kit.Kit, opt *typescfs.TCloudCreateStorageOption) (*typescfs.StorageInfo, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "tcloud storage create option is required")
	}
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := t.clientSet.CfsClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("init tcloud client failed, err: %v, rid: %s", err, kt.Rid)
	}

	createResp, err := client.CreateCfsFileSystemWithContext(kt.Ctx, newCreateCfsFileSystemRequest(opt))
	if err != nil {
		logs.Errorf("tcloud cfs create failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	//respStr, _ := json.MarshalToString(createResp) // debug log
	//logs.Infof("CreateCfs resp: %s", respStr)      // debug log

	ids := []*string{createResp.Response.FileSystemId} // 创建轮询
	createPoller := poller.Poller[*TCloudImpl, []*cfs.FileSystemInfo, poller.BaseDoneResult]{
		Handler: &createCfsPollingHandler{region: opt.Region, kt: kt}}

	resp, err := createPoller.PollUntilDone(t, kt, ids, types.NewCreateCfsPollerOption()) // 执行轮询
	if err != nil {
		return nil, err
	}
	if len(resp.SuccessCloudIDs) == 0 {
		return nil, errors.Errorf("tcloud create cfs storage failed, cloudID: %s, result: %s, rid: %s",
			cvt.PtrToVal(createResp.Response.FileSystemId), resp, kt.Rid)
	}
	cloudID := resp.SuccessCloudIDs[0]
	//logs.Infof("tcloud create cfs storage, cloudID: %s", cloudID) // debug log

	result, err := t.GetStorages(kt, &typescfs.TCloudGetStorageOption{Region: opt.Region, CloudID: cloudID})
	if err != nil {
		logs.Errorf("tcloud get cfs storage failed, cloudID: %s, err: %v, rid: %s", cloudID, err, kt.Rid)
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("tcloud get cfs storage status failed, cloudID: %s, rid: %s", cloudID, kt.Rid)
	}

	return result, nil
}

// DeleteStorage 删除文件存储
// reference: https://cloud.tencent.com/document/api/582/38173
func (t *TCloudImpl) DeleteStorage(kt *kit.Kit, opt *typescfs.TCloudDeleteStorageOption) (
	*typescfs.TCloudDeleteStorageResult, error) {
	result := new(typescfs.TCloudDeleteStorageResult)
	result.Status = false
	result.Name = opt.Name
	result.CloudID = opt.CloudID
	if opt == nil {
		return result, errf.New(errf.InvalidParameter, "tcloud storage delete option is required")
	}
	if err := opt.Validate(); err != nil {
		return result, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := t.clientSet.CfsClient(opt.Region)
	if err != nil {
		return result, fmt.Errorf("init tcloud client failed, err: %v, rid: %s", err, kt.Rid)
	}

	req := newDeleteCfsFileSystemRequest(opt)
	resp, err := client.DeleteCfsFileSystemWithContext(kt.Ctx, req) // note: 删除失败可能存在挂载的情况
	if err != nil {
		logs.Errorf("tcloud cfs delete failed, cloudID: %s, err: %v, rid: %s", opt.CloudID, err, kt.Rid)
		return result, err
	}
	result.Status = true
	result.RequestId = resp.Response.RequestId
	//respStr, _ := json.MarshalToString(resp)  // debug log
	//logs.Infof("DeleteCfs resp: %s", respStr) // debug log

	return result, nil
}

// ListStorages 查询文件存储列表
// reference: https://cloud.tencent.com/document/api/582/38170
func (t *TCloudImpl) ListStorages(kt *kit.Kit, opt *typescfs.TCloudListStorageOption) (*typescfs.TCloudListStorageResult,
	error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "tcloud storage list option is required")
	}
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := t.clientSet.CfsClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("init tcloud client failed, err: %v, rid: %s", err, kt.Rid)
	}

	req := newListCfsFileSystemRequest(opt)
	resp, err := client.DescribeCfsFileSystemsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("tcloud cfs list failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	//respStr, _ := json.MarshalToString(resp) // debug log
	//logs.Infof("ListCfs resp: %s", respStr)  // debug log

	result := new(typescfs.TCloudListStorageResult)
	if *resp.Response.TotalCount == 0 {
		return result, nil
	}
	result.Storages = make([]*typescfs.StorageInfo, *resp.Response.TotalCount)

	for i, info := range resp.Response.FileSystems {
		req := cfs.NewDescribeMountTargetsRequest()
		req.FileSystemId = info.FileSystemId
		resp, err := client.DescribeMountTargetsWithContext(kt.Ctx, req)
		if err != nil {
			logs.Errorf("tcloud cfs mount info get failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		result.Storages[i] = new(typescfs.StorageInfo)
		result.Storages[i].Cfs = info

		if *resp.Response.NumberOfMountTargets > 0 {
			result.Storages[i].MountInfo = resp.Response.MountTargets[0]
		}
	}

	return result, nil
}

// GetStorages 查询文件存储列表
// reference: https://cloud.tencent.com/document/api/582/38170
func (t *TCloudImpl) GetStorages(kt *kit.Kit, opt *typescfs.TCloudGetStorageOption) (*typescfs.StorageInfo, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "tcloud storage list option is required")
	}
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := t.clientSet.CfsClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("init tcloud client failed, err: %v, rid: %s", err, kt.Rid)
	}

	req := newGetCfsFileSystemRequest(opt)
	resp, err := client.DescribeCfsFileSystemsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("tcloud cfs get failed, cloudID: %s, err: %v, rid: %s", opt.CloudID, err, kt.Rid)
		return nil, err
	}
	//respStr, _ := json.MarshalToString(resp) // debug log
	//logs.Infof("GetCfs resp: %s", respStr)   // debug log

	if *resp.Response.TotalCount == 0 {
		return nil, fmt.Errorf("tcloud cfs get failed, resp is nil, cloudID: %s, rid: %s", opt.CloudID, kt.Rid)
	}
	result := new(typescfs.StorageInfo)
	result.Cfs = resp.Response.FileSystems[0]

	mountReq := newDescribeMountTargetsRequest(result.Cfs.FileSystemId)
	mountResp, err := client.DescribeMountTargetsWithContext(kt.Ctx, mountReq)
	if err != nil {
		logs.Errorf("tcloud cfs mount info get failed, cloudID: %s, err: %v, rid: %s", opt.CloudID, err, kt.Rid)
		return nil, err
	}
	if *mountResp.Response.NumberOfMountTargets > 0 {
		result.MountInfo = mountResp.Response.MountTargets[0]
	}

	return result, nil
}

// Poll 轮询结果
func (h *createCfsPollingHandler) Poll(client *TCloudImpl, kt *kit.Kit, cloudIDs []*string) ([]*cfs.FileSystemInfo,
	error) {
	result := make([]*cfs.FileSystemInfo, 0, len(cloudIDs))

	for _, id := range cloudIDs {
		req := cfs.NewDescribeCfsFileSystemsRequest()
		req.FileSystemId = id
		// note: add vpcId and subnetId filter

		cfsCli, err := client.clientSet.CfsClient(h.region)
		if err != nil {
			return nil, fmt.Errorf("init tcloud client failed, err: %v, rid: %s", err, kt.Rid)
		}

		resp, err := cfsCli.DescribeCfsFileSystemsWithContext(kt.Ctx, req)
		if err != nil {
			return nil, errors.Wrapf(err, "tcloud cfs get failed, cloudID: %s, err: %v, rid: %s", *id,
				err, kt.Rid)
		}

		result = append(result, resp.Response.FileSystems...)
	}

	if len(result) != len(cloudIDs) {
		return nil, fmt.Errorf("query cfs count: %d not equal return count: %d, cloudIDs: %s, rid: %s", len(cloudIDs),
			len(result), strings.Join(cvt.PtrToSlice(cloudIDs), ","), kt.Rid)
	}

	return result, nil
}

// Done poll 获取到数据后, 用该方法判断是否符合预期
//
// 文件系统状态, 取值范围:
// creating:创建中; mounting:挂载中; create_failed:创建失败; available:可使用; unserviced:停服中; upgrading:升级中;
//
// 示例值：creating.
func (h *createCfsPollingHandler) Done(data []*cfs.FileSystemInfo) (bool, *poller.BaseDoneResult) {
	result := &poller.BaseDoneResult{
		SuccessCloudIDs: make([]string, 0),
		FailedCloudIDs:  make([]string, 0),
		UnknownCloudIDs: make([]string, 0),
	}
	flag := true

	for _, instance := range data {
		//logs.Infof("tcloud cfs create status, cloudID: %s, status: %s, rid: %s", instance.FileSystemId,
		//	instance.LifeCycleState, h.kt.Rid) // debug log

		// 创建中
		if strings.EqualFold(cvt.PtrToVal(instance.LifeCycleState), "creating") {
			flag = false
			result.UnknownCloudIDs = append(result.UnknownCloudIDs, *instance.FileSystemId)
			continue
		}

		// 生产失败
		if strings.EqualFold(cvt.PtrToVal(instance.LifeCycleState), "create_failed") {
			result.FailedCloudIDs = append(result.FailedCloudIDs, *instance.FileSystemId)
			//result.FailedMessage = cvt.PtrToVal(instance.LatestOperationErrorMsg)
			continue
		}

		result.SuccessCloudIDs = append(result.SuccessCloudIDs, *instance.FileSystemId)
	}

	return flag, result
}

// newCreateCfsFileSystemRequest return *cfs.CreateCfsFileSystemRequest
func newCreateCfsFileSystemRequest(opt *typescfs.TCloudCreateStorageOption) *cfs.CreateCfsFileSystemRequest {
	req := cfs.NewCreateCfsFileSystemRequest()
	req.FsName = &opt.Name
	req.Zone = &opt.Zone
	req.NetInterface = &opt.NetInterface
	req.PGroupId = &opt.PGroupId
	// 选填
	req.VpcId = opt.CloudVpcID
	req.SubnetId = opt.CloudSubnetID
	req.Protocol = opt.Protocol
	req.StorageType = opt.StorageType
	req.Capacity = opt.Capacity
	req.EnableAutoScaleUp = opt.EnableAutoScaleUp
	// 当前版本不支持, 故不启用
	//req.CfsVersion = opt.CfsVersion
	//req.MetaType = opt.MetaType

	for _, tag := range opt.Tags {
		cfsTag := &cfs.TagInfo{
			TagKey:   cvt.ValToPtr(tag.Key),
			TagValue: cvt.ValToPtr(tag.Value),
		}
		req.ResourceTags = append(req.ResourceTags, cfsTag)
	}

	return req
}

// newDeleteCfsFileSystemRequest return *cfs.CreateCfsFileSystemRequest
func newDeleteCfsFileSystemRequest(opt *typescfs.TCloudDeleteStorageOption) *cfs.DeleteCfsFileSystemRequest {
	req := cfs.NewDeleteCfsFileSystemRequest()
	req.FileSystemId = &opt.CloudID
	return req
}

// newDeleteCfsFileSystemRequest return *cfs.CreateCfsFileSystemRequest
func newListCfsFileSystemRequest(opt *typescfs.TCloudListStorageOption) *cfs.DescribeCfsFileSystemsRequest {
	req := cfs.NewDescribeCfsFileSystemsRequest()
	req.FileSystemId = opt.CloudID
	req.VpcId = opt.VpcId
	req.SubnetId = opt.SubnetId
	req.Offset = opt.Offset
	req.Limit = opt.Limit
	req.CreationToken = opt.CreationToken

	return req
}

// newGetCfsFileSystemRequest return *cfs.CreateCfsFileSystemRequest
func newGetCfsFileSystemRequest(opt *typescfs.TCloudGetStorageOption) *cfs.DescribeCfsFileSystemsRequest {
	req := cfs.NewDescribeCfsFileSystemsRequest()
	req.VpcId = opt.VpcId
	req.SubnetId = opt.SubnetId
	req.FileSystemId = cvt.ValToPtr(opt.CloudID)
	req.Limit = cvt.ValToPtr(uint64(core.TCloudQueryLimit))

	return req
}

// newDescribeMountTargetsRequest return *cfs.DescribeMountTargetsRequest
func newDescribeMountTargetsRequest(cfsId *string) *cfs.DescribeMountTargetsRequest {
	mountReq := cfs.NewDescribeMountTargetsRequest()
	mountReq.FileSystemId = cfsId
	return mountReq
}

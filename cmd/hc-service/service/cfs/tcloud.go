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

// Package cfs 如下:
// 处理tcloud cfs请求
package cfs

import (
	"net/http"
	"regexp"

	"hcm/cmd/hc-service/service/capability"
	typecfs "hcm/pkg/adaptor/types/cfs"
	"hcm/pkg/api/core"
	corecfs "hcm/pkg/api/core/cloud/cfs"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service/cfs"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"

	"github.com/pkg/errors"
)

func (svc *cfsSvc) initTCloudCfsService(cap *capability.Capability) {
	h := rest.NewHandler()
	h.Add("CreateTCloudCfs", http.MethodPost, "/vendors/tcloud/cfs/storage/create", svc.CreateTCloudCfs)
	h.Add("DeleteTCloudCfs", http.MethodDelete, "/vendors/tcloud/cfs/storage/delete", svc.DeleteTCloudCfs)
	h.Add("ListTCloudCfs", http.MethodPost, "/vendors/tcloud/cfs/storage/list", svc.ListTCloudCfs)
	h.Add("GetTCloudCfs", http.MethodPost, "/vendors/tcloud/cfs/storage/get", svc.GetTCloudCfs)
	h.Load(cap.WebService)
}

// CreateTCloudCfs create tcloud cfs storage.
func (svc *cfsSvc) CreateTCloudCfs(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.TCloudCreateCfsReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("decode request failed for CreateTCloudCfs request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		logs.Errorf("validation failed for CreateTCloudCfs request, err: %v, req: %+v, rid: %s", err,
			cvt.PtrToVal(req), cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	tSdk, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("request dataservice create tcloud cfs failed, rid: %s", cts.Kit.Rid)
		return nil, errors.Wrapf(err, "request dataservice create tcloud cfs failed, rid: %s", cts.Kit.Rid)
	}

	// 创建云上资源
	tcloudResp, err := tSdk.CreateStorage(cts.Kit, newCreateOption(req))
	if err != nil {
		logs.Errorf("tcloud create cfs storage failed, err: %v, req: %+v, rid: %s", err, cvt.PtrToVal(req),
			cts.Kit.Rid)
		return nil, errors.Wrapf(err, "tcloud create cfs storage failed, rid: %s", cts.Kit.Rid)
	}
	// 创建记录
	createResp, err := svc.dc.TCloud.Cfs.CreateCfs(cts.Kit.Ctx, cts.Kit.Header(), newDBCreateReq(req, tcloudResp))
	if err != nil {
		logs.Errorf("request dataservice create tcloud cfs failed, err: %v, cloudId: %s, rid: %s", err,
			*tcloudResp.Cfs.FileSystemId, cts.Kit.Rid)
		return nil, errors.Wrapf(err, "request dataservice create tcloud cfs failed, cloudId: %s, rid: %s",
			*tcloudResp.Cfs.FileSystemId, cts.Kit.Rid)
	}
	// 获取cfs
	resp, err := svc.dc.TCloud.Cfs.GetCfs(cts.Kit.Ctx, cts.Kit.Header(), createResp.IDs[0])
	if err != nil {
		logs.Errorf("request dataservice get tcloud cfs failed, err: %s, rid: %s", err.Error(), cts.Kit.Rid)
		return nil, errors.Wrapf(err, "request dataservice get tcloud cfs failed, rid: %s", cts.Kit.Rid)
	}

	return resp, nil
}

// DeleteTCloudCfs delete tcloud cfs storage.
func (svc *cfsSvc) DeleteTCloudCfs(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.TCloudDeleteCfsReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("decode request failed for DeleteTCloudCfs request, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		logs.Errorf("validation failed for DeleteTCloudCfs request, err: %v, req: %+v, rid: %s", err)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	// 获取cfs
	resp, err := svc.dc.TCloud.Cfs.GetCfs(cts.Kit.Ctx, cts.Kit.Header(), req.ID)
	if err != nil {
		logs.Errorf("request dataservice get tcloud cfs failed, rid: %s", cts.Kit.Rid)
		return nil, errors.Wrapf(err, "request dataservice get tcloud cfs failed, rid: %s", cts.Kit.Rid)
	}
	if resp.CloudID != req.CloudID { // 无法删除, req的cloud_id 与 当前db记录中的cloud_id不一致
		logs.Errorf("Unable to delete, the cloud_id of the request is inconsistent with the cloud_id in the"+
			" current db record, req cloud_id: %s, db cloud_id: %s, rid: %s", req.CloudID, resp.CloudID, cts.Kit.Rid)
		return nil, errors.Errorf("Unable to delete, the cloud_id of the request is inconsistent with the"+
			" cloud_id in the current db record, req cloud_id: %s, db cloud_id: %s, rid: %s", req.CloudID, resp.CloudID,
			cts.Kit.Rid)
	}
	// 云上资源删除
	tSdk, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("request dataservice delete tcloud cfs failed, rid: %s", cts.Kit.Rid)
		return nil, errors.Wrapf(err, "request dataservice delete tcloud cfs failed, rid: %s", cts.Kit.Rid)
	}
	result, err := tSdk.DeleteStorage(cts.Kit, newDeleteOption(req))
	if err != nil {
		logs.Errorf("tcloud delete storage failed, err: %v, req: %+v, rid: %s", err, cvt.PtrToVal(req),
			cts.Kit.Rid)
		return result, errors.Wrapf(err, "tcloud delete storage failed, rid: %s", cts.Kit.Rid)
	}
	// db删除
	deleteReq := new(protocloud.CfsBatchDeleteReq)
	deleteReq.Filter = tools.ContainersExpression("id", []string{resp.ID})
	if err = svc.dc.TCloud.Cfs.BatchDeleteCfs(cts.Kit.Ctx, cts.Kit.Header(), deleteReq); err != nil {
		logs.Errorf("request dataservice delete tcloud cfs failed, err: %s, ids: %s, rid: %s", err.Error(),
			req.ID, cts.Kit.Rid)
		return nil, errors.Wrapf(err, "request dataservice delete tcloud cfs failed, ids: %s, rid: %s", req.ID,
			cts.Kit.Rid)
	}

	return result, nil
}

// GetTCloudCfs get tcloud cfs storage.
// 从云上接口直接获取
func (svc *cfsSvc) GetTCloudCfs(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.TCloudGetCfsReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("GetTCloudCfs decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		logs.Errorf("GetTCloudCfs validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	// 通过list查找
	result, err := svc.dc.TCloud.Cfs.ListCfsExt(cts.Kit.Ctx, cts.Kit.Header(), newCfsListReq(newTCloudListCfsReq(req)))
	if err != nil {
		logs.Errorf("request dataservice list tcloud cfs failed, err: %s, rid: %s", err.Error(), cts.Kit.Rid)
		return nil, errors.Wrapf(err, "request dataservice list tcloud cfs failed, err: %s, rid: %s",
			err.Error(), cts.Kit.Rid)
	}
	for _, detail := range result.Details {
		if detail.CloudID == req.CloudID || cvt.PtrToVal(req.ID) == detail.ID {
			return detail, nil
		}
	}

	// 数据同步
	if req.ID == nil && req.BkBizID != nil { // 云上资源查询
		tSdk, err := svc.ad.TCloud(cts.Kit, req.AccountID)
		if err != nil {
			logs.Errorf("request dataservice get tcloud cfs failed, rid: %s", cts.Kit.Rid)
			return nil, errors.Wrapf(err, "request dataservice get tcloud cfs failed, rid: %s", cts.Kit.Rid)
		}
		tcloudResp, err := tSdk.GetStorages(cts.Kit, newGetOption(req))
		if err != nil {
			logs.Errorf("tcloud get storage failed, err: %v, req: %+v, rid: %s", err, cvt.PtrToVal(req),
				cts.Kit.Rid)
			return nil, errors.Wrapf(err, "tcloud get storage failed, rid: %s", cts.Kit.Rid)
		}
		createResp, err := svc.dc.TCloud.Cfs.CreateCfs(cts.Kit.Ctx, cts.Kit.Header(), newTCloudSyncDBCreateReq(
			cvt.PtrToVal(req.BkBizID), req.AccountID, tcloudResp))
		if err != nil {
			logs.Errorf("request dataservice create tcloud cfs failed(sync), err: %v, cloudId: %s, rid: %s",
				err, *tcloudResp.Cfs.FileSystemId, cts.Kit.Rid)
			return nil, errors.Wrapf(err, "request dataservice create tcloud cfs failed(sync), cloudId: %s, rid: %s",
				*tcloudResp.Cfs.FileSystemId, cts.Kit.Rid)
		}
		req.ID = cvt.ValToPtr(createResp.IDs[0])
	}

	// 获取cfs
	resp, err := svc.dc.TCloud.Cfs.GetCfs(cts.Kit.Ctx, cts.Kit.Header(), cvt.PtrToVal(req.ID))
	if err != nil {
		logs.Errorf("request dataservice get tcloud cfs failed, err: %s, rid: %s", err.Error(), cts.Kit.Rid)
		return nil, errors.Wrapf(err, "request dataservice get tcloud cfs failed, rid: %s", cts.Kit.Rid)
	}

	return resp, nil
}

// ListTCloudCfs list tcloud cfs storage.
// 从云上接口直接获取
func (svc *cfsSvc) ListTCloudCfs(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.TCloudListCfsReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("ListTCloudCfs decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		logs.Errorf("ListTCloudCfs validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	//// 云上资源查询
	//tSdk, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	//if err != nil {
	//	return nil, err
	//}
	//opt := newListOption(req)
	//result, err := tSdk.ListStorages(cts.Kit, opt)
	//if err != nil {
	//	logs.Errorf("tcloud list storage failed, err: %v, req: %+v, rid: %s", err, converter.PtrToVal(req),
	//		cts.Kit.Rid)
	//	return nil, err
	//}

	// 获取cfs
	resp, err := svc.dc.TCloud.Cfs.ListCfsExt(cts.Kit.Ctx, cts.Kit.Header(), newCfsListReq(req))
	if err != nil {
		logs.Errorf("request dataservice list tcloud cfs failed, err: %s, rid: %s", err.Error(), cts.Kit.Rid)
		return nil, errors.Wrapf(err, "request dataservice list tcloud cfs failed, rid: %s", cts.Kit.Rid)
	}

	return resp, nil
}

// newTCloudListCfsReq 给到get使用 return *hcservice.TCloudListCfsReq
func newTCloudListCfsReq(req *hcservice.TCloudGetCfsReq) *hcservice.TCloudListCfsReq {
	result := &hcservice.TCloudListCfsReq{
		AccountID: req.AccountID,
		Region:    req.Region,
		ID:        req.ID,
		BkBizID:   req.BkBizID,
	}

	return result
}

// newCfsListReq return *protocloud.CfsListReq
func newCfsListReq(req *hcservice.TCloudListCfsReq) *protocloud.CfsListReq {
	page := core.NewDefaultBasePage()
	if req.Offset != nil {
		page.Count = false
		page.Start = uint32(*req.Offset)
	}
	if req.Limit != nil {
		page.Limit = uint(*req.Limit)
	}
	page.Sort = "id"
	page.Order = "DESC"

	rules := make([]filter.RuleFactory, 0)
	rules = append(rules, &filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: req.Region})
	rules = append(rules, &filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: req.AccountID})
	if req.ID != nil {
		rules = append(rules, &filter.AtomRule{Field: "id", Op: filter.Equal.Factory(), Value: cvt.PtrToVal(req.ID)})
	}
	if req.BkBizID != nil {
		rules = append(rules, &filter.AtomRule{Field: "bk_biz_id", Op: filter.Equal.Factory(),
			Value: cvt.PtrToVal(req.BkBizID)})
	}
	if req.CloudID != nil {
		rules = append(rules, &filter.AtomRule{Field: "cloud_id", Op: filter.Equal.Factory(),
			Value: cvt.PtrToVal(req.CloudID)})
	}
	if req.VpcId != nil {
		rules = append(rules, &filter.AtomRule{Field: "vpc_id", Op: filter.Equal.Factory(),
			Value: cvt.PtrToVal(req.VpcId)})
	}
	if req.SubnetId != nil {
		rules = append(rules, &filter.AtomRule{Field: "subnet_id", Op: filter.Equal.Factory(),
			Value: cvt.PtrToVal(req.SubnetId)})
	}
	listReq := &protocloud.CfsListReq{
		Page: page,
		Filter: &filter.Expression{
			Rules: rules,
			Op:    filter.And,
		},
	}

	return listReq
}

// newCreateOption return *typecfs.TCloudStorageCreateOption
func newCreateOption(req *hcservice.TCloudCreateCfsReq) *typecfs.TCloudCreateStorageOption {
	return &typecfs.TCloudCreateStorageOption{
		// 必填
		BkBizID:      req.BkBizID,
		Name:         req.Name,
		Region:       req.Region,
		Zone:         req.Zone,
		NetInterface: req.NetInterface,
		PGroupId:     req.PGroupId,
		// 选填
		CloudVpcID:        req.CloudVpcID,
		CloudSubnetID:     req.CloudSubnetID,
		Protocol:          req.Protocol,
		StorageType:       req.StorageType,
		Capacity:          req.Capacity,
		EnableAutoScaleUp: req.EnableAutoScaleUp,
		CfsVersion:        req.CfsVersion,
		//MetaType: req.MetaType,
		Memo: req.Memo,
		Tags: req.Tags,
	}
}

// newDeleteOption return *typecfs.TCloudStorageDeleteOption
func newDeleteOption(req *hcservice.TCloudDeleteCfsReq) *typecfs.TCloudDeleteStorageOption {
	return &typecfs.TCloudDeleteStorageOption{ // 必填
		//AccountID: req.AccountID,
		CloudID: req.CloudID,
		Name:    req.Name,
		Region:  req.Region,
	}
}

// calcAvailCapacity 计算剩余容量
// limit 是GiB(GB)
// use 是B(byte)
func calcAvailCapacity(limit uint64, use *uint64) uint64 {
	total := limit * 1024 * 1024 * 1024 //  MB * KB * 1024
	return total - *use
}

// zoneToRegion 可用区转换为地域
func zoneToRegion(zone string) string {
	re := regexp.MustCompile(`-\d+$`)
	region := re.ReplaceAllString(zone, "")
	return region
}

// newTCloudSyncDBCreateReq db req
func newTCloudSyncDBCreateReq(bkBizID int64, accountID string, info *typecfs.StorageInfo,
) *protocloud.CfsCreateReq[corecfs.TCloudCfsExtension] {

	extension := &corecfs.TCloudCfsExtension{
		FSID:            cvt.PtrToVal(info.MountInfo.FSID),
		IpAddress:       cvt.PtrToVal(info.MountInfo.IpAddress),
		PGroupId:        cvt.PtrToVal(info.Cfs.PGroup.PGroupId),
		MountInfo:       info.MountInfo,
		AutoScaleUpRule: info.Cfs.AutoScaleUpRule,
		//ExstraPerformanceInfo: info.Cfs.ExstraPerformanceInfo,
		StorageResourcePkg:   info.Cfs.StorageResourcePkg,
		BandwidthResourcePkg: info.Cfs.BandwidthResourcePkg,
		SnapStatus:           info.Cfs.SnapStatus,
		AutoSnapshotPolicyId: info.Cfs.AutoSnapshotPolicyId,
		TieringState:         info.Cfs.TieringState,
		TieringDetail:        info.Cfs.TieringDetail,
		Version:              info.Cfs.Version,
		//MetaType: info.Cfs.MetaType,
	}
	extension.Tags = make(map[string]string)
	for _, tag := range info.Cfs.Tags {
		extension.Tags[cvt.PtrToVal(tag.TagKey)] = cvt.PtrToVal(tag.TagValue)
	}
	capacity := cvt.PtrToVal(info.Cfs.SizeLimit)
	if capacity == 0 {
		capacity = cvt.PtrToVal(info.Cfs.Capacity)
	}

	req := &protocloud.CfsCreateReq[corecfs.TCloudCfsExtension]{
		BkBizID:   bkBizID,
		AccountID: accountID,
		Name:      cvt.PtrToVal(info.Cfs.FsName),
		Region:    zoneToRegion(cvt.PtrToVal(info.Cfs.Zone)),
		Zone:      cvt.PtrToVal(info.Cfs.Zone),
		CloudID:   cvt.PtrToVal(info.Cfs.FileSystemId),
		// 文件系统空间限制. 单位:GiB
		SizeLimit: capacity,
		// 文件系统已使用容量. 单位：Byte
		SizeByte: cvt.PtrToVal(info.Cfs.SizeByte),
		// 文件系统剩余容量. 单位：Byte. 示例值：10
		AvailCapacity: calcAvailCapacity(capacity, info.Cfs.SizeByte),
		// 文件系统吞吐上限, 吞吐上限是根据文件系统当前已使用存储量、绑定的存储资源包以及吞吐资源包一同确定. 单位MiB/s
		BandwidthLimit: cvt.PtrToVal(info.Cfs.BandwidthLimit),
		Protocol:       cvt.PtrToVal(info.Cfs.Protocol),
		StorageType:    cvt.PtrToVal(info.Cfs.StorageType),
		Encrypted:      cvt.PtrToVal(info.Cfs.Encrypted),
		CryptKeyId:     cvt.PtrToVal(info.Cfs.KmsKeyId),
		//
		CloudVpcIDs:    []string{cvt.PtrToVal(info.MountInfo.VpcId)},
		CloudSubnetIDs: []string{cvt.PtrToVal(info.MountInfo.SubnetId)},
		VpcIDs:         []string{cvt.PtrToVal(info.MountInfo.VpcId)},
		SubnetIDs:      []string{cvt.PtrToVal(info.MountInfo.SubnetId)},
		//
		Extension:        extension,
		Status:           cvt.PtrToVal(info.Cfs.LifeCycleState),
		CloudCreatedTime: cvt.PtrToVal(info.Cfs.CreationTime),
	}

	return req
}

// newDBCreateReq db req
func newDBCreateReq(r *hcservice.TCloudCreateCfsReq, info *typecfs.StorageInfo,
) *protocloud.CfsCreateReq[corecfs.TCloudCfsExtension] {

	extension := &corecfs.TCloudCfsExtension{
		FSID:            cvt.PtrToVal(info.MountInfo.FSID),
		IpAddress:       cvt.PtrToVal(info.MountInfo.IpAddress),
		PGroupId:        cvt.PtrToVal(info.Cfs.PGroup.PGroupId),
		MountInfo:       info.MountInfo,
		AutoScaleUpRule: info.Cfs.AutoScaleUpRule,
		//ExstraPerformanceInfo: info.Cfs.ExstraPerformanceInfo,
		StorageResourcePkg:   info.Cfs.StorageResourcePkg,
		BandwidthResourcePkg: info.Cfs.BandwidthResourcePkg,
		SnapStatus:           info.Cfs.SnapStatus,
		AutoSnapshotPolicyId: info.Cfs.AutoSnapshotPolicyId,
		TieringState:         info.Cfs.TieringState,
		TieringDetail:        info.Cfs.TieringDetail,
		Version:              info.Cfs.Version,
		//MetaType: info.Cfs.MetaType,
	}
	extension.Tags = make(map[string]string)
	for _, tag := range info.Cfs.Tags {
		extension.Tags[cvt.PtrToVal(tag.TagKey)] = cvt.PtrToVal(tag.TagValue)
	}
	capacity := cvt.PtrToVal(info.Cfs.SizeLimit)
	if capacity == 0 {
		capacity = cvt.PtrToVal(info.Cfs.Capacity)
	}

	req := &protocloud.CfsCreateReq[corecfs.TCloudCfsExtension]{
		BkBizID:   r.BkBizID,
		AccountID: r.AccountID,
		Name:      r.Name,
		Region:    r.Region,
		Zone:      r.Zone,
		CloudID:   cvt.PtrToVal(info.Cfs.FileSystemId),
		// 文件系统空间限制. 单位:GiB
		SizeLimit: capacity,
		// 文件系统已使用容量. 单位：Byte
		SizeByte: cvt.PtrToVal(info.Cfs.SizeByte),
		// 文件系统剩余容量. 单位：Byte. 示例值：10
		AvailCapacity: calcAvailCapacity(capacity, info.Cfs.SizeByte),
		// 文件系统吞吐上限, 吞吐上限是根据文件系统当前已使用存储量、绑定的存储资源包以及吞吐资源包一同确定. 单位MiB/s
		BandwidthLimit: cvt.PtrToVal(info.Cfs.BandwidthLimit),
		Protocol:       cvt.PtrToVal(info.Cfs.Protocol),
		StorageType:    cvt.PtrToVal(info.Cfs.StorageType),
		Encrypted:      cvt.PtrToVal(info.Cfs.Encrypted),
		CryptKeyId:     cvt.PtrToVal(info.Cfs.KmsKeyId),
		//
		CloudVpcIDs:    []string{cvt.PtrToVal(r.CloudVpcID)},
		CloudSubnetIDs: []string{cvt.PtrToVal(r.CloudSubnetID)},
		VpcIDs:         []string{cvt.PtrToVal(r.CloudVpcID)},
		SubnetIDs:      []string{cvt.PtrToVal(r.CloudSubnetID)},
		//
		Memo:             r.Memo,
		Extension:        extension,
		Status:           cvt.PtrToVal(info.Cfs.LifeCycleState),
		CloudCreatedTime: cvt.PtrToVal(info.Cfs.CreationTime),
	}

	return req
}

// newListOption return *typecfs.TCloudStorageListOption
func newListOption(req *hcservice.TCloudListCfsReq) *typecfs.TCloudListStorageOption {
	return &typecfs.TCloudListStorageOption{
		Region: req.Region,
		// 选填
		CloudID:  req.CloudID,
		VpcId:    req.VpcId,
		SubnetId: req.SubnetId,
		Offset:   req.Offset,
		Limit:    req.Limit,
		//CreationToken: req.CreationToken,
	}
}

// newGetOption return *typecfs.TCloudGetStorageOption
func newGetOption(req *hcservice.TCloudGetCfsReq) *typecfs.TCloudGetStorageOption {
	return &typecfs.TCloudGetStorageOption{
		Region:   req.Region,
		CloudID:  req.CloudID,
		VpcId:    req.VpcId,
		SubnetId: req.SubnetId,
	}
}

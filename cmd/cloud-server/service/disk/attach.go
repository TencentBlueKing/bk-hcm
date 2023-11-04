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

package disk

import (
	"fmt"

	cloudproto "hcm/pkg/api/cloud-server/disk"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// AttachDisk attach disk.
func (svc *diskSvc) AttachDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.attachDisk(cts, handler.ResOperateAuth)
}

// AttachBizDisk  attach biz disk.
func (svc *diskSvc) AttachBizDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.attachDisk(cts, handler.BizOperateAuth)
}

func (svc *diskSvc) attachDisk(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {

	req := new(cloudproto.AttachReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 鉴权和校验资源分配状态和回收状态
	basicInfos, err := svc.associateValidate(cts, validHandler, req)
	if err != nil {
		logs.Errorf("associate validate failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 创建硬盘主机关联审计
	operationInfo := protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.DiskAuditResType,
		ResID:             req.DiskID,
		Action:            protoaudit.Associate,
		AssociatedResType: enumor.CvmAuditResType,
		AssociatedResID:   req.CvmID,
	}
	if err := svc.audit.ResOperationAudit(cts.Kit, operationInfo); err != nil {
		logs.Errorf("create associate disk audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	var vendor enumor.Vendor
	for _, info := range basicInfos {
		vendor = info.Vendor
		break
	}

	switch vendor {
	case enumor.TCloud:
		return svc.tcloudAttachDisk(cts, req)
	case enumor.Aws:
		return svc.awsAttachDisk(cts)
	case enumor.HuaWei:
		return svc.huaweiAttachDisk(cts, req)
	case enumor.Gcp:
		return svc.gcpAttachDisk(cts, req)
	case enumor.Azure:
		return svc.azureAttachDisk(cts)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

func (svc *diskSvc) associateValidate(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler,
	req *cloudproto.AttachReq) (map[string]types.CloudResourceBasicInfo, error) {

	basicReq := &cloud.BatchListResourceBasicInfoReq{
		Items: []cloud.ListResourceBasicInfoReq{
			{ResourceType: enumor.DiskCloudResType, IDs: []string{req.DiskID}, Fields: types.ResWithRecycleBasicFields},
			{ResourceType: enumor.CvmCloudResType, IDs: []string{req.CvmID}, Fields: types.ResWithRecycleBasicFields},
		},
	}

	basicInfos, err := svc.client.DataService().Global.Cloud.BatchListResBasicInfo(cts.Kit, basicReq)
	if err != nil {
		logs.Errorf("batch list resource basic info failed, err: %v, req: %+v, rid: %s", err, basicReq, cts.Kit.Rid)
		return nil, err
	}

	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Eip,
		Action: meta.Associate, BasicInfos: basicInfos})
	if err != nil {
		return nil, err
	}

	return basicInfos, nil
}

func (svc *diskSvc) tcloudAttachDisk(cts *rest.Contexts, req *cloudproto.AttachReq) (interface{}, error) {

	return nil, svc.client.HCService().TCloud.Disk.AttachDisk(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.TCloudDiskAttachReq{
			CvmID:  req.CvmID,
			DiskID: req.DiskID,
		},
	)
}

func (svc *diskSvc) huaweiAttachDisk(cts *rest.Contexts, req *cloudproto.AttachReq) (interface{}, error) {

	return nil, svc.client.HCService().HuaWei.Disk.AttachDisk(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.HuaWeiDiskAttachReq{
			CvmID:  req.CvmID,
			DiskID: req.DiskID,
		},
	)
}

func (svc *diskSvc) gcpAttachDisk(cts *rest.Contexts, req *cloudproto.AttachReq) (interface{}, error) {

	return nil, svc.client.HCService().Gcp.Disk.AttachDisk(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.GcpDiskAttachReq{
			CvmID:  req.CvmID,
			DiskID: req.DiskID,
		},
	)
}

func (svc *diskSvc) azureAttachDisk(cts *rest.Contexts) (interface{}, error) {

	req := new(cloudproto.AzureDiskAttachReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, svc.client.HCService().Azure.Disk.AttachDisk(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.AzureDiskAttachReq{
			CvmID:       req.CvmID,
			DiskID:      req.DiskID,
			CachingType: req.CachingType,
		},
	)
}

func (svc *diskSvc) awsAttachDisk(cts *rest.Contexts) (interface{}, error) {

	req := new(cloudproto.AwsDiskAttachReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return nil, svc.client.HCService().Aws.Disk.AttachDisk(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		&hcproto.AwsDiskAttachReq{
			DiskID:     req.DiskID,
			CvmID:      req.CvmID,
			DeviceName: req.DeviceName,
		},
	)
}

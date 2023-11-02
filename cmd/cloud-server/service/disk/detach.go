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
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// DetachDisk detach disk.
func (svc *diskSvc) DetachDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.detachDisk(cts, handler.ResValidWithAuth)
}

// DetachBizDisk  detach biz disk.
func (svc *diskSvc) DetachBizDisk(cts *rest.Contexts) (interface{}, error) {
	return svc.detachDisk(cts, handler.BizValidWithAuth)
}

func (svc *diskSvc) detachDisk(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	req := new(cloudproto.DiskDetachReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// get cvm id
	rels, err := svc.client.DataService().Global.ListDiskCvmRel(
		cts.Kit,
		&core.ListReq{
			Filter: tools.EqualExpression("disk_id", req.DiskID),
			Page:   core.NewDefaultBasePage(),
		},
	)
	if len(rels.Details) == 0 {
		return nil, fmt.Errorf("disk(%s) not attached", req.DiskID)
	}

	cvmID := rels.Details[0].CvmID

	// 鉴权和校验资源分配状态和回收状态
	basicReq := &cloud.BatchListResourceBasicInfoReq{
		Items: []cloud.ListResourceBasicInfoReq{
			{ResourceType: enumor.DiskCloudResType, IDs: []string{req.DiskID}, Fields: types.ResWithRecycleBasicFields},
			{ResourceType: enumor.CvmCloudResType, IDs: []string{cvmID}, Fields: types.ResWithRecycleBasicFields},
		},
	}

	basicInfos, err := svc.client.DataService().Global.Cloud.BatchListResBasicInfo(cts.Kit, basicReq)
	if err != nil {
		logs.Errorf("batch list resource basic info failed, err: %v, req: %+v, rid: %s", err, basicReq, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Disk,
		Action: meta.Disassociate, BasicInfos: basicInfos})
	if err != nil {
		return nil, err
	}

	var vendor enumor.Vendor
	for _, info := range basicInfos {
		vendor = info.Vendor
		break
	}

	err = svc.diskLgc.DetachDisk(cts.Kit, vendor, cvmID, req.DiskID)
	if err != nil {
		logs.Errorf("detach disk failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

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
 */

package disk

import (
	"encoding/json"
	"fmt"

	"hcm/cmd/cloud-server/service/common"
	cloudserver "hcm/pkg/api/cloud-server"
	csdisk "hcm/pkg/api/cloud-server/disk"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// CreateDisk create disk.
func (svc *diskSvc) CreateDisk(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create disk request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Disk, Action: meta.Create, ResourceID: req.AccountID}}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("create disk auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	info, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch info.Vendor {
	case enumor.TCloud:
		return svc.createTCloudDisk(cts.Kit, req.Data)
	case enumor.Aws:
		return svc.createAwsDisk(cts.Kit, req.Data)
	case enumor.HuaWei:
		return svc.createHuaWeiDisk(cts.Kit, req.Data)
	case enumor.Gcp:
		return svc.createGcpDisk(cts.Kit, req.Data)
	case enumor.Azure:
		return svc.createAzureDisk(cts.Kit, req.Data)
	default:
		return nil, fmt.Errorf("vendor: %s not support", info.Vendor)
	}
}

func (svc *diskSvc) createTCloudDisk(kt *kit.Kit, body json.RawMessage) (interface{}, error) {
	req := new(csdisk.TCloudDiskCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().TCloud.Disk.CreateDisk(kt.Ctx, kt.Header(),
		common.ConvTCloudDiskCreateReq(req))
	if err != nil {
		logs.Errorf("create tcloud disk failed, err: %v, result: %v, rid: %s", err, result, kt.Rid)
		return result, err
	}

	return result, nil
}

func (svc *diskSvc) createHuaWeiDisk(kt *kit.Kit, body json.RawMessage) (interface{}, error) {
	req := new(csdisk.HuaWeiDiskCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().HuaWei.Disk.CreateDisk(kt.Ctx, kt.Header(),
		common.ConvHuaWeiDiskCreateReq(req))
	if err != nil {
		logs.Errorf("create huawei disk failed, err: %v, result: %v, rid: %s", err, result, kt.Rid)
		return result, err
	}

	return nil, nil
}

func (svc *diskSvc) createAwsDisk(kt *kit.Kit, body json.RawMessage) (interface{}, error) {
	req := new(csdisk.AwsDiskCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().Aws.Disk.CreateDisk(kt.Ctx, kt.Header(),
		common.ConvAwsDiskCreateReq(req))
	if err != nil {
		logs.Errorf("create aws disk failed, err: %v, result: %v, rid: %s", err, result, kt.Rid)
		return result, err
	}

	return nil, nil
}

func (svc *diskSvc) createGcpDisk(kt *kit.Kit, body json.RawMessage) (interface{}, error) {
	req := new(csdisk.GcpDiskCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().Gcp.Disk.CreateDisk(kt.Ctx, kt.Header(),
		common.ConvGcpDiskCreateReq(req))
	if err != nil {
		logs.Errorf("create gcp disk failed, err: %v, result: %v, rid: %s", err, result, kt.Rid)
		return result, err
	}

	return nil, nil
}

func (svc *diskSvc) createAzureDisk(kt *kit.Kit, body json.RawMessage) (interface{}, error) {
	req := new(csdisk.AzureDiskCreateReq)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().Azure.Disk.CreateDisk(kt.Ctx, kt.Header(),
		common.ConvAzureDiskCreateReq(req))
	if err != nil {
		logs.Errorf("create azure disk failed, err: %v, result: %v, rid: %s", err, result, kt.Rid)
		return result, err
	}

	return nil, nil
}

/*
 *
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
// 处理cfs list请求
package cfs

import (
	"encoding/json"
	"fmt"

	cloudserver "hcm/pkg/api/cloud-server"
	protocfs "hcm/pkg/api/hc-service/cfs"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// ListCfsStorage 查询cfs
func (svc *cfsSvc) ListCfsStorage(cts *rest.Contexts) (any, error) {
	req := new(cloudserver.ResourceListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("list cfs storage request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.CfsStorage, Action: meta.Find,
		ResourceID: req.AccountID}}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("list cfs storage auth failed, err: %v, account id: %s, rid: %s", err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.AccountCloudResType,
		req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch accountInfo.Vendor {
	case enumor.TCloud:
		return svc.listTCloudCfsStorage(cts.Kit, req.Data)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

// listTCloudCfsStorage 调用hc-service 查询cfs
func (svc *cfsSvc) listTCloudCfsStorage(kt *kit.Kit, rawReq json.RawMessage) (any, error) {
	req := new(protocfs.TCloudListCfsReq)

	if err := json.Unmarshal(rawReq, req); err != nil {
		logs.Errorf("list cfs storage request decode failed, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		logs.Errorf("list cfs storage request validate failed, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().TCloud.Cfs.ListCfsStorage(kt, req)
	if err != nil {
		logs.Errorf("list cfs storage failed, err: %v, req: %v, rid: %s", err, converter.PtrToVal(req), kt.Rid)
		return nil, err
	}

	return result, nil
}

// GetCfsStorage 查询cfs
func (svc *cfsSvc) GetCfsStorage(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudserver.ResourceListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("get cfs storage request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	//// note: 暂时不验证权限
	//authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.CfsStorage, Action: meta.Find,
	//	ResourceID: req.AccountID}}
	//if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
	//	logs.Errorf("list cfs storage auth failed, err: %v, account id: %s, rid: %s", err, req.AccountID, cts.Kit.Rid)
	//	return nil, err
	//}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.AccountCloudResType,
		req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch accountInfo.Vendor {
	case enumor.TCloud:
		return svc.getTCloudCfsStorage(cts.Kit, req.Data)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}
}

// getTCloudCfsStorage 调用hc-service 查询cfs
func (svc *cfsSvc) getTCloudCfsStorage(kt *kit.Kit, rawReq json.RawMessage) (any, error) {
	req := new(protocfs.TCloudGetCfsReq)

	if err := json.Unmarshal(rawReq, req); err != nil {
		logs.Errorf("get cfs storage request decode failed, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		logs.Errorf("get cfs storage request validate failed, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.client.HCService().TCloud.Cfs.GetCfsStorage(kt, req)
	if err != nil {
		logs.Errorf("get cfs storage failed, err: %v, req: %v, rid: %s", err, converter.PtrToVal(req), kt.Rid)
		return nil, err
	}

	return result, nil
}

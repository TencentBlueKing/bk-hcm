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

package cos

import (
	"fmt"
	"net/http"

	"hcm/cmd/hc-service/service/capability"
	tcloudadaptor "hcm/pkg/adaptor/tcloud"
	typecos "hcm/pkg/adaptor/types/cos"
	protocos "hcm/pkg/api/hc-service/cos"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

func (svc *cosSvc) initTCloudZiyanCosService(cap *capability.Capability) {
	h := rest.NewHandler()
	h.Add("CreateTCloudZiyanCosBucket", http.MethodPost, "/vendors/tcloud-ziyan/cos/buckets/create",
		svc.CreateTCloudZiyanCosBucket)
	h.Add("DeleteTCloudZiyanCosBucket", http.MethodDelete, "/vendors/tcloud-ziyan/cos/buckets/delete",
		svc.DeleteTCloudZiyanCosBucket)
	h.Add("ListTCloudZiyanCosBucket", http.MethodPost, "/vendors/tcloud-ziyan/cos/buckets/list",
		svc.ListTCloudZiyanCosBucket)
	h.Load(cap.WebService)
}

// CreateTCloudZiyanCosBucket create tcloud ziyan cos bucket.
func (svc *cosSvc) CreateTCloudZiyanCosBucket(cts *rest.Contexts) (interface{}, error) {
	req := new(protocos.TCloudCreateBucketReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tcloud, err := svc.ad.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	// 获取 AppID
	accountInfo, err := tcloud.GetAccountInfoBySecret(cts.Kit)
	if err != nil {
		logs.Errorf("get account info by secret failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("get account info failed, err: %v", err)
	}

	opt := &typecos.TCloudBucketCreateOption{
		Name:                 req.Name,
		Region:               req.Region,
		AppID:                accountInfo.AppID,
		XCosACL:              req.XCosACL,
		XCosGrantRead:        req.XCosGrantRead,
		XCosGrantWrite:       req.XCosGrantWrite,
		XCosGrantFullControl: req.XCosGrantFullControl,
		XCosGrantReadACP:     req.XCosGrantReadACP,
		XCosGrantWriteACP:    req.XCosGrantWriteACP,
		XCosTagging:          req.XCosTagging,
	}
	if req.CreateBucketConfiguration != nil {
		opt.CreateBucketConfiguration = &typecos.CreateBucketConfiguration{
			BucketAZConfig: req.CreateBucketConfiguration.BucketAZConfig,
		}
	}

	if err = tcloud.CreateBucket(cts.Kit, opt); err != nil {
		logs.Errorf("tcloud ziyan create bucket failed, err: %v, req: %+v, rid: %s", err, converter.PtrToVal(req),
			cts.Kit.Rid)
		return nil, err
	}

	return protocos.TCloudCreateBucketResp{
		CloudName: tcloudadaptor.EnsureBucketNameWithAppID(req.Name, accountInfo.AppID),
	}, nil
}

// DeleteTCloudZiyanCosBucket delete tcloud ziyan cos bucket.
func (svc *cosSvc) DeleteTCloudZiyanCosBucket(cts *rest.Contexts) (interface{}, error) {
	req := new(protocos.TCloudDeleteBucketReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tcloud, err := svc.ad.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	// 获取 AppID
	accountInfo, err := tcloud.GetAccountInfoBySecret(cts.Kit)
	if err != nil {
		logs.Errorf("get account info by secret failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("get account info failed, err: %v", err)
	}

	opt := &typecos.TCloudBucketDeleteOption{
		Name:   req.Name,
		Region: req.Region,
		AppID:  accountInfo.AppID,
	}
	if err = tcloud.DeleteBucket(cts.Kit, opt); err != nil {
		logs.Errorf("tcloud ziyan delete bucket failed, err: %v, req: %+v, rid: %s", err, converter.PtrToVal(req),
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListTCloudZiyanCosBucket list tcloud ziyan cos bucket.
func (svc *cosSvc) ListTCloudZiyanCosBucket(cts *rest.Contexts) (interface{}, error) {
	req := new(protocos.TCloudBucketListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tcloud, err := svc.ad.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecos.TCloudBucketListOption{
		TagKey:     req.TagKey,
		TagValue:   req.TagValue,
		MaxKeys:    req.MaxKeys,
		Marker:     req.Marker,
		Range:      req.Range,
		CreateTime: req.CreateTime,
		Region:     req.Region,
	}
	result, err := tcloud.ListBuckets(cts.Kit, opt)
	if err != nil {
		logs.Errorf("tcloud ziyan list bucket failed, err: %v, req: %+v, rid: %s", err, converter.PtrToVal(req),
			cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

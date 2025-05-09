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

// Package cos ...
package cos

import (
	"net/http"

	"hcm/cmd/hc-service/service/capability"
	typecos "hcm/pkg/adaptor/types/cos"
	protocos "hcm/pkg/api/hc-service/cos"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

func (svc *cosSvc) initTCloudCosService(cap *capability.Capability) {
	h := rest.NewHandler()
	h.Add("CreateTCloudCosBucket", http.MethodPost, "/vendors/tcloud/cos/buckets/create", svc.CreateTCloudCosBucket)
	h.Add("DeleteTCloudCosBucket", http.MethodDelete, "/vendors/tcloud/cos/buckets/delete", svc.DeleteTCloudCosBucket)
	h.Add("ListTCloudCosBucket", http.MethodPost, "/vendors/tcloud/cos/buckets/list", svc.ListTCloudCosBucket)
	h.Load(cap.WebService)
}

// CreateTCloudCosBucket create tcloud cos bucket.
func (svc *cosSvc) CreateTCloudCosBucket(cts *rest.Contexts) (interface{}, error) {
	req := new(protocos.TCloudCreateBucketReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tCloud, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecos.TCloudBucketCreateOption{
		Name:                 req.Name,
		Region:               req.Region,
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

	if err = tCloud.CreateBucket(cts.Kit, opt); err != nil {
		logs.Errorf("tcloud create bucket failed, err: %v, req: %+v, rid: %s", err, converter.PtrToVal(req),
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DeleteTCloudCosBucket delete tcloud cos bucket.
func (svc *cosSvc) DeleteTCloudCosBucket(cts *rest.Contexts) (interface{}, error) {
	req := new(protocos.TCloudDeleteBucketReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tCloud, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecos.TCloudBucketDeleteOption{
		Name:   req.Name,
		Region: req.Region,
	}
	if err = tCloud.DeleteBucket(cts.Kit, opt); err != nil {
		logs.Errorf("tcloud delete bucket failed, err: %v, req: %+v, rid: %s", err, converter.PtrToVal(req),
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListTCloudCosBucket list tcloud cos bucket.
func (svc *cosSvc) ListTCloudCosBucket(cts *rest.Contexts) (interface{}, error) {
	req := new(protocos.TCloudBucketListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tCloud, err := svc.ad.TCloud(cts.Kit, req.AccountID)
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
	result, err := tCloud.ListBuckets(cts.Kit, opt)
	if err != nil {
		logs.Errorf("tcloud list bucket failed, err: %v, req: %+v, rid: %s", err, converter.PtrToVal(req), cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

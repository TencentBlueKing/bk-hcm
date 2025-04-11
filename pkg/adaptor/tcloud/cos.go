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

package tcloud

import (
	typescos "hcm/pkg/adaptor/types/cos"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/tools/converter"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// CreateBucket 创建存储桶
// reference: https://cloud.tencent.com/document/product/436/7738
func (t *TCloudImpl) CreateBucket(kt *kit.Kit, opt *typescos.TCloudBucketCreateOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "tcloud bucket create option is required")
	}
	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	cliOpt := &typescos.ClientOpt{UrlType: typescos.UrlWithNameAndRegion, BucketNameAppID: opt.Name, Region: opt.Region}
	client, err := t.clientSet.CosClient(cliOpt)
	if err != nil {
		return err
	}

	req := &cos.BucketPutOptions{
		XCosACL:              opt.XCosACL,
		XCosGrantRead:        opt.XCosGrantRead,
		XCosGrantWrite:       opt.XCosGrantWrite,
		XCosGrantFullControl: opt.XCosGrantFullControl,
		XCosGrantReadACP:     opt.XCosGrantReadACP,
		XCosGrantWriteACP:    opt.XCosGrantWriteACP,
		XCosTagging:          opt.XCosTagging,
	}
	if opt.CreateBucketConfiguration != nil {
		req.CreateBucketConfiguration = &cos.CreateBucketConfiguration{
			BucketAZConfig: string(opt.CreateBucketConfiguration.BucketAZConfig),
		}
	}
	if _, err = client.Bucket.Put(kt.Ctx, req); err != nil {
		return err
	}

	return nil
}

// DeleteBucket 删除存储桶
// reference: https://cloud.tencent.com/document/api/436/7732
func (t *TCloudImpl) DeleteBucket(kt *kit.Kit, opt *typescos.TCloudBucketDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "tcloud bucket delete option is required")
	}
	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	cliOpt := &typescos.ClientOpt{UrlType: typescos.UrlWithNameAndRegion, BucketNameAppID: opt.Name, Region: opt.Region}
	client, err := t.clientSet.CosClient(cliOpt)
	if err != nil {
		return err
	}

	if _, err = client.Bucket.Delete(kt.Ctx); err != nil {
		return err
	}

	return nil
}

// ListBuckets 查询存储桶列表
// reference: https://cloud.tencent.com/document/product/436/8291
func (t *TCloudImpl) ListBuckets(kt *kit.Kit, opt *typescos.TCloudBucketListOption) (*typescos.TCloudBucketListResult,
	error) {

	cliOpt := &typescos.ClientOpt{UrlType: typescos.NormalUrl}
	client, err := t.clientSet.CosClient(cliOpt)
	if err != nil {
		return nil, err
	}

	var getOpt *cos.ServiceGetOptions
	if opt != nil {
		if err = opt.Validate(); err != nil {
			return nil, err
		}

		if opt.TagKey != nil || opt.TagValue != nil || opt.MaxKeys != nil || opt.Marker != nil || opt.Range != nil ||
			opt.CreateTime != nil || opt.Region != nil {
			getOpt = new(cos.ServiceGetOptions)
		}

		if opt.TagKey != nil {
			getOpt.TagKey = converter.PtrToVal(opt.TagKey)
		}
		if opt.TagValue != nil {
			getOpt.TagValue = converter.PtrToVal(opt.TagValue)
		}
		if opt.MaxKeys != nil {
			getOpt.MaxKeys = converter.PtrToVal(opt.MaxKeys)
		}
		if opt.Marker != nil {
			getOpt.Marker = converter.PtrToVal(opt.Marker)
		}
		if opt.Range != nil {
			getOpt.Range = converter.PtrToVal(opt.Range)
		}
		if opt.CreateTime != nil {
			getOpt.CreateTime = converter.PtrToVal(opt.CreateTime)
		}
		if opt.Region != nil {
			getOpt.Region = converter.PtrToVal(opt.Region)
		}
	}

	getOpts := make([]*cos.ServiceGetOptions, 0)
	if getOpt != nil {
		getOpts = append(getOpts, getOpt)
	}
	res, _, err := client.Service.Get(kt.Ctx, getOpts...)
	if err != nil {
		return nil, err
	}

	result := &typescos.TCloudBucketListResult{
		Marker:      res.Marker,
		NextMarker:  res.NextMarker,
		IsTruncated: res.IsTruncated,
	}
	if res.Owner != nil {
		result.Owner = &typescos.Owner{
			UIN:         res.Owner.UIN,
			ID:          res.Owner.ID,
			DisplayName: res.Owner.DisplayName,
		}
	}
	for _, bucket := range res.Buckets {
		result.Buckets = append(result.Buckets, typescos.Bucket{
			Name:         bucket.Name,
			Region:       bucket.Region,
			CreationDate: bucket.CreationDate,
			BucketType:   bucket.BucketType,
		})
	}

	return result, nil
}

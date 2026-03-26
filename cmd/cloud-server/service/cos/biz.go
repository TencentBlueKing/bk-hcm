/*
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

package cos

import (
	"fmt"
	"net/url"
	"strings"

	ziyanlogic "hcm/cmd/cloud-server/logics/ziyan"
	cloudserver "hcm/pkg/api/cloud-server"
	apicore "hcm/pkg/api/core"
	protocos "hcm/pkg/api/hc-service/cos"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/converter"
	"hcm/pkg/ziyan"
)

// CreateBizCosBucket create cos bucket under biz.
func (svc *cosSvc) CreateBizCosBucket(cts *rest.Contexts) (any, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(cloudserver.TCloudBizCreateCosBucketReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("create biz cos bucket decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err = req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.CosBucket, Action: meta.Create,
		ResourceID: req.AccountID}, BizID: bizID}
	if err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("create biz cos bucket auth failed, err: %v, biz: %d, rid: %s", err, bizID, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch accountInfo.Vendor {
	case enumor.TCloudZiyan:
		return svc.createBizTCloudZiyanCosBucket(cts.Kit, bizID, req)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "vendor %s not support", accountInfo.Vendor)
	}
}

func (svc *cosSvc) createBizTCloudZiyanCosBucket(kt *kit.Kit, bizID int64,
	req *cloudserver.TCloudBizCreateCosBucketReq) (any, error) {

	tags, err := ziyanlogic.GenTagsForBizsWithManager(kt, svc.client.DataService(), cmdb.CmdbClient(),
		bizID, req.Manager, req.BakManager)
	if err != nil {
		logs.Errorf("gen tags for biz cos bucket failed, err: %v, biz: %d, rid: %s", err, bizID, kt.Rid)
		return nil, fmt.Errorf("generate biz tag failed, err: %w", err)
	}

	hcReq := &protocos.TCloudCreateBucketReq{
		AccountID:            req.AccountID,
		Region:               req.Region,
		Name:                 req.Name,
		XCosACL:              req.XCosACL,
		XCosGrantRead:        req.XCosGrantRead,
		XCosGrantWrite:       req.XCosGrantWrite,
		XCosGrantFullControl: req.XCosGrantFullControl,
		XCosGrantReadACP:     req.XCosGrantReadACP,
		XCosGrantWriteACP:    req.XCosGrantWriteACP,
		XCosTagging:          tagsToXCosTagging(tags),
	}
	if req.CreateBucketConfiguration != nil {
		hcReq.CreateBucketConfiguration = &protocos.CreateBucketConfiguration{
			BucketAZConfig: req.CreateBucketConfiguration.BucketAZConfig,
		}
	}

	bucketResp, err := svc.client.HCService().TCloudZiyan.Cos.CreateCosBucket(kt, hcReq)
	if err != nil {
		logs.Errorf("create biz cos bucket failed, err: %v, req: %+v, rid: %s", err, converter.PtrToVal(hcReq),
			kt.Rid)
		return nil, err
	}

	return bucketResp, nil
}

// ListBizCosBucket list cos buckets under biz.
func (svc *cosSvc) ListBizCosBucket(cts *rest.Contexts) (any, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(cloudserver.TCloudBizListCosBucketReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("list biz cos bucket decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err = req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.CosBucket, Action: meta.Find,
		ResourceID: req.AccountID}, BizID: bizID}
	if err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("list biz cos bucket auth failed, err: %v, biz: %d, rid: %s", err, bizID, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch accountInfo.Vendor {
	case enumor.TCloudZiyan:
		return svc.listBizTCloudZiyanCosBucket(cts.Kit, bizID, req)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "vendor %s not support", accountInfo.Vendor)
	}
}

func (svc *cosSvc) listBizTCloudZiyanCosBucket(kt *kit.Kit, bizID int64,
	req *cloudserver.TCloudBizListCosBucketReq) (any, error) {

	filterTag, err := getBizFilterTag(kt, svc, bizID)
	if err != nil {
		return nil, err
	}

	hcReq := &protocos.TCloudBucketListReq{
		AccountID: req.AccountID,
		TagKey:    converter.ValToPtr(filterTag.Key),
		TagValue:  converter.ValToPtr(filterTag.Value),
		Region:    converter.ValToPtr(req.Region),
	}
	if req.MaxKeys != nil {
		hcReq.MaxKeys = req.MaxKeys
	}
	if req.Marker != nil {
		hcReq.Marker = req.Marker
	}
	if req.Range != nil {
		hcReq.Range = req.Range
	}
	if req.CreateTime != nil {
		hcReq.CreateTime = req.CreateTime
	}

	resp, err := svc.client.HCService().TCloudZiyan.Cos.ListCosBucket(kt, hcReq)
	if err != nil {
		logs.Errorf("list biz cos bucket failed, err: %v, biz: %d, rid: %s", err, bizID, kt.Rid)
		return nil, err
	}

	return resp, nil
}

// DeleteBizCosBucket delete cos bucket under biz.
func (svc *cosSvc) DeleteBizCosBucket(cts *rest.Contexts) (any, error) {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(cloudserver.TCloudBizDeleteCosBucketReq)
	if err = cts.DecodeInto(req); err != nil {
		logs.Errorf("delete biz cos bucket decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err = req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.CosBucket, Action: meta.Delete,
		ResourceID: req.AccountID}, BizID: bizID}
	if err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("delete biz cos bucket auth failed, err: %v, biz: %d, rid: %s", err, bizID, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(
		cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch accountInfo.Vendor {
	case enumor.TCloudZiyan:
		return svc.deleteBizTCloudZiyanCosBucket(cts.Kit, bizID, req)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "vendor %s not support", accountInfo.Vendor)
	}
}

func (svc *cosSvc) deleteBizTCloudZiyanCosBucket(kt *kit.Kit, bizID int64,
	req *cloudserver.TCloudBizDeleteCosBucketReq) (any, error) {

	// 先按业务标签查询存储桶列表，校验目标存储桶属于当前业务。
	// 腾讯云 ListBuckets 单次最多返回 2000 条，需循环分页遍历直到找到目标 bucket 或遍历完毕。
	filterTag, err := getBizFilterTag(kt, svc, bizID)
	if err != nil {
		return nil, err
	}

	found := false
	var marker *string
	for {
		listReq := &protocos.TCloudBucketListReq{
			AccountID: req.AccountID,
			TagKey:    converter.ValToPtr(filterTag.Key),
			TagValue:  converter.ValToPtr(filterTag.Value),
			Region:    converter.ValToPtr(req.Region),
			Marker:    marker,
		}
		listResp, err := svc.client.HCService().TCloudZiyan.Cos.ListCosBucket(kt, listReq)
		if err != nil {
			logs.Errorf("list biz cos bucket for delete check failed, err: %v, biz: %d, rid: %s", err, bizID, kt.Rid)
			return nil, err
		}

		for _, bucket := range listResp.Buckets {
			if bucket.CloudName == req.CloudName {
				found = true
				break
			}
		}
		if found {
			break
		}

		// 如果结果未截断，说明已遍历完所有存储桶
		if !listResp.IsTruncated {
			break
		}
		marker = converter.ValToPtr(listResp.NextMarker)
	}

	if !found {
		return nil, errf.Newf(errf.RecordNotFound,
			"bucket %s not found in biz %d or does not belong to this biz", req.CloudName, bizID)
	}

	delReq := &protocos.TCloudDeleteBucketReq{
		AccountID: req.AccountID,
		Region:    req.Region,
		Name:      req.CloudName,
	}

	if err = svc.client.HCService().TCloudZiyan.Cos.DeleteCosBucket(kt, delReq); err != nil {
		logs.Errorf("delete biz cos bucket failed, err: %v, name: %s, biz: %d, rid: %s",
			err, req.CloudName, bizID, kt.Rid)
		return nil, err
	}

	return nil, nil
}

// getBizFilterTag 获取业务的二级业务标签用于过滤
func getBizFilterTag(kt *kit.Kit, svc *cosSvc, bizID int64) (apicore.TagPair, error) {
	resourceMeta, err := ziyan.GetResourceMetaByBiz(kt, svc.client.DataService(), cmdb.CmdbClient(), bizID)
	if err != nil {
		logs.Errorf("get resource meta for biz failed, err: %v, biz: %d, rid: %s", err, bizID, kt.Rid)
		return apicore.TagPair{}, fmt.Errorf("get biz resource meta failed, err: %w", err)
	}
	return resourceMeta.GetSyncFilterTag(), nil
}

// tagsToXCosTagging 将 TagPair 列表转换为 COS XCosTagging header 格式。
// 格式：URL编码的 key1=value1&key2=value2
func tagsToXCosTagging(tags []apicore.TagPair) string {
	if len(tags) == 0 {
		return ""
	}

	parts := make([]string, 0, len(tags))
	for _, tag := range tags {
		parts = append(parts, url.QueryEscape(tag.Key)+"="+url.QueryEscape(tag.Value))
	}
	return strings.Join(parts, "&")
}

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

package tcloud

import (
	"fmt"

	typestag "hcm/pkg/adaptor/types/tag"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	tag "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tag/v20180813"
)

// ListTags list tag list
// reference: https://cloud.tencent.com/document/api/651/72275
func (t *TCloudImpl) ListTags(kt *kit.Kit, listOpt *typestag.TCloudTagListOpt) (*typestag.TCloudTagListResult, error) {
	tagClient, err := t.clientSet.TagClient()
	if err != nil {
		return nil, fmt.Errorf("new tag client failed, err: %v", err)
	}
	req := tag.NewGetTagsRequest()
	req.PaginationToken = listOpt.PaginationToken
	req.TagKeys = cvt.SliceToPtr(listOpt.TagKeys)
	req.Category = listOpt.Category
	req.MaxResults = listOpt.Limit

	resp, err := tagClient.GetTagsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud tag failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list tcloud tag failed, err: %v", err)
	}

	details := make([]typestag.TCloudTag, 0, len(resp.Response.Tags))
	for i := range resp.Response.Tags {
		tmpData := typestag.TCloudTag{Tag: resp.Response.Tags[i]}
		details = append(details, tmpData)
	}

	return &typestag.TCloudTagListResult{PaginationToken: resp.Response.PaginationToken, Details: details}, nil
}

// TagResources 为指定的多个云产品的多个云资源统一创建并绑定标签。给多个资源-打多个标签，已有标签会用新的值覆盖，不存在的标签或者值会自动创建
// 注：该接口需绑定标签的资源不存在也不会报错
// reference: https://cloud.tencent.com/document/api/651/72280
func (t *TCloudImpl) TagResources(kt *kit.Kit, tagOpt *typestag.TCloudTagResOpt) (
	*typestag.TCloudTagResourcesResp, error) {

	tagClient, err := t.clientSet.TagClient()
	if err != nil {
		return nil, fmt.Errorf("new tag client failed, err: %v", err)
	}
	req := tag.NewTagResourcesRequest()
	req.ResourceList = cvt.SliceToPtr(tagOpt.ResourceList)
	req.Tags = make([]*tag.Tag, 0, len(tagOpt.Tags))
	for i := range tagOpt.Tags {
		req.Tags = append(req.Tags, &tag.Tag{
			TagKey:   &tagOpt.Tags[i].Key,
			TagValue: &tagOpt.Tags[i].Value,
		})
	}
	resp, err := tagClient.TagResourcesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("batch tag tcloud resources failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return &typestag.TCloudTagResourcesResp{
		RequestId:       cvt.PtrToVal(resp.Response.RequestId),
		FailedResources: resp.Response.FailedResources,
	}, nil
}

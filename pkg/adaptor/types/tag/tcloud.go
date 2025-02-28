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

package tag

import (
	"hcm/pkg/api/core"

	tag "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tag/v20180813"
)

// TCloudTag tcloud tag
type TCloudTag struct {
	*tag.Tag
}

// TCloudTagListOpt tcloud tag list opt
type TCloudTagListOpt struct {
	Limit           *uint64  `json:"limit" validate:"max=1000"`
	TagKeys         []string `json:"tag_keys" validate:"max=20,dive,required"`
	Category        *string  `json:"category" `
	PaginationToken *string  `json:"pagination_token,omitempty"`
}

// TCloudTagListResult tcloud tag list result
type TCloudTagListResult struct {
	Details         []TCloudTag `json:"details"`
	PaginationToken *string     `json:"pagination_token,omitempty"`
}

// TCloudTagResOpt tag resources option
type TCloudTagResOpt struct {

	// 待绑定的云资源，用标准的资源六段式表示。正确的资源六段式请参考：
	// [标准的资源六段式](https://cloud.tencent.com/document/product/598/10606)和
	// [支持标签的云产品及资源描述方式](https://cloud.tencent.com/document/product/651/89122)。
	// N取值范围：0~9
	ResourceList []string `json:"resource_list,omitempty" validate:"max=10,dive,required"`

	// 标签键和标签值。
	// 如果指定多个标签，则会为指定资源同时创建并绑定该多个标签。
	// 同一个资源上的同一个标签键只能对应一个标签值。如果您尝试添加已有标签键，则对应的标签值会更新为新值。
	// 如果标签不存在会为您自动创建标签。
	// N取值范围：0~9
	Tags []core.TagPair `json:"tags,omitempty" validate:"max=10,dive,required"`
}

// TCloudTagResourcesResp tag resources response
type TCloudTagResourcesResp struct {
	RequestId       string                `json:"request_id"`
	FailedResources []*tag.FailedResource `json:"failed_resources,omitempty" `
}

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

package loadbalancer

import (
	"encoding/json"

	"hcm/pkg/criteria/validator"

	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
)

// TCloudDescribeSlaCapacityOption 定义查询性能保障规格参数
type TCloudDescribeSlaCapacityOption struct {
	Region   string   `json:"region" validate:"required"`
	SlaTypes []string `json:"sla_types" validate:"omitempty"`
}

// Validate 校验查询性能保障规格参数
func (opt TCloudDescribeSlaCapacityOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudDescribeSlaCapacityRequest 腾讯云查询性能保障规格请求
type TCloudDescribeSlaCapacityRequest struct {
	*tchttp.BaseRequest

	SlaTypes []string `json:"SlaTypes,omitnil,omitempty" name:"SlaTypes"`
}

// ToJsonString 序列化请求为 JSON 字符串
func (r *TCloudDescribeSlaCapacityRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString 从 JSON 字符串反序列化请求
func (r *TCloudDescribeSlaCapacityRequest) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

// TCloudSlaItem 性能容量型规格项
type TCloudSlaItem struct {
	SlaType    string `json:"SlaType,omitnil,omitempty" name:"SlaType"`
	SlaName    string `json:"SlaName,omitnil,omitempty" name:"SlaName"`
	MaxConn    int64  `json:"MaxConn,omitnil,omitempty" name:"MaxConn"`
	MaxCps     int64  `json:"MaxCps,omitnil,omitempty" name:"MaxCps"`
	MaxOutBits int64  `json:"MaxOutBits,omitnil,omitempty" name:"MaxOutBits"`
	MaxInBits  int64  `json:"MaxInBits,omitnil,omitempty" name:"MaxInBits"`
	MaxQps     int64  `json:"MaxQps,omitnil,omitempty" name:"MaxQps"`
}

// TCloudDescribeSlaCapacityResponseParams 腾讯云查询性能保障规格响应内容
type TCloudDescribeSlaCapacityResponseParams struct {
	// SlaSet 性能保障规格列表
	SlaSet []*TCloudSlaItem `json:"SlaSet,omitnil,omitempty" name:"SlaSet"`
	// RequestId 唯一请求 ID，由服务端生成，每次请求都会返回
	RequestId *string `json:"RequestId,omitnil,omitempty" name:"RequestId"`
}

// TCloudDescribeSlaCapacityResponse 腾讯云查询性能保障规格响应
type TCloudDescribeSlaCapacityResponse struct {
	*tchttp.BaseResponse

	Response *TCloudDescribeSlaCapacityResponseParams `json:"Response"`
}

// ToJsonString 序列化响应为 JSON 字符串
func (r *TCloudDescribeSlaCapacityResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString 从 JSON 字符串反序列化响应
func (r *TCloudDescribeSlaCapacityResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

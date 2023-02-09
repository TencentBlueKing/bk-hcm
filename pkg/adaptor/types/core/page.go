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

package core

import (
	"hcm/pkg/criteria/errf"
)

const (
	// TCloudQueryLimit is tencent cloud maximum query limit
	TCloudQueryLimit = 100
	// AwsQueryLimit is aws maximum query limit
	AwsQueryLimit = 1000
	// AwsMinimumQueryLimit is aws minimum query limit
	AwsMinimumQueryLimit = 5
	// GcpQueryLimit is gcp maximum query limit
	GcpQueryLimit = 500
	// HuaWeiQueryLimit is huawei maximum query limit
	HuaWeiQueryLimit = 2000
)

// TCloudPage defines tencent cloud page option.
type TCloudPage struct {
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
}

// Validate TCloudPage.
func (t TCloudPage) Validate() error {
	if t.Limit == 0 {
		return errf.New(errf.InvalidParameter, "limit is required")
	}

	if t.Limit > TCloudQueryLimit {
		return errf.New(errf.InvalidParameter, "tcloud.limit should <= 100")
	}

	return nil
}

// AwsPage define aws page option.
type AwsPage struct {
	MaxResults *int64  `json:"maxResults"`
	NextToken  *string `json:"nextToken,omitempty"`
}

// Validate aws page extension.
func (a AwsPage) Validate() error {
	if a.MaxResults == nil {
		return nil
	}

	if *a.MaxResults > AwsQueryLimit || *a.MaxResults < AwsMinimumQueryLimit {
		return errf.New(errf.InvalidParameter, "aws.limit should >=5 and <= 1000")
	}

	return nil
}

// GcpPage defines gcp page option.
type GcpPage struct {
	PageSize  int64  `json:"pageSize"`
	PageToken string `json:"pageToken"`
}

// Validate gcp page option.
func (g GcpPage) Validate() error {
	if g.PageSize == 0 {
		return errf.New(errf.InvalidParameter, "gcp.pageSize is required")
	}

	if g.PageSize > GcpQueryLimit {
		return errf.New(errf.InvalidParameter, "gcp.pageSize should <= 500")
	}

	return nil
}

// HuaWeiPage define huawei page option.
type HuaWeiPage struct {
	Limit  *int32  `json:"limit,omitempty"`
	Marker *string `json:"marker,omitempty"`
}

// Validate huawei page extension.
func (h HuaWeiPage) Validate() error {
	if h.Limit == nil {
		return nil
	}

	if *h.Limit > HuaWeiQueryLimit {
		return errf.New(errf.InvalidParameter, "huawei.pageSize should <= 2000")
	}

	return nil
}

// HuaWeiOffsetPage define huawei offset page option.
type HuaWeiOffsetPage struct {
	Limit  *int32  `json:"limit,omitempty"`
	Offset *string `json:"offset,omitempty"`
}

// Validate huawei offset page extension.
func (h HuaWeiOffsetPage) Validate() error {
	if h.Limit == nil {
		return nil
	}

	if *h.Limit > 1000 {
		return errf.New(errf.InvalidParameter, "huawei.pageSize should <= 1000")
	}

	return nil
}

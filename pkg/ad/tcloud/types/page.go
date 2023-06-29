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

package tcloudtypes

import (
	"errors"
	"fmt"

	"hcm/pkg/ad/provider"
	"hcm/pkg/tools/converter"
)

// QueryMaxLimit is tencent cloud maximum query limit
const QueryMaxLimit = 100

// Page define tcloud page.
type Page struct {
	// 偏移量，默认为0。
	Offset int64 `json:"offset"`

	// 返回数量，必填，最大值为100。
	Limit int64 `json:"limit"`
}

// ConvProviderPage conv provider page.
func (p *Page) ConvProviderPage() *provider.Page {
	return &provider.Page{
		Offset: converter.ValToPtr(p.Offset),
		Limit:  converter.ValToPtr(p.Limit),
	}
}

// Validate page.
func (p Page) Validate() error {
	if p.Limit == 0 {
		return errors.New("limit is required")
	}

	if p.Limit > QueryMaxLimit {
		return fmt.Errorf("limit should <= %d", QueryMaxLimit)
	}

	return nil
}

// ParsePage parse provider page to tcloud page.
func ParsePage(source *provider.Page) *Page {
	return &Page{
		Offset: converter.PtrToVal(source.Offset),
		Limit:  converter.PtrToVal(source.Offset),
	}
}

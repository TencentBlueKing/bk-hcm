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

// Package webserver defines web-server api call protocols.
package webserver

import (
	"hcm/pkg/criteria/errf"
)

// ListCloudAreaOption is list cmdb cloud area option.
type ListCloudAreaOption struct {
	Page *ListCloudAreaPage `json:"page"`
	Name string             `json:"name,omitempty"`
}

// Validate ListCloudAreaOption.
func (l *ListCloudAreaOption) Validate() error {
	if l.Page == nil {
		return errf.New(errf.InvalidParameter, "page is required")
	}

	if err := l.Page.Validate(); err != nil {
		return err
	}

	return nil
}

const (
	// ListCloudAreaLimit is the list cloud area page limit.
	ListCloudAreaLimit = 500
)

// ListCloudAreaPage is list cmdb cloud area paging options.
type ListCloudAreaPage struct {
	Start int `json:"start"`
	Limit int `json:"limit"`
}

// Validate ListCloudAreaPage.
func (l *ListCloudAreaPage) Validate() error {
	if l.Limit == 0 || l.Limit > ListCloudAreaLimit {
		return errf.New(errf.InvalidParameter, "page limit is invalid")
	}

	return nil
}

// ListCloudAreaResult is list cmdb cloud area option.
type ListCloudAreaResult struct {
	Count int64       `json:"count"`
	Info  []CloudArea `json:"info"`
}

// CloudArea is cmdb cloud area info.
type CloudArea struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

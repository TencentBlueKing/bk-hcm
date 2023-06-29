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

package provider

import (
	"hcm/pkg/kit"
	"hcm/pkg/tools/json"
)

// VpcInterface define vpc interface.
type VpcInterface interface {
	CreateVpc(kt *kit.Kit, meta *Vpc, opt *VpcCreateOption) (*Vpc, error)
	UpdateVpc(kt *kit.Kit, meta *Vpc, opt *VpcCreateOption) (*Vpc, error)
	DeleteVpc(kt *kit.Kit, meta *Vpc, opt *VpcDeleteOption) error
	ListVpc(kt *kit.Kit, opt *VpcListOption) (*VpcListResult, error)
}

// VpcCreateOption define vpc create option.
type VpcCreateOption struct{}

// VpcUpdateOption define vpc update option.
type VpcUpdateOption struct{}

// VpcDeleteOption define vpc delete option.
type VpcDeleteOption struct{}

// VpcListOption define vpc list option.
type VpcListOption struct {
	Region   string   `json:"region"`
	CloudIDs []string `json:"cloud_ids"`
	Page     *Page    `json:"page"`
}

// VpcListResult vpc list result.
type VpcListResult struct {
	Items    []Vpc `json:"items"`
	NextPage *Page `json:"next_page"`
}

// Vpc define vpc struct.
type Vpc struct {
	CloudID    string          `json:"cloud_id"`
	Name       string          `json:"name"`
	Region     string          `json:"region"`
	Memo       *string         `json:"memo"`
	ExtMessage json.ExtMessage `json:"ext_message"`
}

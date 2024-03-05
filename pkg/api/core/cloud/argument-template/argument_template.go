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

package argstpl

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/rest"
)

// BaseArgsTpl define base argument template.
type BaseArgsTpl struct {
	ID             string              `json:"id"`
	CloudID        string              `json:"cloud_id"`
	Name           string              `json:"name"`
	Vendor         enumor.Vendor       `json:"vendor"`
	BkBizID        int64               `json:"bk_biz_id"`
	AccountID      string              `json:"account_id"`
	Type           enumor.TemplateType `json:"type"`
	Templates      *[]TemplateInfo     `json:"templates"`
	GroupTemplates *[]string           `json:"group_templates"`

	Memo           *string `json:"memo"`
	*core.Revision `json:",inline"`
}

// TemplateInfo define argument template's template info.
type TemplateInfo struct {
	// ip地址、协议端口等
	Address *string `json:"address,omitempty" name:"address"`
	// 备注。
	Description *string `json:"description,omitempty" name:"description"`
}

// ArgsTpl define argument template.
type ArgsTpl[Ext Extension] struct {
	BaseArgsTpl `json:",inline"`
	Extension   *Ext `json:"extension"`
}

// GetID ...
func (at ArgsTpl[T]) GetID() string {
	return at.BaseArgsTpl.ID
}

// GetCloudID ...
func (at ArgsTpl[T]) GetCloudID() string {
	return at.BaseArgsTpl.CloudID
}

// Extension extension.
type Extension interface{}

// ArgsTplCreateResp ...
type ArgsTplCreateResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ArgsTplCreateResult `json:"data"`
}

// ArgsTplCreateResult ...
type ArgsTplCreateResult struct {
	ID string `json:"id"`
}

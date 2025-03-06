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

package cmdb

import "hcm/pkg/criteria/validator"

// GetBizBriefCacheTopoParams define get biz brief cache topo params.
type GetBizBriefCacheTopoParams struct {
	BkBizID int64 `json:"bk_biz_id" validate:"required"`
}

// Validate get biz brief cache topo params.
func (p *GetBizBriefCacheTopoParams) Validate() error {
	return validator.Validate.Struct(p)
}

// GetBizBriefCacheTopoResult define get biz brief cache topo result.
type GetBizBriefCacheTopoResult struct {
	// basic business info
	Biz *BizBase `json:"biz"`
	// the idle set nodes info
	Idle []Node `json:"idle"`
	// the other common nodes
	Nodes []Node `json:"nds"`
}

// Node define node info.
type Node struct {
	// the object of this node, like set or module
	Object string `json:"object_id"`
	// the node's instance id, like set id or module id
	ID int64 `json:"id"`
	// the node's name, like set name or module name
	Name string `json:"name"`
	// only set, module has this field.
	// describe what kind of set or module this node is.
	// 0: normal module or set.
	// >1: special set or module
	Default *int `json:"type,omitempty"`
	// the sub-nodes of current node
	SubNodes []Node `json:"nds"`
}

// BizBase define biz base.
type BizBase struct {
	// business id
	ID int64 `json:"id" bson:"bk_biz_id"`
	// business name
	Name string `json:"name" bson:"bk_biz_name"`
	// describe it's a resource pool business or normal business.
	// 0: normal business
	// >0: special business, like resource pool business.
	Default int `json:"type" bson:"default"`

	OwnerID string `json:"bk_supplier_account" bson:"bk_supplier_account"`
}

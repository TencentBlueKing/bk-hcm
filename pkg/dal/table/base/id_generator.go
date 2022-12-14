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

package base

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/utils"
)

// IDGeneratorColumns defines all the IDGenerator table's columns.
var IDGeneratorColumns = utils.MergeColumns(utils.InsertWithoutPrimaryID, IDGeneratorColumnDescriptor)

// IDGeneratorColumnDescriptor is IDGenerator's column descriptors.
var IDGeneratorColumnDescriptor = utils.ColumnDescriptors{
	{Column: "resource", NamedC: "resource", Type: enumor.String},
	{Column: "max_id", NamedC: "max_id", Type: enumor.String},
}

// IDGenerator define id generator table struct.
type IDGenerator struct {
	Resource string `db:"resource"`
	MaxID    string `db:"max_id"`
}

// TableName is the IDGenerator's database table name.
func (ig IDGenerator) TableName() table.Name {
	return table.IDGenerator
}

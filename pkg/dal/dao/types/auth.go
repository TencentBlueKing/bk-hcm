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

package types

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/table"
	"hcm/pkg/runtime/filter"
)

// ListInstancesOption list instance options.
type ListInstancesOption struct {
	TableName        table.Name         `json:"table_name"`
	Filter           *filter.Expression `json:"filter"`
	Page             *core.BasePage     `json:"page"`
	Priority         filter.Priority    `json:"priority"`
	ResourceIDField  string             `json:"resource_id_field"`
	DisplayNameField string             `json:"display_name_field"`
}

// Validate list instance options.
func (o *ListInstancesOption) Validate(po *core.PageOption) error {
	if len(o.TableName) == 0 {
		return errf.New(errf.InvalidParameter, "table name is required")
	}

	if o.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter is required")
	}

	if o.Page == nil {
		return errf.New(errf.InvalidParameter, "page is required")
	}

	if err := o.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

// InstanceResource define instances resource for iam pull resource callback.
type InstanceResource struct {
	ID          string `db:"id" json:"id"`
	DisplayName string `db:"name" json:"display_name"`
}

// ListInstanceDetails defines the response details of requested ListInstancesOption.
type ListInstanceDetails struct {
	Count   uint64             `json:"count"`
	Details []InstanceResource `json:"details"`
}

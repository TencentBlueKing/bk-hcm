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

package dataservice

import (
	"hcm/pkg/criteria/errf"
	"hcm/pkg/runtime/filter"
)

// BatchDeleteReq is a standard batch delete operation http request for data-service.
type BatchDeleteReq struct {
	Filter *filter.Expression `json:"filter"`
}

// Validate BatchDeleteReq.
func (d *BatchDeleteReq) Validate() error {
	if d.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter is required")
	}

	return nil
}

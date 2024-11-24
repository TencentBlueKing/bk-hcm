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

package bill

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/rest"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bssintl/v2/model"
)

// -------------------------- List --------------------------

// HuaWeiBillListResult define huawei bill list result.
type HuaWeiBillListResult struct {
	Count    *int32      `json:"count"`
	Details  interface{} `json:"details"`
	Currency *string     `json:"currency"`
}

// HuaWeiBillListResp define huawei bill list resp.
type HuaWeiBillListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *HuaWeiBillListResult `json:"data"`
}

// HuaWeiRootBillListResult define huawei root bill list result.
type HuaWeiRootBillListResult struct {
	Count    int32                  `json:"count"`
	Details  []model.ResFeeRecordV2 `json:"details"`
	Currency enumor.CurrencyCode    `json:"currency"`
}

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
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/kit"
)

const (
	defaultAccountListLimit          = uint64(500)
	defaultControllerSyncDuration    = 30 * time.Second
	defaultControllerSummaryDuration = 30 * time.Second
	defaultDailySummaryDuration      = 30 * time.Second
	defaultDailySplitDuration        = 30 * time.Second
	defaultSleepMillisecond          = 2000
)

func getInternalKit() *kit.Kit {
	newKit := kit.New()
	newKit.User = string(cc.AccountServerName)
	newKit.AppCode = string(cc.AccountServerName)
	// 设置后端操作的租户id
	newKit.SetBackendTenantID()

	return newKit
}

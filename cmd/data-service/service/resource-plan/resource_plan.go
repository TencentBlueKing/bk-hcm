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

// Package resourceplan ...
package resourceplan

import (
	"hcm/cmd/data-service/service/capability"
	resplandemand "hcm/cmd/data-service/service/resource-plan/res-plan-demand"
	demandchangelog "hcm/cmd/data-service/service/resource-plan/res-plan-demand-changelog"
	"hcm/cmd/data-service/service/resource-plan/res-plan-demand-gpu-order"
	demandgputemplate "hcm/cmd/data-service/service/resource-plan/res-plan-demand-gpu-template"
	demandpenaltybase "hcm/cmd/data-service/service/resource-plan/res-plan-demand-penalty-base"
	resplansubticket "hcm/cmd/data-service/service/resource-plan/res-plan-sub-ticket"
	transferappliedrecord "hcm/cmd/data-service/service/resource-plan/res-plan-transfer-applied-record"
	resplanweek "hcm/cmd/data-service/service/resource-plan/res-plan-week"
	shortrentalreturnedrecord "hcm/cmd/data-service/service/resource-plan/short-rental-returned-record"
)

// InitService initial the resource plan service.
func InitService(cap *capability.Capability) {
	resplandemand.InitService(cap)
	resplandemandgpuorder.InitService(cap)
	demandpenaltybase.InitService(cap)
	demandchangelog.InitService(cap)
	demandgputemplate.InitService(cap)
	resplanweek.InitService(cap)
	transferappliedrecord.InitService(cap)
	resplansubticket.InitService(cap)
	shortrentalreturnedrecord.InitService(cap)
}

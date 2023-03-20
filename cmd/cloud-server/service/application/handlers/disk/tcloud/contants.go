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

package tcloud

var ChargeTypeValue = map[string]string{
	"PREPAID":          "包年包月",
	"POSTPAID_BY_HOUR": "按需计费",
}

var DiskTypeValue = map[string]string{
	"CLOUD_BASIC":   "普通云硬盘",
	"CLOUD_PREMIUM": "高性能云硬盘",
	"CLOUD_BSSD":    "通用型SSD云硬盘",
	"CLOUD_SSD":     "SSD云硬盘",
	"CLOUD_HSSD":    "增强型SSD云硬盘",
	"CLOUD_TSSD":    "极速型SSD云硬盘",
}

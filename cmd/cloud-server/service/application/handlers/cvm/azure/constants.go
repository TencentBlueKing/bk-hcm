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

package azure

import (
	typecvm "hcm/pkg/adaptor/types/cvm"
)

var (
	DiskTypeNameMap = map[typecvm.AzureDiskType]string{
		typecvm.PremiumLRS:     "高级SSD",
		typecvm.PremiumV2LRS:   "高级SSDv2",
		typecvm.PremiumZRS:     "高级SSD托管磁盘",
		typecvm.StandardLRS:    "标准HDD",
		typecvm.StandardSSDLRS: "标准SSD",
		typecvm.StandardSSDZRS: "标准SSD托管磁盘",
		typecvm.UltraSSDLRS:    "超级磁盘",
	}
)

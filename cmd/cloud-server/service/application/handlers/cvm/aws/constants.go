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

package aws

import (
	typecvm "hcm/pkg/adaptor/types/cvm"
)

var (
	DiskTypeNameMap = map[typecvm.AwsVolumeType]string{
		typecvm.GP3:      "通用型SSD卷(gp3)",
		typecvm.GP2:      "通用型SSD卷(gp2)",
		typecvm.IO1:      "预置IOPS SSD卷(io1)",
		typecvm.IO2:      "预置IOPS SSD卷(io2)",
		typecvm.ST1:      "吞吐量优化型HDD卷(st1)",
		typecvm.SC1:      "Cold HDD卷(sc1)",
		typecvm.Standard: "上一代磁介质卷(standard)",
	}
)

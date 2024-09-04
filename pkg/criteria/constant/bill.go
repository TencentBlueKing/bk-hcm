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

// Package constant 账单状态package
package constant

// 状态(0:默认1:创建存储桶2:设置存储桶权限3:创建成本报告4:检查yml文件5:创建CloudFormation模版99:重新生成cur成本报告100:正常)
const (
	StatusDefault              = 0
	StatusCreateBucket         = 1
	StatusSetBucketPolicy      = 2
	StatusCreateCur            = 3
	StatusCheckYml             = 4
	StatusCreateCloudFormation = 5
	StatusReCurReport          = 99
	StatusSuccess              = 100
)

const (
	// BillExportFolderPrefix 账单导出文件夹前缀
	BillExportFolderPrefix = "bill_export"
)

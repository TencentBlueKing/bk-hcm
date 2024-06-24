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

package constant

const (
	// 状态(0:默认1:创建存储桶2:设置存储桶权限3:创建成本报告4:检查yml文件5:创建CloudFormation模版100:正常)
	StatusDefault              = 0
	StatusCreateBucket         = 1
	StatusSetBucketPolicy      = 2
	StatusCreateCur            = 3
	StatusCheckYml             = 4
	StatusCreateCloudFormation = 5
	StatusSuccess              = 100

	// 一级账号账单汇总状态
	// RootAccountBillSummaryStateAccounting 核算中
	RootAccountBillSummaryStateAccounting = "accounting"
	// RootAccountBillSummaryStateAccounted 已核算
	RootAccountBillSummaryStateAccounted = "accounted"
	// RootAccountBillSummaryStateConfirmed 已确认
	RootAccountBillSummaryStateConfirmed = "confirmed"
	// RootAccountBillSummaryStateSyncing 同步中
	RootAccountBillSummaryStateSyncing = "syncing"
	// RootAccountBillSummaryStateSynced 已同步
	RootAccountBillSummaryStateSynced = "synced"
	// RootAccountBillSummaryStateStop 停止中
	RootAccountBillSummaryStateStop = "stopped"

	// 二级账号账单汇总状态

	// MainAccountBillSummaryStateAccounting 核算中
	MainAccountBillSummaryStateAccounting = "accounting"
	// MainAccountBillSummaryStateAccounted 已核算
	MainAccountBillSummaryStateAccounted = "accounted"
	// MainAccountBillSummaryStateSyncing 同步中
	MainAccountBillSummaryStateSyncing = "syncing"
	// MainAccountBillSummaryStateSynced 已同步
	MainAccountBillSummaryStateSynced = "synced"
	// MainAccountBillSummaryStateStop 停止中
	MainAccountBillSummaryStateStop = "stopped"

	// 二级账号账单拉取状态
	// MainAccountRawBillPullStatePulling 拉取中
	MainAccountRawBillPullStatePulling = "pulling"
	// MainAccountRawBillPullStatePulled 已拉取
	MainAccountRawBillPullStatePulled = "pulled"
	// MainAccountRawBillPullStateSplitted 已分账
	MainAccountRawBillPullStateSplitted = "splitted"
	// MainAccountRawBillPullStateAccounted
	MainAccountRawBillPullStateAccounted = "accounted"
	// MainAccountRawBillPullStateStop 停止中
	MainAccountRawBillPullStateStop = "stopped"
)

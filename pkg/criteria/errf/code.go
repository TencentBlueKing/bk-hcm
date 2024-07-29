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

package errf

// NOTE: 错误码规则
// 20号段 + 5位错误码共7位
// 注意：
// - 特殊错误码, 2030403（未授权）, 内部保留

// common error code.
const (
	OK               int32 = 0
	PermissionDenied int32 = 2030403
)

// Note:
// this scope's error code ranges at [4000000, 4089999], and works for all the scenario
// except sidecar related scenario.
const (
	// Unknown is unknown error, it is always used when an
	// error is wrapped, but the error code is not parsed.
	Unknown int32 = 2000000
	// InvalidParameter means the request parameter  is invalid
	InvalidParameter int32 = 2000001
	// TooManyRequest means the incoming request have already exceeded the max limit.
	// and the incoming request is rejected.
	TooManyRequest int32 = 2000002
	// RecordNotFound means resource not exist.
	RecordNotFound int32 = 2000003
	// DecodeRequestFailed means decode the request body failed.
	DecodeRequestFailed int32 = 2000004
	// UnHealthy means service health check failed, current service is not healthy.
	UnHealthy int32 = 2000005
	// Aborted means the request is aborted because of some unexpected exceptions.
	Aborted int32 = 2000006
	// DoAuthorizeFailed try to do user's operate authorize, but got an error,
	// so we do not know if the user has the permission or not.
	DoAuthorizeFailed int32 = 2000007
	// PartialFailed means batch operation is partially failed.
	PartialFailed int32 = 2000008
	// UserNoAppAccess user no app access.
	UserNoAppAccess int32 = 2000009
	// RecordNotUpdate DB数据一行都没有被更新
	RecordNotUpdate int32 = 2000010
	// RecordDuplicated 数据重复，对应 MySQL Error 1062 (23000)
	RecordDuplicated int32 = 2000011

	// CloudVendorError 云上错误
	CloudVendorError int32 = 2000013
	// LoadBalancerTaskExecuting 当前负载均衡正在变更中
	LoadBalancerTaskExecuting int32 = 2000014
	// BillItemImportBillDateError 账单导入账单日期错误, bill_year/bill_month 不匹配
	BillItemImportBillDateError int32 = 2000015
	// BillItemImportDataError 账单导入数据错误, 账单数据格式不正确
	BillItemImportDataError int32 = 2000016
	// BillItemImportEmptyDataError 账单导入空列表
	BillItemImportEmptyDataError int32 = 2000017
)

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

package enumor

// AccountType is account type.
type AccountType string

const (
	// ResourceAccount 资源账号是可以用于管理该账号资源的账号。
	ResourceAccount AccountType = "resource"
	// RegistrationAccount 登记账号仅用于账号管理，不用与管理该账号下的资源。
	RegistrationAccount AccountType = "registration"
)

// AccountSiteType is site type.
type AccountSiteType string

const (
	// ChinaSite is china site.
	ChinaSite AccountSiteType = "china"
	// InternationalSite is international site.
	InternationalSite AccountSiteType = "international"
)

// AccountSyncStatus is account sync status.
type AccountSyncStatus string

const (
	// NotStart is account not start sync status.
	// TODO: 同步时候考虑未同步时使用什么名称
	NotStart = "not_start"
)

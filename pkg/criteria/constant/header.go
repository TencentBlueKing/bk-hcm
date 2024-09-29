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

// Note:
// This scope is used to define all the constant keys which is used inside and outside
// the HCM system.
const (
	// RidKey is request id header key.
	RidKey = "X-Bkapi-Request-Id"

	// UserKey is operator name header key.
	UserKey = "X-Bkapi-User-Name"

	// AppCodeKey is blueking application code header key.
	AppCodeKey = "X-Bkapi-App-Code"

	// LanguageKey the language key word.
	LanguageKey = "HTTP_BLUEKING_LANGUAGE"

	// BKGWJWTTokenKey is blueking api gateway jwt header key.
	BKGWJWTTokenKey = "X-Bkapi-JWT"

	// TenantIDKey is tenant id header key. TODO confirm it.
	TenantIDKey = "HTTP_BLUEKING_SUPPLIER_ID"

	// RequestSourceKey is blueking hcm request source header key.
	RequestSourceKey = "X-Bkhcm-Request-Source"

	// BKGWAuthKey is blueking api gateway authorization header key.
	BKGWAuthKey = "X-Bkapi-Authorization"
)

const (
	// BKHTTPCookieLanguageKey ...
	BKHTTPCookieLanguageKey = "blueking_language"
)

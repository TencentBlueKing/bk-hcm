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
	// AvailableYes defines default value for region's available state.
	AvailableYes = 1
	// AvailableNo defines available value for region's available state.
	AvailableNo = 2
	// AwsDefaultRegion defines default value for aws's region.
	AwsDefaultRegion = "ap-northeast-1"
	// TCloudAvailbleState defines default value for tcloud region state
	TCloudAvailbleState = "AVAILABLE"
	// AwsAvailbleState If the Region is enabled by default, the output includes the following
	AwsAvailbleState = "opt-in-not-required"
	// AwsAvailbleStateAllow If the Region is not enabled, the output includes the following
	// https://docs.aws.amazon.com/general/latest/gr/rande-manage.html
	AwsAvailbleStateAllow = "opted-in"
	// GcpAvailbleState defines default value for gcp region state
	GcpAvailbleState = "UP"
)

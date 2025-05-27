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
	// UnassignedBiz 是未分配业务使用的标识
	UnassignedBiz = -1
	// AttachedAllBiz 代表账号关联所有业务
	AttachedAllBiz = int64(-1)
	// HostPoolBiz 代表主机池业务，用于更清晰表明是主机池的主机，并且在实现时，能和处理业务主机一样保持相似的逻辑
	HostPoolBiz = -1
)

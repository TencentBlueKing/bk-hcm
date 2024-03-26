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

package common

import corelb "hcm/pkg/api/core/cloud/load-balancer"

// VpcDB ...
type VpcDB struct {
	VpcCloudID string
	VpcID      string
	BkCloudID  int64
}

// TCloudComposedListener  规则和监听器的符合结构，七层监听器可能没有Rule
type TCloudComposedListener struct {
	*corelb.Listener[corelb.TCloudListenerExtension]
	Rule *corelb.BaseTCloudLbUrlRule
}

// GetID ...
func (l TCloudComposedListener) GetID() string {
	return l.ID
}

// GetCloudID ...
func (l TCloudComposedListener) GetCloudID() string {
	return l.CloudID
}

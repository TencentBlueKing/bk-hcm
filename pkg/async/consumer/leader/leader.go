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

// Package leader 主从判断器
package leader

import (
	"hcm/pkg/serviced"
)

// Leader 选主管理
type Leader interface {
	IsLeader() bool
}

// NewLeader 创建一个主节点控制器
func NewLeader(sd serviced.ServiceDiscover) Leader {
	return &leader{
		sd: sd,
	}
}

// leader ...
type leader struct {
	sd serviced.ServiceDiscover
}

// IsLeader 判断是否是主节点
func (al *leader) IsLeader() bool {
	return al.sd.IsMaster()
}

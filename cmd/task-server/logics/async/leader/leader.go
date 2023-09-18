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

package leader

import (
	"hcm/pkg/serviced"
)

// AsyncLeader ...
type AsyncLeader struct {
	sd serviced.ServiceDiscover
}

// NewLeader 创建一个主节点控制器
func NewLeader(sd serviced.ServiceDiscover) *AsyncLeader {
	return &AsyncLeader{
		sd: sd,
	}
}

// IsLeader 判断是否是主节点
func (al *AsyncLeader) IsLeader() bool {
	ret := false

	if al.sd.IsMaster() {
		ret = true
	}

	return ret
}

// TODO：leader关闭
// Close 主节点关闭操作
func (al *AsyncLeader) Close() {
}

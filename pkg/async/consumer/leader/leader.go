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
	"strings"

	"hcm/pkg/cc"
	"hcm/pkg/logs"
	"hcm/pkg/serviced"
)

// Leader 选主管理
type Leader interface {
	IsLeader() bool
	AliveNodes() ([]string, error)
	CurrNode() string
}

var _ Leader = new(leader)

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

// CurrNode return current node key.
func (al *leader) CurrNode() string {
	split := strings.Split(al.sd.CurrentNodeKey(), "/")

	if len(split) > 0 {
		return split[len(split)-1]
	}

	// this should not be happened
	return ""
}

// AliveNodes return current node key.
func (al *leader) AliveNodes() ([]string, error) {

	keys, err := al.sd.GetServiceAllNodeKeys(cc.TaskServerName)
	if err != nil {
		logs.Errorf("get task server all node keys failed, err: %v", err)
		return nil, err
	}

	// 因为只是需要TaskServer全部节点的唯一标识，所以，仅需要TaskServer节点路径下的UUID即可。
	keyUUIDs := make([]string, 0, len(keys))
	for _, one := range keys {
		split := strings.Split(one, "/")
		keyUUIDs = append(keyUUIDs, split[len(split)-1])
	}

	return keyUUIDs, nil
}

// IsLeader 判断是否是主节点
func (al *leader) IsLeader() bool {
	return al.sd.IsMaster()
}

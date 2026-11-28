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

package lock

import (
	"errors"
	"fmt"
	"strings"

	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	etcd3 "go.etcd.io/etcd/client/v3"
)

// IaCPluginKey 返回 IaC 插件注册、升级、启停共用的锁 key。
func IaCPluginKey(pluginName string) string {
	return fmt.Sprintf("/hcm/lock/iac_plugin/%s", pluginName)
}

// TryLockIaCPlugin 尝试按插件名获取分布式锁（不阻塞），锁被占用时返回 Aborted 错误。
func (mux *EtcdMutex) TryLockIaCPlugin(kt *kit.Kit, pluginName string, ttl int64) (etcd3.LeaseID, error) {
	leaseID, err := mux.TryLockWithTTL(IaCPluginKey(pluginName), ttl)
	if err != nil {
		if errors.Is(err, ErrLockFailed) {
			return 0, errf.Newf(errf.Aborted,
				"iac plugin %s is being registered, upgraded or modified, please retry later", pluginName)
		}
		logs.Errorf("try lock iac plugin failed, err: %v, plugin_name: %s, rid: %s", err, pluginName, kt.Rid)
		return 0, err
	}

	return leaseID, nil
}

// UnlockIaCPlugin 释放插件名分布式锁，租约已过期自动释放时忽略错误。
func (mux *EtcdMutex) UnlockIaCPlugin(kt *kit.Kit, pluginName string, leaseID etcd3.LeaseID) {
	if err := mux.UnLock(leaseID); err != nil {
		if strings.Contains(err.Error(), "requested lease not found") {
			return
		}
		logs.Errorf("unlock iac plugin lock failed, err: %v, plugin_name: %s, lease_id: %d, rid: %s",
			err, pluginName, leaseID, kt.Rid)
	}
}

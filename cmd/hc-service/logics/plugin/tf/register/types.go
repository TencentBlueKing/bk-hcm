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

package register

import (
	"errors"
	"fmt"
	"time"

	pluginschema "hcm/cmd/hc-service/logics/plugin/schema"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/lock"
)

// errPluginVersionConflict 本副本的状态与注册表冲突（版本回退、sha256 不一致），进程应退出，且不修改任何记录。
var errPluginVersionConflict = errors.New("conflicts with the registered plugin")

// errMissingPlugin 注册表里有已启用的 tf 插件，但本副本目录里没有，副本应当退出。
var errMissingPlugin = errors.New("enabled plugin is missing in the local plugin directory")

// Options 注册参数。
type Options struct {
	// LockLeaseSeconds 插件名锁的租约时长。
	LockLeaseSeconds int64
	// LockWaitTimeout 等待插件名锁的最长时间，等不到即按注册失败处理。
	LockWaitTimeout time.Duration
}

// Service 实现 tf_provider 插件的启动注册与升级。
//
// 失败一律返回错误，由 app.Run 让进程非 0 退出、交给编排重启，进程内不做重试。
type Service struct {
	dataCli   *dataservice.Client
	etcdMutex *lock.EtcdMutex
	schemaMgr *pluginschema.SchemaManager
	opts      Options
}

// IaCPluginAction 是 tf_provider 启动注册时判定出的动作，表达上层要做什么，而不是底层怎么做。
type IaCPluginAction string

const (
	// IaCPluginActionExit 版本回退：本副本不应继续服务，不修改任何记录，进程以非 0 退出。
	IaCPluginActionExit IaCPluginAction = "exit"
	// IaCPluginActionRegister 需要注册该版本：没有记录则新建，处于 init / failed 的记录则重置后重入。
	IaCPluginActionRegister IaCPluginAction = "register"
	// IaCPluginActionVerify 该版本已注册完成（enabled / disabled）：只校验本架构 sha256，不拉起进程。
	IaCPluginActionVerify IaCPluginAction = "verify"
)

// Validate 校验动作是否合法。
func (a IaCPluginAction) Validate() error {
	switch a {
	case IaCPluginActionExit, IaCPluginActionRegister, IaCPluginActionVerify:
		return nil
	default:
		return fmt.Errorf("unsupported iac plugin action: %s", a)
	}
}

// IaCPluginHashResult 是本架构 sha256 与注册记录比对的结果。
type IaCPluginHashResult string

const (
	// IaCPluginHashResultMatch 记录里的本架构 sha256 与本地一致。
	IaCPluginHashResultMatch IaCPluginHashResult = "match"
	// IaCPluginHashResultMismatch 记录里的本架构 sha256 与本地不一致：同一版本号下文件内容变了，进程应退出，不修改任何记录。
	IaCPluginHashResultMismatch IaCPluginHashResult = "mismatch"
	// IaCPluginHashResultMissing 记录里没有本架构的 sha256（例如首个 arm64 副本第一次启动）：需要补写。
	IaCPluginHashResultMissing IaCPluginHashResult = "missing"
)

// Validate 校验哈希比对结果是否合法。
func (r IaCPluginHashResult) Validate() error {
	switch r {
	case IaCPluginHashResultMatch, IaCPluginHashResultMismatch, IaCPluginHashResultMissing:
		return nil
	default:
		return fmt.Errorf("unsupported iac plugin hash result: %s", r)
	}
}

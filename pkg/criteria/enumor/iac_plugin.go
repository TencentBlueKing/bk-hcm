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

import "fmt"

// IaCPluginState IaC 插件状态。
type IaCPluginState string

const (
	// IaCPluginStateInit 注册或升级中，Schema 尚未同步完成。
	IaCPluginStateInit IaCPluginState = "init"
	// IaCPluginStateEnabled 可用。
	IaCPluginStateEnabled IaCPluginState = "enabled"
	// IaCPluginStateDisabled 管理员停用。
	IaCPluginStateDisabled IaCPluginState = "disabled"
	// IaCPluginStateFailed 失败，message 写明原因。
	IaCPluginStateFailed IaCPluginState = "failed"
)

// Validate 校验插件状态是否合法。
func (s IaCPluginState) Validate() error {
	switch s {
	case IaCPluginStateInit, IaCPluginStateEnabled, IaCPluginStateDisabled, IaCPluginStateFailed:
		return nil
	default:
		return fmt.Errorf("unsupported iac plugin state: %s", s)
	}
}

// IaCPluginProtocol IaC 插件协议。
type IaCPluginProtocol string

const (
	// IaCPluginProtocolTFProvider Terraform provider 可执行文件，由 hc-service 拉起子进程调用。
	IaCPluginProtocolTFProvider IaCPluginProtocol = "tf_provider"
	// IaCPluginProtocolHTTP 三方独立部署的 HTTP 插件，经蓝鲸 API 网关访问。
	IaCPluginProtocolHTTP IaCPluginProtocol = "http"
)

// Validate 校验插件协议是否合法。
func (p IaCPluginProtocol) Validate() error {
	switch p {
	case IaCPluginProtocolTFProvider, IaCPluginProtocolHTTP:
		return nil
	default:
		return fmt.Errorf("unsupported iac plugin protocol: %s", p)
	}
}

// IaCPluginAction 是 tf_provider 启动注册时判定出的动作。
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
	// IaCPluginHashResultMismatch 记录里的本架构 sha256 与本地不一致。
	IaCPluginHashResultMismatch IaCPluginHashResult = "mismatch"
	// IaCPluginHashResultMissing 记录里没有本架构的 sha256。
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

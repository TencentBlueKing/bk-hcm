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

package pool

import "errors"

var (
	// ErrPoolClosed 池已关闭：关闭后的 Acquire、关闭时仍在排队或准备中的请求都返回该错误。
	ErrPoolClosed = errors.New("plugin process pool is closed")
	// ErrTooManyWaiters 等待队列已达 MaxQueueLength，新请求被直接拒绝。
	ErrTooManyWaiters = errors.New("too many requests waiting for plugin process")
	// ErrAcquireTimeout 请求排队超过 MaxQueueWait 仍未拿到 lease。
	ErrAcquireTimeout = errors.New("wait for plugin process timeout")
	// ErrPluginNotEnabled 插件没有可用的池：未启用，或本副本没有该插件。
	ErrPluginNotEnabled = errors.New("plugin is not enabled on this replica")
	// ErrLeaseReleased lease 已归还，不能再通过它调用插件。
	ErrLeaseReleased = errors.New("plugin process lease has been released")
)

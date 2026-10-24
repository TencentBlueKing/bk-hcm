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

import (
	"time"

	"hcm/pkg/kit"
)

// requestState 请求所处位置。三种取值互斥。
type requestState string

const (
	// requestQueued 在等待队列中。
	requestQueued requestState = "queued"
	// requestPending 已离开等待队列，在某个进程的 pending 中，尚未结束。
	requestPending requestState = "pending"
	// requestDone 已交付、已失败或已取消。
	requestDone requestState = "done"
)

// request 一次 Acquire 的等待记录，除 ch 外的字段受 PluginPool.poolLock 保护。
type request struct {
	// config 插件实例的 properties（含凭证），不得进入日志与错误信息。
	config map[string]any
	// key config 的匹配键，入队前算好。
	key string
	// kt 发起这次 Acquire 的 kit。后台拉起和配置用自己的 taskKit，不用这个 Ctx。
	kt         *kit.Kit
	enqueuedAt time.Time
	// attempts 进程侧失败后的重试次数。
	attempts int
	state    requestState
	// ch 容量为 1，每个请求只写入一次（由 requestDone 保证），所以在锁内写入不会阻塞。
	ch chan result
}

// result 请求的最终结果，lease 与 err 二者其一非空。
type result struct {
	lease *Lease
	err   error
}

// finishRequestLocked 写入请求结果，请求已结束时什么也不做并返回 false。
func finishRequestLocked(w *request, r result) bool {
	if w.state == requestDone {
		return false
	}
	w.state = requestDone
	w.ch <- r
	return true
}

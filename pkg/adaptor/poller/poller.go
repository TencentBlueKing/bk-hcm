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

package poller

import (
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

type PollingHandler[T any, R any] interface {
	Done(pollResult R) bool
	Poll(client T, kt *kit.Kit, ids []*string) (R, error)
}

type Poller[T any, R any] struct {
	Handler PollingHandler[T, R]
}

// PollUntilDone ...
func (poller *Poller[T, R]) PollUntilDone(client T, kt *kit.Kit, ids []*string) error {
	// TODO 增加超时控制等有效结束条件
	for {
		pollResult, err := poller.Handler.Poll(client, kt, ids)
		if err != nil {
			logs.Errorf("failed to finish the request:  %v, cloudIDs: %v, rid: %s", err, ids, kt.Rid)
			return err
		}

		if poller.Handler.Done(pollResult) {
			return nil
		}
	}
}

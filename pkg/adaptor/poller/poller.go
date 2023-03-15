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
	"time"

	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// BaseDoneResult ...
type BaseDoneResult struct {
	SuccessCloudIDs []string
	FailedCloudIDs  []string
	FailedMessage   string
}

// PollingHandler polling handler.
type PollingHandler[T any, R any, Result any] interface {
	Done(pollResult R) (bool, *Result)
	Poll(client T, kt *kit.Kit, ids []*string) (R, error)
}

// Poller ...
type Poller[T any, R any, Result any] struct {
	Handler PollingHandler[T, R, Result]
}

// PollUntilDoneOption ...
type PollUntilDoneOption struct {
}

// PollUntilDone ...
func (poller *Poller[T, R, Result]) PollUntilDone(client T, kt *kit.Kit, ids []*string,
	opt *PollUntilDoneOption) (*Result, error) {

	// TODO 增加超时控制等有效结束条件
	for {
		time.Sleep(1 * time.Second)

		pollResult, err := poller.Handler.Poll(client, kt, ids)
		if err != nil {
			logs.Errorf("failed to finish the request:  %v, cloudIDs: %v, rid: %s", err, ids, kt.Rid)
			time.Sleep(1 * time.Second)
			continue
		}

		done, result := poller.Handler.Done(pollResult)
		if done {
			return result, nil
		}
	}
}

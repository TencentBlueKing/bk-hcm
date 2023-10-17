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

package async

import (
	"errors"
	"fmt"
	"time"

	taskserver "hcm/pkg/client/task-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// WaitTaskToEnd 等待异步任务结束
// TODO: 临时方案，等异步任务上线后，将该逻辑去除
func WaitTaskToEnd(kt *kit.Kit, cli *taskserver.Client, id string) error {

	end := time.Now().Add(5 * time.Minute)
	for {
		if time.Now().After(end) {
			return fmt.Errorf("wait timeout, async task: %s is running", id)
		}

		flow, err := cli.GetFlow(kt, id)
		if err != nil {
			logs.Errorf("get flow failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		if flow.State == enumor.FlowFailed {
			return errors.New(flow.Reason.Message)
		}

		if flow.State == enumor.FlowSuccess {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}
}

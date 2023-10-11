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

package task

import (
	"encoding/json"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/dal/table/types"
)

// AsyncFlow ...
type AsyncFlow struct {
	ID            string                `json:"id"`
	Name          string                `json:"name"`
	State         enumor.FlowState      `json:"state"`
	Tasks         []AsyncFlowTask       `json:"tasks"`
	Memo          string                `json:"memo"`
	Reason        *tableasync.Reason    `json:"reason"`
	ShareData     *tableasync.ShareData `json:"share_data"`
	core.Revision `json:",inline"`
}

// AsyncFlowTask ...
type AsyncFlowTask struct {
	ID          string           `json:"id"`
	FlowID      string           `json:"flow_id"`
	FlowName    string           `json:"flow_name"`
	ActionName  string           `json:"action_name"`
	Params      types.JsonField  `json:"params"`
	RetryCount  int              `json:"retry_count"`
	TimeoutSecs int              `json:"timeout_secs"`
	DependOn    []string         `json:"depend_on"`
	State       enumor.TaskState `json:"state"`
	Memo        string           `json:"memo"`
	Reason      types.JsonField  `json:"reason"`
	ShareData   types.JsonField  `json:"share_data"`
}

// AddFlowParameters ...
type AddFlowParameters struct {
	Params []AddFlowParam `json:"params"`
}

// AddFlowParam ...
type AddFlowParam struct {
	ActionName string          `json:"action_name"`
	Param      json.RawMessage `json:"param"`
}

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

package tpl

import (
	"fmt"

	"hcm/cmd/task-server/logics/async/backends/iface"
	taskserver "hcm/pkg/api/task-server"
	"hcm/pkg/criteria/enumor"
)

var templates = map[enumor.TplName]TemplateFlow{
	enumor.TplFirstTest: &FirstTest{},
}

// TemplateFlow flow template
type TemplateFlow interface {
	makeTemplateFlowTasks(flowID string, req *taskserver.AddFlowReq,
		backend iface.Backend) ([]string, error)
}

// TemplateFlowOperator template flow operator
type TemplateFlowOperator struct {
	tplFlow TemplateFlow
}

// SetTemplateFlow set template flow
func (operator *TemplateFlowOperator) SetTemplateFlow(tplName enumor.TplName) error {
	if _, exsit := templates[tplName]; !exsit {
		return fmt.Errorf("unsupported tpl %s", tplName)
	}

	operator.tplFlow = templates[tplName]
	return nil
}

// MakeTemplateFlowTasks make template flow tasks
func (operator *TemplateFlowOperator) MakeTemplateFlowTasks(flowID string, req *taskserver.AddFlowReq,
	backend iface.Backend) ([]string, error) {

	return operator.tplFlow.makeTemplateFlowTasks(flowID, req, backend)
}

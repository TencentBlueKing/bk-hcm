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

// Package producer 异步任务生产者
package producer

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"

	"hcm/pkg/async/backend"
	"hcm/pkg/kit"
)

// Producer 异步任务生产者，提供异步任务下发相关功能。。
type Producer interface {
	AddTemplateFlow(kt *kit.Kit, opt *AddTemplateFlowOption) (id string, err error)
	AddCustomFlow(kt *kit.Kit, opt *AddCustomFlowOption) (id string, err error)
	BatchUpdateCustomFlowState(kt *kit.Kit, opt *UpdateCustomFlowStateOption) error
	RetryFlowTask(kt *kit.Kit, flowID, taskID string) error
	CloneFlow(kt *kit.Kit, flowId string, opt *CloneFlowOption) (id string, err error)
}

var _ Producer = new(producer)

// NewProducer new producer.
func NewProducer(bd backend.Backend, register prometheus.Registerer) (Producer, error) {
	if bd == nil {
		return nil, errors.New("backend is required")
	}

	if register == nil {
		return nil, errors.New("metrics register is required")
	}

	return &producer{
		backend: bd,
		mc:      initMetric(register),
	}, nil
}

// producer ...
type producer struct {
	backend backend.Backend
	mc      *metric
}

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

package consumer

import (
	"time"

	"hcm/pkg/criteria/validator"
)

type options struct {
	// normalIntervalSec default 10s
	normalIntervalSec time.Duration
	// executorWorkersCnt default 10
	executorWorkersCnt int
	// parserWorkersCnt default 5
	parserWorkersCnt int
	// flowScheduleTimeoutSec default 15s
	flowScheduleTimeoutSec time.Duration
}

// tryDefaultValue 设置默认值。
func (opt *options) tryDefaultValue() {
	if opt.normalIntervalSec == 0 {
		opt.normalIntervalSec = 10 * time.Second
	}

	if opt.executorWorkersCnt == 0 {
		opt.executorWorkersCnt = 10
	}

	if opt.parserWorkersCnt == 0 {
		opt.parserWorkersCnt = 5
	}

	if opt.flowScheduleTimeoutSec == 0 {
		opt.flowScheduleTimeoutSec = 15 * time.Second
	}
}

// Validate define options.
func (opt *options) Validate() error {
	return validator.Validate.Struct(opt)
}

// Option orm option func defines.
type Option func(opt *options)

// NormalIntervalSec set normal interval sec.
func NormalIntervalSec(sec int) Option {
	return func(opt *options) {
		opt.normalIntervalSec = time.Duration(sec) * time.Second
	}
}

// FlowScheduleTimeoutSec set flow schedule timeout.
func FlowScheduleTimeoutSec(sec int) Option {
	return func(opt *options) {
		opt.flowScheduleTimeoutSec = time.Duration(sec) * time.Second
	}
}

// ExecutorWorkersCnt set executor worker sec.
func ExecutorWorkersCnt(cnt int) Option {
	return func(opt *options) {
		opt.executorWorkersCnt = cnt
	}
}

// ParserWorkersCnt set parser worker sec.
func ParserWorkersCnt(cnt int) Option {
	return func(opt *options) {
		opt.parserWorkersCnt = cnt
	}
}

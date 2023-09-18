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
	"github.com/prometheus/client_golang/prometheus"

	"hcm/pkg/criteria/validator"
)

type options struct {
	register prometheus.Registerer `json:"register" validate:"required"`
}

// tryDefaultValue 设置默认值。
func (opt *options) tryDefaultValue() {
	if opt.register == nil {
		opt.register = prometheus.DefaultRegisterer
	}
}

// Validate define options.
func (opt *options) Validate() error {
	return validator.Validate.Struct(opt)
}

// Option orm option func defines.
type Option func(opt *options)

// MetricsRegisterer set metrics registerer.
func MetricsRegisterer(register prometheus.Registerer) Option {
	return func(opt *options) {
		opt.register = register
	}
}

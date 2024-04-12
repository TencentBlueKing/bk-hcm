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

package tableasync

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/tools/retry"
)

// Retry define retry relation setting.
type Retry struct {
	// Enable 是否开启重试。
	Enable bool `json:"enable"`
	// Policy 定义重试策略
	Policy *RetryPolicy `json:"policy,omitempty"`
}

// IsEnable 是否可以重试
func (r Retry) IsEnable() bool {
	return r.Enable
}

// Run retry run func.
func (r Retry) Run(do func() (stop bool, result any, err error)) (result any, err error) {
	if !r.IsEnable() {
		return nil, errors.New("retry not enable")
	}

	rp := retry.NewRetryPolicy(r.Policy.Count, r.Policy.SleepRangeMS)
	var lastErr error
	var lastResult any
	var stop bool
	for {
		if rp.RetryCount() >= uint32(r.Policy.Count) {
			break
		}

		stop, result, err = do()
		if stop {
			// 主动停止
			return result, err
		}
		if err != nil {
			lastErr = err
			lastResult = result
			rp.Sleep()
			continue
		}

		return result, nil
	}

	return lastResult, fmt.Errorf("retry exceed the max number of retryable times: %d, lastErr: %v",
		r.Policy.Count, lastErr)
}

// Validate retry.
func (r Retry) Validate() error {
	if !r.Enable && r.Policy != nil {
		return errors.New("retry not enable, policy can not set")
	}

	if r.Enable {
		if r.Policy == nil {
			return errors.New("retry is enable, policy is required")
		}

		if err := r.Policy.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Scan is used to decode raw message which is read from db into Retry.
func (r *Retry) Scan(raw interface{}) error {
	return types.Scan(raw, r)
}

// Value encode the Retry to a json raw, so that it can be stored to db with json raw.
func (r Retry) Value() (driver.Value, error) {
	return types.Value(r)
}

// RetryPolicy define retry policy.
type RetryPolicy struct {
	// Count 重试次数
	Count uint `json:"count" validate:"required"`
	// SleepRangeMS 重试睡眠周期随机数范围
	SleepRangeMS [2]uint `json:"sleep_range_ms" validate:"required,min=2"`
}

// Validate RetryPolicy.
func (rp RetryPolicy) Validate() error {
	return validator.Validate.Struct(rp)
}

// NewRetryWithPolicy return retry with policy
func NewRetryWithPolicy(count, sleepMsMin, sleepMsMax uint) *Retry {
	return &Retry{
		Enable: true,
		Policy: &RetryPolicy{
			Count:        count,
			SleepRangeMS: [2]uint{sleepMsMin, sleepMsMax},
		},
	}
}

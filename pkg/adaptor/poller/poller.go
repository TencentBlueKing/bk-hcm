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

// Package poller ...
package poller

import (
	"time"

	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/retry"
)

// BaseDoneResult ...
type BaseDoneResult struct {
	SuccessCloudIDs []string `json:"success_cloud_ids"`
	FailedCloudIDs  []string `json:"failed_cloud_ids"`
	UnknownCloudIDs []string `json:"unknown_cloud_ids"`
	FailedMessage   string   `json:"failed_message"`
}

// PollingHandler polling handler.
type PollingHandler[T any, R any, Result any] interface {
	// Done 根据 poll 获取的实例数据，判断是否符合预期，如果有实例状态无法判断，返回false
	Done(pollResult R) (ok bool, ret *Result)
	// Poll 通过实例ids查询实例结果，如果查询不到报错
	Poll(client T, kt *kit.Kit, ids []*string) (R, error)
}

// Poller ...
type Poller[T any, R any, Result any] struct {
	Handler PollingHandler[T, R, Result]
}

const (
	DefaultTimeoutTimeSec = 60
)

// PollUntilDoneOption ...
type PollUntilDoneOption struct {
	TimeoutTimeSecond uint64             `json:"timeout_time_second" validate:"required"`
	Retry             *retry.RetryPolicy `json:"retry" validate:"required"`
}

// TrySetDefaultValue ...
func (opt *PollUntilDoneOption) TrySetDefaultValue() {
	if opt == nil {
		return
	}

	if opt.TimeoutTimeSecond == 0 {
		opt.TimeoutTimeSecond = DefaultTimeoutTimeSec
	}

	if opt.Retry == nil {
		opt.Retry = retry.NewRetryPolicy(0, [2]uint{})
	}
}

// Validate ...
func (opt PollUntilDoneOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// PollUntilDone ...
func (poller *Poller[T, R, Result]) PollUntilDone(client T, kt *kit.Kit, ids []*string,
	opt *PollUntilDoneOption,
) (*Result, error) {
	if opt == nil {
		opt = &PollUntilDoneOption{
			TimeoutTimeSecond: 30 * 60,
			Retry:             retry.NewRetryPolicy(10, [2]uint{2000, 30000}),
		}
	}

	opt.TrySetDefaultValue()

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	// 重试次数归位
	opt.Retry.Reset()

	pollerFunc := func() (bool, *Result, error) {
		pollResult, err := poller.Handler.Poll(client, kt, ids)
		if err != nil {
			return false, nil, err
		}

		ok, result := poller.Handler.Done(pollResult)
		return ok, result, nil
	}

	endTime := time.Now().Add(time.Duration(opt.TimeoutTimeSecond) * time.Second)
	for {
		if time.Now().After(endTime) {
			// 达到超时时间，成功多少返回多少，无法判断的统一放到unknown类
			_, result, err := pollerFunc()
			if err != nil {
				logs.Errorf("poll until done timeout, but exec poller func failed, err: %v, ids: %v, timeout: %ds, "+
					"rid: %s", err, converter.SliceToPtr(ids), opt.TimeoutTimeSecond, kt.Rid)
				return nil, err
			}

			logs.V(2).Infof("poll until done timeout, ids: %v, result: %v, rid: %s", converter.SliceToPtr(ids),
				result, kt.Rid)
			return result, nil
		}

		ok, result, err := pollerFunc()
		if err != nil {
			logs.V(3).Errorf("exec poller func failed, err: %v, retryCount: %d, rid: %s", err,
				opt.Retry.RetryCount(), kt.Rid)
			opt.Retry.Sleep()
			continue
		}

		if !ok {
			opt.Retry.Sleep()
			continue
		}

		return result, nil
	}
}

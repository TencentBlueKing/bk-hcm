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

package types

import (
	"hcm/pkg/adaptor/poller"
	"hcm/pkg/tools/retry"
)

// NewBatchCreateCvmPollerOption 超时时间半小时，10次之内重试间隔时间2s，10次之后重试间隔时间2-30s之间
func NewBatchCreateCvmPollerOption() *poller.PollUntilDoneOption {
	return &poller.PollUntilDoneOption{
		TimeoutTimeSecond: 30 * 60,
		Retry:             retry.NewRetryPolicy(10, [2]uint{2000, 30000}),
	}
}

// NewBatchCreateVpcPollerOption 超时时间10分钟，10次之内重试间隔时间2s，10次之后重试间隔时间2-30s之间
func NewBatchCreateVpcPollerOption() *poller.PollUntilDoneOption {
	return &poller.PollUntilDoneOption{
		TimeoutTimeSecond: 30 * 60,
		Retry:             retry.NewRetryPolicy(10, [2]uint{2000, 30000}),
	}
}

// NewBatchCreateSubnetPollerOption 超时时间10分钟，10次之内重试间隔时间2s，10次之后重试间隔时间2-30s之间
func NewBatchCreateSubnetPollerOption() *poller.PollUntilDoneOption {
	return &poller.PollUntilDoneOption{
		TimeoutTimeSecond: 30 * 60,
		Retry:             retry.NewRetryPolicy(10, [2]uint{2000, 30000}),
	}
}

// NewBatchOperateCvmPollerOpt 超时时间10分钟，10次之内重试间隔时间1s，10次之后重试间隔时间1-5s之间
func NewBatchOperateCvmPollerOpt() *poller.PollUntilDoneOption {
	return &poller.PollUntilDoneOption{
		TimeoutTimeSecond: 5 * 60,
		Retry:             retry.NewRetryPolicy(10, [2]uint{1000, 5000}),
	}
}

// NewBatchUpdateArgsTplPollerOption 超时时间5分钟，10次之内重试间隔时间1s，10次之后重试间隔时间1-5s之间
func NewBatchUpdateArgsTplPollerOption() *poller.PollUntilDoneOption {
	return &poller.PollUntilDoneOption{
		TimeoutTimeSecond: 5 * 60,
		Retry:             retry.NewRetryPolicy(10, [2]uint{1000, 5000}),
	}
}

// NewBatchDeleteArgsTplPollerOption 超时时间5分钟，10次之内重试间隔时间1s，10次之后重试间隔时间1-5s之间
func NewBatchDeleteArgsTplPollerOption() *poller.PollUntilDoneOption {
	return &poller.PollUntilDoneOption{
		TimeoutTimeSecond: 5 * 60,
		Retry:             retry.NewRetryPolicy(10, [2]uint{1000, 5000}),
	}
}

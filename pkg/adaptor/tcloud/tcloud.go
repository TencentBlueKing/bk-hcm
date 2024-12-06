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

// Package tcloud cloud operation for tencent cloud
package tcloud

import (
	"context"
	"errors"
	"strings"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/retry"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

// NewTCloud new tcloud.
func NewTCloud(s *types.BaseSecret) (TCloud, error) {
	prof := profile.NewClientProfile()
	if err := validateSecret(s); err != nil {
		return nil, err
	}

	return &TCloudImpl{clientSet: newClientSet(s, prof)}, nil
}

// TCloudImpl is tencent cloud operator.
type TCloudImpl struct {
	clientSet ClientSet
}

// SetClientSet set new client set
func (t *TCloudImpl) SetClientSet(c ClientSet) {
	t.clientSet = c
}

func validateSecret(s *types.BaseSecret) error {
	if s == nil {
		return errf.New(errf.InvalidParameter, "secret is required")
	}

	if err := s.Validate(); err != nil {
		return err
	}

	return nil
}

// SetRateLimitRetryWithRandomInterval determine whether to set the retry parameter after exceeding the rate limit
func (t *TCloudImpl) SetRateLimitRetryWithRandomInterval(retry bool) {
	if retry {
		t.clientSet.SetRateLimitRetryWithConstInterval()
	}
}

func networkRetryable(err error) bool {
	errStr := err.Error()
	if !(strings.Contains(errStr, constant.TCloudNetworkErrorErrCode)) {
		return false
	}
	if !strings.Contains(errStr, "http: ContentLength=") {
		return false
	}
	return true
}

// NetworkErrRetry auto retry for "ClientError.NetworkError", for read only api
func NetworkErrRetry[I any, O any](apiCall func(context.Context, *I) (*O, error), kt *kit.Kit, req *I) (resp *O,
	err error) {

	var retryTime uint32 = constant.TCloudClientErrRetryTimes
	var sleepMin, sleepMax uint = 700, 1500

	policy := retry.NewRetryPolicy(0, [2]uint{sleepMin, sleepMax})
	for policy.RetryCount() < retryTime {
		resp, err = apiCall(kt.Ctx, req)
		if err == nil {
			// break on success
			break
		}
		// retry for network error
		if policy.RetryCount() < retryTime && networkRetryable(err) {
			logs.ErrorDepthf(1, "call tcloud api failed, req: %+v, retry times: %d, err: %v, rid: %s",
				req, policy.RetryCount(), err, kt.Rid)
			policy.Sleep()
			continue
		}
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("empty response from tcloud")
	}
	return resp, nil
}

func getTagFilterKey(k string) *string {
	return cvt.ValToPtr("tag:" + k)
}

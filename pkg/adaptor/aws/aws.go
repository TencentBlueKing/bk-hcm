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

package aws

import (
	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/errf"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
)

// NewAws new aws.
func NewAws() types.Factory {
	return new(amazon)
}

// NewAwsProxy new aws proxy.
func NewAwsProxy() types.AwsProxy {
	return new(amazon)
}

var (
	_ types.Factory  = new(amazon)
	_ types.AwsProxy = new(amazon)
)

type amazon struct{}

func (am *amazon) ec2Client(secret *types.BaseSecret, region string) (*ec2.EC2, error) {
	cfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(secret.ID, secret.Key, ""),
		DisableSSL:  nil,
		HTTPClient:  nil,
		LogLevel:    nil,
		Logger:      nil,
		MaxRetries:  nil,
		Retryer:     nil,
		SleepDelay:  nil,
	}

	if len(region) != 0 {
		cfg.Region = aws.String(region)
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	return ec2.New(sess), nil
}

func (am *amazon) stsClient(secret *types.BaseSecret) (*sts.STS, error) {
	cfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(secret.ID, secret.Key, ""),
		DisableSSL:  nil,
		HTTPClient:  nil,
		LogLevel:    nil,
		Logger:      nil,
		MaxRetries:  nil,
		Retryer:     nil,
		SleepDelay:  nil,
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	return sts.New(sess), nil
}

func validateSecret(s *types.Secret) error {
	if s == nil {
		return errf.New(errf.InvalidParameter, "secret is required")
	}

	if s.Aws == nil {
		return errf.New(errf.InvalidParameter, "aws secret is required")
	}

	if err := s.Aws.Validate(); err != nil {
		return err
	}

	return nil
}

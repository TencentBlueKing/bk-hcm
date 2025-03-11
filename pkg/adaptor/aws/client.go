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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	curservice "github.com/aws/aws-sdk-go/service/costandusagereportservice"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
)

const (
	ErrDataNotFound       = "InvalidInstanceID.Malformed: Invalid id"
	ErrDryRunSuccess      = "DryRunOperation: Request would have succeeded, but DryRun flag is set"
	ErrSGNotFound         = "InvalidGroup.NotFound"
	ErrRouteTableNotFound = "InvalidRouteTableID.NotFound"
	ErrImageNotFound      = "InvalidAMIID.NotFound"
	ErrVpcNotFound        = "InvalidVpcID.NotFound"
	ErrSubnetNotFound     = "InvalidSubnetID.NotFound"
	ErrDiskNotFound       = "InvalidVolume.NotFound"
	ErrCvmNotFound        = "InvalidInstanceID.NotFound"
)

type clientSet struct {
	credentials *credentials.Credentials
}

func newClientSet(secret *types.BaseSecret) *clientSet {
	return &clientSet{credentials.NewStaticCredentials(secret.CloudSecretID, secret.CloudSecretKey, "")}
}

func (c *clientSet) ec2Client(region string) (*ec2.EC2, error) {
	cfg := &aws.Config{
		Credentials: c.credentials,
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

// sts client, if region is nil, use sdk default region
func (c *clientSet) stsClient(region *string) (*sts.STS, error) {
	cfg := &aws.Config{
		Credentials: c.credentials,
		DisableSSL:  nil,
		HTTPClient:  nil,
		LogLevel:    nil,
		Logger:      nil,
		MaxRetries:  nil,
		Retryer:     nil,
		SleepDelay:  nil,
		Region:      region,
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	return sts.New(sess), nil
}

func (c *clientSet) athenaClient(region string) (*athena.Athena, error) {
	cfg := &aws.Config{
		Credentials: c.credentials,
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

	return athena.New(sess, aws.NewConfig().WithRegion(region)), nil
}

func (c *clientSet) organizations() (*organizations.Organizations, error) {
	cfg := &aws.Config{
		Credentials: c.credentials,
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

	return organizations.New(sess), nil
}

func (c *clientSet) s3Client(region string) (*s3.S3, error) {
	cfg := &aws.Config{
		Credentials: c.credentials,
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

	return s3.New(sess, aws.NewConfig().WithRegion(region)), nil
}

func (c *clientSet) costAndUsageReportClient(region string) (*curservice.CostandUsageReportService, error) {
	cfg := &aws.Config{
		Credentials: c.credentials,
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

	return curservice.New(sess, aws.NewConfig().WithRegion(region)), nil
}

func (c *clientSet) cloudFormationClient(region string) (*cloudformation.CloudFormation, error) {
	cfg := &aws.Config{
		Credentials: c.credentials,
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

	return cloudformation.New(sess, aws.NewConfig().WithRegion(region)), nil
}

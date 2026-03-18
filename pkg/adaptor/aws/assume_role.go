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
	"fmt"
	"time"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/enumor"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

const (
	stsRegionInternational = "us-east-1"
	stsRegionChina         = "cn-north-1"

	arnPartitionInternational = "aws"
	arnPartitionChina         = "aws-cn"
)

// AssumeRoleResult holds the temporary credentials returned by STS AssumeRole.
type AssumeRoleResult struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Expiration      time.Time
}

// AssumeRole calls AWS STS AssumeRole API with the given long-term credentials and role ARN,
// returning temporary credentials for cross-account access. externalId is optional; when
// non-empty it is passed to STS for Trust Policy condition verification.
func AssumeRole(secret *types.BaseSecret, roleArn, sessionName, externalID string,
	site enumor.AccountSiteType) (*AssumeRoleResult, error) {

	region := stsRegionInternational
	if site == enumor.ChinaSite {
		region = stsRegionChina
	}

	creds := credentials.NewStaticCredentials(secret.CloudSecretID, secret.CloudSecretKey,
		secret.CloudSessionToken)
	sess, err := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(region),
	})
	if err != nil {
		return nil, fmt.Errorf("create STS session failed, err: %v", err)
	}

	stsClient := sts.New(sess)
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String(sessionName),
	}
	if externalID != "" {
		input.ExternalId = aws.String(externalID)
	}

	output, err := stsClient.AssumeRole(input)
	if err != nil {
		return nil, err
	}

	if output.Credentials == nil {
		return nil, fmt.Errorf("assume role returned nil credentials, roleArn: %s", roleArn)
	}

	return &AssumeRoleResult{
		AccessKeyID:     aws.StringValue(output.Credentials.AccessKeyId),
		SecretAccessKey: aws.StringValue(output.Credentials.SecretAccessKey),
		SessionToken:    aws.StringValue(output.Credentials.SessionToken),
		Expiration:      aws.TimeValue(output.Credentials.Expiration),
	}, nil
}

// BuildRoleArn constructs a full IAM Role ARN from account ID, role name and site type.
// International: arn:aws:iam::<accountID>:role/<roleName>
// China: arn:aws-cn:iam::<accountID>:role/<roleName>
func BuildRoleArn(accountID, roleName string, site enumor.AccountSiteType) string {
	partition := arnPartitionInternational
	if site == enumor.ChinaSite {
		partition = arnPartitionChina
	}
	return fmt.Sprintf("arn:%s:iam::%s:role/%s", partition, accountID, roleName)
}

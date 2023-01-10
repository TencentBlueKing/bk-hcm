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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// tagKeyForResourceName is tag key that define resource name.
const tagKeyForResourceName = "Name"

// tagResourceType is ec2 tag resource type.
// reference: https://docs.aws.amazon.com/zh_cn/AWSEC2/latest/APIReference/API_TagSpecification.html
type tagResourceType string

const (
	vpcTagResType tagResourceType = "vpc"
)

// genNameTags generate name ec2 tags.
func genNameTags(resourceType tagResourceType, name string) []*ec2.TagSpecification {
	if len(name) == 0 {
		return nil
	}

	tagSpec := &ec2.TagSpecification{
		ResourceType: aws.String(string(resourceType)),
		Tags: []*ec2.Tag{{
			Key:   aws.String(tagKeyForResourceName),
			Value: aws.String(name),
		}},
	}

	return []*ec2.TagSpecification{tagSpec}
}

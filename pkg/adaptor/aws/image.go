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
	"hcm/pkg/adaptor/types/image"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// ListImage ...
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_DescribeImages.html
func (a *Aws) ListImage(
	kt *kit.Kit,
	opt *image.AwsImageListOption,
) (*image.AwsImageListResult, error) {
	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req := &ec2.DescribeImagesInput{MaxResults: opt.Page.MaxResults, NextToken: opt.Page.NextToken}
	req.Filters = []*ec2.Filter{
		{Name: aws.String("is-public"), Values: []*string{aws.String("true")}},
	}

	resp, err := client.DescribeImagesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("describe aws image failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	images := make([]image.AwsImage, 0)
	for _, pImage := range resp.Images {
		images = append(images, image.AwsImage{
			CloudID:      *pImage.ImageId,
			Name:         *pImage.Name,
			State:        *pImage.State,
			Architecture: *pImage.Architecture,
			Type:         "public",
		})
	}
	return &image.AwsImageListResult{Details: images, NextToken: resp.NextToken}, nil
}

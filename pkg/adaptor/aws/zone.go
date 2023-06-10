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
	typeszone "hcm/pkg/adaptor/types/zone"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// ListZone list zone
// reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeAvailabilityZones.html
func (a *Aws) ListZone(kit *kit.Kit, opt *typeszone.AwsZoneListOption) ([]typeszone.AwsZone, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "aws zone list option is required")
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req := new(ec2.DescribeAvailabilityZonesInput)

	resp, err := client.DescribeAvailabilityZones(req)
	if err != nil {
		logs.Errorf("failed to list zone, err: %v, rid: %s", err, kit.Rid)
	}

	if resp == nil {
		return make([]typeszone.AwsZone, 0), nil
	}

	results := make([]typeszone.AwsZone, 0)
	for _, one := range resp.AvailabilityZones {
		results = append(results, typeszone.AwsZone{one})
	}

	return results, nil
}

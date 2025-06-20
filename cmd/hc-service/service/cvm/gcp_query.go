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

package cvm

import (
	"fmt"

	"hcm/pkg/api/core"
	coreimage "hcm/pkg/api/core/cloud/image"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
)

func (svc *cvmSvc) getImageByCloudID(kt *kit.Kit, cloudID string) (
	*coreimage.Image[coreimage.GcpExtension], error) {

	req := &core.ListReq{
		Filter: tools.EqualExpression("cloud_id", cloudID),
		Page:   core.NewDefaultBasePage(),
		Fields: nil,
	}
	images, err := svc.dataCli.Gcp.ListImage(kt, req)
	if err != nil {
		return nil, err
	}

	if len(images.Details) == 0 {
		return nil, fmt.Errorf("image: %s not found", cloudID)
	}

	return images.Details[0], nil
}

func (svc *cvmSvc) getSubnetSelfLinkByCloudID(kt *kit.Kit, cloudID string) (string, error) {
	req := &core.ListReq{
		Filter: tools.EqualExpression("cloud_id", cloudID),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"extension"},
	}
	subnets, err := svc.dataCli.Gcp.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		return "", err
	}

	if len(subnets.Details) == 0 {
		return "", fmt.Errorf("subnet: %s not found", cloudID)
	}

	return subnets.Details[0].Extension.SelfLink, nil
}

func (svc *cvmSvc) getVpcSelfLinkByCloudID(kt *kit.Kit, cloudID string) (string, error) {
	req := &core.ListReq{
		Filter: tools.EqualExpression("cloud_id", cloudID),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"extension"},
	}
	vpcs, err := svc.dataCli.Gcp.Vpc.ListVpcExt(kt, req)
	if err != nil {
		return "", err
	}

	if len(vpcs.Details) == 0 {
		return "", fmt.Errorf("vpc: %s not found", cloudID)
	}

	return vpcs.Details[0].Extension.SelfLink, nil
}

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

package tcloud

import (
	protoclb "hcm/pkg/api/hc-service/clb"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewClbClient create a new clb api client.
func NewClbClient(client rest.ClientInterface) *ClbClient {
	return &ClbClient{
		client: client,
	}
}

// ClbClient is hc service clb api client.
type ClbClient struct {
	client rest.ClientInterface
}

// BatchSetTCloudClbSecurityGroup batch set clb security group resource.
func (cli *ClbClient) BatchSetTCloudClbSecurityGroup(kt *kit.Kit,
	request *protoclb.TCloudSetClbSecurityGroupReq) error {

	return common.RequestNoResp[protoclb.TCloudSetClbSecurityGroupReq](cli.client, rest.POST, kt, request,
		"/clbs/batch/security_groups")
}

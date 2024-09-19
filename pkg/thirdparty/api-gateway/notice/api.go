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

package notice

import (
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	apigateway "hcm/pkg/thirdparty/api-gateway"
)

// GetCurAnn get current announcements
func (n *notice) GetCurAnn(kt *kit.Kit, params map[string]string) (GetCurAnnResp, error) {

	resp, err := apigateway.ApiGatewayCallWithoutReq[GetCurAnnResp](n.client, n.config, rest.GET,
		kt, params, "/announcement/get_current_announcements")
	if err != nil {
		return nil, err
	}
	return *resp, nil
}

// RegApp register application
func (n *notice) RegApp(kt *kit.Kit) (*RegAppData, error) {
	return apigateway.ApiGatewayCallWithoutReq[RegAppData](n.client, n.config, rest.POST,
		kt, nil, "/register")
}

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

package azure

import (
	"strings"

	"hcm/pkg/adaptor/types"
)

// NewAzure new azure.
func NewAzure(credential *types.AzureCredential) (*Azure, error) {
	if err := credential.Validate(); err != nil {
		return nil, err
	}
	return &Azure{clientSet: newClientSet(credential)}, nil
}

// Azure is azure operator.
type Azure struct {
	clientSet *clientSet
}

// parseIDToName parse resource id to name, because id is used for identifier but name is used in api.
// id format: /subscriptions/{subscription}/resourceGroups/{resourceGroup}/providers/{provider}/.../{name}.
// for example, 'test' vpc id is: /subscriptions/ss/resourceGroups/rr/providers/Microsoft.Network/virtualNetworks/test.
func parseIDToName(id string) string {
	idx := strings.LastIndex(id, "/")
	return id[idx+1:]
}

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

package gcp

import (
	"fmt"
	"strings"

	"hcm/pkg/adaptor/types"
)

// NewGcp new gcp.
func NewGcp(credential *types.GcpCredential) (*Gcp, error) {
	if err := credential.Validate(); err != nil {
		return nil, err
	}
	return &Gcp{clientSet: newClientSet(credential)}, nil
}

// Gcp is hcp operator.
type Gcp struct {
	clientSet *clientSet
}

// CloudProjectID return cloud project id.
func (g *Gcp) CloudProjectID() string {
	return g.clientSet.credential.CloudProjectID
}

// generateResourceIDsFilter generate gcp resource ids filter
func generateResourceIDsFilter(resourceIDs []string) string {
	filterExp := ""
	for idx, id := range resourceIDs {
		filterExp += "id=" + id
		if idx != len(resourceIDs)-1 {
			filterExp += " OR "
		}
	}

	return filterExp
}

// generateResourceFilter generate gcp resource ids filter
func generateResourceFilter(field string, values []string) string {
	filterExp := ""
	for idx, one := range values {
		filterExp += fmt.Sprintf(`%s="%s"`, field, one)
		if idx != len(values)-1 {
			filterExp += " OR "
		}
	}

	return filterExp
}

// parseSelfLinkToName parse resource self link to name, because self link is used for relation but name is used in api.
// self link format: https://www.googleapis.com/.../{name}.
// for example, 'us-west1' region's self link is: https://www.googleapis.com/compute/v1/projects/xxx/regions/us-west1.
func parseSelfLinkToName(link string) string {
	idx := strings.LastIndex(link, "/")
	return link[idx+1:]
}

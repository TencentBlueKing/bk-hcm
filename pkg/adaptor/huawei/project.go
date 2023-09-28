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

package huawei

import (
	"errors"
	"fmt"

	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
)

// GetProjectID reference: https://support.huaweicloud.com/intl/zh-cn/api-iam/iam_06_0001.html
func (h *HuaWei) GetProjectID(kt *kit.Kit, name string) (string, error) {

	if len(name) == 0 {
		return "", errors.New("name is required")
	}

	client, err := h.clientSet.iamClient(region.AP_SOUTHEAST_1)
	if err != nil {
		return "", fmt.Errorf("new iam client failed, err: %v", err)
	}

	req := &model.KeystoneListProjectsRequest{
		Name: converter.ValToPtr(name),
	}
	resp, err := client.KeystoneListProjects(req)
	if err != nil {
		logs.Errorf("keystone list project failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	projects := converter.PtrToVal(resp.Projects)
	if len(projects) == 0 {
		return "", fmt.Errorf("region: %s not found", name)
	}

	return projects[0].Id, nil
}

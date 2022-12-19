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
	"hcm/pkg/adaptor/types"
	"hcm/pkg/kit"
)

// AccountCheck check account authentication information and permissions.
func (g *Gcp) AccountCheck(kt *kit.Kit, secret *types.GcpCredential) error {
	// 通过调用获取项目信息接口来验证账号有效性(账号需要有 compute.projects.get 权限)
	if _, err := g.GetProject(kt, secret); err != nil {
		return err
	}

	return nil
}

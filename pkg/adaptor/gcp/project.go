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
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"google.golang.org/api/compute/v1"
)

// getProject 获取项目信息(账号需要有 compute.projects.get 权限)
// 接口参考 https://cloud.google.com/compute/docs/reference/rest/v1/projects/get
func (g *Gcp) getProject(kt *kit.Kit) (*compute.Project, error) {
	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		logs.Errorf("init gcp client failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	cloudProjectID := g.clientSet.credential.CloudProjectID
	project, err := client.Projects.Get(cloudProjectID).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("get project %s failed, err: %v, rid: %s", cloudProjectID, err, kt.Rid)
		return nil, err
	}
	return project, nil
}

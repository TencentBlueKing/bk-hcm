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

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type clientSet struct {
	credential *types.GcpCredential
}

func newClientSet(credential *types.GcpCredential) *clientSet {
	return &clientSet{credential}
}

func (c *clientSet) computeClient(kt *kit.Kit) (*compute.Service, error) {
	opt := option.WithCredentialsJSON(c.credential.Json)
	service, err := compute.NewService(kt.Ctx, opt)
	if err != nil {
		return nil, err
	}

	return service, nil
}

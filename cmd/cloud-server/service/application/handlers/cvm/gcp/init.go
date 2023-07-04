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
	"hcm/cmd/cloud-server/service/application/handlers"
	"hcm/cmd/cloud-server/service/common"
	proto "hcm/pkg/api/cloud-server/cvm"
	hcproto "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
)

// ApplicationOfCreateGcpCvm ...
type ApplicationOfCreateGcpCvm struct {
	handlers.BaseApplicationHandler

	req *proto.GcpCvmCreateReq
}

// NewApplicationOfCreateGcpCvm ...
func NewApplicationOfCreateGcpCvm(
	opt *handlers.HandlerOption, req *proto.GcpCvmCreateReq,
) *ApplicationOfCreateGcpCvm {
	return &ApplicationOfCreateGcpCvm{
		BaseApplicationHandler: handlers.NewBaseApplicationHandler(opt, enumor.CreateCvm, enumor.Gcp),
		req:                    req,
	}
}

func (a *ApplicationOfCreateGcpCvm) toHcProtoGcpBatchCreateReq() *hcproto.GcpBatchCreateReq {
	return common.ConvGcpCvmCreateReq(a.req)
}

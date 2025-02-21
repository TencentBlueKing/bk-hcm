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

// Package logics ...
package logics

import (
	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/logics/cvm"
	"hcm/cmd/cloud-server/logics/disk"
	"hcm/cmd/cloud-server/logics/eip"
	securitygroup "hcm/cmd/cloud-server/logics/security-group"
	"hcm/pkg/client"
	"hcm/pkg/thirdparty/esb"
)

// Logics defines cloud-server common logics.
type Logics struct {
	Audit         audit.Interface
	Disk          disk.Interface
	Cvm           cvm.Interface
	Eip           eip.Interface
	SecurityGroup securitygroup.Interface
}

// NewLogics create a new cloud server logics.
func NewLogics(c *client.ClientSet, esbClient esb.Client) *Logics {
	auditLogics := audit.NewAudit(c.DataService())
	eipLogics := eip.NewEip(c, auditLogics)
	diskLogics := disk.NewDisk(c, auditLogics)
	return &Logics{
		Audit:         auditLogics,
		Disk:          disk.NewDisk(c, auditLogics),
		Cvm:           cvm.NewCvm(c, auditLogics, eipLogics, diskLogics, esbClient),
		Eip:           eip.NewEip(c, auditLogics),
		SecurityGroup: securitygroup.NewSecurityGroup(c, auditLogics),
	}
}

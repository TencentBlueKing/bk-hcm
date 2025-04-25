/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

// Package global ...
package global

import (
	"hcm/pkg/api/core"
	dataorgtopo "hcm/pkg/api/data-service/org_topo"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// OrgTopoClient is data service org topo api client.
type OrgTopoClient struct {
	client rest.ClientInterface
}

// NewOrgTopoClient create a new org topo api client.
func NewOrgTopoClient(client rest.ClientInterface) *OrgTopoClient {
	return &OrgTopoClient{
		client: client,
	}
}

// List ...
func (g *OrgTopoClient) List(kt *kit.Kit, req *dataorgtopo.ListReq) (*dataorgtopo.ListResp, error) {
	return common.Request[dataorgtopo.ListReq, dataorgtopo.ListResp](
		g.client, rest.POST, kt, req, "/org_topos/list")
}

// ListByDeptIDs ...
func (g *OrgTopoClient) ListByDeptIDs(kt *kit.Kit, req *dataorgtopo.ListByDeptIDsReq) (*dataorgtopo.ListResp, error) {
	return common.Request[dataorgtopo.ListByDeptIDsReq, dataorgtopo.ListResp](
		g.client, rest.POST, kt, req, "/org_topos/list/by/dept_ids")
}

// BatchCreate ...
func (g *OrgTopoClient) BatchCreate(kt *kit.Kit, req *dataorgtopo.BatchCreateOrgTopoReq) (
	*core.BatchCreateResult, error) {

	return common.Request[dataorgtopo.BatchCreateOrgTopoReq, core.BatchCreateResult](
		g.client, rest.POST, kt, req, "/org_topos/batch/create")
}

// BatchUpdate ...
func (g *OrgTopoClient) BatchUpdate(kt *kit.Kit, req *dataorgtopo.BatchUpdateOrgTopoReq) error {
	return common.RequestNoResp[dataorgtopo.BatchUpdateOrgTopoReq](
		g.client, rest.PATCH, kt, req, "/org_topos/batch")
}

// BatchDelete ...
func (g *OrgTopoClient) BatchDelete(kt *kit.Kit, req *dataorgtopo.BatchDeleteOrgTopoReq) error {
	return common.RequestNoResp[dataorgtopo.BatchDeleteOrgTopoReq](
		g.client, rest.DELETE, kt, req, "/org_topos/batch")
}

// BatchUpsert ...
func (g *OrgTopoClient) BatchUpsert(kt *kit.Kit, req *dataorgtopo.BatchUpsertOrgTopoReq) (
	*core.BatchCreateResult, error) {

	return common.Request[dataorgtopo.BatchUpsertOrgTopoReq, core.BatchCreateResult](
		g.client, rest.POST, kt, req, "/org_topos/batch/upsert")
}

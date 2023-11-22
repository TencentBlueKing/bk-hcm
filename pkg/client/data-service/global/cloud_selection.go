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

package global

import (
	"hcm/pkg/api/core"
	coreselection "hcm/pkg/api/core/cloud-selection"
	dsselection "hcm/pkg/api/data-service/cloud-selection"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewCloudCloudSelectionClient create a new cloud selection api client.
func NewCloudCloudSelectionClient(client rest.ClientInterface) *CloudSelectionClient {
	return &CloudSelectionClient{
		client: client,
	}
}

// CloudSelectionClient is data service cloud selection api client.
type CloudSelectionClient struct {
	client rest.ClientInterface
}

// ListScheme list scheme.
func (cli *CloudSelectionClient) ListScheme(kt *kit.Kit, req *core.ListReq) (
	*core.ListResultT[coreselection.Scheme], error) {

	return common.Request[core.ListReq, core.ListResultT[coreselection.Scheme]](cli.client, rest.POST, kt, req,
		"/clouds/selections/schemes/list")
}

// CreateScheme create scheme.
func (cli *CloudSelectionClient) CreateScheme(kt *kit.Kit, req *dsselection.SchemeCreateReq) (
	*core.CreateResult, error) {

	return common.Request[dsselection.SchemeCreateReq, core.CreateResult](cli.client, rest.POST, kt, req,
		"/clouds/selections/schemes/create")
}

// BatchDeleteScheme batch delete scheme.
func (cli *CloudSelectionClient) BatchDeleteScheme(kt *kit.Kit, req *core.BatchDeleteReq) error {

	return common.RequestNoResp[core.BatchDeleteReq](cli.client, rest.DELETE, kt, req,
		"/clouds/selections/schemes/batch")
}

// UpdateScheme update scheme.
func (cli *CloudSelectionClient) UpdateScheme(kt *kit.Kit, id string, req *dsselection.SchemeUpdateReq) error {

	return common.RequestNoResp[dsselection.SchemeUpdateReq](cli.client, rest.PATCH, kt, req,
		"/clouds/selections/schemes/%s", id)
}

// ListIdc list idc.
func (cli *CloudSelectionClient) ListIdc(kt *kit.Kit, req *core.ListReq) (
	*core.ListResultT[coreselection.Idc], error) {

	return common.Request[core.ListReq, core.ListResultT[coreselection.Idc]](cli.client, rest.POST, kt, req,
		"/clouds/selections/idcs/list")
}

// ListBizType list biz_type.
func (cli *CloudSelectionClient) ListBizType(kt *kit.Kit, req *core.ListReq) (
	*core.ListResultT[coreselection.BizType], error) {

	return common.Request[core.ListReq, core.ListResultT[coreselection.BizType]](cli.client, rest.POST, kt, req,
		"/clouds/selections/biz_types/list")
}

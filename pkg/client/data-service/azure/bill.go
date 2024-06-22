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

package azure

import (
	"hcm/pkg/api/core"
	billproto "hcm/pkg/api/data-service/bill"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewBillClient create a new bill api client.
func NewBillClient(client rest.ClientInterface) *BillClient {
	return &BillClient{
		client: client,
	}
}

// BillClient is data service bill api client.
type BillClient struct {
	client rest.ClientInterface
}

// ListBillItem list bill item
func (b *BillClient) ListBillItem(kt *kit.Kit, req *core.ListReq) (*billproto.AzureBillItemListResult, error) {

	return common.Request[core.ListReq, billproto.AzureBillItemListResult](
		b.client, rest.POST, kt, req, "/bills/items/list")
}

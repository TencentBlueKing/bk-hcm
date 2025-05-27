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

package other

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
)

// Interface support resource sync.
type Interface interface {
	RemoveHostByCCInfo(kt *kit.Kit, params *DelHostParams) error
	Host(kt *kit.Kit, params *SyncHostParams) error
}

var _ Interface = new(client)

type client struct {
	accountID     string
	dbCli         *dataservice.Client
	ccHostPoolBiz int64
}

// NewClient new client.
func NewClient(dbCli *dataservice.Client, accountID string) Interface {
	return &client{
		dbCli:     dbCli,
		accountID: accountID,
		// todo ccHostPoolBiz后续使用cc提供的api获取
		ccHostPoolBiz: cc.HCService().CCHostPoolBiz,
	}
}

// DelHostParams ...
type DelHostParams struct {
	BizID             int64              `json:"bk_biz_id"`
	CCBizExistHostIDs map[int64]struct{} `json:"cc_exist_host_ids"`
	DelHostIDs        []int64            `json:"delete_host_ids"`
}

// Validate ...
func (opt DelHostParams) Validate() error {
	if len(opt.DelHostIDs) > constant.CloudResourceSyncMaxLimit {
		return fmt.Errorf("host ids should <= %d", constant.CloudResourceSyncMaxLimit)
	}

	return nil
}

// SyncHostParams ...
type SyncHostParams struct {
	AccountID string                         `json:"account_id" validate:"required"`
	BizID     int64                          `json:"bk_biz_id" validate:"required"`
	HostIDs   []int64                        `json:"bk_host_ids" validate:"required"`
	HostCache map[int64]cmdb.HostWithCloudID `json:"host_cache"`
}

// Validate ...
func (opt SyncHostParams) Validate() error {
	if len(opt.HostIDs) > int(core.DefaultMaxPageLimit) {
		return fmt.Errorf("host ids should <= %d", int(core.DefaultMaxPageLimit))
	}

	return validator.Validate.Struct(opt)
}

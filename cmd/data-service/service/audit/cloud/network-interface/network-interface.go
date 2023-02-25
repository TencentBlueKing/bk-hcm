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

package networkinterface

import (
	"hcm/pkg/api/core"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableni "hcm/pkg/dal/table/cloud/network-interface"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// ListNetworkInterface list network interface.
func ListNetworkInterface(kt *kit.Kit, dao dao.Set, ids []string) (map[string]tableni.NetworkInterfaceTable, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.DefaultBasePage,
	}
	list, err := dao.NetworkInterface().List(kt, opt)
	if err != nil {
		logs.Errorf("list network interface failed, err: %v, ids: %v, rid: %ad", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tableni.NetworkInterfaceTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}

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

package actcli

import (
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/dal/dao"
)

var (
	cliSet *client.ClientSet
	daoSet dao.Set
)

// SetClientSet set client set.
func SetClientSet(cli *client.ClientSet) {
	cliSet = cli
}

// GetHCService get hc service.
func GetHCService() *hcservice.Client {
	return cliSet.HCService()
}

// GetDataService get data service.
func GetDataService() *dataservice.Client {
	return cliSet.DataService()
}

// SetDaoSet set dao set.
func SetDaoSet(cli dao.Set) {
	daoSet = cli
}

// GetDaoSet get dao set.
func GetDaoSet() dao.Set {
	return daoSet
}

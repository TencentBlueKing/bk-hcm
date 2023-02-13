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

package routetable

import (
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
)

// Route defines route dao operations.
type Route interface {
	TCloud() TCloudRoute
	Aws() AwsRoute
	Azure() AzureRoute
	HuaWei() HuaWeiRoute
	Gcp() GcpRoute
}

var _ Route = new(routeDao)

// routeDao route dao.
type routeDao struct {
	tcloud TCloudRoute
	aws    AwsRoute
	azure  AzureRoute
	huawei HuaWeiRoute
	gcp    GcpRoute
}

// NewRouteDao create a route dao.
func NewRouteDao(orm orm.Interface, idGen idgenerator.IDGenInterface, audit audit.Interface) Route {
	return &routeDao{
		tcloud: NewTCloudRouteDao(orm, idGen, audit),
		aws:    NewAwsRouteDao(orm, idGen, audit),
		azure:  NewAzureRouteDao(orm, idGen, audit),
		huawei: NewHuaWeiRouteDao(orm, idGen, audit),
		gcp:    NewGcpRouteDao(orm, idGen, audit),
	}
}

func (r *routeDao) TCloud() TCloudRoute {
	return r.tcloud
}

func (r *routeDao) Aws() AwsRoute {
	return r.aws
}

func (r *routeDao) Azure() AzureRoute {
	return r.azure
}

func (r *routeDao) HuaWei() HuaWeiRoute {
	return r.huawei
}

func (r *routeDao) Gcp() GcpRoute {
	return r.gcp
}

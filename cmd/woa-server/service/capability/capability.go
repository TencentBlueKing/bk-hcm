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

// Package capability ...
package capability

import (
	"hcm/cmd/woa-server/logics/biz"
	"hcm/cmd/woa-server/logics/config"
	cvmlogic "hcm/cmd/woa-server/logics/cvm"
	"hcm/cmd/woa-server/logics/dissolve"
	gclogics "hcm/cmd/woa-server/logics/green-channel"
	"hcm/cmd/woa-server/logics/plan"
	ressync "hcm/cmd/woa-server/logics/res-sync"
	rslogic "hcm/cmd/woa-server/logics/rolling-server"
	taskLogics "hcm/cmd/woa-server/logics/task"
	"hcm/cmd/woa-server/logics/task/informer"
	"hcm/cmd/woa-server/logics/task/operation"
	"hcm/cmd/woa-server/logics/task/recycler"
	"hcm/cmd/woa-server/logics/task/scheduler"
	taskStatistics "hcm/cmd/woa-server/logics/task/statistics"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/dal/dao"
	"hcm/pkg/iam/auth"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/es"

	"github.com/emicklei/go-restful/v3"
)

// Capability defines the service's capability
type Capability struct {
	Client         *client.ClientSet
	Dao            dao.Set
	WebService     *restful.WebService
	PlanController plan.Logics
	CmdbCli        cmdb.Client
	ThirdCli       *thirdparty.Client
	Authorizer     auth.Authorizer
	Conf           cc.WoaServerSetting
	SchedulerIf    scheduler.Interface
	InformerIf     informer.Interface
	RecyclerIf     recycler.Interface
	OperationIf    operation.Interface
	EsCli          *es.EsCli
	RsLogic        rslogic.Logics
	GcLogic        gclogics.Logics
	BizLogic       biz.Logics
	DissolveLogic  dissolve.Logics
	ResSyncLogic   ressync.Logics
	ConfigLogics   config.Logics
	TaskLogic      taskLogics.Logics
	TaskStatistics taskStatistics.Interface
	CvmLogic       cvmlogic.Logics
}

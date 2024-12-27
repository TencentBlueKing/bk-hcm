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

// Package cvm ...
package cvm

import (
	"hcm/cmd/hc-service/logics/cloud-adaptor"
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// InitCvmService initial the cvm service.
func InitCvmService(cap *capability.Capability) {
	svc := &cvmSvc{
		ad:      cap.CloudAdaptor,
		dataCli: cap.ClientSet.DataService(),
	}

	svc.initTCloudCvmService(cap)
	svc.initAwsCvmService(cap)
	svc.initAzureCvmService(cap)
	svc.initGcpCvmService(cap)
	svc.initHuaWeiCvmService(cap)
}

type cvmSvc struct {
	ad      *cloudadaptor.CloudAdaptorClient
	dataCli *dataservice.Client
	client  *client.ClientSet
}

func (svc *cvmSvc) getCvms(kt *kit.Kit, vendor enumor.Vendor, region string, cvmIDs []string) ([]corecvm.BaseCvm,
	error) {
	if len(cvmIDs) == 0 {
		return nil, nil
	}

	result := make([]corecvm.BaseCvm, 0, len(cvmIDs))
	for _, ids := range slice.Split(cvmIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("vendor", vendor),
				tools.RuleEqual("region", region),
				tools.RuleIn("id", ids),
			),
			Page: core.NewDefaultBasePage(),
		}
		listResp, err := svc.dataCli.Global.Cvm.ListCvm(kt, listReq)
		if err != nil {
			logs.Errorf("request dataservice list cvm failed, err: %v, ids: %v, rid: %s", err, cvmIDs, kt.Rid)
			return nil, err
		}
		result = append(result, listResp.Details...)
	}

	return result, nil
}

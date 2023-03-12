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

package huawei

import (
	"time"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcapiproto "hcm/pkg/api/hc-service"
	dataservice "hcm/pkg/client/data-service"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// SyncNetworkInterface 网络接口同步
func SyncNetworkInterface(kt *kit.Kit, hcCli *hcservice.Client, dataCli *dataservice.Client, accountID string,
	regions []string) error {

	start := time.Now()
	logs.V(3).Infof("cloud-server-sync-%s account[%s] sync network interface start, time: %v, rid: %s",
		enumor.HuaWei, accountID, start, kt.Rid)

	defer func() {
		logs.V(3).Infof("cloud-server-sync-%s account[%s] sync network interface end, cost: %v, rid: %s",
			enumor.HuaWei, accountID, time.Since(start), kt.Rid)
	}()

	for _, region := range regions {
		cvmCloudIDs, err := getCvmListAll(kt, dataCli, accountID, region)
		if err != nil {
			continue
		}

		req := &hcapiproto.HuaWeiNetworkInterfaceSyncReq{
			AccountID:   accountID,
			Region:      region,
			CloudCvmIDs: cvmCloudIDs,
		}
		if err = hcCli.HuaWei.NetworkInterface.SyncNetworkInterface(kt.Ctx, kt.Header(), req); err != nil {
			logs.Errorf("cloud-server-sync-%s network interface failed, req: %v, err: %v, rid: %s",
				enumor.HuaWei, req, err, kt.Rid)
			return err
		}
	}

	return nil
}

func getCvmListAll(kt *kit.Kit, dataCli *dataservice.Client, accountID, region string) ([]string, error) {
	listReq := &dataproto.CvmListReq{
		Field: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: accountID,
				},
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: enumor.HuaWei,
				},
				&filter.AtomRule{
					Field: "region",
					Op:    filter.Equal.Factory(),
					Value: region,
				},
			},
		},
		Page: core.DefaultBasePage,
	}
	listResp, err := dataCli.Global.Cvm.ListCvm(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("cloud-server-sync-%s, request dataservice list cvm failed, ids: %v, err: %v, rid: %s",
			enumor.HuaWei, err, kt.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	return cloudIDs, nil
}

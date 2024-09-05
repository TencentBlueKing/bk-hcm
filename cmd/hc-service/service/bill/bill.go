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

// Package bill defines bill service.
package bill

import (
	"fmt"

	cloudadaptor "hcm/cmd/hc-service/logics/cloud-adaptor"
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/client"
	dataserviceclient "hcm/pkg/client/data-service"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// InitBillService initial the bill service
func InitBillService(cap *capability.Capability) {
	v := &bill{
		ad: cap.CloudAdaptor,
		cs: cap.ClientSet,
	}

	h := rest.NewHandler()

	h.Add("AwsGetBillList", "POST", "/vendors/aws/bills/list", v.AwsGetBillList)
	h.Add("AwsBillsPipeline", "POST", "/vendors/aws/bills/pipeline", v.AwsBillPipeline)
	h.Add("AwsBillConfigDelete", "DELETE", "/vendors/aws/bills/{id}", v.AwsBillConfigDelete)
	h.Add("TCloudGetBillList", "POST", "/vendors/tcloud/bills/list", v.TCloudGetBillList)
	h.Add("HuaWeiGetBillList", "POST", "/vendors/huawei/bills/list", v.HuaWeiGetBillList)
	h.Add("HuaWeiGetFeeRecordList", "POST", "/vendors/huawei/feerecords/list", v.HuaWeiGetFeeRecordList)
	h.Add("AzureGetBillList", "POST", "/vendors/azure/bills/list", v.AzureGetBillList)
	h.Add("GcpGetBillList", "POST", "/vendors/gcp/bills/list", v.GcpGetBillList)
	h.Add("GcpGetRootAccountBillList", "POST", "/vendors/gcp/root_account_bills/list", v.GcpGetRootAccountBillList)
	h.Add("AwsGetRootAccountBillList", "POST", "/vendors/aws/root_account_bills/list", v.AwsGetRootAccountBillList)
	h.Add("AzureGetRootAccountBillList", "POST",
		"/vendors/azure/root_account_bills/list", v.AzureGetRootAccountBillList)
	h.Add("AwsGetRootAccountSpTotalUsage", "GET",
		"/vendors/aws/root_account_bills/sp_usage_total", v.AwsGetRootAccountSpTotalUsage)

	h.Load(cap.WebService)
}

type bill struct {
	ad *cloudadaptor.CloudAdaptorClient
	cs *client.ClientSet
}

// getBillInfo get bill info.
func getBillInfo[T cloud.AccountBillConfigExtension](kt *kit.Kit, accountID string,
	dataCli *dataserviceclient.Client) (*cloud.AccountBillConfig[T], error) {

	// 查询gcp账单基础表
	billList, err := dataCli.Global.Bill.List(kt.Ctx, kt.Header(), &core.ListReq{
		Filter: tools.EqualExpression("account_id", accountID),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	})
	if err != nil {
		logs.Errorf("aws get base info from db failed, accountID: %s, err: %+v", accountID, err)
		return nil, err
	}
	if len(billList.Details) == 0 {
		return nil, nil
	}

	billData := billList.Details[0]
	extension := new(T)
	if billData.Extension != "" {
		err = json.UnmarshalFromString(string(billData.Extension), extension)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalFromString account bill config hc extension failed, err: %+v", err)
		}
	}

	return &cloud.AccountBillConfig[T]{
		BaseAccountBillConfig: billData,
		Extension:             extension,
	}, nil
}

// getRootAccountBillConfigInfo get root account bill config info.
func getRootAccountBillConfigInfo[T billcore.RootAccountBillConfigExtension](kt *kit.Kit, rootAccountID string,
	dataCli *dataserviceclient.Client) (*billcore.RootAccountBillConfig[T], error) {

	// 查询gcp账单基础表
	billList, err := dataCli.Global.Bill.ListRootAccountBillConfig(kt.Ctx, kt.Header(), &core.ListReq{
		Filter: tools.EqualExpression("root_account_id", rootAccountID),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	})
	if err != nil {
		logs.Errorf("aws get base info from db failed, rootAccountID: %s, err: %+v", rootAccountID, err)
		return nil, err
	}
	if len(billList.Details) == 0 {
		return nil, nil
	}

	billData := billList.Details[0]
	extension := new(T)
	if billData.Extension != "" {
		err = json.UnmarshalFromString(string(billData.Extension), extension)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalFromString root account bill config hc extension failed, err: %+v", err)
		}
	}

	return &billcore.RootAccountBillConfig[T]{
		BaseRootAccountBillConfig: billData,
		Extension:                 extension,
	}, nil
}

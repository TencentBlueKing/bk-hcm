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

package eip

import (
	"hcm/pkg/adaptor/types/eip"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	protoeip "hcm/pkg/api/hc-service/eip"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// TCloudEipSync
type TCloudEipSync struct {
	IsUpdate bool
	Eip      *eip.TCloudEip
}

// TCloudDSEipSync ...
type TCloudDSEipSync struct {
	Eip *dataproto.EipExtResult[dataproto.TCloudEipExtensionResult]
}

// AwsEipSync
type AwsEipSync struct {
	IsUpdate bool
	Eip      *eip.AwsEip
}

// AwsDSEipSync
type AwsDSEipSync struct {
	Eip *dataproto.EipExtResult[dataproto.AwsEipExtensionResult]
}

// HuaWeiEipSync
type HuaWeiEipSync struct {
	IsUpdate bool
	Eip      *eip.HuaWeiEip
}

// HuaWeiDSEipSync
type HuaWeiDSEipSync struct {
	Eip *dataproto.EipExtResult[dataproto.HuaWeiEipExtensionResult]
}

// GcpEipSync
type GcpEipSync struct {
	IsUpdate bool
	Eip      *eip.GcpEip
}

// GcpDSEipSync
type GcpDSEipSync struct {
	Eip *dataproto.EipExtResult[dataproto.GcpEipExtensionResult]
}

// AzureEipSync
type AzureEipSync struct {
	IsUpdate bool
	Eip      *eip.AzureEip
}

// AzureDSEipSync
type AzureDSEipSync struct {
	Eip *dataproto.EipExtResult[dataproto.AzureEipExtensionResult]
}

func (ea *eipAdaptor) decodeEipSyncReq(cts *rest.Contexts) (*protoeip.EipSyncReq, error) {

	req := new(protoeip.EipSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return req, nil
}

func (ea *eipAdaptor) syncEipDelete(cts *rest.Contexts, deleteCloudIDs []string) error {

	batchDeleteReq := &dataproto.EipDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", deleteCloudIDs),
	}

	if _, err := ea.dataCli.Global.DeleteEip(cts.Kit.Ctx, cts.Kit.Header(), batchDeleteReq); err != nil {
		logs.Errorf("request dataservice delete tcloud eip failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return err
	}

	return nil
}

func getStrPtrVal(val *string) string {

	if val == nil {
		return ""
	}

	return *val
}

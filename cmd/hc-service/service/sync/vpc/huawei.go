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

// Package vpc defines vpc service.
package vpc

import (
	vpclogic "hcm/cmd/hc-service/logics/sync/vpc"
	"hcm/pkg/adaptor/huawei"
	"hcm/pkg/adaptor/types"
	typecore "hcm/pkg/adaptor/types/core"
	dataproto "hcm/pkg/api/data-service"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// SyncHuaWeiVpc sync huawei vpc to hcm.
func (v syncVpcSvc) SyncHuaWeiVpc(cts *rest.Contexts) (interface{}, error) {
	req := new(sync.HuaWeiSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := v.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	cloudTotalMap := make(map[string]struct{}, 0)
	listOpt := &types.HuaWeiVpcListOption{
		HuaWeiListOption: typecore.HuaWeiListOption{
			Region:   req.Region,
			CloudIDs: nil,
			Page: &typecore.HuaWeiPage{
				Marker: nil,
				Limit:  converter.ValToPtr(int32(constant.BatchOperationMaxLimit)),
			},
		},
	}
	for {
		vpcResult, err := cli.ListVpc(cts.Kit, listOpt)
		if err != nil {
			logs.Errorf("request adaptor list huawei vpc failed, err: %v, opt: %v, rid: %s", err, listOpt, cts.Kit.Rid)
			return nil, err
		}

		if len(vpcResult.Details) == 0 {
			break
		}

		cloudIDs := make([]string, 0, len(vpcResult.Details))
		for _, one := range vpcResult.Details {
			cloudTotalMap[one.CloudID] = struct{}{}
			cloudIDs = append(cloudIDs, one.CloudID)
		}

		syncOpt := &vpclogic.SyncHuaWeiOption{
			AccountID: req.AccountID,
			Region:    req.Region,
			CloudIDs:  cloudIDs,
		}
		_, err = vpclogic.HuaWeiVpcSync(cts.Kit, syncOpt, v.ad, v.dataCli)
		if err != nil {
			logs.Errorf("request to sync huawei vpc failed, err: %v, opt: %v, rid: %s", err, syncOpt, cts.Kit.Rid)
			return nil, err
		}

		if vpcResult.NextMarker == nil || len(*vpcResult.NextMarker) == 0 {
			break
		}

		listOpt.Page.Marker = vpcResult.NextMarker
	}

	if err = v.delDBNotExistHuaWeiVpc(cts.Kit, cli, req, cloudTotalMap); err != nil {
		logs.Errorf("remove db not exist vpc failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (v *syncVpcSvc) delDBNotExistHuaWeiVpc(kt *kit.Kit, huawei *huawei.HuaWei, req *sync.HuaWeiSyncReq,
	cloudCvmMap map[string]struct{}) error {

	// 找出云上已经不存在的VpcID
	delCloudVpcIDs, err := v.queryDeleteVpc(kt, req.Region, req.AccountID, cloudCvmMap)
	if err != nil {
		return err
	}

	// 再用这部分VpcID去云上确认是否存在，如果不存在，删除db数据，存在的忽略，等同步修复
	start, end := 0, typecore.HuaWeiQueryLimit
	for {
		if start+end > len(delCloudVpcIDs) {
			end = len(delCloudVpcIDs)
		} else {
			end = int(start) + typecore.HuaWeiQueryLimit
		}
		tmpCloudIDs := delCloudVpcIDs[start:end]

		if len(tmpCloudIDs) == 0 {
			break
		}

		listOpt := &types.HuaWeiVpcListOption{
			HuaWeiListOption: typecore.HuaWeiListOption{
				Region:   req.Region,
				CloudIDs: tmpCloudIDs,
			},
		}
		vpcResult, err := huawei.ListVpc(kt, listOpt)
		if err != nil {
			logs.Errorf("list vpc from huawei failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
			return err
		}

		tmpMap := converter.StringSliceToMap(tmpCloudIDs)
		for _, vpc := range vpcResult.Details {
			delete(tmpMap, vpc.CloudID)
		}

		if len(tmpMap) == 0 {
			start = end
			continue
		}

		if err = v.dataCli.Global.Vpc.BatchDelete(kt.Ctx, kt.Header(), &dataproto.BatchDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", converter.MapKeyToStringSlice(tmpMap)),
		}); err != nil {
			logs.Errorf("batch delete db vpc failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		start = end
		if start == len(delCloudVpcIDs) {
			break
		}
	}

	return nil
}

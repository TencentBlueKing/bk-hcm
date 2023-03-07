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

// Package subnet defines subnet service.
package subnet

import (
	"hcm/cmd/hc-service/logics/sync/subnet"
	"hcm/pkg/adaptor/huawei"
	"hcm/pkg/adaptor/types"
	typecore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
)

// SyncHuaWeiSubnet sync huawei subnet to hcm.
func (v *syncSubnetSvc) SyncHuaWeiSubnet(cts *rest.Contexts) (interface{}, error) {
	req := new(sync.HuaWeiSubnetSyncReq)
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
	listOpt := &types.HuaWeiSubnetListOption{
		Region:     req.Region,
		CloudVpcID: req.CloudVpcID,
		Page: &typecore.HuaWeiPage{
			Marker: nil,
			Limit:  converter.ValToPtr(int32(constant.BatchOperationMaxLimit)),
		},
	}
	for {
		subnetResult, err := cli.ListSubnet(cts.Kit, listOpt)
		if err != nil {
			logs.Errorf("request adaptor list huawei subnet failed, err: %v, opt: %v, rid: %s", err,
				listOpt, cts.Kit.Rid)
			return nil, err
		}

		if len(subnetResult.Details) == 0 {
			break
		}

		cloudIDs := make([]string, 0, len(subnetResult.Details))
		for _, one := range subnetResult.Details {
			cloudTotalMap[one.CloudID] = struct{}{}
			cloudIDs = append(cloudIDs, one.CloudID)
		}

		syncOpt := &subnet.SyncHuaWeiOption{
			AccountID:  req.AccountID,
			Region:     req.Region,
			CloudVpcID: req.CloudVpcID,
			CloudIDs:   cloudIDs,
		}
		_, err = subnet.SyncHuaWeiSubnet(cts.Kit, syncOpt, v.ad, v.dataCli)
		if err != nil {
			logs.Errorf("request to sync huawei subnet failed, err: %v, opt: %v, rid: %s", err, syncOpt, cts.Kit.Rid)
			return nil, err
		}

		if len(subnetResult.Details) < constant.BatchOperationMaxLimit {
			break
		}

		listOpt.Page.Marker = converter.ValToPtr(subnetResult.Details[len(subnetResult.Details)-1].CloudID)
	}

	if err = v.delDBNotExistHuaWeiSubnet(cts.Kit, cli, req, cloudTotalMap); err != nil {
		logs.Errorf("remove db not exist subnet failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (v *syncSubnetSvc) delDBNotExistHuaWeiSubnet(kt *kit.Kit, huawei *huawei.HuaWei, req *sync.HuaWeiSubnetSyncReq,
	cloudCvmMap map[string]struct{}) error {

	// 找出云上已经不存在的SubnetID
	start := uint32(0)
	delCloudIDs := make([]string, 0)
	for {
		listReq := &core.ListReq{
			Fields: []string{"cloud_id"},
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: req.AccountID},
					&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: req.Region},
					&filter.AtomRule{Field: "cloud_vpc_id", Op: filter.Equal.Factory(), Value: req.CloudVpcID},
				},
			},
			Page: &core.BasePage{
				Start: start,
				Limit: core.DefaultMaxPageLimit,
			},
		}
		result, err := v.dataCli.Global.Subnet.List(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("list subnet from db failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		for _, one := range result.Details {
			if _, exist := cloudCvmMap[one.CloudID]; !exist {
				delCloudIDs = append(delCloudIDs, one.CloudID)
			}
		}

		if len(result.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		start += uint32(core.DefaultMaxPageLimit)
	}

	// 再用这部分SubnetID去云上确认是否存在，如果不存在，删除db数据，存在的忽略，等同步修复
	start, end := 0, typecore.HuaWeiQueryLimit
	for {
		if int(start)+end > len(delCloudIDs) {
			end = len(delCloudIDs)
		} else {
			end = int(start) + typecore.HuaWeiQueryLimit
		}
		tmpCloudIDs := delCloudIDs[start:end]

		if len(tmpCloudIDs) == 0 {
			break
		}

		listOpt := &types.HuaWeiSubnetListOption{
			Region:     req.Region,
			CloudVpcID: req.CloudVpcID,
		}
		subnetResult, err := huawei.ListSubnet(kt, listOpt)
		if err != nil {
			logs.Errorf("list subnet from huawei failed, err: %v, opt: %v, rid: %s", err, listOpt, kt.Rid)
			return err
		}

		tmpMap := converter.StringSliceToMap(tmpCloudIDs)
		for _, one := range subnetResult.Details {
			delete(tmpMap, one.CloudID)
		}

		if len(tmpMap) == 0 {
			start = uint32(end)
			continue
		}

		if err = v.dataCli.Global.Subnet.BatchDelete(kt.Ctx, kt.Header(), &dataproto.BatchDeleteReq{
			Filter: tools.ContainersExpression("cloud_id", converter.MapKeyToStringSlice(tmpMap)),
		}); err != nil {
			logs.Errorf("batch delete db subnet failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		start = uint32(end)
		if int(start) == len(delCloudIDs) {
			break
		}
	}

	return nil
}

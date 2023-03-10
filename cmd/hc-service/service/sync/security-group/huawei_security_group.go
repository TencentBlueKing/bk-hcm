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

package securitygroup

import (
	securitygroup "hcm/cmd/hc-service/logics/sync/security-group"
	"hcm/pkg/adaptor/huawei"
	typcore "hcm/pkg/adaptor/types/core"
	typessg "hcm/pkg/adaptor/types/security-group"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SyncHuaWeiSecurityGroup sync huawei security group to hcm.
func (svc *syncSecurityGroupSvc) SyncHuaWeiSecurityGroup(cts *rest.Contexts) (interface{}, error) {

	req := new(securitygroup.SyncHuaWeiSecurityGroupOption)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.adaptor.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	allCloudIDs := make(map[string]struct{})
	limit := int32(typcore.HuaWeiQueryLimit)
	var marker *string = nil
	for {
		opt := &typessg.HuaWeiListOption{
			Region: req.Region,
			Page:   &typcore.HuaWeiPage{Limit: &limit, Marker: marker},
		}

		datas, pageInfo, err := client.ListSecurityGroup(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list huawei security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		cloudIDs := make([]string, 0, len(*datas))
		for _, one := range *datas {
			cloudIDs = append(cloudIDs, one.Id)
			allCloudIDs[one.Id] = struct{}{}
		}

		if len(cloudIDs) > 0 {
			req.CloudIDs = cloudIDs
		}
		_, err = securitygroup.SyncHuaWeiSecurityGroup(cts.Kit, req, svc.adaptor, svc.dataCli)
		if err != nil {
			logs.Errorf("request to sync huawei security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		marker = pageInfo.NextMarker
		if len(*datas) == 0 || pageInfo.NextMarker == nil {
			break
		}
	}

	commonReq := &hcservice.SecurityGroupSyncReq{
		AccountID: req.AccountID,
		Region:    req.Region,
	}
	dsIDs, err := securitygroup.GetDatasFromDSForSecurityGroupSync(cts.Kit, commonReq, svc.dataCli)
	if err != nil {
		return nil, err
	}

	deleteIDs := make([]string, 0)
	for id := range dsIDs {
		if _, ok := allCloudIDs[id]; !ok {
			deleteIDs = append(deleteIDs, id)
		}
	}

	err = svc.deleteHuaWeiSG(cts, client, req, deleteIDs)
	if err != nil {
		logs.Errorf("request deleteHuaWeiSG failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *syncSecurityGroupSvc) deleteHuaWeiSG(cts *rest.Contexts, client *huawei.HuaWei,
	req *securitygroup.SyncHuaWeiSecurityGroupOption, deleteIDs []string) error {

	if len(deleteIDs) > 0 {
		realDeleteIDs := make([]string, 0)
		limit := int32(typcore.HuaWeiQueryLimit)
		var marker *string = nil
		for {
			opt := &typessg.HuaWeiListOption{
				Region: req.Region,
				Page:   &typcore.HuaWeiPage{Limit: &limit, Marker: marker},
			}

			datas, pageInfo, err := client.ListSecurityGroup(cts.Kit, opt)
			if err != nil {
				logs.Errorf("request adaptor to list huawei security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return err
			}

			for _, id := range deleteIDs {
				realDeleteFlag := true
				for _, data := range *datas {
					if data.Id == id {
						realDeleteFlag = false
						break
					}
				}

				if realDeleteFlag {
					realDeleteIDs = append(realDeleteIDs, id)
				}
			}

			marker = pageInfo.NextMarker
			if len(*datas) == 0 || pageInfo.NextMarker == nil {
				break
			}
		}

		if len(realDeleteIDs) > 0 {
			err := securitygroup.DiffSecurityGroupSyncDelete(cts.Kit, realDeleteIDs, svc.dataCli)
			if err != nil {
				logs.Errorf("sync delete huawei security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return err
			}
		}
	}

	return nil
}

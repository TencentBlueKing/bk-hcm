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
	"hcm/pkg/adaptor/tcloud"
	typcore "hcm/pkg/adaptor/types/core"
	typessg "hcm/pkg/adaptor/types/security-group"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SyncTCloudSecurityGroup sync tcloud security group to hcm.
func (svc *syncSecurityGroupSvc) SyncTCloudSecurityGroup(cts *rest.Contexts) (interface{}, error) {

	req := new(securitygroup.SyncTCloudSecurityGroupOption)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.adaptor.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	allCloudIDs := make(map[string]struct{})
	offset := 0
	for {
		opt := &typessg.TCloudListOption{
			Region: req.Region,
			Page:   &typcore.TCloudPage{Offset: uint64(offset), Limit: uint64(typcore.TCloudQueryLimit)},
		}

		datas, err := client.ListSecurityGroup(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list tcloud security group failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
			return nil, err
		}

		cloudIDs := make([]string, 0, len(datas))
		for _, one := range datas {
			cloudIDs = append(cloudIDs, *one.SecurityGroupId)
			allCloudIDs[*one.SecurityGroupId] = struct{}{}
		}

		if len(cloudIDs) > 0 {
			req.CloudIDs = cloudIDs
		}
		_, err = securitygroup.SyncTCloudSecurityGroup(cts.Kit, req, svc.adaptor, svc.dataCli)
		if err != nil {
			logs.Errorf("request to sync tcloud security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		offset += len(datas)
		if uint(len(datas)) < typcore.TCloudQueryLimit {
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

	err = svc.deleteTCloudSG(cts, client, req, deleteIDs)
	if err != nil {
		logs.Errorf("request deleteTCloudSG failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *syncSecurityGroupSvc) deleteTCloudSG(cts *rest.Contexts, client *tcloud.TCloud,
	req *securitygroup.SyncTCloudSecurityGroupOption, deleteIDs []string) error {

	if len(deleteIDs) > 0 {
		realDeleteIDs := make([]string, 0)
		offset := 0
		for {
			opt := &typessg.TCloudListOption{
				Region: req.Region,
				Page:   &typcore.TCloudPage{Offset: uint64(offset), Limit: uint64(typcore.TCloudQueryLimit)},
			}

			datas, err := client.ListSecurityGroup(cts.Kit, opt)
			if err != nil {
				logs.Errorf("request adaptor to list tcloud security group failed, err: %v, opt: %v, rid: %s",
					err, opt, cts.Kit.Rid)
				return err
			}

			for _, id := range deleteIDs {
				realDeleteFlag := true
				for _, data := range datas {
					if *data.SecurityGroupId == id {
						realDeleteFlag = false
						break
					}
				}

				if realDeleteFlag {
					realDeleteIDs = append(realDeleteIDs, id)
				}
			}

			offset += len(datas)
			if uint(len(datas)) < typcore.TCloudQueryLimit {
				break
			}
		}

		if len(realDeleteIDs) > 0 {
			err := securitygroup.DiffSecurityGroupSyncDelete(cts.Kit, realDeleteIDs, svc.dataCli)
			if err != nil {
				logs.Errorf("sync delete tcloud security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return err
			}
		}
	}

	return nil
}

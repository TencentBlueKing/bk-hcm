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
	"hcm/pkg/adaptor/aws"
	typcore "hcm/pkg/adaptor/types/core"
	typessg "hcm/pkg/adaptor/types/security-group"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
)

// SyncAwsSecurityGroup sync aws security group to hcm.
func (svc *syncSecurityGroupSvc) SyncAwsSecurityGroup(cts *rest.Contexts) (interface{}, error) {

	syncReq := new(sync.SyncAwsSecurityGroupReq)
	if err := cts.DecodeInto(syncReq); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := syncReq.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &securitygroup.SyncAwsSecurityGroupOption{
		AccountID: syncReq.AccountID,
		Region:    syncReq.Region,
	}
	client, err := svc.adaptor.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	allCloudIDs := make(map[string]struct{})
	nextToken := ""
	for {
		listOpt := &typessg.AwsListOption{
			Region: req.Region,
			Page: &typcore.AwsPage{
				MaxResults: converter.ValToPtr(int64(filter.DefaultMaxInLimit)),
			},
		}
		if nextToken != "" {
			listOpt.Page.NextToken = converter.ValToPtr(nextToken)
		}

		results, err := client.ListSecurityGroup(cts.Kit, listOpt)
		if err != nil {
			logs.Errorf("request adaptor to list aws security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		cloudIDs := make([]string, 0, len(results.SecurityGroups))
		for _, one := range results.SecurityGroups {
			cloudIDs = append(cloudIDs, *one.GroupId)
			allCloudIDs[*one.GroupId] = struct{}{}
		}

		if len(cloudIDs) > 0 {
			req.CloudIDs = cloudIDs
		}
		_, err = securitygroup.SyncAwsSecurityGroup(cts.Kit, req, svc.adaptor, svc.dataCli)
		if err != nil {
			logs.Errorf("request to sync aws security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		if results.NextToken == nil {
			break
		}
		nextToken = *results.NextToken
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

	err = svc.deleteAwsSG(cts, client, req, deleteIDs)
	if err != nil {
		logs.Errorf("request deleteAwsSG failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *syncSecurityGroupSvc) deleteAwsSG(cts *rest.Contexts, client *aws.Aws,
	req *securitygroup.SyncAwsSecurityGroupOption, deleteIDs []string) error {

	if len(deleteIDs) > 0 {
		realDeleteIDs := make([]string, 0)
		nextToken := ""
		for {
			listOpt := &typessg.AwsListOption{
				Region: req.Region,
				Page: &typcore.AwsPage{
					MaxResults: converter.ValToPtr(int64(filter.DefaultMaxInLimit)),
				},
			}
			if nextToken != "" {
				listOpt.Page.NextToken = converter.ValToPtr(nextToken)
			}

			results, err := client.ListSecurityGroup(cts.Kit, listOpt)
			if err != nil {
				logs.Errorf("request adaptor to list aws security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return err
			}

			for _, id := range deleteIDs {
				realDeleteFlag := true
				for _, data := range results.SecurityGroups {
					if *data.GroupId == id {
						realDeleteFlag = false
						break
					}
				}

				if realDeleteFlag {
					realDeleteIDs = append(realDeleteIDs, id)
				}
			}

			if results.NextToken == nil {
				break
			}
			nextToken = *results.NextToken
		}

		if len(realDeleteIDs) > 0 {
			err := securitygroup.DiffSecurityGroupSyncDelete(cts.Kit, realDeleteIDs, svc.dataCli)
			if err != nil {
				logs.Errorf("sync delete aws security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return err
			}
		}
	}

	return nil
}

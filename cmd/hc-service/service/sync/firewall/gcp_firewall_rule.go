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

package firewall

import (
	"fmt"

	firewall "hcm/cmd/hc-service/logics/sync/firewall"
	"hcm/pkg/adaptor/gcp"
	adcore "hcm/pkg/adaptor/types/core"
	firewallrule "hcm/pkg/adaptor/types/firewall-rule"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SyncGcpFirewallRule sync gcp firewall rule to hcm.
func (svc *syncFireWallSvc) SyncGcpFirewallRule(cts *rest.Contexts) (interface{}, error) {

	syncReq := new(sync.GcpFirewallSyncReq)
	if err := cts.DecodeInto(syncReq); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := syncReq.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &firewall.GcpFirewallSyncOption{
		AccountID: syncReq.AccountID,
	}

	client, err := svc.adaptor.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	allCloudIDs := make(map[string]struct{})
	nextToken := ""
	for {
		opt := &firewallrule.ListOption{
			Page: &adcore.GcpPage{
				PageSize: int64(adcore.GcpQueryLimit),
			},
		}

		if nextToken != "" {
			opt.Page.PageToken = nextToken
		}

		results, token, err := client.ListFirewallRule(cts.Kit, opt)
		if err != nil {
			return nil, err
		}

		cloudIDs := make([]string, 0, len(results))
		for _, one := range results {
			cloudIDs = append(cloudIDs, fmt.Sprint(one.Id))
			allCloudIDs[fmt.Sprint(one.Id)] = struct{}{}
		}

		if len(cloudIDs) > 0 {
			req.CloudIDs = cloudIDs
		}
		_, err = firewall.SyncGcpFirewallRule(cts.Kit, req, svc.adaptor, svc.dataCli)
		if err != nil {
			logs.Errorf("request to sync gcp firewall rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		if len(token) == 0 {
			break
		}
		nextToken = token
	}

	commonReq := &firewall.GcpFirewallSyncOption{
		AccountID: req.AccountID,
	}
	dsIDs, err := firewall.GetDatasFromDSForGcpFireWallSync(cts.Kit, commonReq, svc.dataCli)
	if err != nil {
		return nil, err
	}

	deleteIDs := make([]string, 0)
	for id := range dsIDs {
		if _, ok := allCloudIDs[id]; !ok {
			deleteIDs = append(deleteIDs, id)
		}
	}

	err = svc.deleteGcpFireWall(cts, client, deleteIDs)
	if err != nil {
		logs.Errorf("request deleteGcpFireWall failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *syncFireWallSvc) deleteGcpFireWall(cts *rest.Contexts, client *gcp.Gcp,
	deleteIDs []string) error {

	if len(deleteIDs) > 0 {
		realDeleteIDs := make([]string, 0)
		nextToken := ""
		for {
			opt := &firewallrule.ListOption{
				Page: &adcore.GcpPage{
					PageSize: int64(adcore.GcpQueryLimit),
				},
			}

			if nextToken != "" {
				opt.Page.PageToken = nextToken
			}

			results, token, err := client.ListFirewallRule(cts.Kit, opt)
			if err != nil {
				return err
			}

			for _, id := range deleteIDs {
				realDeleteFlag := true
				for _, data := range results {
					if fmt.Sprint(data.Id) == id {
						realDeleteFlag = false
						break
					}
				}

				if realDeleteFlag {
					realDeleteIDs = append(realDeleteIDs, id)
				}
			}

			if len(token) == 0 {
				break
			}
			nextToken = token
		}

		if len(realDeleteIDs) > 0 {
			err := firewall.DiffFireWallSyncDelete(cts.Kit, realDeleteIDs, svc.dataCli)
			if err != nil {
				logs.Errorf("sync delete gcp firewall failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return err
			}
		}
	}

	return nil
}

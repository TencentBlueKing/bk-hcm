/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package matcher provide the matcher for task
package matcher

import (
	"fmt"

	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
)

// transferHostAndSetOperator transfer host and set operator in cc 3.0
func (m *Matcher) transferHostAndSetOperator(info *types.DeviceInfo, order *types.ApplyOrder) error {
	hostId, err := m.cc.GetHostIDByAssetID(m.kt, info.AssetId)
	if err != nil {
		logs.Errorf("failed to get host id by asset id: %s, err: %v", info.AssetId, err)
		return err
	}

	idMap, err := m.cc.GetHostBizIds(m.kt, []int64{hostId})
	if err != nil {
		logs.Errorf("failed to get host biz id, ip: %s, err: %v", info.Ip, err)
		return err
	}
	bizID, ok := idMap[hostId]
	if !ok {
		logs.Errorf("failed to get host biz id, ip: %s, err: %v", info.Ip, err)
		return fmt.Errorf("failed to get host biz id, ip: %s", info.Ip)
	}
	if bizID != enumor.ResourcePoolBiz {
		logs.Errorf("host is not in biz, ip: %s, biz id: %d", info.Ip, bizID)
		return nil
	}

	targetBizID := order.BkBizId
	if order.Source == enumor.ApplyTicketSrcPurchaseToResPool {
		targetBizID = cc.WoaServer().ApplyTicketConfig.PurchaseToResourcePool.BizID
	}
	if err := m.transferHost(info, hostId, targetBizID); err != nil {
		logs.Errorf("failed to transfer host, ip: %s, err: %v", info.Ip, err)
		return err
	}

	if err := m.UpdateHostOperator(info, hostId, order.User); err != nil {
		logs.Errorf("failed to update host operator, ip: %s, err: %v", info.Ip, err)
		return err
	}

	return nil
}

// transferHost transfer host to target business in cc 3.0
func (m *Matcher) transferHost(info *types.DeviceInfo, hostId, bizId int64) error {
	transferReq := &cmdb.TransferHostReq{
		From: cmdb.TransferHostSrcInfo{
			FromBizID: enumor.ResourcePoolBiz,
			HostIDs:   []int64{hostId},
		},
		To: cmdb.TransferHostDstInfo{
			ToBizID: bizId,
		},
	}

	err := m.cc.TransferHost(m.kt, transferReq)
	if err != nil {
		return err
	}

	return nil
}

// UpdateHostOperator update host operator in cc 3.0
func (m *Matcher) UpdateHostOperator(info *types.DeviceInfo, hostId int64, operator string) error {
	update := &cmdb.UpdateHostProperty{
		HostID: hostId,
		Properties: map[string]interface{}{
			"operator":        operator,
			"bk_bak_operator": operator,
		},
	}
	req := &cmdb.UpdateHostsReq{
		Update: []*cmdb.UpdateHostProperty{
			update,
		},
	}

	_, err := m.cc.UpdateHosts(m.kt, req)
	if err != nil {
		return err
	}
	return nil
}

func (m *Matcher) getBizName(bizId int64) string {
	req := &cmdb.SearchBizParams{
		BizPropertyFilter: &cmdb.QueryFilter{
			Rule: cmdb.CombinedRule{
				Condition: cmdb.ConditionAnd,
				Rules: []cmdb.Rule{
					cmdb.AtomRule{
						Field:    "bk_biz_id",
						Operator: cmdb.OperatorEqual,
						Value:    bizId,
					},
				},
			},
		},
		Fields: []string{"bk_biz_id", "bk_biz_name"},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: 1,
		},
	}

	resp, err := m.cc.SearchBusiness(m.kt, req)
	if err != nil {
		logs.Warnf("failed to get cc business info, err: %v", err)
		return ""
	}

	cnt := len(resp.Info)
	if cnt != 1 {
		logs.Warnf("get invalid cc business info count %d != 1", cnt)
		return ""
	}
	return resp.Info[0].BizName
}

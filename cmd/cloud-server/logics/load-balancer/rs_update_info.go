/*
 *
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package lblogic

import (
	"fmt"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client/data-service/global"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// RSUpdateInfo 单个RS的修改参数
type RSUpdateInfo struct {
	InstType  enumor.InstType `json:"inst_type"`
	IP        string          `json:"rsip"`
	Port      int             `json:"rsport"`
	EndPort   int             `json:"-"`
	OldWeight int             `json:"old_weight"`
	NewWeight int             `json:"new_weight"`
}

// ValidateOldWeight 校验RSInfo是否存在, 并获取其原有的权重进行比较
func (r *RSUpdateInfo) ValidateOldWeight(kt *kit.Kit, client *global.LoadBalancerClient, tgID string) error {

	target, err := r.GetTarget(kt, client, tgID, "")
	if err != nil {
		logs.Errorf("get target %s:%d error: %s, rid: %s", r.IP, r.Port, err, kt.Rid)
		return err
	}

	weight := converter.PtrToVal(target.Weight)
	if weight != int64(r.OldWeight) {
		logs.Errorf("target %s:%d oldWeight not match, input %d, actual %d, rid: %s",
			r.IP, r.Port, r.OldWeight, weight, kt.Rid)
		return fmt.Errorf("target %s:%d oldWeight not match, input %d, actual %d",
			r.IP, r.Port, r.OldWeight, weight)
	}

	return nil
}

// GetTarget 获取dataservice的Target信息
func (r *RSUpdateInfo) GetTarget(kt *kit.Kit, client *global.LoadBalancerClient,
	tgID, accountID string) (*dataproto.TargetBaseReq, error) {

	expression, err := tools.And(
		tools.RuleEqual("target_group_id", tgID),
		tools.RuleEqual("port", r.Port),
		tools.ExpressionOr(
			tools.RuleJSONContains("private_ip_address", r.IP),
			tools.RuleJSONContains("public_ip_address", r.IP),
		),
	)
	if err != nil {
		return nil, err
	}
	listReq := &core.ListReq{
		Filter: expression,
		Page:   core.NewDefaultBasePage(),
	}
	target, err := client.ListTarget(kt, listReq)
	if err != nil {
		logs.Errorf("list target error: %s, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(target.Details) == 0 {
		return nil, fmt.Errorf("target %s:%d not found", r.IP, r.Port)
	}
	item := target.Details[0]
	r.InstType = item.InstType
	return &dataproto.TargetBaseReq{
		ID:               item.ID,
		InstType:         item.InstType,
		CloudInstID:      item.CloudInstID,
		Port:             item.Port,
		Weight:           item.Weight,
		AccountID:        accountID,
		TargetGroupID:    tgID,
		InstName:         item.InstName,
		PrivateIPAddress: item.PrivateIPAddress,
		PublicIPAddress:  item.PublicIPAddress,
		CloudVpcIDs:      item.CloudVpcIDs,
		Zone:             item.Zone,
		NewWeight:        converter.ValToPtr(int64(r.NewWeight)),
	}, nil
}

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

package cslb

import (
	"errors"
	"fmt"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
)

// ListTargetByCondReq ...
type ListTargetByCondReq struct {
	Vendor        enumor.Vendor     `json:"vendor" validate:"required,min=1"`
	AccountID     string            `json:"account_id" validate:"required,min=1"`
	RuleQueryList []TargetQueryLine `json:"rule_query_list" validate:"required,min=1,max=50"`
}

// Validate ...
func (req *ListTargetByCondReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for _, line := range req.RuleQueryList {
		if err := line.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// TargetQueryLine ...
type TargetQueryLine struct {
	Protocol      enumor.ProtocolType `json:"protocol" validate:"required"`
	Region        string              `json:"region" validate:"required"`
	ClbVipDomains []string            `json:"clb_vip_domains" validate:"required,min=1,max=50"`
	CloudLbIds    []string            `json:"cloud_lb_ids" validate:"required,min=1,max=50"`
	ListenerPorts []int               `json:"listener_ports" validate:"omitempty,max=50"`
	RsIps         []string            `json:"rs_ips" validate:"omitempty,max=500"`
	RsPorts       []int               `json:"rs_ports" validate:"omitempty,max=500"`
	Domains       []string            `json:"domains" validate:"omitempty,max=50"`
	Urls          []string            `json:"urls" validate:"omitempty,max=50"`
}

// Validate ...
func (item *TargetQueryLine) Validate() error {
	if err := validator.Validate.Struct(item); err != nil {
		return err
	}
	if item.Protocol.IsLayer4Protocol() && (len(item.Domains) > 0 || len(item.Urls) > 0) {
		return errors.New("layer4 protocol should not have domains or urls")
	}

	if len(item.ClbVipDomains) != len(item.CloudLbIds) {
		return errors.New("clb_vip_domains and cloud_lb_ids num must be equal")
	}

	// 监听器端口和RSIP，不能同时为ANY
	if len(item.ListenerPorts) == 0 && len(item.RsIps) == 0 {
		return errors.New("listener_ports and rs_ips can not be empty at the same time")
	}

	return nil
}

// ListTargetByCondResp ...
type ListTargetByCondResp struct {
	Details []*ListTargetByCondResult `json:"details"`
}

// ListTargetByCondResult ...
type ListTargetByCondResult struct {
	ClbId        string              `json:"clb_id"`
	CloudLbId    string              `json:"cloud_lb_id"`
	ClbVipDomain string              `json:"clb_vip_domain"`
	BkBizId      int64               `json:"bk_biz_id"`
	Region       string              `json:"region"`
	Vendor       enumor.Vendor       `json:"vendor"`
	LblId        string              `json:"lbl_id"`
	CloudLblId   string              `json:"cloud_lbl_id"`
	Protocol     enumor.ProtocolType `json:"protocol"`
	Domain       string              `json:"domain,omitempty"`
	Url          string              `json:"url,omitempty"`
	Port         int64               `json:"port"`
	InstType     string              `json:"inst_type"`
	RsIp         string              `json:"rs_ip"`
	RsPort       int64               `json:"rs_port"`
	RsWeight     int64               `json:"rs_weight"`
}

// BatchModifyTargetWeightReq ...
type BatchModifyTargetWeightReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	TargetIDs []string `json:"target_ids" validate:"min=1"`
	NewWeight *int64   `json:"new_weight" validate:"required,min=0,max=100"`
}

// Validate ...
func (b *BatchModifyTargetWeightReq) Validate() error {
	if len(b.TargetIDs) > constant.BatchOperateModifyTargetWeightLimit {
		return fmt.Errorf("target_ids length count should <= %d", constant.BatchOperateModifyTargetWeightLimit)
	}
	return validator.Validate.Struct(b)
}

// BatchRemoveTargetReq ...
type BatchRemoveTargetReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	TargetIDs []string `json:"target_ids" validate:"min=1"`
}

// Validate ...
func (b *BatchRemoveTargetReq) Validate() error {
	if len(b.TargetIDs) > constant.BatchOperateRemoveTargetLimit {
		return fmt.Errorf("the number of target IDs cannot exceed %d", constant.BatchOperateRemoveTargetLimit)
	}
	return validator.Validate.Struct(b)
}

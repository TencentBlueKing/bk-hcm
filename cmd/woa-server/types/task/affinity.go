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

// Package task affinity check types
package task

import (
	"fmt"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/thirdparty/cvmapi"
)

// AffinityMatchReq 亲和性匹配请求
type AffinityMatchReq struct {
	BkBizID int64               `json:"bk_biz_id" validate:"required,gt=0"`      // CC业务ID
	Specs   []AffinityMatchSpec `json:"specs" validate:"required,min=1,max=100"` // 资源申请子需求单信息
}

// AffinityMatchSpec 资源申请子需求单信息
type AffinityMatchSpec struct {
	Region     string   `json:"region" validate:"required"`                            // 区域
	Zones      []string `json:"zones" validate:"required,min=1,max=100,dive,required"` // 可用区数组，最大最多传100个 (选"全部"时传all)
	DeviceType string   `json:"device_type" validate:"required"`                       // 机型
	Replicas   int      `json:"replicas" validate:"required,gt=0"`                     // 该机型对应的数量
}

// AffinityMatchResp 亲和性匹配响应
type AffinityMatchResp struct {
	Details []AffinityMatchDetail `json:"details"` // 查询返回的数据
}

// AffinityMatchDetail 亲和性匹配详情
type AffinityMatchDetail struct {
	Region     string                `json:"region" validate:"required"`      // 区域
	Zone       string                `json:"zone" validate:"required"`        // 可用区
	DeviceType string                `json:"device_type" validate:"required"` // 机型
	Replicas   int                   `json:"replicas" validate:"required"`    // 该机型对应的数量
	Status     enumor.AffinityStatus `json:"status" validate:"required"`      // 匹配状态（1:CRP预检有数据 2:CRP预检无数据）
	MaxCutNum  int                   `json:"max_cut_num"`                     // 最大切片数量，虚拟比 = 1:max_cut_num
	IPs        []string              `json:"ips"`                             // 申请后的母机IP分布
}

// CampusInfo 园区信息
type CampusInfo struct {
	CampusID   string `json:"campus_id"`   // 园区ID
	CampusName string `json:"campus_name"` // 园区名称
	Zone       string `json:"zone"`        // 可用区
}

// Validate 验证请求参数
func (req *AffinityMatchReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for i, spec := range req.Specs {
		if err := spec.Validate(); err != nil {
			return fmt.Errorf("specs[%d]: %w", i, err)
		}
	}

	return nil
}

// Validate 验证资源申请子需求单信息
func (spec *AffinityMatchSpec) Validate() error {
	if err := validator.Validate.Struct(spec); err != nil {
		return err
	}
	return nil
}

// IsCVMSeparateCampus 是否分Campus申请单
func (spec *AffinityMatchSpec) IsCVMSeparateCampus() bool {
	for _, zone := range spec.Zones {
		if zone == cvmapi.CvmSeparateCampus {
			return true
		}
	}
	return false
}

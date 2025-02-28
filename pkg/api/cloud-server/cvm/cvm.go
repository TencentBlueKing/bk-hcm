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

package cscvm

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	rr "hcm/pkg/api/core/recycle-record"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"
)

// AssignCvmToBizReq define assign cvm to biz req.
type AssignCvmToBizReq struct {
	Cvms []AssignCvmToBizData `json:"cvms" validate:"required,min=1,max=500"`
}

// Validate assign cvm to biz request.
func (req *AssignCvmToBizReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for _, cvm := range req.Cvms {
		if err := cvm.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// AssignCvmToBizData define assign cvm to biz data.
type AssignCvmToBizData struct {
	CvmID     string `json:"cvm_id" validate:"required"`
	BkBizID   int64  `json:"bk_biz_id" validate:"required"`
	BkCloudID *int64 `json:"bk_cloud_id" validate:"required"`
}

// Validate assign cvm to biz data.
func (req *AssignCvmToBizData) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.BkBizID <= 0 {
		return errors.New("bk_biz_id should >= 0")
	}

	// todo 本期暂不支持管控区域为0
	if converter.PtrToVal(req.BkCloudID) == 0 {
		return errors.New("bk_cloud_id should != 0")
	}

	return nil
}

// AssignCvmToBizPreviewReq define assign cvm to biz preview req.
type AssignCvmToBizPreviewReq struct {
	CvmIDs []string `json:"cvm_ids"`
}

// Validate assign cvm to biz preview request.
func (req *AssignCvmToBizPreviewReq) Validate() error {
	if len(req.CvmIDs) == 0 {
		return errors.New("cvm ids is required")
	}

	if len(req.CvmIDs) > int(core.DefaultMaxPageLimit) {
		return fmt.Errorf("cvm ids length should <= %d", core.DefaultMaxPageLimit)
	}

	return nil
}

// AssignCvmToBizPreviewData define assign cvm to biz preview data.
type AssignCvmToBizPreviewData struct {
	Details []AssignCvmToBizPreviewDetail `json:"details"`
}

// AssignCvmToBizPreviewDetail define assign cvm to biz preview detail.
type AssignCvmToBizPreviewDetail struct {
	CvmID     string              `json:"cvm_id"`
	MatchType enumor.CvmMatchType `json:"match_type"`
	BizID     int64               `json:"bk_biz_id,omitempty"`
	BkCloudID *int64              `json:"bk_cloud_id,omitempty"`
}

// ListAssignedCvmMatchHostReq list assigned cvm match host req.
type ListAssignedCvmMatchHostReq struct {
	AccountID            string   `json:"account_id" validate:"required"`
	PrivateIPv4Addresses []string `json:"private_ipv4_addresses" validate:"required,min=1,max=10"`
}

// Validate list assigned cvm match host request.
func (req *ListAssignedCvmMatchHostReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ListAssignedCvmMatchHostData define list assigned cvm match host data.
type ListAssignedCvmMatchHostData struct {
	Details []ListAssignedCvmMatchHostDetail `json:"details"`
}

// ListAssignedCvmMatchHostDetail define list assigned cvm match host detail.
type ListAssignedCvmMatchHostDetail struct {
	BkHostID             int64    `json:"bk_host_id"`
	PrivateIPv4Addresses []string `json:"private_ipv4_addresses"`
	PublicIPv4Addresses  []string `json:"public_ipv4_addresses"`
	BkCloudID            int64    `json:"bk_cloud_id"`
	BkBizID              int64    `json:"bk_biz_id"`
	Region               string   `json:"region"`
	BkHostName           string   `json:"bk_host_name"`
	BkOsName             string   `json:"bk_os_name"`
	CreateTime           string   `json:"create_time"`
}

// BatchStartCvmReq batch start cvm req.
type BatchStartCvmReq struct {
	IDs []string `json:"ids" validate:"required"`
}

// Validate batch start cvm request.
func (req *BatchStartCvmReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cvm ids should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// BatchStopCvmReq batch stop cvm req.
type BatchStopCvmReq struct {
	IDs []string `json:"ids" validate:"required"`
}

// Validate batch stop cvm request.
func (req *BatchStopCvmReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cvm ids should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// BatchRebootCvmReq batch reboot cvm req.
type BatchRebootCvmReq struct {
	IDs []string `json:"ids" validate:"required"`
}

// Validate batch reboot cvm request.
func (req *BatchRebootCvmReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cvm ids should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// -------------------------- Recycle ------------------------

// CvmRecycleReq recycle cvm request.
type CvmRecycleReq struct {
	Infos []CvmRecycleInfo `json:"infos" validate:"min=1,max=100"`
}

// CvmRecycleInfo defines recycle one cvm info.
type CvmRecycleInfo struct {
	ID                   string `json:"id" validate:"required"`
	rr.CvmRecycleOptions `json:",inline" validate:"required"`
}

// Validate CvmRecycleReq
func (req CvmRecycleReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Recover ------------------------

// CvmRecoverReq recover cvm request.
type CvmRecoverReq struct {
	RecordIDs []string `json:"record_ids" validate:"min=1,max=100"`
}

// Validate CvmRecoverReq
func (req CvmRecoverReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Recover ------------------------

// CvmDeleteRecycledReq delete recycled cvm request.
type CvmDeleteRecycledReq struct {
	RecordIDs []string `json:"record_ids" validate:"min=1,max=100"`
}

// Validate CvmDeleteRecycledReq
func (req CvmDeleteRecycledReq) Validate() error {
	return validator.Validate.Struct(req)
}

// BatchQueryCvmRelatedReq  批量查询cvm关联资源请求
type BatchQueryCvmRelatedReq struct {
	IDs []string `json:"ids" validate:"min=1,max=100"`
}

// Validate ...
func (req BatchQueryCvmRelatedReq) Validate() error {
	return validator.Validate.Struct(req)
}

// CvmRelatedResult  批量查询cvm关联资源响应
type CvmRelatedResult struct {
	Detail []CvmRelatedInfo
}

// CvmRelatedInfo Cvm 关联资源信息
type CvmRelatedInfo struct {
	DiskCount int      `json:"disk_count"`
	EipCount  int      `json:"eip_count"`
	Eip       []string `json:"eip"`
}

// BatchGetCvmSecurityGroupsReq ...
type BatchGetCvmSecurityGroupsReq struct {
	CvmIDs []string `json:"cvm_ids" validate:"required,min=1,max=500"`
}

// Validate ...
func (req BatchGetCvmSecurityGroupsReq) Validate() error {
	return validator.Validate.Struct(req)
}

// BatchListCvmSecurityGroupsResp ...
type BatchListCvmSecurityGroupsResp struct {
	CvmID          string                               `json:"cvm_id"`
	SecurityGroups []BatchListCvmSecurityGroupsRespItem `json:"security_groups"`
}

// BatchListCvmSecurityGroupsRespItem ...
type BatchListCvmSecurityGroupsRespItem struct {
	ID      string `json:"id"`
	CloudId string `json:"cloud_id"`
	Name    string `json:"name"`
}

// BatchAssociateSecurityGroupsReq ...
type BatchAssociateSecurityGroupsReq struct {
	SecurityGroupIDs []string `json:"security_group_ids" validate:"required,min=1,max=500"`
}

// Validate ...
func (req BatchAssociateSecurityGroupsReq) Validate() error {
	return validator.Validate.Struct(req)
}

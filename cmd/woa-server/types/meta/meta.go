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

// Package meta ...
package meta

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// DiskTypeItem defines disk type item.
type DiskTypeItem struct {
	DiskType     enumor.DiskType `json:"disk_type"`
	DiskTypeName string          `json:"disk_type_name"`
}

// TicketTypeItem defines ticket type item.
type TicketTypeItem struct {
	TicketType     enumor.RPTicketType `json:"ticket_type"`
	TicketTypeName string              `json:"ticket_type_name"`
}

// ListZoneReq defines list zone request.
type ListZoneReq struct {
	RegionIDs []string `json:"region_ids"`
}

// Validate whether ListZoneReq is valid.
func (r *ListZoneReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	return nil
}

// ListDeviceTypeReq defines list device type request.
type ListDeviceTypeReq struct {
	DeviceClasses []string `json:"device_classes"`
	DeviceTypes   []string `json:"device_types"`
}

// Validate whether ListDeviceTypeReq is valid.
func (r *ListDeviceTypeReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	return nil
}

// ListDeviceTypeRst defines list device type result.
type ListDeviceTypeRst struct {
	DeviceType   string `json:"device_type"`
	CoreType     string `json:"core_type"`
	CpuCore      int64  `json:"cpu_core"`
	Memory       int64  `json:"memory"`
	DeviceClass  string `json:"device_class"`
	DeviceFamily string `json:"device_family"`
}

// ListBizsByOpProdReq defines list bizs by op product request.
type ListBizsByOpProdReq struct {
	OpProductID int64 `json:"op_product_id" validate:"required"`
}

// Validate whether ListBizsByOpProdReq is valid.
func (r *ListBizsByOpProdReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if r.OpProductID <= 0 {
		return errors.New("invalid op product id, should be > 0")
	}

	return nil
}

// BizOrgRel is GetBizOrgRel result.
type BizOrgRel struct {
	BkBizID         int64  `json:"bk_biz_id"`
	BkBizName       string `json:"bk_biz_name"`
	OpProductID     int64  `json:"op_product_id"`
	OpProductName   string `json:"op_product_name"`
	PlanProductID   int64  `json:"plan_product_id"`
	PlanProductName string `json:"plan_product_name"`
	VirtualDeptID   int64  `json:"virtual_dept_id"`
	VirtualDeptName string `json:"virtual_dept_name"`
}

// Biz is GetBizs result.
type Biz struct {
	BkBizID   int64  `json:"bk_biz_id"`
	BkBizName string `json:"bk_biz_name"`
}

// OpProduct is GetOpProducts result.
type OpProduct struct {
	OpProductID   int64  `json:"op_product_id"`
	OpProductName string `json:"op_product_name"`
}

// PlanProduct is GetPlanProducts result.
type PlanProduct struct {
	PlanProductID   int64  `json:"plan_product_id"`
	PlanProductName string `json:"plan_product_name"`
}

// OrgTopoReq is the get org topo request
type OrgTopoReq struct {
	View enumor.View `json:"view" validate:"required"`
}

// Validate OrgTopoReq
func (req *OrgTopoReq) Validate() error {
	return req.View.Validate()
}

// OrgInfo define cost organization information.
type OrgInfo struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	FullName    string     `json:"full_name"`
	Level       int64      `json:"level"`
	HasChildren bool       `json:"has_children"`
	Children    []*OrgInfo `json:"children"`
	TofDeptID   int        `json:"tof_dept_id"`
	Parent      string     `json:"parent"`
}

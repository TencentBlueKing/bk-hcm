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
	"fmt"

	"hcm/cmd/data-service/service/audit/cloud/cvm"
	auditlb "hcm/cmd/data-service/service/audit/cloud/load-balancer"
	networkinterface "hcm/cmd/data-service/service/audit/cloud/network-interface"
	"hcm/cmd/data-service/service/audit/cloud/subnet"
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// NewSecurityGroup new firewall.
func NewSecurityGroup(dao dao.Set) *SecurityGroup {
	return &SecurityGroup{
		dao: dao,
	}
}

// SecurityGroup define firewall audit.
type SecurityGroup struct {
	dao dao.Set
}

// SecurityGroupUpdateAuditBuild security group update audit.
func (s *SecurityGroup) SecurityGroupUpdateAuditBuild(kt *kit.Kit, updates []protoaudit.CloudResourceUpdateInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(updates))
	for _, one := range updates {
		ids = append(ids, one.ResID)
	}
	idSgMap, err := s.listSecurityGroup(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		sg, exist := idSgMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: sg.CloudID,
			ResName:    sg.Name,
			ResType:    enumor.SecurityGroupAuditResType,
			Action:     enumor.Update,
			BkBizID:    sg.BkBizID,
			Vendor:     sg.Vendor,
			AccountID:  sg.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data:    sg,
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil
}

// SecurityGroupDeleteAuditBuild security group delete audit.
func (s *SecurityGroup) SecurityGroupDeleteAuditBuild(kt *kit.Kit, deletes []protoaudit.CloudResourceDeleteInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(deletes))
	for _, one := range deletes {
		ids = append(ids, one.ResID)
	}
	idSgMap, err := s.listSecurityGroup(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(deletes))
	for _, one := range deletes {
		sg, exist := idSgMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: sg.CloudID,
			ResName:    sg.Name,
			ResType:    enumor.SecurityGroupAuditResType,
			Action:     enumor.Delete,
			BkBizID:    sg.BkBizID,
			Vendor:     sg.Vendor,
			AccountID:  sg.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: sg,
			},
		})
	}

	return audits, nil
}

// SecurityGroupAssignAuditBuild security group assign audit.
func (s *SecurityGroup) SecurityGroupAssignAuditBuild(kt *kit.Kit, assigns []protoaudit.CloudResourceAssignInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(assigns))
	for _, one := range assigns {
		ids = append(ids, one.ResID)
	}
	idSgMap, err := s.listSecurityGroup(kt, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(assigns))
	for _, one := range assigns {
		sg, exist := idSgMap[one.ResID]
		if !exist {
			continue
		}

		if one.AssignedResType != enumor.BizAuditAssignedResType {
			return nil, errf.New(errf.InvalidParameter, "assigned resource type is invalid")
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: sg.CloudID,
			ResName:    sg.Name,
			ResType:    enumor.SecurityGroupAuditResType,
			Action:     enumor.Assign,
			BkBizID:    sg.BkBizID,
			Vendor:     sg.Vendor,
			AccountID:  sg.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Changed: map[string]interface{}{
					"bk_biz_id": one.AssignedResID,
				},
			},
		})
	}

	return audits, nil
}

// listSecurityGroup lists security groups by their IDs.
func (s *SecurityGroup) listSecurityGroup(kt *kit.Kit, ids []string) (map[string]tablecloud.SecurityGroupTable, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := s.dao.SecurityGroup().List(kt, opt)
	if err != nil {
		logs.Errorf("list security group failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablecloud.SecurityGroupTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}

// OperationAuditBuild ...
func (s *SecurityGroup) OperationAuditBuild(kt *kit.Kit, operations []protoaudit.CloudResourceOperationInfo) (
	[]*tableaudit.AuditTable, error) {

	cvmAssOperations := make([]protoaudit.CloudResourceOperationInfo, 0)
	subnetAssOperations := make([]protoaudit.CloudResourceOperationInfo, 0)
	niAssOperations := make([]protoaudit.CloudResourceOperationInfo, 0)
	clbAssOperations := make([]protoaudit.CloudResourceOperationInfo, 0)
	for _, operation := range operations {
		switch operation.Action {
		case protoaudit.Associate, protoaudit.Disassociate:
			switch operation.AssociatedResType {
			case enumor.CvmAuditResType:
				cvmAssOperations = append(cvmAssOperations, operation)
			case enumor.SubnetAuditResType:
				subnetAssOperations = append(subnetAssOperations, operation)
			case enumor.NetworkInterfaceAuditResType:
				niAssOperations = append(niAssOperations, operation)
			case enumor.LoadBalancerAuditResType:
				clbAssOperations = append(clbAssOperations, operation)
			default:
				return nil, fmt.Errorf("audit associated resource type: %s not support", operation.AssociatedResType)
			}

		default:
			return nil, fmt.Errorf("audit action: %s not support", operation.Action)
		}
	}

	audits := make([]*tableaudit.AuditTable, 0, len(operations))
	if len(cvmAssOperations) != 0 {
		audit, err := s.cvmAssOperationAuditBuild(kt, cvmAssOperations)
		if err != nil {
			return nil, err
		}

		audits = append(audits, audit...)
	}

	if len(subnetAssOperations) != 0 {
		audit, err := s.subnetAssOperationAuditBuild(kt, subnetAssOperations)
		if err != nil {
			return nil, err
		}

		audits = append(audits, audit...)
	}

	if len(niAssOperations) != 0 {
		audit, err := s.niAssOperationAuditBuild(kt, niAssOperations)
		if err != nil {
			return nil, err
		}

		audits = append(audits, audit...)
	}

	if len(clbAssOperations) != 0 {
		audit, err := s.clbAssOperationAuditBuild(kt, clbAssOperations)
		if err != nil {
			return nil, err
		}

		audits = append(audits, audit...)
	}

	return audits, nil
}

// cvmAssOperationAuditBuild builds audit for security group associated with CVM.
func (s *SecurityGroup) cvmAssOperationAuditBuild(kt *kit.Kit, operations []protoaudit.CloudResourceOperationInfo) (
	[]*tableaudit.AuditTable, error) {

	sgIDs := make([]string, 0)
	cvmIDs := make([]string, 0)
	for _, one := range operations {
		sgIDs = append(sgIDs, one.ResID)
		cvmIDs = append(cvmIDs, one.AssociatedResID)
	}

	sgIDMap, err := s.listSecurityGroup(kt, sgIDs)
	if err != nil {
		return nil, err
	}

	cvmIDMap, err := cvm.ListCvm(kt, s.dao, cvmIDs)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(operations))
	for _, one := range operations {
		sg, exist := sgIDMap[one.ResID]
		if !exist {
			return nil, errf.Newf(errf.RecordNotFound, "security group: %s not found", one.ResID)
		}

		cvmInfo, exist := cvmIDMap[one.AssociatedResID]
		if !exist {
			return nil, errf.Newf(errf.RecordNotFound, "cvm: %s not found", one.AssociatedResID)
		}

		action, err := one.Action.ConvAuditAction()
		if err != nil {
			return nil, err
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      sg.ID,
			CloudResID: sg.CloudID,
			ResName:    sg.Name,
			ResType:    enumor.SecurityGroupAuditResType,
			Action:     action,
			BkBizID:    sg.BkBizID,
			Vendor:     sg.Vendor,
			AccountID:  sg.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: &tableaudit.AssociatedOperationAudit{
					AssResType:    enumor.CvmAuditResType,
					AssResID:      cvmInfo.ID,
					AssResCloudID: cvmInfo.CloudID,
					AssResName:    cvmInfo.Name,
				},
			},
		})
	}

	return audits, nil
}

// cvmAssOperationAuditBuild builds audit for security group associated with CVM.
func (s *SecurityGroup) subnetAssOperationAuditBuild(kt *kit.Kit, operations []protoaudit.CloudResourceOperationInfo) (
	[]*tableaudit.AuditTable, error) {

	sgIDs := make([]string, 0)
	subnetIDs := make([]string, 0)
	for _, one := range operations {
		sgIDs = append(sgIDs, one.ResID)
		subnetIDs = append(subnetIDs, one.AssociatedResID)
	}

	sgIDMap, err := s.listSecurityGroup(kt, sgIDs)
	if err != nil {
		return nil, err
	}

	subnetIDMap, err := subnet.ListSubnet(kt, s.dao, subnetIDs)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(operations))
	for _, one := range operations {
		sg, exist := sgIDMap[one.ResID]
		if !exist {
			return nil, errf.Newf(errf.RecordNotFound, "security group: %s not found", one.ResID)
		}

		subnetInfo, exist := subnetIDMap[one.AssociatedResID]
		if !exist {
			return nil, errf.Newf(errf.RecordNotFound, "subnet: %s not found", one.AssociatedResID)
		}

		action, err := one.Action.ConvAuditAction()
		if err != nil {
			return nil, err
		}

		subnetName := ""
		if subnetInfo.Name != nil {
			subnetName = *subnetInfo.Name
		}
		audits = append(audits, &tableaudit.AuditTable{
			ResID:      sg.ID,
			CloudResID: sg.CloudID,
			ResName:    sg.Name,
			ResType:    enumor.SecurityGroupAuditResType,
			Action:     action,
			BkBizID:    sg.BkBizID,
			Vendor:     sg.Vendor,
			AccountID:  sg.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: &tableaudit.AssociatedOperationAudit{
					AssResType:    enumor.SubnetAuditResType,
					AssResID:      subnetInfo.ID,
					AssResCloudID: subnetInfo.CloudID,
					AssResName:    subnetName,
				},
			},
		})
	}

	return audits, nil
}

// niAssOperationAuditBuild builds audit for security group associated with network interface.
func (s *SecurityGroup) niAssOperationAuditBuild(kt *kit.Kit, operations []protoaudit.CloudResourceOperationInfo) (
	[]*tableaudit.AuditTable, error) {

	sgIDs := make([]string, 0)
	niIDs := make([]string, 0)
	for _, one := range operations {
		sgIDs = append(sgIDs, one.ResID)
		niIDs = append(niIDs, one.AssociatedResID)
	}

	sgIDMap, err := s.listSecurityGroup(kt, sgIDs)
	if err != nil {
		return nil, err
	}

	niIDMap, err := networkinterface.ListNetworkInterface(kt, s.dao, niIDs)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(operations))
	for _, one := range operations {
		sg, exist := sgIDMap[one.ResID]
		if !exist {
			return nil, errf.Newf(errf.RecordNotFound, "security group: %s not found", one.ResID)
		}

		ni, exist := niIDMap[one.AssociatedResID]
		if !exist {
			return nil, errf.Newf(errf.RecordNotFound, "network interface: %s not found", one.AssociatedResID)
		}

		action, err := one.Action.ConvAuditAction()
		if err != nil {
			return nil, err
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      sg.ID,
			CloudResID: sg.CloudID,
			ResName:    sg.Name,
			ResType:    enumor.SecurityGroupAuditResType,
			Action:     action,
			BkBizID:    sg.BkBizID,
			Vendor:     sg.Vendor,
			AccountID:  sg.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: &tableaudit.AssociatedOperationAudit{
					AssResType:    enumor.NetworkInterfaceAuditResType,
					AssResID:      ni.ID,
					AssResCloudID: ni.CloudID,
					AssResName:    ni.Name,
				},
			},
		})
	}

	return audits, nil
}

// clbAssOperationAuditBuild builds audit for security group associated with load balancer.
func (s *SecurityGroup) clbAssOperationAuditBuild(kt *kit.Kit, operations []protoaudit.CloudResourceOperationInfo) (
	[]*tableaudit.AuditTable, error) {

	sgIDs := make([]string, 0)
	clbIDs := make([]string, 0)
	for _, one := range operations {
		sgIDs = append(sgIDs, one.ResID)
		clbIDs = append(clbIDs, one.AssociatedResID)
	}

	sgIDMap, err := s.listSecurityGroup(kt, sgIDs)
	if err != nil {
		return nil, err
	}

	clbIDMap, err := auditlb.ListLoadBalancer(kt, s.dao, clbIDs)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(operations))
	for _, one := range operations {
		sg, exist := sgIDMap[one.ResID]
		if !exist {
			return nil, errf.Newf(errf.RecordNotFound, "security group: %s not found", one.ResID)
		}

		clbInfo, exist := clbIDMap[one.AssociatedResID]
		if !exist {
			return nil, errf.Newf(errf.RecordNotFound, "clb: %s not found", one.AssociatedResID)
		}

		action, err := one.Action.ConvAuditAction()
		if err != nil {
			return nil, err
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      sg.ID,
			CloudResID: sg.CloudID,
			ResName:    sg.Name,
			ResType:    enumor.SecurityGroupAuditResType,
			Action:     action,
			BkBizID:    sg.BkBizID,
			Vendor:     sg.Vendor,
			AccountID:  sg.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: &tableaudit.AssociatedOperationAudit{
					AssResType:    enumor.LoadBalancerAuditResType,
					AssResID:      clbInfo.ID,
					AssResCloudID: clbInfo.CloudID,
					AssResName:    clbInfo.Name,
				},
			},
		})
	}

	return audits, nil
}

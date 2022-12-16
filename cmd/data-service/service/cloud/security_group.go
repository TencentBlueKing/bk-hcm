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

package cloud

import (
	"fmt"
	"net/http"
	"reflect"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	daotypes "hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// InitSecurityGroupService initial the security group service
func InitSecurityGroupService(cap *capability.Capability) {
	svc := &securityGroupSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("CreateSecurityGroup", http.MethodPost, "/vendors/{vendor}/security_groups/create", svc.CreateSecurityGroup)
	h.Add("UpdateSecurityGroup", http.MethodPatch, "/vendors/{vendor}/security_groups/{security_group_id}",
		svc.UpdateSecurityGroup)
	h.Add("GetSecurityGroup", http.MethodGet, "/vendors/{vendor}/security_groups/{security_group_id}",
		svc.GetSecurityGroup)
	h.Add("ListSecurityGroup", http.MethodPost, "/security_groups/list", svc.ListSecurityGroup)
	h.Add("DeleteSecurityGroup", http.MethodDelete, "/security_groups/batch", svc.DeleteSecurityGroup)
	h.Add("GetSecurityGroupVendor", http.MethodGet, "/security_groups/batch", svc.DeleteSecurityGroup)

	h.Load(cap.WebService)
}

type securityGroupSvc struct {
	dao dao.Set
}

// CreateSecurityGroup create security group.
func (svc *securityGroupSvc) CreateSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return createSecurityGroup[corecloud.TCloudSecurityGroupExtension](vendor, svc, cts)
	case enumor.Aws:
		return createSecurityGroup[corecloud.AwsSecurityGroupExtension](vendor, svc, cts)
	case enumor.HuaWei:
		return createSecurityGroup[corecloud.HuaWeiSecurityGroupExtension](vendor, svc, cts)
	case enumor.Azure:
		return createSecurityGroup[corecloud.AzureSecurityGroupExtension](vendor, svc, cts)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

// UpdateSecurityGroup update security group.
func (svc *securityGroupSvc) UpdateSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	switch vendor {
	case enumor.TCloud:
		return updateSecurityGroup[corecloud.TCloudSecurityGroupExtension](id, svc, cts)
	case enumor.Aws:
		return updateSecurityGroup[corecloud.AwsSecurityGroupExtension](id, svc, cts)
	case enumor.HuaWei:
		return updateSecurityGroup[corecloud.HuaWeiSecurityGroupExtension](id, svc, cts)
	case enumor.Azure:
		return updateSecurityGroup[corecloud.AzureSecurityGroupExtension](id, svc, cts)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

// ListSecurityGroup list security group.
func (svc *securityGroupSvc) ListSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.SecurityGroupListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Field,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.SecurityGroup().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list security group failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.SecurityGroupListResult{Count: result.Count}, nil
	}

	details := make([]*corecloud.BaseSecurityGroup, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, &corecloud.BaseSecurityGroup{
			ID:     one.ID,
			Vendor: enumor.Vendor(one.Vendor),
			Spec: &corecloud.SecurityGroupSpec{
				CloudID:   one.CloudID,
				Assigned:  one.Assigned,
				Region:    one.Region,
				Name:      one.Name,
				Memo:      one.Memo,
				AccountID: one.AccountID,
			},
			Revision: &core.Revision{
				Creator:   one.Creator,
				Reviser:   one.Reviser,
				CreatedAt: one.CreatedAt,
				UpdatedAt: one.UpdatedAt,
			},
		})
	}

	return &protocloud.SecurityGroupListResult{Details: details}, nil
}

// DeleteSecurityGroup delete security group.
func (svc *securityGroupSvc) DeleteSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.SecurityGroupDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: []string{"id"},
		Filter: req.Filter,
		Page: &types.BasePage{
			Start: 0,
			Limit: types.DefaultMaxPageLimit,
		},
	}
	listResp, err := svc.dao.SecurityGroup().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list security group failed, err: %v, rid: %s", cts.Kit.Rid)
		return nil, fmt.Errorf("list security group failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delFilter := tools.ContainersExpression("id", delIDs)
		if err := svc.dao.SecurityGroup().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
			return nil, err
		}

		// TODO: add delete relation operation.

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// GetSecurityGroup get security group detail.
func (svc *securityGroupSvc) GetSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	sgTable, err := getSecurityGroupByID(id, svc, cts)
	if err != nil {
		return nil, err
	}

	// TODO: 添加查询管理信息逻辑

	base := &corecloud.BaseSecurityGroup{
		ID:     sgTable.ID,
		Vendor: enumor.Vendor(sgTable.Vendor),
		Spec: &corecloud.SecurityGroupSpec{
			CloudID:   sgTable.CloudID,
			Assigned:  sgTable.Assigned,
			Region:    sgTable.Region,
			Name:      sgTable.Name,
			Memo:      sgTable.Memo,
			AccountID: sgTable.AccountID,
		},
		Revision: &core.Revision{
			Creator:   sgTable.Creator,
			Reviser:   sgTable.Reviser,
			CreatedAt: sgTable.CreatedAt,
			UpdatedAt: sgTable.UpdatedAt,
		},
	}

	switch enumor.Vendor(sgTable.Vendor) {
	case enumor.TCloud:
		return convertToSGResult[corecloud.TCloudSecurityGroupExtension](base, nil, sgTable.Extension)
	case enumor.Aws:
		return convertToSGResult[corecloud.AwsSecurityGroupExtension](base, nil, sgTable.Extension)
	case enumor.HuaWei:
		return convertToSGResult[corecloud.HuaWeiSecurityGroupExtension](base, nil, sgTable.Extension)
	case enumor.Azure:
		return convertToSGResult[corecloud.AwsSecurityGroupExtension](base, nil, sgTable.Extension)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

func convertToSGResult[T corecloud.SecurityGroupExtension](base *corecloud.BaseSecurityGroup, atm *corecloud.
	SecurityGroupAttachment, extJson tabletype.JsonField) (*corecloud.SecurityGroup[T], error) {

	var extension *T
	err := json.UnmarshalFromString(string(extJson), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString security group json extension failed, err: %v", err)
	}

	return &corecloud.SecurityGroup[T]{
		BaseSecurityGroup: *base,
		Attachment:        atm,
		Extension:         extension,
	}, nil
}

func updateSecurityGroup[T corecloud.SecurityGroupExtension](id string, svc *securityGroupSvc,
	cts *rest.Contexts) (interface{}, error) {

	req := new(protocloud.SecurityGroupUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sg := new(tablecloud.SecurityGroupTable)
	if req.Spec != nil {
		sg.Name = req.Spec.Name
		sg.Memo = req.Spec.Memo
		sg.Assigned = req.Spec.Assigned
	}

	if req.Extension != nil {
		sgTable, err := getSecurityGroupByID(id, svc, cts)
		if err != nil {
			return nil, err
		}

		merge, err := json.UpdateMerge(req.Extension, string(sgTable.Extension))
		if err != nil {
			return nil, fmt.Errorf("json UpdateMerge extension failed, err: %v", err)
		}

		sg.Extension = tabletype.JsonField(merge)
	}

	if err := svc.dao.SecurityGroup().Update(cts.Kit, tools.EqualExpression("id", id), sg); err != nil {
		logs.Errorf("update security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("update security group failed, err: %v", err)
	}

	return nil, nil
}

func getSecurityGroupByID(id string, svc *securityGroupSvc, cts *rest.Contexts) (*tablecloud.SecurityGroupTable, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", id),
		Page:   &daotypes.BasePage{Count: false, Start: 0, Limit: 1},
	}
	result, err := svc.dao.SecurityGroup().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list account failed, err: %v, rid: %s", cts.Kit.Rid)
		return nil, fmt.Errorf("list account failed, err: %v", err)
	}

	if len(result.Details) != 1 {
		return nil, errf.New(errf.RecordNotFound, "security group not found")
	}

	return &result.Details[0], nil
}

func createSecurityGroup[T corecloud.SecurityGroupExtension](vendor enumor.Vendor, svc *securityGroupSvc,
	cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.SecurityGroupCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	sgID, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		extension, err := json.MarshalToString(req.Extension)
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		sg := &tablecloud.SecurityGroupTable{
			Vendor:    string(vendor),
			CloudID:   req.Spec.CloudID,
			Assigned:  req.Spec.Assigned,
			Region:    req.Spec.Region,
			Name:      req.Spec.Name,
			Memo:      req.Spec.Memo,
			AccountID: req.Spec.AccountID,
			Extension: tabletype.JsonField(extension),
			Creator:   cts.Kit.User,
			Reviser:   cts.Kit.User,
		}
		sgID, err := svc.dao.SecurityGroup().CreateWithTx(cts.Kit, txn, sg)
		if err != nil {
			return nil, fmt.Errorf("create security group failed, err: %v", err)
		}

		return sgID, nil
	})
	if err != nil {
		return nil, err
	}

	id, ok := sgID.(string)
	if !ok {
		return nil, fmt.Errorf("create security group but return id type is not string, id type: %v",
			reflect.TypeOf(sgID).String())
	}

	return &core.CreateResult{ID: id}, nil
}

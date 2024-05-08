/*
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

package securitygroup

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
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// InitSecurityGroupService initial the security group service
func InitSecurityGroupService(cap *capability.Capability) {
	initSecurityGroupService(cap)
	initTCloudSGRuleService(cap)
	initHuaWeiSGRuleService(cap)
	initAzureSGRuleService(cap)
	initAwsSGRuleService(cap)

	initSGServiceHook(cap)
}

// initSecurityGroupService initial the security group service
func initSecurityGroupService(cap *capability.Capability) {
	svc := &securityGroupSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("BatchCreateSecurityGroup", http.MethodPost, "/vendors/{vendor}/security_groups/batch/create",
		svc.BatchCreateSecurityGroup)
	h.Add("BatchUpdateSecurityGroup", http.MethodPatch, "/vendors/{vendor}/security_groups/batch/update",
		svc.BatchUpdateSecurityGroup)
	h.Add("GetSecurityGroup", http.MethodGet, "/vendors/{vendor}/security_groups/{id}",
		svc.GetSecurityGroup)
	h.Add("ListSecurityGroup", http.MethodPost, "/security_groups/list", svc.ListSecurityGroup)
	h.Add("ListSecurityGroupExt", http.MethodPost, "/vendors/{vendor}/security_groups/list", svc.ListSecurityGroupExt)
	h.Add("BatchDeleteSecurityGroup", http.MethodDelete, "/security_groups/batch", svc.BatchDeleteSecurityGroup)
	h.Add("BatchUpdateSecurityGroupCommonInfo", http.MethodPatch, "/security_groups/common/info/batch/update",
		svc.BatchUpdateSecurityGroupCommonInfo)

	h.Load(cap.WebService)
}

type securityGroupSvc struct {
	dao dao.Set
}

// BatchCreateSecurityGroup create security group.
func (svc *securityGroupSvc) BatchCreateSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateSecurityGroup[corecloud.TCloudSecurityGroupExtension](vendor, svc, cts)
	case enumor.Aws:
		return batchCreateSecurityGroup[corecloud.AwsSecurityGroupExtension](vendor, svc, cts)
	case enumor.HuaWei:
		return batchCreateSecurityGroup[corecloud.HuaWeiSecurityGroupExtension](vendor, svc, cts)
	case enumor.Azure:
		return batchCreateSecurityGroup[corecloud.AzureSecurityGroupExtension](vendor, svc, cts)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

// BatchUpdateSecurityGroup update security group.
func (svc *securityGroupSvc) BatchUpdateSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchUpdateSecurityGroup[corecloud.TCloudSecurityGroupExtension](cts, svc)
	case enumor.Aws:
		return batchUpdateSecurityGroup[corecloud.AwsSecurityGroupExtension](cts, svc)
	case enumor.HuaWei:
		return batchUpdateSecurityGroup[corecloud.HuaWeiSecurityGroupExtension](cts, svc)
	case enumor.Azure:
		return batchUpdateSecurityGroup[corecloud.AzureSecurityGroupExtension](cts, svc)
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

	details := make([]corecloud.BaseSecurityGroup, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, corecloud.BaseSecurityGroup{
			ID:        one.ID,
			Vendor:    one.Vendor,
			CloudID:   one.CloudID,
			BkBizID:   one.BkBizID,
			Region:    one.Region,
			Name:      one.Name,
			Memo:      one.Memo,
			AccountID: one.AccountID,
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		})
	}

	return &protocloud.SecurityGroupListResult{Details: details}, nil
}

// BatchDeleteSecurityGroup delete security group.
func (svc *securityGroupSvc) BatchDeleteSecurityGroup(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.SecurityGroupBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: []string{"id", "vendor"},
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.SecurityGroup().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
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
		if err := svc.deleteSecurityGroupRule(cts.Kit, txn, listResp.Details); err != nil {
			return nil, err
		}

		delFilter := tools.ContainersExpression("id", delIDs)
		if err := svc.dao.SecurityGroup().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete security group failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *securityGroupSvc) deleteSecurityGroupRule(kt *kit.Kit, txn *sqlx.Tx,
	details []tablecloud.SecurityGroupTable) error {

	vendorSGMap := make(map[enumor.Vendor][]string)
	for _, one := range details {
		if _, exist := vendorSGMap[one.Vendor]; !exist {
			vendorSGMap[one.Vendor] = make([]string, 0)
		}

		vendorSGMap[one.Vendor] = append(vendorSGMap[one.Vendor], one.ID)
	}

	var err error
	for vendor, sgIDs := range vendorSGMap {
		switch vendor {
		case enumor.TCloud:
			err = svc.dao.TCloudSGRule().DeleteWithTx(kt, txn, tools.ContainersExpression("security_group_id", sgIDs))
		case enumor.Aws:
			err = svc.dao.AwsSGRule().DeleteWithTx(kt, txn, tools.ContainersExpression("security_group_id", sgIDs))
		case enumor.HuaWei:
			err = svc.dao.HuaWeiSGRule().DeleteWithTx(kt, txn, tools.ContainersExpression("security_group_id", sgIDs))
		case enumor.Azure:
			err = svc.dao.AzureSGRule().DeleteWithTx(kt, txn, tools.ContainersExpression("security_group_id", sgIDs))
		default:
			return fmt.Errorf("vendor: %s not support", vendor)
		}
		if err != nil {
			return err
		}
	}

	return nil
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

	sgTable, err := getSecurityGroupByID(cts.Kit, id, svc)
	if err != nil {
		return nil, err
	}

	base := convTableToBaseSG(sgTable)
	switch sgTable.Vendor {
	case enumor.TCloud:
		return convertToSGResult[corecloud.TCloudSecurityGroupExtension](base, sgTable.Extension)
	case enumor.Aws:
		return convertToSGResult[corecloud.AwsSecurityGroupExtension](base, sgTable.Extension)
	case enumor.HuaWei:
		return convertToSGResult[corecloud.HuaWeiSecurityGroupExtension](base, sgTable.Extension)
	case enumor.Azure:
		return convertToSGResult[corecloud.AzureSecurityGroupExtension](base, sgTable.Extension)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

func convTableToBaseSG(sgTable *tablecloud.SecurityGroupTable) *corecloud.BaseSecurityGroup {
	return &corecloud.BaseSecurityGroup{
		ID:        sgTable.ID,
		Vendor:    sgTable.Vendor,
		CloudID:   sgTable.CloudID,
		BkBizID:   sgTable.BkBizID,
		Region:    sgTable.Region,
		Name:      sgTable.Name,
		Memo:      sgTable.Memo,
		AccountID: sgTable.AccountID,
		Creator:   sgTable.Creator,
		Reviser:   sgTable.Reviser,
		CreatedAt: sgTable.CreatedAt.String(),
		UpdatedAt: sgTable.UpdatedAt.String(),
	}
}

func convertToSGResult[T corecloud.SecurityGroupExtension](base *corecloud.BaseSecurityGroup,
	extJson tabletype.JsonField) (*corecloud.SecurityGroup[T], error) {

	extension := new(T)
	err := json.UnmarshalFromString(string(extJson), &extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString security group json extension failed, err: %v", err)
	}

	return &corecloud.SecurityGroup[T]{
		BaseSecurityGroup: *base,
		Extension:         extension,
	}, nil
}

func batchUpdateSecurityGroup[T corecloud.SecurityGroupExtension](cts *rest.Contexts, svc *securityGroupSvc) (
	interface{}, error) {

	req := new(protocloud.SecurityGroupBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0, len(req.SecurityGroups))
	for _, one := range req.SecurityGroups {
		ids = append(ids, one.ID)
	}
	extensionMap, err := listSecurityGroupExtension(cts, svc, ids)
	if err != nil {
		return nil, err
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, sg := range req.SecurityGroups {
			update := &tablecloud.SecurityGroupTable{
				BkBizID: sg.BkBizID,
				Name:    sg.Name,
				Memo:    sg.Memo,
				Reviser: cts.Kit.User,
			}

			if sg.Extension != nil {
				extension, exist := extensionMap[sg.ID]
				if !exist {
					continue
				}

				merge, err := json.UpdateMerge(sg.Extension, string(extension))
				if err != nil {
					return nil, fmt.Errorf("json UpdateMerge extension failed, err: %v", err)
				}
				update.Extension = tabletype.JsonField(merge)
			}

			if err := svc.dao.SecurityGroup().UpdateByIDWithTx(cts.Kit, txn, sg.ID, update); err != nil {
				logs.Errorf("update security group by id failed, err: %v, id: %s, rid: %s", err, sg.ID, cts.Kit.Rid)
				return nil, fmt.Errorf("update security group failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func listSecurityGroupExtension(cts *rest.Contexts, svc *securityGroupSvc, ids []string) (
	map[string]tabletype.JsonField, error) {

	opt := &types.ListOption{
		Fields: []string{"id", "extension"},
		Filter: tools.ContainersExpression("id", ids),
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}
	list, err := svc.dao.SecurityGroup().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	result := make(map[string]tabletype.JsonField, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one.Extension
	}

	return result, nil
}

func getSecurityGroupByID(kt *kit.Kit, id string, svc *securityGroupSvc) (*tablecloud.SecurityGroupTable, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.dao.SecurityGroup().List(kt, opt)
	if err != nil {
		logs.Errorf("list security group failed, err: %v, rid: %s", kt.Rid)
		return nil, fmt.Errorf("list security group failed, err: %v", err)
	}

	if len(result.Details) != 1 {
		return nil, errf.New(errf.RecordNotFound, "security group not found")
	}

	return &result.Details[0], nil
}

func batchCreateSecurityGroup[T corecloud.SecurityGroupExtension](vendor enumor.Vendor, svc *securityGroupSvc,
	cts *rest.Contexts) (interface{}, error) {

	req := new(protocloud.SecurityGroupBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		sgs := make([]*tablecloud.SecurityGroupTable, 0, len(req.SecurityGroups))
		for _, sg := range req.SecurityGroups {
			extension, err := json.MarshalToString(sg.Extension)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}

			sgs = append(sgs, &tablecloud.SecurityGroupTable{
				Vendor:    vendor,
				CloudID:   sg.CloudID,
				BkBizID:   sg.BkBizID,
				Region:    sg.Region,
				Name:      sg.Name,
				Memo:      sg.Memo,
				AccountID: sg.AccountID,
				Extension: tabletype.JsonField(extension),
				Creator:   cts.Kit.User,
				Reviser:   cts.Kit.User,
			})
		}

		ids, err := svc.dao.SecurityGroup().BatchCreateWithTx(cts.Kit, txn, sgs)
		if err != nil {
			return nil, fmt.Errorf("create security group failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create security group but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchUpdateSecurityGroupCommonInfo batch update security group common info.
func (svc *securityGroupSvc) BatchUpdateSecurityGroupCommonInfo(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.SecurityGroupCommonInfoBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	updateFilter := tools.ContainersExpression("id", req.IDs)
	updateFiled := &tablecloud.SecurityGroupTable{
		BkBizID: req.BkBizID,
	}
	if err := svc.dao.SecurityGroup().Update(cts.Kit, updateFilter, updateFiled); err != nil {
		return nil, err
	}

	return nil, nil
}

// ListSecurityGroupExt list security group with extension.
func (svc *securityGroupSvc) ListSecurityGroupExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}
	listResp, err := svc.dao.SecurityGroup().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list security group failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return convSecurityGroupExtListResult[corecloud.TCloudSecurityGroupExtension](listResp.Details)
	case enumor.Aws:
		return convSecurityGroupExtListResult[corecloud.AwsSecurityGroupExtension](listResp.Details)
	case enumor.Azure:
		return convSecurityGroupExtListResult[corecloud.AzureSecurityGroupExtension](listResp.Details)
	case enumor.HuaWei:
		return convSecurityGroupExtListResult[corecloud.HuaWeiSecurityGroupExtension](listResp.Details)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

func convSecurityGroupExtListResult[T corecloud.SecurityGroupExtension](tables []tablecloud.SecurityGroupTable) (
	*protocloud.SecurityGroupExtListResult[T], error) {

	details := make([]corecloud.SecurityGroup[T], 0, len(tables))
	for _, one := range tables {
		extension := new(T)
		err := json.UnmarshalFromString(string(one.Extension), &extension)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalFromString security group json extension failed, err: %v", err)
		}

		details = append(details, corecloud.SecurityGroup[T]{
			BaseSecurityGroup: *convTableToBaseSG(&one),
			Extension:         extension,
		})
	}

	return &protocloud.SecurityGroupExtListResult[T]{
		Details: details,
	}, nil
}

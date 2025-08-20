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

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"
)

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

	// 查询使用范围
	sgDetails := result.Details
	var sgBizInfo []types.ResBizInfo
	if len(sgDetails) > 0 {
		sgIDs := slice.Map(sgDetails, tablecloud.SecurityGroupTable.GetID)
		sgBizInfo, err = svc.dao.ResUsageBizRel().ListUsageBizs(cts.Kit, enumor.SecurityGroupCloudResType, sgIDs)
		if err != nil {
			logs.Errorf("fail to get security group usage bizs, err: %v, sg: %v, rid: %s", err, sgIDs, cts.Kit.Rid)
			return nil, fmt.Errorf("fail to get security group usage bizs, err: %w", err)
		}
	}

	details := make([]corecloud.BaseSecurityGroup, 0, len(result.Details))
	for i := range sgDetails {
		sg := sgDetails[i]
		details = append(details, corecloud.BaseSecurityGroup{
			ID:               sg.ID,
			Vendor:           sg.Vendor,
			CloudID:          sg.CloudID,
			BkBizID:          sg.BkBizID,
			Region:           sg.Region,
			Name:             sg.Name,
			Memo:             sg.Memo,
			AccountID:        sg.AccountID,
			MgmtType:         sg.MgmtType,
			MgmtBizID:        sg.MgmtBizID,
			Manager:          sg.Manager,
			BakManager:       sg.BakManager,
			UsageBizIDs:      sgBizInfo[i].BizIDs,
			Creator:          sg.Creator,
			Reviser:          sg.Reviser,
			CreatedAt:        sg.CreatedAt.String(),
			UpdatedAt:        sg.UpdatedAt.String(),
			CloudCreatedTime: sg.CloudCreatedTime,
			CloudUpdateTime:  sg.CloudUpdateTime,
			Tags:             core.TagMap(sg.Tags),
		})
	}

	return &protocloud.SecurityGroupListResult{Details: details}, nil
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

	sgBizInfo, err := svc.dao.ResUsageBizRel().ListUsageBizs(cts.Kit, enumor.SecurityGroupCloudResType,
		[]string{sgTable.ID})
	if err != nil {
		logs.Errorf("fail to get security group usage bizs, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("fail to get security group usage bizs, err: %w", err)
	}

	base := convTableToBaseSG(sgTable, sgBizInfo[0])
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

func convTableToBaseSG(sgTable *tablecloud.SecurityGroupTable, bizInfo types.ResBizInfo) *corecloud.BaseSecurityGroup {
	return &corecloud.BaseSecurityGroup{
		ID:               sgTable.ID,
		Vendor:           sgTable.Vendor,
		CloudID:          sgTable.CloudID,
		BkBizID:          sgTable.BkBizID,
		Region:           sgTable.Region,
		Name:             sgTable.Name,
		Memo:             sgTable.Memo,
		CloudCreatedTime: sgTable.CloudCreatedTime,
		CloudUpdateTime:  sgTable.CloudUpdateTime,
		Tags:             core.TagMap(sgTable.Tags),
		AccountID:        sgTable.AccountID,
		MgmtType:         sgTable.MgmtType,
		MgmtBizID:        sgTable.MgmtBizID,
		Manager:          sgTable.Manager,
		BakManager:       sgTable.BakManager,
		UsageBizIDs:      bizInfo.BizIDs,
		Creator:          sgTable.Creator,
		Reviser:          sgTable.Reviser,
		CreatedAt:        sgTable.CreatedAt.String(),
		UpdatedAt:        sgTable.UpdatedAt.String(),
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

// listSecurityGroupExtension list security group extension by ids.
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

// getSecurityGroupByID get security group by id.
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

	// 查询使用范围
	sgDetails := listResp.Details
	var sgBizInfo []types.ResBizInfo
	if req.Page.Count {
		return &protocloud.SecurityGroupListResult{Count: listResp.Count}, nil
	}
	if len(sgDetails) == 0 {
		return listResp, nil
	}
	sgIDs := slice.Map(sgDetails, tablecloud.SecurityGroupTable.GetID)
	sgBizInfo, err = svc.dao.ResUsageBizRel().ListUsageBizs(cts.Kit, enumor.SecurityGroupCloudResType, sgIDs)
	if err != nil {
		logs.Errorf("fail to get security group usage bizs, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("fail to get security group usage bizs, err: %w", err)
	}
	switch vendor {
	case enumor.TCloud:
		return convSecurityGroupExtListResult[corecloud.TCloudSecurityGroupExtension](sgDetails, sgBizInfo)
	case enumor.Aws:
		return convSecurityGroupExtListResult[corecloud.AwsSecurityGroupExtension](sgDetails, sgBizInfo)
	case enumor.Azure:
		return convSecurityGroupExtListResult[corecloud.AzureSecurityGroupExtension](sgDetails, sgBizInfo)
	case enumor.HuaWei:
		return convSecurityGroupExtListResult[corecloud.HuaWeiSecurityGroupExtension](sgDetails, sgBizInfo)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

func convSecurityGroupExtListResult[T corecloud.SecurityGroupExtension](tables []tablecloud.SecurityGroupTable,
	bizInfos []types.ResBizInfo) (*protocloud.SecurityGroupExtListResult[T], error) {

	details := make([]corecloud.SecurityGroup[T], 0, len(tables))
	for i, one := range tables {
		var extension *T
		if one.Extension != "" {
			extension = new(T)
			err := json.UnmarshalFromString(string(one.Extension), &extension)
			if err != nil {
				return nil, fmt.Errorf("unmarshal security group json extension failed, err: %v", err)
			}
		}

		details = append(details, corecloud.SecurityGroup[T]{
			BaseSecurityGroup: *convTableToBaseSG(&one, bizInfos[i]),
			Extension:         extension,
		})
	}

	return &protocloud.SecurityGroupExtListResult[T]{
		Details: details,
	}, nil
}

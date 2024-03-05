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

package argstpl

import (
	"fmt"

	"hcm/pkg/api/core"
	coreargstpl "hcm/pkg/api/core/cloud/argument-template"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	tableargstpl "hcm/pkg/dal/table/cloud/argument-template"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"
)

// ListArgsTpl list argument template.
func (svc *argsTplSvc) ListArgsTpl(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.ArgsTplListReq)
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
	result, err := svc.dao.ArgsTpl().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list argument template failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list argument template failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.ArgsTplListResult{Count: result.Count}, nil
	}

	details := make([]coreargstpl.BaseArgsTpl, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne, tErr := convTableToBaseArgsTpl(&one)
		if tErr != nil {
			logs.Errorf("list loop argument template detail failed, err: %v, rid: %s", tErr, cts.Kit.Rid)
			continue
		}

		details = append(details, *tmpOne)
	}

	return &protocloud.ArgsTplListResult{Details: details}, nil
}

func convTableToBaseArgsTpl(one *tableargstpl.ArgumentTemplateTable) (*coreargstpl.BaseArgsTpl, error) {
	templates := new([]coreargstpl.TemplateInfo)
	if len(one.Templates) > 0 && one.Templates != "{}" {
		err := json.UnmarshalFromString(string(one.Templates), templates)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalFromString db templates failed, err: %v", err)
		}
	}

	groupTemplates := new([]string)
	if len(one.GroupTemplates) > 0 && one.GroupTemplates != "{}" {
		err := json.UnmarshalFromString(string(one.GroupTemplates), groupTemplates)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalFromString db group templates failed, err: %v", err)
		}
	}

	base := &coreargstpl.BaseArgsTpl{
		ID:             one.ID,
		CloudID:        one.CloudID,
		Name:           one.Name,
		Vendor:         one.Vendor,
		BkBizID:        one.BkBizID,
		AccountID:      one.AccountID,
		Type:           one.Type,
		Templates:      templates,
		GroupTemplates: groupTemplates,
		Memo:           one.Memo,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}

	return base, nil
}

// ListArgsTplExt list argument template with extension.
func (svc *argsTplSvc) ListArgsTplExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(protocloud.ArgsTplListReq)
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
	result, err := svc.dao.ArgsTpl().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list argument template failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list argument template failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.ArgsTplExtListResult[coreargstpl.TCloudArgsTplExtension]{Count: result.Count}, nil
	}

	switch vendor {
	case enumor.TCloud:
		return convArgsTplListResult[coreargstpl.TCloudArgsTplExtension](cts.Kit, result.Details)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

func convArgsTplListResult[T coreargstpl.Extension](kt *kit.Kit, tables []tableargstpl.ArgumentTemplateTable) (
	*protocloud.ArgsTplExtListResult[T], error) {

	details := make([]*coreargstpl.ArgsTpl[T], 0, len(tables))
	for _, one := range tables {
		tmpData, err := convTableToBaseArgsTpl(&one)
		if err != nil {
			logs.Errorf("list loop argument template detail failed, err: %v, rid: %s", err, kt.Rid)
			continue
		}

		extension := new(T)
		details = append(details, &coreargstpl.ArgsTpl[T]{
			BaseArgsTpl: *tmpData,
			Extension:   extension,
		})
	}

	return &protocloud.ArgsTplExtListResult[T]{
		Details: details,
	}, nil
}

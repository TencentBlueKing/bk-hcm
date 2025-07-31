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

package rootaccountbillconfig

import (
	"fmt"

	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
	dsbill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablebill "hcm/pkg/dal/table/bill"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

// ListRootAccountBillConfig list root account bill config.
func (svc *rootBillConfigSvc) ListRootAccountBillConfig(cts *rest.Contexts) (interface{}, error) {
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
	daoResp, err := svc.dao.RootAccountBillConfig().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list root account bill config failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list root bill config failed, err: %v", err)
	}

	if req.Page.Count {
		return &dsbill.RootAccountBillConfigListResult{Count: daoResp.Count}, nil
	}

	details := make([]billcore.BaseRootAccountBillConfig, 0, len(daoResp.Details))
	for _, item := range daoResp.Details {
		details = append(details, converter.PtrToVal(convertBaseAccountBillConfig(&item)))
	}

	return &dsbill.RootAccountBillConfigListResult{Details: details}, nil
}

// ListRootAccountBillConfigExt list bill config ext.
func (svc *rootBillConfigSvc) ListRootAccountBillConfigExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
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
	listResp, err := svc.dao.RootAccountBillConfig().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list bill config extension failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.Aws:
		return cvtAccountBillConfigExtListResult[billcore.AwsBillConfigExtension](listResp.Details)
	case enumor.Gcp:
		return cvtAccountBillConfigExtListResult[billcore.GcpBillConfigExtension](listResp.Details)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

func cvtAccountBillConfigExtListResult[T billcore.RootAccountBillConfigExtension](
	tables []tablebill.RootAccountBillConfigTable) (*dsbill.RootAccountBillConfigExtListResult[T], error) {

	details := make([]billcore.RootAccountBillConfig[T], 0, len(tables))
	for _, one := range tables {
		extension := new(T)
		if one.Extension != "" {
			err := json.UnmarshalFromString(string(one.Extension), &extension)
			if err != nil {
				return nil, fmt.Errorf("UnmarshalFromString account bill json extension failed, err: %+v, "+
					"extension: %s", err, one.Extension)
			}
		}

		details = append(details, billcore.RootAccountBillConfig[T]{
			BaseRootAccountBillConfig: *convertBaseAccountBillConfig(&one),
			Extension:                 extension,
		})
	}

	return &dsbill.RootAccountBillConfigExtListResult[T]{
		Details: details,
	}, nil
}

func convertBaseAccountBillConfig(dbDetail *tablebill.RootAccountBillConfigTable) *billcore.BaseRootAccountBillConfig {
	if dbDetail == nil {
		return nil
	}

	var tmpErrMsg []string
	if dbDetail.ErrMsg != "" {
		err := json.UnmarshalFromString(string(dbDetail.ErrMsg), &tmpErrMsg)
		if err != nil {
			logs.Errorf("convert account bill config errmsg failed, errmsg: %s, err: %v", dbDetail.ErrMsg, err)
			tmpErrMsg = append(tmpErrMsg, string(dbDetail.ErrMsg))
		}
	}

	return &billcore.BaseRootAccountBillConfig{
		ID:                dbDetail.ID,
		Vendor:            dbDetail.Vendor,
		RootAccountID:     dbDetail.RootAccountID,
		CloudDatabaseName: dbDetail.CloudDatabaseName,
		CloudTableName:    dbDetail.CloudTableName,
		ErrMsg:            tmpErrMsg,
		Extension:         dbDetail.Extension,

		Revision: &core.Revision{
			Creator:   dbDetail.Creator,
			Reviser:   dbDetail.Reviser,
			CreatedAt: dbDetail.CreatedAt.String(),
			UpdatedAt: dbDetail.UpdatedAt.String(),
		},
	}
}

// GetRootAccountBillConfig get bill config detail.
func (svc *rootBillConfigSvc) GetRootAccountBillConfig(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	netID := cts.PathParameter("id").String()
	dbDetail, err := getRootAccountBillConfigFromTable(cts.Kit, svc.dao, netID)
	if err != nil {
		return nil, err
	}

	base := convertBaseAccountBillConfig(dbDetail)
	switch vendor {
	case enumor.Aws:
		return cvtToRootAccountBillConfigResult[billcore.AwsBillConfigExtension](base, dbDetail.Extension)
	case enumor.Gcp:
		return cvtToRootAccountBillConfigResult[billcore.GcpBillConfigExtension](base, dbDetail.Extension)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

func getRootAccountBillConfigFromTable(kt *kit.Kit, dao dao.Set, billID string) (
	*tablebill.RootAccountBillConfigTable, error) {

	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", billID),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}
	res, err := dao.RootAccountBillConfig().List(kt, opt)
	if err != nil {
		logs.Errorf("get root account bill config failed, netID: %s, err: %v, rid: %s", billID, kt.Rid)
		return nil, fmt.Errorf("get root account bill config failed, err: %v", err)
	}

	details := res.Details
	if len(details) != 1 {
		return nil, fmt.Errorf("get list root account bill config failed, bill(id=%s) doesn't exist", billID)
	}

	return &details[0], nil
}

func cvtToRootAccountBillConfigResult[T billcore.RootAccountBillConfigExtension](
	baseNI *billcore.BaseRootAccountBillConfig,
	dbExtension tabletype.JsonField) (
	*billcore.RootAccountBillConfig[T], error) {

	extension := new(T)
	err := json.UnmarshalFromString(string(dbExtension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString root account bill config db extension failed, err: %+v", err)
	}

	return &billcore.RootAccountBillConfig[T]{
		BaseRootAccountBillConfig: *baseNI,
		Extension:                 extension,
	}, nil
}

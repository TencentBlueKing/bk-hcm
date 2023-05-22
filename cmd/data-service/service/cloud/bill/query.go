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

package bill

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	dsbill "hcm/pkg/api/data-service/cloud/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablebill "hcm/pkg/dal/table/cloud/bill"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

// ListBillConfig list bill config.
func (svc *billConfigSvc) ListBillConfig(cts *rest.Contexts) (interface{}, error) {
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
	daoResp, err := svc.dao.AccountBillConfig().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list account bill config failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list bill failed, err: %v", err)
	}

	if req.Page.Count {
		return &dsbill.AccountBillConfigListResult{Count: daoResp.Count}, nil
	}

	details := make([]cloud.BaseAccountBillConfig, 0, len(daoResp.Details))
	for _, item := range daoResp.Details {
		details = append(details, converter.PtrToVal(convertBaseAccountBillConfig(&item)))
	}

	return &dsbill.AccountBillConfigListResult{Details: details}, nil
}

// ListBillConfigExt list bill config ext.
func (svc *billConfigSvc) ListBillConfigExt(cts *rest.Contexts) (interface{}, error) {
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
	listResp, err := svc.dao.AccountBillConfig().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list bill config extension failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.Aws:
		return convertAccountBillConfigExtListResult[cloud.AwsBillConfigExtension](listResp.Details)
	case enumor.Gcp:
		return convertAccountBillConfigExtListResult[cloud.GcpBillConfigExtension](listResp.Details)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

func convertAccountBillConfigExtListResult[T cloud.AccountBillConfigExtension](
	tables []tablebill.AccountBillConfigTable) (*dsbill.AccountBillConfigExtListResult[T], error) {

	details := make([]cloud.AccountBillConfig[T], 0, len(tables))
	for _, one := range tables {
		extension := new(T)
		if one.Extension != "" {
			err := json.UnmarshalFromString(string(one.Extension), &extension)
			if err != nil {
				return nil, fmt.Errorf("UnmarshalFromString account bill json extension failed, err: %+v, "+
					"extension: %s", err, one.Extension)
			}
		}

		details = append(details, cloud.AccountBillConfig[T]{
			BaseAccountBillConfig: *convertBaseAccountBillConfig(&one),
			Extension:             extension,
		})
	}

	return &dsbill.AccountBillConfigExtListResult[T]{
		Details: details,
	}, nil
}

func convertBaseAccountBillConfig(dbDetail *tablebill.AccountBillConfigTable) *cloud.BaseAccountBillConfig {
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

	return &cloud.BaseAccountBillConfig{
		ID:                dbDetail.ID,
		Vendor:            dbDetail.Vendor,
		AccountID:         dbDetail.AccountID,
		CloudDatabaseName: dbDetail.CloudDatabaseName,
		CloudTableName:    dbDetail.CloudTableName,
		Status:            dbDetail.Status,
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

// GetBillConfig get bill config detail.
func (svc *billConfigSvc) GetBillConfig(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	netID := cts.PathParameter("id").String()
	dbDetail, err := getAccountBillConfigFromTable(cts.Kit, svc.dao, netID)
	if err != nil {
		return nil, err
	}

	base := convertBaseAccountBillConfig(dbDetail)
	switch vendor {
	case enumor.Aws:
		return convertToAccountBillConfigResult[cloud.AwsBillConfigExtension](base, dbDetail.Extension)
	case enumor.Gcp:
		return convertToAccountBillConfigResult[cloud.GcpBillConfigExtension](base, dbDetail.Extension)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

func getAccountBillConfigFromTable(kt *kit.Kit, dao dao.Set, billID string) (*tablebill.AccountBillConfigTable, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", billID),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}
	res, err := dao.AccountBillConfig().List(kt, opt)
	if err != nil {
		logs.Errorf("get account bill config failed, netID: %s, err: %v, rid: %s", billID, kt.Rid)
		return nil, fmt.Errorf("get account bill config failed, err: %v", err)
	}

	details := res.Details
	if len(details) != 1 {
		return nil, fmt.Errorf("get list account bill config failed, bill(id=%s) doesn't exist", billID)
	}

	return &details[0], nil
}

func convertToAccountBillConfigResult[T cloud.AccountBillConfigExtension](baseNI *cloud.BaseAccountBillConfig,
	dbExtension tabletype.JsonField) (*cloud.AccountBillConfig[T], error) {

	extension := new(T)
	err := json.UnmarshalFromString(string(dbExtension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString account bill config db extension failed, err: %+v", err)
	}

	return &cloud.AccountBillConfig[T]{
		BaseAccountBillConfig: *baseNI,
		Extension:             extension,
	}, nil
}

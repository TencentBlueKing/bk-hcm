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

package eip

import (
	"fmt"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	eipdao "hcm/pkg/dal/dao/cloud/eip"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud/eip"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

type eipSvc struct {
	dao.Set
	objectDao *eipdao.EipDao
}

// Init 注册 eip.EipDao 到 Capability.Dao, 并设置 objectDao
func (svc *eipSvc) Init() {
	d := &eipdao.EipDao{}
	registeredDao := svc.GetObjectDao(d.Name())
	if registeredDao == nil {
		d.ObjectDaoManager = new(dao.ObjectDaoManager)
		svc.RegisterObjectDao(d)
	}

	svc.objectDao = svc.GetObjectDao(d.Name()).(*eipdao.EipDao)
	svc.objectDao.Audit = svc.Audit()
}

// BatchCreateEipExt ...
func (svc *eipSvc) BatchCreateEipExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	switch vendor {
	case enumor.TCloud:
		return batchCreateEipExt[dataproto.TCloudEipExtensionCreateReq](cts, svc, vendor)
	case enumor.Aws:
		return batchCreateEipExt[dataproto.AwsEipExtensionCreateReq](cts, svc, vendor)
	case enumor.Gcp:
		return batchCreateEipExt[dataproto.GcpEipExtensionCreateReq](cts, svc, vendor)
	case enumor.HuaWei:
		return batchCreateEipExt[dataproto.HuaWeiEipExtensionCreateReq](cts, svc, vendor)
	case enumor.Azure:
		return batchCreateEipExt[dataproto.AzureEipExtensionCreateReq](cts, svc, vendor)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// RetrieveEipExt ...
func (svc *eipSvc) RetrieveEipExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	eipID := cts.PathParameter("id").String()
	opt := &types.ListOption{
		Filter: tools.EqualWithOpExpression(
			filter.And,
			map[string]interface{}{"id": eipID, "vendor": string(vendor)},
		),
		Page: &core.BasePage{Count: false, Start: 0, Limit: 1},
	}

	data, err := svc.objectDao.List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	if count := len(data.Details); count != 1 {
		return nil, fmt.Errorf("retrieve eip failed: query id(%s) return total %d", eipID, count)
	}

	eipData := data.Details[0]
	switch vendor {
	case enumor.TCloud:
		return toProtoEipExtResult[dataproto.TCloudEipExtensionResult](eipData)
	case enumor.Aws:
		return toProtoEipExtResult[dataproto.AwsEipExtensionResult](eipData)
	case enumor.Gcp:
		return toProtoEipExtResult[dataproto.GcpEipExtensionResult](eipData)
	case enumor.Azure:
		return toProtoEipExtResult[dataproto.AzureEipExtensionResult](eipData)
	case enumor.HuaWei:
		return toProtoEipExtResult[dataproto.HuaWeiEipExtensionResult](eipData)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// ListEip ...
func (svc *eipSvc) ListEip(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.EipListReq)
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

	data, err := svc.objectDao.List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	details := make([]*dataproto.EipResult, len(data.Details))
	for indx, d := range data.Details {
		details[indx] = toProtoEipResult(d)
	}

	return &dataproto.EipListResult{Details: details, Count: data.Count}, nil
}

// ListEipExt ...
func (svc *eipSvc) ListEipExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(dataproto.EipListReq)
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

	data, err := svc.objectDao.List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return toProtoEipExtListResult[dataproto.TCloudEipExtensionResult](data)
	case enumor.Aws:
		return toProtoEipExtListResult[dataproto.AwsEipExtensionResult](data)
	case enumor.Gcp:
		return toProtoEipExtListResult[dataproto.GcpEipExtensionResult](data)
	case enumor.HuaWei:
		return toProtoEipExtListResult[dataproto.HuaWeiEipExtensionResult](data)
	case enumor.Azure:
		return toProtoEipExtListResult[dataproto.AzureEipExtensionResult](data)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// BatchUpdateEipExt ...
func (svc *eipSvc) BatchUpdateEipExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	switch vendor {
	case enumor.TCloud:
		return batchUpdateEipExt[dataproto.TCloudEipExtensionUpdateReq](cts, svc)
	case enumor.Aws:
		return batchUpdateEipExt[dataproto.AwsEipExtensionUpdateReq](cts, svc)
	case enumor.Gcp:
		return batchUpdateEipExt[dataproto.GcpEipExtensionUpdateReq](cts, svc)
	case enumor.Azure:
		return batchUpdateEipExt[dataproto.AzureEipExtensionUpdateReq](cts, svc)
	case enumor.HuaWei:
		return batchUpdateEipExt[dataproto.HuaWeiEipExtensionUpdateReq](cts, svc)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// BatchUpdateEip ...
func (svc *eipSvc) BatchUpdateEip(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.EipBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	updateData := &tablecloud.EipModel{
		BkBizID: int64(req.BkBizID),
		Status:  req.Status,
	}
	if err := svc.objectDao.Update(cts.Kit, tools.ContainersExpression("id", req.IDs), updateData); err != nil {
		return nil, err
	}
	return nil, nil
}

// BatchDeleteEip ...
func (svc *eipSvc) BatchDeleteEip(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.EipDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		return nil, svc.objectDao.DeleteWithTx(cts.Kit, txn, req.Filter)
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// rawExtensions 根据条件查询原始的 extension 字段, 返回字典结构 {"EIP ID": "原始的 extension 字段"}
// TODO 不同资源可以复用 rawExtensions 逻辑
func (svc *eipSvc) rawExtensions(
	cts *rest.Contexts,
	filterExp *filter.Expression,
) (map[string]tabletype.JsonField, error) {
	opt := &types.ListOption{
		Filter: filterExp,
		Page:   &core.BasePage{Limit: core.DefaultMaxPageLimit},
		Fields: []string{"id", "extension"},
	}
	data, err := svc.objectDao.List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	extensions := make(map[string]tabletype.JsonField)
	for _, d := range data.Details {
		extensions[d.ID] = d.Extension
	}

	return extensions, nil
}

func batchCreateEipExt[T dataproto.EipExtensionCreateReq](
	cts *rest.Contexts,
	svc *eipSvc,
	vendor enumor.Vendor,
) (interface{}, error) {
	req := new(dataproto.EipExtBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	eipIDs, err := svc.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		eips := make([]*tablecloud.EipModel, len(*req))
		for idx, eipReq := range *req {
			extensionJson, err := json.MarshalToString(eipReq.Extension)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}

			bkBizID := int64(constant.UnassignedBiz)
			if eipReq.BkBizID > 0 {
				bkBizID = eipReq.BkBizID
			}
			eips[idx] = &tablecloud.EipModel{
				Vendor:    string(vendor),
				AccountID: eipReq.AccountID,
				CloudID:   eipReq.CloudID,
				BkBizID:   bkBizID,
				Name:      eipReq.Name,
				Region:    eipReq.Region,
				Status:    eipReq.Status,
				PublicIp:  eipReq.PublicIp,
				PrivateIp: eipReq.PrivateIp,
				Extension: tabletype.JsonField(extensionJson),
				Creator:   cts.Kit.User,
				Reviser:   cts.Kit.User,
			}
		}
		return svc.objectDao.BatchCreateWithTx(cts.Kit, txn, eips)
	})
	if err != nil {
		return nil, err
	}

	return &core.BatchCreateResult{IDs: eipIDs.([]string)}, nil
}

func batchUpdateEipExt[T dataproto.EipExtensionUpdateReq](cts *rest.Contexts, svc *eipSvc) (interface{}, error) {
	req := new(dataproto.EipExtBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	queryIDs := make([]string, len(*req))
	for idx, eipReq := range *req {
		queryIDs[idx] = eipReq.ID
	}
	rawExtensions, err := svc.rawExtensions(cts, tools.ContainersExpression("id", queryIDs))
	if err != nil {
		return nil, err
	}

	_, err = svc.Set.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, eipReq := range *req {
			updateData := &tablecloud.EipModel{
				BkBizID: int64(eipReq.BkBizID),
				Status:  eipReq.Status,
			}

			if eipReq.Extension != nil {
				rawExtension, exist := rawExtensions[eipReq.ID]
				if !exist {
					return nil, fmt.Errorf("eip id (%s) not exit", eipReq.ID)
				}
				mergedExtension, err := json.UpdateMerge(eipReq.Extension, string(rawExtension))
				if err != nil {
					return nil, fmt.Errorf("eip id (%s) merge extension failed, err: %v", eipReq.ID, err)
				}
				updateData.Extension = tabletype.JsonField(mergedExtension)
			}

			if err := svc.objectDao.UpdateByIDWithTx(cts.Kit, txn, eipReq.ID, updateData); err != nil {
				return nil, fmt.Errorf("update eip failed, err: %v", err)
			}
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

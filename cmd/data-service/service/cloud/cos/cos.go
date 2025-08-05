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

// Package cos 提供cos相关接口
package cos

import (
	"fmt"
	"net/http"
	"reflect"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	corecos "hcm/pkg/api/core/cloud/cos"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	daotypes "hcm/pkg/dal/dao/types"
	tablecos "hcm/pkg/dal/table/cloud/cos"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

type cosSvc struct {
	dao dao.Set
}

var svc *cosSvc

// InitSvc init cos service
func InitSvc(cap *capability.Capability) {
	svc = &cosSvc{dao: cap.Dao}
	h := rest.NewHandler()
	h.Add("ListBuckets", http.MethodGet, "/vendors/{vendor}/buckets", svc.ListCos)
	h.Add("GetBucket", http.MethodGet, "/vendors/{vendor}/bucket/{id}", svc.GetCos)
	h.Add("BatchCreateBucket", http.MethodPost, "/vendors/{vendor}/buckets", svc.BatchCreateCos)
	h.Add("BatchUpdateBucket", http.MethodPut, "/vendors/{vendor}/bucket/{id}", svc.BatchUpdateCos)
	h.Add("BatchDeleteBucket", http.MethodDelete, "/vendors/{vendor}/bucket/{id}", svc.BatchDeleteCos)

	h.Load(cap.WebService)
}

// ListCos list cos buckets.
func (svc *cosSvc) ListCos(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &daotypes.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.Cos().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list cos failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, fmt.Errorf("list cos failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.CosListResult{Count: result.Count}, nil
	}

	details := make([]corecos.BaseCos, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne := convTableToBaseCos(&one)
		details = append(details, *tmpOne)
	}

	return &protocloud.CosListResult{Details: details}, nil
}

// GetCos get cos bucket.
func (svc *cosSvc) GetCos(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "cos id is required")
	}

	opt := &daotypes.ListOption{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.dao.Cos().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list cos(%s) failed, err: %v, rid: %s", err, id, cts.Kit.Rid)
		return nil, fmt.Errorf("list cos failed, err: %v", err)
	}

	if len(result.Details) != 1 {
		return nil, errf.New(errf.RecordNotFound, "cos not found")
	}

	cosTable := result.Details[0]
	switch cosTable.Vendor {
	case enumor.TCloud:
		return convCosWithExt[corecos.CosExtension](&cosTable)
	default:
		return nil, fmt.Errorf("unsupport vendor: %s", vendor)
	}
}

// BatchCreateCos batch create cos buckets.
func (svc *cosSvc) BatchCreateCos(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateCos[corecos.CosExtension](cts, svc, vendor)
	default:
		return nil, errf.New(errf.InvalidParameter, "unsupported vendor: "+string(vendor))
	}
}

func batchCreateCos[T corecos.CosExtension](cts *rest.Contexts, svc *cosSvc, vendor enumor.Vendor) (any, error) {
	req := new(protocloud.CosBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		models := make([]*tablecos.CosTable, 0, len(req.Cos))
		for _, cos := range req.Cos {
			cosTable, err := convCosReqToTable(cts.Kit, vendor, cos)
			if err != nil {
				return nil, err
			}
			models = append(models, cosTable)
		}

		ids, err := svc.dao.Cos().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("[%s]fail to batch create cos, err: %v, rid:%s", vendor, err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create cos failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create ccos but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchUpdateCos batch update cos buckets.
func (svc *cosSvc) BatchUpdateCos(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchUpdateCos[corecos.CosExtension](cts, svc)
	default:
		return nil, errf.New(errf.InvalidParameter, "unsupported vendor: "+string(vendor))
	}
}

func batchUpdateCos[T corecos.CosExtension](cts *rest.Contexts, svc *cosSvc) (any, error) {
	req := new(protocloud.CosExtBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cosIds := slice.Map(req.Cos, func(one *protocloud.CosExtUpdateReq[T]) string { return one.ID })

	extensionMap, err := svc.listCosExt(cts.Kit, cosIds)
	if err != nil {
		return nil, err
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		for _, cos := range req.Cos {
			update := &tablecos.CosTable{
				Name:    cos.Name,
				BkBizID: cos.BkBizID,
				Domain:  cos.Domain,
				Status:  cos.Status,

				ACL:                       cos.ACL,
				GrantFullControl:          cos.GrantFullControl,
				GrantRead:                 cos.GrantRead,
				GrantWrite:                cos.GrantWrite,
				GrantReadACP:              cos.GrantReadACP,
				GrantWriteACP:             cos.GrantWriteACP,
				CreateBucketConfiguration: cos.CreateBucketConfiguration,

				CloudCreatedTime: cos.CloudCreatedTime,
				CloudStatusTime:  cos.CloudStatusTime,
				CloudExpiredTime: cos.CloudExpiredTime,
				SyncTime:         cos.SyncTime,
				Tags:             types.StringMap(cos.Tags),
				Reviser:          cts.Kit.User,
			}

			if cos.Extension != nil {
				extension, exist := extensionMap[cos.ID]
				if !exist {
					continue
				}

				merge, err := json.UpdateMerge(cos.Extension, string(extension))
				if err != nil {
					return nil, fmt.Errorf("json UpdateMerge extension failed, err: %v", err)
				}
				update.Extension = types.JsonField(merge)
			}

			if err := svc.dao.Cos().UpdateByIDWithTx(cts.Kit, txn, cos.ID, update); err != nil {
				logs.Errorf("update cos by id failed, err: %v, id: %s, rid: %s", err, cos.ID, cts.Kit.Rid)
				return nil, fmt.Errorf("update cos failed, err: %v", err)
			}
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (svc *cosSvc) listCosExt(kt *kit.Kit, ids []string) (map[string]types.JsonField, error) {
	opt := &daotypes.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   &core.BasePage{Limit: core.DefaultMaxPageLimit},
	}

	resp, err := svc.dao.Cos().List(kt, opt)
	if err != nil {
		return nil, err
	}

	return converter.SliceToMap(resp.Details, func(t tablecos.CosTable) (string, types.JsonField) {
		return t.ID, t.Extension
	}), nil

}

// BatchDeleteCos batch delete cos buckets.
func (svc *cosSvc) BatchDeleteCos(cts *rest.Contexts) (any, error) {
	req := new(protocloud.CosBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &daotypes.ListOption{
		Fields: []string{"id", "vendor", "cloud_id", "bk_biz_id"},
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.Cos().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list cos failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list cos failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	cosIds := slice.Map(listResp.Details, func(one tablecos.CosTable) string { return one.ID })

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		delFilter := tools.ContainersExpression("id", cosIds)
		return nil, svc.dao.Cos().DeleteWithTx(cts.Kit, txn, delFilter)
	})
	if err != nil {
		logs.Errorf("delete cos(ids=%v) failed, err: %v, rid: %s", cosIds, err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

func convCosWithExt[T corecos.CosExtension](tableCos *tablecos.CosTable) (*corecos.Cos[T], error) {
	base := convTableToBaseCos(tableCos)
	extension := new(T)
	if tableCos.Extension != "" {
		if err := json.UnmarshalFromString(string(tableCos.Extension), extension); err != nil {
			return nil, fmt.Errorf("fail unmarshal cos extension, err: %v", err)
		}
	}
	return &corecos.Cos[T]{
		BaseCos:   base,
		Extension: extension,
	}, nil
}

func convTableToBaseCos(one *tablecos.CosTable) *corecos.BaseCos {

	return &corecos.BaseCos{
		ID:        one.ID,
		CloudID:   one.CloudID,
		Name:      one.Name,
		Vendor:    one.Vendor,
		AccountID: one.AccountID,
		BkBizID:   one.BkBizID,
		Region:    one.Region,

		ACL:                       one.ACL,
		GrantFullControl:          one.GrantFullControl,
		GrantRead:                 one.GrantRead,
		GrantWrite:                one.GrantWrite,
		GrantReadACP:              one.GrantReadACP,
		GrantWriteACP:             one.GrantWriteACP,
		CreateBucketConfiguration: one.CreateBucketConfiguration,

		Domain:           one.Domain,
		Status:           one.Status,
		CloudCreatedTime: one.CloudCreatedTime,
		CloudStatusTime:  one.CloudStatusTime,
		CloudExpiredTime: one.CloudExpiredTime,
		SyncTime:         one.SyncTime,
		Tags:             core.TagMap(one.Tags),
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}
}

func convCosReqToTable[T corecos.CosExtension](kt *kit.Kit, vendor enumor.Vendor, cos protocloud.CosCreate[T]) (
	*tablecos.CosTable, error) {
	extension, err := json.MarshalToString(cos.Extension)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return &tablecos.CosTable{
		CloudID:   cos.CloudID,
		Name:      cos.Name,
		Vendor:    vendor,
		AccountID: cos.AccountID,
		BkBizID:   cos.BkBizID,
		Region:    cos.Region,

		Domain:           cos.Domain,
		Status:           cos.Status,
		CloudCreatedTime: cos.CloudCreatedTime,
		CloudStatusTime:  cos.CloudStatusTime,
		CloudExpiredTime: cos.CloudExpiredTime,
		SyncTime:         cos.SyncTime,

		Extension: types.JsonField(extension),
		Tags:      types.StringMap(cos.Tags),

		Creator: kt.User,
		Reviser: kt.User,
	}, nil
}

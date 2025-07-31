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
	"reflect"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/cloud"
	dataservice "hcm/pkg/api/data-service"
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
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// InitVpcService initialize the vpc service.
func InitVpcService(cap *capability.Capability) {
	svc := &vpcSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("BatchCreateVpc", "POST", "/vendors/{vendor}/vpcs/batch/create", svc.BatchCreateVpc)
	h.Add("BatchUpdateVpc", "PATCH", "/vendors/{vendor}/vpcs/batch", svc.BatchUpdateVpc)
	h.Add("BatchUpdateVpcBaseInfo", "PATCH", "/vpcs/base/batch", svc.BatchUpdateVpcBaseInfo)
	h.Add("GetVpc", "GET", "/vendors/{vendor}/vpcs/{id}", svc.GetVpc)
	h.Add("ListVpc", "POST", "/vpcs/list", svc.ListVpc)
	h.Add("ListVpcExt", "POST", "/vendors/{vendor}/vpcs/list", svc.ListVpcExt)
	h.Add("DeleteVpc", "DELETE", "/vpcs/batch", svc.BatchDeleteVpc)

	h.Load(cap.WebService)
}

type vpcSvc struct {
	dao dao.Set
}

// BatchCreateVpc batch create vpc.
func (svc *vpcSvc) BatchCreateVpc(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateVpc[protocloud.TCloudVpcCreateExt](cts, vendor, svc)
	case enumor.Aws:
		return batchCreateVpc[protocloud.AwsVpcCreateExt](cts, vendor, svc)
	case enumor.Gcp:
		return batchCreateVpc[protocloud.GcpVpcCreateExt](cts, vendor, svc)
	case enumor.HuaWei:
		return batchCreateVpc[protocloud.HuaWeiVpcCreateExt](cts, vendor, svc)
	case enumor.Azure:
		return batchCreateVpc[protocloud.AzureVpcCreateExt](cts, vendor, svc)
	}

	return nil, nil
}

// batchCreateVpc batch create vpc.
func batchCreateVpc[T protocloud.VpcCreateExtension](cts *rest.Contexts, vendor enumor.Vendor, svc *vpcSvc) (
	interface{}, error) {

	req := new(protocloud.VpcBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vpcIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		vpcs := make([]tablecloud.VpcTable, 0, len(req.Vpcs))
		for _, createReq := range req.Vpcs {
			ext, err := tabletype.NewJsonField(createReq.Extension)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}

			vpc := tablecloud.VpcTable{
				Vendor:    vendor,
				AccountID: createReq.AccountID,
				CloudID:   createReq.CloudID,
				Name:      createReq.Name,
				Region:    createReq.Region,
				Category:  createReq.Category,
				Memo:      createReq.Memo,
				Extension: ext,
				BkBizID:   createReq.BkBizID,
				Creator:   cts.Kit.User,
				Reviser:   cts.Kit.User,
			}

			vpcs = append(vpcs, vpc)
		}

		vpcID, err := svc.dao.Vpc().BatchCreateWithTx(cts.Kit, txn, vpcs)
		if err != nil {
			return nil, fmt.Errorf("create vpc failed, err: %v", err)
		}

		return vpcID, nil
	})

	if err != nil {
		return nil, err
	}

	ids, ok := vpcIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create vpc but return ids type %s is not string array",
			reflect.TypeOf(vpcIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchUpdateVpc batch update vpc.
func (svc *vpcSvc) BatchUpdateVpc(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchUpdateVpc[protocloud.TCloudVpcUpdateExt](cts, svc)
	case enumor.Aws:
		return batchUpdateVpc[protocloud.AwsVpcUpdateExt](cts, svc)
	case enumor.Gcp:
		return batchUpdateVpc[protocloud.GcpVpcUpdateExt](cts, svc)
	case enumor.HuaWei:
		return batchUpdateVpc[protocloud.HuaWeiVpcUpdateExt](cts, svc)
	case enumor.Azure:
		return batchUpdateVpc[protocloud.AzureVpcUpdateExt](cts, svc)
	}

	return nil, nil
}

// batchUpdateVpc batch update vpc.
func batchUpdateVpc[T protocloud.VpcUpdateExtension](cts *rest.Contexts, svc *vpcSvc) (
	interface{}, error) {
	req := new(protocloud.VpcBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0, len(req.Vpcs))
	for _, vpc := range req.Vpcs {
		ids = append(ids, vpc.ID)
	}

	// check if all vpcs exists
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   &core.BasePage{Count: true},
	}
	listRes, err := svc.dao.Vpc().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list vpc failed, err: %v", err)
	}
	if listRes.Count != uint64(len(req.Vpcs)) {
		return nil, fmt.Errorf("list vpc failed, some vpc(ids=%+v) doesn't exist", ids)
	}

	// update vpc
	vpc := &tablecloud.VpcTable{
		Reviser: cts.Kit.User,
	}

	for _, updateReq := range req.Vpcs {
		vpc.Name = updateReq.Name
		vpc.Category = updateReq.Category
		vpc.Memo = updateReq.Memo
		vpc.BkBizID = updateReq.BkBizID

		// update extension
		if updateReq.Extension != nil {
			// TODO get vpcs before for expression
			dbVpc, err := getVpcFromTable(cts.Kit, svc.dao, updateReq.ID)
			if err != nil {
				return nil, err
			}

			updatedExtension, err := json.UpdateMerge(updateReq.Extension, string(dbVpc.Extension))
			if err != nil {
				return nil, fmt.Errorf("extension update merge failed, err: %v", err)
			}

			vpc.Extension = tabletype.JsonField(updatedExtension)
		}

		err = svc.updateVpc(cts.Kit, []string{updateReq.ID}, vpc)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (svc *vpcSvc) updateVpc(kt *kit.Kit, ids []string, vpc *tablecloud.VpcTable) error {
	if len(ids) == 0 || vpc == nil {
		return errf.New(errf.InvalidParameter, "update vpc ids or update data is not set")
	}

	// update vpc
	err := svc.dao.Vpc().Update(kt, tools.ContainersExpression("id", ids), vpc)
	if err != nil {
		logs.Errorf("update vpc failed, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("update vpc failed, err: %v", err)
	}

	return nil
}

// BatchUpdateVpcBaseInfo batch update vpc base info.
func (svc *vpcSvc) BatchUpdateVpcBaseInfo(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.VpcBaseInfoBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0)
	for _, vpc := range req.Vpcs {
		ids = append(ids, vpc.IDs...)
	}

	// check if all vpcs exists
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   &core.BasePage{Count: true},
	}
	listRes, err := svc.dao.Vpc().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list vpc failed, err: %v", err)
	}
	if listRes.Count != uint64(len(ids)) {
		return nil, fmt.Errorf("list vpc failed, some vpc(ids=%+v) doesn't exist", ids)
	}

	// update vpc
	vpc := &tablecloud.VpcTable{
		Reviser: cts.Kit.User,
	}

	for _, updateReq := range req.Vpcs {
		vpc.Name = updateReq.Data.Name
		vpc.Category = updateReq.Data.Category
		vpc.Memo = updateReq.Data.Memo
		vpc.BkBizID = updateReq.Data.BkBizID

		err = svc.updateVpc(cts.Kit, updateReq.IDs, vpc)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// GetVpc get vpc details.
func (svc *vpcSvc) GetVpc(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vpcID := cts.PathParameter("id").String()

	dbVpc, err := getVpcFromTable(cts.Kit, svc.dao, vpcID)
	if err != nil {
		return nil, err
	}

	base := convertBaseVpc(dbVpc)

	switch vendor {
	case enumor.TCloud:
		return convertToVpcResult[protocore.TCloudVpcExtension](base, dbVpc.Extension)
	case enumor.Aws:
		return convertToVpcResult[protocore.AwsVpcExtension](base, dbVpc.Extension)
	case enumor.Gcp:
		return convertToVpcResult[protocore.GcpVpcExtension](base, dbVpc.Extension)
	case enumor.HuaWei:
		return convertToVpcResult[protocore.HuaWeiVpcExtension](base, dbVpc.Extension)
	case enumor.Azure:
		return convertToVpcResult[protocore.AzureVpcExtension](base, dbVpc.Extension)
	}

	return nil, nil
}

// convertToVpcResult converts the base VPC and its extension from the database into a VPC result.
func convertToVpcResult[T protocore.VpcExtension](baseVpc *protocore.BaseVpc, dbExtension tabletype.JsonField) (
	*protocore.Vpc[T], error) {

	extension := new(T)
	err := json.UnmarshalFromString(string(dbExtension), extension)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString db extension failed, err: %v", err)
	}
	return &protocore.Vpc[T]{
		BaseVpc:   *baseVpc,
		Extension: extension,
	}, nil
}

// getVpcFromTable retrieves a VPC from the database by its ID.
func getVpcFromTable(kt *kit.Kit, dao dao.Set, vpcID string) (*tablecloud.VpcTable, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", vpcID),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	}
	res, err := dao.Vpc().List(kt, opt)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, rid: %s", kt.Rid)
		return nil, fmt.Errorf("list vpc failed, err: %v", err)
	}

	details := res.Details
	if len(details) != 1 {
		return nil, fmt.Errorf("list vpc failed, vpc(id=%s) doesn't exist", vpcID)
	}

	return &details[0], nil
}

// ListVpc list vpcs.
func (svc *vpcSvc) ListVpc(cts *rest.Contexts) (interface{}, error) {
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
	daoVpcResp, err := svc.dao.Vpc().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list vpc failed, err: %v", err)
	}
	if req.Page.Count {
		return &protocloud.VpcListResult{Count: daoVpcResp.Count}, nil
	}

	details := make([]protocore.BaseVpc, 0, len(daoVpcResp.Details))
	for _, vpc := range daoVpcResp.Details {
		details = append(details, converter.PtrToVal(convertBaseVpc(&vpc)))
	}

	return &protocloud.VpcListResult{Details: details}, nil
}

func convertBaseVpc(dbVpc *tablecloud.VpcTable) *protocore.BaseVpc {
	if dbVpc == nil {
		return nil
	}

	return &protocore.BaseVpc{
		ID:        dbVpc.ID,
		Vendor:    dbVpc.Vendor,
		AccountID: dbVpc.AccountID,
		CloudID:   dbVpc.CloudID,
		Name:      converter.PtrToVal(dbVpc.Name),
		Region:    dbVpc.Region,
		Category:  dbVpc.Category,
		Memo:      dbVpc.Memo,
		BkBizID:   dbVpc.BkBizID,
		Revision: &core.Revision{
			Creator:   dbVpc.Creator,
			Reviser:   dbVpc.Reviser,
			CreatedAt: dbVpc.CreatedAt.String(),
			UpdatedAt: dbVpc.UpdatedAt.String(),
		},
	}
}

// BatchDeleteVpc batch delete vpcs.
func (svc *vpcSvc) BatchDeleteVpc(cts *rest.Contexts) (interface{}, error) {
	req := new(dataservice.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}
	listResp, err := svc.dao.Vpc().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list vpc failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delVpcIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delVpcIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delVpcFilter := tools.ContainersExpression("id", delVpcIDs)
		if err := svc.dao.Vpc().BatchDeleteWithTx(cts.Kit, txn, delVpcFilter); err != nil {
			return nil, err
		}

		// delete vpc relations
		delSubnetFilter := tools.ContainersExpression("vpc_id", delVpcIDs)
		if err := svc.dao.Subnet().BatchDeleteWithTx(cts.Kit, txn, delSubnetFilter); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete vpc failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListVpcExt ...
func (svc *vpcSvc) ListVpcExt(cts *rest.Contexts) (interface{}, error) {
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
	listResp, err := svc.dao.Vpc().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return conVpcExtListResult[protocore.TCloudVpcExtension](listResp.Details)
	case enumor.Aws:
		return conVpcExtListResult[protocore.AwsVpcExtension](listResp.Details)
	case enumor.Azure:
		return conVpcExtListResult[protocore.AzureVpcExtension](listResp.Details)
	case enumor.HuaWei:
		return conVpcExtListResult[protocore.HuaWeiVpcExtension](listResp.Details)
	case enumor.Gcp:
		return conVpcExtListResult[protocore.GcpVpcExtension](listResp.Details)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

// conVpcExtListResult converts the list of VPC tables to a VPC extension list result.
func conVpcExtListResult[T protocore.VpcExtension](tables []tablecloud.VpcTable) (
	*protocloud.VpcExtListResult[T], error) {

	details := make([]protocore.Vpc[T], 0, len(tables))
	for _, one := range tables {
		extension := new(T)
		if len(one.Extension) != 0 {
			err := json.UnmarshalFromString(string(one.Extension), &extension)
			if err != nil {
				return nil, fmt.Errorf("UnmarshalFromString vpc json extension failed, err: %v", err)
			}
		}

		details = append(details, protocore.Vpc[T]{
			BaseVpc:   *convertBaseVpc(&one),
			Extension: extension,
		})
	}

	return &protocloud.VpcExtListResult[T]{
		Details: details,
	}, nil
}

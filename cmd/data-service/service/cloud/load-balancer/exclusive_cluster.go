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

package loadbalancer

import (
	rawjson "encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	daotypes "hcm/pkg/dal/dao/types"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// InitExclusiveClusterService initial the load balancer exclusive cluster service.
func InitExclusiveClusterService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("ListExclusiveCluster", http.MethodPost, "/load_balancer_exclusive_clusters/list", svc.ListExclusiveCluster)
	h.Add("BatchCreateExclusiveCluster", http.MethodPost,
		"/vendors/{vendor}/load_balancer_exclusive_clusters/batch/create", svc.BatchCreateExclusiveCluster)
	h.Add("BatchUpdateExclusiveCluster", http.MethodPatch, "/vendors/{vendor}/load_balancer_exclusive_clusters",
		svc.BatchUpdateExclusiveCluster)
	h.Add("BatchUpdateExclusiveClusterBizID", http.MethodPatch, "/load_balancer_exclusive_clusters/biz",
		svc.BatchUpdateExclusiveClusterBizID)
	h.Add("BatchDeleteExclusiveCluster", http.MethodDelete, "/load_balancer_exclusive_clusters/batch",
		svc.BatchDeleteExclusiveCluster)

	h.Load(cap.WebService)
}

// ListExclusiveCluster list load balancer exclusive cluster, extension is returned as raw json since this
// interface is not vendor-scoped.
func (svc *lbSvc) ListExclusiveCluster(cts *rest.Contexts) (any, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &daotypes.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.LoadBalancerExclusiveCluster().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list load balancer exclusive cluster failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list load balancer exclusive cluster failed, err: %v", err)
	}

	if req.Page.Count {
		return &dataproto.ExclusiveClusterListResult{Count: result.Count}, nil
	}

	details := make([]corelb.ExclusiveClusterRaw, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, corelb.ExclusiveClusterRaw{
			BaseExclusiveCluster: *convExclusiveClusterTableToBase(&one),
			Extension:            convExclusiveClusterExtensionToRaw(one.Extension),
		})
	}

	return &dataproto.ExclusiveClusterListResult{Details: details}, nil
}

// convExclusiveClusterExtensionToRaw converts the extension column to raw json.
// When fields does not include extension, the column is not selected and stays empty.
// An empty string is not valid json and fails response encoding, so fall back to {}.
func convExclusiveClusterExtensionToRaw(ext tabletype.JsonField) rawjson.RawMessage {
	if len(ext) == 0 {
		return rawjson.RawMessage("{}")
	}
	return rawjson.RawMessage(ext)
}

// convExclusiveClusterTableToBase convert LoadBalancerExclusiveClusterTable to BaseExclusiveCluster.
func convExclusiveClusterTableToBase(one *tablelb.LoadBalancerExclusiveClusterTable) *corelb.BaseExclusiveCluster {
	return &corelb.BaseExclusiveCluster{
		ID:               one.ID,
		CloudID:          one.CloudID,
		Name:             one.Name,
		Vendor:           one.Vendor,
		AccountID:        one.AccountID,
		BkBizID:          one.BkBizID,
		Region:           one.Region,
		Zone:             one.Zone,
		ClusterType:      one.ClusterType,
		ClusterTag:       one.ClusterTag,
		Network:          one.Network,
		Isp:              one.Isp,
		Egress:           one.Egress,
		IPVersion:        one.IPVersion,
		MaxConn:          one.MaxConn,
		ClbResourceCount: one.ClbResourceCount,
		Memo:             one.Memo,
		Creator:          one.Creator,
		Reviser:          one.Reviser,
		CreatedAt:        one.CreatedAt.String(),
		UpdatedAt:        one.UpdatedAt.String(),
	}
}

// BatchCreateExclusiveCluster batch create load balancer exclusive cluster, bk_biz_id is fixed to
// constant.UnassignedBiz, business assignment is done by BatchUpdateExclusiveClusterBizID.
func (svc *lbSvc) BatchCreateExclusiveCluster(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateExclusiveCluster[corelb.TCloudExclusiveClusterExtension](cts, svc, vendor)
	default:
		return nil, errf.New(errf.InvalidParameter, "unsupported vendor: "+string(vendor))
	}
}

func batchCreateExclusiveCluster[T corelb.ExclusiveClusterExtension](cts *rest.Contexts, svc *lbSvc,
	vendor enumor.Vendor) (any, error) {

	req := new(dataproto.ExclusiveClusterBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		models := make([]tablelb.LoadBalancerExclusiveClusterTable, 0, len(req.Clusters))
		for _, cluster := range req.Clusters {
			model, err := convExclusiveClusterCreateToTable(cts.Kit, vendor, cluster)
			if err != nil {
				return nil, err
			}
			models = append(models, *model)
		}

		ids, err := svc.dao.LoadBalancerExclusiveCluster().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("[%s]fail to batch create load balancer exclusive cluster, err: %v, rid: %s", vendor,
				err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create load balancer exclusive cluster failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create exclusive cluster but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

func convExclusiveClusterCreateToTable[T corelb.ExclusiveClusterExtension](kt *kit.Kit, vendor enumor.Vendor,
	cluster dataproto.ExclusiveClusterCreate[T]) (*tablelb.LoadBalancerExclusiveClusterTable, error) {

	extension, err := json.MarshalToString(cluster.Extension)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return &tablelb.LoadBalancerExclusiveClusterTable{
		CloudID:          cluster.CloudID,
		Name:             cluster.Name,
		Vendor:           vendor,
		AccountID:        cluster.AccountID,
		BkBizID:          constant.UnassignedBiz,
		Region:           cluster.Region,
		Zone:             cluster.Zone,
		ClusterType:      cluster.ClusterType,
		ClusterTag:       cluster.ClusterTag,
		Network:          cluster.Network,
		Isp:              cluster.Isp,
		Egress:           cluster.Egress,
		IPVersion:        cluster.IPVersion,
		MaxConn:          cluster.MaxConn,
		ClbResourceCount: cluster.ClbResourceCount,
		Extension:        tabletype.JsonField(extension),
		Memo:             cluster.Memo,
		Creator:          kt.User,
		Reviser:          kt.User,
	}, nil
}

// BatchUpdateExclusiveCluster batch update load balancer exclusive cluster's cloud attributes, bk_biz_id is not
// touched by this interface, keep its zero value on the model so it is skipped by RearrangeSQLDataWithOption.
func (svc *lbSvc) BatchUpdateExclusiveCluster(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchUpdateExclusiveCluster[corelb.TCloudExclusiveClusterExtension](cts, svc)
	default:
		return nil, errf.New(errf.InvalidParameter, "unsupported vendor: "+string(vendor))
	}
}

func batchUpdateExclusiveCluster[T corelb.ExclusiveClusterExtension](cts *rest.Contexts, svc *lbSvc) (any, error) {
	req := new(dataproto.ExclusiveClusterBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		models := make([]tablelb.LoadBalancerExclusiveClusterTable, 0, len(req.Clusters))
		for _, cluster := range req.Clusters {
			model, err := convExclusiveClusterUpdateToTable(cts.Kit, cluster)
			if err != nil {
				return nil, err
			}
			models = append(models, *model)
		}

		if err := svc.dao.LoadBalancerExclusiveCluster().BatchUpdateWithTx(cts.Kit, txn, models); err != nil {
			logs.Errorf("batch update load balancer exclusive cluster failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch update load balancer exclusive cluster failed, err: %v", err)
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func convExclusiveClusterUpdateToTable[T corelb.ExclusiveClusterExtension](kt *kit.Kit,
	cluster dataproto.ExclusiveClusterUpdate[T]) (*tablelb.LoadBalancerExclusiveClusterTable, error) {

	model := &tablelb.LoadBalancerExclusiveClusterTable{
		ID:               cluster.ID,
		Name:             cluster.Name,
		Zone:             cluster.Zone,
		ClusterType:      cluster.ClusterType,
		ClusterTag:       cluster.ClusterTag,
		Network:          cluster.Network,
		Isp:              cluster.Isp,
		Egress:           cluster.Egress,
		IPVersion:        cluster.IPVersion,
		MaxConn:          cluster.MaxConn,
		ClbResourceCount: cluster.ClbResourceCount,
		Memo:             cluster.Memo,
		Reviser:          kt.User,
	}

	if cluster.Extension != nil {
		extension, err := json.MarshalToString(cluster.Extension)
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
		model.Extension = tabletype.JsonField(extension)
	}

	return model, nil
}

// BatchUpdateExclusiveClusterBizID batch assign load balancer exclusive cluster to a business. This interface
// does NOT do "whether already assigned" business pre-check, authentication or audit, it is only supposed to be
// called internally by cloud-server, and it reuses the same DAO BatchUpdateWithTx as cloud attribute update,
// isolated purely by the narrow field set (id/bk_biz_id/reviser) of the model built here.
func (svc *lbSvc) BatchUpdateExclusiveClusterBizID(cts *rest.Contexts) (any, error) {
	req := new(dataproto.ExclusiveClusterBatchUpdateBizIDReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		models := make([]tablelb.LoadBalancerExclusiveClusterTable, 0, len(req.ClusterIDs))
		for _, id := range req.ClusterIDs {
			models = append(models, tablelb.LoadBalancerExclusiveClusterTable{
				ID:      id,
				BkBizID: req.BkBizID,
				Reviser: cts.Kit.User,
			})
		}

		if err := svc.dao.LoadBalancerExclusiveCluster().BatchUpdateWithTx(cts.Kit, txn, models); err != nil {
			logs.Errorf("batch update load balancer exclusive cluster bk_biz_id failed, err: %v, rid: %s",
				err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch update load balancer exclusive cluster bk_biz_id failed, err: %v", err)
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// BatchDeleteExclusiveCluster batch delete load balancer exclusive cluster.
func (svc *lbSvc) BatchDeleteExclusiveCluster(cts *rest.Contexts) (any, error) {
	req := new(dataproto.ExclusiveClusterBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		return nil, svc.dao.LoadBalancerExclusiveCluster().BatchDeleteWithTx(cts.Kit, txn, req.Filter)
	})
	if err != nil {
		logs.Errorf("delete load balancer exclusive cluster failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

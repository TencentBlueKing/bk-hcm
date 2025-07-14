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

package networkinterface

import (
	"fmt"
	"reflect"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/api/data-service"
	datacloudniproto "hcm/pkg/api/data-service/cloud/network-interface"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableni "hcm/pkg/dal/table/cloud/network-interface"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// InitNetInterfaceService initial the network interface service
func InitNetInterfaceService(cap *capability.Capability) {
	svc := &NetworkInterfaceSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("BatchCreateNetworkInterface", "POST", "/vendors/{vendor}/network_interfaces/batch/create",
		svc.BatchCreateNetworkInterface)
	h.Add("BatchUpdateNetworkInterface", "PATCH", "/vendors/{vendor}/network_interfaces/batch",
		svc.BatchUpdateNetworkInterface)
	h.Add("BatchUpdateNetworkInterfaceCommonInfo", "PATCH",
		"/network_interfaces/common/info/batch/update", svc.BatchUpdateNetworkInterfaceCommonInfo)
	h.Add("BatchDeleteNetworkInterface", "DELETE", "/network_interfaces/batch",
		svc.BatchDeleteNetworkInterface)
	h.Add("ListNetworkInterface", "POST", "/network_interfaces/list", svc.ListNetworkInterface)
	h.Add("ListNetworkInterfaceAssociate", "POST", "/network_interfaces/associate/list",
		svc.ListNetworkInterfaceAssociate)
	h.Add("ListNetworkInterfaceExt", "POST", "/vendors/{vendor}/network_interfaces/list",
		svc.ListNetworkInterfaceExt)
	h.Add("GetNetworkInterface", "GET", "/vendors/{vendor}/network_interfaces/{id}",
		svc.GetNetworkInterface)

	h.Load(cap.WebService)
}

// NetworkInterfaceSvc ...
type NetworkInterfaceSvc struct {
	dao dao.Set
}

// BatchCreateNetworkInterface batch create network interface.
func (svc *NetworkInterfaceSvc) BatchCreateNetworkInterface(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.Azure:
		return batchCreateNI[datacloudniproto.AzureNICreateExt](cts, vendor, svc)
	case enumor.Gcp:
		return batchCreateNI[datacloudniproto.GcpNICreateExt](cts, vendor, svc)
	case enumor.HuaWei:
		return batchCreateNI[datacloudniproto.HuaWeiNICreateExt](cts, vendor, svc)
	default:
		return nil, errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

// batchCreateNI create network interface.
func batchCreateNI[T datacloudniproto.NetworkInterfaceCreateExtension](cts *rest.Contexts, vendor enumor.Vendor,
	svc *NetworkInterfaceSvc) (interface{}, error) {

	req := new(datacloudniproto.NetworkInterfaceBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	niIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		nis := make([]tableni.NetworkInterfaceTable, 0, len(req.NetworkInterfaces))
		for _, createReq := range req.NetworkInterfaces {
			ext, err := tabletype.NewJsonField(createReq.Extension)
			if err != nil {
				return nil, errf.NewFromErr(errf.InvalidParameter, err)
			}

			ni := tableni.NetworkInterfaceTable{
				Vendor:        vendor,
				Name:          createReq.Name,
				AccountID:     createReq.AccountID,
				Region:        createReq.Region,
				Zone:          createReq.Zone,
				CloudID:       createReq.CloudID,
				VpcID:         createReq.VpcID,
				CloudVpcID:    createReq.CloudVpcID,
				SubnetID:      createReq.SubnetID,
				CloudSubnetID: createReq.CloudSubnetID,
				BkBizID:       createReq.BkBizID,
				InstanceID:    createReq.InstanceID,
				Extension:     ext,
				Creator:       cts.Kit.User,
				Reviser:       cts.Kit.User,
			}
			convertNetworkInterfaceIPInfo(&ni, createReq.PrivateIPv4, createReq.PrivateIPv6,
				createReq.PublicIPv4, createReq.PublicIPv6)

			nis = append(nis, ni)
		}

		niID, err := svc.dao.NetworkInterface().CreateWithTx(cts.Kit, txn, nis)
		if err != nil {
			return nil, fmt.Errorf("create network interface failed, err: %v", err)
		}

		return niID, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := niIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create azure network interface but return id type is not string, "+
			"id type: %v", reflect.TypeOf(niIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchUpdateNetworkInterface batch update network interface.
func (svc *NetworkInterfaceSvc) BatchUpdateNetworkInterface(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.Azure:
		return batchUpdateNI[datacloudniproto.AzureNICreateExt](cts, svc)
	case enumor.Gcp:
		return batchUpdateNI[datacloudniproto.GcpNICreateExt](cts, svc)
	case enumor.HuaWei:
		return batchUpdateNI[datacloudniproto.HuaWeiNICreateExt](cts, svc)
	}

	return nil, nil
}

// batchUpdateNI batch update network interface.
func batchUpdateNI[T datacloudniproto.NetworkInterfaceCreateExtension](cts *rest.Contexts, svc *NetworkInterfaceSvc) (
	interface{}, error) {

	req := new(datacloudniproto.NetworkInterfaceBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0, len(req.NetworkInterfaces))
	for _, niItem := range req.NetworkInterfaces {
		ids = append(ids, niItem.ID)
	}

	// check if all network interface exists
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   &core.BasePage{Count: true},
	}
	listRes, err := svc.dao.NetworkInterface().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("batch update list network interface failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list network interface failed, err: %v", err)
	}

	if listRes.Count != uint64(len(req.NetworkInterfaces)) {
		return nil, fmt.Errorf("list network interface failed, some ni(ids=%+v) doesn't exist", ids)
	}

	// update network interface
	ni := &tableni.NetworkInterfaceTable{
		Reviser: cts.Kit.User,
	}

	for _, updateReq := range req.NetworkInterfaces {
		ni.Name = updateReq.Name
		ni.Region = updateReq.Region
		ni.Zone = updateReq.Zone
		ni.CloudID = updateReq.CloudID
		ni.VpcID = updateReq.VpcID
		ni.CloudVpcID = updateReq.CloudVpcID
		ni.SubnetID = updateReq.SubnetID
		ni.CloudSubnetID = updateReq.CloudSubnetID
		ni.BkBizID = updateReq.BkBizID
		ni.InstanceID = updateReq.InstanceID

		convertNetworkInterfaceIPInfo(ni, updateReq.PrivateIPv4, updateReq.PrivateIPv6,
			updateReq.PublicIPv4, updateReq.PublicIPv6)

		// update extension
		if updateReq.Extension != nil {
			// get network interface before for expression
			dbNI, err := getNetworkInterfaceFromTable(cts.Kit, svc.dao, updateReq.ID)
			if err != nil {
				return nil, err
			}

			updatedExtension, err := json.UpdateMerge(updateReq.Extension, string(dbNI.Extension))
			if err != nil {
				return nil, fmt.Errorf("extension update network interface merge failed, err: %v", err)
			}

			ni.Extension = tabletype.JsonField(updatedExtension)
		}

		err = svc.dao.NetworkInterface().Update(cts.Kit, tools.EqualExpression("id", updateReq.ID), ni)
		if err != nil {
			logs.Errorf("batch update network interface failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("update network interface failed, err: %v", err)
		}
	}

	return nil, nil
}

// convertNetworkInterfaceIPInfo convert network interface ip info to table.
func convertNetworkInterfaceIPInfo(dbDetail *tableni.NetworkInterfaceTable, privateIPv4, privateIPv6,
	publicIPv4, publicIPv6 []string) {

	tmpIPStr, err := convertArrToTableJSON(privateIPv4)
	if err != nil {
		return
	}
	dbDetail.PrivateIPv4 = tmpIPStr

	tmpIPStr, err = convertArrToTableJSON(privateIPv6)
	if err != nil {
		return
	}
	dbDetail.PrivateIPv6 = tmpIPStr

	tmpIPStr, err = convertArrToTableJSON(publicIPv4)
	if err != nil {
		return
	}
	dbDetail.PublicIPv4 = tmpIPStr

	tmpIPStr, err = convertArrToTableJSON(publicIPv6)
	if err != nil {
		return
	}
	dbDetail.PublicIPv6 = tmpIPStr
}

func convertArrToTableJSON(arr []string) (tabletype.JsonField, error) {
	if len(arr) == 0 {
		return "[]", nil
	}

	str, err := json.Marshal(arr)
	if err != nil {
		logs.Errorf("convert arr to table json, json.Marshal failed, arr: %v, err: %v", arr, err)
		return "[]", err
	}

	return tabletype.JsonField(str), nil
}

// BatchUpdateNetworkInterfaceCommonInfo batch update network interface common info.
func (svc *NetworkInterfaceSvc) BatchUpdateNetworkInterfaceCommonInfo(cts *rest.Contexts) (interface{}, error) {
	req := new(datacloudniproto.NetworkInterfaceCommonInfoBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	updateFilter := tools.ContainersExpression("id", req.IDs)
	updateFiled := &tableni.NetworkInterfaceTable{
		BkBizID: req.BkBizID,
		Reviser: cts.Kit.User,
	}
	if err := svc.dao.NetworkInterface().Update(cts.Kit, updateFilter, updateFiled); err != nil {
		logs.Errorf("batch update network interface common info failed, req: %+v, err: %v, rid: %s",
			req, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchDeleteNetworkInterface delete network interface.
func (svc *NetworkInterfaceSvc) BatchDeleteNetworkInterface(cts *rest.Contexts) (interface{}, error) {
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
	listResp, err := svc.dao.NetworkInterface().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("delete list network interface failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("delete list network interface failed, err: %v", err)
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
		if err := svc.dao.NetworkInterface().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete network interface failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

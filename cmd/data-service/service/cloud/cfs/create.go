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

// Package cfs 如下:
// cfs handler
package cfs

import (
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	corecfs "hcm/pkg/api/core/cloud/cfs"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablecfs "hcm/pkg/dal/table/cloud/cfs"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/rest"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// CreateCfs handles the creation of CFS (Cloud File System) based on vendor information from the request context.
func (svc *cfsSvc) CreateCfs(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud: // 只支持腾讯云
		return createCfs[corecfs.TCloudCfsExtension](cts, svc, vendor)
	//case enumor.Aws:
	//	return batchCreateCfs[corecfs.AwsCfsExtension](cts, svc, vendor)
	//case enumor.HuaWei:
	//	return batchCreateCfs[corecfs.HuaWeiCfsExtension](cts, svc, vendor)
	//case enumor.Azure:
	//	return batchCreateCfs[corecfs.AzureCfsExtension](cts, svc, vendor)
	//case enumor.Gcp:
	//	return batchCreateCfs[corecfs.GcpCfsExtension](cts, svc, vendor)
	//case enumor.Other:
	//	return batchCreateCfs[corecfs.OtherCfsExtension](cts, svc, vendor)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

// createCfs creates a new CFS (Cloud File System) entry using the provided context, service, and vendor information.
// It decodes incoming requests, validates input, and performs a database transaction to persist the CFS entry.
// Returns the ID of the created entry or an error if the operation fails.
func createCfs[T corecfs.Extension](cts *rest.Contexts, svc *cfsSvc, vendor enumor.Vendor) (interface{}, error) {
	req := new(protocloud.CfsCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		extension, err := json.MarshalToString(req.Extension)
		if err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		id, err := svc.dao.Cfs().CreateWithTx(cts.Kit, txn, newTCloudCfsModel(req, vendor, extension, cts))
		if err != nil {
			return nil, fmt.Errorf("create cfs failed, err: %v", err)
		}
		return id, nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "insert cfs record failed, cloudID: %s, rid: %s", req.CloudID, cts.Kit.Rid)
	}

	id, ok := result.(string)
	if !ok {
		return nil, errors.Errorf("create cfs but return id type is not string, cloudID: %s, result: %+v, id type: %v",
			req.CloudID, result, reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: []string{id}}, nil
}

// newTCloudCfsModel creates and initializes a Table resource for a TCloudCfsModel based on the provided request and
// contexts.
func newTCloudCfsModel[T corecfs.Extension](req *protocloud.CfsCreateReq[T], vendor enumor.Vendor, extension string,
	cts *rest.Contexts) *tablecfs.Table {
	return &tablecfs.Table{
		CloudID:          req.CloudID,
		Name:             req.Name,
		Vendor:           vendor,
		BkBizID:          req.BkBizID,
		AccountID:        req.AccountID,
		Region:           req.Region,
		Zone:             req.Zone,
		SizeLimit:        req.SizeLimit,
		SizeByte:         req.SizeByte,
		AvailCapacity:    req.AvailCapacity,
		BandwidthLimit:   req.BandwidthLimit,
		Protocol:         req.Protocol,
		StorageType:      req.StorageType,
		Encrypted:        req.Encrypted,
		CryptKeyId:       req.CryptKeyId,
		CloudVpcIDs:      req.CloudVpcIDs,
		VpcIDs:           req.VpcIDs,
		CloudSubnetIDs:   req.CloudSubnetIDs,
		SubnetIDs:        req.SubnetIDs,
		Memo:             req.Memo,
		Status:           req.Status,
		Extension:        tabletype.JsonField(extension),
		CloudCreatedTime: req.CloudCreatedTime,
		Creator:          cts.Kit.User,
		Reviser:          cts.Kit.User,
	}
}

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

package cvm

import (
	"fmt"

	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecvm "hcm/pkg/dal/table/cloud/cvm"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateCvm cvm.
func (svc *cvmSvc) BatchUpdateCvm(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchUpdateCvm[corecvm.TCloudCvmExtension](cts, svc, vendor)
	case enumor.Aws:
		return batchUpdateCvm[corecvm.AwsCvmExtension](cts, svc, vendor)
	case enumor.HuaWei:
		return batchUpdateCvm[corecvm.HuaWeiCvmExtension](cts, svc, vendor)
	case enumor.Azure:
		return batchUpdateCvm[corecvm.AzureCvmExtension](cts, svc, vendor)
	case enumor.Gcp:
		return batchUpdateCvm[corecvm.GcpCvmExtension](cts, svc, vendor)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

func batchUpdateCvm[T corecvm.Extension](cts *rest.Contexts, svc *cvmSvc, vendor enumor.Vendor) (interface{}, error) {

	req := new(protocloud.CvmBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ids := make([]string, 0, len(req.Cvms))
	for _, one := range req.Cvms {
		ids = append(ids, one.ID)
	}
	// TODO list extension and cloud id
	existCvmMap, err := listCvmInfo(cts, svc, ids)
	if err != nil {
		return nil, err
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]*tablecvm.Table, 0, len(req.Cvms))

		for _, one := range req.Cvms {
			update := &tablecvm.Table{
				Name:                 one.Name,
				BkBizID:              one.BkBizID,
				BkCloudID:            one.BkCloudID,
				CloudVpcIDs:          one.CloudVpcIDs,
				CloudSubnetIDs:       one.CloudSubnetIDs,
				CloudImageID:         one.CloudImageID,
				ImageID:              one.ImageID,
				Memo:                 one.Memo,
				Status:               one.Status,
				PrivateIPv4Addresses: one.PrivateIPv4Addresses,
				PrivateIPv6Addresses: one.PrivateIPv6Addresses,
				PublicIPv4Addresses:  one.PublicIPv4Addresses,
				PublicIPv6Addresses:  one.PublicIPv6Addresses,
				CloudLaunchedTime:    one.CloudLaunchedTime,
				CloudExpiredTime:     one.CloudExpiredTime,
				Reviser:              cts.Kit.User,
			}

			existCvm, exist := existCvmMap[one.ID]
			if !exist {
				continue
			}

			if one.Extension != nil {
				merge, err := json.UpdateMerge(one.Extension, string(existCvm.Extension))
				if err != nil {
					return nil, fmt.Errorf("json UpdateMerge extension failed, err: %v", err)
				}
				update.Extension = tabletype.JsonField(merge)
			}

			if err := svc.dao.Cvm().UpdateByIDWithTx(cts.Kit, txn, one.ID, update); err != nil {
				logs.Errorf("update cvm by id failed, err: %v, id: %s, rid: %s", err, one.ID, cts.Kit.Rid)
				return nil, fmt.Errorf("update cvm failed, err: %v", err)
			}

			if update.BkCloudID == 0 {
				update.BkCloudID = existCvm.BkCloudID
			}

			if len(update.PrivateIPv4Addresses) == 0 {
				update.PrivateIPv4Addresses = existCvm.PrivateIPv4Addresses
			}

			if len(update.PrivateIPv6Addresses) == 0 {
				update.PrivateIPv6Addresses = existCvm.PrivateIPv6Addresses
			}

			if len(update.PublicIPv4Addresses) == 0 {
				update.PublicIPv4Addresses = existCvm.PublicIPv4Addresses
			}

			if len(update.PublicIPv6Addresses) == 0 {
				update.PublicIPv6Addresses = existCvm.PublicIPv6Addresses
			}

			update.CloudID = existCvm.CloudID
			update.BkBizID = existCvm.BkBizID
			models = append(models, update)
		}

		// upsert cmdb cloud hosts
		err = upsertCmdbHosts[T](svc, cts.Kit, vendor, models)
		if err != nil {
			logs.Errorf("upsert cmdb hosts failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, nil
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func listCvmInfo(cts *rest.Contexts, svc *cvmSvc, ids []string) (map[string]tablecvm.Table, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}
	list, err := svc.dao.Cvm().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	result := make(map[string]tablecvm.Table, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = one
	}

	return result, nil
}

// BatchUpdateCvmCommonInfo cvm.
func (svc *cvmSvc) BatchUpdateCvmCommonInfo(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.CvmCommonInfoBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	updateFilter := tools.ContainersExpression("id", req.IDs)
	updateFiled := &tablecvm.Table{
		BkBizID: req.BkBizID,
	}
	if err := svc.dao.Cvm().Update(cts.Kit, updateFilter, updateFiled); err != nil {
		return nil, err
	}

	// upsert cmdb cloud hosts
	opt := &types.ListOption{
		Filter: updateFilter,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.Cvm().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list cvm failed, err: %v", err)
	}

	for idx := range listResp.Details {
		listResp.Details[idx].BkBizID = req.BkBizID
	}

	err = upsertBaseCmdbHosts(svc, cts.Kit, converter.SliceToPtr(listResp.Details))
	if err != nil {
		logs.Errorf("upsert base cmdb hosts failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, nil
	}

	return nil, nil
}

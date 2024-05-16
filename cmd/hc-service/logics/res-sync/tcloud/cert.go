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

package tcloud

import (
	"fmt"
	"strconv"
	"time"

	"hcm/cmd/hc-service/logics/res-sync/common"
	typecert "hcm/pkg/adaptor/types/cert"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	corecert "hcm/pkg/api/core/cloud/cert"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
)

// SyncCertOption ...
type SyncCertOption struct {
	BkBizID int64 `json:"bk_biz_id" validate:"omitempty"`
}

// Validate ...
func (opt SyncCertOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Cert ...
func (cli *client) Cert(kt *kit.Kit, params *SyncBaseParams, opt *SyncCertOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	certFromCloud, err := cli.listCertFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	logs.Infof("[%s] hcservice sync cert listCertFromCloud success, params: %+v, cloud_cert_count: %d, rid: %s",
		enumor.TCloud, params, len(certFromCloud), kt.Rid)

	certFromDB, err := cli.listCertFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	logs.Infof("[%s] hcservice sync cert listCertFromDB success, db_cert_count: %d, rid: %s",
		enumor.TCloud, len(certFromDB), kt.Rid)

	if len(certFromCloud) == 0 && len(certFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typecert.TCloudCert, *corecert.Cert[corecert.TCloudCertExtension]](
		certFromCloud, certFromDB, isCertChange)

	logs.Infof("[%s] hcservice sync cert diff success, addNum: %d, updateNum: %d, delNum: %d, rid: %s",
		enumor.TCloud, len(addSlice), len(updateMap), len(delCloudIDs), kt.Rid)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteCert(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createCert(kt, params.AccountID, opt, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateCert(kt, params.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	return new(SyncResult), nil
}

func (cli *client) deleteCert(kt *kit.Kit, accountID, region string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("hcservice resource sync failed, delCloudIDs is <= 0, not delete")
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  delCloudIDs,
	}
	delFromCloud, err := cli.listCertFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delFromCloud) > 0 {
		logs.Errorf("[%s] validate cert not exist failed, before delete, opt: %v, failed_count: %d, rid: %s",
			enumor.TCloud, checkParams, len(delFromCloud), kt.Rid)
		return fmt.Errorf("validate cert not exist failed, before delete")
	}

	deleteReq := &protocloud.CertBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.BatchDeleteCert(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete cert failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cert to delete cert success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		accountID, len(delCloudIDs), kt.Rid)

	return nil
}

func (cli *client) updateCert(kt *kit.Kit, accountID string, updateMap map[string]typecert.TCloudCert) error {
	if len(updateMap) <= 0 {
		return fmt.Errorf("hcservice resource sync failed, updateMap is <= 0, not update")
	}

	certs := make([]*protocloud.CertExtUpdateReq[corecert.TCloudCertExtension], 0)

	for id, one := range updateMap {
		domainJson, err := types.NewJsonField(one.SubjectAltName)
		if err != nil {
			return fmt.Errorf("json marshal extension failed, err: %w", err)
		}

		cert := &protocloud.CertExtUpdateReq[corecert.TCloudCertExtension]{
			ID:               id,
			Name:             converter.PtrToVal(one.Alias),
			Vendor:           string(enumor.TCloud),
			AccountID:        accountID,
			Domain:           domainJson,
			CertType:         enumor.CertType(converter.PtrToVal(one.CertificateType)),
			EncryptAlgorithm: converter.PtrToVal(one.EncryptAlgorithm),
			CertStatus:       strconv.FormatUint(converter.PtrToVal(one.Status), 10),
			CloudCreatedTime: converter.PtrToVal(one.InsertTime),
			CloudExpiredTime: converter.PtrToVal(one.CertEndTime),
		}

		certs = append(certs, cert)
	}

	var updateReq protocloud.CertExtBatchUpdateReq[corecert.TCloudCertExtension]
	for _, item := range certs {
		updateReq = append(updateReq, item)
	}
	if _, err := cli.dbCli.TCloud.BatchUpdateCert(kt.Ctx, kt.Header(), &updateReq); err != nil {
		logs.Errorf("[%s] request dataservice BatchUpdateCert failed, err: %v, rid: %s", enumor.TCloud,
			err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cert to update cert success, accountID: %s, count: %d, rid: %s", enumor.TCloud,
		accountID, len(updateMap), kt.Rid)

	return nil
}

func (cli *client) createCert(kt *kit.Kit, accountID string, opt *SyncCertOption, addSlice []typecert.TCloudCert) error {
	if len(addSlice) <= 0 {
		return fmt.Errorf("hcservice resource sync failed, addSlice is <= 0, not create")
	}

	var createReq = new(protocloud.CertBatchCreateReq[corecert.TCloudCertExtension])

	for _, one := range addSlice {
		domainJson, err := types.NewJsonField(one.SubjectAltName)
		if err != nil {
			return fmt.Errorf("json marshal extension failed, err: %w", err)
		}

		cert := []protocloud.CertBatchCreate[corecert.TCloudCertExtension]{
			{
				CloudID:          one.GetCloudID(),
				Name:             converter.PtrToVal(one.Alias),
				Vendor:           string(enumor.TCloud),
				AccountID:        accountID,
				BkBizID:          opt.BkBizID,
				Domain:           domainJson,
				CertType:         enumor.CertType(converter.PtrToVal(one.CertificateType)),
				EncryptAlgorithm: converter.PtrToVal(one.EncryptAlgorithm),
				CertStatus:       strconv.FormatUint(converter.PtrToVal(one.Status), 10),
				CloudCreatedTime: converter.PtrToVal(one.InsertTime),
				CloudExpiredTime: converter.PtrToVal(one.CertEndTime),
			},
		}

		createReq.Certs = append(createReq.Certs, cert...)
	}

	newIDs, err := cli.dbCli.TCloud.BatchCreateCert(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create tcloud cert failed, createReq: %+v, err: %v, rid: %s",
			enumor.TCloud, createReq, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync cert to create cert success, accountID: %s, count: %d, newIDs: %v, opt: %+v, rid: %s", enumor.TCloud,
		accountID, len(addSlice), newIDs, opt, kt.Rid)

	return nil
}

func (cli *client) listCertFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]typecert.TCloudCert, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	list := make([]typecert.TCloudCert, 0)
	for _, tmpCloudID := range params.CloudIDs {
		opt := &typecert.TCloudListOption{
			SearchKey: tmpCloudID,
			Page:      &adcore.TCloudPage{Offset: 0, Limit: 1},
		}
		result, err := cli.cloudCli.ListCert(kt, opt)
		if err != nil {
			logs.Errorf("[%s] list cert from cloud failed, account: %s, opt: %v, err: %v, rid: %s", enumor.TCloud,
				params.AccountID, opt, err, kt.Rid)
			return nil, err
		}

		list = append(list, result...)
	}

	return list, nil
}

func (cli *client) listCertFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]*corecert.Cert[corecert.TCloudCertExtension], error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: params.AccountID,
				},
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: params.CloudIDs,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.TCloud.ListCert(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list cert from db failed, account: %s, req: %v, err: %v, rid: %s",
			enumor.TCloud, params.AccountID, req, err, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func isCertChange(cloud typecert.TCloudCert, db *corecert.Cert[corecert.TCloudCertExtension]) bool {
	if converter.PtrToVal(cloud.Alias) != db.Name {
		return true
	}

	if !assert.IsPtrStringSliceEqual(cloud.SubjectAltName, db.Domain) {
		return true
	}

	if enumor.CertType(converter.PtrToVal(cloud.CertificateType)) != db.CertType {
		return true
	}

	statusCloud := strconv.FormatUint(converter.PtrToVal(cloud.Status), 10)
	if statusCloud != db.CertStatus {
		return true
	}

	cloudEndTime := converter.PtrToVal(cloud.CertEndTime)
	if len(cloudEndTime) == 0 && len(db.CloudExpiredTime) > 0 {
		return true
	}

	if len(cloudEndTime) > 0 && len(db.CloudExpiredTime) == 0 {
		return true
	}

	expireTime, err := time.Parse(constant.TimeStdFormat, db.CloudExpiredTime)
	if err != nil {
		logs.Errorf("cert sync expired time parse failed, dbExpireTime: %s, err: %v", db.CloudExpiredTime, err)
		return true
	}

	if cloudEndTime != expireTime.Format(constant.DateTimeLayout) {
		return true
	}

	if converter.PtrToVal(cloud.EncryptAlgorithm) != db.EncryptAlgorithm {
		return true
	}

	return false
}

// RemoveCertDeleteFromCloud ...
func (cli *client) RemoveCertDeleteFromCloud(kt *kit.Kit, accountID, region string) error {
	req := &core.ListReq{
		Fields: []string{"id", "cloud_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		resultFromDB, err := cli.dbCli.Global.ListCert(kt, req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list cert failed, req: %v, err: %v, rid: %s",
				enumor.TCloud, req, err, kt.Rid)
			return err
		}

		cloudIDs := make([]string, 0)
		for _, one := range resultFromDB.Details {
			cloudIDs = append(cloudIDs, one.CloudID)
		}

		if len(cloudIDs) == 0 {
			break
		}

		params := &SyncBaseParams{
			AccountID: accountID,
			Region:    region,
			CloudIDs:  cloudIDs,
		}
		resultFromCloud, err := cli.listCertFromCloud(kt, params)
		if err != nil {
			return err
		}

		// 如果有资源没有查询出来，说明数据被从云上删除
		if len(resultFromCloud) != len(cloudIDs) {
			cloudIDMap := converter.StringSliceToMap(cloudIDs)
			for _, one := range resultFromCloud {
				delete(cloudIDMap, converter.PtrToVal(one.CertificateId))
			}

			cloudIDs = converter.MapKeyToStringSlice(cloudIDMap)
			if err = cli.deleteCert(kt, accountID, region, cloudIDs); err != nil {
				return err
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

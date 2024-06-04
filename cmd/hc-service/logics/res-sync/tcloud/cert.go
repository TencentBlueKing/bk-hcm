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
	"hcm/pkg/tools/slice"
)

// SyncCertOption ...
type SyncCertOption struct {
	BkBizID int64 `json:"bk_biz_id" validate:"omitempty"`
	// should match params' cloud id
	PreCachedCertList []typecert.TCloudCert
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
	certFromCloud := opt.PreCachedCertList
	if certFromCloud == nil {
		var err error
		certFromCloud, err = cli.listCertFromCloud(kt, params)
		if err != nil {
			return nil, err
		}
	}

	certFromDB, err := cli.listCertFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(certFromCloud) == 0 && len(certFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typecert.TCloudCert, *corecert.Cert[corecert.TCloudCertExtension]](
		certFromCloud, certFromDB, isCertChange)

	if err = cli.deleteCert(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
		return nil, err
	}

	if err = cli.createCert(kt, params.AccountID, opt, addSlice); err != nil {
		return nil, err
	}

	if err = cli.updateCert(kt, params.AccountID, updateMap); err != nil {
		return nil, err
	}

	return new(SyncResult), nil
}

func (cli *client) deleteCert(kt *kit.Kit, accountID, region string, delCloudIDs []string) error {
	if len(delCloudIDs) <= 0 {
		return nil
	}

	deleteReq := &protocloud.CertBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err := cli.dbCli.Global.BatchDeleteCert(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete cert failed, err: %v, rid: %s",
			enumor.TCloud, err, kt.Rid)
		return err
	}

	return nil
}

func (cli *client) updateCert(kt *kit.Kit, accountID string, updateMap map[string]typecert.TCloudCert) error {
	if len(updateMap) <= 0 {
		return nil
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
			CloudCreatedTime: convTCloudTimeStd(converter.PtrToVal(one.InsertTime)),
			CloudExpiredTime: convTCloudTimeStd(converter.PtrToVal(one.CertEndTime)),
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

func (cli *client) createCert(kt *kit.Kit, accountID string, opt *SyncCertOption,
	addSlice []typecert.TCloudCert) error {

	if len(addSlice) <= 0 {
		return nil
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
				CloudCreatedTime: convTCloudTimeStd(converter.PtrToVal(one.InsertTime)),
				CloudExpiredTime: convTCloudTimeStd(converter.PtrToVal(one.CertEndTime)),
			},
		}

		createReq.Certs = append(createReq.Certs, cert...)
	}

	_, err := cli.dbCli.TCloud.BatchCreateCert(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create tcloud cert failed, createReq: %+v, err: %v, rid: %s",
			enumor.TCloud, createReq, err, kt.Rid)
		return err
	}

	return nil
}

// 不支持按id批量获取，直接获取全部数据可以降低调用腾讯云api次数
func (cli *client) listAllCertFromCloud(kt *kit.Kit) ([]typecert.TCloudCert, error) {

	list := make([]typecert.TCloudCert, 0, 100)
	opt := &typecert.TCloudListOption{
		Page: &adcore.TCloudPage{Offset: 0, Limit: adcore.TCloudQueryLimit},
	}
	for {
		result, err := cli.cloudCli.ListCert(kt, opt)
		if err != nil {
			logs.Errorf("[%s] list all cert from cloud failed, account: %s, opt: %v, err: %v, rid: %s",
				enumor.TCloud, cli.accountID, opt, err, kt.Rid)
			return nil, err
		}

		list = append(list, result...)

		if uint64(len(result)) < opt.Page.Limit {
			break
		}
		opt.Page.Offset += opt.Page.Limit

	}

	return list, nil
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

	if db.CloudExpiredTime != convTCloudTimeStd(converter.PtrToVal(cloud.CertEndTime)) {
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
	// 全量获取一次云端证书数据
	allResultFromCloud, err := cli.listAllCertFromCloud(kt)
	certCloudIDMap := make(map[string]struct{}, len(allResultFromCloud))
	for _, cert := range allResultFromCloud {
		certCloudIDMap[cert.GetCloudID()] = struct{}{}
	}
	if err != nil {
		return err
	}
	delCloudIDs := make([]string, 0)
	for {
		resultFromDB, err := cli.dbCli.Global.ListCert(kt, req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list cert failed, req: %v, err: %v, rid: %s",
				enumor.TCloud, req, err, kt.Rid)
			return err
		}
		for _, detail := range resultFromDB.Details {
			if _, ok := certCloudIDMap[detail.CloudID]; !ok {
				delCloudIDs = append(delCloudIDs, detail.CloudID)
			}
		}

		if len(resultFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}
	if len(delCloudIDs) == 0 {
		return nil
	}

	for _, delCloudBatch := range slice.Split(delCloudIDs, constant.BatchOperationMaxLimit) {
		if err = cli.deleteCert(kt, accountID, region, delCloudBatch); err != nil {
			return err
		}

	}

	return nil
}

func convTCloudTimeStd(t string) string {
	parse, err := time.Parse(constant.DateTimeLayout, t)
	if err != nil {
		logs.Errorf("[%s] parse time failed, time: %s, err: %v", enumor.TCloud, t, err)
		return ""
	}
	return parse.Format(constant.TimeStdFormat)
}

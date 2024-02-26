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
	"errors"
	"fmt"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	typecert "hcm/pkg/adaptor/types/cert"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	terr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

// CreateCert 上传证书
// reference: https://cloud.tencent.com/document/api/400/41665
func (t *TCloudImpl) CreateCert(kt *kit.Kit, opt *typecert.TCloudCreateOption) (*poller.BaseDoneResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "create option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.certClient()
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud cert client failed, err: %v", err)
	}

	req := ssl.NewUploadCertificateRequest()
	req.Alias = common.StringPtr(opt.Name)
	req.CertificateType = common.StringPtr(opt.CertType)
	req.CertificatePublicKey = common.StringPtr(opt.PublicKey)
	req.CertificatePrivateKey = common.StringPtr(opt.PrivateKey)
	req.Repeatable = common.BoolPtr(opt.Repeatable)
	resp, err := client.UploadCertificateWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("run tencent cloud cert instance failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return nil, err
	}

	handler := &createCertPollingHandler{}
	respPoller := poller.Poller[*TCloudImpl, []typecert.TCloudCert, poller.BaseDoneResult]{Handler: handler}
	result, err := respPoller.PollUntilDone(t, kt, []*string{resp.Response.CertificateId},
		types.NewBatchCreateCertPollerOption())
	if err != nil {
		logs.Errorf("run tencent cloud cert poller failed, resp: %+v, err: %v, rid: %s", resp.Response, err, kt.Rid)
		return nil, err
	}

	return result, nil
}

// ListCert list cert.
// reference: https://cloud.tencent.com/document/api/400/41671
func (t *TCloudImpl) ListCert(kt *kit.Kit, opt *typecert.TCloudListOption) ([]typecert.TCloudCert, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.certClient()
	if err != nil {
		return nil, fmt.Errorf("new tcloud cert client failed, err: %v", err)
	}

	req := ssl.NewDescribeCertificatesRequest()
	if len(opt.SearchKey) > 0 {
		req.SearchKey = common.StringPtr(opt.SearchKey)
	}

	if opt.Page != nil {
		req.Offset = common.Uint64Ptr(opt.Page.Offset)
		req.Limit = common.Uint64Ptr(opt.Page.Limit)
	}

	resp, err := client.DescribeCertificatesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud cert failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	certs := make([]typecert.TCloudCert, 0, len(resp.Response.Certificates))
	for _, one := range resp.Response.Certificates {
		certs = append(certs, typecert.TCloudCert{Certificates: one})
	}

	return certs, nil
}

// CountCert count cert in given region
// reference: https://cloud.tencent.com/document/api/400/41671
func (t *TCloudImpl) CountCert(kt *kit.Kit) (int32, error) {
	client, err := t.clientSet.certClient()
	if err != nil {
		return 0, fmt.Errorf("new tcloud cert client failed, err: %v", err)
	}

	req := ssl.NewDescribeCertificatesRequest()
	req.Limit = common.Uint64Ptr(1)
	resp, err := client.DescribeCertificatesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("count tcloud cert failed, err: %v, rid: %s", err, kt.Rid)
		return 0, err
	}

	return int32(*resp.Response.TotalCount), nil
}

// DeleteCert delete cert
// reference: https://cloud.tencent.com/document/api/213/15723
func (t *TCloudImpl) DeleteCert(kt *kit.Kit, opt *typecert.TCloudDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "delete cert option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.certClient()
	if err != nil {
		return fmt.Errorf("init tencent cloud cert client failed, err: %v", err)
	}

	req := ssl.NewDeleteCertificateRequest()
	req.CertificateId = common.StringPtr(opt.CloudID)

	_, err = client.DeleteCertificateWithContext(kt.Ctx, req)
	if err != nil {
		// 兼容证书不存在
		var tErr *terr.TencentCloudSDKError
		if errors.As(err, &tErr) && tErr.GetCode() == "FailedOperation.CertificateNotFound" {
			logs.Errorf("delete cert instance failed, cert not exist, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
			return nil
		}

		logs.Errorf("delete cert instance failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return err
	}

	return nil
}

type createCertPollingHandler struct{}

func (h *createCertPollingHandler) Done(certs []typecert.TCloudCert) (bool, *poller.BaseDoneResult) {
	result := &poller.BaseDoneResult{
		SuccessCloudIDs: make([]string, 0),
		FailedCloudIDs:  make([]string, 0),
		UnknownCloudIDs: make([]string, 0),
	}
	flag := true
	for _, instance := range certs {
		// 审核中
		if converter.PtrToVal(instance.Status) == 0 {
			flag = false
			result.UnknownCloudIDs = append(result.UnknownCloudIDs, *instance.CertificateId)
			continue
		}

		// 不是[已通过、已过期]的状态
		if converter.PtrToVal(instance.Status) != 1 && converter.PtrToVal(instance.Status) != 3 {
			result.FailedCloudIDs = append(result.FailedCloudIDs, *instance.CertificateId)
			result.FailedMessage = converter.PtrToVal(instance.StatusMsg)
			continue
		}

		result.SuccessCloudIDs = append(result.SuccessCloudIDs, *instance.CertificateId)
	}

	return flag, result
}

func (h *createCertPollingHandler) Poll(client *TCloudImpl, kt *kit.Kit, cloudIDs []*string) (
	[]typecert.TCloudCert, error) {

	cloudIDSplit := slice.Split(cloudIDs, core.TCloudQueryLimit)
	certs := make([]typecert.TCloudCert, 0, len(cloudIDs))
	for _, partIDs := range cloudIDSplit {
		for idx, certCloudID := range partIDs {
			opt := &typecert.TCloudListOption{
				SearchKey: converter.PtrToVal(certCloudID),
				Page:      &core.TCloudPage{Offset: uint64(idx), Limit: 1},
			}
			resp, err := client.ListCert(kt, opt)
			if err != nil {
				return nil, err
			}

			certs = append(certs, resp...)
		}
	}

	if len(certs) != len(cloudIDs) {
		return nil, fmt.Errorf("query cert count: %d not equal return count: %d", len(cloudIDs), len(certs))
	}

	return certs, nil
}

var _ poller.PollingHandler[*TCloudImpl, []typecert.TCloudCert, poller.BaseDoneResult] = new(createCertPollingHandler)

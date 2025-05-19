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

package other

import (
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewCloudCvmClient create a new cvm api client.
func NewCloudCvmClient(client rest.ClientInterface) *CvmClient {
	return &CvmClient{
		client: client,
	}
}

// CvmClient is data service cvm api client.
type CvmClient struct {
	client rest.ClientInterface
}

// BatchCreateCvm batch create cvm rule.
func (cli *CvmClient) BatchCreateCvm(kt *kit.Kit,
	request *protocloud.CvmBatchCreateReq[corecvm.OtherCvmExtension]) (*core.BatchCreateResult, error) {

	resp, err := common.Request[protocloud.CvmBatchCreateReq[corecvm.OtherCvmExtension], core.BatchCreateResp](
		cli.client, rest.POST, kt, request, "/cvms/batch/create")
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// BatchUpdateCvm batch update cvm.
func (cli *CvmClient) BatchUpdateCvm(kt *kit.Kit,
	request *protocloud.CvmBatchUpdateReq[corecvm.OtherCvmExtension]) error {

	err := common.RequestNoResp[protocloud.CvmBatchUpdateReq[corecvm.OtherCvmExtension]](
		cli.client, rest.PATCH, kt, request, "/cvms/batch/update")
	if err != nil {
		return err
	}
	return nil
}

// GetCvm get cvm.
func (cli *CvmClient) GetCvm(kt *kit.Kit, id string) (
	*corecvm.Cvm[corecvm.OtherCvmExtension], error) {

	resp := new(protocloud.CvmGetResp[corecvm.OtherCvmExtension])

	resp, err := common.Request[interface{}, protocloud.CvmGetResp[corecvm.OtherCvmExtension]](
		cli.client, rest.GET, kt, nil, "/cvms/%s", id)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// ListCvmExt list cvm with extension.
func (cli *CvmClient) ListCvmExt(kt *kit.Kit, request *protocloud.CvmListReq) (
	*protocloud.CvmExtListResult[corecvm.OtherCvmExtension], error) {

	resp, err := common.Request[protocloud.CvmListReq, protocloud.CvmExtListResp[corecvm.OtherCvmExtension]](cli.client,
		rest.POST, kt, request, "/cvms/list")
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

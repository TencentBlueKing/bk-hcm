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

package common

import (
	"bytes"
	"io/ioutil"

	cloudserver "hcm/pkg/api/cloud-server"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetCloudResourceBasicInfo get cloud resource basicInfo
func GetCloudResourceBasicInfo(accountID string, bizID int64) *types.CloudResourceBasicInfo {
	var basicInfo *types.CloudResourceBasicInfo

	if bizID == int64(constant.UnassignedBiz) {
		basicInfo = &types.CloudResourceBasicInfo{
			AccountID: accountID,
		}
	} else {
		basicInfo = &types.CloudResourceBasicInfo{
			BkBizID: bizID,
		}
	}

	return basicInfo
}

// ExtractAccountID extract accountID
func ExtractAccountID(cts *rest.Contexts) (string, error) {
	req := new(cloudserver.AccountReq)
	reqData, err := ioutil.ReadAll(cts.Request.Request.Body)
	if err != nil {
		logs.Errorf("read request body failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return "", err
	}

	cts.Request.Request.Body = ioutil.NopCloser(bytes.NewReader(reqData))
	if err := cts.DecodeInto(req); err != nil {
		return "", err
	}

	if err := req.Validate(); err != nil {
		return "", errf.NewFromErr(errf.InvalidParameter, err)
	}

	cts.Request.Request.Body = ioutil.NopCloser(bytes.NewReader(reqData))

	return req.AccountID, nil
}

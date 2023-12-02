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

package azure

import (
	"errors"
	"fmt"
	"io/ioutil"

	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
	"github.com/tidwall/gjson"
)

// errorf azure的error格式不可读，需要进行一层转换，取错误原因做为error返回。
/*
e.g:
PUT https://management.azure.com/subscriptions/xx/resourceGroups/xx/providers/Microsoft.Compute/virtualMachines/xx
--------------------------------------------------------------------------------
RESPONSE 409: 409 Conflict
ERROR CODE: OperationNotAllowed
--------------------------------------------------------------------------------
{
  "error": {
    "code": "OperationNotAllowed",
    "message": "Managed disk storage account type change through Virtual Machine 'xx' is not allowed.",
    "target": "osDisk.managedDisk.storageAccountType"
  }
}
--------------------------------------------------------------------------------
*/
func errorf(azureError error) error {
	azureErr, ok := azureError.(*azcore.ResponseError)
	if ok {
		all, err := ioutil.ReadAll(azureErr.RawResponse.Body)
		if err != nil {
			logs.ErrorDepthf(1, "read azure error body failed, skip read error message, err: %v")
			return azureError
		}

		message := gjson.GetBytes(all, "error.message").String()
		if len(message) == 0 {
			return azureErr
		}

		return errors.New(message)
	} else {
		return azureErr
	}
}

// azure 的graph api sdk的错误被层层包裹，需要通过特殊手段解开。
// ref: https://github.com/microsoftgraph/msgraph-sdk-go/issues/510
// TODO: 验证升级到1.19.0是否能够直接返回错误信息
func extractGraphError(graphErr error) error {
	var oDataError *odataerrors.ODataError
	switch {
	case errors.As(graphErr, &oDataError):
		if terr := oDataError.GetErrorEscaped(); terr != nil {
			return fmt.Errorf("%v,%v,%v", oDataError, converter.PtrToVal(terr.GetCode()),
				converter.PtrToVal(terr.GetMessage()))
		}
		return oDataError
	default:
		logs.Errorf("fail to get azure applications, error(%T): %#v", graphErr, graphErr)
		return graphErr
	}
}

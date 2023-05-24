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
	"hcm/pkg/kit"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

// ResultHandlerInterface 海垒对azure返回的数据的封装的处理器
type ResultHandlerInterface[AzureRespType any, ResultType any] interface {
	BuildResult(AzureRespType) []ResultType
}

// Pager 因为azure不支持limit获取分页数据，所以构建了pager去统一管理查询接口分页。
type Pager[AzureRespType any, ResultType any] struct {
	pager         *runtime.Pager[AzureRespType]
	resultHandler ResultHandlerInterface[AzureRespType, ResultType]
}

// More ...
func (pager *Pager[AzureRespType, ResultType]) More() bool {
	return pager.pager.More()
}

// NextPage ...
func (pager *Pager[AzureRespType, ResultType]) NextPage(kt *kit.Kit) ([]ResultType, error) {
	if !pager.pager.More() {
		return make([]ResultType, 0), nil
	}

	page, err := pager.pager.NextPage(kt.Ctx)
	if err != nil {
		return nil, err
	}

	return pager.resultHandler.BuildResult(page), nil
}

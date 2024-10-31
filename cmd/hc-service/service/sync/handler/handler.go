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

// Package handler ...
package handler

import (
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// Handler 定义了全量同步操作函数。
type Handler interface {
	// Prepare 解析请求体，构建同步所需客户端。
	Prepare(cts *rest.Contexts) error
	// Next 去云上分页查询资源云ID，用于同步，每次分页查询 constant.CloudResourceSyncMaxLimit 条数据。
	Next(kt *kit.Kit) ([]string, error)
	// Sync 同步传入的 cloudIDs 的资源数据。
	Sync(kt *kit.Kit, cloudIDs []string) error
	// RemoveDeleteFromCloud 进行db和云上数据的全量对比，删除已经从云上删除的数据。
	RemoveDeleteFromCloud(kt *kit.Kit) error

	Name() enumor.CloudResourceType
}

// ResourceSync 资源同步流程。
func ResourceSync(cts *rest.Contexts, handler Handler) error {
	kt := cts.Kit

	// 解析请求参数到handler实现中，构建同步需要的客户端
	if err := handler.Prepare(cts); err != nil {
		logs.Errorf("%s sync handler to prepare failed, err: %v, rid: %s", handler.Name(), err, kt.Rid)
		return err
	}

	if err := handler.RemoveDeleteFromCloud(kt); err != nil {
		logs.Errorf("%s sync handler to removeDeleteFromCloud failed, err: %v, rid: %s", handler.Name(), err, kt.Rid)
		return err
	}

	for {
		cloudIDs, err := handler.Next(kt)
		if err != nil {
			logs.Errorf("%s sync handler to next failed, err: %v, rid: %s", handler.Name(), err, kt.Rid)
			return err
		}

		if len(cloudIDs) == 0 {
			break
		}

		if err = handler.Sync(kt, cloudIDs); err != nil {
			logs.Errorf("%s sync handler to sync failed, err: %v, rid: %s", handler.Name(), err, kt.Rid)
			return err
		}

		if len(cloudIDs) < constant.CloudResourceSyncMaxLimit {
			break
		}
	}

	return nil
}

// HandlerV2 定义了全量同步操作函数。实验性全量同步
type HandlerV2 interface {
	// Prepare 解析请求体，构建同步所需客户端。
	Prepare(cts *rest.Contexts) error
	// Next 去云上分页查询资源云ID，用于同步，每次分页查询 constant.CloudResourceSyncMaxLimit 条数据。
	Next(kt *kit.Kit) ([]string, error)
	// Sync 同步传入的 cloudIDs 的资源数据。
	Sync(kt *kit.Kit, cloudIDs []string) error

	// RemoveDeleteFromCloudV2 进行db和云上数据的全量对比，删除已经从云上删除的数据。
	RemoveDeleteFromCloudV2(kt *kit.Kit, allCloudIDMap map[string]struct{}) error

	Name() enumor.CloudResourceType
}

// ResourceSyncV2 资源同步流程。
func ResourceSyncV2(cts *rest.Contexts, handler HandlerV2) error {
	kt := cts.Kit

	// 解析请求参数到handler实现中，构建同步需要的客户端
	if err := handler.Prepare(cts); err != nil {
		logs.Errorf("%s sync handler to prepare failed, err: %v, rid: %s", handler.Name(), err, kt.Rid)
		return err
	}

	allCloudIDMap := make(map[string]struct{}, 1024)
	for {
		cloudIDs, err := handler.Next(kt)
		if err != nil {
			logs.Errorf("%s sync handler to next failed, err: %v, rid: %s", handler.Name(), err, kt.Rid)
			return err
		}

		if len(cloudIDs) == 0 {
			break
		}
		for i := range cloudIDs {
			allCloudIDMap[cloudIDs[i]] = struct{}{}
		}

		if err = handler.Sync(kt, cloudIDs); err != nil {
			logs.Errorf("%s sync handler to sync failed, err: %v, rid: %s", handler.Name(), err, kt.Rid)
			return err
		}

		if len(cloudIDs) < constant.CloudResourceSyncMaxLimit {
			break
		}
	}
	// 删除云上已删除数据
	if err := handler.RemoveDeleteFromCloudV2(kt, allCloudIDMap); err != nil {
		logs.Errorf("%s sync handler to removeDeleteFromCloud failed, err: %v, rid: %s", handler.Name(), err, kt.Rid)
		return err
	}
	return nil
}

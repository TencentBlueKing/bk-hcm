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

package sync

// Diff 对比源数据和目标数据的增/删/改数据。
func Diff[SourceDataType SourceData, TargetDataType TargetData](
	sourceData []SourceDataType, targetData []TargetDataType, isChange func(SourceDataType, TargetDataType) bool) (
	createData []SourceDataType, idUpdateDataMap map[string]SourceDataType, delIDs []string) {

	uuidTargetDataMap := make(map[string]TargetDataType, len(targetData))
	for _, one := range targetData {
		uuidTargetDataMap[one.GetUUID()] = one
	}

	for _, oneFromSource := range sourceData {
		oneFromTarget, exist := uuidTargetDataMap[oneFromSource.GetUUID()]
		if !exist {
			createData = append(createData, oneFromSource)
			continue
		}

		delete(uuidTargetDataMap, oneFromSource.GetUUID())
		if isChange(oneFromSource, oneFromTarget) {
			idUpdateDataMap[oneFromTarget.GetUUID()] = oneFromSource
		}
	}

	for _, one := range uuidTargetDataMap {
		delIDs = append(delIDs, one.GetID())
	}

	return createData, idUpdateDataMap, delIDs
}

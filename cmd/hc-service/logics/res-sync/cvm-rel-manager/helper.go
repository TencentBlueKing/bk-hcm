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

package cvmrelmgr

func diffCvmWithAssResRel(cloud map[string][]string, db map[string]map[string]cvmRelInfo) ([]cvmRelInfo, []uint64) {

	addRels := make([]cvmRelInfo, 0)
	delIDs := make([]uint64, 0)
	for cvmID, assResIDsFromCloud := range cloud {
		assResIDMapFromDB, exist := db[cvmID]
		// db不存在当前cvm的关系映射，代表是新增的关系
		if !exist {
			for _, assResID := range assResIDsFromCloud {
				addRels = append(addRels, cvmRelInfo{
					AssResID: assResID,
					CvmID:    cvmID,
				})
			}
			continue
		}

		delete(db, cvmID)
		for _, assResID := range assResIDsFromCloud {
			_, exist := assResIDMapFromDB[assResID]
			if !exist {
				addRels = append(addRels, cvmRelInfo{
					AssResID: assResID,
					CvmID:    cvmID,
				})
				continue
			}

			delete(assResIDMapFromDB, assResID)
		}

		for _, rel := range assResIDMapFromDB {
			delIDs = append(delIDs, rel.RelID)
		}
	}

	for _, relMap := range db {
		for _, rel := range relMap {
			delIDs = append(delIDs, rel.RelID)
		}
	}

	return addRels, delIDs
}

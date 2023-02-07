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

package cloud

import (
	"fmt"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// Cloud only used for cloud common operation.
type Cloud interface {
	ListResourceBasicInfo(kt *kit.Kit, resType enumor.CloudResourceType, ids []string) (
		[]types.CloudResourceBasicInfo, error)
}

var _ Cloud = new(CloudDao)

// CloudDao cloud dao.
type CloudDao struct {
	Orm orm.Interface
}

// ListResourceBasicInfo list cloud resource basic info.
func (dao CloudDao) ListResourceBasicInfo(kt *kit.Kit, resType enumor.CloudResourceType, ids []string) (
	[]types.CloudResourceBasicInfo, error) {

	tableName, err := resType.ConvTableName()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if len(ids) == 0 {
		return nil, errf.New(errf.InvalidParameter, "ids is required")
	}

	sql := fmt.Sprintf("select id, vendor, account_id from %s where id in (:ids)", tableName)
	if tableName == table.AccountTable {
		sql = fmt.Sprintf("select id, vendor, id as account_id from %s where id in (:ids)", tableName)
	}

	list := make([]types.CloudResourceBasicInfo, 0)
	args := map[string]interface{}{
		"ids": ids,
	}
	if err := dao.Orm.Do().Select(kt.Ctx, &list, sql, args); err != nil {
		logs.Errorf("select resource vendor failed, err: %v, table: %s, id: %v, rid: %s", err, resType, ids, kt.Rid)
		return nil, err
	}

	return list, nil
}

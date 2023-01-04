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
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// Cloud only used for cloud common operation.
type Cloud interface {
	GetResourceVendor(kt *kit.Kit, resType enumor.CloudResourceType, id string) (enumor.Vendor, error)
}

var _ Cloud = new(CloudDao)

// CloudDao cloud dao.
type CloudDao struct {
	Orm orm.Interface
}

// GetResourceVendor get cloud resource vendor.
func (dao CloudDao) GetResourceVendor(kt *kit.Kit, resType enumor.CloudResourceType, id string) (enumor.Vendor, error) {
	tableName, err := resType.ConvTableName()
	if err != nil {
		return "", errf.NewFromErr(errf.InvalidParameter, err)
	}

	if len(id) == 0 {
		return "", errf.New(errf.InvalidParameter, "id is required")
	}

	sql := fmt.Sprintf("select vendor from %s where id = %s", tableName, id)

	var vendor enumor.Vendor = ""
	if err := dao.Orm.Do().Get(kt.Ctx, &vendor, sql); err != nil {
		logs.Errorf("get resource vendor failed, table: %s, id: %s, err: %v, rid: %s", resType, id, err, kt.Rid)
		return "", err
	}

	return vendor, nil
}

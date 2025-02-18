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

package huawei

import (
	"fmt"

	typeszone "hcm/pkg/adaptor/types/zone"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dcs/v2/model"
)

// ListZone list zone.
// reference: https://support.huaweicloud.com/api-dcs/ListAvailableZones.html
func (h *HuaWei) ListZone(kt *kit.Kit, opt *typeszone.HuaWeiZoneListOption) ([]typeszone.HuaWeiZone, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "huawei zone list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.dcsClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new huawei dcs client failed, err: %v", err)
	}

	req := &model.ListAvailableZonesRequest{}
	resp, err := client.ListAvailableZones(req)
	if err != nil {
		logs.Errorf("list huawei zone failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp == nil {
		return make([]typeszone.HuaWeiZone, 0), nil
	}

	results := make([]typeszone.HuaWeiZone, 0)
	for _, one := range converter.PtrToVal(resp.AvailableZones) {
		results = append(results, typeszone.HuaWeiZone{one})
	}

	return results, nil
}

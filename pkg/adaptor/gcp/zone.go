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

package gcp

import (
	typeszone "hcm/pkg/adaptor/types/zone"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"strings"

	"google.golang.org/api/compute/v1"
)

// ListZone list zone
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/zones/list
func (g *Gcp) ListZone(kit *kit.Kit, opt *typeszone.GcpZoneListOption) ([]typeszone.GcpZone, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "gcp zone list option is required")
	}

	client, err := g.clientSet.computeClient(kit)
	if err != nil {
		return nil, err
	}

	zones := make([]typeszone.GcpZone, 0)
	req := client.Zones.List(g.CloudProjectID())
	if err := req.Pages(kit.Ctx, func(page *compute.ZoneList) error {
		for _, item := range page.Items {
			parts := strings.Split(item.Region, "/")
			// strings.Split 至少返回长度为1的空串slice, 如果非空则替换为截断后的字符串
			if lastPart := parts[len(parts)-1]; len(lastPart) > 0 {
				item.Region = lastPart
			}
			zones = append(zones, typeszone.GcpZone{item})
		}
		return nil
	}); err != nil {
		logs.Errorf("failed to list zone, err: %v, rid: %s", err, kit.Rid)
	}

	return zones, nil
}

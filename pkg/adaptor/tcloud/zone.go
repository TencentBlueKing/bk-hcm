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

package tcloud

import (
	"fmt"

	typeszone "hcm/pkg/adaptor/types/zone"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

// ListZone list zone.
// reference: https://cloud.tencent.com/document/product/213/15707
func (t *TCloudImpl) ListZone(kt *kit.Kit, opt *typeszone.TCloudZoneListOption) ([]typeszone.TCloudZone, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "tcloud zone list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.CvmClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud cvm client failed, err: %v", err)
	}

	req := cvm.NewDescribeZonesRequest()
	resp, err := client.DescribeZones(req)
	if err != nil {
		logs.Errorf("list tcloud zone failed, err: %v, rid: %s", err, kt.Rid)
	}

	if resp == nil || resp.Response == nil {
		return make([]typeszone.TCloudZone, 0), nil
	}

	results := make([]typeszone.TCloudZone, 0)
	for _, one := range resp.Response.ZoneSet {
		results = append(results, typeszone.TCloudZone{one})
	}

	return results, nil
}

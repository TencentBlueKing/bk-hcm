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

package zone

import (
	"hcm/pkg/api/core/cloud/zone"
	protozone "hcm/pkg/api/data-service/cloud/zone"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/rest"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dcs/v2/model"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	"google.golang.org/api/compute/v1"
)

// TCloudZoneSync ...
type TCloudZoneSync struct {
	IsUpdate bool
	Zone     *cvm.ZoneInfo
}

// HuaWeiZoneSync ...
type HuaWeiZoneSync struct {
	IsUpdate bool
	Zone     model.AvailableZones
}

// GcpZoneSync ...
type GcpZoneSync struct {
	IsUpdate bool
	Zone     *compute.Zone
}

// AwsZoneSync ...
type AwsZoneSync struct {
	IsUpdate bool
	Zone     *ec2.AvailabilityZone
}

// DSZoneSync ...
type DSZoneSync struct {
	Zone zone.BaseZone
}

func (z *zoneHC) syncZoneDelete(cts *rest.Contexts, deleteCloudIDs []string) error {

	deleteReq := &protozone.ZoneBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", deleteCloudIDs),
	}

	err := z.dataCli.Global.Zone.BatchDeleteZone(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	if err != nil {
		return err
	}

	return nil
}

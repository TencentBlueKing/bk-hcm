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

import (
	"net/http"

	"hcm/pkg/api/core"
	protoregion "hcm/pkg/api/data-service/cloud/region"
	proto "hcm/pkg/api/hc-service"
	protodisk "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncAzureAll sync azure all resource
func SyncAzureAll(c *client.ClientSet, kit *kit.Kit, header http.Header, accountID string) error {

	resourceGroups, err := c.DataService().Azure.ResourceGroup.ListResourceGroup(
		kit.Ctx,
		header,
		&protoregion.AzureRGListReq{
			Filter: tools.EqualExpression("type", constant.SyncTimingListAzureRG),
			Page:   core.DefaultBasePage,
		},
	)
	if err != nil {
		logs.Errorf("sync list resourceGroups failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	regions, err := c.DataService().Azure.Region.ListRegion(
		kit.Ctx,
		header,
		&protoregion.AzureRegionListReq{
			Filter: tools.EqualExpression("type", constant.SyncTimingListAzureRegion),
			Page:   core.DefaultBasePage,
		},
	)
	if err != nil {
		logs.Errorf("sync list regions failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	for _, resourceGroup := range resourceGroups.Details {
		err := syncAzureWithResourceGroup(c, kit, resourceGroup.Name, header, regions, accountID)
		if err != nil {
			logs.Errorf("sync lsyncAzureWithResourceGroup failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}
	}

	return nil
}

func syncAzureWithResourceGroup(c *client.ClientSet, kit *kit.Kit, resourceGroup string,
	header http.Header, regions *protoregion.AzureRegionListResult, accountID string) error {

	var err error
	for _, region := range regions.Details {

		// sg
		err = c.HCService().Azure.SecurityGroup.SyncSecurityGroup(
			kit.Ctx,
			header,
			&proto.SecurityGroupSyncReq{
				AccountID:         accountID,
				Region:            region.Name,
				ResourceGroupName: resourceGroup,
			},
		)
		if err != nil {
			logs.Errorf("sync do azure sync sg failed, err: %v, regionID: %s, rid: %s",
				err, region.Name, kit.Rid)
		}

		// disk
		err = c.HCService().Azure.Disk.SyncDisk(
			kit.Ctx,
			header,
			&protodisk.DiskSyncReq{
				AccountID:         accountID,
				Region:            region.Name,
				ResourceGroupName: resourceGroup,
			},
		)
		if err != nil {
			logs.Errorf("sync do azure sync disk failed, err: %v, regionID: %s, rid: %s",
				err, region.Name, kit.Rid)
		}

		// Vpc(Azure在这里只需要同步Vpc即可，Vpc会自动同步关联的子网数据)
		err = c.HCService().Azure.Vpc.SyncVpc(
			kit.Ctx,
			header,
			&proto.ResourceSyncReq{
				AccountID:         accountID,
				Region:            region.Name,
				ResourceGroupName: resourceGroup,
			},
		)
		if err != nil {
			logs.Errorf("sync do azure sync vpc failed, err: %v, accountID: %s, regionID: %s, "+
				"resourceGroup: %s, rid: %s", err, accountID, region.Name, resourceGroup, kit.Rid)
		}
	}

	if err != nil {
		return err
	}

	return nil
}

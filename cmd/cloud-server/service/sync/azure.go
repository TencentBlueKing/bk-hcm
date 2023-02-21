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
	dataproto "hcm/pkg/api/data-service/cloud"
	protoregion "hcm/pkg/api/data-service/cloud/region"
	proto "hcm/pkg/api/hc-service"
	protodisk "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
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
			logs.Errorf("sync syncAzureWithResourceGroup failed, err: %v, rid: %s", err, kit.Rid)
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

		// sg rule
		err = SyncAzureSGRule(c, kit, header, region.Name, accountID)
		if err != nil {
			logs.Errorf("sync do zure sync sg rule failed, err: %v, regionID: %s,  rid: %s",
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

	return err
}

// SyncAzureImage ...
func SyncAzureImage(c *client.ClientSet, kit *kit.Kit, accountID string, header http.Header) error {

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

	for _, resourceGroup := range resourceGroups.Details {
		err = c.HCService().Azure.Image.SyncImage(
			kit.Ctx,
			header,
			&protodisk.DiskSyncReq{
				AccountID:         accountID,
				Region:            resourceGroup.Location,
				ResourceGroupName: resourceGroup.Name,
			},
		)
		// sync only one time
		if err == nil {
			break
		} else {
			logs.Errorf("sync azure image failed, err: %v, rid: %s", err, kit.Rid)
		}
	}

	return err
}

// SyncAzureSGRule ...
func SyncAzureSGRule(c *client.ClientSet, kit *kit.Kit, header http.Header,
	region string, accountID string) error {

	start := 0
	for {
		results, err := c.DataService().Global.SecurityGroup.ListSecurityGroup(
			kit.Ctx,
			header,
			&dataproto.SecurityGroupListReq{
				Filter: &filter.Expression{
					Op: filter.And,
					Rules: []filter.RuleFactory{
						&filter.AtomRule{
							Field: "vendor",
							Op:    filter.Equal.Factory(),
							Value: enumor.Azure,
						},
						&filter.AtomRule{
							Field: "region",
							Op:    filter.Equal.Factory(),
							Value: region,
						},
						&filter.AtomRule{
							Field: "account_id",
							Op:    filter.Equal.Factory(),
							Value: accountID,
						},
					},
				},
				Page: &core.BasePage{
					Start: uint32(start),
					Limit: core.DefaultMaxPageLimit,
				},
			},
		)
		if err != nil {
			logs.Errorf("list azure security group failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}

		if len(results.Details) == 0 {
			break
		}

		for _, v := range results.Details {
			err = c.HCService().Azure.SecurityGroup.SyncSecurityGroupRule(
				kit.Ctx,
				header,
				&proto.SecurityGroupSyncReq{
					AccountID: v.AccountID,
					Region:    v.Region,
				},
				v.ID,
			)
			if err != nil {
				logs.Errorf("sync do azure sync sg rule failed, err: %v, regionID: %s, rid: %s",
					err, v.Region, kit.Rid)
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < core.DefaultMaxPageLimit {
			break
		}
	}

	return nil
}
